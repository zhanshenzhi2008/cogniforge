package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// =============================================================================
// 个人设置 API
// =============================================================================

// GetSettings 获取当前用户的设置
// GET /api/settings
func GetSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	var settings model.UserSettings
	if err := database.DB.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 首次访问，创建默认设置
			settings = model.UserSettings{
				ID:        generateID(),
				UserID:    userID.(string),
				AvatarURL: "",
				Theme:     "light",
				Language:  "zh-CN",
				Timezone:  "Asia/Shanghai",
				Metadata:  make(model.JSONBMap),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			database.DB.Create(&settings)
		} else {
			model.InternalError(c, "查询失败")
			return
		}
	}

	model.Success(c, settings)
}

// UpdateSettings 更新用户设置（基本信息）
// PUT /api/settings
func UpdateSettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	var req struct {
		AvatarURL string                 `json:"avatar_url"`
		Theme     string                 `json:"theme" binding:"omitempty,oneof=light dark auto"`
		Language  string                 `json:"language"`
		Timezone  string                 `json:"timezone"`
		Metadata  map[string]interface{} `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 查找现有设置
	var settings model.UserSettings
	if err := database.DB.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新设置
			settings = model.UserSettings{
				ID:        generateID(),
				UserID:    userID.(string),
				AvatarURL: req.AvatarURL,
				Theme:     req.Theme,
				Language:  req.Language,
				Timezone:  req.Timezone,
				Metadata:  req.Metadata,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := database.DB.Create(&settings).Error; err != nil {
				model.InternalError(c, "创建设置失败")
				return
			}
			model.Success(c, settings)
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 更新设置
	updates := make(map[string]interface{})
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}
	if req.Theme != "" {
		updates["theme"] = req.Theme
	}
	if req.Language != "" {
		updates["language"] = req.Language
	}
	if req.Timezone != "" {
		updates["timezone"] = req.Timezone
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(&settings).Updates(updates).Error; err != nil {
			model.InternalError(c, "更新失败")
			return
		}
	}

	// 返回更新后的设置
	if err := database.DB.Where("id = ?", settings.ID).First(&settings).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, settings)
}

// ChangePassword 修改密码
// POST /api/settings/password
func ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 查找用户
	var user model.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		model.Fail(c, http.StatusUnauthorized, "旧密码错误")
		return
	}

	// 检查新旧密码是否相同
	if req.OldPassword == req.NewPassword {
		model.Fail(c, http.StatusBadRequest, "新旧密码不能相同")
		return
	}

	// 验证新密码强度
	isValid, msg := checkPasswordStrength(req.NewPassword)
	if !isValid {
		model.BadRequest(c, msg)
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		model.InternalError(c, "密码加密失败")
		return
	}

	// 更新密码
	if err := database.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		model.InternalError(c, "密码更新失败")
		return
	}

	model.SuccessWithMessage(c, nil, "密码修改成功")
}

// UploadAvatar 上传头像
// POST /api/settings/avatar
func UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		model.BadRequest(c, "请选择头像文件")
		return
	}

	// 验证文件大小（最大2MB）
	if file.Size > 2*1024*1024 {
		model.BadRequest(c, "文件大小不能超过2MB")
		return
	}

	// 验证文件类型
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	// 这里简化处理，实际应该检查 MIME type
	_ = allowedTypes

	// 生成保存路径（简化版：实际应该使用 OSS/S3）
	fileName := generateID() + ".jpg"
	// TODO: 实现文件上传到存储服务
	// 目前只返回文件名，实际项目需要配置存储路径
	avatarURL := "/uploads/avatars/" + fileName

	// 保存文件到本地（示例）
	// dst := "/path/to/uploads/avatars/" + fileName
	// if err := c.SaveUploadedFile(file, dst); err != nil {
	//     model.InternalError(c, "文件保存失败")
	//     return
	// }

	// 更新用户头像
	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error; err != nil {
		model.InternalError(c, "头像更新失败")
		return
	}

	model.Success(c, gin.H{
		"avatar_url": avatarURL,
		"message":    "头像上传成功",
	})
}

// =============================================================================
// 会话管理 API
// =============================================================================

// GetSessions 获取当前用户的会话列表
// GET /api/sessions
func GetSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	var sessions []model.UserSession
	if err := database.DB.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used DESC").Find(&sessions).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	// 调试日志
	slog.Debug("GetSessions", "userID", userID, "sessionsCount", len(sessions), "sessions", sessions)

	model.Success(c, sessions)
}

// RevokeSession 撤销会话（远程登出）
// DELETE /api/sessions/:id
func RevokeSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		model.BadRequest(c, "会话ID不能为空")
		return
	}

	// 验证会话归属
	var session model.UserSession
	if err := database.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "会话不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 软删除会话（标记为不活跃）
	if err := database.DB.Model(&session).Update("is_active", false).Error; err != nil {
		model.InternalError(c, "操作失败")
		return
	}

	model.SuccessWithMessage(c, nil, "会话已撤销")
}
