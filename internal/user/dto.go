package user

import (
	"cogniforge/internal/model"
)

// ============ 请求结构 ============

type ListUsersRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Search   string `form:"search"`
	Status   string `form:"status"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"oneof=admin user"`
	Status   string `json:"status" binding:"oneof=active disabled locked"`
}

type UpdateUserRequest struct {
	Name      string `json:"name" binding:"min=2,max=100"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status" binding:"oneof=active disabled locked"`
	Role      string `json:"role" binding:"oneof=admin user"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled locked"`
}

// ============ Settings 请求/响应结构 ============

type UpdateSettingsRequest struct {
	AvatarURL string                 `json:"avatar_url"`
	Theme     string                 `json:"theme"`
	Language  string                 `json:"language"`
	Timezone  string                 `json:"timezone"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type SettingsResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	AvatarURL string                 `json:"avatar_url"`
	Theme     string                 `json:"theme"`
	Language  string                 `json:"language"`
	Timezone  string                 `json:"timezone"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type AvatarUploadResponse struct {
	AvatarURL string `json:"avatar_url"`
	Message   string `json:"message"`
}

// ============ 响应结构 ============

type ListUsersResponse struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Users    []model.User `json:"users"`
}

// ============ 辅助函数 ============

// ToListUsersResponse 将用户列表转换为响应格式
func ToListUsersResponse(total int64, page, pageSize int, users []model.User) ListUsersResponse {
	return ListUsersResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Users:    users,
	}
}

// ToSettingsResponse 将 UserSettings 转换为响应格式
func ToSettingsResponse(settings *model.UserSettings) SettingsResponse {
	metadata := make(map[string]interface{})
	for k, v := range settings.Metadata {
		metadata[k] = v
	}

	return SettingsResponse{
		ID:        settings.ID,
		UserID:    settings.UserID,
		AvatarURL: settings.AvatarURL,
		Theme:     settings.Theme,
		Language:  settings.Language,
		Timezone:  settings.Timezone,
		Metadata:  metadata,
		CreatedAt: settings.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: settings.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
