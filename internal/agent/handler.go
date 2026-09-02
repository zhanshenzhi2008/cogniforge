package agent

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/chat"
	"cogniforge/internal/database"
	"cogniforge/internal/memory"
	"cogniforge/internal/provider"
	"cogniforge/internal/quota"
	"cogniforge/internal/response"
)

type AgentHandler struct {
	service     *AgentService
	providerSvc *provider.Service
	chatSvc     *chat.ChatService
	quota       *quota.Service
}

func NewAgentHandler(providerSvc *provider.Service, chatSvc *chat.ChatService, quotaSvc *quota.Service) *AgentHandler {
	return &AgentHandler{
		service:     NewAgentService(),
		providerSvc: providerSvc,
		chatSvc:     chatSvc,
		quota:       quotaSvc,
	}
}

// ListAgents 获取 Agent 列表
func (h *AgentHandler) ListAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	agents, err := h.service.ListAgents(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, agents)
}

// CreateAgent 创建 Agent
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

// GetAgent 获取 Agent 详情
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

// UpdateAgent 更新 Agent
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

// DeleteAgent 删除 Agent
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

// AgentChat Agent 对话
func (h *AgentHandler) AgentChat(c *gin.Context) {
	agentID := c.Param("id")
	userID := c.GetString("user_id")

	var dbAgent struct {
		ID           string
		UserID       string
		Model        string
		SystemPrompt string
		MemoryType   string
		MemoryTurns  int
	}
	if err := database.DB.Where("id = ? AND user_id = ?", agentID, userID).
		Select("id, user_id, model, system_prompt, memory_type, memory_turns").
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

	// 服务端滑动窗口裁剪（阶段十四 14.1）
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

	systemPrompt := dbAgent.SystemPrompt
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

	svcReq := toServiceRequest(&ChatRequest{
		Model:       model,
		Messages:    messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})

	if req.Stream {
		usage, err := h.chatSvc.ChatStream(c, svcReq)
		if err != nil {
			if h.quota != nil {
				h.quota.RefundRequest(c.Request.Context(), userID)
			}
			if chat.WriteNoActiveProvider(c, err) {
				return
			}
			h.commitAgent(c, userID, model, nil, "error", started)
			slog.Error("AgentChat stream failed",
				"error", err,
				"agent_id", agentID,
				"model", model,
			)
			fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
			c.Writer.Flush()
			return
		}
		h.commitAgent(c, userID, model, usage, "ok", started)
	} else {
		resp, err := h.chatSvc.Chat(svcReq)
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
		u := resp.Usage
		h.commitAgent(c, userID, model, &u, "ok", started)
		response.Success(c, resp)
	}
}

func (h *AgentHandler) commitAgent(c *gin.Context, userID, model string, usage *chat.ChatUsage, status string, started time.Time) {
	if h.quota == nil || userID == "" {
		return
	}
	in := quota.CommitInput{
		UserID:    userID,
		Source:    "agent",
		Model:     model,
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

type ChatRequest struct {
	Model       string             `json:"model"`
	Messages    []chat.ChatMessage `json:"messages"`
	Stream      bool               `json:"stream"`
	Temperature *float64           `json:"temperature,omitempty"`
	MaxTokens   *int              `json:"max_tokens,omitempty"`
}

func toServiceRequest(req *ChatRequest) *chat.ChatRequest {
	return &chat.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
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

// RegisterRoutes 注册路由
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
