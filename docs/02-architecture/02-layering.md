# Cogniforge 单体分层约定（不拆微服务）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-17 | v1.2 | 增 governance；orchestration/shared 物理迁入 | orjrs |
| 2026-11-17 | v1.1 | capability：provider/memory/mcp/skill/knowledge/webfetch 迁入 | orjrs |
| 2026-11-17 | v1.0 | develop 启动分层：文档 + infra 物理迁入 | orjrs |

## [变更] governance + orchestration + shared（2026-11-17）

- **变更原因**：监控/配额/权限/评测同属「治理」；编排与公共层继续收敛
- **变更后**：
  - `internal/governance/{monitor,quota,rbac,eval}`（**评测在治理层**）
  - `internal/orchestration/{engine,workflow}`
  - `internal/shared/{response,util}`

## [变更] capability 包迁入（2026-11-17）

- **变更原因**：能力层与业务 HTTP 用例分离
- **变更后**：`internal/capability/{provider,memory,mcp,skill,skillimport,knowledge,webfetch}`

## [变更] develop 启动分层（2026-11-17）

- **变更原因**：按 Agent 平台能力理清依赖，但不拆微服务、不新建仓库
- **分支策略**：`release-20261117` 备份当前 main；迁移在 `develop`；单测/e2e 通过后再合 `main`
- **本阶段**：`internal/infra/{database,crypto,logger,httpclient,trace,mail}` 物理搬家；其余层先建 `doc.go` 占位

---

## 1. 目标分层（Go 扁平风格）

依赖只能向下（governance / orchestration / runtime 同级，都在 app 之下）：

```text
app                 业务应用（HTTP 用例：chat/agent/auth/user…）
  ↓
runtime             Agent/Chat 执行循环（使用 Memory）
orchestration       工作流编排（engine + workflow）
governance          监控治理：monitor / quota / rbac / eval（评测）
  ↓
capability          能力：provider / memory / mcp / skill / knowledge / webfetch …
  ↓
infra               database / crypto / logger / httpclient / mail / trace …
shared              response / util（极瘦）
model               持久化实体（跨层共享，暂保持顶层）
```

说明：

- **不新建 GitHub 仓**；三仓仍是 web / go / ai
- Go 包保持相对扁平：`internal/infra/database`，避免过深嵌套
- **Memory**：能力在 capability，调用方向 runtime → memory
- **Eval 评测**：归 **governance**（质量门禁/治理），不是业务 CRUD，也不是 runtime

## 2. 包 → 层对照

| 包 | 目标层 | 状态 |
|----|--------|------|
| `infra/{database,crypto,logger,httpclient,trace,mail}` | infra | ✅ |
| `capability/{provider,memory,mcp,skill,skillimport,knowledge,webfetch}` | capability | ✅ |
| `governance/{monitor,quota,rbac,eval}` | governance | ✅ |
| `orchestration/{engine,workflow}` | orchestration | ✅ |
| `shared/{response,util}` | shared | ✅ |
| `config` `middleware` `token` `modelcache` | infra / 边界 | 待定 |
| `chat` `agent` 执行内核 | runtime | 待抽 |
| `chat` `agent` `auth` `user` `prompt`… HTTP | app | 待整理 |
| `model` | 共享实体 | 暂顶层 |
| `router` | app 组装 | 暂顶层 |

## 3. Import 禁止规则

1. `infra/*` 不得 import 业务 / capability / governance / orchestration  
2. `capability` 不得 import `router` 或未来 `app/*`  
3. `governance` 可依赖 capability（评测调模型），勿依赖具体页面编排细节  
4. `memory` 新增引用应来自 runtime/chat/agent  
5. 新代码优先落目标层目录  

## 4. 迁移节奏

1. ✅ infra  
2. ✅ capability  
3. ✅ governance（含 eval）+ orchestration + shared  
4. runtime 抽出执行循环  
5. app 收敛 handler（chat/agent/auth/user/prompt…）  
6. `cogniforge-ai` 对齐  

每一步：`go test` + Web 单测；合 `main` 前再跑 e2e。

## 5. 分支

| 分支 | 用途 |
|------|------|
| `main` | 生产稳定 |
| `release-20261117` | 迁移前备份 |
| `develop` | 分层迁移与联调 |
