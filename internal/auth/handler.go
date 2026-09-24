package auth

import (
	"errors"
	"net/http"
	"strings"

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

func NewAuthHandlerWithService(svc *AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
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

// PasswordResetOptions 查询是否已开通邮件重置
func (h *AuthHandler) PasswordResetOptions(c *gin.Context) {
	response.Success(c, h.service.PasswordResetOptions())
}

// ForgotPassword 发送重置邮件
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的邮箱地址")
		return
	}

	// 动态获取当前请求的 origin，避免依赖静态 APP_PUBLIC_URL
	publicURL := getRequestOrigin(c)

	err := h.service.RequestPasswordReset(c.Request.Context(), req.Email, publicURL)
	if err != nil {
		if errors.Is(err, errMailDisabled) {
			response.FailWithHTTPStatus(c, http.StatusServiceUnavailable, response.CodeServiceUnavailable, "邮件重置尚未配置，请联系管理员在「用户」页重置密码")
			return
		}
		if strings.Contains(err.Error(), "有效的邮箱") {
			response.BadRequest(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "过于频繁") {
			response.FailWithHTTPStatus(c, http.StatusTooManyRequests, response.CodeRateLimitExceeded, err.Error())
			return
		}
		if strings.Contains(err.Error(), "邮件发送失败") || strings.Contains(err.Error(), "暂时不可用") {
			response.FailWithHTTPStatus(c, http.StatusBadGateway, response.CodeNetworkError, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "如果该邮箱已注册，你将收到一封重置邮件（请同时检查垃圾箱）")
}

// getRequestOrigin 从请求 Header 动态获取 origin
// 优先级：X-Forwarded-Host > Host > Origin
func getRequestOrigin(c *gin.Context) string {
	// 反向代理场景
	if host := c.GetHeader("X-Forwarded-Host"); host != "" {
		scheme := c.GetHeader("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "https" // 默认 https
		}
		return scheme + "://" + host
	}

	// 直接请求场景
	if origin := c.GetHeader("Origin"); origin != "" {
		return origin
	}

	// Fallback: 使用 Host
	if host := c.GetHeader("Host"); host != "" {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		return scheme + "://" + host
	}

	// 兜底：使用配置的默认值
	return ""
}

// ResetPassword 用邮件令牌设新密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的重置信息")
		return
	}
	if err := h.service.ResetPasswordWithToken(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "无效或已过期") {
			response.FailWithHTTPStatus(c, http.StatusGone, response.CodeVerifyCodeInvalid, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "密码已重置，请使用新密码登录")
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetCurrentUser)
		auth.GET("/password-reset-options", h.PasswordResetOptions)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		auth.GET("/apikeys", h.ListApiKeys)
		auth.POST("/apikeys", h.CreateApiKey)
		auth.DELETE("/apikeys/:id", h.DeleteApiKey)
	}
}
