# 个人设置模块详细设计文档

## 1. 模块概述

### 1.1 功能范围

本模块包含：
- ✅ 个人资料管理（头像、姓名、邮箱、联系方式）
- ✅ 密码修改（旧密码验证、强度校验）
- ✅ 头像上传（裁剪、压缩、存储）
- ✅ 安全设置（2FA 预留、会话管理）
- ✅ 用户管理（管理员：用户列表、创建、编辑、状态）
- ✅ 角色管理（RBAC：角色 CRUD、权限配置）

### 1.2 权限边界

| 功能 | 访问权限 | 说明 |
|------|---------|------|
| 查看/编辑个人资料 | 所有登录用户 | 仅限自己 |
| 修改密码 | 所有登录用户 | 需提供旧密码 |
| 上传头像 | 所有登录用户 | 文件大小限制 2MB |
| 查看会话列表 | 所有登录用户 | 仅自己的会话 |
| 远程登出 | 所有登录用户 | 仅自己的会话 |
| 用户列表页 | admin/org_admin | 管理员可见 |
| 创建/编辑用户 | admin/org_admin | 仅管理员 |
| 禁用/启用用户 | admin/org_admin | 仅管理员 |
| 角色管理 | super_admin/org_admin | 超级管理员或组织管理员 |
| 权限配置 | super_admin | 仅超级管理员 |

---

## 2. 数据库设计

### 2.1 用户表扩展（已有表，无需新增）

**表名**：`cf_users`（已有）

**新增字段**（如需）：

```sql
-- 当前表结构已包含：
-- id, organization_id, email, name, password_hash, avatar_url, phone, status, email_verified, last_login_at, created_at, updated_at, deleted_at

-- 如需扩展，可添加：
ALTER TABLE cf_users ADD COLUMN IF NOT EXISTS
    -- 时区偏好
    timezone VARCHAR(50) DEFAULT 'UTC',
    -- 语言偏好
    locale VARCHAR(10) DEFAULT 'zh-CN',
    -- 主题偏好（light/dark）
    theme VARCHAR(20) DEFAULT 'light',
    -- 邮件通知开关
    email_notifications BOOLEAN DEFAULT TRUE,
    -- 2FA 启用状态
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    -- 2FA 密钥（加密存储）
    two_factor_secret VARCHAR(255);
```

### 2.2 用户会话表（新增）

**表名**：`cf_user_sessions`

```sql
CREATE TABLE cf_user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES cf_users(id) ON DELETE CASCADE,
    session_id VARCHAR(100) NOT NULL UNIQUE,  -- JWT ID 或 Redis session ID
    ip_address INET,
    user_agent TEXT,
    device_info JSONB DEFAULT '{}',  -- {os, browser, device_type}
    location VARCHAR(100),  -- 城市/国家（可选，IP 解析）
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_session_user ON cf_user_sessions(user_id);
CREATE INDEX idx_session_expires ON cf_user_sessions(expires_at);
CREATE INDEX idx_session_active ON cf_user_sessions(last_active_at DESC);
```

**说明**：
- 用户每次登录生成一条 session 记录
- 定期清理过期 session（TTL）
- 用户可查看所有登录设备并远程登出

### 2.3 角色表（已有）

**表名**：`cf_roles`（已在数据库设计中）

```sql
-- 已有字段：
-- id, organization_id, name, description, permissions JSONB, is_system, created_at, updated_at

-- 预置角色数据（已在 migrations 中插入）：
-- 1. super_admin: 超级管理员，权限 ["*"]
-- 2. org_admin: 组织管理员，权限 ["users:*", "agents:*", "workflows:*", "billing:*"]
-- 3. developer: 开发者，权限 ["agents:*", "workflows:*", "knowledge_bases:*"]
-- 4. analyst: 分析师，权限 ["usage:read", "logs:read"]
```

### 2.4 用户角色关联表（已有）

**表名**：`cf_user_roles`

```sql
-- 已有字段：
-- id, user_id, role_id, created_at
-- 复合唯一索引：UNIQUE(user_id, role_id)
```

---

## 3. 后端 API 设计

### 3.1 个人设置 API（`internal/handler/settings.go`）

#### 3.1.1 获取个人资料

```go
// GET /api/v1/settings/profile
// Response:
{
  "code": 0,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "张三",
    "avatar_url": "https://cdn.example.com/avatars/xxx.png",
    "phone": "+86 13800138000",
    "email_verified": true,
    "timezone": "Asia/Shanghai",
    "locale": "zh-CN",
    "theme": "light"
  }
}
```

#### 3.1.2 更新个人资料

```go
// PUT /api/v1/settings/profile
// Request:
{
  "name": "李四",
  "phone": "+86 13800138001",
  "timezone": "Asia/Shanghai",
  "locale": "zh-CN"
}

// Response:
{
  "code": 0,
  "data": {
    "updated": true
  }
}
```

#### 3.1.3 上传头像

```go
// POST /api/v1/settings/avatar
// Content-Type: multipart/form-data
// Body: file (image/*, max 2MB)

// Response:
{
  "code": 0,
  "data": {
    "avatar_url": "https://cdn.example.com/avatars/user-xxx.png"
  }
}
```

**后端处理逻辑**：
```go
func (h *Handler) UploadAvatar(c *gin.Context) {
    userID := c.GetString("user_id")

    file, err := c.FormFile("file")
    if err != nil {
        model.FailBadRequest(c, "Invalid file")
        return
    }

    // 1. 验证文件大小（2MB 限制）
    if file.Size > 2*1024*1024 {
        model.FailBadRequest(c, "File too large (max 2MB)")
        return
    }

    // 2. 验证文件类型（仅图片）
    if !isImageFile(file.Filename) {
        model.FailBadRequest(c, "Only image files allowed")
        return
    }

    // 3. 打开文件
    src, err := file.Open()
    if err != nil { ... }
    defer src.Close()

    // 4. 解码图片（验证有效性 + 裁剪）
    img, err := imaging.Decode(src)
    if err != nil {
        model.FailBadRequest(c, "Invalid image format")
        return
    }

    // 5. 裁剪为圆形或正方形（200x200）
    thumb := imaging.Fill(img, 200, 200, imaging.Center, imaging.Lanczos)

    // 6. 保存到本地/对象存储
    filename := fmt.Sprintf("avatars/%s.png", userID)
    savePath := filepath.Join(h.config.AvatarUploadDir, filename)

    if err := imaging.Save(thumb, savePath); err != nil {
        model.FailInternalError(c, "Failed to save avatar")
        return
    }

    // 7. 更新用户 avatar_url
    avatarURL := fmt.Sprintf("%s/%s", h.config.AvatarBaseURL, filename)
    if err := h.userService.UpdateAvatar(userID, avatarURL); err != nil {
        model.FailInternalError(c, err.Error())
        return
    }

    model.Success(c, gin.H{"avatar_url": avatarURL})
}
```

#### 3.1.4 修改密码

```go
// POST /api/v1/settings/password
// Request:
{
  "current_password": "旧密码",
  "new_password": "新密码",
  "confirm_password": "确认密码"
}

// Response:
{
  "code": 0,
  "message": "Password updated successfully"
}
```

**逻辑**：
```go
func (h *Handler) ChangePassword(c *gin.Context) {
    userID := c.GetString("user_id")
    var req ChangePasswordRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        model.FailBadRequest(c, err.Error())
        return
    }

    // 1. 验证新密码一致性
    if req.NewPassword != req.ConfirmPassword {
        model.FailBadRequest(c, "Passwords do not match")
        return
    }

    // 2. 验证新密码强度（最小长度、复杂度）
    if err := validatePasswordStrength(req.NewPassword); err != nil {
        model.FailBadRequest(c, err.Error())
        return
    }

    // 3. 验证旧密码
    user, err := h.userService.GetByID(userID)
    if err != nil { ... }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
        model.FailUnauthorized(c, "Current password is incorrect")
        return
    }

    // 4. 哈希新密码
    newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
    if err != nil { ... }

    // 5. 更新数据库
    if err := h.userService.UpdatePassword(userID, string(newHash)); err != nil {
        model.FailInternalError(c, err.Error())
        return
    }

    // 6. 使当前 Token 失效（可选：强制重新登录）
    h.tokenService.RevokeUserTokens(userID)

    model.SuccessWithMessage(c, gin.H{"message": "Password updated"}, "密码修改成功")
}
```

#### 3.1.5 获取会话列表

```go
// GET /api/v1/settings/sessions
// Response:
{
  "code": 0,
  "data": {
    "sessions": [
      {
        "id": "session-uuid",
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "device_info": {"os": "macOS", "browser": "Chrome", "device_type": "desktop"},
        "location": "上海, 中国",
        "last_active_at": "2026-04-11T10:30:00Z",
        "is_current": true,
        "created_at": "2026-04-10T08:00:00Z"
      }
    ]
  }
}
```

#### 3.1.6 远程登出会话

```go
// DELETE /api/v1/settings/sessions/:sessionId
// Response:
{
  "code": 0,
  "message": "Session revoked"
}
```

### 3.2 用户管理 API（`internal/handler/user.go`）

#### 3.2.1 用户列表（分页、筛选）

```go
// GET /api/v1/admin/users
// Query: ?page=1&limit=20&status=active&search=张三&role=developer

// Response:
{
  "code": 0,
  "data": {
    "users": [
      {
        "id": "uuid",
        "email": "user@example.com",
        "name": "张三",
        "avatar_url": "...",
        "status": "active",  // active / disabled
        "role": "developer",
        "email_verified": true,
        "last_login_at": "2026-04-11T...",
        "created_at": "2026-04-01T..."
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 20
  }
}
```

#### 3.2.2 创建用户（管理员）

```go
// POST /api/v1/admin/users
// Request:
{
  "email": "newuser@example.com",
  "password": "TempPass123!",
  "name": "新用户",
  "role": "developer",  // 角色名称
  "organization_id": "org-uuid"  // 可选，默认当前组织
}

// Response:
{
  "code": 0,
  "data": {
    "id": "new-uuid",
    "email": "newuser@example.com",
    "name": "新用户",
    "status": "active"
  }
}
```

#### 3.2.3 编辑用户

```go
// PUT /api/v1/admin/users/:id
// Request:
{
  "name": "修改后的姓名",
  "phone": "+86 13800138000",
  "role": "analyst",
  "status": "active"  // 可启用/禁用
}

// Response:
{
  "code": 0,
  "data": {
    "updated": true
  }
}
```

#### 3.2.4 删除用户（软删除）

```go
// DELETE /api/v1/admin/users/:id
// Response:
{
  "code": 0,
  "message": "User deleted"
}
```

### 3.3 角色管理 API（`internal/handler/roles.go`）

#### 3.3.1 角色列表

```go
// GET /api/v1/admin/roles
// Response:
{
  "code": 0,
  "data": {
    "roles": [
      {
        "id": "role-uuid",
        "name": "developer",
        "description": "开发者角色",
        "permissions": ["agents:*", "workflows:*", "knowledge_bases:*"],
        "is_system": true,  // 系统预置角色不可删除
        "created_at": "..."
      }
    ]
  }
}
```

#### 3.3.2 创建角色

```go
// POST /api/v1/admin/roles
// Request:
{
  "name": "custom_role",
  "description": "自定义角色",
  "permissions": ["agents:read", "agents:write", "workflows:read"]
}

// Response:
{
  "code": 0,
  "data": {
    "id": "new-role-uuid",
    "name": "custom_role"
  }
}
```

#### 3.3.3 更新角色权限

```go
// PUT /api/v1/admin/roles/:id/permissions
// Request:
{
  "permissions": ["agents:*", "workflows:*"]
}

// Response:
{
  "code": 0,
  "data": {
    "updated": true
  }
}
```

#### 3.3.4 删除角色

```go
// DELETE /api/v1/admin/roles/:id
// 仅可删除自定义角色（is_system=false）

// Response:
{
  "code": 0,
  "message": "Role deleted"
}
```

### 3.4 权限中间件（`internal/middleware/rbac.go`）

```go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
)

// RequirePermission 中间件：检查用户是否拥有指定权限
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            model.FailUnauthorized(c, "Unauthorized")
            c.Abort()
            return
        }

        // 1. 从数据库查询用户角色和权限
        permissions, err := userService.GetUserPermissions(userID)
        if err != nil {
            model.FailInternalError(c, err.Error())
            c.Abort()
            return
        }

        // 2. 检查权限（支持通配符 *）
        if !hasPermission(permissions, permission) {
            model.FailForbidden(c, "Insufficient permissions")
            c.Abort()
            return
        }

        c.Next()
    }
}

// hasPermission 检查用户权限列表是否包含目标权限
// 支持：agents:read, agents:write, agents:*（通配符）
func hasPermission(userPermissions []string, required string) bool {
    for _, p := range userPermissions {
        // 精确匹配
        if p == required {
            return true
        }
        // 通配符匹配：agents:* 匹配 agents:read/agents:write
        if strings.HasSuffix(p, ":*") {
            prefix := strings.TrimSuffix(p, ":*")
            reqPrefix := strings.SplitN(required, ":", 2)[0]
            if prefix == reqPrefix {
                return true
            }
        }
    }
    return false
}
```

**使用示例**：

```go
// 路由注册
r.GET("/api/v1/admin/users", middleware.RequirePermission("users:read"), handler.ListUsers)
r.POST("/api/v1/admin/users", middleware.RequirePermission("users:write"), handler.CreateUser)
r.DELETE("/api/v1/admin/users/:id", middleware.RequirePermission("users:delete"), handler.DeleteUser)

r.GET("/api/v1/agents", middleware.RequirePermission("agents:read"), handler.ListAgents)
r.POST("/api/v1/agents", middleware.RequirePermission("agents:write"), handler.CreateAgent)
```

---

## 4. 前端设计

### 4.1 视觉规范

#### 4.1.1 通用设置模块样式规范

所有设置页面（个人资料、偏好设置、登录会话、安全设置）采用统一的视觉规范，确保体验一致性：

```css
/* 页面容器 */
.section-container {
  animation: fadeIn 0.3s ease;
}

/* 标题区域 */
.section-header {
  margin-bottom: 16px;
}

.section-title {
  font-size: 18px;        /* 标题字号 */
  font-weight: 600;       /* 标题字重 */
  color: #0f172a;         /* 标题颜色 */
  margin: 0 0 4px 0;     /* 标题下边距 */
}

.section-desc {
  font-size: 13px;        /* 描述字号 */
  color: #64748b;         /* 描述颜色 */
  margin: 0;
}

/* 内容卡片 */
.content-card {
  background: #ffffff;
  border-radius: 10px;    /* 圆角 10px */
  border: 1px solid #e2e8f0;
  overflow: hidden;
  margin-bottom: 12px;    /* 卡片间距 12px */
}

/* 卡片头部 */
.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
}

.card-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-icon :deep(.n-icon) {
  font-size: 18px;
  color: #ffffff;
}

.card-title-area {
  flex: 1;
}

.card-title-area h3 {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.card-title-area p {
  font-size: 12px;
  color: #64748b;
  margin: 2px 0 0 0;
}

/* 表单样式 */
.form-section {
  padding: 16px;
}

.form-row {
  margin-bottom: 12px;
}

.form-row:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 6px;
}

/* 输入控件 - 使用 small 尺寸 */
.form-section :deep(.n-input),
.form-section :deep(.n-select) {
  --n-height: 32px;  /* 统一高度 */
}

/* 按钮样式 */
.form-actions {
  display: flex;
  justify-content: flex-end;  /* 按钮右对齐 */
  gap: 8px;
  padding: 12px 16px;
  background: #f8fafc;
  border-top: 1px solid #f1f5f9;
}

.form-actions :deep(.n-button) {
  font-size: 13px;  /* 按钮字体 */
}
```

#### 4.1.2 间距系统

| 元素 | 间距值 |
|------|--------|
| 标题与内容间距 | 16px |
| 卡片之间间距 | 12px |
| 表单项间距 | 12px |
| 标签与输入框间距 | 6px |
| 按钮间距 | 8px |
| 卡片内边距 | 16px |
| 卡片头部内边距 | 14px 16px |
| 卡片底部内边距 | 12px 16px |

#### 4.1.3 字体规范

| 元素 | 字号 | 字重 | 颜色 |
|------|------|------|------|
| 页面标题 | 18px | 600 | #0f172a |
| 页面描述 | 13px | 400 | #64748b |
| 卡片标题 | 14px | 600 | #0f172a |
| 卡片描述 | 12px | 400 | #64748b |
| 表单标签 | 13px | 500 | #374151 |
| 提示文字 | 11-12px | 400 | #94a3b8 |
| 按钮文字 | 13px | 500 | - |

#### 4.1.4 卡片图标渐变色规范

```css
/* 密码相关 - 靛蓝色 */
.password-icon {
  background: linear-gradient(135deg, #6366f1, #4f46e5);
}

/* 2FA相关 - 琥珀色 */
.tfa-icon {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

/* 会话相关 - 绿色 */
.sessions-icon {
  background: linear-gradient(135deg, #10b981, #059669);
}

/* 当前设备 - 绿色 */
.current-icon {
  background: linear-gradient(135deg, #10b981, #059669);
}

/* 个人资料 - 蓝色 */
.profile-icon {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}

/* 偏好设置 - 紫色 */
.preferences-icon {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
}
```

#### 4.1.5 危险操作区域样式

```css
.danger-zone {
  border-color: #fecaca;
  background: #fef2f2;
}

.danger-header {
  padding: 14px 16px;
  border-bottom: 1px solid #fecaca;
}

.danger-header h4 {
  font-size: 13px;
  font-weight: 600;
  color: #991b1b;
  margin: 0 0 2px 0;
}

.danger-header p {
  font-size: 12px;
  color: #b91c1c;
  margin: 0;
}
```

### 4.2 路由结构

```typescript
// pages/settings/
├── index.vue            # 重定向到 /settings/profile
├── profile.vue          # 个人资料
├── security.vue         # 安全设置（密码、2FA）
└── sessions.vue         # 会话管理

// pages/admin/（管理员可见）
├── users.vue            # 用户管理
├── users/[id].vue       # 用户编辑
├── roles.vue            # 角色管理
└── roles/[id].vue       # 角色编辑
```

### 4.3 状态管理（Pinia）

```typescript
// stores/settings.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useApi } from '@/composables/useApi'

export const useSettingsStore = defineStore('settings', () => {
  const api = useApi()

  // State
  const profile = ref<UserProfile | null>(null)
  const sessions = ref<Session[]>([])
  const loading = ref(false)

  // Getters
  const currentUser = computed(() => profile.value)

  // Actions
  async function fetchProfile() {
    loading.value = true
    try {
      const res = await api.get<UserProfile>('/api/v1/settings/profile')
      if (res.error) throw new Error(res.error)
      profile.value = res.data!
    } catch (error) {
      console.error('Failed to fetch profile:', error)
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(data: Partial<UserProfile>) {
    loading.value = true
    try {
      const res = await api.put<UserProfile>('/api/v1/settings/profile', data)
      if (res.error) throw new Error(res.error)
      profile.value = { ...profile.value!, ...res.data }
      return { success: true }
    } catch (error) {
      return { error: error.message }
    } finally {
      loading.value = false
    }
  }

  async function uploadAvatar(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const res = await api.post<{ avatar_url: string }>('/api/v1/settings/avatar', formData)
    if (!res.error) {
      profile.value!.avatar_url = res.data!.avatar_url
    }
    return res
  }

  async function changePassword(current: string, newPassword: string) {
    const res = await api.post('/api/v1/settings/password', {
      current_password: current,
      new_password: newPassword,
      confirm_password: newPassword
    })
    return res
  }

  async function fetchSessions() {
    const res = await api.get<Session[]>('/api/v1/settings/sessions')
    if (!res.error) {
      sessions.value = res.data!
    }
    return res
  }

  async function revokeSession(sessionId: string) {
    const res = await api.delete(`/api/v1/settings/sessions/${sessionId}`)
    if (!res.error) {
      sessions.value = sessions.value.filter(s => s.id !== sessionId)
    }
    return res
  }

  return {
    profile,
    sessions,
    loading,
    currentUser,
    fetchProfile,
    updateProfile,
    uploadAvatar,
    changePassword,
    fetchSessions,
    revokeSession,
  }
})
```

### 4.4 个人资料页面（`pages/settings/profile.vue`）

```vue
<template>
  <div class="settings-page">
    <n-card title="个人资料" class="profile-card">
      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-placement="left"
        label-width="100px"
      >
        <!-- 头像上传 -->
        <n-form-item label="头像">
          <div class="avatar-upload">
            <n-avatar
              :size="80"
              :src="form.avatar_url"
              fallback-src="https://cdn.example.com/avatars/default.png"
            />
            <n-upload
              :custom-request="handleAvatarUpload"
              :show-file-list="false"
              accept="image/*"
            >
              <n-button size="small" type="primary" class="upload-btn">
                更换头像
              </n-button>
            </n-upload>
            <p class="hint">支持 JPG、PNG，最大 2MB</p>
          </div>
        </n-form-item>

        <!-- 姓名 -->
        <n-form-item label="姓名" path="name">
          <n-input v-model:value="form.name" placeholder="请输入姓名" />
        </n-form-item>

        <!-- 邮箱（只读） -->
        <n-form-item label="邮箱">
          <n-input :value="form.email" disabled />
        </n-form-item>

        <!-- 手机号 -->
        <n-form-item label="手机号" path="phone">
          <n-input v-model:value="form.phone" placeholder="+86 13800138000" />
        </n-form-item>

        <!-- 时区 -->
        <n-form-item label="时区" path="timezone">
          <n-select v-model:value="form.timezone" :options="timezoneOptions" />
        </n-form-item>

        <!-- 语言 -->
        <n-form-item label="语言" path="locale">
          <n-select v-model:value="form.locale" :options="localeOptions" />
        </n-form-item>

        <!-- 主题 -->
        <n-form-item label="主题" path="theme">
          <n-radio-group v-model:value="form.theme">
            <n-radio-button value="light">浅色</n-radio-button>
            <n-radio-button value="dark">深色</n-radio-button>
            <n-radio-button value="auto">跟随系统</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <!-- 邮件通知 -->
        <n-form-item label="邮件通知">
          <n-switch v-model:value="form.email_notifications" />
        </n-form-item>

        <!-- 保存按钮 -->
        <n-form-item>
          <n-space>
            <n-button type="primary" :loading="loading" @click="handleSave">
              保存修改
            </n-button>
            <n-button @click="resetForm">重置</n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useMessage, useDialog } from 'naive-ui'
import { useSettingsStore } from '@/stores/settings'

const message = useMessage()
const dialog = useDialog()
const settingsStore = useSettingsStore()

const formRef = ref()
const loading = ref(false)

// 表单数据
const form = reactive({
  id: '',
  email: '',
  name: '',
  avatar_url: '',
  phone: '',
  timezone: 'Asia/Shanghai',
  locale: 'zh-CN',
  theme: 'light' as 'light' | 'dark' | 'auto',
  email_notifications: true,
})

// 校验规则
const rules = {
  name: [
    { required: true, message: '姓名不能为空', trigger: 'blur' },
    { max: 50, message: '姓名不能超过50个字符', trigger: 'blur' },
  ],
  phone: [
    { pattern: /^\+?[1-9]\d{1,14}$/, message: '手机号格式不正确', trigger: 'blur' },
  ],
}

// 时区选项
const timezoneOptions = [
  { label: '中国标准时间 (UTC+8)', value: 'Asia/Shanghai' },
  { label: '美国东部时间 (UTC-5)', value: 'America/New_York' },
  { label: '欧洲中部时间 (UTC+1)', value: 'Europe/Berlin' },
  // ... 更多时区
]

// 语言选项
const localeOptions = [
  { label: '简体中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' },
]

// 头像上传处理
const handleAvatarUpload = async (options: { file: File }) => {
  const { file } = options

  // 验证文件大小
  if (file.size > 2 * 1024 * 1024) {
    message.error('文件大小不能超过 2MB')
    return
  }

  // 验证文件类型
  if (!file.type.startsWith('image/')) {
    message.error('只支持图片文件')
    return
  }

  loading.value = true
  try {
    const result = await settingsStore.uploadAvatar(file)
    if (result.error) {
      message.error(result.error)
    } else {
      message.success('头像上传成功')
    }
  } finally {
    loading.value = false
  }
}

// 保存表单
const handleSave = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    const result = await settingsStore.updateProfile({
      name: form.name,
      phone: form.phone,
      timezone: form.timezone,
      locale: form.locale,
      theme: form.theme,
      email_notifications: form.email_notifications,
    })

    if (result.error) {
      message.error(result.error)
    } else {
      message.success('保存成功')
    }
  } catch (error) {
    // 表单验证失败
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  settingsStore.fetchProfile()
}

// 初始化
onMounted(() => {
  settingsStore.fetchProfile().then(() => {
    if (settingsStore.profile) {
      Object.assign(form, settingsStore.profile)
    }
  })
})
</script>

<style scoped>
.profile-card {
  max-width: 600px;
  margin: 0 auto;
}

.avatar-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.upload-btn {
  margin-top: 8px;
}

.hint {
  font-size: 12px;
  color: var(--n-text-color-3);
  margin: 0;
}
</style>
```

### 4.5 安全设置页面（`pages/settings/security.vue`）

```vue
<template>
  <div class="settings-page">
    <!-- 修改密码卡片 -->
    <n-card title="修改密码" class="security-card">
      <n-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-placement="left"
        label-width="120px"
      >
        <n-form-item label="当前密码" path="current_password">
          <n-input
            v-model:value="passwordForm.current_password"
            type="password"
            show-password-on="click"
            placeholder="请输入当前密码"
          />
        </n-form-item>

        <n-form-item label="新密码" path="new_password">
          <n-input
            v-model:value="passwordForm.new_password"
            type="password"
            show-password-on="click"
            placeholder="请输入新密码"
          />
          <template #help>
            <PasswordStrength :password="passwordForm.new_password" />
          </template>
        </n-form-item>

        <n-form-item label="确认新密码" path="confirm_password">
          <n-input
            v-model:value="passwordForm.confirm_password"
            type="password"
            show-password-on="click"
            placeholder="请再次输入新密码"
          />
        </n-form-item>

        <n-form-item>
          <n-button
            type="primary"
            :loading="loading"
            @click="handleChangePassword"
          >
            修改密码
          </n-button>
        </n-form-item>
      </n-form>
    </n-card>

    <!-- 登录会话卡片 -->
    <n-card title="登录会话" class="sessions-card">
      <n-data-table
        :columns="sessionColumns"
        :data="sessions"
        :pagination="false"
        :bordered="false"
      >
        <template #device-type="{ row }">
          <n-icon :size="16" :component="getDeviceIcon(row.device_info?.device_type)" />
          {{ getDeviceLabel(row.device_info?.device_type) }}
        </template>

        <template #is-current="{ row }">
          <n-tag v-if="row.is_current" type="success" size="small">
            当前设备
          </n-tag>
        </template>

        <template #actions="{ row }">
          <n-button
            v-if="!row.is_current"
            size="tiny"
            type="error"
            @click="handleRevokeSession(row.id)"
          >
            登出
          </n-button>
        </template>
      </n-data-table>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useMessage } from 'naive-ui'
import { useSettingsStore } from '@/stores/settings'
import PasswordStrength from '@/components/common/PasswordStrength.vue'
import {
  DesktopOutline,
  MobileOutline,
  TabletOutline,
  WarningOutline
} from '@vicons/ionicons5'

const message = useMessage()
const settingsStore = useSettingsStore()

const passwordFormRef = ref()
const loading = ref(false)

const passwordForm = reactive({
  current_password: '',
  new_password: '',
  confirm_password: '',
})

const passwordRules = {
  current_password: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码至少8个字符', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        // 至少包含大小写字母和数字
        const hasUpper = /[A-Z]/.test(value)
        const hasLower = /[a-z]/.test(value)
        const hasDigit = /\d/.test(value)
        return hasUpper && hasLower && hasDigit
      },
      message: '密码需包含大小写字母和数字',
      trigger: 'blur'
    },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string) => {
        return value === passwordForm.new_password
      },
      message: '两次密码输入不一致',
      trigger: 'blur'
    },
  ],
}

// 会话表格列定义
const sessionColumns = [
  { title: '设备类型', key: 'device-type', width: 120 },
  { title: 'IP 地址', key: 'ip_address', width: 150 },
  { title: '位置', key: 'location', width: 150 },
  { title: '最后活动', key: 'last_active_at', width: 180 },
  { title: '状态', key: 'is_current', width: 100 },
  { title: '操作', key: 'actions', width: 100 },
]

const sessions = ref<Session[]>([])

// 获取设备图标
const getDeviceIcon = (type?: string) => {
  switch (type) {
    case 'mobile': return MobileOutline
    case 'tablet': return TabletOutline
    default: return DesktopOutline
  }
}

const getDeviceLabel = (type?: string) => {
  switch (type) {
    case 'mobile': return '手机'
    case 'tablet': return '平板'
    case 'desktop': return '桌面'
    default: return '未知设备'
  }
}

// 修改密码
const handleChangePassword = async () => {
  try {
    await passwordFormRef.value?.validate()
    loading.value = true

    const res = await settingsStore.changePassword(
      passwordForm.current_password,
      passwordForm.new_password
    )

    if (res.error) {
      message.error(res.error)
    } else {
      message.success('密码修改成功，请重新登录')
      // 清空表单
      passwordForm.current_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
      // 跳转到登录页
      setTimeout(() => navigateTo('/login'), 2000)
    }
  } catch (error) {
    // 验证失败
  } finally {
    loading.value = false
  }
}

// 远程登出会话
const handleRevokeSession = async (sessionId: string) => {
  dialog.warning({
    title: '确认登出',
    content: '确定要登出该设备吗？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await settingsStore.revokeSession(sessionId)
      if (res.error) {
        message.error(res.error)
      } else {
        message.success('已登出该设备')
      }
    },
  })
}

onMounted(async () => {
  await settingsStore.fetchSessions()
  sessions.value = settingsStore.sessions
})
</script>

<style scoped>
.settings-page {
  max-width: 800px;
  margin: 0 auto;
}

.security-card,
.sessions-card {
  margin-bottom: 24px;
}
</style>
```

### 4.6 密码强度组件（`components/common/PasswordStrength.vue`）

```vue
<template>
  <div class="password-strength">
    <div class="strength-bar">
      <div
        class="bar-fill"
        :class="strengthClass"
        :style="{ width: strengthPercent + '%' }"
      />
    </div>
    <div class="strength-label" :class="strengthClass">
      {{ strengthLabel }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  password: string
}>()

// 计算强度（0-100）
const strengthScore = computed(() => {
  let score = 0
  const p = props.password

  if (!p) return 0

  // 长度得分
  if (p.length >= 8) score += 25
  if (p.length >= 12) score += 10

  // 包含小写字母
  if (/[a-z]/.test(p)) score += 15

  // 包含大写字母
  if (/[A-Z]/.test(p)) score += 15

  // 包含数字
  if (/\d/.test(p)) score += 15

  // 包含特殊字符
  if (/[^A-Za-z0-9]/.test(p)) score += 20

  return Math.min(score, 100)
})

const strengthPercent = computed(() => strengthScore.value)

const strengthClass = computed(() => {
  if (strengthScore.value < 40) return 'weak'
  if (strengthScore.value < 70) return 'medium'
  if (strengthScore.value < 90) return 'strong'
  return 'very-strong'
})

const strengthLabel = computed(() => {
  if (strengthScore.value < 40) return '弱'
  if (strengthScore.value < 70) return '中等'
  if (strengthScore.value < 90) return '强'
  return '非常强'
})
</script>

<style scoped>
.password-strength {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

.strength-bar {
  flex: 1;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.bar-fill.weak {
  background-color: #ef4444;
}

.bar-fill.medium {
  background-color: #f59e0b;
}

.bar-fill.strong {
  background-color: #22c55e;
}

.bar-fill.very-strong {
  background-color: #3b82f6;
}

.strength-label {
  font-size: 12px;
  min-width: 48px;
  text-align: right;
}

.strength-label.weak {
  color: #ef4444;
}

.strength-label.medium {
  color: #f59e0b;
}

.strength-label.strong {
  color: #22c55e;
}

.strength-label.very-strong {
  color: #3b82f6;
}
</style>
```

### 4.7 用户管理页面（`pages/admin/users.vue`）

```vue
<template>
  <div class="admin-users-page">
    <n-card title="用户管理">
      <!-- 搜索栏 -->
      <n-space class="mb-4" vertical>
        <n-space>
          <n-input
            v-model:value="searchQuery"
            placeholder="搜索姓名/邮箱"
            clearable
            style="width: 300px"
            @update:value="handleSearch"
          />
          <n-select
            v-model:value="filterStatus"
            :options="statusOptions"
            placeholder="状态筛选"
            clearable
            style="width: 150px"
            @update:value="handleSearch"
          />
          <n-select
            v-model:value="filterRole"
            :options="roleOptions"
            placeholder="角色筛选"
            clearable
            style="width: 150px"
            @update:value="handleSearch"
          />
          <n-button type="primary" @click="handleCreate">
            <template #icon>
              <n-icon :component="AddOutline" />
            </template>
            创建用户
          </n-button>
        </n-space>
      </n-space>

      <!-- 用户表格 -->
      <n-data-table
        :columns="columns"
        :data="users"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: User) => row.id"
        @update:page="handlePageChange"
      />
    </n-card>

    <!-- 创建/编辑用户模态框 -->
    <n-modal
      v-model:show="modalVisible"
      preset="card"
      :title="isEditing ? '编辑用户' : '创建用户'"
      style="width: 500px"
      :segmented="{ content: true, footer: true }"
    >
      <n-form
        ref="userFormRef"
        :model="userForm"
        :rules="userRules"
        label-placement="left"
        label-width="100px"
      >
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="userForm.email" :disabled="isEditing" placeholder="user@example.com" />
        </n-form-item>

        <n-form-item label="姓名" path="name">
          <n-input v-model:value="userForm.name" placeholder="张三" />
        </n-form-item>

        <n-form-item v-if="!isEditing" label="密码" path="password">
          <n-input
            v-model:value="userForm.password"
            type="password"
            show-password-on="click"
            placeholder="初始密码（至少8位）"
          />
        </n-form-item>

        <n-form-item label="角色" path="role">
          <n-select v-model:value="userForm.role" :options="availableRoles" placeholder="选择角色" />
        </n-form-item>

        <n-form-item label="状态">
          <n-switch v-model:value="userForm.status" :checked-value="'active'" :unchecked-value="'disabled'">
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </n-switch>
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="modalVisible = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEditing ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useMessage, NIcon } from 'naive-ui'
import { AddOutline, EditOutline, DeleteOutline } from '@vicons/ionicons5'
import { useAdminStore } from '@/stores/admin'

const message = useMessage()
const adminStore = useAdminStore()

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const isEditing = ref(false)
const userFormRef = ref()

const searchQuery = ref('')
const filterStatus = ref<string | null>(null)
const filterRole = ref<string | null>(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  itemCount: 0,
})

const users = ref<User[]>([])
const roles = ref<Role[]>([])
const currentUserId = ref<string | null>(null)

const userForm = reactive({
  id: '',
  email: '',
  name: '',
  password: '',
  role: '',
  status: 'active' as 'active' | 'disabled',
})

const userRules = {
  email: [
    { required: true, message: '邮箱不能为空', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  name: [
    { required: true, message: '姓名不能为空', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '密码不能为空', trigger: 'blur' },
    { min: 8, message: '密码至少8个字符', trigger: 'blur' },
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' },
  ],
}

const columns = [
  { title: '姓名', key: 'name', width: 150 },
  { title: '邮箱', key: 'email', width: 200 },
  {
    title: '角色',
    key: 'role',
    width: 120,
    render(row: User) {
      const role = roles.value.find(r => r.name === row.role)
      return h(NIcon, null, {
        default: () => h('span', {
          style: { color: role?.color || '#666' }
        }, role?.label || row.role)
      })
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render(row: User) {
      const type = row.status === 'active' ? 'success' : 'error'
      const label = row.status === 'active' ? '启用' : '禁用'
      return h('n-tag', { type, size: 'small' }, { default: () => label })
    }
  },
  { title: '最后登录', key: 'last_login_at', width: 180 },
  { title: '创建时间', key: 'created_at', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render(row: User) {
      return h('n-space', {}, {
        default: () => [
          h('n-button', {
            size: 'tiny',
            type: 'primary',
            onClick: () => handleEdit(row)
          }, { default: () => '编辑' }),
          h('n-button', {
            size: 'tiny',
            type: 'error',
            onClick: () => handleDelete(row)
          }, { default: () => '删除' }),
        ]
      })
    }
  },
]

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '禁用', value: 'disabled' },
]

const roleOptions = computed(() =>
  roles.value.map(r => ({ label: r.label, value: r.name }))
)

const availableRoles = computed(() =>
  roles.value.map(r => ({
    label: r.label,
    value: r.name,
    disabled: r.is_system  // 系统角色不可编辑
  }))
)

// 加载数据
onMounted(async () => {
  loading.value = true
  try {
    await Promise.all([
      fetchUsers(),
      fetchRoles(),
    ])
  } finally {
    loading.value = false
  }
})

const fetchUsers = async () => {
  const res = await adminStore.fetchUsers({
    page: pagination.page,
    limit: pagination.pageSize,
    search: searchQuery.value || undefined,
    status: filterStatus.value || undefined,
    role: filterRole.value || undefined,
  })

  if (!res.error) {
    users.value = res.data!.users
    pagination.itemCount = res.data!.total
  }
}

const fetchRoles = async () => {
  const res = await adminStore.fetchRoles()
  if (!res.error) {
    roles.value = res.data!.roles
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchUsers()
}

// 分页
const handlePageChange = (page: number) => {
  pagination.page = page
  fetchUsers()
}

// 创建用户
const handleCreate = () => {
  isEditing.value = false
  currentUserId.value = null
  Object.assign(userForm, {
    id: '',
    email: '',
    name: '',
    password: '',
    role: '',
    status: 'active',
  })
  modalVisible.value = true
}

// 编辑用户
const handleEdit = (user: User) => {
  isEditing.value = true
  currentUserId.value = user.id
  Object.assign(userForm, {
    id: user.id,
    email: user.email,
    name: user.name,
    password: '',
    role: user.role,
    status: user.status,
  })
  modalVisible.value = true
}

// 删除用户
const handleDelete = (user: User) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除用户 "${user.name}" 吗？此操作不可撤销。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await adminStore.deleteUser(user.id)
      if (res.error) {
        message.error(res.error)
      } else {
        message.success('删除成功')
        fetchUsers()
      }
    },
  })
}

// 提交表单
const handleSubmit = async () => {
  try {
    await userFormRef.value?.validate()
    submitting.value = true

    let res
    if (isEditing.value) {
      res = await adminStore.updateUser(currentUserId.value!, {
        name: userForm.name,
        role: userForm.role,
        status: userForm.status,
      })
    } else {
      res = await adminStore.createUser({
        email: userForm.email,
        name: userForm.name,
        password: userForm.password,
        role: userForm.role,
      })
    }

    if (res.error) {
      message.error(res.error)
    } else {
      message.success(isEditing.value ? '更新成功' : '创建成功')
      modalVisible.value = false
      fetchUsers()
    }
  } catch (error) {
    // 表单验证失败
  } finally {
    submitting.value = false
  }
}
</script>
```

---

## 5. 安全考虑

### 5.1 密码策略

| 规则 | 说明 |
|------|------|
| 最小长度 | 8 字符 |
| 复杂度要求 | 至少包含大小写字母和数字 |
| 密码历史 | 禁止使用最近 5 次密码 |
| 锁定策略 | 连续 5 次失败，锁定 30 分钟 |
| 过期时间 | 建议 90 天强制修改（可选） |

### 5.2 会话安全

- Session ID 随机生成（UUID v4）
- Session 过期时间：7天（记住我）或 24小时
- 每次登录生成新 session，旧 session 保留（可查看）
- 密码修改后，使所有 session 失效（强制重新登录）

### 5.3 头像上传安全

- 文件大小限制：2MB
- 文件类型限制：`image/jpeg`, `image/png`, `image/gif`
- 文件名随机化：使用 UUID 命名，避免路径遍历
- 病毒扫描（可选）：集成 ClamAV
- CDN 存储：上传到 MinIO/S3，而非本地文件系统

### 5.4 权限校验

- **后端**：每个 API 都必须通过 `RequirePermission` 中间件
- **前端**：自定义指令 `v-permission="'agents:write'"` 控制按钮显示
- **数据库层**：Row-Level Security（可选，PostgreSQL RLS）

---

## 6. 前端权限指令

```typescript
// plugins/permission.ts
import { App, Directive } from 'vue'

const permission: Directive = {
  mounted(el: HTMLElement, binding: any) {
    const requiredPermission = binding.value
    const userPermissions = useAuthStore().permissions  // 从 Pinia 获取

    if (!hasPermission(userPermissions, requiredPermission)) {
      el.parentNode?.removeChild(el)  // 移除 DOM
    }
  },
}

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.directive('permission', permission)
})
```

**使用示例**：

```vue
<template>
  <n-button
    v-permission="'agents:write'"
    type="primary"
    @click="createAgent"
  >
    创建 Agent
  </n-button>
</template>
```

---

## 7. 测试计划

### 7.1 后端单元测试

```go
// internal/handler/settings_test.go
func TestChangePassword(t *testing.T) {
    tests := []struct {
        name           string
        currentPass    string
        newPass        string
        confirmPass   string
        wantErr        bool
        expectedErrMsg string
    }{
        {"正确修改密码", "old123", "NewPass123!", "NewPass123!", false, ""},
        {"旧密码错误", "wrong", "NewPass123!", "NewPass123!", true, "当前密码不正确"},
        {"密码不一致", "old123", "NewPass123!", "Different123!", true, "两次密码输入不一致"},
        {"密码太短", "old123", "abc", "abc", true, "至少8个字符"},
        {"缺少大写字母", "old123", "abcdefg1!", "abcdefg1!", true, "大小写字母和数字"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := ChangePasswordRequest{
                CurrentPassword: tt.currentPass,
                NewPassword:     tt.newPass,
                ConfirmPassword: tt.confirmPass,
            }

            err := validateChangePassword(req)
            if (err != nil) != tt.wantErr {
                t.Errorf("expected error %v, got %v", tt.wantErr, err)
            }
        })
    }
}
```

### 7.2 前端组件测试

```typescript
// __tests__/components/ProfileForm.spec.ts
import { mount } from '@vue/test-utils'
import ProfileForm from '@/pages/settings/profile.vue'

describe('ProfileForm', () => {
  it('renders correctly', () => {
    const wrapper = mount(ProfileForm)
    expect(wrapper.find('n-input[placeholder="请输入姓名"]').exists()).toBe(true)
  })

  it('validates required fields', async () => {
    const wrapper = mount(ProfileForm)
    const submitBtn = wrapper.find('n-button[type="primary"]')

    await submitBtn.trigger('click')

    // 应该显示验证错误
    expect(wrapper.find('.n-form-item--error').exists()).toBe(true)
  })

  it('submits form successfully', async () => {
    // Mock API
    const mockApi = vi.fn().mockResolvedValue({ data: { ... } })
    // ...
  })
})
```

---

## 8. 接口文档

### 8.1 Swagger/OpenAPI 定义（Go 端）

```go
// internal/handler/settings.go
//go:generate swag generate -g handler.go -t ./docs -f ./internal/handler/*.go

// @Summary 获取个人资料
// @Tags Settings
// @Produce json
// @Success 200 {object} ApiResponse{data=UserProfile}
// @Router /api/v1/settings/profile [get]
func (h *Handler) GetProfile(c *gin.Context) { ... }

// @Summary 更新个人资料
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "更新请求"
// @Success 200 {object} ApiResponse{data=UserProfile}
// @Router /api/v1/settings/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) { ... }

// @Summary 上传头像
// @Tags Settings
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "头像文件"
// @Success 200 {object} ApiResponse{data=AvatarResponse}
// @Router /api/v1/settings/avatar [post]
func (h *Handler) UploadAvatar(c *gin.Context) { ... }

// @Summary 修改密码
// @Tags Settings
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "密码修改请求"
// @Success 200 {object} ApiResponse
// @Router /api/v1/settings/password [post]
func (h *Handler) ChangePassword(c *gin.Context) { ... }

// @Summary 获取会话列表
// @Tags Settings
// @Produce json
// @Success 200 {object} ApiResponse{data=[]Session}
// @Router /api/v1/settings/sessions [get]
func (h *Handler) GetSessions(c *gin.Context) { ... }

// @Summary 远程登出会话
// @Tags Settings
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} ApiResponse
// @Router /api/v1/settings/sessions/{sessionId} [delete]
func (h *Handler) RevokeSession(c *gin.Context) { ... }
```

访问地址：`http://localhost:8080/swagger/index.html`

---

## 9. 部署配置

### 9.1 环境变量

```bash
# .env
# 头像上传配置
AVATAR_UPLOAD_DIR=/var/www/cogniforge/avatars
AVATAR_BASE_URL=https://cdn.cogniforge.com/avatars

# 文件大小限制
MAX_UPLOAD_SIZE_MB=2

# 密码策略
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPER=true
PASSWORD_REQUIRE_NUMBER=true
```

### 9.2 Nginx 配置（头像访问）

```nginx
location /avatars/ {
    alias /var/www/cogniforge/avatars/;
    expires 30d;
    add_header Cache-Control "public, immutable";
    # 可选：图片压缩
    # image_filter resize 200 200;
}
```

---

## 10. 监控指标

| 指标 | 说明 | 类型 |
|------|------|------|
| `settings_profile_views_total` | 个人资料页访问次数 | Counter |
| `avatar_upload_count` | 头像上传次数 | Counter |
| `password_change_count` | 密码修改次数 | Counter |
| `session_revoke_count` | 会话登出次数 | Counter |
| `avatar_upload_failures` | 头像上传失败次数 | Counter |
| `password_change_failures` | 密码修改失败次数 | Counter |

---

## 11. 后续扩展

### 11.1 双因素认证（2FA）

- [ ] TOTP 生成（Google Authenticator 兼容）
- [ ] QR 码展示
- [ ] 验证码验证
- [ ] 备用恢复码
- [ ] 2FA 强制启用（管理员策略）

### 11.2 单点登录（SSO）

- [ ] OAuth2 / OIDC 支持（Google、GitHub、企业 IdP）
- [ ] SAML 支持
- [ ] 自动账户关联

### 11.3 通知偏好设置

- [ ] 邮件通知开关
- [ ] Webhook 通知配置
- [ ] 站内信通知

### 11.4 API 密钥管理（已在 stage 3 完成，可集成到设置页）

- [ ] 显示 API Key 列表
- [ ] 创建新 Key
- [ ] 撤销 Key

---

## 12. 常见��题 FAQ

**Q1：用户忘记密码怎么办？**
A：实现密码重置功能（通过邮箱验证码），暂不纳入阶段九范围。

**Q2：头像存储用本地还是对象存储？**
A：建议 MinIO（本地开发）或 S3（生产），便于 CDN 加速。

**Q3：会话管理需要 WebSocket 吗？**
A：不需要。会话列表通过 HTTP GET 获取，远程登出通过 DELETE。

**Q4：用户删除是软删除还是硬删除？**
A：软删除（`deleted_at` 字段），避免关联数据丢失。

**Q5：2FA 是否必须？**
A：阶段九不强制，Q3 企业版可开启。

---

## 13. 开发排期

| 任务 | 预计时间 | 优先级 |
|------|---------|--------|
| 后端：User/Settings/Roles Handler | 2天 | P0 |
| 后端：RBAC 中间件 | 1天 | P0 |
| 前端：个人资料页面 | 1天 | P0 |
| 前端：头像上传组件 | 1天 | P1 |
| 前端：安全设置页面 | 1天 | P1 |
| 前端：用户管理页面 | 1.5天 | P1 |
| 前端：角色管理页面 | 1.5天 | P1 |
| 联调测试 | 1.5天 | P0 |
| Bug 修复 + 文档 | 1天 | P0 |

**总计**：约 **10 个工作日**（2周，4.15-4.28）

---

**文档版本**: v1.0  
**最后更新**: 2026-04-11  
**维护团队**: CogniForge 前端团队
