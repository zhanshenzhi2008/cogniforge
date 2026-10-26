package engine

import (
	"gorm.io/gorm"

	"cogniforge/internal/chat"
	"cogniforge/internal/mcp"
	"cogniforge/internal/quota"
)

// AIDeps 工作流 AI 节点依赖（阶段十八 W0）
type AIDeps struct {
	Chat  *chat.ChatService
	MCP   *mcp.Registry
	Quota *quota.Service
	DB    *gorm.DB
}
