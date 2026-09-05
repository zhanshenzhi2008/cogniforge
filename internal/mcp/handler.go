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

// ListServers 获取 MCP Server 列表
func (h *Handler) ListServers(c *gin.Context) {
	userID := c.GetString("user_id")
	servers, err := h.svc.ListServers(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 合并内置 server
	allServers := servers
	for _, b := range model.BuiltInMcpServers {
		b := b
		allServers = append(allServers, b)
	}

	response.Success(c, allServers)
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

	// 刷新 Registry
	h.registry.Refresh([]model.McpServer{*srv})

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

	// 刷新 Registry
	servers, _ := h.svc.ListServers(userID)
	h.registry.Refresh(servers)

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

	// 刷新 Registry
	servers, _ := h.svc.ListServers(userID)
	h.registry.Refresh(servers)

	response.SuccessWithMessage(c, nil, "MCP Server 已删除")
}

// ListTools 列出某 Server 的工具
func (h *Handler) ListTools(c *gin.Context) {
	serverID := c.Param("id")
	tools := h.svc.GetServerTools(serverID, h.registry)
	response.Success(c, tools)
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
		servers.POST("/:id/test", h.TestTool)
	}
}
