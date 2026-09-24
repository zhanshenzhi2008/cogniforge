package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// MCP Models - Model Context Protocol
// =============================================================================

// McpServer MCP Server 配置表（阶段十五 15.1）
type McpServer struct {
	ID         string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	UserID     string         `gorm:"type:varchar(64);not null;index" json:"user_id"`
	Name       string         `gorm:"type:varchar(255)" json:"name"`                  // 显示名，如"网页抓取"
	Type       string         `gorm:"type:varchar(50)" json:"type"`                  // "built-in" | "http"
	URL        string         `gorm:"type:varchar(500)" json:"url"`                  // HTTP URL（built-in 留空）
	AuthHeader string         `gorm:"type:varchar(500)" json:"auth_header"`           // 可选：Authorization header
	Enabled    bool           `gorm:"default:true" json:"enabled"`
	Metadata   JSONBMap       `gorm:"type:jsonb" json:"metadata"`                    // 扩展配置
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (McpServer) TableName() string {
	return "mcp_servers"
}

// AgentMcpServer Agent 与 MCP Server 关联表
type AgentMcpServer struct {
	AgentID  string `gorm:"primaryKey;type:varchar(64)" json:"agent_id"`
	ServerID string `gorm:"primaryKey;type:varchar(64)" json:"server_id"`
	Enabled  bool   `gorm:"default:true" json:"enabled"`
	Priority int    `gorm:"default:0" json:"priority"` // 工具优先级
}

func (AgentMcpServer) TableName() string {
	return "agent_mcp_servers"
}

// =============================================================================
// MCP Server 类型常量
// =============================================================================

const (
	McpServerTypeBuiltIn = "built-in"
	McpServerTypeHTTP    = "http"
)

// BuiltInMcpServers 内置 MCP Server 定义（代码内置，无需数据库记录）
var BuiltInMcpServers = []McpServer{
	{
		ID:   "fetch",
		Type: McpServerTypeBuiltIn,
		Name: "网页抓取",
		Metadata: JSONBMap{
			"description": "获取网页内容并提取正文",
			"icon":        "🌐",
		},
	},
	{
		ID:   "web_search",
		Type: McpServerTypeBuiltIn,
		Name: "网络搜索",
		Metadata: JSONBMap{
			"description": "搜索互联网获取相关信息",
			"icon":        "🔍",
		},
	},
	{
		ID:   "code_execute",
		Type: McpServerTypeBuiltIn,
		Name: "代码执行",
		Metadata: JSONBMap{
			"description": "在沙箱中执行代码（Python/Go/JS）",
			"icon":        "💻",
		},
	},
}

// GetBuiltInServer 根据 ID 查找内置 Server
func GetBuiltInServer(id string) *McpServer {
	for i := range BuiltInMcpServers {
		if BuiltInMcpServers[i].ID == id {
			return &BuiltInMcpServers[i]
		}
	}
	return nil
}

// IsBuiltInServer 判断是否为内置 Server
func IsBuiltInServer(id string) bool {
	return GetBuiltInServer(id) != nil
}
