package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// Prompt Models - 提示词工程（阶段十七 / 工业落地）
// =============================================================================

const (
	PromptVersionDraft     = "draft"
	PromptVersionPublished = "published"
)

// PromptVarDef 提示词变量定义（工业：声明式，评估/运行时代入）
type PromptVarDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptVarDefs 变量定义列表（Postgres JSONB / SQLite TEXT）
type PromptVarDefs []PromptVarDef

func (j *PromptVarDefs) Scan(value any) error {
	if value == nil {
		*j = PromptVarDefs{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			bytes = []byte(s)
		} else {
			return fmt.Errorf("failed to scan PromptVarDefs: not a byte slice")
		}
	}
	if len(bytes) == 0 {
		*j = PromptVarDefs{}
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (j PromptVarDefs) Value() (driver.Value, error) {
	if j == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(j)
}

// Prompt 提示词资产
type Prompt struct {
	ID               string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID           string         `gorm:"type:varchar(64);not null;index:idx_prompts_user_id" json:"user_id"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description"`
	CurrentVersionID string         `gorm:"type:varchar(64)" json:"current_version_id"`
	Variables        PromptVarDefs  `gorm:"type:jsonb" json:"variables"` // 声明式变量
	Tags             JSONBArray     `gorm:"type:jsonb" json:"tags"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Prompt) TableName() string { return "prompts" }

// PromptVersion 提示词版本（draft → published；current 指向已发布版）
type PromptVersion struct {
	ID          string     `gorm:"primaryKey;type:varchar(64)" json:"id"`
	PromptID    string     `gorm:"type:varchar(64);not null;index:idx_prompt_versions_prompt_id" json:"prompt_id"`
	UserID      string     `gorm:"type:varchar(64);not null;index:idx_prompt_versions_user_id" json:"user_id"`
	Version     int        `gorm:"not null;default:1" json:"version"`
	Status      string     `gorm:"type:varchar(32);not null;default:'published';index:idx_prompt_ver_status" json:"status"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	Changelog   string     `gorm:"type:text" json:"changelog"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (PromptVersion) TableName() string { return "prompt_versions" }
