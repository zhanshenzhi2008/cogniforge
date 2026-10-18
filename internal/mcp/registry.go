package mcp

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"cogniforge/internal/database"
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
			Description: "搜索互联网获取实时信息。用户问天气、新闻、股价、赛事比分、最新事件等时效性问题时必须调用本工具，不要让用户去 App 自行搜索。",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词，如「北京今天天气」「今日热点新闻」",
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

// Refresh 从数据库重新加载用户 MCP Server，并对 HTTP 类型尝试 tools/list
func (r *Registry) Refresh(servers []model.McpServer) {
	r.mu.Lock()
	// 清空用户 server（保留内置）
	for k := range r.servers {
		if !model.IsBuiltInServer(k) {
			delete(r.servers, k)
			prefix := k + ":"
			for toolKey := range r.tools {
				if len(toolKey) >= len(prefix) && toolKey[:len(prefix)] == prefix {
					delete(r.tools, toolKey)
				}
			}
		}
	}
	enabledHTTP := make([]model.McpServer, 0)
	for _, srv := range servers {
		if !srv.Enabled {
			continue
		}
		s := srv
		r.servers[s.ID] = &s
		if s.Type == model.McpServerTypeHTTP && s.URL != "" {
			enabledHTTP = append(enabledHTTP, s)
		}
	}
	r.mu.Unlock()

	for _, srv := range enabledHTTP {
		r.discoverAndRegisterHTTP(srv)
	}
	slog.Info("MCP registry refreshed", "total_servers", len(r.GetServers()), "http_discovered", len(enabledHTTP))
}

// discoverAndRegisterHTTP 拉取远程工具并写入 Registry（失败只打日志，不阻断列表）
func (r *Registry) discoverAndRegisterHTTP(srv model.McpServer) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tools, err := DiscoverHTTPTools(ctx, srv.ID, srv.URL, srv.AuthHeader)
	if err != nil {
		slog.Warn("MCP HTTP tools/list failed", "server_id", srv.ID, "url", srv.URL, "error", err)
		return
	}
	r.RegisterHTTPTools(srv.ID, tools)
	slog.Info("MCP HTTP tools registered", "server_id", srv.ID, "count", len(tools))
}

// RegisterHTTPTools 覆盖注册某 HTTP Server 的工具列表
func (r *Registry) RegisterHTTPTools(serverID string, tools []ToolDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	prefix := serverID + ":"
	for toolKey := range r.tools {
		if len(toolKey) >= len(prefix) && toolKey[:len(prefix)] == prefix {
			delete(r.tools, toolKey)
		}
	}
	for _, t := range tools {
		t.ServerID = serverID
		t.ServerType = model.McpServerTypeHTTP
		r.tools[serverID+":"+t.Name] = t
	}
}

// EnsureHTTPTools 若该 Server 尚无工具则尝试发现一次
func (r *Registry) EnsureHTTPTools(srv *model.McpServer) {
	if srv == nil || srv.Type != model.McpServerTypeHTTP || srv.URL == "" {
		return
	}
	if len(r.GetToolsForServer(srv.ID)) > 0 {
		r.mu.Lock()
		r.servers[srv.ID] = srv
		r.mu.Unlock()
		return
	}
	r.mu.Lock()
	cp := *srv
	r.servers[srv.ID] = &cp
	r.mu.Unlock()
	r.discoverAndRegisterHTTP(cp)
}

// EnsureHTTPToolsByIDs 对话前按 ID 拉取用户 HTTP Server 并发现工具（进程重启后也能用）
func (r *Registry) EnsureHTTPToolsByIDs(userID string, serverIDs []string) {
	if r == nil || userID == "" || len(serverIDs) == 0 || database.DB == nil {
		return
	}
	for _, id := range serverIDs {
		if id == "" || model.IsBuiltInServer(id) {
			continue
		}
		var srv model.McpServer
		if err := database.DB.Where("id = ? AND user_id = ? AND enabled = ?", id, userID, true).First(&srv).Error; err != nil {
			continue
		}
		r.EnsureHTTPTools(&srv)
	}
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
	prefix := serverID + ":"
	var out []ToolDefinition
	for key, t := range r.tools {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
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

// GetToolsForAgent 获取 Agent 绑定的所有工具（按 MCP Server ID 过滤）
func (r *Registry) GetToolsForAgent(serverIDs []string) []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var out []ToolDefinition
	for _, sid := range serverIDs {
		if sid == "" {
			continue
		}
		prefix := sid + ":"
		for key, t := range r.tools {
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				if !seen[key] {
					seen[key] = true
					out = append(out, t)
				}
			}
		}
	}
	return out
}

// GetBuiltInTool 获取内置工具定义。
// id 可为 registry key（web_search:search）或函数名（web_search_search）。
func (r *Registry) GetBuiltInTool(id string) *ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for key, t := range r.tools {
		if key == id || t.Name == id {
			cp := t
			return &cp
		}
	}
	return nil
}
