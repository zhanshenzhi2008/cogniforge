package chat

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/provider"
	"cogniforge/internal/quota"
	"cogniforge/internal/response"
)

type ChatHandler struct {
	service *ChatService
	conv    *ConversationService
	quota   *quota.Service
}

func NewChatHandler(providerSvc *provider.Service, db *gorm.DB, quotaSvc *quota.Service) *ChatHandler {
	return &ChatHandler{
		service: NewChatService(providerSvc),
		conv:    NewConversationService(db),
		quota:   quotaSvc,
	}
}

func (h *ChatHandler) Service() *ChatService {
	return h.service
}

func (h *ChatHandler) ListModels(c *gin.Context) {
	result := h.service.ListModels()
	response.Success(c, result)
}

func (h *ChatHandler) GetModel(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get model"})
}

func (h *ChatHandler) gate(c *gin.Context, source string) bool {
	if h.quota == nil {
		return true
	}
	userID := c.GetString("user_id")
	if userID == "" {
		return true
	}
	if err := h.quota.Allow(c.Request.Context(), userID, source); err != nil {
		quota.WriteError(c, err)
		return false
	}
	return true
}

func (h *ChatHandler) commit(c *gin.Context, source, model string, usage *ChatUsage, status string, started time.Time) {
	if h.quota == nil {
		return
	}
	userID := c.GetString("user_id")
	if userID == "" {
		return
	}
	in := quota.CommitInput{
		UserID:    userID,
		Source:    source,
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

func (h *ChatHandler) refund(c *gin.Context) {
	if h.quota == nil {
		return
	}
	h.quota.RefundRequest(c.Request.Context(), c.GetString("user_id"))
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}
	if !h.gate(c, "playground") {
		return
	}
	started := time.Now()
	resp, err := h.service.Chat(&req)
	if err != nil {
		h.refund(c)
		if WriteNoActiveProvider(c, err) {
			return
		}
		h.commit(c, "playground", req.Model, nil, "error", started)
		response.Fail(c, http.StatusBadGateway, "AI provider error: "+err.Error())
		return
	}
	usage := resp.Usage
	h.commit(c, "playground", req.Model, &usage, "ok", started)
	response.Success(c, resp)
}

func (h *ChatHandler) ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		response.BadRequest(c, "messages 不能为空")
		return
	}
	if !h.gate(c, "playground") {
		return
	}
	started := time.Now()
	usage, err := h.service.ChatStream(c, &req)
	if err != nil {
		h.refund(c)
		if WriteNoActiveProvider(c, err) {
			return
		}
		h.commit(c, "playground", req.Model, nil, "error", started)
		slog.Error("ChatStream failed",
			"error", err,
			"model", req.Model,
			"messages_count", len(req.Messages),
		)
		fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
		c.Writer.Flush()
		return
	}
	h.commit(c, "playground", req.Model, usage, "ok", started)
}

func WriteNoActiveProvider(c *gin.Context, err error) bool {
	if err == nil || !errors.Is(err, ErrNoActiveProvider) || c.Writer.Written() {
		return false
	}
	response.FailWithHTTPStatus(c, http.StatusServiceUnavailable, response.CodeNoActiveProvider, err.Error())
	return true
}

func (h *ChatHandler) Embeddings(c *gin.Context) {
	var req EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Input == nil {
		response.BadRequest(c, "input 不能为空")
		return
	}

	resp, err := h.service.Embeddings(&req)
	if err != nil {
		if WriteNoActiveProvider(c, err) {
			return
		}
		slog.Error("Embeddings failed", "error", err, "model", req.Model)
		response.FailWithHTTPStatus(c, http.StatusBadGateway, response.CodeAIProviderError, "AI provider error: "+err.Error())
		return
	}
	response.Success(c, resp)
}

func (h *ChatHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/models", h.ListModels)
	rg.GET("/models/:id", h.GetModel)
	rg.POST("/embeddings", h.Embeddings)
}

func (h *ChatHandler) RegisterRoutes(rg *gin.RouterGroup) {
	h.RegisterPublicRoutes(rg)
}
