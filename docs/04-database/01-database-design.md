# CogniForge 数据库设计文档

## 1. 数据库架构概述

### 1.1 存储选型

| 存储类型 | 数据库 | 用途 |
|---------|-------|------|
| **关系型** | PostgreSQL 15+ | 核心业务数据、事务数据 |
| **缓存** | Redis 7+ | 会话、限流、实时数据 |
| **向量** | Milvus/Qdrant | 知识库向量存储 |
| **消息队列** | Kafka 3.6+ | 异步任务、事件流 |
| **对象存储** | S3/MinIO | 文件、日志归档 |
| **分析** | ClickHouse | 用量统计、成本分析 |

### 1.2 命名规范

```sql
-- 表名: 小写字母 + 下划线 (snake_case)
-- 表前缀: cf_ (cogniforge)

-- 示例:
-- 用户表: cf_users
-- API密钥表: cf_api_keys
-- Agent表: cf_agents

-- 字段名: 小写字母 + 下划线
-- 主键: id (UUID)
-- 外键: xxx_id
-- 创建时间: created_at
-- 更新时间: updated_at
-- 删除时间: deleted_at (软删除)
```

---

## 2. 核心业务表

### 2.1 组织与用户

#### 2.1.1 组织表 (cf_organizations)

```sql
CREATE TABLE cf_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    plan VARCHAR(50) DEFAULT 'free',  -- free, pro, enterprise
    settings JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',  -- active, suspended
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_org_slug ON cf_organizations(slug);
CREATE INDEX idx_org_status ON cf_organizations(status);
```

#### 2.1.2 用户表 (cf_users)

```sql
CREATE TABLE cf_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url VARCHAR(500),
    phone VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',  -- active, disabled
    email_verified BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_user_email ON cf_users(email);
CREATE INDEX idx_user_org ON cf_users(organization_id);
CREATE INDEX idx_user_status ON cf_users(status);

-- 软删除查询
-- SELECT * FROM cf_users WHERE organization_id = ? AND deleted_at IS NULL;
```

#### 2.1.3 角色表 (cf_roles)

```sql
CREATE TABLE cf_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system BOOLEAN DEFAULT FALSE,  -- 系统预置角色不可删除
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, name)
);

-- 预置角色
INSERT INTO cf_roles (id, organization_id, name, description, permissions, is_system) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'super_admin', '超级管理员', '["*"]', TRUE),
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'org_admin', '组织管理员', '["users:*", "agents:*", "workflows:*", "billing:*"]', TRUE),
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000000', 'developer', '开发者', '["agents:*", "workflows:*", "knowledge_bases:*"]', TRUE),
('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000000', 'analyst', '分析师', '["usage:read", "logs:read"]', TRUE);
```

#### 2.1.4 用户角色关联表 (cf_user_roles)

```sql
CREATE TABLE cf_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES cf_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES cf_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_role_user ON cf_user_roles(user_id);
CREATE INDEX idx_user_role_role ON cf_user_roles(role_id);
```

#### 2.1.5 用户会话表 (cf_user_sessions)

```sql
-- 用户登录会话管理
-- 每次用户登录时创建记录，用于展示"登录设备列表"和远程登出
CREATE TABLE cf_user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES cf_users(id) ON DELETE CASCADE,
    session_id VARCHAR(100) NOT NULL UNIQUE,  -- JWT ID 或 Redis Session ID
    ip_address INET,                          -- 登录 IP
    user_agent TEXT,                          -- User-Agent 原始字符串
    device_info JSONB DEFAULT '{}',          -- 设备信息 {os, browser, device_type}
    location VARCHAR(100),                   -- 地理位置（可选，IP 解析）
    last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 索引：快速查询用户会话、清理过期会话
CREATE INDEX idx_session_user ON cf_user_sessions(user_id);
CREATE INDEX idx_session_expires ON cf_user_sessions(expires_at);
CREATE INDEX idx_session_active ON cf_user_sessions(last_active_at DESC);
```

**说明**：
- `session_id`：从 JWT Token 的 `jti` 声明或 Session ID 生成
- `device_info`：解析 User-Agent 得到 `{os: "macOS", browser: "Chrome 120", device_type: "desktop"}`
- `last_active_at`：每次请求更新，用于判断活跃状态
- `expires_at`：Session 过期时间（Token 有效期）
- 定期清理任务：`DELETE FROM cf_user_sessions WHERE expires_at < NOW()`

---

### 2.2 API密钥与认证

#### 2.2.1 API密钥表 (cf_api_keys)

```sql
CREATE TABLE cf_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    user_id UUID NOT NULL REFERENCES cf_users(id),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,  -- SHA256 hash
    key_prefix VARCHAR(20) NOT NULL,  -- sk-cf-xxxx 前缀
    ip_whitelist TEXT[],  -- IP白名单
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active',  -- active, revoked
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_api_key_org ON cf_api_keys(organization_id);
CREATE INDEX idx_api_key_hash ON cf_api_keys(key_hash);
CREATE INDEX idx_api_key_prefix ON cf_api_keys(key_prefix);
CREATE INDEX idx_api_key_status ON cf_api_keys(status);
```

---

### 2.3 Agent服务

#### 2.3.1 Agent表 (cf_agents)

```sql
CREATE TABLE cf_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    user_id UUID NOT NULL REFERENCES cf_users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    model VARCHAR(100) NOT NULL,
    system_prompt TEXT,
    tools JSONB DEFAULT '[]',  -- ["tool_id1", "tool_id2"]
    memory_config JSONB DEFAULT '{"type": "short_term", "max_turns": 10}',
    guardrails JSONB DEFAULT '{"input_filter": true, "output_filter": true}',
    status VARCHAR(20) DEFAULT 'active',  -- active, disabled
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_agent_org ON cf_agents(organization_id);
CREATE INDEX idx_agent_user ON cf_agents(user_id);
CREATE INDEX idx_agent_status ON cf_agents(status);
```

#### 2.3.2 Agent对话历史表 (cf_agent_conversations)

```sql
CREATE TABLE cf_agent_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES cf_agents(id) ON DELETE CASCADE,
    session_id VARCHAR(100) NOT NULL,
    messages JSONB NOT NULL DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_conv_agent ON cf_agent_conversations(agent_id);
CREATE INDEX idx_conv_session ON cf_agent_conversations(session_id);
```

---

### 2.4 工作流服务

#### 2.4.1 工作流表 (cf_workflows)

```sql
CREATE TABLE cf_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    user_id UUID NOT NULL REFERENCES cf_users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    nodes JSONB NOT NULL DEFAULT '[]',  -- 节点定义
    edges JSONB NOT NULL DEFAULT '[]',    -- 连线定义
    version INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'draft',  -- draft, published, archived
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_workflow_org ON cf_workflows(organization_id);
CREATE INDEX idx_workflow_user ON cf_workflows(user_id);
CREATE INDEX idx_workflow_status ON cf_workflows(status);
```

#### 2.4.2 工作流版本表 (cf_workflow_versions)

```sql
CREATE TABLE cf_workflow_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES cf_workflows(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    nodes JSONB NOT NULL,
    edges JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(workflow_id, version)
);

CREATE INDEX idx_wfv_workflow ON cf_workflow_versions(workflow_id);
```

#### 2.4.3 工作流执行记录表 (cf_workflow_executions)

```sql
CREATE TABLE cf_workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES cf_workflows(id),
    execution_id VARCHAR(100) NOT NULL UNIQUE,
    input JSONB DEFAULT '{}',
    output JSONB DEFAULT '{}',
    node_executions JSONB DEFAULT '[]',  -- 每个节点的执行结果
    status VARCHAR(20) DEFAULT 'pending',  -- pending, running, completed, failed, cancelled
    error_message TEXT,
    tokens_used INTEGER DEFAULT 0,
    duration_ms INTEGER DEFAULT 0,
    triggered_by VARCHAR(50),  -- api, webhook, schedule
    webhook_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_exec_workflow ON cf_workflow_executions(workflow_id);
CREATE INDEX idx_exec_status ON cf_workflow_executions(status);
CREATE INDEX idx_exec_id ON cf_workflow_executions(execution_id);
CREATE INDEX idx_exec_created ON cf_workflow_executions(created_at DESC);
```

---

### 2.5 知识库服务

#### 2.5.1 知识库表 (cf_knowledge_bases)

```sql
CREATE TABLE cf_knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    user_id UUID NOT NULL REFERENCES cf_users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-small',
    chunk_size INTEGER DEFAULT 512,
    chunk_overlap INTEGER DEFAULT 50,
    status VARCHAR(20) DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_kb_org ON cf_knowledge_bases(organization_id);
CREATE INDEX idx_kb_status ON cf_knowledge_bases(status);
```

#### 2.5.2 文档表 (cf_knowledge_documents)

```sql
CREATE TABLE cf_knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES cf_knowledge_bases(id) ON DELETE CASCADE,
    filename VARCHAR(500) NOT NULL,
    file_path VARCHAR(1000) NOT NULL,  -- S3路径
    file_size BIGINT NOT NULL,
    file_type VARCHAR(50) NOT NULL,  -- pdf, docx, md, txt
    status VARCHAR(20) DEFAULT 'pending',  -- pending, processing, completed, failed
    chunk_count INTEGER DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_doc_kb ON cf_knowledge_documents(knowledge_base_id);
CREATE INDEX idx_doc_status ON cf_knowledge_documents(status);

-- 注意: 向量数据使用 PostgreSQL pgvector 扩展存储
-- 文档Chunk表存储分块文本及其向量（vector列）
CREATE TABLE cf_knowledge_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES cf_knowledge_documents(id) ON DELETE CASCADE,
    knowledge_base_id UUID NOT NULL,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    vector vector(1536),  -- pgvector列，OpenAI embedding维度
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- HNSW 索引：加速向量相似度检索
-- 适用于高维向量近似最近邻搜索（ANN）
CREATE INDEX idx_chunk_vector ON cf_knowledge_chunks
USING hnsw (vector vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- 辅助索引
CREATE INDEX idx_chunk_doc ON cf_knowledge_chunks(document_id);
CREATE INDEX idx_chunk_kb ON cf_knowledge_chunks(knowledge_base_id);
```

---

### 2.6 微调训练

#### 2.6.1 数据集表 (cf_datasets)

```sql
CREATE TABLE cf_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    user_id UUID NOT NULL REFERENCES cf_users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    file_path VARCHAR(1000) NOT NULL,
    file_size BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'validating',  -- validating, ready, failed
    sample_count INTEGER DEFAULT 0,
    validation_result JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_dataset_org ON cf_datasets(organization_id);
CREATE INDEX idx_dataset_status ON cf_datasets(status);
```

#### 2.6.2 训练任务表 (cf_fine_tunes)

```sql
CREATE TABLE cf_fine_tunes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    user_id UUID NOT NULL REFERENCES cf_users(id),
    dataset_id UUID NOT NULL REFERENCES cf_datasets(id),
    base_model VARCHAR(100) NOT NULL,
    fine_tuned_model VARCHAR(100),  -- 训练完成后填充
    name VARCHAR(255),
    hyperparameters JSONB DEFAULT '{"epochs": 3, "batch_size": "auto", "learning_rate_multiplier": 1.0}',
    status VARCHAR(20) DEFAULT 'queued',  -- queued, running, completed, failed, cancelled
    result JSONB,  -- 训练结果
    error_message TEXT,
    training_job_id VARCHAR(100),  -- 外部训练平台Job ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_ft_org ON cf_fine_tunes(organization_id);
CREATE INDEX idx_ft_dataset ON cf_fine_tunes(dataset_id);
CREATE INDEX idx_ft_status ON cf_fine_tunes(status);
```

---

### 2.7 模型配置

#### 2.7.1 模型供应商表 (cf_model_providers)

```sql
CREATE TABLE cf_model_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    api_base_url VARCHAR(500),
    logo_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO cf_model_providers (name, display_name, api_base_url) VALUES
('openai', 'OpenAI', 'https://api.openai.com/v1'),
('anthropic', 'Anthropic', 'https://api.anthropic.com/v1'),
('google', 'Google', 'https://generativelanguage.googleapis.com/v1'),
('cohere', 'Cohere', 'https://api.cohere.ai/v1'),
('local', '本地部署', NULL);
```

#### 2.7.2 模型表 (cf_models)

```sql
CREATE TABLE cf_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES cf_model_providers(id),
    model_id VARCHAR(100) NOT NULL,  -- 如 gpt-4o
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    model_type VARCHAR(50),  -- chat, embedding, image, audio
    context_window INTEGER,
    max_output_tokens INTEGER,
    pricing_input DECIMAL(10, 6),  -- 每1K token价格
    pricing_output DECIMAL(10, 6),
    capabilities JSONB DEFAULT '[]',  -- ["streaming", "function_call"]
    status VARCHAR(20) DEFAULT 'active',
    deprecated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(provider_id, model_id)
);
```

#### 2.7.3 组织模型配置表 (cf_organization_models)

```sql
CREATE TABLE cf_organization_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    model_id UUID NOT NULL REFERENCES cf_models(id),
    api_key_encrypted TEXT,  -- 加密存储
    status VARCHAR(20) DEFAULT 'active',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, model_id)
);
```

---

## 3. 日志与统计表

### 3.1 请求日志表 (cf_request_logs)

```sql
CREATE TABLE cf_request_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    api_key_id UUID REFERENCES cf_api_keys(id),
    user_id UUID REFERENCES cf_users(id),
    trace_id VARCHAR(100) NOT NULL,  -- 链路追踪ID
    span_id VARCHAR(50),
    parent_span_id VARCHAR(50),
    
    -- 请求信息
    method VARCHAR(10) NOT NULL,
    path VARCHAR(500) NOT NULL,
    request_body JSONB,
    request_headers JSONB,
    
    -- 模型信息
    model VARCHAR(100),
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    
    -- 响应信息
    status_code INTEGER,
    response_body JSONB,
    error_message TEXT,
    
    -- 性能信息
    latency_ms INTEGER,
    tokens_per_second DECIMAL(10, 2),
    
    -- 成本
    input_cost DECIMAL(10, 6),
    output_cost DECIMAL(10, 6),
    total_cost DECIMAL(10, 6),
    
    -- 其他
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- 按月分区
CREATE TABLE cf_request_logs_2026_01 PARTITION OF cf_request_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE INDEX idx_log_org ON cf_request_logs(organization_id);
CREATE INDEX idx_log_trace ON cf_request_logs(trace_id);
CREATE INDEX idx_log_model ON cf_request_logs(model);
CREATE INDEX idx_log_created ON cf_request_logs(created_at DESC);
CREATE INDEX idx_log_api_key ON cf_request_logs(api_key_id);
```

### 3.2 用量统计表 (cf_usage_stats)

```sql
CREATE TABLE cf_usage_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES cf_organizations(id),
    
    -- 时间维度
    stat_date DATE NOT NULL,
    stat_hour INTEGER,  -- 0-23
    
    -- 统计维度
    model VARCHAR(100),
    provider VARCHAR(50),
    
    -- 用量数据
    request_count BIGINT DEFAULT 0,
    success_count BIGINT DEFAULT 0,
    error_count BIGINT DEFAULT 0,
    input_tokens BIGINT DEFAULT 0,
    output_tokens BIGINT DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    total_cost DECIMAL(15, 6) DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, stat_date, stat_hour, model)
);

CREATE INDEX idx_usage_org_date ON cf_usage_stats(organization_id, stat_date DESC);
CREATE INDEX idx_usage_date ON cf_usage_stats(stat_date DESC);
```

---

## 4. 审计日志表

### 4.1 审计日志表 (cf_audit_logs)

```sql
CREATE TABLE cf_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID REFERENCES cf_organizations(id),
    user_id UUID REFERENCES cf_users(id),
    
    -- 操作信息
    action VARCHAR(100) NOT NULL,  -- create, update, delete, login, logout
    resource_type VARCHAR(50) NOT NULL,  -- user, agent, workflow, etc.
    resource_id UUID,
    
    -- 请求信息
    ip_address INET,
    user_agent TEXT,
    request_method VARCHAR(10),
    request_path VARCHAR(500),
    
    -- 变更内容
    changes JSONB,  -- {"old": {...}, "new": {...}}
    result VARCHAR(20),  -- success, failure
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_org ON cf_audit_logs(organization_id);
CREATE INDEX idx_audit_user ON cf_audit_logs(user_id);
CREATE INDEX idx_audit_action ON cf_audit_logs(action);
CREATE INDEX idx_audit_resource ON cf_audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_created ON cf_audit_logs(created_at DESC);
```

---

## 5. Redis 缓存设计

### 5.1 缓存键设计

```redis
# 用户会话
session:{user_id} -> JSON {token, expires_at}

# API限流
rate_limit:{api_key_id}:{minute} -> counter
rate_limit:{api_key_id}:{hour} -> counter

# 模型配置缓存
model_config:{org_id}:{model_id} -> JSON {api_key, settings}

# Agent对话缓存
agent_conv:{agent_id}:{session_id} -> JSON {messages}

# 工作流执行状态
workflow_exec:{execution_id} -> JSON {status, result}

# Token计数
token_count:{org_id}:{date} -> counter
```

---

## 6. 向量数据库设计 (Milvus/Qdrant)

### 6.1 Collection设计

```yaml
知识库向量集合:
  名称: cf_knowledge_vectors
  向量维度: 1536 (text-embedding-3-small)
  
  字段:
    - id: string (主键)
    - document_id: string (文档ID)
    - knowledge_base_id: string (知识库ID)
    - chunk_index: int (分块索引)
    - content: string (文本内容)
    - metadata: json (元数据)
    - vector: float vector (向量)
  
  索引: HNSW
  距离度量: cosine
```

---

## 7. 数据库迁移策略

### 7.1 迁移工具

- **Go**: 使用 `golang-migrate` 或 `Gorm`
- **Java**: 使用 Flyway 或 Liquibase

### 7.2 迁移原则

1. 始终向后兼容
2. 大表修改使用 `ALTER TABLE` 的 `LOCK` 选项
3. 索引创建使用 `CONCURRENTLY` 选项
4. 敏感数据加密存储
5. 定期清理过期数据

---

## 8. pgvector 向量扩展使用说明

### 8.1 安装与配置

**PostgreSQL 版本要求**：PostgreSQL 15+（推荐 15 或 16）

```bash
# 检查 PostgreSQL 版本
psql --version

# 进入 PostgreSQL 容器或本地实例
docker exec -it cogniforge-postgres psql -U cogniforge -d cogniforge

# 安装 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;
```

**验证安装**：
```sql
SELECT * FROM pg_extension WHERE extname = 'vector';
-- 应返回 vector 扩展信息
```

### 8.2 向量表创建

已在 `§2.5 知识库服务` 中定义 `cf_knowledge_chunks` 表，包含 `vector` 列类型为 `vector(1536)`。

**维度说明**：
- OpenAI `text-embedding-3-small`：1536 维
- OpenAI `text-embedding-ada-002`：1536 维
- 本地模型（如 BGE-M3）：1024 维（需调整表定义）

如需修改维度：
```sql
-- 修改列类型（需要重建索引）
ALTER TABLE cf_knowledge_chunks
    ALTER COLUMN vector TYPE vector(1024);
```

### 8.3 索引优化

**HNSW 索引参数调优**：

| 场景 | m | ef_construction | 说明 |
|------|---|----------------|------|
| 快速原型 | 8 | 32 | 建索引快，精度略低 |
| 生产推荐 | 16 | 64 | 平衡精度与速度（默认）|
| 高精度 | 32 | 200 | 精度高，建索引慢，内存占用大 |

**创建索引示例**：
```sql
-- 查看当前索引
\d cf_knowledge_chunks

-- 删除旧索引
DROP INDEX IF EXISTS idx_chunk_vector;

-- 创建新索引（推荐参数）
CREATE INDEX idx_chunk_vector ON cf_knowledge_chunks
USING hnsw (vector vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

**索引大小估算**：
```sql
-- 查看索引大小
SELECT pg_size_pretty(pg_relation_size('idx_chunk_vector'));

-- 规则：索引大小 ≈ 向量数量 × 维度 × 4字节 × 1.2（HNSW 开销）
-- 例如：100万向量 × 1536维 × 4字节 ≈ 5.8 GB
```

### 8.4 Go 端操作示例

**依赖**：
```go
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
)
```

**插入向量**：
```go
vec := pgtype.Vector{
    Dimensions: 1536,
    Elements:   myVector[:],  // [1536]float64
}

_, err := db.Exec(ctx, `
    INSERT INTO cf_knowledge_chunks (id, document_id, knowledge_base_id, chunk_index, content, vector)
    VALUES ($1, $2, $3, $4, $5, $6)
`, id, docID, kbID, chunkIndex, content, vec)
```

**检索向量**：
```go
rows, err := db.Query(ctx, `
    SELECT id, content, 1 - (vector <=> $1) AS similarity
    FROM cf_knowledge_chunks
    WHERE knowledge_base_id = $2
      AND 1 - (vector <=> $1) >= 0.7
    ORDER BY vector <=> $1
    LIMIT 5
`, queryVec, kbID)
```

### 8.5 性能监控

```sql
-- 查看 pgvector 索引统计
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan AS index_scans,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched
FROM pg_stat_user_indexes
WHERE tablename = 'cf_knowledge_chunks';

-- 查看表大小
SELECT
    pg_size_pretty(pg_total_relation_size('cf_knowledge_chunks')) AS total,
    pg_size_pretty(pg_relation_size('cf_knowledge_chunks')) AS table_size,
    pg_size_pretty(pg_total_relation_size('idx_chunk_vector')) AS index_size;

-- 查询 QPS
SELECT
    date_trunc('minute', created_at) AS minute,
    COUNT(*) AS queries_per_min
FROM pg_stat_statements
WHERE query LIKE '%cf_knowledge_chunks%'
GROUP BY 1
ORDER BY 1 DESC
LIMIT 10;
```

### 8.6 备份与恢复

```bash
# 导出向量���（pg_dump 支持 vector 列类型）
pg_dump -U cogniforge -d cogniforge -t cf_knowledge_chunks > chunks.sql

# 恢复
psql -U cogniforge -d cogniforge < chunks.sql
```

**注意**：pgvector 数据以二进制格式存储，pg_dump 能正确处理。

### 8.7 常见问题

**Q1：向量检索很慢怎么办？**
- 检查 HNSW 索引是否存在：`\d cf_knowledge_chunks`
- 增加 `ef_search` 参数：`SET hnsw.ef_search = 100;`
- 确保 `vector` 列有索引，且查询使用了索引（EXPLAIN ANALYZE）

**Q2：插入向量时提示 "column vector is of type vector but expression is of type double precision[]"？**
- 需要使用 `pgtype.Vector` 类型，不能直接传 `[]float64`

**Q3：pgvector 支持的最大向量维度？**
- 最大 16000 维（PostgreSQL 限制），足够 OpenAI 的 1536 维

**Q4：可以同时存储多个 embedding 模型吗？**
- 可以，添加 `model_version` 字段区分，或创建不同表

---

## 9. 未来扩展：专用向量数据库

### 9.1 Milvus 集成（可选扩展点）

**适用场景**：向量数量 > 1000 万，或需要更高性能（< 10ms P99）

**迁移策略**：
1. 抽象向量存储接口：
```go
type VectorStore interface {
    Insert(ctx context.Context, chunks []Chunk, vectors [][]float64) error
    Search(ctx context.Context, kbID string, query []float64, topK int, threshold float64) ([]Result, error)
    Delete(ctx context.Context, documentID string) error
}
```

2. 实现两个版本：
   - `PgVectorStore`：当前默认（pgvector）
   - `MilvusStore`：未来扩展（Milvus）

3. 配置切换：
```yaml
vector:
  provider: pgvector  # 或 milvus、qdrant
  pgvector:
    dsn: postgres://...
  milvus:
    host: localhost
    port: 19530
```

**数据迁移脚本**（pgvector → Milvus）：
```python
# 读取 pgvector 数据
cur.execute("SELECT id, document_id, vector FROM cf_knowledge_chunks")
rows = cur.fetchall()

# 插入 Milvus
milvus.insert(
    collection_name="cf_knowledge_vectors",
    vectors=[row[2] for row in rows],
    ids=[row[0] for row in rows],
    metadata=[{"document_id": row[1]} for row in rows]
)
```

---

**文档版本**: v1.0  
**最后更新**: 2026-03-16（pgvector 更新：2026-04-11）  
**维护团队**: CogniForge 数据库团队
