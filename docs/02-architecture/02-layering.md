# Cogniforge 单体分层约定（不拆微服务）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-17 | v1.0 | develop 启动分层：文档 + infra 物理迁入 | orjrs |

## [变更] develop 启动分层（2026-11-17）

- **变更原因**：按 Agent 平台能力理清依赖，但不拆微服务、不新建仓库
- **分支策略**：`release-20261117` 备份当前 main；迁移在 `develop`；单测/e2e 通过后再合 `main`
- **本阶段**：`internal/infra/{database,crypto,logger,httpclient,trace,mail}` 物理搬家；其余层先建 `doc.go` 占位

---

## 1. 目标分层（Go 扁平风格）

依赖只能向下：

```text
app                 业务应用（HTTP 用例）
  ↓
runtime             Agent/Chat 执行循环（使用 Memory）
orchestration       工作流编排（engine + workflow API）  ← 与 runtime 并列
  ↓
capability          能力：provider / memory / mcp / skill / knowledge / webfetch …
  ↓
infra               database / crypto / logger / httpclient / mail / trace …
shared              response / util（极瘦）
model               持久化实体（跨层共享，暂保持顶层）
```

说明：

- **不新建 GitHub 仓**；三仓仍是 web / go / ai
- Go 包保持相对扁平：`internal/infra/database`，避免 `a/b/c/d` 过深
- **Memory 属于 runtime 使用的能力**：目录可在 capability，调用方向必须是 runtime → memory

## 2. 现有包 → 目标层对照

| 现有包 | 目标层 | 状态 |
|--------|--------|------|
| `infra/database` `infra/crypto` `infra/logger` `infra/httpclient` `infra/trace` `infra/mail` | infra | ✅ 已迁 |
| `config` `middleware` `token` `modelcache` | infra / 边界 | 待定（下一轮） |
| `provider` `memory` `mcp` `skill` `skillimport` `knowledge` `webfetch` | capability | 待迁 |
| `engine` `workflow` | orchestration | 待迁/改名 |
| `chat` `agent` 执行内核 | runtime | 待抽 |
| `chat` `agent` `workflow` `knowledge` `auth` `user`… 的 HTTP | app | 待整理 |
| `response` `util` | shared | 待迁 |
| `model` | 共享实体 | 暂顶层 |
| `router` | app 组装 | 暂顶层 |

## 3. Import 禁止规则（人工/评审约束）

1. `infra/*` 不得 import `chat` / `agent` / `workflow` / `knowledge` / `provider` 等业务包  
2. `capability` 不得 import `router` 或具体 HTTP handler 包（未来 `app/*`）  
3. `memory` 不得被「非 runtime/chat/agent」的无关业务随便引用新增  
4. 新代码优先按目标层落目录，避免继续在顶层摊大饼  

## 4. 迁移节奏

1. ✅ infra 搬家 + 本文档  
2. capability 搬家（provider/memory/mcp/…）  
3. orchestration 目录收敛（engine/workflow）  
4. runtime 抽出执行循环；Memory 调用关系钉死  
5. app 收敛 handler；shared 收 response/util  
6. `cogniforge-ai` 对齐：rag/tools/memory → capability；agent → runtime  

每一步：`go test` + Web 单测/e2e 绿后再考虑合入 `main`。

## 5. 分支

| 分支 | 用途 |
|------|------|
| `main` | 生产稳定 |
| `release-20261117` | 迁移前备份（= 启动迁移时的 main） |
| `develop` | 分层迁移与联调 |
