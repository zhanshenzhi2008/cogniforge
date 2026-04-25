package model

import (
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// RBAC Models - 用户管理与权限系统
// =============================================================================

// Permission 权限点表
type Permission struct {
	ID        string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Code      string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"` // 权限代码（如：user:create, role:edit）
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`             // 权限名称
	Group     string         `gorm:"type:varchar(100)" json:"group"`                     // 分组（如：用户管理、角色管理）
	Desc      string         `gorm:"type:text" json:"description"`                       // 描述
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Role 角色表
type Role struct {
	ID          string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`             // 角色名称
	Code        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"` // 角色代码（如：admin, user）
	Description string         `gorm:"type:text" json:"description"`                       // 描述
	IsSystem    bool           `gorm:"default:false" json:"is_system"`                     // 是否系统预置角色（不可删除）
	IsDefault   bool           `gorm:"default:false" json:"is_default"`                    // 是否默认角色
	Permissions string         `gorm:"type:text" json:"-"`                                 // 权限列表（缓存用，不直接查询）
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}

// RolePermission 角色权限关联表（多对多）
type RolePermission struct {
	ID           string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	RoleID       string         `gorm:"type:varchar(64);not null;index" json:"role_id"`
	PermissionID string         `gorm:"type:varchar(64);not null;index" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
