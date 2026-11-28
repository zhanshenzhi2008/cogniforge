# Cogniforge — Agent 入口

本仓库按 **colincalnan 风格 SDD** 开发。动手改行为前，先读活规格，再写代码。

## 先读（宪法三件套）

1. [`specs/mission.md`](./specs/mission.md) — 为谁、做什么、非目标、MVP
2. [`specs/tech-stack.md`](./specs/tech-stack.md) — 技术红线（以本仓代码为准）
3. [`specs/roadmap.md`](./specs/roadmap.md) — Now / Next / Later

方法约定见 [`specs/method.md`](./specs/method.md)（指向全局 `~/.cursor/rules/sdd-development.mdc`）。

## 仓库角色

| 仓库 | 职责 |
|------|------|
| **cogniforge**（本仓） | Go 单体 API：网关、对话、Agent、Team、工作流引擎、配额、RBAC、模型配置出口 |
| **cogniforge-ai** | Python FastAPI：RAG 解析/分块/pgvector；LLM/embedding **回调本仓**，不自持供应商 Key |
| **cogniforge-web** | Nuxt 4 控制台：Playground、工作室、画布、知识库 UI |

跨仓行为变更：在本仓 `specs/roadmap.md` 挂上条目，并在相关仓用 feature 三件套或同步备注；**不要**为对齐 SDD 去改旧 `docs/` 产品树。

## `docs/` 怎么看待

`docs/` **不是** SDD 活规格目录，默认作 **归档 / 参考**（需求草案、历史架构、变更流水、已知问题）。冲突时以 `specs/` + 代码为准。

**仍维护的非产品规格（运维）**：

- [`docs/06-devops/01-deployment.md`](./docs/06-devops/01-deployment.md) — 部署说明、冒烟清单、环境开关

其余 `docs/01-requirements/`、`02-architecture/`、`03-api/`、`04-database/`、`05-frontend/`、`99-dev-plan/` 等仅供查阅；新功能规格写在 `specs/YYYY-MM-DD-<slug>/`。

## 开发门禁（摘要）

1. Catalog check：mission 冲突？roadmap 有无？tech-stack 红线？
2. 行为变更 → 挂 roadmap → 写/更新 feature 的 `requirements.md` / `plan.md` / `validation.md` → 再改代码
3. 可跳过 feature spec：typo、纯注释、纯格式、不改对外行为的重构
4. 未知标 `_TBD_`，禁止编造技术栈或产品承诺

## 硬约束速记

- Redis：只用 **db0**；键前缀必须 `cogniforge:`（已有 `cogniforge:modelcfg:rev` / `snapshot` 等）
- 明文 API Key **禁止**进 Redis；聊天密钥只经本仓 `ai_providers` 出口
- 敏感值不进规则文件 / commit message / 对话（`.env` 只谈键名）
