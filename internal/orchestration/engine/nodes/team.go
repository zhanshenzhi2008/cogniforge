package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/capability/mcp"
	"cogniforge/internal/capability/skill"
	"cogniforge/internal/governance/quota"
	"cogniforge/internal/model"
	"cogniforge/internal/orchestration/engine/core"
	orchteam "cogniforge/internal/orchestration/team"
	"cogniforge/internal/runtime/chat"
)

// TeamNodeConfig 画布小队节点（绑定 agent_teams，内部走 supervisor/handoff）
type TeamNodeConfig struct {
	TeamID string `json:"team_id"`
	Task   string `json:"task,omitempty"`
	Prompt string `json:"prompt,omitempty"` // 与 task 同义，兼容画布 prompt 字段
	Goal   string `json:"goal,omitempty"`
}

// TeamNodeExecutor 在工作流节点内跑一轮小队任务；结果进 node_runs / checkpoint，不落 chat 预览
type TeamNodeExecutor struct {
	Chat  *chat.ChatService
	MCP   *mcp.Registry
	Quota *quota.Service
	DB    *gorm.DB
	Skill *skill.Service
}

func (e *TeamNodeExecutor) Execute(ctx *core.ExecutionContext, config json.RawMessage) (any, error) {
	var cfg TeamNodeConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid team config: %w", err)
	}
	if e == nil || e.Chat == nil {
		return nil, fmt.Errorf("Team node: ChatService not configured")
	}
	if e.DB == nil {
		return nil, fmt.Errorf("Team node: DB not configured")
	}
	if strings.TrimSpace(cfg.TeamID) == "" {
		return nil, fmt.Errorf("Team node: team_id is required")
	}
	userID := ctx.UserID
	if userID == "" {
		return nil, fmt.Errorf("Team node: missing user_id on execution context")
	}

	var team model.AgentTeam
	if err := e.DB.Where("id = ? AND user_id = ?", cfg.TeamID, userID).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found: %s", cfg.TeamID)
		}
		return nil, fmt.Errorf("load team: %w", err)
	}
	if team.Status != "" && team.Status != "active" {
		return nil, fmt.Errorf("team is not active: %s", cfg.TeamID)
	}

	task := resolveTemplate(cfg.Task, ctx)
	if task == "" {
		task = resolveTemplate(cfg.Prompt, ctx)
	}
	if task == "" {
		for _, key := range []string{"query", "message", "prompt", "input", "task"} {
			if v, ok := ctx.Input[key]; ok {
				task = fmt.Sprint(v)
				if task != "" {
					break
				}
			}
		}
	}
	// 流水线：上游 LLM/Agent/Team 写入的 last_agent_content
	if task == "" {
		if v, ok := ctx.GetVariable("last_agent_content"); ok {
			task = strings.TrimSpace(fmt.Sprint(v))
		}
	}
	if task == "" {
		return nil, fmt.Errorf("Team node: empty task (set config.task/prompt or input.query)")
	}
	goal := resolveTemplate(cfg.Goal, ctx)

	bg := context.Background()
	if e.Quota != nil {
		if err := e.Quota.Allow(bg, userID, "workflow"); err != nil {
			return nil, err
		}
	}

	skillSvc := e.Skill
	if skillSvc == nil {
		skillSvc = skill.NewService(e.DB)
	}
	deps := &orchteam.ChatDeps{
		ChatSvc: e.Chat,
		MCP:     e.MCP,
		DB:      e.DB,
		Skill:   skillSvc,
	}
	if ctx.OnProgress != nil {
		nodeID := ctx.NodeID
		onProg := ctx.OnProgress
		deps.OnProgress = func(message string) {
			onProg(nodeID, message)
		}
		deps.OnProgress(fmt.Sprintf("team 开始：%s", team.Name))
	}

	started := time.Now()
	result, err := deps.RunTask(bg, userID, &team, task, goal, nil)
	if err != nil {
		if e.Quota != nil {
			e.Quota.RefundRequest(bg, userID)
		}
		return nil, fmt.Errorf("team run failed: %w", err)
	}

	content := ""
	modelOut := ""
	var usage chat.ChatUsage
	if result != nil && result.Response != nil {
		content = firstChoiceContent(result.Response)
		modelOut = result.Response.Model
		usage = result.Response.Usage
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

	active := ""
	if result != nil && result.RunState != nil {
		active = result.RunState.ActiveAgentID
	}
	if content != "" {
		ctx.SetVariable("last_agent_content", content)
	}
	ctx.Logger.Logf(ctx.NodeID, "info", "Team ok id=%s active=%s model=%s latency=%dms", cfg.TeamID, active, modelOut, latency)
	if deps.OnProgress != nil {
		deps.OnProgress(fmt.Sprintf("team 完成 active=%s latency=%dms", active, latency))
	}

	out := map[string]any{
		"team_id":         cfg.TeamID,
		"active_agent_id": active,
		"task":            task,
		"content":         content,
		"response":        content,
		"model":           modelOut,
		"usage": map[string]any{
			"prompt_tokens":     usage.PromptTokens,
			"completion_tokens": usage.CompletionTokens,
			"total_tokens":      usage.TotalTokens,
		},
		"latency_ms": latency,
	}
	if result != nil && result.RunState != nil {
		out["run_state"] = result.RunState
	}
	return out, nil
}
