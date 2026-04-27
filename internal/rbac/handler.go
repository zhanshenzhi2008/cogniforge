package rbac

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/response"
)

type RBACHandler struct {
	service *RBACService
}

func NewRBACHandler() *RBACHandler {
	return &RBACHandler{
		service: NewRBACService(),
	}
}

func (h *RBACHandler) InitDefaultRoles() {
	h.service.InitDefaultRoles()
}

// ============ 权限管理 ============

// ListPermissions 获取权限列表
func (h *RBACHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.service.ListPermissions()
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}
	response.Success(c, permissions)
}

// CreatePermission 创建权限
func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	permission, err := h.service.CreatePermission(&req)
	if err != nil {
		response.Fail(c, http.StatusConflict, err.Error())
		return
	}
	response.Created(c, permission)
}

// DeletePermission 删除权限
func (h *RBACHandler) DeletePermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		response.BadRequest(c, "权限ID不能为空")
		return
	}

	err := h.service.DeletePermission(permissionID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "权限已删除")
}

// ============ 角色管理 ============

// ListRoles 获取角色列表
func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles()
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}
	response.Success(c, roles)
}

// GetRole 获取单个角色
func (h *RBACHandler) GetRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.BadRequest(c, "角色ID不能为空")
		return
	}

	role, err := h.service.GetRole(roleID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, role)
}

// CreateRole 创建角色
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.service.CreateRole(&req)
	if err != nil {
		response.Fail(c, http.StatusConflict, err.Error())
		return
	}
	response.Created(c, gin.H{"message": "角色创建成功"})
}

// UpdateRole 更新角色
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.BadRequest(c, "角色ID不能为空")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.service.UpdateRole(roleID, &req)
	if err != nil {
		if err.Error() == "系统预置角色不可修改" {
			response.Fail(c, http.StatusForbidden, err.Error())
		} else {
			response.NotFound(c, err.Error())
		}
		return
	}
	response.SuccessWithMessage(c, nil, "角色更新成功")
}

// DeleteRole 删除角色
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		response.BadRequest(c, "角色ID不能为空")
		return
	}

	err := h.service.DeleteRole(roleID)
	if err != nil {
		if err.Error() == "系统预置角色不可删除" {
			response.Fail(c, http.StatusForbidden, err.Error())
		} else {
			response.NotFound(c, err.Error())
		}
		return
	}
	response.SuccessWithMessage(c, nil, "角色已删除")
}

// ============ 用户角色分配 ============

// AssignRole 为用户分配角色
func (h *RBACHandler) AssignRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	err := h.service.AssignRole(userID, req.RoleCode)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, nil, "角色分配成功")
}

// GetUserRole 获取用户角色
func (h *RBACHandler) GetUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户ID不能为空")
		return
	}

	result, err := h.service.GetUserRole(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, result)
}

// RegisterRoutes 注册路由
func (h *RBACHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// 角色管理
	rg.GET("/roles", h.ListRoles)
	rg.POST("/roles", h.CreateRole)
	rg.GET("/roles/:id", h.GetRole)
	rg.PUT("/roles/:id", h.UpdateRole)
	rg.DELETE("/roles/:id", h.DeleteRole)

	// 权限管理
	rg.GET("/permissions", h.ListPermissions)
	rg.POST("/permissions", h.CreatePermission)
	rg.DELETE("/permissions/:id", h.DeletePermission)

	// 用户角色分配
	rg.POST("/users/:id/roles", h.AssignRole)
	rg.GET("/users/:id/role", h.GetUserRole)
}
