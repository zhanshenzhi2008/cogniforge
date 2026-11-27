package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cogniforge/internal/governance/quota"
	"cogniforge/internal/orchestration/engine/core"
	"cogniforge/internal/runtime/chat"
)

type LLMNodeConfig struct {
	Model        string            `json:"model"`
	Prompt       string            `json:"prompt"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
	Temperature  float64           `json:"temperature,omitempty"`
	MaxTokens    int               `json:"max_tokens,omitempty"`
	Stream       bool              `json:"stream,omitempty"`
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

	prompt := resolveLLMUserPrompt(cfg.Prompt, ctx)
	systemPrompt := resolveTemplate(cfg.SystemPrompt, ctx)
	if prompt == "" && systemPrompt == "" {
		return nil, fmt.Errorf("LLM node: prompt is empty")
	}
	if prompt == "" {
		prompt = systemPrompt
		systemPrompt = ""
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

	userID := ctx.UserID
	req := &chat.ChatRequest{
		Model:    model,
		Messages: messages,
		UserID:   userID,
		Stream:   cfg.Stream,
	}
	if cfg.Temperature > 0 {
		t := cfg.Temperature
		req.Temperature = &t
	}
	if cfg.MaxTokens > 0 {
		m := cfg.MaxTokens
		req.MaxTokens = &m
	}

	bg := context.Background()
	if e.Quota != nil && userID != "" {
		if err := e.Quota.Allow(bg, userID, "workflow"); err != nil {
			return nil, err
		}
	}

	started := time.Now()
	var (
		content string
		outModel string
		usageTokens chat.ChatUsage
	)

	if cfg.Stream {
		lastReport := time.Time{}
		var totalRunes int
		collected, err := e.Chat.ChatCollectStream(bg, req, func(delta string) {
			totalRunes += len([]rune(delta))
			if ctx.OnProgress == nil {
				return
			}
			now := time.Now()
			if now.Sub(lastReport) < 400*time.Millisecond {
				return
			}
			lastReport = now
			preview := delta
			if rs := []rune(preview); len(rs) > 48 {
				preview = string(rs[len(rs)-48:])
			}
			ctx.OnProgress(ctx.NodeID, fmt.Sprintf("streaming… %d chars …%s", totalRunes, preview))
		})
		if err != nil {
			if e.Quota != nil && userID != "" {
				e.Quota.RefundRequest(bg, userID)
			}
			return nil, fmt.Errorf("LLM stream failed: %w", err)
		}
		content = collected.Content
		outModel = collected.Model
		usageTokens = collected.Usage
	} else {
		resp, err := e.Chat.Chat(req)
		if err != nil {
			if e.Quota != nil && userID != "" {
				e.Quota.RefundRequest(bg, userID)
			}
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}
		content = firstChoiceContent(resp)
		if resp != nil {
			outModel = resp.Model
			usageTokens = resp.Usage
		}
	}

	latency := time.Since(started).Milliseconds()
	usage := map[string]any{
		"prompt_tokens":     usageTokens.PromptTokens,
		"completion_tokens": usageTokens.CompletionTokens,
		"total_tokens":      usageTokens.TotalTokens,
		"estimated":         usageTokens.Estimated,
	}
	if e.Quota != nil && userID != "" {
		e.Quota.Commit(bg, quota.CommitInput{
			UserID:           userID,
			Source:           "workflow",
			Model:            outModel,
			PromptTokens:     usageTokens.PromptTokens,
			CompletionTokens: usageTokens.CompletionTokens,
			TotalTokens:      usageTokens.TotalTokens,
			Estimated:        usageTokens.Estimated,
			Status:           "ok",
			LatencyMS:        latency,
			TraceID:          ctx.TraceID,
		})
	}

	mode := "sync"
	if cfg.Stream {
		mode = "stream"
	}
	// 与 Agent 共用：下游空 prompt / 检索问句可接 last_agent_content
	ctx.SetVariable("last_agent_content", content)
	ctx.Logger.Logf(ctx.NodeID, "info", "LLM ok mode=%s model=%s latency=%dms content_len=%d", mode, outModel, latency, len(content))

	return map[string]any{
		"model":      outModel,
		"content":    content,
		"response":   content, // 兼容旧调试字段名
		"usage":      usage,
		"latency_ms": latency,
		"streamed":   cfg.Stream,
	}, nil
}

// resolveLLMUserPrompt：配置模板 → 上游 last_agent_content → 默认拼上下文
func resolveLLMUserPrompt(cfgPrompt string, ctx *core.ExecutionContext) string {
	prompt := resolveTemplate(cfgPrompt, ctx)
	if strings.TrimSpace(prompt) != "" {
		return prompt
	}
	if ctx != nil {
		if v, ok := ctx.GetVariable("last_agent_content"); ok {
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return s
			}
		}
	}
	return defaultLLMUserPrompt(ctx)
}

func defaultLLMUserPrompt(ctx *core.ExecutionContext) string {
	if ctx == nil {
		return ""
	}
	payload := map[string]any{}
	if len(ctx.Input) > 0 {
		payload["input"] = ctx.Input
	}
	if len(ctx.Output) > 0 {
		payload["previous_outputs"] = ctx.Output
	}
	if len(payload) == 0 {
		return "请根据系统提示完成任务。"
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Sprint(payload)
	}
	return "请根据以下上下文完成任务：\n" + string(b)
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
