package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/chat"
	"cogniforge/internal/database"
	"cogniforge/internal/mcp"
	"cogniforge/internal/memory"
	agentmodel "cogniforge/internal/model"
	"cogniforge/internal/provider"
	"cogniforge/internal/quota"
	"cogniforge/internal/response"
	"cogniforge/internal/skill"
)

// GetBuiltInSkill 在 skill 包里单独导出（避免循环引用）
func getBuiltInSkill(id string) *agentmodel.Skill {
	return agentmodel.GetBuiltInSkill(id)
}

type AgentHandler struct {
	service     *AgentService
	providerSvc *provider.Service
	chatSvc     *chat.ChatService
	quota       *quota.Service
	mcpRegistry *mcp.Registry
	skillSvc   *skill.Service
}

func NewAgentHandler(providerSvc *provider.Service, chatSvc *chat.ChatService, quotaSvc *quota.Service) *AgentHandler {
	return &AgentHandler{
		service:     NewAgentService(),
		providerSvc: providerSvc,
		chatSvc:     chatSvc,
		quota:       quotaSvc,
		mcpRegistry: nil,
		skillSvc:   skill.NewService(database.DB),
	}
}

func (h *AgentHandler) SetMCPRegistry(reg *mcp.Registry) {
	h.mcpRegistry = reg
}

func (h *AgentHandler) ListAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	agents, err := h.service.ListAgents(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, agents)
}

func (h *AgentHandler) CreateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	agent, err := h.service.CreateAgent(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, agent)
}

func (h *AgentHandler) GetAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")
	agent, err := h.service.GetAgent(userID, agentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err.Error() == "Agent 不存在" {
			response.NotFound(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, agent)
}

func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")
	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	agent, err := h.service.UpdateAgent(userID, agentID, &req)
	if err != nil {
		if err.Error() == "Agent 不存在" {
			response.NotFound(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.Success(c, agent)
}

func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")
	err := h.service.DeleteAgent(userID, agentID)
	if err != nil {
		if err.Error() == "Agent 不存在" {
			response.NotFound(c, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}
	response.SuccessWithMessage(c, nil, "Agent 已删除")
}

func (h *AgentHandler) AgentChat(c *gin.Context) {
	agentID := c.Param("id")
	userID := c.GetString("user_id")

	var dbAgent struct {
		ID           string
		UserID       string
		Model        string
		SystemPrompt string
		SkillID      string
		MemoryType   string
		MemoryTurns  int
		Tools        agentmodel.JSONBArray
	}
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).
		Select("id, user_id, model, system_prompt, skill_id, memory_type, memory_turns, tools").
		First(&dbAgent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Agent 不存在")
		} else {
			response.InternalError(c, "查询 Agent 失败")
		}
		return
	}

	var req struct {
		Model       string             `json:"model"`
		Messages    []chat.ChatMessage `json:"messages" binding:"required"`
		Stream      bool               `json:"stream"`
		Temperature *float64           `json:"temperature,omitempty"`
		MaxTokens   *int              `json:"max_tokens,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}

	if dbAgent.MemoryType != "off" {
		turns := dbAgent.MemoryTurns
		if turns <= 0 {
			turns = 10
		}
		windowMsgs := toWindowMessages(req.Messages)
		windowMsgs = memory.ApplyWindow(windowMsgs, turns, 0)
		req.Messages = fromWindowMessages(windowMsgs)
	}

	model := req.Model
	if model == "" {
		model = dbAgent.Model
	}
	if model == "" {
		model = h.defaultModel()
	}

	// 阶段十五：优先用 Skill.Instructions 构建 system prompt
	systemPrompt := dbAgent.SystemPrompt
	if dbAgent.SkillID != "" {
		if s := agentmodel.GetBuiltInSkill(dbAgent.SkillID); s != nil {
			systemPrompt = s.BuildSystemPrompt()
		} else if h.skillSvc != nil {
			if dbSkill, err := h.skillSvc.GetSkill(userID, dbAgent.SkillID); err == nil {
				systemPrompt = dbSkill.BuildSystemPrompt()
			}
		}
	}
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant."
	}
	messages := append([]chat.ChatMessage{{Role: "system", Content: systemPrompt}}, req.Messages...)

	if h.quota != nil {
		if err := h.quota.Allow(c.Request.Context(), userID, "agent"); err != nil {
			quota.WriteError(c, err)
			return
		}
	}
	started := time.Now()

	svcReq := &chat.ChatRequest{
		Model:       model,
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	// 阶段十五 15.2：注入 MCP 工具
	var hasTools bool
	if h.mcpRegistry != nil && len(dbAgent.Tools) > 0 {
		for _, toolDef := range h.mcpRegistry.GetAllTools() {
			svcReq.Tools = append(svcReq.Tools, chat.ToolCallSpec{
				Type: "function",
				Function: chat.ToolFunctionSpec{
					Name:        toolDef.Name,
					Description: toolDef.Description,
					Parameters:  toolDef.InputSchema,
				},
			})
		}
		hasTools = len(svcReq.Tools) > 0
	}

	var resp any
	var err error

	if hasTools && !req.Stream {
		executor := buildToolExecutor(h.mcpRegistry)
		resp, err = h.chatSvc.ChatWithTools(svcReq, executor)
	} else {
		resp, err = h.chatSvc.Chat(svcReq)
	}

	if err != nil {
		if h.quota != nil {
			h.quota.RefundRequest(c.Request.Context(), userID)
		}
		if chat.WriteNoActiveProvider(c, err) {
			return
		}
		h.commitAgent(c, userID, model, nil, "error", started)
		response.Fail(c, http.StatusBadGateway, "AI provider error: "+err.Error())
		return
	}

	if chatResp, ok := resp.(*chat.ChatResponse); ok {
		h.commitAgent(c, userID, model, &chatResp.Usage, "ok", started)
	}
	response.Success(c, resp)
}

func (h *AgentHandler) commitAgent(c *gin.Context, userID, modelName string, usage *chat.ChatUsage, status string, started time.Time) {
	if h.quota == nil || userID == "" {
		return
	}
	in := quota.CommitInput{
		UserID:    userID,
		Source:    "agent",
		Model:     modelName,
		Status:    status,
		TraceID:   quota.TraceID(c),
		LatencyMS: time.Since(started).Milliseconds(),
	}
	if usage != nil {
		in.PromptTokens = usage.PromptTokens
		in.CompletionTokens = usage.CompletionTokens
		in.TotalTokens = usage.TotalTokens
		in.Estimated = usage.Estimated
	}
	h.quota.Commit(c.Request.Context(), in)
}

func (h *AgentHandler) defaultModel() string {
	if h.providerSvc == nil {
		return ""
	}
	active, err := h.providerSvc.GetActive()
	if err == nil && active.DefaultModel != "" {
		return active.DefaultModel
	}
	return ""
}

func toWindowMessages(in []chat.ChatMessage) []memory.WindowMessage {
	out := make([]memory.WindowMessage, len(in))
	for i, m := range in {
		out[i] = memory.WindowMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func fromWindowMessages(in []memory.WindowMessage) []chat.ChatMessage {
	out := make([]chat.ChatMessage, len(in))
	for i, m := range in {
		out[i] = chat.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func buildToolExecutor(registry *mcp.Registry) *mcpToolExecutor {
	return &mcpToolExecutor{registry: registry, executor: mcp.NewBuiltInExecutor()}
}

type mcpToolExecutor struct {
	registry  *mcp.Registry
	executor *mcp.BuiltInExecutor
}

func (e *mcpToolExecutor) Execute(call chat.ToolCall) (string, error) {
	toolDef := e.registry.GetBuiltInTool(call.Function.Name)
	if toolDef == nil {
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
	var args map[string]interface{}
	if call.Function.Arguments != "" {
		_ = json.Unmarshal([]byte(call.Function.Arguments), &args)
	}
	if toolDef.ServerType == "built-in" {
		result, err := e.executor.Execute(toolDef.Name, args)
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(result)
		return string(b), nil
	}
	srv := e.registry.GetServer(toolDef.ServerID)
	if srv == nil || srv.URL == "" {
		return "", fmt.Errorf("MCP server not available: %s", toolDef.ServerID)
	}
	result, err := e.executor.CallHTTPClient(srv.URL, srv.AuthHeader, toolDef.Name, args)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}

func (h *AgentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	agents := rg.Group("/agents")
	{
		agents.GET("", h.ListAgents)
		agents.POST("", h.CreateAgent)
		agents.GET("/:id", h.GetAgent)
		agents.PUT("/:id", h.UpdateAgent)
		agents.DELETE("/:id", h.DeleteAgent)
		agents.POST("/:id/chat", h.AgentChat)
	}
}
