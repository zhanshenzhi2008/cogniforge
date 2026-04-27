package rbac

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"cogniforge/internal/database"
	"cogniforge/internal/model"
)

type RBACService struct {
	db *gorm.DB
}

func NewRBACService() *RBACService {
	return &RBACService{db: database.DB}
}

// ============ 权限管理 ============

// ListPermissions 获取权限列表
func (s *RBACService) ListPermissions() ([]model.Permission, error) {
	var permissions []model.Permission
	if err := s.db.Order("`group`, code").Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CreatePermission 创建权限
func (s *RBACService) CreatePermission(req *CreatePermissionRequest) (*model.Permission, error) {
	var existing model.Permission
	if err := s.db.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("权限代码已存在")
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

	if err := s.db.Create(&permission).Error; err != nil {
		return nil, err
	}
	return &permission, nil
}

// DeletePermission 删除权限
func (s *RBACService) DeletePermission(permissionID string) error {
	var permission model.Permission
	if err := s.db.Where("id = ?", permissionID).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("权限不存在")
		}
		return err
	}

	if err := s.db.Delete(&permission).Error; err != nil {
		return err
	}
	return nil
}

// ============ 角色管理 ============

// ListRoles 获取角色列表
func (s *RBACService) ListRoles() ([]RoleResponse, error) {
	var roles []model.Role
	if err := s.db.Order("is_system DESC, name").Find(&roles).Error; err != nil {
		return nil, err
	}

	responses := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		permissions, _ := s.getRolePermissions(role.ID)
		responses = append(responses, toRoleResponse(&role, permissions))
	}
	return responses, nil
}

// GetRole 获取单个角色
func (s *RBACService) GetRole(roleID string) (*RoleResponse, error) {
	var role model.Role
	if err := s.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, err
	}

	permissions, _ := s.getRolePermissions(role.ID)
	resp := toRoleResponse(&role, permissions)
	return &resp, nil
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(req *CreateRoleRequest) error {
	var existing model.Role
	if err := s.db.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		return fmt.Errorf("角色代码已存在")
	}

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

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}

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
}

// UpdateRole 更新角色
func (s *RBACService) UpdateRole(roleID string, req *UpdateRoleRequest) error {
	var role model.Role
	if err := s.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("角色不存在")
		}
		return err
	}

	if role.IsSystem {
		return fmt.Errorf("系统预置角色不可修改")
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&role).Updates(updates).Error; err != nil {
			return err
		}
	}

	if len(req.PermissionIDs) > 0 {
		s.db.Where("role_id = ?", roleID).Delete(&model.RolePermission{})
		for _, permID := range req.PermissionIDs {
			rp := model.RolePermission{
				ID:           generateID(),
				RoleID:       roleID,
				PermissionID: permID,
				CreatedAt:    time.Now(),
			}
			if err := s.db.Create(&rp).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteRole 删除角色
func (s *RBACService) DeleteRole(roleID string) error {
	var role model.Role
	if err := s.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("角色不存在")
		}
		return err
	}

	if role.IsSystem {
		return fmt.Errorf("系统预置角色不可删除")
	}

	if err := s.db.Delete(&role).Error; err != nil {
		return err
	}
	return nil
}

// ============ 用户角色分配 ============

// AssignRole 为用户分配角色
func (s *RBACService) AssignRole(userID, roleCode string) error {
	var role model.Role
	if err := s.db.Where("code = ?", roleCode).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("角色不存在")
		}
		return err
	}

	if err := s.db.Model(&model.User{}).Where("id = ?", userID).Update("role", role.Code).Error; err != nil {
		return fmt.Errorf("分配失败")
	}
	return nil
}

// GetUserRole 获取用户角色
func (s *RBACService) GetUserRole(userID string) (*UserRoleResponse, error) {
	var user model.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, err
	}
	return &UserRoleResponse{Role: user.Role}, nil
}

// ============ 默认数据初始化 ============

// InitDefaultRoles 初始化默认角色和权限
func (s *RBACService) InitDefaultRoles() error {
	var count int64
	s.db.Model(&model.Role{}).Count(&count)
	if count > 0 {
		return nil
	}

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
		s.db.Create(&p)
	}

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
	s.db.Create(&adminRole)

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
	s.db.Create(&userRole)

	for _, p := range permissions {
		rp := model.RolePermission{
			ID:           generateID(),
			RoleID:       adminRole.ID,
			PermissionID: p.ID,
			CreatedAt:    time.Now(),
		}
		s.db.Create(&rp)
	}

	userPermissions := []string{"settings:view", "knowledge:create", "knowledge:edit", "knowledge:delete", "document:upload", "workflow:create", "workflow:edit", "workflow:execute"}
	for _, code := range userPermissions {
		var p model.Permission
		if s.db.Where("code = ?", code).First(&p).Error == nil {
			rp := model.RolePermission{
				ID:           generateID(),
				RoleID:       userRole.ID,
				PermissionID: p.ID,
				CreatedAt:    time.Now(),
			}
			s.db.Create(&rp)
		}
	}

	s.db.Model(&model.User{}).Where("role = ? OR role = ''").Update("role", "user")
	return nil
}

// ============ 辅助函数 ============

func (s *RBACService) getRolePermissions(roleID string) ([]model.Permission, error) {
	var permissions []model.Permission
	if err := s.db.
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Where("rp.role_id = ? AND rp.deleted_at IS NULL", roleID).
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func toRoleResponse(role *model.Role, permissions []model.Permission) RoleResponse {
	permResponses := make([]PermissionResponse, 0, len(permissions))
	for _, p := range permissions {
		permResponses = append(permResponses, PermissionResponse{
			ID:    p.ID,
			Code:  p.Code,
			Name:  p.Name,
			Group: p.Group,
			Desc:  p.Desc,
		})
	}

	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		IsDefault:   role.IsDefault,
		Permissions: permResponses,
		CreatedAt:   role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   role.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// generateID 生成唯一ID
func generateID() string {
	return newID()
}
