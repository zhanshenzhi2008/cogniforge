package model

import (
	"time"

	"gorm.io/gorm"
)

// ConversationMessage 一条已保存的对话消息（Playground 历史）
type ConversationMessage struct {
	ID      string   `json:"id"`
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"` // data URL 或 http(s) 图，用于回显
	Time    string   `json:"time,omitempty"`
}

// ConversationQueueItem 待发送排队（流式中 Enter 入队）
type ConversationQueueItem struct {
	ID      string   `json:"id"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
	Status  string   `json:"status"` // queued | sending
	Sort    int      `json:"sort"`
}

// ChatConversation 用户聊天历史（Playground；不依赖 Agent）
type ChatConversation struct {
	ID           string                   `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID       string                   `gorm:"type:varchar(64);not null;index" json:"user_id"`
	AgentID      string                   `gorm:"type:varchar(64);index" json:"agent_id"`
	Title        string                   `gorm:"type:varchar(255)" json:"title"`
	Model        string                   `gorm:"type:varchar(128)" json:"model"`
	Pinned       bool                     `gorm:"not null;default:false;index" json:"pinned"`
	Messages     []ConversationMessage    `gorm:"serializer:json;type:jsonb" json:"messages"`
	MessageQueue []ConversationQueueItem  `gorm:"serializer:json;type:jsonb" json:"message_queue"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	DeletedAt    gorm.DeletedAt           `gorm:"index" json:"-"`
}

func (ChatConversation) TableName() string {
	return "chat_conversations"
}
