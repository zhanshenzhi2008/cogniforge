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

	"cogniforge/internal/config"
)

type ChatService struct {
	cfg           *config.Config
	builtInModels []string
}

func NewChatService(cfg *config.Config) *ChatService {
	return &ChatService{
		cfg: cfg,
		builtInModels: []string{
			"gpt-3.5-turbo",
			"gpt-3.5-turbo-0125",
			"gpt-3.5-turbo-0301",
			"gpt-3.5-turbo-0302",
			"gpt-3.5-turbo-0613",
			"gpt-3.5-turbo-0615",
			"gpt-3.5-turbo-1106",
			"gpt-3.5-turbo-1107",
			"gpt-3.5-turbo-16k",
			"gpt-3.5-turbo-16k-0613",
			"gpt-3.5-turbo-instruct",
			"text-davinci-003",
			"text-embedding-3-large",
			"text-embedding-3-small",
			"text-embedding-ada-002",
			"tts-1",
			"tts-1-1106",
			"tts-1-hd",
			"tts-1-hd-1106",
			"whisper-1",
		},
	}
}

// ListModels 获取模型列表
func (s *ChatService) ListModels() *ListModelsResponse {
	models := make([]ModelInfo, 0, len(s.builtInModels)+1)
	seen := make(map[string]struct{}, len(s.builtInModels)+1)

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
	for _, m := range s.builtInModels {
		add(m)
	}

	return &ListModelsResponse{Models: models}
}

func (s *ChatService) defaultModel() string {
	if s.cfg != nil && strings.TrimSpace(s.cfg.AI.DefaultModel) != "" {
		return s.cfg.AI.DefaultModel
	}
	return "gpt-3.5-turbo"
}

// Chat 非流式对话
func (s *ChatService) Chat(req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}

	if s.cfg == nil || s.cfg.AI.APIKey == "" {
		slog.Info("using mock AI provider (api_key empty)")
		return s.mockChatResponse(req)
	}

	providerURL := s.aiChatCompletionsURL(s.cfg.AI.BaseURL)
	slog.Info("calling AI provider API", "url", providerURL, "model", req.Model)

	payload := s.buildPayload(req)
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.APIKey)

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

	if s.cfg == nil || s.cfg.AI.APIKey == "" {
		slog.Info("using mock AI provider (api_key empty)")
		return s.mockStreamResponse(c, req)
	}

	providerURL := s.aiChatCompletionsURL(s.cfg.AI.BaseURL)
	slog.Info("streaming AI provider API", "url", providerURL, "model", req.Model)

	payload := s.buildPayload(req)
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.APIKey)

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
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return "/v1/chat/completions"
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

func (s *ChatService) buildPayload(req *ChatRequest) map[string]any {
	payload := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
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
