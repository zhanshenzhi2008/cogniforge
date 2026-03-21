# CogniForge 技术债务与待优化项

## 📋 概述

本文档用于记录开发过程中发现的技术债务、已知问题和待优化项，便于后期跟进和优化。

---

## 🐛 已知问题

### 1. Element Plus 与 Nuxt SSR 水合不匹配

**问题描述**：

在使用 Nuxt SSR 模式时，Element Plus 组件（如 `ElDropdown`、`ElTooltip`、`ElEmpty`）在服务端渲染和客户端渲染时生成的 DOM ID 不一致，导致大量 Hydration Mismatch 警告：

```
[Vue warn]: Hydration attribute mismatch on <span class="user-dropdown">
  - rendered on server: id="el-id-7461-20"
  - expected on client: id="el-id-6630-0"
```

**影响范围**：
- `layouts/default.vue` 中的下拉菜单组件
- `pages/index.vue` 中的空状态组件
- 任何使用 Element Plus 弹窗类组件的页面

**当前解决方案**：
在 `nuxt.config.ts` 中设置 `ssr: false`，禁用服务端渲染。

**影响**：
- ✅ 功能正常
- ⚠️ SEO 略有影响（首屏需要 JS 执行后才能渲染）
- ⚠️ 首屏加载略慢

**后续优化方案**（按优先级）：

| 优先级 | 方案 | 说明 |
|-------|------|------|
| P1 | 升级 Element Plus | 检查新版本是否已修复此问题 |
| P2 | 使用 `<ClientOnly>` 包裹 | 只在客户端渲染特定组件 |
| P3 | 自定义 Element Plus ID 生成策略 | 需要修改 Element Plus 源码或配置 |

**相关文件**：
- `nuxt.config.ts` - 已设置 `ssr: false`
- `layouts/default.vue` - 使用 ElDropdown 组件
- `pages/index.vue` - 使用 ElEmpty 组件

**状态**：🟡 临时方案，待长期优化

---

### 2. 重复导入警告

**问题描述**：

控制台出现重复导入警告：

```
[warn] Duplicated imports "HealthResponse", the one from "composables/useApi.ts" has been ignored
[warn] Duplicated imports "ApiResponse", the one from "composables/useApi.ts" has been ignored
```

**原因**：
`composables/useApi.ts` 和 `utils/apiClient.ts` 中定义了相同的类型，但 Nuxt 自动导入功能导致类型被重复导入。

**影响**：
- ⚠️ 控制台警告
- ✅ 功能正常

**优化方案**：
- 统一类型定义位置，避免重复导出
- 将共享类型统一放置在 `types/` 目录

**相关文件**：
- `composables/useApi.ts`
- `utils/apiClient.ts`
- `types/` 目录

**状态**：🟡 待优化

---

## 🔧 技术债务

### 1. 内存数据存储

**描述**：
当前后端使用内存存储用户和会话数据，服务重启后数据会丢失。

```go
// gateway/pkg/orjrs/gw/handler/auth.go
var users = map[string]*User{}  // 内存存储
```

**影响**：
- ⚠️ 开发测试受影响
- 🚨 生产环境不可用

**解决方案**：
- 集成 PostgreSQL 数据库
- 使用 Redis 存储会话

**优先级**：P0（生产必需）

**状态**：🔴 待实现

---

### 2. API 端点占位符

**描述**：
以下 API 端点在 `main.go` 中已注册路由，但处理器仅返回空数据或占位符响应：

| 端点 | 模块 | 状态 |
|------|------|------|
| `/api/v1/users/*` | 用户管理 | 占位符 |
| `/api/v1/keys/*` | API密钥 | 占位符 |
| `/api/v1/models/*` | 模型网关 | 占位符 |
| `/api/v1/agents/*` | Agent引擎 | 占位符 |
| `/api/v1/workflows/*` | 工作流 | 占位符 |
| `/api/v1/knowledge/*` | 知识库 | 占位符 |

**相关文件**：
- `gateway/cmd/server/main.go`

**状态**：🔴 待实现

---

## 📊 优化建议

### 1. 前端优化

| 建议 | 说明 | 优先级 |
|------|------|-------|
| 添加错误边界 | 捕获组件渲染错误 | P2 |
| 骨架屏加载 | 改善首屏体验 | P2 |
| 图片懒加载 | 减少初始加载时间 | P3 |
| API 请求缓存 | 使用 Vue Query 配置缓存策略 | P2 |

### 2. 后端优化

| 建议 | 说明 | 优先级 |
|------|------|-------|
| 添加数据库连接池 | 提高并发性能 | P1 |
| 实现 Redis 缓存 | 减少数据库压力 | P1 |
| 添加请求超时 | 防止慢请求占用资源 | P1 |
| 日志结构化 | 便于日志分析和查询 | P2 |

### 3. 安全优化

| 建议 | 说明 | 优先级 |
|------|------|-------|
| 添加请求速率限制 | 防止 DDoS | P1 |
| 实现 CSRF 防护 | Web 安全 | P1 |
| 添加请求签名验证 | API 安全 | P2 |
| 敏感信息加密存储 | 数据安全 | P1 |

---

## 📅 跟进记录

| 日期 | 问题/优化项 | 处理状态 | 备注 |
|------|-----------|---------|------|
| 2026-03-21 | Element Plus SSR 水合问题 | 🟡 临时方案 | 禁用 SSR |
| 2026-03-21 | 重复导入警告 | 🟡 待优化 | - |
| 2026-03-21 | 内存数据存储 | 🔴 待实现 | 需要数据库 |

---

## 📝 更新日志

- **2026-03-21**: 初始创建本文档，记录 Element Plus SSR 问题、重复导入警告和内存存储债务
