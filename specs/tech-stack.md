# Tech Stack（以代码为准）

> README 里部分「规划栈」（Next.js、Milvus/Kafka 必选、Java 用户/计费现网等）**已过时**。下表只写本仓与已接线的兄弟仓现状。未知标 `_TBD_`。

## 本仓（cogniforge）

| 层 | 选型 | 说明 |
|----|------|------|
| 语言 | Go **1.24.5**（`go.mod`） | 单体，不拆微服务 |
| HTTP | Gin | `cmd/server` |
| ORM / DB | GORM + **PostgreSQL**（测试可用 SQLite） | 业务表在本仓迁移/模型 |
| 缓存 | Redis（go-redis） | **只用 db0**；键前缀 `cogniforge:` |
| 调度 | robfig/cron | 工作流 Cron |
| 配置 | Viper + `.env` / `configs/config.yaml` | 业务密钥不进 git |
| 容器 | 根目录 `docker-compose.yml` + `Dockerfile` | 镜像 GHCR；经 Traefik/前端反代 |

### 包分层（强制依赖方向）

```text
gateway → app → runtime / orchestration / governance
                ↓
            capability → infra / shared / model
```

| 目录 | 职责 |
|------|------|
| `internal/gateway` | JWT、CORS、限流门禁 |
| `internal/app` | HTTP 用例（chat/agent/team/workflow/knowledge/…） |
| `internal/runtime` | 对话/Agent 执行核、SSE |
| `internal/orchestration` | 工作流引擎、Team 编排 |
| `internal/capability` | provider / memory / mcp / knowledge / prompt / skill… |
| `internal/governance` | 配额、RBAC、评测、监控策略实现 |
| `internal/infra` | DB、Redis、crypto、modelcache、scheduler… |
| `internal/shared` | 响应封装等 |

**红线**：`capability` 不得反向依赖 `runtime`；明文 API Key 不出 Go 进程内存进 Redis。

## 兄弟仓（对接，非本仓代码）

| 仓 | 栈（代码现状） | 与本仓关系 |
|----|----------------|------------|
| cogniforge-web | **Nuxt 4** + Vue 3 + Pinia + **Nuxt UI 4** + Vue Flow；Playwright e2e | 浏览器控制台；`/api` 打本仓 |
| cogniforge-ai | **FastAPI** + pgvector RAG；Python ≥3.10 | 解析/分块/向量；聊天与 embedding **回调**本仓 |

## 数据与基础设施

| 组件 | 现状 |
|------|------|
| PostgreSQL | 主库；向量用 **pgvector**（在 AI 仓） |
| Redis | 配额计数、模型配置 rev/snapshot、Team run_state 等 |
| 对象存储 | 本地/卷挂载 uploads（`./data/uploads`） |
| Milvus / Qdrant | **未作为现网必选**；规划可选（见 roadmap Later） |
| Kafka / ClickHouse | **未接入** |
| Java `services/*` | 存根/历史目录；**非现网路径** |

## LLM 与密钥

- 供应商与 Key：表 `ai_providers`；管理页写入；本仓解密后调上游
- Redis：`cogniforge:modelcfg:rev`、`cogniforge:modelcfg:snapshot`（只存密文 Key）
- Python：**禁止**再读库解密 LLM Key；经 `COGNIFORGE_API_URL` 回调

## 部署与 CI

| 项 | 现状 |
|----|------|
| CI | `.github/workflows/ci.yml` |
| CD | `.github/workflows/cd.yml`（`workflow_run` 等 CI 成功） |
| 运维手册 | `docs/06-devops/01-deployment.md`（**仍维护**） |
| 路由 | Traefik + 前端 Nginx 反代 `/api`（细节见运维文档） |

## 明确不用 / 本期禁止

- 嵌 Temporal、n8n 作工作流运行时
- 以 LangGraph 等框架替代本仓 `orchestration`
- Redis `SELECT` 切库、无前缀键、明文 Key 进 Redis
- 为「文档里的理想架构」引入未接线组件却不改 roadmap
