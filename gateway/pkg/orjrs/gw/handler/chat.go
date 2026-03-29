package handler

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
	"github.com/orjrs/gateway/pkg/orjrs/gw/config"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`

	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	Choices           []ChatChoice `json:"choices"`
	Usage             ChatUsage    `json:"usage"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
}

type SSEEvent struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	Choices           []SSEChoice `json:"choices"`
	SystemFingerprint string      `json:"system_fingerprint,omitempty"`
}

type SSEChoice struct {
	Index        int            `json:"index"`
	Delta        map[string]any `json:"delta"`
	FinishReason string         `json:"finish_reason,omitempty"`
}

var chatCfg *config.Config

func ListModels(c *gin.Context) {
	var models []map[string]string
	if chatCfg != nil && chatCfg.AI.APIKey != "" {
		models = []map[string]string{
			{"id": chatCfg.AI.DefaultModel, "name": chatCfg.AI.DefaultModel},
		}
	} else {
		models = []map[string]string{
			{"id": "mock-gpt-4o", "name": "Mock GPT-4o (no API key)"},
		}
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func SetChatConfig(cfg *config.Config) {
	chatCfg = cfg
	slog.Info("AI config loaded",
		"provider", cfg.AI.Provider,
		"base_url", cfg.AI.BaseURL,
		"api_key_set", cfg.AI.APIKey != "",
		"default_model", cfg.AI.DefaultModel,
	)
}

// aiChatCompletionsURL builds the provider URL. Base may be host only (https://api.openai.com)
// or already include /v1 (https://api.xty.app/v1).
func aiChatCompletionsURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return "/v1/chat/completions"
	}
	if strings.HasSuffix(base, "/v1") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

// Chat handles non-streaming chat completions.
func Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages is required and cannot be empty"})
		return
	}

	if req.Model == "" {
		req.Model = "gpt-4o"
	}

	resp, err := callAIProvider(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI provider error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ChatStream handles streaming chat completions (SSE).
func ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages is required and cannot be empty"})
		return
	}

	if req.Model == "" {
		req.Model = "gpt-4o"
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	if err := streamAIProvider(c, req); err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
		c.Writer.Flush()
	}
}

// callAIProvider forwards the chat request to the configured AI provider.
// If AI_API_KEY is not set, returns a mock response for development.
func callAIProvider(req ChatRequest) (*ChatResponse, error) {
	if chatCfg == nil || chatCfg.AI.APIKey == "" {
		slog.Info("using mock AI provider (api_key empty)")
		return mockChatResponse(req)
	}

	providerURL := aiChatCompletionsURL(chatCfg.AI.BaseURL)
	slog.Info("calling AI provider API", "url", providerURL, "model", req.Model)

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

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+chatCfg.AI.APIKey)

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

// streamAIProvider forwards a streaming request to the AI provider and streams SSE to the client.
// If AI_API_KEY is not set, streams a mock response.
func streamAIProvider(c *gin.Context, req ChatRequest) error {
	if chatCfg == nil || chatCfg.AI.APIKey == "" {
		slog.Info("using mock AI provider (api_key empty)")
		return mockStreamResponse(c, req)
	}

	providerURL := aiChatCompletionsURL(chatCfg.AI.BaseURL)
	slog.Info("streaming AI provider API", "url", providerURL, "model", req.Model)

	payload := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
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

	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", providerURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+chatCfg.AI.APIKey)

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

// mockChatResponse returns a simulated response for development without an AI API key.
func mockChatResponse(req ChatRequest) (*ChatResponse, error) {
	lastUserMsg := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}
	if lastUserMsg == "" {
		lastUserMsg = "hello"
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

// mockStreamResponse streams a mock SSE response for development.
func mockStreamResponse(c *gin.Context, req ChatRequest) error {
	lastUserMsg := "hello"
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	fullText := fmt.Sprintf("Mock stream response to: %s (model: %s)", lastUserMsg, req.Model)
	// Chunk by runes so UTF-8 (e.g. Chinese) is not split mid-character; fullText keeps spaces readable.
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
