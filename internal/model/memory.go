package model

import (
	"time"

	"gorm.io/gorm"
)

// ChatMemory 长期记忆条目（阶段十四 14.4/14.5）
// 向量存 Python pgvector（collection=chat_memories），Go 只存文本 + 过滤字段。
type ChatMemory struct {
	ID              string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID          string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	AgentID         string         `gorm:"type:varchar(64);index" json:"agent_id,omitempty"` // 空=用户全局，有值=仅该 Agent
	ConversationID  string         `gorm:"type:varchar(64)" json:"conversation_id,omitempty"` // 来源会话
	Kind           string         `gorm:"type:varchar(32);not null" json:"kind"`             // profile / preference / decision / episode
	Content         string         `gorm:"type:text;not null" json:"content"`                // 记忆文本（≤80字/条）
	Importance      int            `gorm:"default:5" json:"importance"` // 1-10，检索加权
	Metadata        []byte         `gorm:"type:text;default:'{}'" json:"metadata"` // 扩展字段（兼容 sqlite/pg）
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	LastAccessedAt  *time.Time     `json:"last_accessed_at,omitempty"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChatMemory) TableName() string {
	return "chat_memories"
}
