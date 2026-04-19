# CogniForge 技术架构设计文档

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-04-04 | v2.0 | 后端架构由 gateway 独立目录收敛为 monolith；删除 go-standards/dev-environment rules；rules 文档变更记录规范 | orjrs |
| 2026-03-16 | v1.0 | 初始版本 | orjrs |

## [变更] 项目结构变更 + rules 规则整理（2026-04-04）

变更原因：后端架构由 gateway 独立目录收敛为统一 monolith 服务；精简 rules 文件，删除已不适用的 go-standards.mdc 和 dev-environment.mdc
包含代码：`cmd/server/main.go`、`internal/` 各包、`.cursor/rules/`
影响范围：架构设计、rules 规则文件

### 变更前

后端为分立服务规划：gateway/model/agent/workflow 各自独立目录；rules 有 3 个文件。

```
cogniforge/
├── gateway/                      # 独立目录
├── internal/                     # 各模块
├── services/                     # Java 微服务
└── llm/                         # Python ML

.cursor/rules/
├── dev-rules.mdc
├── dev-environment.mdc   # 过时
└── go-standards.mdc     # 过时
```

### 变更后

后端收敛为单体架构；rules 精简为 1 个文件。

```
cogniforge/
├── cmd/server/main.go           # 唯一入口
├── internal/                    # 所有 handler 同进程
├── services/                    # Java 微服务存根
└── llm/                       # Python ML 存根

.cursor/rules/
└── dev-rules.mdc               # 唯一规则文件
```

### 关键差异

- **移除**：`gateway/` 独立目录、`pkg/orjrs/gw/` 代码
- **新增**：`cmd/server/` 作为唯一入口
- **合并**：所有 handler 收敛到 `internal/handler/` 同进程
- **移除 rules 文件**：`go-standards.mdc`（已由变更记录规范覆盖）、`dev-environment.mdc`（端口和启动方式已过时）
- **端口**：原 gateway 8080 + model 8081 + ... → 统一 8080
- **通信**：原 gRPC → 纯 REST 同进程

---

## 1. 技术架构概览

### 1.1 设计原则

| 原则 | 描述 |
|-----|------|
| **高性能** | AI推理延迟要求极高，核心路径采用Go实现 |
| **可扩展** | 微服务架构，支持水平扩展 |
| **多语言融合** | 根据服务特性选择最优技术栈 |
| **云原生** | 容器化部署，Kubernetes编排 |
| **可观测** | 全链路追踪、指标监控、日志聚合 |

### 1.2 技术栈选择理由

| 技术 | 选型理由 | 适用场景 |
|-----|---------|---------|
| **Go** | 高并发、低延迟、编译型无GC暂停 | API网关（monolith）、模型调用、Agent引擎 |
| **Java (Spring Boot 3)** | 成熟稳定、丰富生态 | 用户中心、计费系统（存根） |
| **Python** | AI/ML事实标准 | 模型微调、向量 embedding、RAG处理（存根） |
| **Vue 3 + Nuxt 3** | 组合式 API、SSR/SSG、内置 API 路由 | Web控制台前端 |
| **TypeScript** | 强类型支持、IDE友好 | 全栈类型安全 |
| **PostgreSQL** | ACID事务、JSON支持 | 核心业务数据存储（GORM） |
| **Redis** | 高速缓存、会话存储 | 已配置，暂未使用 |
| **Milvus/Qdrant** | 高效向量检索 | 知识库语义搜索（规划） |
| **Kafka** | 高吞吐消息队列 | 异步任务、事件流（规划） |

---

## 2. 服务架构图（当前实际）

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        全球加速层 (CDN/WAF)                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌────────────────────────────────────────────────────────────────┐     │
│   │               API 层 (Go - 单体 monolith)                        │     │
│   │        cmd/server/main.go → Gin → internal/handler/             │     │
│   └────────────────────────────────────────────────────────────────┘     │
│                                       │                                  │
│           ┌──────────────────────────┼──────────────────────────┐       │
│           │                          ▼                          │       │
│   ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    │
│   │   用户中心      │    │   计费中心       │    │   监控服务      │    │
│   │   (Java存根)   │    │   (Java存根)    │    │   (未实现)      │    │
│   └─────────────────┘    └─────────────────┘    └─────────────────┘    │
│           │                          │                          │       │
│           └──────────────────────────┼──────────────────────────┘       │
│                                      ▼                                 │
│   ┌────────────────────────────────────────────────────────────────┐    │
│   │                    核心服务 (Go - 同一进程)                         │    │
│   │                                                                  │    │
│   │  handler/chat.go  │ handler/agent.go │ handler/workflow.go │      │    │
│   │  handler/auth.go  │ handler/knowledge.go │ middleware/ │         │    │
│   │                                                                  │    │
│   └────────────────────────────────────────────────────────────────┘    │
│                                      ▼                                 │
│   ┌────────────────────────────────────────────────────────────────┐    │
│   │               AI/ML 层 (Python - llm/knowledge/)                  │    │
│   └────────────────────────────────────────────────────────────────┘    │
│                                      ▼                                 │
│      ┌────────────────┐        ┌────────────────┐       ┌────────────┐  │
│      │   PostgreSQL    │        │    Redis      │       │   Kafka   │  │
│      │   5432/5433     │        │    (已配置)   │       │  (未接入)  │  │
│      └────────────────┘        └────────────────┘       └────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. 核心服务设计

### 3.0 API 服务 (Go - 单体 monolith)

**当前状态**：已收敛为单一进程，不再有独立 `gateway/` 目录。

```yaml
入口: cmd/server/main.go
语言: Go 1.22+
框架: Gin
端口: 8080

已实现的 handler (internal/handler/):
  - auth.go:       注册/登录/登出/当前用户
  - chat.go:       聊天流式输出 (OpenAI 兼容 SSE)
  - agent.go:      Agent CRUD + 对话
  - workflow.go:   工作流 CRUD + 执行
  - knowledge.go:  知识库 CRUD (存根)
  - health.go:     健康/就绪/存活探针

中间件:
  - CorsMiddleware: 允许跨域
  - LoggerMiddleware: slog JSON 日志
  - AuthMiddleware: JWT 验证

未实现:
  - 请求限流 (令牌桶)
  - 熔断器模式
  - Prometheus 指标
```

### 3.1 模型网关服务

**当前状态**：功能内嵌在 `internal/handler/chat.go`。

```yaml
实际位置: internal/handler/chat.go
端口: 8080 (与 monolith 共享)

核心功能:
  - OpenAI 兼容 API 格式
  - SSE 流式响应
  - Mock 降级 (无 API Key 时)

支持的模型: OpenAI gpt-4o, gpt-4o-mini, gpt-3.5-turbo 等
```

### 3.2 Agent 引擎服务

**当前状态**：`internal/handler/agent.go` 提供 CRUD + 对话。

```yaml
实际位置: internal/handler/agent.go
端口: 8080 (与 monolith 共享)

已实现:
  - Agent CRUD
  - Agent 对话 (复用 chat 流式输出)
  - 推理模式配置 (存储在 DB)

未实现:
  - 工具注册与发现
  - 短期/长期记忆管理
  - 护栏检查
  - 人机协作
```

### 3.3 工作流编排服务

**当前状态**：`internal/handler/workflow.go` 提供 CRUD + 执行。

```yaml
实际位置: internal/handler/workflow.go
端口: 8080 (与 monolith 共享)

已实现:
  - 工作流 CRUD
  - 执行触发 (POST /workflows/:id/execute)
  - 执行状态存储

未实现:
  - 可视化流程设计器后端
  - 节点执行引擎
  - 条件分支/并行/循环
  - 定时执行 (Cron)
```

### 3.4 知识库服务

**当前状态**：`internal/handler/knowledge.go` 提供 CRUD，Python 处理层待开发。

```yaml
实际位置: internal/handler/knowledge.go + llm/knowledge/
端口: 8080 (Go) + 8081 (Python)

已实现:
  - 知识库 CRUD (Go)
  - 文档列表/删除 (Go)
  - 文档上传接口 (Go，接收 multipart/form-data)

待实现 (Python 处理层):
  - 文件解析：PDF/DOCX/MD/TXT/HTML
  - 文本智能分块（RecursiveCharacterTextSplitter）
  - Embedding 生成（OpenAI API 或本地模型）
  - 向量存储（PostgreSQL pgvector 扩展）
  - 语义检索（向量相似度查询）

技术栈:
  - Python 3.11+
  - FastAPI（Web 框架）
  - unstructured / pypdf（PDF 解析）
  - python-docx（DOCX 解析）
  - openai（Embedding API）
  - pgvector（向量存储，基于 PostgreSQL）
```

**架构流程**：

```
用户上传文档
    │
    ▼
Go Handler (internal/handler/knowledge.go)
    │ - 接收 multipart/form-data
    │ - 保存文件到临时目录
    │ - 创建 Document 记录（status=pending）
    │ - 异步触发 Python 处理（goroutine 或 Kafka）
    ▼
Python FastAPI 服务 (llm/knowledge/)
    │ - 读取文件
    │ - 解析文本内容
    │ - 智能分块（chunk_size=512, overlap=50）
    │ - 调用 OpenAI embedding API 生成向量
    ▼
PostgreSQL (pgvector)
    │ - 存储 chunk 元数据（cf_knowledge_chunks）
    │ - 存储向量（cf_knowledge_vectors 表，vector 列）
    │ - 创建 HNSW 索引（加速检索）
    ▼
完成（status=completed）
```

**数据表设计**（见数据库文档）：
- `cf_knowledge_bases`：知识库元数据
- `cf_knowledge_documents`：文档元数据（状态、chunk 数）
- `cf_knowledge_chunks`：文本分块（content、vector_id）
- `cf_knowledge_vectors`（pgvector）：向量存储（vector 列类型）

**向量检索**：

```sql
-- 使用 pgvector 进行相似度检索
SELECT
    id,
    document_id,
    content,
    1 - (vector <=> $1) as similarity  -- $1 是 query embedding
FROM cf_knowledge_vectors
WHERE knowledge_base_id = $2
ORDER BY vector <=> $1  -- 余弦距离
LIMIT $3;
```

**后续扩展**：当向量数据超过 1000 万条时，可考虑迁移到 Milvus 或 Qdrant（见附录 9.5）。

### 3.5 用户中心服务 (Java) - 存根

```yaml
目录: services/user-service/
语言: Java 21 / Spring Boot 3.2
端口: 8085
实际状态: 仅目录，未启动
```

### 3.6 计费中心服务 (Java) - 存根

```yaml
目录: services/billing/
语言: Java 21 / Spring Boot 3.2
端口: 8086
实际状态: 仅目录，未启动
```

### 3.7 监控服务 - 未实现

```yaml
当前日志: slog JSON 输出到 stdout，由运维层收集
Prometheus/Jaeger/Loki: 未接入
```

### 3.8 AI/ML 处理服务 (Python) - 存根

```yaml
目录: llm/knowledge/
框架: FastAPI
实际状态: 仅目录结构，未启动
```

---

## 4. 服务间通信

### 4.1 通信模式

| 模式 | 技术 | 实际状态 |
|-----|------|---------|
| **同步调用** | REST | 当前唯一方式，所有 handler 同进程 |
| **异步消息** | Kafka | 未接入 |
| **服务发现** | Consul/etcd | 未接入 |

### 4.2 API 路由（当前实际）

```
/health | /ready | /live           GET   健康检查
/auth/register                     POST  用户注册
/auth/login                        POST  用户登录
/auth/logout                       POST  登出
/auth/me                           GET   当前用户
/users/:id                         GET/PUT/DELETE  用户管理
/keys                              POST/GET         API Key 管理
/keys/:id                          DELETE           删除 Key
/models                            GET              模型列表
/models/:id                        GET              模型详情
/chat/stream                       POST             流式聊天（核心）
/agents                            GET/POST          Agent 列表/创建
/agents/:id                        GET/PUT/DELETE   Agent CRUD
/agents/:id/chat                   POST             Agent 对话
/workflows                         GET/POST          工作流列表/创建
/workflows/:id                     GET/PUT/DELETE   工作流 CRUD
/workflows/:id/execute             POST             执行工作流
/knowledge                         GET/POST          知识库列表/创建
/knowledge/:id                     GET/PUT/DELETE   知识库 CRUD
/knowledge/:id/documents           GET              知识库文档列表
```

### 4.3 微服务存根（未激活）

```
services/
├── user-service/       # Java Spring Boot，用户管理
├── auth-service/       # Java Spring Boot，认证（存根）
├── agent-service/      # Java Spring Boot，Agent（存根）
└── workflow-service/   # Java Spring Boot，工作流（存根）

# 未来规划：从 Go monolith 逐步拆分到独立 Java 服务
```

### 4.4 消息队列 - 未接入

```yaml
Kafka Topics (规划，未接入):
  - ai.request: AI请求日志
  - ai.usage: 用量统计
  - ai.workflow.execute: 工作流执行
  - system.audit: 审计日志
```

---

## 5. 数据架构（当前实际）

### 5.1 存储分层

| 层级 | 存储类型 | 实际状态 |
|-----|---------|---------|
| **热数据** | Redis | 已配置，未使用 |
| **温数据** | PostgreSQL | 核心业务数据，实际使用 (GORM) |
| **冷数据** | S3/对象存储 | 未接入 |
| **向量数据** | Milvus/Qdrant | 未接入 |
| **分析数据** | ClickHouse | 未接入 |

### 5.2 数据库表（当前实际）

通过 GORM 自动迁移，核心表：

```
users, api_keys, agents, workflows,
workflow_nodes, workflow_edges, workflow_executions
```

### 5.3 数据隔离

```yaml
隔离策略:
  - 租户隔离: 逻辑隔离 (租户ID)
  - 数据加密: 传输 TLS (生产环境)
  - 备份策略: 手动管理
```

---

## 6. 部署架构（规划中）

### 6.1 当前部署

```yaml
当前: 单体服务，直接运行
  - 后端: go run cmd/server/main.go 或编译后 ./server
  - 前端: nuxt dev / nuxt build
  - 数据库: PostgreSQL 5433 (dev)

未来规划:
  - K8s 部署
  - Docker 容器化
```

### 6.2 高可用配置（规划）

```yaml
目标:
  - 后端: 3 副本 + LoadBalancer
  - 数据库: 主从 + 自动故障转移
  - 缓存: Redis Cluster
  - 消息队列: Kafka (待接入)
  - 向量库: Milvus Cluster (待接入)
```

---

## 7. 安全架构（当前实际）

### 7.1 认证授权

```yaml
已实现:
  - JWT Token: 用户会话验证
  - API 密钥: 程序调用验证
  - 密码: bcrypt 哈希存储

未实现:
  - OAuth2/OIDC SSO
  - RBAC 权限模型
  - 细粒度权限控制
```

### 7.2 安全防护

```yaml
已实现:
  - CORS 跨域配置
  - 密码 bcrypt 加密
  - JWT 签名验证
  - 基础日志审计

未实现:
  - WAF 防护
  - 请求限流 (令牌桶)
  - PII 数据脱敏
  - HashiCorp Vault 密钥管理
```

---

## 8. 技术选型总结（当前实际）

### 8.1 服务技术矩阵

| 服务 | 语言 | 框架 | 实际状态 |
|-----|------|------|---------|
| **monolith** | Go | Gin | 收敛为单一服务，端口 8080 |
| user | Java | Spring Boot | 仅目录存根 |
| billing | Java | Spring Boot | 仅目录存根 |
| ml | Python | FastAPI | 仅目录存根 |
| **前端** | TypeScript | Nuxt 3 + Vue 3 | 实际运行 |

### 8.2 核心原则

> **"Go for serving, Python for training"**
>
> - 实时推理路径使用 Go (monolith)
> - AI/ML 处理预留 Python 层
> - 企业级业务预留 Java 微服务
> - 前端使用 Nuxt 3 + Vue 3 + TypeScript

### 8.3 实际项目结构

```
cogniforge/                          # 后端 Go monolith
├── cmd/server/main.go               # 单一入口
├── configs/config.yaml              # 配置文件
├── internal/
│   ├── config/                      # Viper 配置加载
│   ├── database/                    # GORM PostgreSQL
│   ├── model/                       # 数据模型
│   ├── handler/                     # HTTP 处理器 (auth/chat/agent/workflow/knowledge/health)
│   ├── middleware/                  # CORS/日志/JWT 中间件
│   └── logger/                      # slog JSON 日志
├── services/                        # Java 微服务存根
├── llm/                             # Python ML 存根
└── docs/                            # 文档

cogniforge-web/                     # 前端 Nuxt 3
├── pages/                           # 页面路由
├── composables/                     # 组合式函数
├── layouts/                         # 布局组件
└── assets/                          # 静态资源
```

---

## 9. 附录

### 9.1 端口分配（当前实际）

| 服务 | 端口 | 协议 | 实际状态 |
|-----|------|------|---------|
| **Go monolith** | 8080 | HTTP | 唯一服务，承载全部 API |
| PostgreSQL | 5432 / 5433 | TCP | 开发: 5433 |
| Redis | 6379 | TCP | 已配置，未使用 |
| Java 微服务 | - | - | 未启动 |
| Python ML | - | - | 未启动 |

### 9.2 依赖版本

```yaml
Go: 1.22+
Java: 21
Python: 3.11+
Node.js: 20+
Nuxt: 3.14+
Vue: 3.4+
TypeScript: 5+
PostgreSQL: 15+
Redis: 7+
Spring Boot: 3.2+
```

---

**文档版本**: v2.0
**最后更新**: 2026-04-04
**维护团队**: orjrs
