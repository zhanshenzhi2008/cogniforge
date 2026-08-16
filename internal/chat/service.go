package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/provider"
)

type ChatService struct {
	providerSvc *provider.Service
}

func NewChatService(providerSvc *provider.Service) *ChatService {
	return &ChatService{providerSvc: providerSvc}
}

// ListModels 返回已启用供应商在模型模块里配置的默认模型（不再使用环境变量或内置 GPT 列表）
func (s *ChatService) ListModels() *ListModelsResponse {
	models := make([]ModelInfo, 0)
	seen := make(map[string]struct{})

	add := func(m string) {
		m = strings.TrimSpace(m)
		if m == "" {
			return
		}
		if _, ok := seen[m]; ok {
			return
		}
		seen[m] = struct{}{}
		models = append(models, ModelInfo{ID: m, Name: m})
	}

	add(s.defaultModel())
	if s.providerSvc != nil {
		if cached := s.providerSvc.CachedModels(); len(cached) > 0 {
			for _, m := range cached {
				add(m.ID)
			}
			return &ListModelsResponse{Models: models}
		}
		list, err := s.providerSvc.List()
		if err == nil {
			for _, p := range list {
				if p.IsEnabled {
					add(p.DefaultModel)
				}
			}
		}
	}

	return &ListModelsResponse{Models: models}
}

func (s *ChatService) defaultModel() string {
	if s.providerSvc == nil {
		return ""
	}
	active, err := s.providerSvc.GetActive()
	if err == nil && active.DefaultModel != "" {
		return active.DefaultModel
	}
	return ""
}

// Chat 非流式对话
func (s *ChatService) Chat(req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}

	baseURL, apiKey, extraHeaders, err := s.providerSvc.GetActiveForChat()
	if err != nil {
		slog.Info("using mock AI provider (no active provider)")
		return s.mockChatResponse(req)
	}

	providerURL := s.aiChatCompletionsURL(baseURL)
	slog.Info("calling AI provider API", "url", providerURL, "model", req.Model)

	payload := s.buildPayload(req, false)
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	for k, v := range extraHeaders {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	return &chatResp, nil
}

// ChatStream 流式对话
func (s *ChatService) ChatStream(c *gin.Context, req *ChatRequest) error {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	baseURL, apiKey, extraHeaders, err := s.providerSvc.GetActiveForChat()
	if err != nil {
		slog.Info("using mock AI provider (no active provider)")
		return s.mockStreamResponse(c, req)
	}

	providerURL := s.aiChatCompletionsURL(baseURL)
	slog.Info("streaming AI provider API", "url", providerURL, "model", req.Model, "stream", true)

	payload := s.buildPayload(req, true)
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	for k, v := range extraHeaders {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := resp.Body.Read(buf)
		if n > 0 {
			c.Writer.Write(buf[:n])
			c.Writer.Flush()
			return true
		}
		return err == nil
	})
	return nil
}

func (s *ChatService) aiChatCompletionsURL(base string) string {
	return s.aiOpenAIPath(base, "chat/completions")
}

func (s *ChatService) aiEmbeddingsURL(base string) string {
	return s.aiOpenAIPath(base, "embeddings")
}

func (s *ChatService) aiOpenAIPath(base, path string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return "/v1/" + path
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/" + path
	}
	return base + "/v1/" + path
}

// Embeddings 用当前启用的 ai_providers 调上游 /v1/embeddings（与聊天同一套密钥/模型配置）
func (s *ChatService) Embeddings(req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}
	if s.providerSvc == nil {
		return nil, fmt.Errorf("no active AI provider")
	}

	baseURL, apiKey, extraHeaders, err := s.providerSvc.GetActiveForChat()
	if err != nil {
		return nil, fmt.Errorf("no active AI provider: %w", err)
	}

	providerURL := s.aiEmbeddingsURL(baseURL)
	slog.Info("calling AI embeddings API", "url", providerURL, "model", req.Model)

	payload := map[string]any{
		"model": req.Model,
		"input": req.Input,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	for k, v := range extraHeaders {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var embResp EmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}
	return &embResp, nil
}

func (s *ChatService) buildPayload(req *ChatRequest, stream bool) map[string]any {
	payload := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   stream,
	}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		payload["max_tokens"] = *req.MaxTokens
	}
	if req.TopP != nil {
		payload["top_p"] = *req.TopP
	}
	return payload
}

func (s *ChatService) mockChatResponse(req *ChatRequest) (*ChatResponse, error) {
	lastUserMsg := "hello"
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	content := fmt.Sprintf("Mock response to: %s (model: %s)", lastUserMsg, req.Model)
	if len(content) > 500 {
		content = content[:500]
	}

	usage := ChatUsage{
		PromptTokens:     len(lastUserMsg) * 2,
		CompletionTokens: len(content) / 4,
		TotalTokens:      len(lastUserMsg)*2 + len(content)/4,
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []ChatChoice{{Index: 0, Message: ChatMessage{Role: "assistant", Content: content}, FinishReason: "stop"}},
		Usage:   usage,
	}, nil
}

func (s *ChatService) mockStreamResponse(c *gin.Context, req *ChatRequest) error {
	lastUserMsg := "hello"
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	fullText := fmt.Sprintf("Mock stream response to: %s (model: %s)", lastUserMsg, req.Model)
	const chunkRunes = 12
	runes := []rune(fullText)
	words := make([]string, 0, (len(runes)+chunkRunes-1)/chunkRunes)
	for i := 0; i < len(runes); i += chunkRunes {
		end := i + chunkRunes
		if end > len(runes) {
			end = len(runes)
		}
		words = append(words, string(runes[i:end]))
	}

	eventID := fmt.Sprintf("chatcmpl-%d", time.Now().Unix())
	for i, word := range words {
		finish := ""
		if i == len(words)-1 {
			finish = "stop"
		}
		event := SSEEvent{
			ID:      eventID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []SSEChoice{{
				Index:        0,
				Delta:        map[string]any{"content": word},
				FinishReason: finish,
			}},
		}
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
		c.Writer.Flush()
	}
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	c.Writer.Flush()
	return nil
}
