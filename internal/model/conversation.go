package model

import (
	"time"

	"gorm.io/gorm"
)

// ConversationMessage 一条已保存的对话消息（Playground 历史）
type ConversationMessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
	Time    string `json:"time,omitempty"`
}

// ChatConversation 用户聊天历史（Playground；不依赖 Agent）
type ChatConversation struct {
	ID        string                `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID    string                `gorm:"type:varchar(64);not null;index" json:"user_id"`
	AgentID   string                `gorm:"type:varchar(64);index" json:"agent_id"`
	Title     string                `gorm:"type:varchar(255)" json:"title"`
	Model     string                `gorm:"type:varchar(128)" json:"model"`
	Messages  []ConversationMessage `gorm:"serializer:json;type:jsonb" json:"messages"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (ChatConversation) TableName() string {
	return "chat_conversations"
}
