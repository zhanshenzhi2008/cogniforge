package chat

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/config"
	"cogniforge/internal/response"
)

type ChatHandler struct {
	service *ChatService
}

func NewChatHandler(cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		service: NewChatService(cfg),
	}
}

// ListModels 获取模型列表
func (h *ChatHandler) ListModels(c *gin.Context) {
	result := h.service.ListModels()
	response.Success(c, result)
}

// GetModel 获取模型信息
func (h *ChatHandler) GetModel(c *gin.Context) {
	response.Success(c, gin.H{"message": "Get model"})
}

// Chat 对话
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

	resp, err := h.service.Chat(&req)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, "AI provider error: "+err.Error())
		return
	}
	response.Success(c, resp)
}

// ChatStream 流式对话
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

	if err := h.service.ChatStream(c, &req); err != nil {
		slog.Error("ChatStream failed",
			"error", err,
			"model", req.Model,
			"messages_count", len(req.Messages),
		)
		fmt.Fprintf(c.Writer, "data: {\"error\": \"AI provider error: %s\"}\n\n", err.Error())
		c.Writer.Flush()
	}
}

// RegisterRoutes 注册路由
func (h *ChatHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/models", h.ListModels)
	rg.GET("/models/:id", h.GetModel)
	rg.POST("/chat/completions", h.Chat)
	rg.POST("/chat/stream", h.ChatStream)
}
