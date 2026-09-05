package mcp

import (
	"log/slog"
	"sync"

	"cogniforge/internal/model"
)

// ToolDefinition 工具的 JSON Schema 描述
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"` // JSON Schema
	ServerID    string                 `json:"server_id"`
	ServerType  string                 `json:"server_type"` // "built-in" | "http"
}

// Registry MCP Server + 工具注册表（内存缓存）
type Registry struct {
	mu      sync.RWMutex
	servers map[string]*model.McpServer // serverID -> server
	tools   map[string]ToolDefinition   // "serverID:toolName" -> tool
}

// NewRegistry 创建 Registry 并注册内置工具
func NewRegistry() *Registry {
	r := &Registry{
		servers: make(map[string]*model.McpServer),
		tools:   make(map[string]ToolDefinition),
	}
	// 注册内置工具
	for _, srv := range model.BuiltInMcpServers {
		s := srv
		r.servers[s.ID] = &s
		r.registerBuiltInTools(&s)
	}
	slog.Info("MCP registry initialized", "built_in_count", len(model.BuiltInMcpServers))
	return r
}

// registerBuiltInTools 注册内置工具的 schema
func (r *Registry) registerBuiltInTools(srv *model.McpServer) {
	switch srv.ID {
	case "fetch":
		r.tools["fetch:get_page"] = ToolDefinition{
			Name:        "fetch_get_page",
			Description: "获取指定 URL 的网页内容并提取正文文本",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "目标网页 URL，必须是 http:// 或 https:// 开头",
					},
					"max_length": map[string]interface{}{
						"type":        "integer",
						"description": "最大返回字符数，默认 5000",
						"default":     5000,
					},
				},
				"required": []string{"url"},
			},
			ServerID:   srv.ID,
			ServerType: srv.Type,
		}
	case "web_search":
		r.tools["web_search:search"] = ToolDefinition{
			Name:        "web_search_search",
			Description: "搜索互联网获取与关键词相关的信息列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词，建议简洁具体",
					},
					"max_results": map[string]interface{}{
						"type":        "integer",
						"description": "最大返回结果数，默认 5",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
			ServerID:   srv.ID,
			ServerType: srv.Type,
		}
	case "code_execute":
		for _, lang := range []string{"python", "go", "javascript"} {
			langName := lang
			r.tools["code_execute:"+lang] = ToolDefinition{
				Name:        "code_execute_" + lang,
				Description: "在沙箱中执行 " + langName + " 代码并返回输出",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"code": map[string]interface{}{
							"type":        "string",
							"description": langName + " 代码内容",
						},
						"timeout": map[string]interface{}{
							"type":        "integer",
							"description": "超时时间（秒），默认 30",
							"default":     30,
						},
					},
					"required": []string{"code"},
				},
				ServerID:   srv.ID,
				ServerType: srv.Type,
			}
		}
	}
}

// Refresh 从数据库重新加载用户 MCP Server
func (r *Registry) Refresh(servers []model.McpServer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 清空用户 server（保留内置）
	for k := range r.servers {
		if !model.IsBuiltInServer(k) {
			delete(r.servers, k)
			// 清空该 server 的工具
			for toolKey := range r.tools {
				if len(toolKey) > len(k) && toolKey[:len(k)] == k {
					delete(r.tools, toolKey)
				}
			}
		}
	}
	// 重新注册用户 server（暂不注册 HTTP 工具，待 tool list 接口实现）
	for _, srv := range servers {
		if srv.Enabled {
			s := srv
			r.servers[s.ID] = &s
		}
	}
	slog.Info("MCP registry refreshed", "total_servers", len(r.servers))
}

// GetServers 返回所有已注册 Server
func (r *Registry) GetServers() []*model.McpServer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.McpServer, 0, len(r.servers))
	for _, s := range r.servers {
		out = append(out, s)
	}
	return out
}

// GetServer 根据 ID 获取 Server
func (r *Registry) GetServer(id string) *model.McpServer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.servers[id]
}

// GetToolsForServer 获取某 Server 的所有工具
func (r *Registry) GetToolsForServer(serverID string) []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []ToolDefinition
	for key, t := range r.tools {
		if len(key) > len(serverID) && key[:len(serverID)] == serverID {
			out = append(out, t)
		}
	}
	return out
}

// GetAllTools 获取所有工具
func (r *Registry) GetAllTools() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// GetToolsForAgent 获取 Agent 绑定的所有工具
func (r *Registry) GetToolsForAgent(serverIDs []string) []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var out []ToolDefinition
	for _, sid := range serverIDs {
		// 内置工具直接注册
		for key, t := range r.tools {
			if len(key) > len(sid) && key[:len(sid)] == sid {
				if !seen[key] {
					seen[key] = true
					out = append(out, t)
				}
			}
		}
	}
	return out
}

// GetBuiltInTool 获取内置工具定义
func (r *Registry) GetBuiltInTool(id string) *ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// id 格式：serverID:toolName
	for key, t := range r.tools {
		if key == id {
			return &t
		}
	}
	return nil
}
