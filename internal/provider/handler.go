package provider

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

// Handler HTTP处理层
type Handler struct {
	service *Service
}

// NewHandler 创建 Handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListProviders 获取供应商列表
func (h *Handler) ListProviders(c *gin.Context) {
	providers, err := h.service.List()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	safeList := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		resp := toResponse(&p)
		resp.APIKey = maskKey(resp.APIKey)
		resp.ExtraHeaders = nil
		safeList[i] = resp
	}
	response.Success(c, safeList)
}

// GetProvider 获取单个供应商
func (h *Handler) GetProvider(c *gin.Context) {
	id := c.Param("id")
	p, err := h.service.Get(id)
	if err != nil {
		response.NotFound(c, "provider not found")
		return
	}
	resp := toResponse(p)
	response.Success(c, resp)
}

// CreateProvider 创建供应商
func (h *Handler) CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.service.Create(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, toResponse(p))
}

// UpdateProvider 更新供应商
func (h *Handler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	req := new(UpdateProviderRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.service.Update(id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, toResponse(p))
}

// DeleteProvider 删除供应商
func (h *Handler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "deleted")
}

// SetDefault 设置默认供应商
func (h *Handler) SetDefault(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.SetDefault(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "set as default")
}

// TestProvider 测试供应商连接
func (h *Handler) TestProvider(c *gin.Context) {
	id := c.Param("id")
	result, err := h.service.TestConnection(id)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}

// GetActive 获取当前生效配置
func (h *Handler) GetActive(c *gin.Context) {
	p, err := h.service.GetActive()
	if err != nil {
		response.Fail(c, http.StatusServiceUnavailable, "no active provider")
		return
	}
	resp := toResponse(p)
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
