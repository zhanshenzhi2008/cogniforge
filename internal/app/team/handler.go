package team

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/capability/mcp"
	"cogniforge/internal/capability/skill"
	"cogniforge/internal/governance/quota"
	"cogniforge/internal/infra/database"
	orchteam "cogniforge/internal/orchestration/team"
	rtchat "cogniforge/internal/runtime/chat"
	"cogniforge/internal/shared/response"
)

// Handler 小队 HTTP
type Handler struct {
	svc   *orchteam.Service
	chat  *rtchat.ChatService
	mcp   *mcp.Registry
	quota *quota.Service
	skill *skill.Service
}

func NewHandler(chatSvc *rtchat.ChatService, quotaSvc *quota.Service) *Handler {
	return &Handler{
		svc:   orchteam.NewService(),
		chat:  chatSvc,
		quota: quotaSvc,
		skill: skill.NewService(database.DB),
	}
}

func NewHandlerWithService(svc *orchteam.Service, chatSvc *rtchat.ChatService, quotaSvc *quota.Service) *Handler {
	return &Handler{svc: svc, chat: chatSvc, quota: quotaSvc, skill: skill.NewService(database.DB)}
}

func (h *Handler) SetMCPRegistry(reg *mcp.Registry) {
	h.mcp = reg
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	rows, err := h.svc.List(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rows)
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	var req orchteam.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	row, err := h.svc.Create(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, row)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	row, err := h.svc.Get(userID, id)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") || err == gorm.ErrRecordNotFound {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, row)
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	var req orchteam.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	row, err := h.svc.Update(userID, id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, row)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	if err := h.svc.Delete(userID, id); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// Chat 小队预览对话（主管/入口 Agent + delegate/handoff）
func (h *Handler) Chat(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.Param("id")
	t, err := h.svc.Get(userID, teamID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	if t.Status != "" && t.Status != "active" {
		response.BadRequest(c, "小队未启用")
		return
	}

	var req struct {
		Messages []rtchat.ChatMessage `json:"messages" binding:"required"`
		Stream   bool                 `json:"stream"`
		Goal     string               `json:"goal"`
		RunState *orchteam.RunState   `json:"run_state"` // 客户端回传，handoff 后换人设续聊
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}

	if h.quota != nil {
		if err := h.quota.Allow(c.Request.Context(), userID, "team"); err != nil {
			quota.WriteError(c, err)
			return
		}
	}
	started := time.Now()

	deps := &orchteam.ChatDeps{
		ChatSvc: h.chat,
		MCP:     h.mcp,
		DB:      database.DB,
		Skill:   h.skill,
	}
	result, err := deps.RunChat(c.Request.Context(), userID, t, req.Messages, req.Goal, req.RunState)
	if err != nil {
		if h.quota != nil {
			h.quota.RefundRequest(c.Request.Context(), userID)
		}
		if rtchat.WriteNoActiveProvider(c, err) {
			return
		}
		h.commitTeam(c, userID, "", nil, "error", started)
		response.FailWithHTTPStatus(c, http.StatusBadGateway, response.CodeAIProviderError, "AI provider error: "+err.Error())
		return
	}
	modelName := ""
	if result.Response != nil {
		modelName = result.Response.Model
		h.commitTeam(c, userID, modelName, &result.Response.Usage, "ok", started)
	} else {
		h.commitTeam(c, userID, modelName, nil, "ok", started)
	}

	payload := gin.H{
		"response":  result.Response,
		"run_state": result.RunState,
	}
	if req.Stream && result.Response != nil {
		rtchat.WriteResponseSSE(c, result.Response)
		return
	}
	response.Success(c, payload)
}

func (h *Handler) commitTeam(c *gin.Context, userID, modelName string, usage *rtchat.ChatUsage, status string, started time.Time) {
	if h.quota == nil || userID == "" {
		return
	}
	in := quota.CommitInput{
		UserID:    userID,
		Source:    "team",
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

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	teams := rg.Group("/teams")
	{
		teams.GET("", h.List)
		teams.POST("", h.Create)
		teams.GET("/:id", h.Get)
		teams.PUT("/:id", h.Update)
		teams.DELETE("/:id", h.Delete)
		teams.POST("/:id/chat", h.Chat)
	}
}
