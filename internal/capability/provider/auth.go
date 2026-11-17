package provider

import (
	"errors"

	"cogniforge/internal/infra/database"
	"cogniforge/internal/model"
)

var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrProviderForbidden = errors.New("forbidden")
	ErrProviderBadRequest = errors.New("bad request")
)

// Actor 当前操作者
type Actor struct {
	UserID  string
	IsAdmin bool
}

// ResolveActor 从 userID 查是否 admin
func ResolveActor(userID string) Actor {
	a := Actor{UserID: userID}
	if userID == "" || database.DB == nil {
		return a
	}
	var u model.User
	if err := database.DB.Where("id = ?", userID).Select("id, role").First(&u).Error; err == nil {
		a.IsAdmin = u.Role == "admin"
	}
	return a
}
