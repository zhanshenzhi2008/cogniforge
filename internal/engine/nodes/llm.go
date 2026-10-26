package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cogniforge/internal/chat"
	"cogniforge/internal/engine/core"
	"cogniforge/internal/quota"
)

type LLMNodeConfig struct {
	Model        string            `json:"model"`
	Prompt       string            `json:"prompt"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Temperature  float64           `json:"temperature,omitempty"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	Variables    map[string]string `json:"variables,omitempty"`
}

// LLMNodeExecutor 调用真实 ChatService（阶段十八 W0）
type LLMNodeExecutor struct {
	Chat  *chat.ChatService
	Quota *quota.Service
}

func (e *LLMNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg LLMNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid LLM config: %w", err)
	}
	if e == nil || e.Chat == nil {
		return nil, fmt.Errorf("LLM node: ChatService not configured")
	}

	// 节点级 variables 写入上下文，便于 {{var}}
	for k, v := range cfg.Variables {
		ctx.SetVariable(k, v)
	}

	prompt := resolveTemplate(cfg.Prompt, ctx)
	systemPrompt := resolveTemplate(cfg.SystemPrompt, ctx)
	if prompt == "" {
		return nil, fmt.Errorf("LLM node: prompt is empty")
	}

	model := cfg.Model
	if model == "" {
		if m, ok := ctx.Variables["model"].(string); ok {
			model = m
		}
	}

	messages := make([]chat.ChatMessage, 0, 2)
	if systemPrompt != "" {
		messages = append(messages, chat.ChatMessage{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, chat.ChatMessage{Role: "user", Content: prompt})

	req := &chat.ChatRequest{
		Model:    model,
		Messages: messages,
	}
	if cfg.Temperature > 0 {
		t := cfg.Temperature
		req.Temperature = &t
	}
	if cfg.MaxTokens > 0 {
		m := cfg.MaxTokens
		req.MaxTokens = &m
	}

	userID := ctx.UserID
	bg := context.Background()
	if e.Quota != nil && userID != "" {
		if err := e.Quota.Allow(bg, userID, "workflow"); err != nil {
			return nil, err
		}
	}

	started := time.Now()
	resp, err := e.Chat.Chat(req)
	if err != nil {
		if e.Quota != nil && userID != "" {
			e.Quota.RefundRequest(bg, userID)
		}
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	content := firstChoiceContent(resp)
	latency := time.Since(started).Milliseconds()
	usage := map[string]any{}
	if resp != nil {
		usage["prompt_tokens"] = resp.Usage.PromptTokens
		usage["completion_tokens"] = resp.Usage.CompletionTokens
		usage["total_tokens"] = resp.Usage.TotalTokens
		if e.Quota != nil && userID != "" {
			e.Quota.Commit(bg, quota.CommitInput{
				UserID:           userID,
				Source:           "workflow",
				Model:            resp.Model,
				PromptTokens:     resp.Usage.PromptTokens,
				CompletionTokens: resp.Usage.CompletionTokens,
				TotalTokens:      resp.Usage.TotalTokens,
				Estimated:        resp.Usage.Estimated,
				Status:           "ok",
				LatencyMS:        latency,
				TraceID:          ctx.TraceID,
			})
		}
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "LLM ok model=%s latency=%dms content_len=%d", model, latency, len(content))

	return map[string]any{
		"model":      resp.Model,
		"content":    content,
		"response":   content, // 兼容旧调试字段名
		"usage":      usage,
		"latency_ms": latency,
	}, nil
}

func firstChoiceContent(resp *chat.ChatResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	c := resp.Choices[0].Message.Content
	switch v := c.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
