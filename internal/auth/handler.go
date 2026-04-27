package auth

import (
	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		service: NewAuthService(),
	}
}

func (h *AuthHandler) InitDefaultAdmin() {
	h.service.InitDefaultAdmin()
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	authData, err := h.service.Register(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, authData)
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	authData, err := h.service.Login(c)
	if err != nil {
		if err.Error() == "用户名或密码错误" {
			response.Unauthorized(c, err.Error())
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.Success(c, authData)
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	response.SuccessWithMessage(c, nil, "已退出登录")
}

// GetCurrentUser 获取当前用户
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.service.GetCurrentUser(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, user)
}

// ListApiKeys 获取 API Key 列表
func (h *AuthHandler) ListApiKeys(c *gin.Context) {
	userID := c.GetString("user_id")
	keys, err := h.service.ListApiKeys(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"keys": keys})
}

// CreateApiKey 创建 API Key
func (h *AuthHandler) CreateApiKey(c *gin.Context) {
	userID := c.GetString("user_id")
	var req ApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	apiKey, err := h.service.CreateApiKey(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, gin.H{
		"id":         apiKey.ID,
		"name":       apiKey.Name,
		"key":        apiKey.Key,
		"created_at": apiKey.CreatedAt,
	})
}

// DeleteApiKey 删除 API Key
func (h *AuthHandler) DeleteApiKey(c *gin.Context) {
	userID := c.GetString("user_id")
	keyID := c.Param("id")

	err := h.service.DeleteApiKey(userID, keyID)
	if err != nil {
		switch err.Error() {
		case "API Key 不存在":
			response.NotFound(c, err.Error())
		case "无权删除此 API Key":
			response.Forbidden(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}
	response.SuccessWithMessage(c, nil, "API Key 已撤销")
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetCurrentUser)
		auth.GET("/apikeys", h.ListApiKeys)
		auth.POST("/apikeys", h.CreateApiKey)
		auth.DELETE("/apikeys/:id", h.DeleteApiKey)
	}
}
