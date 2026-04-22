package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// =============================================================================
// 用户管理 API
// =============================================================================

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"min=1"`              // 页码，从1开始
	PageSize int    `form:"page_size" binding:"min=1,max=100"` // 每页数量
	Search   string `form:"search"`                            // 搜索关键词（邮箱/姓名）
	Status   string `form:"status"`                            // 状态筛选
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Users    []model.User `json:"users"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"oneof=admin user"`               // admin, user
	Status   string `json:"status" binding:"oneof=active disabled locked"` // active, disabled, locked
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name      string `json:"name" binding:"min=2,max=100"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status" binding:"oneof=active disabled locked"`
	Role      string `json:"role" binding:"oneof=admin user"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// GetUsers 获取用户列表（管理员）
// GET /api/users
func GetUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// 构建查询
	query := database.DB.Model(&model.User{})

	// 搜索
	if req.Search != "" {
		query = query.Where("email LIKE ? OR name LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	// 状态筛选
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	var users []model.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, ListUsersResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Users:    users,
	})
}

// CreateUser 创建用户（管理员）
// POST /api/users
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 检查邮箱是否已存在
	var existing model.User
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		model.Fail(c, http.StatusConflict, "该邮箱已被注册")
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		model.InternalError(c, "密码加密失败")
		return
	}

	// 创建用户
	user := model.User{
		ID:        generateID(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		Status:    req.Status,
		Role:      req.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		model.InternalError(c, "创建用户失败")
		return
	}

	// 创建默认用户设置
	settings := model.UserSettings{
		ID:        generateID(),
		UserID:    user.ID,
		AvatarURL: "",
		Theme:     "light",
		Language:  "zh-CN",
		Timezone:  "Asia/Shanghai",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	database.DB.Create(&settings)

	model.Created(c, user)
}

// GetUser 获取单个用户信息（管理员）
// GET /api/users/:id
func GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "用户ID不能为空")
		return
	}

	var user model.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, user)
}

// UpdateUser 更新用户信息（管理员）
// PUT /api/users/:id
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "用户ID不能为空")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 查找用户
	var user model.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			model.InternalError(c, "更新失败")
			return
		}
	}

	// 返回更新后的用户信息
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, user)
}

// DeleteUser 删除用户（管理员）- 软删除
// DELETE /api/users/:id
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "用户ID不能为空")
		return
	}

	// 不允许删除默认管理员
	if id == "admin" {
		model.Fail(c, http.StatusForbidden, "不能删除默认管理员")
		return
	}

	// 查找用户
	var user model.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 软删除
	if err := database.DB.Delete(&user).Error; err != nil {
		model.InternalError(c, "删除失败")
		return
	}

	model.SuccessWithMessage(c, nil, "用户已删除")
}

// UpdateUserStatus 更新用户状态（管理员）
// PATCH /api/users/:id/status
func UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "用户ID不能为空")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active disabled locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 不允许修改默认管理员状态
	if id == "admin" {
		model.Fail(c, http.StatusForbidden, "不能修改默认管理员状态")
		return
	}

	// 更新状态
	if err := database.DB.Model(&model.User{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "更新失败")
		return
	}

	model.SuccessWithMessage(c, nil, "状态已更新")
}

// UpdateCurrentUser 更新当前用户信息（个人设置）
// PUT /api/me
func UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		model.Unauthorized(c, "未登录")
		return
	}

	var req UpdateUserRequest
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

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			model.InternalError(c, "更新失败")
			return
		}
	}

	// 返回更新后的用户信息
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, user)
}

// =============================================================================
// 密码强度检查辅助函数
// =============================================================================

// checkPasswordStrength 检查密码强度
func checkPasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "密码长度至少8位"
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case ch >= 33 && ch <= 47 || ch >= 58 && ch <= 64 || ch >= 91 && ch <= 96 || ch >= 123 && ch <= 126:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "密码必须包含大写字母"
	}
	if !hasLower {
		return false, "密码必须包含小写字母"
	}
	if !hasDigit {
		return false, "密码必须包含数字"
	}
	if !hasSpecial {
		return false, "密码必须包含特殊字符"
	}

	return true, "密码强度合格"
}
