# MCP + SKILL 设计方案

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-09-11 | v1.3 | 内置工具 web_search 接入 Tavily：支持 keyless 模式（无需 API Key）和 API Key 模式（每月 1000 次免费）；移除 BING_API_KEY / GOOGLE_API_KEY 方案 | orjrs |
| 2026-09-05 | v1.2 | 新增 SKILL 外部导入：支持 Claude Skills 格式 SKILL.md（Markdown 单文件 / ZIP 压缩包）；ZIP 支持批量导入多个 Skill；内置 URL 安全校验（协议白名单、内网 IP 阻断） | orjrs |
| 2026-09-04 | v1.1 | 修复：工具调用循环改为动态停止（无 tool_calls 即停）；SKILL 结构升级（meta/instructions/references/constraints/examples） | orjrs |
| 2026-09-04 | v1.0 | 立项：MCP + SKILL，让 Agent 能调用外部工具，可复用技能模板 | orjrs |

## [变更] 立项（2026-09-04）

- **变更原因**：现有 Agent 只有 System Prompt，没有真正调用外部工具的能力；Playground 也无法让模型访问实时网页/数据库/文件系统
- **包含代码**：无（本期先定方案）
- **影响范围**：Go `internal/mcp/`、`internal/skill/`；Python MCP 客户端；Web Agent 表单；PostgreSQL 新表；与阶段十四记忆互不干扰
- **变更前 vs 变更后**：~~Agent 只能对话~~（2026-09-04）→ Agent 可通过 MCP 调用外部工具；可绑定 SKILL 模板

## [变更] 工具调用循环 + SKILL 结构（2026-09-04）

- **工具调用**：循环改为动态停止——每轮发 LLM → 检查 tool_calls → 没有则直接返回，有则执行工具 → 继续下一轮；MaxToolCalls=5 作为硬上限防止无限循环；达到上限返回错误而非再发一次请求
- **SKILL 结构升级**：参照 Claude Skills / GPTs，分离 meta（名称/描述/版本/标签）、instructions（核心指令）、constraints（约束规则）、examples（Few-shot）、references（参考数据：URL/文件/内嵌文本）；Skill.BuildSystemPrompt() 负责拼接；Agent 绑定 SkillID 时优先用 Skill.Instructions 构建 system prompt

---

## 1. 两个概念

### 1.1 MCP（Model Context Protocol）

让 AI 模型能调用**外部工具**的协议。

简单理解：Agent 想说「帮我查一下今天天气」，MCP 就是让模型能真正调用天气 API 并拿到结果。

| | 现在（无 MCP） | 有 MCP 之后 |
|--|--|--|
| Agent 对话 | 只靠预训练知识 | 可实时查网页/数据库/文件系统 |
| Playground | 模型只能回答训练数据内的内容 | 可访问外部实时数据 |
| 工具调用 | 没有 | 模型自主决定调用哪个工具 |

**MCP 不是大模型本身**，它是一个协议，让大模型能调用外部服务。

### 1.2 SKILL（技能模板）

把「一套工具组合 + System Prompt + 预设配置」打包成一个**可复用模板**。

| | 现在 | 有 SKILL 之后 |
|--|--|--|
| 创建 Agent | 每次都要配 System Prompt、选工具 | 选一个 SKILL，一键带入 |
| 常用场景 | 「代码助手」每次新建都要重配 | 选「代码助手」SKILL，自动带工具+人设 |

---

## 2. 架构总览

```
用户发消息
    │
    ▼
Go  /agents/:id/chat
    │
    ├─ 1. 鉴权 + 配额（已有）
    ├─ 2. 短期/长期记忆组装（阶段十四，已有）
    ├─ 3. 工具列表注入：
    │     ├─ 查 Agent 绑定的 SKILL → 得到 MCP Servers
    │     ├─ 查 Agent 直接配置的 MCP Servers
    │     └─ 把工具列表以 JSON schema 格式注入 system prompt 或 tool_calls
    │
    ├─ 4. 发给上游 LLM（已有）
    │     └─ 模型决定调用 tool_calls
    │
    ├─ 5. 拦截 tool_calls：
    │     ├─ 解析工具名 + 参数
    │     ├─ 从 Registry 拿 MCP Server 连接
    │     ├─ 调用 MCP Server（HTTP / stdio）
    │     └─ 把工具结果塞回 messages
    │
    ├─ 6. 再次发给 LLM（拿完整结果）
    │
    └─ 7. 流式返回前端（已有）
```

**SKILL 是配置层，MCP 是执行层。**

---

## 3. MCP 模块设计

### 3.1 三种 MCP Server 类型

| 类型 | 说明 | 示例 |
|--|--|--|
| **内置（built-in）** | Python/Go 自带，无需用户提供地址 | `fetch`（网页抓取）、`web_search`（搜索）、`filesystem`（文件操作） |
| **远程 HTTP** | 用户提供 MCP Server HTTP URL + 认证 | 第三方 MCP Server（Zapier、Slack 等） |
| **本地 stdio** | 用户提供本地可执行文件路径 | Cursor MCP、Custom 本地服务 |

### 3.2 数据模型

#### McpServer 表（PostgreSQL）

```go
type McpServer struct {
    ID          string    `gorm:"primaryKey" json:"id"`
    UserID      string    `gorm:"not null;index" json:"user_id"`
    Name        string    `gorm:"size:255" json:"name"`      // 显示名，如"网页抓取"
    Type        string    `gorm:"size:50" json:"type"`       // "built-in" | "http" | "stdio"
    URL         string    `gorm:"size:500" json:"url"`        // HTTP URL 或 stdio 路径
    AuthHeader  string    `gorm:"size:500" json:"auth_header"` // 可选：Authorization header
    Enabled     bool      `gorm:"default:true" json:"enabled"`
    Metadata    JSONBMap  `gorm:"type:jsonb" json:"metadata"` // 扩展配置
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### AgentMcpServer 关联表

```go
type AgentMcpServer struct {
    AgentID    string `gorm:"primaryKey" json:"agent_id"`
    ServerID   string `gorm:"primaryKey" json:"server_id"`
    Enabled    bool   `gorm:"default:true" json:"enabled"`
    Priority   int    `gorm:"default:0" json:"priority"` // 工具优先级
}
```

#### 内置 MCP Server（代码内置，无需数据库记录）

| 名称 | 类型 | 说明 |
|--|--|--|
| `fetch` | built-in | 网页抓取（调用 Go `httpclient`） |
| `web_search` | built-in | 搜索（调用 Bing/Google API） |
| `code_execute` | built-in | 代码执行（沙箱容器） |

### 3.3 MCP Registry（内存缓存）

```go
type Registry struct {
    mu      sync.RWMutex
    servers map[string]*McpServer       // serverID -> server
    tools   map[string]ToolDefinition     // "serverID:toolName" -> tool
}

// ToolDefinition：工具的 JSON Schema 描述
type ToolDefinition struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{}  `json:"input_schema"`  // JSON Schema
    ServerID    string                 `json:"server_id"`
}
```

- 启动时加载数据库中的 Server 到内存
- 内置工具直接注册到 Registry
- `Refresh()` 方法从数据库重新加载（Agent 编辑后调用）
- 所有缓存**不过期**，以数据库为唯一真相

### 3.4 MCP 客户端

#### HTTP MCP Server 客户端

```go
type HttpClient struct {
    server  *McpServer
    client  *http.Client
}

func (c *HttpClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)
```

调用格式（符合 MCP 协议）：

```
POST {url}/tools/{name}/call
Authorization: Bearer {auth_header}
Content-Type: application/json

{"arguments": {...}}
```

#### Built-in 客户端

```go
type BuiltInToolExecutor struct {
    fetch         *FetchTool
    webSearch     *WebSearchTool
    codeExecute   *CodeExecuteTool
}
```

### 3.5 ChatService 集成（工具调用循环）

```go
// MaxToolCalls = 5 硬上限，防止无限循环
// 实际次数由模型自主决定：没有 tool_calls → 直接返回，不继续发请求
const MaxToolCalls = 5

func (s *ChatService) ChatWithTools(req *ChatRequest, toolExecutor ToolExecutor) (*ChatResponse, error) {
    for i := 0; i < MaxToolCalls; i++ {
        resp, err := s.callLLM(req)           // 发 LLM
        if err != nil { return nil, err }

        toolCalls := extractToolCalls(resp)     // 提取 tool_calls
        if len(toolCalls) == 0 {
            return resp, nil                    // 模型决定结束，直接返回
        }

        for _, tc := range toolCalls {
            result, err := toolExecutor.Execute(tc) // 执行工具
            msg := "Error: " + err.Error()
            if err == nil { msg = result }
            req.Messages = append(req.Messages, newToolMsg(tc.ID, tc.Name, msg))
        }
        // 继续下一次循环
    }
    return nil, fmt.Errorf("tool call limit (%d) reached", MaxToolCalls)
}
```

---

## 4. SKILL 模块设计

### 4.1 概念

SKILL 参照 Claude Skills / GPTs，分以下几层：

```
SKILL = {
    // --- meta 元数据 ---
    name: "代码助手",
    description: "你是一个资深全栈工程师...",
    icon: "💻",
    category: "coding",
    tags: ["go", "python", "sql"],
    version: "v1.0",
    author: "系统",

    // --- instructions 核心指令 ---
    instructions: `请遵循以下工作方式：
1. 代码优先，注释清晰
2. 解释简洁，直接说明做什么
3. 错误排查给出根因和方向...`,

    // --- constraints 约束规则 ---
    constraints: [
        "禁止直接给出未经验证的代码",
        "SQL 操作必须防注入",
        "敏感信息不要出现在回复里",
    ],

    // --- examples Few-shot 示例 ---
    examples: [
        "User: 帮我写一个 Go 中间件...",
        "Assistant: 以下是完整实现...",
    ],

    // --- references 参考数据 ---
    references: [
        {"type":"url","title":"Go 标准库","url":"https://pkg.go.dev"},
    ],

    // --- 工具与配置 ---
    mcp_servers: ["code_execute"],
    model: "",
    memory_type: "short_term",
    memory_turns: 10,
}
```

**BuildSystemPrompt()** 拼接顺序：`description → constraints → instructions → examples → references`

### 4.2 数据模型

#### Skill 表

```go
type Skill struct {
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"` // 空 = 内置，全局可用
    Name         string    `json:"name"`
    Description  string    `json:"description"` // 角色描述，作为 system prompt 开头
    Icon         string    `json:"icon"`       // emoji

    Instructions string    `json:"instructions"` // 核心指令
    Examples     []string  `json:"examples"`     // Few-shot 示例
    References   []RefItem `json:"references"`   // 参考数据
    Constraints  []string  `json:"constraints"`  // 约束规则

    Model       string    `json:"model"`
    McpServers  []string  `json:"mcp_servers"`  // MCP Server ID 列表
    MemoryType  string    `json:"memory_type"`
    MemoryTurns int       `json:"memory_turns"`

    Category    string    `json:"category"`     // coding/data/research
    Tags        []string  `json:"tags"`        // ["go","sql"]
    Version     string    `json:"version"`     // v1.0
    Author      string    `json:"author"`
    IsBuiltIn   bool      `json:"is_built_in"`
    SortOrder   int       `json:"sort_order"`
    UsageCount  int       `json:"usage_count"`
}

type RefItem struct {
    Type    string `json:"type"`    // "url" | "file" | "text"
    Title   string `json:"title"`
    Content string `json:"content"`  // type=text 时
    URL     string `json:"url"`     // type=url 时
}
```

#### 内置 SKILL（数据库种子数据）

| 名称 | 描述 | 工具 | 约束（部分） |
|--|--|--|--|
| 代码助手 | 写代码、调试、Code Review | `code_execute` | 禁止未验证代码；SQL 防注入 |
| 网页助手 | 查资料、写报告 | `fetch`, `web_search` | 必须实际访问 URL；不能虚构引用 |
| 数据分析 | 跑 SQL、画图表 | `code_execute` | 结论先行；数据估算要说明假设 |

### 4.3 Agent 创建流程（绑定 SKILL）

```
用户点「创建 Agent」
    │
    ├─ 选择「从空白创建」→ 手动填所有字段
    │
    └─ 选择「从 SKILL 创建」→ 选择一个 SKILL
              │
              ├─ 自动带入：name / system_prompt / model / mcp_servers / memory_*
              └─ 用户可后续修改（SKILL 只作模板，不锁定）
```

### 4.4 SkillService

```go
type SkillService struct {
    db *gorm.DB
}

// ListSkills 列出当前用户可用的 SKILL（内置 + 自己创建的）
func (s *SkillService) ListSkills(userID string) ([]Skill, error)

// CreateSkill 创建自定义 SKILL
func (s *SkillService) CreateSkill(userID string, req *CreateSkillRequest) (*Skill, error)

// GetSkillForAgent 加载 Agent 使用的完整配置（合并 SKILL + Agent 覆盖）
func (s *SkillService) GetSkillForAgent(agentID string) (*AgentConfig, error)
```

---

## 5. API 设计

### 5.1 MCP Server API

| 方法 | 路径 | 说明 |
|-----|------|------|
| `GET` | `/api/v1/mcp/servers` | 列出当前用户的 MCP Server |
| `POST` | `/api/v1/mcp/servers` | 创建 MCP Server |
| `GET` | `/api/v1/mcp/servers/:id` | 获取详情 |
| `PUT` | `/api/v1/mcp/servers/:id` | 更新 |
| `DELETE` | `/api/v1/mcp/servers/:id` | 删除 |
| `GET` | `/api/v1/mcp/servers/:id/tools` | 列出该 Server 支持的工具 |
| `POST` | `/api/v1/mcp/servers/:id/test` | 测试调用（调试用） |

### 5.2 SKILL API

| 方法 | 路径 | 说明 |
|-----|------|------|
| `GET` | `/api/v1/skills` | 列出内置 + 自定义 SKILL |
| `POST` | `/api/v1/skills` | 创建自定义 SKILL |
| `GET` | `/api/v1/skills/:id` | 获取详情 |
| `PUT` | `/api/v1/skills/:id` | 更新 |
| `DELETE` | `/api/v1/skills/:id` | 删除（自定义的） |

### 5.3 SKILL 外部导入（Markdown / ZIP）

> 阶段十五 15.3 扩展：支持导入 Claude Skills 开放标准格式。

**接口：**

| 方法 | 路径 | 说明 |
|-----|------|------|
| `POST` | `/api/v1/import/skills/markdown` | Markdown 文本导入 |
| `POST` | `/api/v1/import/skills/file` | 文件上传（.md / .zip） |

**支持格式：** Claude SKILL.md（[官方标准](https://code.claude.com/docs/en/skills.md)，被 Claude Code、Cursor、Codex CLI 等多 agent 支持）

```markdown
---
name: my-skill               # 必需
description: 技能描述...      # 必需
when_to_use: 适用场景...      # 可选
allowed-tools:                # 可选
  - Bash(gh pr list *)
  - Read
  - Write
arguments: branch message     # 可选
context: fork                 # 可选（⚠️ 子代理未完整实现）
model: claude-opus-4-5     # 可选
effort: high                  # 可选
version: "1.0.0"            # 可选
---
## Instructions

你的 AI 技能指令内容...
```

**ZIP 包结构（支持批量，每个子目录一个 Skill）：

```
skill-pack.zip
├── SKILL.md          # 根目录 → 单个 Skill
├── skill-a/          # 子目录 → 单独一个 Skill
│   ├── SKILL.md
│   └── references/
│       └── doc.md
└── skill-b/
    ├── SKILL.md
    └── scripts/
        └── run.sh
```

**批量导入响应（ZIP）：**
```json
{
  "total": 3,
  "success": 2,
  "failed": 1,
  "results": [
    { "skill": {...}, "warnings": [] },
    { "skill": {...}, "warnings": ["description 超过 1536 字符"] },
    { "errors": ["解析失败: ..."] }
  ]
}
```

**安全校验规则：**

| 校验项 | 规则 |
|---|---|
| 协议白名单 | 仅 `https://`；信任域名可 `http://` |
| 内网 IP 阻断 | `127.0.0.1`/`localhost`/`10.x`/`192.168.x` 等 |
| DNS 反查 | 域名解析到内网 IP → 拒绝 |
| 危险协议 | `javascript:`/`data:`/`file:`/`vbscript:` → 拒绝 |
| 文件大小 | 默认 ≤ 10 MB |
| ZIP 安全 | 路径穿越检测；仅允许 `references/`/`scripts/`/`assets/` 子目录 |
| 引用数量 | 正文 URL > 50 个 → 警告 |
| `context: fork` | ⚠️ 提示功能未完整实现（警告） |
| 描述长度 | `description` > 1536 字符 → 警告 |

### 5.4 Agent 改动

Agent 创建/更新请求新增字段：

```go
type CreateAgentRequest struct {
    Name         string   `json:"name" binding:"required"`
    Description  string   `json:"description"`
    Model        string   `json:"model" binding:"required"`
    SystemPrompt string   `json:"system_prompt"`
    // 新增：
    SkillID      string   `json:"skill_id"`       // 从 SKILL 创建（可选）
    McpServers   []string `json:"mcp_servers"`     // 直接配置 MCP Server ID 列表
}
```

---

## 6. 前端改动

### 6.1 Agent 表单（pages/agents.vue）

- 「从 SKILL 创建」按钮 → 弹出 SKILL 选择器
- 「工具配置」区块 → 展开 MCP Server 列表 + 内置工具开关
- 工具列表显示：名称 / 类型 / 状态

### 6.2 Playground 集成（可选，P2）

Playground 后期也可以配 MCP Server，用于高级用户调试工具调用。

### 6.3 MCP Server 管理页（pages/mcp.vue）

- 列表 / 创建 / 编辑 / 删除
- 类型选择：HTTP / stdio / 内置
- 工具列表预览（创建后显示该 Server 支持哪些工具）

---

## 7. 分步落地

### 第 1 步 — MCP 数据模型 + Registry（约 1 天）

- `internal/model/mcp.go`
- `internal/mcp/registry.go`
- `McpServer` CRUD API
- 内置工具注册（fetch / web_search / code_execute）

### 第 2 步 — MCP 客户端 + ChatService 集成（约 2 天）

- `internal/mcp/http_client.go`
- `internal/mcp/executor.go`（工具调用循环）
- `ChatService.AgentChat` 集成工具调用
- 限制 MaxToolCalls = 5

### 第 3 步 — SKILL 模型 + Service（约 1 天）

- `internal/model/skill.go`
- `internal/skill/service.go`
- SKILL CRUD API
- Agent 绑定 SKILL 逻辑

### 第 4 步 — 前端（约 2 天）

- `pages/mcp.vue`（MCP Server 管理）
- `pages/agents.vue` 集成 SKILL 选择器
- Agent 表单露出工具配置区块

### 第 5 步 — 内置工具实现（约 2 天）

- `fetch` 实现（Go httpclient）
- `web_search` 实现（Tavily API，支持 keyless 模式免费调用）
- `code_execute` 实现（沙箱容器或 Python exec）

---

## 8. 明确不做

- 不做 MCP Server 的实现（只做客户端，Server 由用户提供）
- 不做本地 stdio 类型的 MCP Server 管理（安全风险大，后期再考虑）
- 不做 Agent 工具的手动调用（纯模型自主决定）
- 不做 Playground 的 MCP 配置（一期只做 Agent）

---

## 9. 风险

| 风险 | 处理 |
|--|--|
| 模型乱调工具浪费配额 | MaxToolCalls = 5；工具调用不计用户额度（类似 RAG embedding） |
| 用户配置不安全的 MCP Server | HTTP + 可选 Auth；不执行本地命令 |
| 内置工具安全（代码执行） | 沙箱容器；超时 + 内存限制；禁止网络请求 |
| MCP Server 挂了阻塞对话 | 超时 10s 降级返回错误，不卡死对话 |

---

## 10. 与现有模块的关系

| 模块 | 关系 |
|--|--|
| Agent | 新增 `mcp_servers` 字段；工具调用在 `AgentChat` 里 |
| ChatService | 工具调用循环在 ChatService 层，不影响 Playground 聊天 |
| 阶段十四记忆 | 独立模块，MCP/SKILL 叠加在记忆之上 |
| 知识库 RAG | 独立模块，RAG 注入和工具调用可同时生效 |
| 配额 | 工具调用默认不计个人额度；代码执行/网页抓取按 Token 或系统统计 |
