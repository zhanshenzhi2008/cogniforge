package chat

import (
	"errors"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	rows, err := h.conv.List(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, rows)
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	row, err := h.conv.Create(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, row)
}

func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	row, err := h.conv.Get(userID, c.Param("id"))
	if err != nil {
		if errors.Is(err, errConversationNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, row)
}

func (h *ChatHandler) UpdateConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	row, err := h.conv.Update(userID, c.Param("id"), &req)
	if err != nil {
		if errors.Is(err, errConversationNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, row)
}

func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.conv.Delete(userID, c.Param("id")); err != nil {
		if errors.Is(err, errConversationNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "对话已删除")
}

func (h *ChatHandler) RegisterConversationRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/conversations")
	g.GET("", h.ListConversations)
	g.POST("", h.CreateConversation)
	g.GET("/:id", h.GetConversation)
	g.PUT("/:id", h.UpdateConversation)
	g.DELETE("/:id", h.DeleteConversation)
}
