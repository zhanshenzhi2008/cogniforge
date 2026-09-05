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

	"cogniforge/internal/model"
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
				if p.IsEnabled && model.HasCapability(p.Capabilities, model.CapChat) {
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

// activeChatConfig 取当前可用供应商。没有默认模型或 Key 时返回 ErrNoActiveProvider，不再走 mock。
func (s *ChatService) activeChatConfig() (baseURL, apiKey string, headers map[string]string, err error) {
	if s.providerSvc == nil {
		return "", "", nil, ErrNoActiveProvider
	}
	baseURL, apiKey, headers, err = s.providerSvc.GetActiveForChat()
	if err != nil {
		slog.Warn("chat skipped: no usable AI provider", "error", err)
		return "", "", nil, ErrNoActiveProvider
	}
	if strings.TrimSpace(apiKey) == "" {
		slog.Warn("chat skipped: active provider has empty API key")
		return "", "", nil, ErrNoActiveProvider
	}
	return baseURL, apiKey, headers, nil
}

// Chat 非流式对话
func (s *ChatService) Chat(req *ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}

	baseURL, apiKey, extraHeaders, err := s.activeChatConfig()
	if err != nil {
		return nil, err
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

// ChatWithTools 带工具调用的对话（阶段十五 15.2）
// 调用循环：发消息 → 检查 tool_calls → 执行工具 → 把结果塞回 → 继续发
// MaxToolCalls = 5 硬上限，防止无限循环；实际次数由模型自主决定（没有 tool_calls 即停）
const MaxToolCalls = 5

func (s *ChatService) ChatWithTools(req *ChatRequest, toolExecutor ToolExecutor) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}

	baseURL, apiKey, extraHeaders, err := s.activeChatConfig()
	if err != nil {
		return nil, err
	}

	providerURL := s.aiChatCompletionsURL(baseURL)

	// 动态循环：最多 MaxToolCalls 次，模型说没有 tool_calls 就停
	for i := 0; i < MaxToolCalls; i++ {
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

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
		}

		var chatResp ChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		// 优先提取 tool_use_message_groups（Claude 格式）
		toolCalls := extractToolCalls(&chatResp)
		if len(toolCalls) == 0 {
			// 没有任何工具调用 → 模型决定结束，返回
			return &chatResp, nil
		}

		// 执行工具并把结果追加到 messages
		for _, tc := range toolCalls {
			result, execErr := toolExecutor.Execute(tc)
			content := "Error: " + execErr.Error()
			if execErr == nil {
				content = result
			}
			req.Messages = append(req.Messages, ChatMessage{
				Role:    "tool",
				Content: content,
			})
		}
		// 继续下一次循环，把结果发给模型让它决定下一步
	}

	// 达到 MaxToolCalls 硬上限，不再发请求，直接返回已有结果
	// 实际应用中建议此时给出警告日志
	return nil, fmt.Errorf("tool call limit (%d) reached, please reduce tool usage", MaxToolCalls)
}

// extractToolCalls 从 chat response 中提取 tool_calls
func extractToolCalls(resp *ChatResponse) []ToolCall {
	var calls []ToolCall
	for _, choice := range resp.Choices {
		for _, tcRaw := range choice.ToolCalls {
			tc := ToolCall{}
			if id, ok := tcRaw["id"].(string); ok {
				tc.ID = id
			}
			if t, ok := tcRaw["type"].(string); ok {
				tc.Type = t
			}
			if fn, ok := tcRaw["function"].(map[string]any); ok {
				tc.Function.Name, _ = fn["name"].(string)
				tc.Function.Arguments, _ = fn["arguments"].(string)
			}
			calls = append(calls, tc)
		}
	}
	return calls
}

// ChatStream 流式对话。返回本次 usage（上游没有则估算）。
func (s *ChatService) ChatStream(c *gin.Context, req *ChatRequest) (*ChatUsage, error) {
	if req.Model == "" {
		req.Model = s.defaultModel()
	}

	baseURL, apiKey, extraHeaders, err := s.activeChatConfig()
	if err != nil {
		return nil, err
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	providerURL := s.aiChatCompletionsURL(baseURL)
	slog.Info("streaming AI provider API", "url", providerURL, "model", req.Model, "stream", true)

	payload := s.buildPayload(req, true)
	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI provider returned status %d: %s", resp.StatusCode, string(respBody))
	}

	scan := &sseUsageScan{}
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := resp.Body.Read(buf)
		if n > 0 {
			scan.feed(buf[:n])
			_, _ = c.Writer.Write(buf[:n])
			c.Writer.Flush()
			return true
		}
		return err == nil
	})
	usage := scan.usage
	if usage.TotalTokens == 0 {
		usage = EstimateUsage(req.Messages, scan.text.String())
		usage.Estimated = true
	}
	return &usage, nil
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

// Embeddings 用「向量默认」供应商调上游 /v1/embeddings，不复用对话模型。
func (s *ChatService) Embeddings(req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	if s.providerSvc == nil {
		return nil, fmt.Errorf("no embedding provider configured")
	}
	baseURL, apiKey, extraHeaders, embedModel, err := s.providerSvc.GetActiveForEmbedding()
	if err != nil {
		return nil, fmt.Errorf("no embedding provider: %w", err)
	}
	req.Model = embedModel
	if req.Model == "" {
		return nil, fmt.Errorf("no embedding model configured")
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
	if stream {
		payload["stream_options"] = map[string]any{"include_usage": true}
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
	// 阶段十五 15.2：注入 MCP 工具
	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
	}
	return payload
}
