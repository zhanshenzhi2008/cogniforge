package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// Prompt Models - 提示词工程（阶段十七）
// =============================================================================

// Prompt 提示词资产
type Prompt struct {
	ID               string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID           string         `gorm:"type:varchar(64);not null;index:idx_prompts_user_id" json:"user_id"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description"`
	CurrentVersionID string         `gorm:"type:varchar(64)" json:"current_version_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Prompt) TableName() string { return "prompts" }

// PromptVersion 提示词版本
type PromptVersion struct {
	ID        string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	PromptID  string    `gorm:"type:varchar(64);not null;index:idx_prompt_versions_prompt_id" json:"prompt_id"`
	UserID    string    `gorm:"type:varchar(64);not null;index:idx_prompt_versions_user_id" json:"user_id"`
	Version   int       `gorm:"not null;default:1" json:"version"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Changelog string    `gorm:"type:text" json:"changelog"`
	CreatedAt time.Time `json:"created_at"`
}

func (PromptVersion) TableName() string { return "prompt_versions" }
