package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// =============================================================================
// RBAC 权限中间件
// =============================================================================

// RequirePermission 要求特定权限的中间件
// 用法：router.GET("/api/admin/users", middleware.RequirePermission("user:list"), handler.GetUsers)
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			model.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 获取用户角色
		var user model.User
		if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			model.NotFound(c, "用户不存在")
			c.Abort()
			return
		}

		// 如果是系统管理员，拥有所有权限
		if user.Role == "admin" {
			c.Next()
			return
		}

		// 检查权限
		hasPermission, err := checkUserPermission(userID.(string), permissionCode)
		if err != nil {
			model.InternalError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			model.Forbidden(c, "权限不足")
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
		userID, exists := c.Get("userID")
		if !exists {
			model.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		var user model.User
		if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			model.NotFound(c, "用户不存在")
			c.Abort()
			return
		}

		// 检查角色
		if user.Role != roleCode {
			model.Forbidden(c, fmt.Sprintf("需要 %s 角色", roleCode))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin 要求管理员权限的中间件
// 用法：router.GET("/api/admin/users", middleware.RequireAdmin())
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireAnyPermission 要求任一权限即可通过
func RequireAnyPermission(permissionCodes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			model.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		// 获取用户角色
		var user model.User
		if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			model.NotFound(c, "用户不存在")
			c.Abort()
			return
		}

		// 如果是系统管理员，拥有所有权限
		if user.Role == "admin" {
			c.Next()
			return
		}

		// 检查是否拥有任一权限
		for _, code := range permissionCodes {
			hasPermission, err := checkUserPermission(userID.(string), code)
			if err != nil {
				model.InternalError(c, "权限检查���败")
				c.Abort()
				return
			}
			if hasPermission {
				c.Next()
				return
			}
		}

		model.Forbidden(c, "权限不足")
		c.Abort()
	}
}

// =============================================================================
// 权限检查辅助函数
// =============================================================================

// checkUserPermission 检查用户是否拥有指定权限
func checkUserPermission(userID, permissionCode string) (bool, error) {
	var count int64

	// 通过角色关联查询权限
	err := database.DB.Model(&model.RolePermission{}).
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN users ON users.role = (SELECT code FROM roles WHERE id = role_permissions.role_id)").
		Where("users.id = ? AND permissions.code = ? AND role_permissions.deleted_at IS NULL", userID, permissionCode).
		Count(&count).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	return count > 0, nil
}

// getUserPermissions 获取用户所有权限代码
func getUserPermissions(userID string) ([]string, error) {
	var permissions []string

	err := database.DB.Model(&model.Permission{}).
		Select("permissions.code").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN users ON users.role = (SELECT code FROM roles WHERE id = rp.role_id)").
		Where("users.id = ? AND rp.deleted_at IS NULL", userID).
		Pluck("code", &permissions).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return permissions, nil
}

// GetPermissionCheckFunction 返回一个权限检查函数（用于前端传递）
// 这个函数可以用来在需要时动态检查权限
func GetPermissionCheckFunction() func(permissionCode string) bool {
	return func(permissionCode string) bool {
		// 这个函数需要在上下文中获取用户ID
		// 实际使用时应该从 JWT token 或 session 中获取
		// 这里返回一个占位实现
		return false
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
	permissions, err := getUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	permMap := make(map[string]bool)
	for _, p := range permissions {
		permMap[p] = true
	}

	return &PermissionDirective{
		UserPermissions: permMap,
	}, nil
}

// HasPermission 检查是否有权限
func (pd *PermissionDirective) HasPermission(permissionCode string) bool {
	return pd.UserPermissions[permissionCode]
}

// HasAllPermissions 检查是否拥有所有权限
func (pd *PermissionDirective) HasAllPermissions(permissionCodes []string) bool {
	for _, code := range permissionCodes {
		if !pd.UserPermissions[code] {
			return false
		}
	}
	return true
}

// HasAnyPermission 检查是否拥有任一权限
func (pd *PermissionDirective) HasAnyPermission(permissionCodes []string) bool {
	for _, code := range permissionCodes {
		if pd.UserPermissions[code] {
			return true
		}
	}
	return false
}

// =============================================================================
// 权限缓存（可选优化）
// =============================================================================

// CachedPermissionChecker 带缓存的权限检查器
type CachedPermissionChecker struct {
	cache map[string]bool
}

// NewCachedPermissionChecker 创建带缓存的权限检查器
func NewCachedPermissionChecker() *CachedPermissionChecker {
	return &CachedPermissionChecker{
		cache: make(map[string]bool),
	}
}

// CheckPermission 检查权限（带缓存）
func (cpc *CachedPermissionChecker) CheckPermission(userID, permissionCode string) (bool, error) {
	key := fmt.Sprintf("%s:%s", userID, permissionCode)
	if cached, ok := cpc.cache[key]; ok {
		return cached, nil
	}

	hasPerm, err := checkUserPermission(userID, permissionCode)
	if err != nil {
		return false, err
	}

	cpc.cache[key] = hasPerm
	return hasPerm, nil
}

// ClearCache 清除缓存
func (cpc *CachedPermissionChecker) ClearCache() {
	cpc.cache = make(map[string]bool)
}
