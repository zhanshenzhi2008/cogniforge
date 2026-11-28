package team

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"cogniforge/internal/capability/mcp"
	capprompt "cogniforge/internal/capability/prompt"
	"cogniforge/internal/capability/skill"
	"cogniforge/internal/model"
	"cogniforge/internal/runtime/chat"
)

// ChatDeps 小队对话依赖（由 app 层注入）
type ChatDeps struct {
	ChatSvc *chat.ChatService
	MCP     *mcp.Registry
	DB      *gorm.DB
	Skill   *skill.Service
	// OnProgress 可选：delegate/handoff 子步骤进度（画布 team 节点 → ExecutionContext.OnProgress）
	OnProgress func(message string)
}

// ChatResult 一次小队对话结果
type ChatResult struct {
	Response *chat.ChatResponse `json:"response"`
	RunState *RunState          `json:"run_state"`
}

// RunChat 以小队当前 active agent 跑一轮；prior 非空时续跑（handoff 后换人设）
func (d *ChatDeps) RunChat(ctx context.Context, userID string, t *model.AgentTeam, messages []chat.ChatMessage, goal string, prior *RunState) (*ChatResult, error) {
	_ = ctx
	if d == nil || d.ChatSvc == nil || d.DB == nil {
		return nil, fmt.Errorf("team chat deps incomplete")
	}
	if t == nil {
		return nil, fmt.Errorf("team is nil")
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages 不能为空")
	}
	state, err := ResumeRunState(t, prior, goal)
	if err != nil {
		return nil, err
	}
	if state.Goal == "" {
		state.Goal = lastUserText(messages)
	}

	ag, err := d.loadAgent(userID, state.ActiveAgentID)
	if err != nil {
		return nil, err
	}
	systemPrompt, err := d.resolveSystemPrompt(userID, ag, nil)
	if err != nil {
		return nil, err
	}
	isSupervisor := t.SupervisorAgentID != "" && state.ActiveAgentID == t.SupervisorAgentID
	roster := d.buildRosterHint(userID, t, isSupervisor)
	systemPrompt = systemPrompt + "\n\n" + roster

	svcReq := &chat.ChatRequest{
		Model:    ag.Model,
		Messages: append([]chat.ChatMessage{{Role: "system", Content: systemPrompt}}, messages...),
		UserID:   userID,
		AgentID:  ag.ID,
	}
	chat.InjectMCPTools(svcReq, d.MCP, []string(ag.Tools), userID)
	if isSupervisor || t.Mode == model.TeamModeHandoff || t.Mode == model.TeamModeSupervisor {
		svcReq.Tools = append(svcReq.Tools, OrchestrationToolSpecs()...)
	}

	exec := &compositeExecutor{
		deps:   d,
		userID: userID,
		team:   t,
		state:  state,
		inner:  chat.NewMCPToolExecutor(d.MCP),
	}

	var resp *chat.ChatResponse
	if len(svcReq.Tools) > 0 {
		resp, err = d.ChatSvc.ChatWithTools(svcReq, exec)
	} else {
		resp, err = d.ChatSvc.Chat(svcReq)
	}
	if err != nil {
		return nil, err
	}
	return &ChatResult{Response: resp, RunState: state}, nil
}

func (d *ChatDeps) loadAgent(userID, agentID string) (*model.Agent, error) {
	var ag model.Agent
	if err := d.DB.Where("id = ? AND user_id = ?", agentID, userID).
		Select("id, user_id, name, model, system_prompt, prompt_id, skill_id, memory_type, memory_turns, tools, status").
		First(&ag).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("Agent 不存在: %s", agentID)
		}
		return nil, fmt.Errorf("查询 Agent 失败")
	}
	return &ag, nil
}

func (d *ChatDeps) resolveSystemPrompt(userID string, ag *model.Agent, variables map[string]string) (string, error) {
	systemPrompt := ag.SystemPrompt
	if ag.PromptID != "" {
		rendered, err := capprompt.NewService(d.DB).ResolvePublishedContent(userID, ag.PromptID, variables)
		if err != nil {
			return "", err
		}
		if rendered != "" {
			systemPrompt = rendered
		}
	} else if ag.SkillID != "" {
		if s := model.GetBuiltInSkill(ag.SkillID); s != nil {
			systemPrompt = s.BuildSystemPrompt()
		} else if d.Skill != nil {
			if dbSkill, err := d.Skill.GetSkill(userID, ag.SkillID); err == nil {
				systemPrompt = dbSkill.BuildSystemPrompt()
			}
		}
	}
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = "You are a helpful AI assistant."
	}
	return systemPrompt, nil
}

func (d *ChatDeps) buildRosterHint(userID string, t *model.AgentTeam, isSupervisor bool) string {
	var b strings.Builder
	b.WriteString("【小队编排】你属于一支多 Agent 小队。\n")
	if isSupervisor {
		b.WriteString("你是主管。可用工具：delegate_to_agent（委派工人）、handoff_to_agent（交接对话权）。\n")
	} else {
		b.WriteString("你当前持有对话权（交接后接手）。可用工具：handoff_to_agent（再交棒）；不要假装主管去派活，除非用户明确要求。\n")
	}
	b.WriteString(fmt.Sprintf("模式：%s；小队：%s；当前角色即你。\n", t.Mode, t.Name))
	if t.SupervisorAgentID != "" {
		b.WriteString(fmt.Sprintf("主管 Agent ID：%s\n", t.SupervisorAgentID))
	}
	b.WriteString("成员：\n")
	ids := append([]string{}, t.MemberAgentIDs...)
	if t.SupervisorAgentID != "" {
		ids = append(ids, t.SupervisorAgentID)
	}
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		name := id
		if ag, err := d.loadAgent(userID, id); err == nil && ag.Name != "" {
			name = ag.Name
		}
		b.WriteString(fmt.Sprintf("- %s (%s)\n", name, id))
	}
	b.WriteString("只能委派/交接给上述成员。工人失败时说明原因，不要编造成果。")
	return b.String()
}

func lastUserText(messages []chat.ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "user" {
			continue
		}
		switch v := messages[i].Content.(type) {
		case string:
			return v
		default:
			b, _ := json.Marshal(v)
			return string(b)
		}
	}
	return ""
}

func teamAllowsAgent(t *model.AgentTeam, agentID string) bool {
	if t == nil || agentID == "" {
		return false
	}
	if t.SupervisorAgentID == agentID {
		return true
	}
	for _, id := range t.MemberAgentIDs {
		if id == agentID {
			return true
		}
	}
	return false
}

type compositeExecutor struct {
	deps   *ChatDeps
	userID string
	team   *model.AgentTeam
	state  *RunState
	inner  chat.ToolExecutor
}

func (e *compositeExecutor) report(msg string) {
	if e == nil || e.deps == nil || e.deps.OnProgress == nil || strings.TrimSpace(msg) == "" {
		return
	}
	e.deps.OnProgress(msg)
}

func (e *compositeExecutor) Execute(call chat.ToolCall) (string, error) {
	name := call.Function.Name
	switch name {
	case ToolDelegate:
		return e.delegate(call)
	case ToolHandoff:
		return e.handoff(call)
	default:
		if e.inner == nil {
			return "", fmt.Errorf("unknown tool: %s", name)
		}
		return e.inner.Execute(call)
	}
}

func (e *compositeExecutor) delegate(call chat.ToolCall) (string, error) {
	args, err := parseToolArgs(call.Function.Arguments)
	if err != nil {
		return "", err
	}
	workerID := argString(args, "agent_id")
	task := argString(args, "task")
	if workerID == "" || task == "" {
		return "", fmt.Errorf("delegate_to_agent 需要 agent_id 与 task")
	}
	if !teamAllowsAgent(e.team, workerID) {
		return "", fmt.Errorf("agent %s 不在小队成员中", workerID)
	}
	if err := e.state.RecordDelegate(workerID, task, ""); err != nil {
		return "", err
	}
	workerName := workerID
	if ag, err := e.deps.loadAgent(e.userID, workerID); err == nil {
		workerName = ag.Name
	}
	e.report(fmt.Sprintf("delegate → %s：%s", workerName, truncate(task, 80)))
	summary, err := e.runWorker(workerID, task)
	if err != nil && workerRetryOnce(e.team) {
		e.report(fmt.Sprintf("delegate 重试 → %s", workerName))
		summary, err = e.runWorker(workerID, task)
	}
	if err != nil {
		// 软失败：回传主管，不中断工具循环（设计 §12.4）
		msg := "工人执行失败（已回传，请改派或说明）：" + err.Error()
		if n := len(e.state.DelegateHistory); n > 0 {
			e.state.DelegateHistory[n-1].ResultSummary = truncate(msg, 500)
		}
		e.state.SetScratch("last_delegate_error_"+workerID, msg)
		e.report(fmt.Sprintf("delegate 失败 ← %s：%s", workerName, truncate(err.Error(), 80)))
		return msg, nil
	}
	// 回填摘要
	if n := len(e.state.DelegateHistory); n > 0 {
		e.state.DelegateHistory[n-1].ResultSummary = truncate(summary, 500)
	}
	e.state.SetScratch("last_delegate_"+workerID, summary)
	e.report(fmt.Sprintf("delegate 完成 ← %s（%d 字）", workerName, len([]rune(summary))))
	return summary, nil
}

func workerRetryOnce(t *model.AgentTeam) bool {
	if t == nil || t.Config == nil {
		return false
	}
	v, ok := t.Config["worker_retry_once"]
	if !ok {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return strings.EqualFold(x, "true") || x == "1"
	default:
		return false
	}
}

func (e *compositeExecutor) handoff(call chat.ToolCall) (string, error) {
	args, err := parseToolArgs(call.Function.Arguments)
	if err != nil {
		return "", err
	}
	toID := argString(args, "agent_id")
	reason := argString(args, "reason")
	if !teamAllowsAgent(e.team, toID) {
		return "", fmt.Errorf("agent %s 不在小队成员中", toID)
	}
	if err := e.state.Handoff(toID, reason); err != nil {
		return "", err
	}
	name := toID
	if ag, err := e.deps.loadAgent(e.userID, toID); err == nil {
		name = ag.Name
	}
	reasonHint := reason
	if reasonHint == "" {
		reasonHint = "无说明"
	}
	e.report(fmt.Sprintf("handoff → %s：%s", name, truncate(reasonHint, 80)))
	return fmt.Sprintf("已交接给 %s (%s)。请用简短话告知用户后续由对方接手；本轮不要再长篇作答。", name, toID), nil
}

func (e *compositeExecutor) runWorker(workerID, task string) (string, error) {
	ag, err := e.deps.loadAgent(e.userID, workerID)
	if err != nil {
		return "", err
	}
	systemPrompt, err := e.deps.resolveSystemPrompt(e.userID, ag, nil)
	if err != nil {
		return "", err
	}
	systemPrompt += "\n\n你是小队工人。只完成委派任务，直接给出结果，不要扮演主管或再委派他人。"
	userMsg := task
	if e.state.Goal != "" {
		userMsg = "总目标：" + e.state.Goal + "\n\n本轮任务：" + task
	}
	req := &chat.ChatRequest{
		Model: ag.Model,
		Messages: []chat.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMsg},
		},
		UserID:  e.userID,
		AgentID: ag.ID,
	}
	chat.InjectMCPTools(req, e.deps.MCP, []string(ag.Tools), e.userID)
	if len(req.Tools) > 0 {
		resp, err := e.deps.ChatSvc.ChatWithTools(req, chat.NewMCPToolExecutor(e.deps.MCP))
		if err != nil {
			return "", fmt.Errorf("工人执行失败: %w", err)
		}
		return extractAssistantText(resp), nil
	}
	resp, err := e.deps.ChatSvc.Chat(req)
	if err != nil {
		return "", fmt.Errorf("工人执行失败: %w", err)
	}
	return extractAssistantText(resp), nil
}

func extractAssistantText(resp *chat.ChatResponse) string {
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

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
