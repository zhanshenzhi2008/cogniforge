# CogniForge API 接口设计文档

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-04-09 | v1.1 | 新增文档上传接口、语义检索接口实现说明 | orjrs |
| 2026-03-16 | v1.0 | 初始版本 | orjrs |

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
认证: API密钥
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
接口组: /v1/embeddings

POST /v1/embeddings
描述: 生成文本向量
认证: API密钥
请求体:
  {
    "model": "text-embedding-3-small",
    "input": "要向量化的文本"
  }
响应:
  {
    "object": "list",
    "data": [
      {
        "object": "embedding",
        "embedding": [0.123, -0.456, ...],
        "index": 0
      }
    ],
    "model": "text-embedding-3-small",
    "usage": {
      "prompt_tokens": 10,
      "total_tokens": 10
    }
  }
```

### 3.3 模型列表

```yaml
接口组: /v1/models

GET /v1/models
描述: 获取可用模型列表
响应:
  {
    "data": [
      {
        "id": "gpt-4o",
        "object": "model",
        "created": 1234567890,
        "owned_by": "openai",
        "name": "GPT-4o",
        "description": "最新一代GPT-4模型"
      }
    ]
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
描述: 删除文档
认证: JWT
```

### 6.3 检索

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

## 10. 错误码定义

| 错误码 | 类型 | 描述 |
|-------|------|------|
| 400 | invalid_request_error | 请求参数错误 |
| 401 | authentication_error | 认证失败 |
| 403 | permission_error | 权限不足 |
| 404 | not_found_error | 资源不存在 |
| 429 | rate_limit_error | 请求频率超限 |
| 500 | server_error | 服务器内部错误 |
| 503 | service_unavailable | 服务不可用 |

---

**文档版本**: v1.0  
**最后更新**: 2026-03-16  
**维护团队**: CogniForge API 团队
