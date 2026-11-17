# Cogniforge 单体分层约定（不拆微服务）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-17 | v1.3 | runtime/app 迁入；infra 收 middleware/config/token 等 | orjrs |
| 2026-11-17 | v1.2 | 增 governance；orchestration/shared 物理迁入 | orjrs |
| 2026-11-17 | v1.1 | capability：provider/memory/mcp/skill/knowledge/webfetch 迁入 | orjrs |
| 2026-11-17 | v1.0 | develop 启动分层：文档 + infra 物理迁入 | orjrs |

## [变更] runtime + app + infra 边界收口（2026-11-17）

- **变更后**：
  - `runtime/{chat,agent}`
  - `app/{auth,user,prompt,router}`
  - `infra` 增补 `middleware` / `config` / `token` / `modelcache` / `scheduler`
- **说明**：chat/agent 的 HTTP handler 暂仍与执行同包；后续可再拆到 app

## [变更] governance + orchestration + shared（2026-11-17）

- **变更后**：`governance/{monitor,quota,rbac,eval}`；`orchestration/{engine,workflow}`；`shared/{response,util}`
- **评测**：属治理层，不是业务 CRUD / runtime

## [变更] capability 包迁入（2026-11-17）

- **变更后**：`capability/{provider,memory,mcp,skill,skillimport,knowledge,webfetch}`

## [变更] develop 启动分层（2026-11-17）

- **分支**：`release-20261117` 备份；迁移在 `develop`；测通后再合 `main`

---

## 1. 目标分层

```text
app                 auth / user / prompt / router
  ↓
runtime             chat / agent（执行循环；用 Memory）
orchestration       engine / workflow
governance          monitor / quota / rbac / eval（评测）
  ↓
capability          provider / memory / mcp / skill / knowledge / webfetch …
  ↓
infra               database / crypto / logger / httpclient / mail / trace /
                    middleware / config / token / modelcache / scheduler
shared              response / util
model               持久化实体（暂顶层）
interfaces          预留跨模块接口（暂顶层）
```

## 2. 包对照（当前）

| 路径 | 层 | 状态 |
|------|----|------|
| `infra/*` | infra | ✅ |
| `capability/*` | capability | ✅ |
| `governance/*` | governance | ✅ |
| `orchestration/*` | orchestration | ✅ |
| `runtime/{chat,agent}` | runtime | ✅ |
| `app/{auth,user,prompt,router}` | app | ✅ |
| `shared/{response,util}` | shared | ✅ |
| `model` `interfaces` | 共享/预留 | 暂顶层 |

## 3. Import 规则

1. `infra` 不得依赖上层业务  
2. `capability` 不得依赖 `app` / `router`  
3. `governance` 可依赖 capability（评测调模型）  
4. `runtime` → `capability/memory`（不得反向）  
5. 新代码落目标层目录  

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
