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

**状态**：🟢 已统一（见「§4 统一 API 响应格式」）

---

### 3. Nuxt 3 动态路由：目录结构、不刷新与常见坑（cogniforge-web / 工作流）

**背景**：工作流列表进入画布页时曾出现「地址变了但仍是列表」、`[id].vue` 404、HMR 拉取已删除文件、以及 `Cannot access 'loadWorkflow' before initialization` 等问题。以下汇总为可复用的约定与排障要点。

#### 3.1 目录结构（最佳实践）

| 场景 | 推荐结构 | 路由 |
|------|----------|------|
| **A. 仅详情** | `pages/user/[id].vue` | `/user/123` |
| **B. 列表 + 详情（推荐）** | `pages/user/index.vue` + `pages/user/[id].vue` | `/user`、`/user/:id` |
| **C. 详情内多子页** | `pages/project/[id]/index.vue`、`overview.vue`、`files.vue` 等 | `/project/1`、`/project/1/overview` … |

**工作流当前约定**（与 B 一致）：

```text
cogniforge-web/pages/workflows/
  index.vue    → /workflows（列表）
  [id].vue     → /workflows/:id（画布/详情）
```

**避免**：同时存在 `pages/workflows.vue` 与目录 `pages/workflows/`，会与 `/workflows` 解析冲突，并易与 Vite HMR 缓存纠缠；列表应放在 `workflows/index.vue`。

#### 3.2 动态路由跳转后「不刷新」

**原因**：`/workflows/1` → `/workflows/2` 时仍是同一个页面组件实例，Vue 为性能会复用组件，不会自动重新执行仅依赖 `onMounted` 的一次性加载逻辑。

**做法**：监听路由参数变化后再拉数，例如：

```ts
const route = useRoute()

watch(
  () => route.params.id,
  async (newId) => {
    if (newId) await loadData(String(newId))
  },
  { immediate: true },
)
```

或使用 `watchEffect(() => { loadData(route.params.id) })`（在 effect 内读取 `params.id` 即可建立依赖）。

#### 3.3 `<script setup>` 与 `immediate: true` 的顺序陷阱

**现象**：`ReferenceError: Cannot access 'loadWorkflow' before initialization`。

**原因**：`watch(..., { immediate: true })` 在 setup 阶段会**立刻**执行回调；若回调里调用的函数（如 `loadWorkflow`）写在 `watch` **之后** 的 `const loadWorkflow = async () => {}`，会触发 TDZ（暂时性死区），与「函数声明提升」无关。

**做法**：把 `loadWorkflow`（或被 watch 调用的异步函数）定义在 `watch` **之前**，或去掉 `immediate` 改在 `onMounted` 里首次加载（仍建议保留对 `params.id` 的 watch 以处理同页换 id）。

#### 3.4 开发期 404 / 仍请求已删页面

删除或重命名 `pages/**` 后，若控制台仍请求旧的 `pages/workflows/[id].vue` 等资源：

- 停止 dev server，删除 `cogniforge-web/.nuxt` 后重新 `pnpm dev`；
- 浏览器硬刷新，避免旧 HMR 状态。

#### 3.5 本地验证画布（不调后端）

列表页可提供固定测试 id（如 `__canvas_smoke__`）进入 `/workflows/__canvas_smoke__`，在 `[id].vue` 内短路为本地节点数据，用于单独验证动态路由与 Vue Flow 渲染，与网关 401 解耦。

**相关文件**：

- `cogniforge-web/pages/workflows/index.vue`
- `cogniforge-web/pages/workflows/[id].vue`

**状态**：🟢 已按上述约定修复；本文档作后续排障与新人说明

---

### 4. 统一 API 响应格式（前后端）

**背景**：各 handler 原本返回结构混乱——成功有时裸结构、有时 `gin.H{"data":...}`；失败有时 `gin.H{"error":...}`、有时 `gin.H{"message":...}`。前端 composable 也有 `res.data.data` 双层嵌套。

**后端统一结构**（Go，`internal/model/model.go`）：

```go
type ApiResponse struct {
    Code    int         `json:"code"`             // 0=成功，非0=失败
    Message string      `json:"message,omitempty"` // 描述信息
    Data    interface{} `json:"data,omitempty"`   // 业务数据
    Error   string      `json:"error,omitempty"`  // 错误信息（code!=0 时）
}
```

**封装方法**：

| 方法 | HTTP 状态 | 用途 |
|------|----------|------|
| `Success(c, data)` | 200 | 标准成功 |
| `SuccessWithMessage(c, data, msg)` | 200 | 带提示成功 |
| `Created(c, data)` | 201 | 资源创建成功 |
| `Accepted(c, data)` | 202 | 异步接受 |
| `Fail(c, httpStatus, errMsg)` | 自定义 | 通用失败 |
| `FailBadRequest / FailUnauthorized / FailForbidden / FailNotFound / FailInternal` | 4xx/5xx | 快捷失败 |

**前端统一格式**（`utils/apiClient.ts`）：

```ts
export interface ApiResponse<T = unknown> {
  code: number
  data?: T
  error?: string
  message?: string
}
```

**已改造文件**：

| 文件 | 说明 |
|------|------|
| `internal/model/model.go` | 新增 ApiResponse 及封装方法 |
| `internal/handler/auth.go` | 全部接口改用 `model.Success/Fail*` |
| `internal/handler/workflow.go` | 全部接口改用 `model.Success/Fail*` |
| `internal/handler/agent.go` | 全部接口改用 `model.Success/Fail*` |
| `internal/handler/chat.go` | ListModels / Chat / ChatStream 改造 |
| `internal/handler/knowledge.go` | 全部接口改用 `model.Success/Fail*` |
| `utils/apiClient.ts` | ApiResponse 增加 `code` 字段；根据 `code` 自动解析 |
| `composables/useWorkflows.ts` | 去掉双层 `.data.data`，改为 `if (res.error)` 判断 |
| `composables/useAgents.ts` | 同上 |
| `composables/useModels.ts` | 同上 |

**前端调用示例**：

```ts
const res = await api.get<Workflow[]>('/api/v1/workflows/')
if (res.error) {
  message.error(res.error)
  return
}
const workflows = res.data || []
```

**状态**：🟢 已完成

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
| 2026-04-09 | 知识库文档上传功能 | 🟢 已完成 | 阶段七 7.2 |
| 2026-04-09 | 知识库语义检索 API | 🟢 已完成 | 阶段七 7.4 |
| 2026-04-09 | 知识库检索测试页面 | 🟢 已完成 | 阶段七 7.5 |
| 2026-03-21 | Element Plus SSR 水合问题 | 🟡 临时方案 | 禁用 SSR |
| 2026-03-21 | 重复导入警告 | 🟡 待优化 | - |
| 2026-03-21 | 内存数据存储 | 🔴 待实现 | 需要数据库 |
| 2026-04-06 | Nuxt 动态路由（工作流） | 🟢 已修复并文档化 | 见「§3」 |
| 2026-04-06 | 统一 API 响应格式（前后端） | 🟢 已完成 | 见「§4」 |
| 2026-04-06 | 业务 Code 规范重构 | 🟢 已完成 | 2xxx 成功/4xxx 系统异常/5xxx 业务异常 |
| 2026-04-06 | 知识库 CRUD + 文档列表/删除 | 🟢 已完成 | 阶段七 7.1/7.3 |

---

## 📝 更新日志

- **2026-04-09**: 阶段七知识库服务全部完成：文档上传（支持 PDF/TXT/MD/DOCX/HTML）、异步分块处理、基于关键词的语义检索 API、前端检索测试页面
- **2026-04-06**: 业务 Code 规范重构：响应结构拆分 `code.go`/`response.go`/`model.go`、2xxx 成功/4xxx 系统异常/5xxx 业务异常
- **2026-04-06**: 知识库服务（阶段七 7.1/7.3）：新增 `KnowledgeBase`/`Document` 模型、知识库 CRUD API、文档列表/删除 API、前端知识库管理页面
- **2026-04-06**: 新增「§4 统一 API 响应格式」：Go `model.ApiResponse` 结构（`code`/`data`/`error`）、全部 handler 改造、前端 `apiClient` 对齐、后端 composable 去掉双层 `.data`
- **2026-04-06**: 新增「§3 Nuxt 3 动态路由」：目录结构 A/B/C、同组件复用需 watch、`immediate` 与函数声明顺序、`.nuxt` 缓存与 smoke 测试说明
- **2026-03-21**: 初始创建本文档，记录 Element Plus SSR 问题、重复导入警告和内存存储债务
