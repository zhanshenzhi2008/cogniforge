package user

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: database.DB}
}

// ListUsers 获取用户列表（管理员）
func (s *UserService) ListUsers(req *ListUsersRequest) (*ListUsersResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	query := s.db.Model(&model.User{})

	if req.Search != "" {
		query = query.Where("email LIKE ? OR name LIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	query.Count(&total)

	var users []model.User
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("查询失败")
	}

	return &ListUsersResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Users:    users,
	}, nil
}

// CreateUser 创建用户（管理员）
func (s *UserService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	var existing model.User
	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("该邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败")
	}

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

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("创建用户失败")
	}

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
	s.db.Create(&settings)

	return &user, nil
}

// GetUser 获取单个用户信息
func (s *UserService) GetUser(userID string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询失败")
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(userID string, req *UpdateUserRequest) (*model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询失败")
	}

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
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新失败")
		}
	}

	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("查询失败")
	}

	return &user, nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(userID string) error {
	if userID == "admin" {
		return fmt.Errorf("不能删除默认管理员")
	}

	result := s.db.Where("id = ?", userID).Delete(&model.User{})
	if result.Error != nil {
		return fmt.Errorf("删除失败")
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	return nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(userID string, req *UpdateStatusRequest) error {
	if userID == "admin" {
		return fmt.Errorf("不能修改默认管理员状态")
	}

	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Update("status", req.Status).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("用户不存在")
		}
		return fmt.Errorf("更新失败")
	}

	return nil
}

// UpdateCurrentUser 更新当前用户信息
func (s *UserService) UpdateCurrentUser(userID string, req *UpdateUserRequest) (*model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询失败")
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新失败")
		}
	}

	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("查询失败")
	}

	return &user, nil
}

// CheckPasswordStrength 检查密码强度
func CheckPasswordStrength(password string) (bool, string) {
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
		case (ch >= 33 && ch <= 47) || (ch >= 58 && ch <= 64) || (ch >= 91 && ch <= 96) || (ch >= 123 && ch <= 126):
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

// ============ Settings 服务 ============

// GetSettings 获取用户设置
func (s *UserService) GetSettings(userID string) (*model.UserSettings, error) {
	var settings model.UserSettings
	if err := s.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 首次访问，创建默认设置
			settings = model.UserSettings{
				ID:        generateID(),
				UserID:    userID,
				AvatarURL: "",
				Theme:     "light",
				Language:  "zh-CN",
				Timezone:  "Asia/Shanghai",
				Metadata:  make(model.JSONBMap),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			s.db.Create(&settings)
			return &settings, nil
		}
		return nil, fmt.Errorf("查询失败")
	}
	return &settings, nil
}

// UpdateSettings 更新用户设置
func (s *UserService) UpdateSettings(userID string, req *UpdateSettingsRequest) (*model.UserSettings, error) {
	var settings model.UserSettings
	if err := s.db.Where("user_id = ?", userID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新设置
			settings = model.UserSettings{
				ID:        generateID(),
				UserID:    userID,
				AvatarURL: req.AvatarURL,
				Theme:     req.Theme,
				Language:  req.Language,
				Timezone:  req.Timezone,
				Metadata:  req.Metadata,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := s.db.Create(&settings).Error; err != nil {
				return nil, fmt.Errorf("创建设置失败")
			}
			return &settings, nil
		}
		return nil, fmt.Errorf("查询失败")
	}

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
		if err := s.db.Model(&settings).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新失败")
		}
	}

	// 返回更新后的设置
	s.db.Where("id = ?", settings.ID).First(&settings)
	return &settings, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID string, req *ChangePasswordRequest) error {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("用户不存在")
		}
		return fmt.Errorf("查询失败")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 检查新旧密码是否相同
	if req.OldPassword == req.NewPassword {
		return fmt.Errorf("新旧密码不能相同")
	}

	// 验证新密码强度
	isValid, msg := CheckPasswordStrength(req.NewPassword)
	if !isValid {
		return fmt.Errorf("%s", msg)
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败")
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("密码更新失败")
	}

	return nil
}

// UploadAvatar 上传头像
func (s *UserService) UploadAvatar(userID string, filename string) (string, error) {
	avatarURL := "/uploads/avatars/" + filename
	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error; err != nil {
		return "", fmt.Errorf("头像更新失败")
	}
	return avatarURL, nil
}

// GetSessions 获取用户的会话列表
func (s *UserService) GetSessions(userID string) ([]model.UserSession, error) {
	var sessions []model.UserSession
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	return sessions, nil
}

// RevokeSession 撤销会话
func (s *UserService) RevokeSession(userID, sessionID string) error {
	var session model.UserSession
	if err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("会话不存在")
		}
		return fmt.Errorf("查询失败")
	}

	if err := s.db.Model(&session).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("操作失败")
	}
	return nil
}

// generateID 生成唯一ID
func generateID() string {
	return newID()
}
