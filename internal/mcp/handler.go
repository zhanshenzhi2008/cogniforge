package mcp

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/model"
	"cogniforge/internal/response"
)

// Handler MCP HTTP 处理
type Handler struct {
	svc      *Service
	registry *Registry
}

// NewHandler 创建 MCP Handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		svc:      NewService(db),
		registry: NewRegistry(),
	}
}

// Registry 返回 registry 实例（供其他模块使用）
func (h *Handler) Registry() *Registry {
	return h.registry
}

// ServerListItem 列表项：Server + 已发现工具数（对标 Cursor / Comis 卡片信息密度）
type ServerListItem struct {
	model.McpServer
	ToolCount int `json:"tool_count"`
}

// ListServers 获取 MCP Server 列表
func (h *Handler) ListServers(c *gin.Context) {
	userID := c.GetString("user_id")
	servers, err := h.svc.ListServers(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 内置在前，便于列表一眼看到工具能力；再接用户自定义
	allServers := make([]model.McpServer, 0, len(model.BuiltInMcpServers)+len(servers))
	for _, b := range model.BuiltInMcpServers {
		allServers = append(allServers, b)
	}
	allServers = append(allServers, servers...)

	items := make([]ServerListItem, 0, len(allServers))
	for _, srv := range allServers {
		items = append(items, ServerListItem{
			McpServer: srv,
			ToolCount: h.registry.ToolCount(srv.ID),
		})
	}

	response.Success(c, items)
}

// CreateServer 创建 MCP Server
func (h *Handler) CreateServer(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	srv, err := h.svc.CreateServer(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	h.refreshUserRegistry(userID)
	response.Created(c, srv)
}

// GetServer 获取 MCP Server 详情
func (h *Handler) GetServer(c *gin.Context) {
	userID := c.GetString("user_id")
	serverID := c.Param("id")

	srv, err := h.svc.GetServer(userID, serverID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, srv)
}

// UpdateServer 更新 MCP Server
func (h *Handler) UpdateServer(c *gin.Context) {
	userID := c.GetString("user_id")
	serverID := c.Param("id")

	var req UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	srv, err := h.svc.UpdateServer(userID, serverID, &req)
	if err != nil {
		if err.Error() == "内置 MCP Server 不可编辑" {
			response.Forbidden(c, err.Error())
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}

	h.refreshUserRegistry(userID)
	response.Success(c, srv)
}

// DeleteServer 删除 MCP Server
func (h *Handler) DeleteServer(c *gin.Context) {
	userID := c.GetString("user_id")
	serverID := c.Param("id")

	err := h.svc.DeleteServer(userID, serverID)
	if err != nil {
		if err.Error() == "内置 MCP Server 不可删除" {
			response.Forbidden(c, err.Error())
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}

	h.refreshUserRegistry(userID)
	response.SuccessWithMessage(c, nil, "MCP Server 已删除")
}

// ListTools 列出某 Server 的工具；?refresh=1 强制重新发现 HTTP 工具
func (h *Handler) ListTools(c *gin.Context) {
	userID := c.GetString("user_id")
	serverID := c.Param("id")
	force := c.Query("refresh") == "1" || c.Query("refresh") == "true"

	srv, err := h.svc.GetServer(userID, serverID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if srv.Type == model.McpServerTypeHTTP && srv.URL != "" {
		if force {
			if err := h.registry.RediscoverHTTP(srv); err != nil {
				response.BadRequest(c, "刷新工具失败: "+err.Error())
				return
			}
		} else {
			h.registry.EnsureHTTPTools(srv)
		}
	}

	tools := h.svc.GetServerTools(serverID, h.registry)
	response.Success(c, tools)
}

// RefreshTools 强制重新 tools/list（HTTP）
func (h *Handler) RefreshTools(c *gin.Context) {
	userID := c.GetString("user_id")
	serverID := c.Param("id")
	srv, err := h.svc.GetServer(userID, serverID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	if srv.Type != model.McpServerTypeHTTP {
		response.BadRequest(c, "内置 Server 无需刷新")
		return
	}
	if err := h.registry.RediscoverHTTP(srv); err != nil {
		response.BadRequest(c, "刷新工具失败: "+err.Error())
		return
	}
	tools := h.svc.GetServerTools(serverID, h.registry)
	response.Success(c, tools)
}

// refreshUserRegistry 用当前用户全部自定义 Server 刷新（含 HTTP 工具发现）
func (h *Handler) refreshUserRegistry(userID string) {
	servers, err := h.svc.ListServers(userID)
	if err != nil {
		return
	}
	h.registry.Refresh(servers)
}

// TestTool 测试调用工具（调试用）
func (h *Handler) TestTool(c *gin.Context) {
	serverID := c.Param("id")
	var req struct {
		ToolName string                 `json:"tool_name" binding:"required"`
		Args     map[string]interface{} `json:"args"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	toolKey := serverID + ":" + req.ToolName
	tool := h.registry.GetBuiltInTool(toolKey)
	if tool == nil {
		response.NotFound(c, "工具不存在")
		return
	}

	// 内置工具直接执行
	executor := NewBuiltInExecutor()
	result, err := executor.Execute(tool.Name, req.Args)
	if err != nil {
		response.InternalError(c, "工具执行失败: "+err.Error())
		return
	}
	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	servers := rg.Group("/mcp/servers")
	{
		servers.GET("", h.ListServers)
		servers.POST("", h.CreateServer)
		servers.GET("/:id", h.GetServer)
		servers.PUT("/:id", h.UpdateServer)
		servers.DELETE("/:id", h.DeleteServer)
		servers.GET("/:id/tools", h.ListTools)
		servers.POST("/:id/tools/refresh", h.RefreshTools)
		servers.POST("/:id/test", h.TestTool)
	}
}
