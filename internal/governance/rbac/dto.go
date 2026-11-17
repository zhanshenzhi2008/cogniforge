package rbac

// ============ 请求结构 ============

type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required,min=2,max=100"`
	Code          string   `json:"code" binding:"required,alpha,min=2,max=50"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          string   `json:"name" binding:"min=2,max=100"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

type CreatePermissionRequest struct {
	Code  string `json:"code" binding:"required,min=2,max=100"`
	Name  string `json:"name" binding:"required,min=2,max=255"`
	Group string `json:"group"`
	Desc  string `json:"description"`
}

type AssignRoleRequest struct {
	RoleCode string `json:"role_code" binding:"required"`
}

// ============ 响应结构 ============

type PermissionResponse struct {
	ID    string `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Group string `json:"group"`
	Desc  string `json:"description"`
}

type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Code        string               `json:"code"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	IsDefault   bool                 `json:"is_default"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

type UserRoleResponse struct {
	Role string `json:"role"`
}
