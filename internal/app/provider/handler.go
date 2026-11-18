package provider

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/model"
	"cogniforge/internal/shared/response"
)

// Handler HTTP处理层
type Handler struct {
	service *Service
}

// NewHandler 创建 Handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) actor(c *gin.Context) Actor {
	return ResolveActor(c.GetString("user_id"))
}

func (h *Handler) writeErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	forbidden, notFound, badRequest := MapServiceError(err)
	switch {
	case forbidden:
		response.Forbidden(c, err.Error())
	case notFound:
		response.NotFound(c, "provider not found")
	case badRequest:
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, err.Error())
	}
	return true
}

// ListProviders 获取供应商列表（可见范围内）
func (h *Handler) ListProviders(c *gin.Context) {
	actor := h.actor(c)
	providers, err := h.service.ListFor(actor)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	prefChat, prefEmbed := h.service.PreferredIDs(actor.UserID)
	safeList := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		resp := toResponseFor(&p, actor, prefChat, prefEmbed)
		resp.APIKey = maskKey(resp.APIKey)
		resp.ExtraHeaders = nil
		safeList[i] = resp
	}
	response.Success(c, safeList)
}

// GetProvider 获取单个供应商
func (h *Handler) GetProvider(c *gin.Context) {
	actor := h.actor(c)
	id := c.Param("id")
	p, err := h.service.GetFor(id, actor)
	if h.writeErr(c, err) {
		return
	}
	prefChat, prefEmbed := h.service.PreferredIDs(actor.UserID)
	resp := toResponseFor(p, actor, prefChat, prefEmbed)
	resp.APIKey = maskKey(resp.APIKey)
	response.Success(c, resp)
}

// CreateProvider 创建供应商
func (h *Handler) CreateProvider(c *gin.Context) {
	actor := h.actor(c)
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.service.CreateFor(&req, actor)
	if h.writeErr(c, err) {
		return
	}
	resp := toResponseFor(p, actor, "", "")
	resp.APIKey = maskKey(resp.APIKey)
	response.Created(c, resp)
}

// UpdateProvider 更新供应商
func (h *Handler) UpdateProvider(c *gin.Context) {
	actor := h.actor(c)
	id := c.Param("id")
	req := new(UpdateProviderRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.service.UpdateFor(id, req, actor)
	if h.writeErr(c, err) {
		return
	}
	prefChat, prefEmbed := h.service.PreferredIDs(actor.UserID)
	resp := toResponseFor(p, actor, prefChat, prefEmbed)
	resp.APIKey = maskKey(resp.APIKey)
	response.Success(c, resp)
}

// DeleteProvider 删除供应商
func (h *Handler) DeleteProvider(c *gin.Context) {
	actor := h.actor(c)
	id := c.Param("id")
	if err := h.service.DeleteFor(id, actor); h.writeErr(c, err) {
		return
	}
	response.SuccessWithMessage(c, nil, "deleted")
}

// SetDefault 设置默认供应商（admin 全站；普通用户个人偏好）
func (h *Handler) SetDefault(c *gin.Context) {
	actor := h.actor(c)
	id := c.Param("id")
	var req SetDefaultRequest
	_ = c.ShouldBindJSON(&req)
	personal := false
	if req.Personal != nil {
		personal = *req.Personal
	} else if !actor.IsAdmin {
		personal = true
	}
	if err := h.service.SetDefaultFor(id, req.Purpose, actor, personal); h.writeErr(c, err) {
		return
	}
	msg := "set as default"
	if personal || !actor.IsAdmin {
		msg = "set as personal default"
	}
	response.SuccessWithMessage(c, nil, msg)
}

// TestProvider 测试供应商连接
func (h *Handler) TestProvider(c *gin.Context) {
	actor := h.actor(c)
	id := c.Param("id")
	result, err := h.service.TestConnectionFor(id, actor)
	if err != nil {
		if h.writeErr(c, err) {
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}

// GetActive 获取当前生效配置（按用户偏好解析）
func (h *Handler) GetActive(c *gin.Context) {
	actor := h.actor(c)
	p, err := h.service.GetActiveForUser(actor.UserID)
	if err != nil {
		p, err = h.service.GetActive()
		if err != nil {
			response.Fail(c, http.StatusServiceUnavailable, "no active provider")
			return
		}
	}
	if !model.ProviderVisibleTo(p, actor.UserID, actor.IsAdmin) {
		response.Fail(c, http.StatusServiceUnavailable, "no active provider")
		return
	}
	prefChat, prefEmbed := h.service.PreferredIDs(actor.UserID)
	resp := toResponseFor(p, actor, prefChat, prefEmbed)
	resp.APIKey = maskKey(resp.APIKey)
	response.Success(c, resp)
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/providers", h.ListProviders)
	rg.GET("/providers/active", h.GetActive)
	rg.GET("/providers/:id", h.GetProvider)
	rg.POST("/providers", h.CreateProvider)
	rg.PUT("/providers/:id", h.UpdateProvider)
	rg.DELETE("/providers/:id", h.DeleteProvider)
	rg.POST("/providers/:id/default", h.SetDefault)
	rg.POST("/providers/:id/test", h.TestProvider)
}

// maskKey 将 API Key 中间部分替换为 ***
func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}
