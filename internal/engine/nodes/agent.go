package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/chat"
	"cogniforge/internal/engine/core"
	"cogniforge/internal/mcp"
	"cogniforge/internal/model"
	promptpkg "cogniforge/internal/prompt"
	"cogniforge/internal/quota"
	"cogniforge/internal/skill"
)

type AgentNodeConfig struct {
	AgentID     string            `json:"agent_id"`
	Prompt      string            `json:"prompt,omitempty"`
	Tools       []string          `json:"tools,omitempty"` // 可选覆盖；空则用 Agent.tools
	MemoryTurns int               `json:"memory_turns,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Model       string            `json:"model,omitempty"`
}

// AgentNodeExecutor 加载 Agent 配置并走 Chat / ChatWithTools（阶段十八 W0）
type AgentNodeExecutor struct {
	Chat  *chat.ChatService
	MCP   *mcp.Registry
	Quota *quota.Service
	DB    *gorm.DB
}

func (e *AgentNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg AgentNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}
	if e == nil || e.Chat == nil {
		return nil, fmt.Errorf("Agent node: ChatService not configured")
	}
	if e.DB == nil {
		return nil, fmt.Errorf("Agent node: DB not configured")
	}
	if cfg.AgentID == "" {
		return nil, fmt.Errorf("Agent node: agent_id is required")
	}
	userID := ctx.UserID
	if userID == "" {
		return nil, fmt.Errorf("Agent node: missing user_id on execution context")
	}

	for k, v := range cfg.Variables {
		ctx.SetVariable(k, v)
	}

	var ag model.Agent
	if err := e.DB.Where("id = ? AND user_id = ?", cfg.AgentID, userID).
		Select("id, user_id, model, system_prompt, prompt_id, skill_id, memory_type, memory_turns, tools").
		First(&ag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("agent not found: %s", cfg.AgentID)
		}
		return nil, fmt.Errorf("load agent: %w", err)
	}

	userPrompt := resolveTemplate(cfg.Prompt, ctx)
	if userPrompt == "" {
		// 兼容：从执行 input 取 query / message / prompt
		for _, key := range []string{"query", "message", "prompt", "input"} {
			if v, ok := ctx.Input[key]; ok {
				userPrompt = fmt.Sprint(v)
				if userPrompt != "" {
					break
				}
			}
		}
	}
	if userPrompt == "" {
		return nil, fmt.Errorf("Agent node: empty user prompt (set config.prompt or input.query)")
	}

	systemPrompt := e.resolveSystemPrompt(userID, &ag, cfg.Variables)
	modelName := cfg.Model
	if modelName == "" {
		modelName = ag.Model
	}

	messages := []chat.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	req := &chat.ChatRequest{
		Model:    modelName,
		Messages: messages,
	}

	toolIDs := cfg.Tools
	if len(toolIDs) == 0 {
		toolIDs = []string(ag.Tools)
	}
	chat.InjectMCPTools(req, e.MCP, toolIDs, userID)
	hasTools := len(req.Tools) > 0

	bg := context.Background()
	if e.Quota != nil {
		if err := e.Quota.Allow(bg, userID, "workflow"); err != nil {
			return nil, err
		}
	}

	started := time.Now()
	var (
		content string
		usage   chat.ChatUsage
		modelOut string
		toolsUsed int
	)

	if hasTools {
		resp, err := e.Chat.ChatWithTools(req, chat.NewMCPToolExecutor(e.MCP))
		if err != nil {
			if e.Quota != nil {
				e.Quota.RefundRequest(bg, userID)
			}
			return nil, fmt.Errorf("agent tool chat failed: %w", err)
		}
		if resp == nil {
			if e.Quota != nil {
				e.Quota.RefundRequest(bg, userID)
			}
			return nil, fmt.Errorf("empty agent tool chat response")
		}
		content = firstChoiceContent(resp)
		usage = resp.Usage
		modelOut = resp.Model
		toolsUsed = len(toolIDs)
	} else {
		resp, err := e.Chat.Chat(req)
		if err != nil {
			if e.Quota != nil {
				e.Quota.RefundRequest(bg, userID)
			}
			return nil, fmt.Errorf("agent chat failed: %w", err)
		}
		content = firstChoiceContent(resp)
		if resp != nil {
			usage = resp.Usage
			modelOut = resp.Model
		}
	}

	latency := time.Since(started).Milliseconds()
	if e.Quota != nil {
		e.Quota.Commit(bg, quota.CommitInput{
			UserID:           userID,
			Source:           "workflow",
			Model:            modelOut,
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
			Estimated:        usage.Estimated,
			Status:           "ok",
			LatencyMS:        latency,
			TraceID:          ctx.TraceID,
		})
	}

	ctx.Logger.Logf(ctx.NodeID, "info", "Agent ok id=%s model=%s tools=%d latency=%dms", cfg.AgentID, modelOut, toolsUsed, latency)

	return map[string]any{
		"agent_id":     cfg.AgentID,
		"model":        modelOut,
		"prompt":       userPrompt,
		"content":      content,
		"response":     content,
		"tools_used":   toolIDs,
		"memory_turns": ag.MemoryTurns,
		"usage": map[string]any{
			"prompt_tokens":     usage.PromptTokens,
			"completion_tokens": usage.CompletionTokens,
			"total_tokens":      usage.TotalTokens,
		},
		"latency_ms": latency,
	}, nil
}

func (e *AgentNodeExecutor) resolveSystemPrompt(userID string, ag *model.Agent, vars map[string]string) string {
	systemPrompt := ag.SystemPrompt
	if ag.PromptID != "" {
		rendered, err := promptpkg.NewService(e.DB).ResolvePublishedContent(userID, ag.PromptID, vars)
		if err == nil && rendered != "" {
			systemPrompt = rendered
		}
	} else if ag.SkillID != "" {
		if s := model.GetBuiltInSkill(ag.SkillID); s != nil {
			systemPrompt = s.BuildSystemPrompt()
		} else {
			if dbSkill, err := skill.NewService(e.DB).GetSkill(userID, ag.SkillID); err == nil && dbSkill != nil {
				systemPrompt = dbSkill.BuildSystemPrompt()
			}
		}
	}
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}
	return systemPrompt
}
