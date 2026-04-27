package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		service: NewUserService(),
	}
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.ListUsers(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.service.CreateUser(&req)
	if err != nil {
		if err.Error() == "该邮箱已被注册" {
			response.Fail(c, http.StatusConflict, err.Error())
		} else {
			response.InternalError(c, err.Error())
		}
		return
	}

	response.Created(c, user)
}

// GetUser 获取单个用户
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.service.UpdateUser(userID, &req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	err := h.service.DeleteUser(userID)
	if err != nil {
		if err.Error() == "不能删除默认管理员" {
			response.Fail(c, http.StatusForbidden, err.Error())
		} else {
			response.NotFound(c, err.Error())
		}
		return
	}

	response.SuccessWithMessage(c, nil, "用户已删除")
}

// UpdateUserStatus 更新用户状态
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.service.UpdateUserStatus(userID, &req)
	if err != nil {
		if err.Error() == "不能修改默认管理员状态" {
			response.Fail(c, http.StatusForbidden, err.Error())
		} else {
			response.NotFound(c, err.Error())
		}
		return
	}

	response.SuccessWithMessage(c, nil, "状态已更新")
}

// UpdateCurrentUser 更新当前用户
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.service.UpdateCurrentUser(userID.(string), &req)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, user)
}

// ============ Settings 相关 ============

// GetSettings 获取当前用户设置
func (h *UserHandler) GetSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	settings, err := h.service.GetSettings(userID.(string))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToSettingsResponse(settings))
}

// UpdateSettings 更新当前用户设置
func (h *UserHandler) UpdateSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	settings, err := h.service.UpdateSettings(userID.(string), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToSettingsResponse(settings))
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.service.ChangePassword(userID.(string), &req)
	if err != nil {
		switch err.Error() {
		case "旧密码错误":
			response.Fail(c, http.StatusUnauthorized, err.Error())
		case "新旧密码不能相同":
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.SuccessWithMessage(c, nil, "密码修改成功")
}

// UploadAvatar 上传头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "请选择头像文件")
		return
	}

	if file.Size > 2*1024*1024 {
		response.BadRequest(c, "文件大小不能超过2MB")
		return
	}

	filename := generateID() + ".jpg"
	_, err = h.service.UploadAvatar(userID.(string), filename)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, AvatarUploadResponse{
		AvatarURL: "/uploads/avatars/" + filename,
		Message:   "头像上传成功",
	})
}

// GetSessions 获取会话列表
func (h *UserHandler) GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	sessions, err := h.service.GetSessions(userID.(string))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, sessions)
}

// RevokeSession 撤销会话
func (h *UserHandler) RevokeSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		response.BadRequest(c, "会话ID不能为空")
		return
	}

	err := h.service.RevokeSession(userID.(string), sessionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, nil, "会话已撤销")
}

// RegisterRoutes 注册路由
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", h.GetUsers)
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.PATCH("/:id/status", h.UpdateUserStatus)
	}

	// 当前用户
	me := rg.Group("/me")
	{
		me.PUT("", h.UpdateCurrentUser)
	}

	// 设置
	settings := rg.Group("/settings")
	{
		settings.GET("", h.GetSettings)
		settings.PUT("", h.UpdateSettings)
		settings.POST("/password", h.ChangePassword)
		settings.POST("/avatar", h.UploadAvatar)
		settings.GET("/sessions", h.GetSessions)
		settings.DELETE("/sessions/:id", h.RevokeSession)
	}
}
