package auth

import (
	"cogniforge/internal/model"
)

// ============ 请求结构 ============

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

type ApiKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// ============ 响应结构 ============

type AuthData struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ToUserResponse 将 model.User 转换为 UserResponse
func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToAuthData 将 model.User 转换为 AuthData
func ToAuthData(token string, user *model.User) AuthData {
	return AuthData{
		Token: token,
		User:  ToUserResponse(user),
	}
}
