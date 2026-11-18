# Cogniforge 单体分层约定（不拆微服务）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-18 | v1.5 | 锁定：provider/memory 留 capability；context 由 runtime 组装 | orjrs |
| 2026-11-17 | v1.4 | 增 gateway：边缘鉴权/限流中间件迁入；配额实现仍在 governance | orjrs |
| 2026-11-17 | v1.3 | runtime/app 迁入；infra 收 middleware/config/token 等 | orjrs |
| 2026-11-17 | v1.2 | 增 governance；orchestration/shared 物理迁入 | orjrs |
| 2026-11-17 | v1.1 | capability：provider/memory/mcp/skill/knowledge/webfetch 迁入 | orjrs |
| 2026-11-17 | v1.0 | develop 启动分层：文档 + infra 物理迁入 | orjrs |

## [变更] 锁定 Memory / Provider 归属（2026-11-18）

- **变更原因**：容易误以为「对话相关都该进 runtime」；实际 runtime 是执行循环，Memory/Provider 是可复用积木。
- **约定（已确认，不挪包）**：
  | 概念 | 落点 | 说明 |
  |------|------|------|
  | Provider | `capability/provider` | 连哪家模型、Key、模型列表；管理页与执行共用 |
  | Memory | `capability/memory` | 存取与窗口裁剪；非「跑一轮」本身 |
  | Context | **不单独成包** | 由 `runtime/chat|agent` 用 Memory + RAG + 系统提示当场拼 |
  | 执行循环 | `runtime` | 调模型、工具、流式；`runtime → capability`，禁止反向 |
- **影响范围**：仅文档与包注释；目录不变

## [变更] gateway 边缘层（2026-11-17）

- **变更原因**：鉴权校验、CORS、RBAC 门禁、HTTP 限流属于「请求入口」职责，不应堆在 infra 或业务包里。
- **变更后**：
  - `gateway/middleware`：CORS / Recovery / 请求日志 / JWT / RequireAdmin
  - `gateway`：`RateLimit` / `RateLimitFromQuota`（调用 `governance/quota`，不存策略）
- **边界**：
  | 能力 | 落点 |
  |------|------|
  | 登录/注册/改密 | `app/auth`（业务用例） |
  | JWT 校验、CORS、Admin 门禁 | `gateway` |
  | 配额策略 / 计数 / RPM | `governance/quota` |
  | 聊天路径内 `Allow` | 仍在 `runtime/chat|agent`（避免网关再挂导致双重扣次） |
- **包含代码**：`internal/gateway/**`；`cmd/server`、`app/router`、auth/user/token 的 import 改指向 gateway
- **影响范围**：包路径变更；对外 HTTP 行为不变

## [变更] runtime + app + infra 边界收口（2026-11-17）

- **变更后**：
  - `runtime/{chat,agent}`
  - `app/{auth,user,prompt,router}`
  - `infra`：`config` / `token` / `modelcache` / `scheduler`（middleware 已迁出）
- **说明**：chat/agent 的 HTTP handler 暂仍与执行同包；后续可再拆到 app

## [变更] governance + orchestration + shared（2026-11-17）

- **变更后**：`governance/{monitor,quota,rbac,eval}`；`orchestration/{engine,workflow}`；`shared/{response,util}`
- **评测**：属治理层，不是业务 CRUD / runtime

## [变更] capability 包迁入（2026-11-17）

- **变更后**：`capability/{provider,memory,mcp,skill,knowledge,webfetch}`

## [变更] develop 启动分层（2026-11-17）

- **分支**：`release-20261117` 备份；迁移在 `develop`；测通后再合 `main`

---

## 1. 目标分层

```text
gateway             CORS / JWT 校验 / RBAC 门禁 / HTTP 限流中间件
  ↓
app                 auth / user / prompt / router
  ↓
runtime             chat / agent（执行循环；用 Memory；内嵌业务配额 Allow）
orchestration       engine / workflow
governance          monitor / quota / rbac / eval（评测）
  ↓
capability          provider / memory / mcp / skill / knowledge / webfetch …
  ↓
infra               database / crypto / logger / httpclient / mail / trace /
                    config / token / modelcache / scheduler
shared              response / util
model               持久化实体（暂顶层）
interfaces          预留跨模块接口（暂顶层）
```

## 2. 包对照（当前）

| 路径 | 层 | 状态 |
|------|----|------|
| `gateway` `gateway/middleware` | gateway | ✅ |
| `infra/*`（无 middleware） | infra | ✅ |
| `capability/*` | capability | ✅ |
| `governance/*` | governance | ✅ |
| `orchestration/*` | orchestration | ✅ |
| `runtime/{chat,agent}` | runtime | ✅ |
| `app/{auth,user,prompt,router}` | app | ✅ |
| `shared/{response,util}` | shared | ✅ |
| `model` `interfaces` | 共享/预留 | 暂顶层 |

## 3. Import 规则

1. `infra` 不得依赖上层业务与 `gateway`  
2. `capability` 不得依赖 `app` / `router` / `gateway` / `runtime`  
3. `governance` 可依赖 capability（评测调模型）  
4. `gateway` 可依赖 `governance`（限流调配额）与 `shared`；不得依赖 `app`/`runtime`  
5. `runtime` → `capability/{memory,provider,…}`（不得反向）  
6. 新代码落目标层目录  

## 3.1 runtime vs capability（易混点）

- **runtime** = 这一轮 **怎么跑**（循环、拼 context、调工具、SSE）
- **capability** = 跑的时候 **有什么积木**（provider / memory / mcp / skill / knowledge / webfetch）
- 即使今天几乎只有 chat 用 Memory/Provider，也仍留 capability，便于 eval、workflow、embedding 复用，并避免 runtime 变成杂物间

## 4. 后续

1. chat/agent handler 与执行核再拆（handler → app）  
2. `cogniforge-ai` 对齐 capability/runtime  
3. e2e 绿后合 `main`  

## 5. 分支

| 分支 | 用途 |
|------|------|
| `main` | 生产 |
| `release-20261117` | 迁移前备份 |
| `develop` | 分层迁移 |
