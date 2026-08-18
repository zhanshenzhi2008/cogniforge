# CogniForge API 接口设计文档

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-08-18 | v1.7 | 配额接口 /quota/*；聊天须登录；新增业务码 5016 | orjrs |
| 2026-08-16 | v1.6 | 未配置默认模型时对话返回 4010，不再 mock | orjrs |
| 2026-08-16 | v1.5 | DeepSeek 下拉增加 V4，同时保留 deepseek-chat / reasoner | orjrs |
| 2026-08-16 | v1.4 | 新增登录用户聊天历史 CRUD：/api/v1/conversations | orjrs |
| 2026-08-15 | v1.3 | 新增 POST /api/v1/embeddings；Python RAG 回调 Go，不自己拿 Key | orjrs |
| 2026-08-15 | v1.2 | GET /v1/models 改为返回已启用供应商的 default_model（不再写死 GPT 列表） | orjrs |
| 2026-04-09 | v1.1 | 新增文档上传接口、语义检索接口实现说明 | orjrs |
| 2026-03-16 | v1.0 | 初始版本 | orjrs |

## [变更] 配额与聊天须登录（2026-08-18）

- **变更原因**：`/chat/stream` 目前公开，Playground 可无限刷共用 Key
- **详细设计**：`docs/01-requirements/02-quota-design.md`
- **接口**：`GET /api/v1/quota/me`、`GET /api/v1/quota/usage`、`GET/PUT /api/v1/admin/quota/policy`、`PUT /api/v1/admin/quota/users/:id`
- **行为变更**：浏览器调用 `/chat/stream`、`/chat/completions`、`/agents/:id/chat` 必须 JWT，超限返回 HTTP 429 / `code=5016`
- **不改**：`/embeddings` 仍供内网 Python 回调，不算用户额度
- **变更前 vs 变更后**：~~匿名可打对话~~（2026-08-18）→ 未登录 401；额度用尽 5016；上游供应商没钱仍是 4008

## [变更] 未配置默认模型时返回 4010（2026-08-16）

- **变更原因**：缺 Key / 缺默认模型时接口假装在流式输出 mock 文本
- **包含代码**：`internal/chat/handler.go`、`internal/chat/service.go`、`internal/response/response.go`
- **接口**：`POST /api/v1/chat/stream`、`POST /api/v1/chat/completions`、`POST /api/v1/agents/:id/chat`、`POST /api/v1/embeddings`
- **变更后**：HTTP 503，`code=4010`（`CodeNoActiveProvider`），`message` 为「请先到「模型」页填写 API Key，并设置一个默认模型后再对话」
- **不再**：把 mock 句子写进 SSE `choices[].delta.content`

## [变更] DeepSeek 下拉增加 V4（2026-08-16）

- **变更原因**：对话里要能选 V4，但日常仍用更便宜的 `deepseek-chat`
- **包含代码**：`internal/model/provider.go`、`internal/provider/service.go`；Web 模型页下拉
- **变更后**：列表为 `deepseek-chat`、`deepseek-reasoner`、`deepseek-v4-flash`、`deepseek-v4-pro`；**不**自动改掉库里的默认模型

~~先前写过「启动时把 chat 迁成 flash」——已撤销，不删旧选项。~~（2026-08-16）

## [变更] Playground 对话历史接口（2026-08-16）

- **变更原因**：对话页需要历史列表；参数不再占左侧。消息按用户落库，刷新后还能打开
- **包含代码**：`internal/model/conversation.go`、`internal/chat/conversation.go`、`internal/chat/conversation_handler.go`
- **影响范围**：`GET/POST /api/v1/conversations`、`GET/PUT/DELETE /api/v1/conversations/:id`（需 JWT）

### 变更前 vs 变更后

- **变更前**：只有 `/chat/stream`，消息只在浏览器内存里
- **变更后**：登录用户可增删改查自己的对话；列表不带 messages，详情才带全文

## [变更] 新增 Embeddings 内部接口（2026-08-15）

- **变更原因**：Python RAG 不应自己持有供应商 Key；与聊天共用 `ai_providers`
- **包含代码**：`internal/chat/handler.go`、`internal/chat/service.go`
- **影响范围**：`POST /api/v1/embeddings`（公开、无登录，供内网 Python 调用）

### 变更前 vs 变更后

- **变更前**：文档写了 `/v1/embeddings` 但 Go 未实现
- **变更后**：Go 用当前启用供应商转发上游 `/v1/embeddings`，响应包在统一 `{code,data}` 里

## [变更] 模型列表改为数据库配置（2026-08-15）

- **变更原因**：Playground/Agent 下拉仍写死 GPT，且 `AI_*` 环境变量与「模型」页重复
- **包含代码**：`internal/chat/service.go`、`internal/agent/handler.go`、`configs/config.yaml`
- **影响范围**：`GET /api/v1/models`、Agent 缺省模型；密钥只走 `ai_providers`

### 变更前 vs 变更后

- **变更前**：`ListModels` 内置一串 gpt-3.5；Agent 缺省读 `AI_DEFAULT_MODEL`
- **变更后**：只返回已启用供应商的 `default_model`；无环境变量兜底

## [变更] 文档上传与检索接口实现（2026-04-09）

变更原因：补充文档上传和语义检索接口的实现细节
包含代码：`internal/handler/knowledge.go`
影响范围：API 设计文档

### 变更前

- 文档上传接口仅有占位说明
- 检索接口未实现

### 变更后

- 文档上传：支持 multipart/form-data，支持 PDF/TXT/MD/DOCX/HTML
- 检索接口：基于关键词的文本检索，支持相似度评分

### 关键差异

- **文档上传**：`POST /api/v1/knowledge/:id/documents`
- **检索**：`POST /api/v1/knowledge/:id/search`

## 1. API 设计规范

### 1.1 设计原则

| 原则 | 描述 |
|-----|------|
| **RESTful** | 资源导向的URL设计 |
| **版本控制** | URL路径包含版本号 (v1) |
| **标准化错误** | 统一的错误响应格式 |
| **OpenAPI 3.0** | 使用OpenAPI规范文档化 |
| **OpenAI兼容** | 核心接口兼容OpenAI API |

### 1.2 认证方式

```yaml
认证方式:
  - API密钥: Header "Authorization: Bearer {api_key}"
  - JWT Token: Header "Authorization: Bearer {jwt_token}"
  
速率限制:
  - 免费版: 60请求/分钟
  - 专业版: 600请求/分钟
  - 企业版: 自定义
```

### 1.3 错误响应格式

```json
{
  "error": {
    "message": "错误描述",
    "type": "invalid_request_error",
    "code": "400",
    "param": "具体参数名"
  }
}
```

---

## 2. 认证接口

### 2.1 用户认证

```yaml
接口组: /v1/auth

POST /v1/auth/register
描述: 用户注册
请求体:
  {
    "email": "user@example.com",
    "password": "password123",
    "name": "张三"
  }
响应:
  {
    "id": "user_xxx",
    "email": "user@example.com",
    "name": "张三",
    "created_at": "2026-03-16T10:00:00Z"
  }

---

POST /v1/auth/login
描述: 用户登录
请求体:
  {
    "email": "user@example.com",
    "password": "password123"
  }
响应:
  {
    "access_token": "eyJxxx",
    "token_type": "Bearer",
    "expires_in": 86400
  }

---

POST /v1/auth/logout
描述: 用户登出
认证: 需要JWT
响应: 204 No Content
```

### 2.2 API密钥管理

```yaml
接口组: /v1/api-keys

GET /v1/api-keys
描述: 获取API密钥列表
认证: 需要JWT
响应:
  {
    "data": [
      {
        "id": "key_xxx",
        "name": "我的密钥",
        "prefix": "sk-cf-xxxx",
        "created_at": "2026-03-16T10:00:00Z",
        "last_used_at": "2026-03-16T12:00:00Z"
      }
    ]
  }

---

POST /v1/api-keys
描述: 创建API密钥
认证: 需要JWT
请求体:
  {
    "name": "生产环境密钥",
    "expires_in": 7776000  # 90天，单位秒
  }
响应:
  {
    "id": "key_xxx",
    "name": "生产环境密钥",
    "secret": "sk-cf-xxxxxx",  # 仅返回一次
    "created_at": "2026-03-16T10:00:00Z"
  }

---

DELETE /v1/api-keys/{key_id}
描述: 撤销API密钥
认证: 需要JWT
响应: 204 No Content
```

---

## 3. 模型网关接口

### 3.1 聊天补全

```yaml
接口组: /v1/chat

POST /v1/chat/completions
描述: 聊天补全（OpenAI兼容）
认证: JWT（2026-08-18 起浏览器必须登录；匿名调用返回 401）
配额: 计入用户额度；用尽 HTTP 429 / code=5016；每分钟过快 code=5014
注意: Python 内部回调走单独约定，不算 Playground 用户额度，见配额设计文档
请求体:
  {
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的助手"},
      {"role": "user", "content": "你好"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000,
    "top_p": 1.0,
    "frequency_penalty": 0.0,
    "presence_penalty": 0.0,
    "stream": false,
    "stop": null,
    "tools": null,
    "tool_choice": null
  }
响应:
  {
    "id": "chatcmpl-xxx",
    "object": "chat.completion",
    "created": 1234567890,
    "model": "gpt-4o",
    "choices": [
      {
        "index": 0,
        "message": {
          "role": "assistant",
          "content": "你好！有什么可以帮助你的吗？"
        },
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 20,
      "completion_tokens": 50,
      "total_tokens": 70
    }
  }

---

POST /v1/chat/completions (流式)
描述: 流式聊天补全
请求体:
  {"model": "gpt-4o", "messages": [...], "stream": true}
响应: Server-Sent Events (SSE)
  data: {"id":"chatcmpl-xxx","choices":[{"index":0,"delta":{"role":"assistant","content":"你"},"finish_reason":null}]}
  data: {"id":"chatcmpl-xxx","choices":[{"index":0,"delta":{"content":"好"},"finish_reason":null}]}
  data: {"id":"chatcmpl-xxx","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
  data: [DONE]
```

### 3.2 Embeddings

```yaml
接口组: /api/v1/embeddings

POST /api/v1/embeddings
描述: 生成文本向量。用当前启用的 ai_providers 调上游 /v1/embeddings（与聊天同一套配置）
认证: 无（内网 Python RAG 回调；勿对公网暴露 Go 8080）
请求体:
  {
    "model": "可选，空则用供应商 default_model",
    "input": "要向量化的文本，或字符串数组"
  }
响应（统一信封）:
  {
    "code": 2000,
    "data": {
      "object": "list",
      "data": [
        {
          "object": "embedding",
          "embedding": [0.123, -0.456],
          "index": 0
        }
      ],
      "model": "text-embedding-3-small",
      "usage": {
        "prompt_tokens": 10,
        "total_tokens": 10
      }
    }
  }
```

~~原规划的独立 OpenAI 风格 `/v1/embeddings`（需 API 密钥认证）尚未作为对外网关实现。~~（2026-08-15）

### 3.3 模型列表

```yaml
接口组: /v1/models

GET /v1/models
描述: 获取可用模型列表（来自已启用的 ai_providers.default_model）
认证: 无
响应:
  {
    "code": 2000,
    "data": {
      "models": [
        {"id": "deepseek-chat", "name": "deepseek-chat"},
        {"id": "deepseek-reasoner", "name": "deepseek-reasoner"},
        {"id": "deepseek-v4-flash", "name": "deepseek-v4-flash"},
        {"id": "deepseek-v4-pro", "name": "deepseek-v4-pro"}
      ]
    }
  }

---

GET /v1/models/{model_id}
描述: 获取模型详情
响应:
  {
    "id": "gpt-4o",
    "object": "model",
    "created": 1234567890,
    "owned_by": "openai",
    "name": "GPT-4o",
    "description": "最新一代GPT-4模型",
    "context_window": 128000,
    "max_output_tokens": 16384,
    "pricing": {
      "input": 0.005,
      "output": 0.015
    }
  }
```

### 3.4 聊天历史（Playground）

```yaml
接口组: /api/v1/conversations
认证: JWT（只看得到当前登录用户自己的对话）

GET /api/v1/conversations
描述: 对话列表（不含 messages，按 updated_at 倒序，最多 100 条）
响应:
  {
    "code": 2000,
    "data": [
      {
        "id": "uuid",
        "title": "今天天气怎么样",
        "agent_id": "",
        "model": "deepseek-chat",
        "created_at": "2026-08-16T00:00:00Z",
        "updated_at": "2026-08-16T00:00:00Z"
      }
    ]
  }

POST /api/v1/conversations
描述: 新建对话。title 为空时用第一条用户消息前 40 字
请求体:
  {
    "title": "可选",
    "agent_id": "可选，空表示通用对话",
    "model": "deepseek-chat",
    "messages": [
      {"id": "uuid", "role": "user", "content": "你好", "time": "ISO8601"}
    ]
  }

GET /api/v1/conversations/{id}
描述: 对话详情（含 messages）
响应: { "code": 2000, "data": { "id": "...", "messages": [...], "...": "..." } }

PUT /api/v1/conversations/{id}
描述: 更新标题 / 模型 / Agent / 消息全文
请求体: 字段均可选；messages 为整份覆盖，不是增量

DELETE /api/v1/conversations/{id}
描述: 软删除
```

~~规划中的 `cf_agent_conversations`（必须挂 Agent）未作为 Playground 历史表落地。~~（2026-08-16）

---

## 4. Agent接口

### 4.1 Agent管理

```yaml
接口组: /v1/agents

GET /v1/agents
描述: 获取Agent列表
认证: JWT
响应:
  {
    "data": [
      {
        "id": "agent_xxx",
        "name": "客服助手",
        "description": "处理客户咨询",
        "model": "gpt-4o",
        "status": "active",
        "created_at": "2026-03-16T10:00:00Z",
        "updated_at": "2026-03-16T10:00:00Z"
      }
    ]
  }

---

POST /v1/agents
描述: 创建Agent
认证: JWT
请求体:
  {
    "name": "客服助手",
    "description": "处理客户咨询",
    "model": "gpt-4o",
    "system_prompt": "你是一个专业的客服助手...",
    "tools": ["search_kb", "create_ticket"],
    "memory": {
      "type": "short_term",
      "max_turns": 10
    },
    "guardrails": {
      "input_filter": true,
      "output_filter": true
    }
  }
响应:
  {
    "id": "agent_xxx",
    "name": "客服助手",
    "description": "处理客户咨询",
    "model": "gpt-4o",
    "status": "active",
    "created_at": "2026-03-16T10:00:00Z"
  }

---

GET /v1/agents/{agent_id}
描述: 获取Agent详情
认证: JWT

---

PUT /v1/agents/{agent_id}
描述: 更新Agent
认证: JWT

---

DELETE /v1/agents/{agent_id}
描述: 删除Agent
认证: JWT
```

### 4.2 Agent对话

```yaml
接口组: /v1/agents/{agent_id}

POST /v1/agents/{agent_id}/chat
描述: 与Agent对话
认证: API密钥
请求体:
  {
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "stream": false
  }
响应:
  {
    "id": "agent_chat_xxx",
    "agent_id": "agent_xxx",
    "message": {
      "role": "assistant",
      "content": "你好！有什么可以帮助你的？"
    },
    "usage": {
      "total_tokens": 100
    }
  }

---

POST /v1/agents/{agent_id}/chat (流式)
描述: 流式与Agent对话
请求体: {"messages": [...], "stream": true}
响应: SSE流
```

---

## 5. 工作流接口

### 5.1 工作流管理

```yaml
接口组: /v1/workflows

GET /v1/workflows
描述: 获取工作流列表
认证: JWT

---

POST /v1/workflows
描述: 创建工作流
认证: JWT
请求体:
  {
    "name": "客服工作流",
    "description": "智能客服流程",
    "nodes": [
      {
        "id": "node_1",
        "type": "start",
        "position": {"x": 0, "y": 0}
      },
      {
        "id": "node_2",
        "type": "llm",
        "model": "gpt-4o",
        "prompt": "你是客服助手...",
        "position": {"x": 100, "y": 0}
      },
      {
        "id": "node_3",
        "type": "end",
        "position": {"x": 200, "y": 0}
      }
    ],
    "edges": [
      {"source": "node_1", "target": "node_2"},
      {"source": "node_2", "target": "node_3"}
    ]
  }
响应:
  {
    "id": "workflow_xxx",
    "name": "客服工作流",
    "version": 1,
    "status": "draft",
    "created_at": "2026-03-16T10:00:00Z"
  }

---

GET /v1/workflows/{workflow_id}
描述: 获取工作流详情

---

PUT /v1/workflows/{workflow_id}
描述: 更新工作流
认证: JWT

---

DELETE /v1/workflows/{workflow_id}
描述: 删除工作流
认证: JWT
```

### 5.2 工作流执行

```yaml
接口组: /v1/workflows/{workflow_id}

POST /v1/workflows/{workflow_id}/execute
描述: 执行工作流
认证: API密钥
请求体:
  {
    "input": {"query": "你好"},
    "sync": true
  }
响应:
  {
    "execution_id": "exec_xxx",
    "status": "completed",
    "output": {"result": "你好，我是..."},
    "tokens_used": 500,
    "duration_ms": 2000
  }

---

POST /v1/workflows/{workflow_id}/execute (异步)
请求体:
  {
    "input": {"query": "你好"},
    "sync": false,
    "webhook_url": "https://example.com/callback"
  }
响应:
  {
    "execution_id": "exec_xxx",
    "status": "running"
  }

---

GET /v1/workflows/{workflow_id}/executions
描述: 获取执行历史
认证: JWT

---

GET /v1/executions/{execution_id}
描述: 获取执行详情
认证: JWT
响应:
  {
    "id": "exec_xxx",
    "workflow_id": "workflow_xxx",
    "status": "completed",
    "input": {"query": "你好"},
    "output": {"result": "你好，我是..."},
    "node_executions": [
      {
        "node_id": "node_1",
        "status": "completed",
        "input": {},
        "output": {},
        "duration_ms": 10
      },
      {
        "node_id": "node_2",
        "status": "completed",
        "input": {"query": "你好"},
        "output": {"result": "你好，我是..."},
        "duration_ms": 1990
      }
    ],
    "tokens_used": 500,
    "duration_ms": 2000,
    "created_at": "2026-03-16T10:00:00Z",
    "completed_at": "2026-03-16T10:00:02Z"
  }
```

---

## 6. 知识库接口

### 6.1 知识库管理

```yaml
接口组: /api/v1/knowledge

GET /api/v1/knowledge
描述: 获取知识库列表
认证: JWT
响应:
  {
    "code": 2000,
    "data": [
      {
        "id": "kb_xxx",
        "name": "产品文档",
        "description": "产品帮助文档",
        "vector_db": "chroma",
        "embedding_model": "text-embedding-ada-002",
        "doc_count": 5,
        "status": "active",
        "created_at": "2026-03-16T10:00:00Z"
      }
    ]
  }

---

POST /api/v1/knowledge
描述: 创建知识库
认证: JWT
请求体:
  {
    "name": "产品文档",
    "description": "产品帮助文档",
    "vector_db": "chroma",
    "embedding_model": "text-embedding-ada-002"
  }
响应:
  {
    "code": 2001,
    "data": {
      "id": "kb_xxx",
      "name": "产品文档",
      "description": "产品帮助文档",
      "doc_count": 0,
      "status": "active",
      "created_at": "2026-03-16T10:00:00Z"
    }
  }

---

GET /api/v1/knowledge/{kb_id}
描述: 获取知识库详情
认证: JWT

---

PUT /api/v1/knowledge/{kb_id}
描述: 更新知识库
认证: JWT

---

DELETE /api/v1/knowledge/{kb_id}
描述: 删除知识库（软删除）
认证: JWT
```

### 6.2 文档管理

```yaml
接口组: /api/v1/knowledge/{kb_id}

POST /api/v1/knowledge/{kb_id}/documents
描述: 上传文档
认证: JWT
请求: multipart/form-data
  - file: PDF/TXT/MD/DOCX/HTML 文件（必填）
响应:
  {
    "code": 2001,
    "data": {
      "id": "doc_xxx",
      "knowledge_base_id": "kb_xxx",
      "name": "产品手册.pdf",
      "file_name": "产品手册.pdf",
      "file_size": 1024000,
      "file_type": "pdf",
      "file_path": "uploads/documents/xxx/kb_xxx/doc_xxx.pdf",
      "status": "pending",
      "chunk_count": 0,
      "vector_count": 0,
      "created_at": "2026-04-09T10:00:00Z"
    }
  }

---

GET /api/v1/knowledge/{kb_id}/documents
描述: 获取文档列表
认证: JWT
响应:
  {
    "code": 2000,
    "data": [
      {
        "id": "doc_xxx",
        "name": "产品手册.pdf",
        "status": "completed",
        "chunk_count": 10,
        "file_size": 1024000,
        "created_at": "2026-04-09T10:00:00Z"
      }
    ]
  }

---

DELETE /api/v1/knowledge/{kb_id}/documents/{doc_id}
描述: 删除文档（软删除，同时删除向量）
认证: JWT

---

### 6.3 文档处理状态

```yaml
GET /api/v1/knowledge/{kb_id}/documents/{doc_id}/status
描述: 查询文档处理状态（pending/processing/completed/failed）
认证: JWT
响应:
  {
    "code": 2000,
    "data": {
      "id": "doc_xxx",
      "status": "completed",  // pending, processing, completed, failed
      "chunk_count": 15,
      "vector_count": 15,
      "error_message": null,  // 失败时包含错误信息
      "updated_at": "2026-04-11T10:30:00Z"
    }
  }
```

---

### 6.4 检索

```yaml
POST /api/v1/knowledge/{kb_id}/search
描述: 语义检索（基于关键词）
认证: JWT
请求体:
  {
    "query": "如何重置密码",
    "top_k": 5,
    "min_score": 0.3
  }
响应:
  {
    "code": 2000,
    "data": {
      "results": [
        {
          "document_id": "doc_xxx",
          "document_name": "产品手册.pdf",
          "chunk_id": "doc_xxx_chunk_0",
          "content": "重置密码步骤：...",
          "score": 0.85
        }
      ],
      "total": 1,
      "query": "如何重置密码",
      "duration_ms": 125
    }
  }
```

---

## 7. 微调训练接口

### 7.1 数据集管理

```yaml
接口组: /v1/datasets

GET /v1/datasets
描述: 获取数据集列表
认证: JWT

---

POST /v1/datasets
描述: 上传数据集
认证: JWT
请求: multipart/form-data
  - file: JSONL文件
  - name: 数据集名称
响应:
  {
    "id": "dataset_xxx",
    "name": "客服对话数据",
    "status": "validating",
    "sample_count": 1000,
    "created_at": "2026-03-16T10:00:00Z"
  }

---

GET /v1/datasets/{dataset_id}
描述: 获取数据集详情
响应:
  {
    "id": "dataset_xxx",
    "name": "客服对话数据",
    "status": "ready",
    "sample_count": 1000,
    "validation_result": {
      "valid": true,
      "errors": []
    },
    "preview": [
      {"messages": [...]}
    ]
  }
```

### 7.2 训练任务

```yaml
接口组: /v1/fine-tunes

POST /v1/fine-tunes
描述: 创建训练任务
认证: JWT
请求体:
  {
    "model": "gpt-4o-mini",
    "dataset_id": "dataset_xxx",
    "name": "客服模型v1",
    "hyperparameters": {
      "epochs": 3,
      "batch_size": "auto",
      "learning_rate_multiplier": 1.0
    }
  }
响应:
  {
    "id": "fine_tune_xxx",
    "model": "gpt-4o-mini",
    "dataset_id": "dataset_xxx",
    "status": "queued",
    "created_at": "2026-03-16T10:00:00Z"
  }

---

GET /v1/fine-tunes
描述: 获取训练任务列表

---

GET /v1/fine-tunes/{fine_tune_id}
描述: 获取训练任务详情
响应:
  {
    "id": "fine_tune_xxx",
    "model": "gpt-4o-mini",
    "status": "completed",
    "result": {
      "training_loss": 0.5,
      "eval_loss": 0.3
    },
    "fine_tuned_model": "gpt-4o-mini:ft-xxx",
    "created_at": "2026-03-16T10:00:00Z",
    "completed_at": "2026-03-16T12:00:00Z"
  }

---

POST /v1/fine-tunes/{fine_tune_id}/cancel
描述: 取消训练任务
```

---

## 8. 监控接口

### 8.1 用量统计

```yaml
接口组: /v1/usage

GET /v1/usage
描述: 获取用量统计
认证: JWT
参数:
  - start_date: 2026-01-01
  - end_date: 2026-01-31
  - granularity: daily|monthly
响应:
  {
    "data": [
      {
        "date": "2026-01-01",
        "requests": 1000,
        "input_tokens": 500000,
        "output_tokens": 300000,
        "cost": 15.00
      }
    ],
    "summary": {
      "total_requests": 30000,
      "total_input_tokens": 15000000,
      "total_output_tokens": 9000000,
      "total_cost": 450.00
    }
  }
```

### 8.2 请求日志

```yaml
接口组: /v1/logs

GET /v1/logs
描述: 获取请求日志
认证: JWT
参数:
  - start_time: 2026-01-01T00:00:00Z
  - end_time: 2026-01-01T23:59:59Z
  - model: gpt-4o
  - status: success|error
  - limit: 100
  - offset: 0
响应:
  {
    "data": [
      {
        "id": "req_xxx",
        "model": "gpt-4o",
        "prompt_tokens": 100,
        "completion_tokens": 200,
        "status": "success",
        "latency_ms": 1500,
        "cost": 0.005,
        "created_at": "2026-01-01T10:00:00Z"
      }
    ],
    "total": 1000,
    "limit": 100,
    "offset": 0
  }

---

GET /v1/logs/{log_id}
描述: 获取日志详情
响应:
  {
    "id": "req_xxx",
    "model": "gpt-4o",
    "messages": [...],
    "response": {...},
    "usage": {...},
    "status": "success",
    "latency_ms": 1500,
    "created_at": "2026-01-01T10:00:00Z"
  }
```

### 8.3 配额与用量（2026-08-18 设计，待落地）

详细规则见 `docs/01-requirements/02-quota-design.md`。统一包在 `{ code, message, trace_id, data }`。

```yaml
GET /api/v1/quota/me
描述: 当前用户剩余额度（Playground / Dashboard 轮询）
认证: JWT
响应 data:
  {
    "unlimited": false,
    "day": {
      "requests_used": 12,
      "requests_limit": 30,
      "tokens_used": 32000,
      "tokens_limit": 100000,
      "resets_at": "2026-08-19T00:00:00+08:00"
    },
    "month": {
      "tokens_used": 120000,
      "tokens_limit": 1000000,
      "resets_at": "2026-09-01T00:00:00+08:00"
    },
    "warn": false
  }

---

GET /api/v1/quota/usage
描述: 用量序列，供柱状图 / 饼图
认证: JWT
参数:
  - range: 7d | 30d
  - metric: requests | tokens
  - user_id: 仅 admin，看指定用户
  - scope: self | all（all 仅 admin）
响应 data:
  {
    "points": [{ "date": "2026-08-18", "requests": 12, "tokens": 32000 }],
    "by_model": [{ "model": "deepseek-chat", "tokens": 24000 }],
    "top_users": [{ "user_id": "...", "name": "...", "tokens": 180000 }]
  }

---

GET /api/v1/admin/quota/policy
PUT /api/v1/admin/quota/policy
描述: 读/改全站默认限额
认证: JWT + role=admin
请求体:
  {
    "daily_requests": 30,
    "daily_tokens": 100000,
    "monthly_tokens": 1000000,
    "rpm": 8,
    "admin_unlimited": true
  }

---

PUT /api/v1/admin/quota/users/:id
描述: 覆盖某用户限额；字段传 null 表示取消覆盖、回到默认
认证: JWT + role=admin
```

超限时对话接口：

| HTTP | code | 含义 |
|------|------|------|
| 401 | 5005 | 未登录 |
| 429 | 5014 | 每分钟过快（不扣每日次数） |
| 429 | 5016 | 日/月额度用尽（新增 `CodeUserQuotaExceeded`） |
| 502/503 | 4008 | 上游供应商自己没额度（不是本平台限额） |

---

## 9. 用户与组织接口

### 9.1 用户管理

```yaml
接口组: /v1/users

GET /v1/users
描述: 获取用户列表（仅管理员）
认证: JWT

---

GET /v1/users/me
描述: 获取当前用户信息
认证: JWT

---

PUT /v1/users/me
描述: 更新当前用户信息
认证: JWT

---

POST /v1/users
描述: 创建用户（仅管理员）
认证: JWT
```

### 9.2 组织管理

```yaml
接口组: /v1/organizations

GET /v1/organizations
描述: 获取组织列表
认证: JWT

---

POST /v1/organizations
描述: 创建组织
认证: JWT

---

GET /v1/organizations/{org_id}
描述: 获取组织详情

---

PUT /v1/organizations/{org_id}
描述: 更新组织信息

---

POST /v1/organizations/{org_id}/members
描述: 邀请成员
请求体:
  {
    "email": "user@example.com",
    "role": "developer"
  }
```

### 9.3 角色权限

```yaml
接口组: /v1/roles

GET /v1/roles
描述: 获取角色列表

---

GET /v1/roles/{role_id}
描述: 获取角色详情

---

POST /v1/roles
描述: 创建自定义角色
认证: JWT (仅管理员)
```

---

## 10. 个人设置接口

### 10.1 个人资料

```yaml
接口组: /api/v1/settings

GET /api/v1/settings/profile
描述: 获取当前用户个人资料
认证: JWT
响应:
  {
    "code": 2000,
    "data": {
      "id": "user-uuid",
      "email": "user@example.com",
      "name": "张三",
      "avatar_url": "https://cdn.example.com/avatars/user.png",
      "phone": "+86 13800138000",
      "timezone": "Asia/Shanghai",
      "locale": "zh-CN",
      "theme": "light",
      "email_verified": true,
      "created_at": "2026-03-16T10:00:00Z"
    }
  }

---

PUT /api/v1/settings/profile
描述: 更新个人资料
认证: JWT
请求体:
  {
    "name": "李四",
    "phone": "+86 13800138001",
    "timezone": "Asia/Shanghai",
    "locale": "zh-CN",
    "theme": "dark"
  }
响应:
  {
    "code": 2000,
    "data": {
      "updated": true
    }
  }
```

### 10.2 头像上传

```yaml
POST /api/v1/settings/avatar
描述: 上传头像
认证: JWT
请求: multipart/form-data
  - file: 图片文件（JPG/PNG/GIF，最大 2MB）
响应:
  {
    "code": 2000,
    "data": {
      "avatar_url": "https://cdn.example.com/avatars/user-abc123.png"
    }
  }
```

**处理逻辑**：
1. 验证文件大小（≤ 2MB）
2. 验证文件类型（仅图片）
3. 图片解码并裁剪为 200×200 正方形
4. 保存到对象存储（MinIO/S3）
5. 更新用户 `avatar_url` 字段

### 10.3 密码修改

```yaml
POST /api/v1/settings/password
描述: 修改密码
认证: JWT
请求体:
  {
    "current_password": "旧密码（必填）",
    "new_password": "新密码（至少8位，含大小写字母和数字）",
    "confirm_password": "确认密码（必须与new_password一致）"
  }
响应:
  {
    "code": 2000,
    "message": "密码修改成功，请重新登录"
  }
```

**验证规则**：
- 旧密码正确（bcrypt 比对）
- 新密码长度 ≥ 8
- 新密码包含大写字母、小写字母、数字
- 确认密码一致
- 修改成功后，使当前 Token 失效（强制重新登录）

### 10.4 会话管理

```yaml
GET /api/v1/settings/sessions
描述: 获取当前用户的登录会话列表
认证: JWT
响应:
  {
    "code": 2000,
    "data": {
      "sessions": [
        {
          "id": "session-uuid",
          "ip_address": "192.168.1.100",
          "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...",
          "device_info": {
            "os": "macOS",
            "browser": "Chrome 120",
            "device_type": "desktop"
          },
          "location": "上海, 中国",
          "last_active_at": "2026-04-11T10:30:00Z",
          "is_current": true,
          "created_at": "2026-04-10T08:00:00Z"
        }
      ]
    }
  }

---

DELETE /api/v1/settings/sessions/{session_id}
描述: 远程登出指定会话（设备）
认证: JWT
路径参数:
  - session_id: 会话 ID（从列表页获取）
响应:
  {
    "code": 2000,
    "message": "会话已登出"
  }
```

**会话创建时机**：
- 用户登录成功后，生成 JWT Token 的同时在 `cf_user_sessions` 表插入一条记录
- `session_id` 使用 JWT 的 `jti` 声明（JWT ID）
- 每次请求更新 `last_active_at`

---

## 11. 错误码定义

HTTP 层（兼容旧表）：

| 错误码 | 类型 | 描述 |
|-------|------|------|
| 400 | invalid_request_error | 请求参数错误 |
| 401 | authentication_error | 认证失败 |
| 403 | permission_error | 权限不足 |
| 404 | not_found_error | 资源不存在 |
| 429 | rate_limit_error | 请求频率超限 **或** 平台额度用尽 |
| 500 | server_error | 服务器内部错误 |
| 503 | service_unavailable | 服务不可用 |

业务码（以 `internal/response` 为准）：

| code | 常量 | 描述 |
|------|------|------|
| 4008 | CodeAIQuotaExhausted | 上游 AI 供应商额度用尽 |
| 4010 | CodeNoActiveProvider | 未配置默认模型或 Key |
| 5005 | CodeUnauthorized | 未登录 |
| 5014 | CodeRateLimitExceeded | 每分钟请求过快 |
| 5016 | CodeUserQuotaExceeded | 本平台用户日/月额度用尽（2026-08-18 新增） |

---

**文档版本**: v1.7  
**最后更新**: 2026-08-18  
**维护团队**: CogniForge API 团队
