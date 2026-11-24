package team

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"cogniforge/internal/capability/mcp"
	"cogniforge/internal/capability/skill"
	"cogniforge/internal/governance/quota"
	"cogniforge/internal/infra/database"
	"cogniforge/internal/model"
	orchteam "cogniforge/internal/orchestration/team"
	rtchat "cogniforge/internal/runtime/chat"
	"cogniforge/internal/shared/response"
)

// Handler 小队 HTTP
type Handler struct {
	svc     *orchteam.Service
	chat    *rtchat.ChatService
	conv    *rtchat.ConversationService
	runs    *orchteam.RunStore
	mcp     *mcp.Registry
	quota   *quota.Service
	skill   *skill.Service
}

func NewHandler(chatSvc *rtchat.ChatService, quotaSvc *quota.Service) *Handler {
	return &Handler{
		svc:   orchteam.NewService(),
		chat:  chatSvc,
		conv:  rtchat.NewConversationService(database.DB),
		quota: quotaSvc,
		skill: skill.NewService(database.DB),
	}
}

func NewHandlerWithService(svc *orchteam.Service, chatSvc *rtchat.ChatService, quotaSvc *quota.Service) *Handler {
	return &Handler{
		svc:   svc,
		chat:  chatSvc,
		conv:  rtchat.NewConversationService(database.DB),
		quota: quotaSvc,
		skill: skill.NewService(database.DB),
	}
}

func (h *Handler) SetMCPRegistry(reg *mcp.Registry) {
	h.mcp = reg
}

func (h *Handler) SetRedis(rdb *redis.Client) {
	h.runs = orchteam.NewRunStore(rdb)
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
		Messages       []rtchat.ChatMessage `json:"messages" binding:"required"`
		Stream         bool                 `json:"stream"`
		Goal           string               `json:"goal"`
		RunState       *orchteam.RunState   `json:"run_state"`
		ConversationID string               `json:"conversation_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}

	prior := req.RunState
	sessionKey := strings.TrimSpace(req.ConversationID)
	if prior == nil && h.runs != nil && sessionKey != "" {
		if loaded, loadErr := h.runs.Load(c.Request.Context(), userID, teamID, sessionKey); loadErr == nil && loaded != nil {
			prior = loaded
		}
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
	result, err := deps.RunChat(c.Request.Context(), userID, t, req.Messages, req.Goal, prior)
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

	convID, persistErr := h.persistPreview(userID, teamID, sessionKey, modelName, req.Messages, result)
	if persistErr != nil {
		// 预览主路径已成功；落库失败不回滚对话
		_ = persistErr
	}
	if h.runs != nil && result.RunState != nil {
		key := convID
		if key == "" {
			key = sessionKey
		}
		if key != "" {
			_ = h.runs.Save(c.Request.Context(), userID, teamID, key, result.RunState)
		}
	}

	payload := gin.H{
		"response":        result.Response,
		"run_state":       result.RunState,
		"conversation_id": convID,
	}
	if req.Stream && result.Response != nil {
		rtchat.WriteResponseSSE(c, result.Response)
		return
	}
	response.Success(c, payload)
}

// Run 任务型小队执行（不落 chat 预览；正式可恢复任务仍建议走工作流 team 节点）
func (h *Handler) Run(c *gin.Context) {
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
		Task     string             `json:"task"`
		Goal     string             `json:"goal"`
		RunState *orchteam.RunState `json:"run_state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	task := strings.TrimSpace(req.Task)
	if task == "" {
		task = strings.TrimSpace(req.Goal)
	}
	if task == "" {
		response.BadRequest(c, "task 不能为空")
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
	result, err := deps.RunTask(c.Request.Context(), userID, t, task, req.Goal, req.RunState)
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

	response.Success(c, gin.H{
		"response":  result.Response,
		"run_state": result.RunState,
	})
}

func (h *Handler) persistPreview(
	userID, teamID, conversationID, modelName string,
	reqMsgs []rtchat.ChatMessage,
	result *orchteam.ChatResult,
) (string, error) {
	if h.conv == nil || result == nil {
		return conversationID, nil
	}
	active := ""
	if result.RunState != nil {
		active = result.RunState.ActiveAgentID
	}
	assistantText := ""
	if result.Response != nil {
		assistantText = extractChatText(result.Response)
	}

	var stored []model.ConversationMessage
	for _, m := range reqMsgs {
		if m.Role == "system" {
			continue
		}
		stored = append(stored, model.ConversationMessage{
			ID:            uuid.New().String(),
			Role:          m.Role,
			Content:       contentToString(m.Content),
			ActiveAgentID: active,
			Time:          time.Now().Format(time.RFC3339),
		})
	}
	if assistantText != "" {
		stored = append(stored, model.ConversationMessage{
			ID:            uuid.New().String(),
			Role:          "assistant",
			Content:       assistantText,
			ActiveAgentID: active,
			Time:          time.Now().Format(time.RFC3339),
		})
	}

	if conversationID != "" {
		row, err := h.conv.Get(userID, conversationID)
		if err != nil {
			return "", err
		}
		if row.TeamID != "" && row.TeamID != teamID {
			return "", nil
		}
		msgs := append([]model.ConversationMessage{}, row.Messages...)
		// 请求里通常已含完整历史；以请求侧为准覆盖，避免重复叠加
		if len(reqMsgs) > 0 {
			msgs = stored
		} else {
			msgs = append(msgs, stored...)
		}
		updated, err := h.conv.Update(userID, conversationID, &rtchat.UpdateConversationRequest{
			Messages: &msgs,
			Model:    &modelName,
			TeamID:   &teamID,
		})
		if err != nil {
			return conversationID, err
		}
		return updated.ID, nil
	}

	created, err := h.conv.Create(userID, &rtchat.CreateConversationRequest{
		TeamID:   teamID,
		Model:    modelName,
		Messages: stored,
		Title:    "",
	})
	if err != nil {
		return "", err
	}
	return created.ID, nil
}

func contentToString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		return ""
	}
}

func extractChatText(resp *rtchat.ChatResponse) string {
	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}
	return contentToString(resp.Choices[0].Message.Content)
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
		teams.POST("/:id/run", h.Run)
		teams.POST("/:id/expand-pipeline", h.ExpandPipeline)
	}
}

// ExpandPipeline 一键从小队生成工作流图并绑定 workflow_id
func (h *Handler) ExpandPipeline(c *gin.Context) {
	userID := c.GetString("user_id")
	teamID := c.Param("id")
	result, err := h.svc.ExpandToWorkflow(userID, teamID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}
