package handler

import "github.com/google/uuid"

// generateID 生成唯一ID
func generateID() string {
	return uuid.New().String()
}
