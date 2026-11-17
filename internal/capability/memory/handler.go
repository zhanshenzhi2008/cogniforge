package memory

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/response"
)

type MemoryHandler struct {
	svc *MemoryService
}

func NewMemoryHandler(db *gorm.DB) *MemoryHandler {
	return &MemoryHandler{svc: NewMemoryService(db)}
}

func (h *MemoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/memories")
	g.Use(func(c *gin.Context) {
		c.Next()
	})
	g.GET("", h.ListMemories)
	g.GET("/:id", h.GetMemory)
	g.DELETE("/:id", h.DeleteMemory)
	g.DELETE("", h.ClearAllMemories)
}

// ListMemories GET /api/v1/memories?kind=&agent_id=&limit=
func (h *MemoryHandler) ListMemories(c *gin.Context) {
	userID := c.GetString("user_id")
	kind := c.Query("kind")
	agentID := c.Query("agent_id")
	limit := 50

	rows, err := h.svc.List(c.Request.Context(), userID, kind, agentID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": rows, "total": len(rows)})
}

// GetMemory GET /api/v1/memories/:id
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	row, err := h.svc.Get(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "记忆不存在")
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, row)
}

// DeleteMemory DELETE /api/v1/memories/:id
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.svc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "记忆已删除")
}

// ClearAllMemories DELETE /api/v1/memories
func (h *MemoryHandler) ClearAllMemories(c *gin.Context) {
	userID := c.GetString("user_id")
	if err := h.svc.ClearAll(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "全部记忆已清空")
}
