package token

import (
	"github.com/gin-gonic/gin"
)

// Handler HTTP 路由封装
type Handler struct {
	svc *Service
}

// NewHandler 创建 handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(authenticated *gin.RouterGroup) {
	t := authenticated.Group("/token")
	{
		// 获取 LLM 临时凭证（供 Python 直调用）
		t.GET("/llm", h.svc.IssueHandler())
		// 验证 token（Go 侧验证，Python 回调用）
		t.POST("/validate", h.svc.ValidateHandler())
	}
}
