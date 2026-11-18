package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"cogniforge/internal/governance/rbac"
	"cogniforge/internal/shared/response"
)

// rbacSvc 门禁委托治理层（单一权限真相）。
var rbacSvc = rbac.NewRBACService()

// RequirePermission 要求特定权限的中间件
// 用法：router.GET("/api/admin/users", middleware.RequirePermission("user:list"), handler.GetUsers)
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		uid, _ := userID.(string)
		ok, err := rbacSvc.UserHasPermission(uid, permissionCode)
		if err != nil {
			if err.Error() == "用户不存在" {
				response.NotFound(c, "用户不存在")
			} else {
				response.InternalError(c, "权限检查失败")
			}
			c.Abort()
			return
		}
		if !ok {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
// 用法：router.GET("/api/admin/*", middleware.RequireRole("admin"))
func RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		uid, _ := userID.(string)
		ok, err := rbacSvc.HasRole(uid, roleCode)
		if err != nil {
			if err.Error() == "用户不存在" {
				response.NotFound(c, "用户不存在")
			} else {
				response.InternalError(c, "角色检查失败")
			}
			c.Abort()
			return
		}
		if !ok {
			response.Forbidden(c, fmt.Sprintf("需要 %s 角色", roleCode))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireAnyPermission 要求任一权限即可通过
func RequireAnyPermission(permissionCodes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		uid, _ := userID.(string)
		for _, code := range permissionCodes {
			ok, err := rbacSvc.UserHasPermission(uid, code)
			if err != nil {
				if err.Error() == "用户不存在" {
					response.NotFound(c, "用户不存在")
				} else {
					response.InternalError(c, "权限检查失败")
				}
				c.Abort()
				return
			}
			if ok {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}

// =============================================================================
// 权限指令（用于 Vue 前端）
// =============================================================================

// PermissionDirective Vue 权限指令的辅助数据
type PermissionDirective struct {
	UserPermissions map[string]bool `json:"permissions"`
}

// NewPermissionDirective 创建权限指令实例
func NewPermissionDirective(userID string) (*PermissionDirective, error) {
	permissions, err := rbacSvc.ListUserPermissionCodes(userID)
	if err != nil {
		return nil, err
	}
	permMap := make(map[string]bool)
	for _, p := range permissions {
		permMap[p] = true
	}
	return &PermissionDirective{UserPermissions: permMap}, nil
}

func (pd *PermissionDirective) HasPermission(permissionCode string) bool {
	return pd.UserPermissions[permissionCode]
}

func (pd *PermissionDirective) HasAllPermissions(permissionCodes []string) bool {
	for _, code := range permissionCodes {
		if !pd.UserPermissions[code] {
			return false
		}
	}
	return true
}

func (pd *PermissionDirective) HasAnyPermission(permissionCodes []string) bool {
	for _, code := range permissionCodes {
		if pd.UserPermissions[code] {
			return true
		}
	}
	return false
}

// CachedPermissionChecker 简单进程内缓存（按需使用）
type CachedPermissionChecker struct {
	cache map[string]map[string]bool
}

func NewCachedPermissionChecker() *CachedPermissionChecker {
	return &CachedPermissionChecker{cache: map[string]map[string]bool{}}
}

func (cpc *CachedPermissionChecker) CheckPermission(userID, permissionCode string) (bool, error) {
	if m, ok := cpc.cache[userID]; ok {
		if v, ok2 := m[permissionCode]; ok2 {
			return v, nil
		}
	}
	ok, err := rbacSvc.UserHasPermission(userID, permissionCode)
	if err != nil {
		return false, err
	}
	if _, exists := cpc.cache[userID]; !exists {
		cpc.cache[userID] = map[string]bool{}
	}
	cpc.cache[userID][permissionCode] = ok
	return ok, nil
}

func (cpc *CachedPermissionChecker) ClearCache() {
	cpc.cache = map[string]map[string]bool{}
}
