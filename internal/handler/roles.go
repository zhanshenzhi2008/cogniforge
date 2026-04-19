package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

// =============================================================================
// 角色权限管理 API
// =============================================================================

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Group string `json:"group"`
	Desc  string `json:"description"`
}

// RoleResponse 角色响应（带权限列表）
type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Code        string               `json:"code"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	IsDefault   bool                 `json:"is_default"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required,min=2,max=100"`
	Code          string   `json:"code" binding:"required,alpha,min=2,max=50"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"` // 权限ID列表
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"min=2,max=100"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Code  string `json:"code" binding:"required,min=2,max=100"`
	Name  string `json:"name" binding:"required,min=2,max=255"`
	Group string `json:"group"`
	Desc  string `json:"description"`
}

// =============================================================================
// 权限管理
// =============================================================================

// ListPermissions 获取权限列表
// GET /api/permissions
func ListPermissions(c *gin.Context) {
	var permissions []model.Permission
	if err := database.DB.Order("group, code").Find(&permissions).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, permissions)
}

// CreatePermission 创建权限（管理员）
// POST /api/permissions
func CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 检查权限代码是否已存在
	var existing model.Permission
	if err := database.DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		model.Fail(c, http.StatusConflict, "权限代码已存在")
		return
	}

	permission := model.Permission{
		ID:        generateID(),
		Code:      req.Code,
		Name:      req.Name,
		Group:     req.Group,
		Desc:      req.Desc,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.Create(&permission).Error; err != nil {
		model.InternalError(c, "创建失败")
		return
	}

	model.Created(c, permission)
}

// DeletePermission 删除权限（管理员）
// DELETE /api/permissions/:id
func DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "权限ID不能为空")
		return
	}

	var permission model.Permission
	if err := database.DB.Where("id = ?", id).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "权限不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 删除权限
	if err := database.DB.Delete(&permission).Error; err != nil {
		model.InternalError(c, "删除失败")
		return
	}

	model.SuccessWithMessage(c, nil, "权限已删除")
}

// =============================================================================
// 角色管理
// =============================================================================

// ListRoles 获取角色列表
// GET /api/roles
func ListRoles(c *gin.Context) {
	var roles []model.Role
	if err := database.DB.Order("is_system DESC, name").Find(&roles).Error; err != nil {
		model.InternalError(c, "查询失败")
		return
	}

	// 为每个角色加载权限
	roleResponses := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		var permissions []model.Permission
		database.DB.Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
			Where("rp.role_id = ? AND rp.deleted_at IS NULL", role.ID).
			Find(&permissions)

		permissionResponses := make([]PermissionResponse, 0, len(permissions))
		for _, p := range permissions {
			permissionResponses = append(permissionResponses, PermissionResponse{
				ID:    p.ID,
				Code:  p.Code,
				Name:  p.Name,
				Group: p.Group,
				Desc:  p.Desc,
			})
		}

		roleResponses = append(roleResponses, RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			IsDefault:   role.IsDefault,
			Permissions: permissionResponses,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	model.Success(c, roleResponses)
}

// GetRole 获取单个角色详情
// GET /api/roles/:id
func GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "角色ID不能为空")
		return
	}

	var role model.Role
	if err := database.DB.Where("id = ?", id).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "角色不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 加载权限
	var permissions []model.Permission
	database.DB.Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ? AND rp.deleted_at IS NULL", role.ID).
		Find(&permissions)

	permissionResponses := make([]PermissionResponse, 0, len(permissions))
	for _, p := range permissions {
		permissionResponses = append(permissionResponses, PermissionResponse{
			ID:    p.ID,
			Code:  p.Code,
			Name:  p.Name,
			Group: p.Group,
			Desc:  p.Desc,
		})
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		IsDefault:   role.IsDefault,
		Permissions: permissionResponses,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	model.Success(c, response)
}

// CreateRole 创建角色
// POST /api/roles
func CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 检查角色代码是否已存在
	var existing model.Role
	if err := database.DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		model.Fail(c, http.StatusConflict, "角色代码已存在")
		return
	}

	// 开始事务
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 创建角色
		role := model.Role{
			ID:          generateID(),
			Name:        req.Name,
			Code:        req.Code,
			Description: req.Description,
			IsSystem:    false,
			IsDefault:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}

		// 关联权限
		for _, permID := range req.PermissionIDs {
			rp := model.RolePermission{
				ID:           generateID(),
				RoleID:       role.ID,
				PermissionID: permID,
				CreatedAt:    time.Now(),
			}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		model.InternalError(c, "创建失败")
		return
	}

	model.Created(c, gin.H{"message": "角色创建成功"})
}

// UpdateRole 更新角色
// PUT /api/roles/:id
func UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "角色ID不能为空")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 查找角色
	var role model.Role
	if err := database.DB.Where("id = ?", id).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "角色不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 系统预置角色不可修改（除了权限）
	if role.IsSystem {
		model.Fail(c, http.StatusForbidden, "系统预置角色不可修改")
		return
	}

	// 更新基本信息
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(&role).Updates(updates).Error; err != nil {
			model.InternalError(c, "更新失败")
			return
		}
	}

	// 更新权限关联
	if len(req.PermissionIDs) > 0 {
		// 删除旧的关联
		database.DB.Where("role_id = ?", id).Delete(&model.RolePermission{})

		// 创建新的关联
		for _, permID := range req.PermissionIDs {
			rp := model.RolePermission{
				ID:           generateID(),
				RoleID:       id,
				PermissionID: permID,
				CreatedAt:    time.Now(),
			}
			if err := database.DB.Create(&rp).Error; err != nil {
				model.InternalError(c, "权限更新失败")
				return
			}
		}
	}

	model.SuccessWithMessage(c, nil, "角色更新成功")
}

// DeleteRole 删除角色
// DELETE /api/roles/:id
func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		model.BadRequest(c, "角色ID不能为空")
		return
	}

	var role model.Role
	if err := database.DB.Where("id = ?", id).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "角色不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 系统预置角色不可删除
	if role.IsSystem {
		model.Fail(c, http.StatusForbidden, "系统预置角色不可删除")
		return
	}

	// 删除角色（软删除）
	if err := database.DB.Delete(&role).Error; err != nil {
		model.InternalError(c, "删除失败")
		return
	}

	model.SuccessWithMessage(c, nil, "角色已删除")
}

// =============================================================================
// 用户角色分配
// =============================================================================

// AssignRole 为用户分配角色
// POST /api/users/:id/roles
func AssignRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		model.BadRequest(c, "用户ID不能为空")
		return
	}

	var req struct {
		RoleCode string `json:"role_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		model.BadRequest(c, err.Error())
		return
	}

	// 查找角色
	var role model.Role
	if err := database.DB.Where("code = ?", req.RoleCode).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "角色不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	// 更新用户角色
	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("role", role.Code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "分配失败")
		return
	}

	model.SuccessWithMessage(c, nil, "角色分配成功")
}

// GetUserRole 获取用户角色
// GET /api/users/:id/role
func GetUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		model.BadRequest(c, "用户ID不能为空")
		return
	}

	var user model.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			model.NotFound(c, "用户不存在")
			return
		}
		model.InternalError(c, "查询失败")
		return
	}

	model.Success(c, gin.H{
		"role": user.Role,
	})
}

// =============================================================================
// 预置数据初始化
// =============================================================================

// InitDefaultRoles 初始化默认角色和权限
func InitDefaultRoles() {
	// 检查是否已经初始化
	var count int64
	database.DB.Model(&model.Role{}).Count(&count)
	if count > 0 {
		return
	}

	// 创建默认权限
	permissions := []model.Permission{
		{ID: generateID(), Code: "user:list", Name: "查看用户列表", Group: "用户管理", Desc: "查看所有用户列表", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "user:create", Name: "创建用户", Group: "用户管理", Desc: "创建新用户", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "user:edit", Name: "编辑用户", Group: "用户管理", Desc: "编辑用户信息", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "user:delete", Name: "删除用户", Group: "用户管理", Desc: "删除用户", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "user:status", Name: "修改用户状态", Group: "用户管理", Desc: "启用/禁用用户", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "role:list", Name: "查看角色列表", Group: "角色权限", Desc: "查看所有角色", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "role:create", Name: "创建角色", Group: "角色权限", Desc: "创建新角色", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "role:edit", Name: "编辑角色", Group: "角色权限", Desc: "编辑角色信息", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "role:delete", Name: "删除角色", Group: "角色权限", Desc: "删除角色", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "role:assign", Name: "分配角色", Group: "角色权限", Desc: "为用户分配角色", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "settings:view", Name: "查看设置", Group: "系统设置", Desc: "查看系统设置页面", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "settings:edit", Name: "编辑设置", Group: "系统设置", Desc: "修改系统设置", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "monitor:view", Name: "查看监控", Group: "监控中心", Desc: "查看监控仪表板", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "knowledge:create", Name: "创建知识库", Group: "知识库", Desc: "创建知识库", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "knowledge:edit", Name: "编辑知识库", Group: "知识库", Desc: "编辑知识库", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "knowledge:delete", Name: "删除知识库", Group: "知识库", Desc: "删除知识库", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "document:upload", Name: "上传文档", Group: "知识库", Desc: "上传文档到知识库", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "workflow:create", Name: "创建工作流", Group: "工作流", Desc: "创建工作流", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "workflow:edit", Name: "编辑工作流", Group: "工作流", Desc: "编辑工作流", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: generateID(), Code: "workflow:execute", Name: "执行工作流", Group: "工作流", Desc: "执行工作流", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, p := range permissions {
		database.DB.Create(&p)
	}

	// 创建默认角色
	adminRole := model.Role{
		ID:          generateID(),
		Name:        "系统管理员",
		Code:        "admin",
		Description: "系统管理员，拥有所有权限",
		IsSystem:    true,
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	database.DB.Create(&adminRole)

	userRole := model.Role{
		ID:          generateID(),
		Name:        "普通用户",
		Code:        "user",
		Description: "普通用户，基础功能",
		IsSystem:    true,
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	database.DB.Create(&userRole)

	// 为admin角色分配所有权限
	for _, p := range permissions {
		rp := model.RolePermission{
			ID:           generateID(),
			RoleID:       adminRole.ID,
			PermissionID: p.ID,
			CreatedAt:    time.Now(),
		}
		database.DB.Create(&rp)
	}

	// 为user角色分配基础权限
	userPermissions := []string{"settings:view", "knowledge:create", "knowledge:edit", "knowledge:delete", "document:upload", "workflow:create", "workflow:edit", "workflow:execute"}
	for _, code := range userPermissions {
		var p model.Permission
		if database.DB.Where("code = ?", code).First(&p).Error == nil {
			rp := model.RolePermission{
				ID:           generateID(),
				RoleID:       userRole.ID,
				PermissionID: p.ID,
				CreatedAt:    time.Now(),
			}
			database.DB.Create(&rp)
		}
	}

	// 将默认用户的角色设置为user
	database.DB.Model(&model.User{}).Where("role = ? OR role = ''").Update("role", "user")
}
