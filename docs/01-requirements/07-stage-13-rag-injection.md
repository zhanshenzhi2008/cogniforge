# 阶段十三：Playground 通用文件 + RAG

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-09-05 | v1.1 | 详细设计：发送框文件、RAG 注入、会话关联 | orjrs |

---

## 1. 开工门槛验收

| # | 门槛 | 状态 | 验收方式 |
|---|------|------|---------|
| 1 | 知识库上传→解析→向量稳定 | ✅ | 上传 PDF，重复 reprocess 成功 |
| 2 | Embedding 与聊天分开 | ✅ | DeepSeek Chat 单独开，向量不挂 |
| 3 | 向量语义检索 | ✅ | `POST /kb/:id/search` 返回片段+分数 |
| 4 | 对话时召回知识库 | ❌ | **本期完成** |
| 5 | Embedding 不算额度 | ✅ | 走独立 `/embeddings`，不计 quota |

---

## 2. 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         Playground 前端                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐│
│  │  发送框文件   │  │  会话文件列表 │  │  气泡文件名展示       ││
│  │  粘贴/拖入    │  │  关联知识库   │  │  附件图标            ││
│  └──────────────┘  └──────────────┘  └──────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Go ChatService                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐│
│  │ 附件解析      │  │ RAG 检索     │  │ 上下文组装           ││
│  │ attachments  │  │ Top-K 片段   │  │ system + history +   ││
│  │ (file_refs)  │  │ → fragments  │  │ RAG fragments       ││
│  └──────────────┘  └──────────────┘  └──────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                   ▼
┌──────────────────┐  ┌────────────────┐  ┌────────────────┐
│   知识库检索       │  │  Python 文档   │  │  上游 LLM      │
│ POST /kb/:id/     │  │  处理服务       │  │  (流式响应)    │
│     search        │  │  /api/upload   │  │                │
└──────────────────┘  └────────────────┘  └────────────────┘
```

---

## 3. 数据模型变更

### 3.1 ChatMessage 新增 attachments 字段

```go
// internal/model/message.go
type ChatMessage struct {
    // ... 现有字段 ...
    Images    []string   `json:"images,omitempty"`     // 已有：base64 或 URL
    Attachments []MessageAttachment `json:"attachments,omitempty"` // 新增：文件引用
}

type MessageAttachment struct {
    ID           string `json:"id"`           // 文档 ID
    Name         string `json:"name"`         // 文件名
    FileType     string `json:"file_type"`    // pdf/docx/txt/md
    Status       string `json:"status"`       // pending/processing/completed/failed
    ChunkCount   int    `json:"chunk_count"`  // 成功后填充
}
```

### 3.2 ChatConversation 新增 rag_kb_ids 字段

```go
// internal/model/conversation.go
type ChatConversation struct {
    // ... 现有字段 ...
    RAGKnowledgeBaseIDs []string `json:"rag_knowledge_base_ids,omitempty"` // RAG 关联的知识库 ID 列表
    Pinned             bool     `json:"pinned"`
}
```

---

## 4. API 设计

### 4.1 文件上传（复用知识库）

**现有**：`POST /api/v1/knowledge-bases/:id/documents/upload`
- 上传到知识库，返回 `document_id`

**新流程**：
1. 前端上传文件到用户默认知识库（或选定知识库）
2. 返回 `document_id` + `status`
3. 前端把 `document_id` 放进 `attachments[]`
4. 发送消息时带上 `attachments[]`

### 4.2 发送消息（新增 attachments 字段）

**请求**：`POST /api/v1/chat/stream`

```json
{
  "conversation_id": "conv_xxx",
  "messages": [...],
  "attachments": [
    {
      "id": "doc_xxx",
      "name": "产品说明书.pdf",
      "file_type": "pdf"
    }
  ],
  "rag_kb_ids": ["kb_xxx"],  // 可选：指定 RAG 知识库
  "model": "deepseek-chat"
}
```

**响应**：`text/event-stream`（同现有格式）

### 4.3 RAG 检索（复用现有）

**已有**：`POST /api/v1/knowledge-bases/:id/search`

```json
{
  "query": "产品退换政策是什么",
  "top_k": 3,
  "min_score": 0.5
}
```

---

## 5. 后端实现

### 5.1 ChatService 变更

```
internal/chat/
├── dto.go          # + Attachments, RAGKnowledgeBaseIDs
├── service.go      # + RAG检索 + 上下文组装
└── handler.go      # + attachments 解析
```

**ChatService.Chat() 流程**：

```go
func (s *ChatService) Chat(req *ChatRequest) {
    // 1. 解析 attachments
    attachments := parseAttachments(req.Attachments)

    // 2. 如果有附件或 rag_kb_ids，进行 RAG 检索
    var ragFragments []string
    if len(attachments) > 0 || len(req.RAGKnowledgeBaseIDs) > 0 {
        ragFragments = s.retrieveRAGFragments(
            ctx,
            userID,
            req.RAGKnowledgeBaseIDs, // 优先用请求中的，否则用会话绑定的
            req.Messages[-1].Content, // 用最新用户消息检索
        )
    }

    // 3. 组装增强上下文
    enhancedSystem := s.buildRAGSystemPrompt(ragFragments)

    // 4. 调用 LLM（system + history + enhanced_system）
    stream := s.callLLM(ctx, req, enhancedSystem)
}
```

### 5.2 RAG 检索逻辑

```go
func (s *ChatService) retrieveRAGFragments(ctx, userID string, kbIDs []string, query string) []string {
    if len(kbIDs) == 0 {
        // 取用户默认知识库
        kbIDs = s.getUserDefaultKnowledgeBases(userID)
    }

    var fragments []string
    for _, kbID := range kbIDs {
        resp, err := s.knowledgeClient.Search(&SearchRequest{
            Query:   query,
            TopK:    3, // 可配置
            MinScore: 0.5,
        })
        if err != nil {
            slog.Warn("RAG 检索失败", "kb_id", kbID, "error", err)
            continue
        }
        for _, r := range resp.Results {
            fragments = append(fragments, fmt.Sprintf("[%s] %s", r.DocumentName, r.Content))
        }
    }
    return fragments
}
```

### 5.3 RAG System Prompt 组装

```go
func (s *ChatService) buildRAGSystemPrompt(fragments []string) string {
    if len(fragments) == 0 {
        return ""
    }
    var sb strings.Builder
    sb.WriteString("\n\n## 知识库参考\n")
    sb.WriteString("以下是相关文档片段，请基于这些信息回答：\n\n")
    for i, f := range fragments {
        sb.WriteString(fmt.Sprintf("【片段 %d】\n%s\n\n", i+1, f))
    }
    sb.WriteString("如果上述信息无法回答问题，请如实说明，不要编造。")
    return sb.String()
}
```

---

## 6. 前端实现

### 6.1 发送框文件支持

**文件上传流程**：
```
用户粘贴/拖入/选择文件
        │
        ▼
前端：POST /api/v1/knowledge-bases/:id/documents/upload (FormData)
        │
        ▼
返回：{ id, name, file_type, status }
        │
        ▼
前端：attachments[] 追加该文件引用
        │
        ▼
用户发送消息
```

**前端组件**：

```
pages/playground.vue
├── PlaygroundChat.vue
│   ├── USendEmailInput (改造)
│   │   ├── [ ] 文本输入
│   │   ├── [+] 附件按钮 → 文件选择器
│   │   ├── [📎] 附件列表（可删除）
│   │   └── [发送] 按钮
│   └── 消息气泡
│       └── 显示附件图标 + 文件名
└── PlaygroundSidebar.vue
    └── RAG 知识库选择器（可选）
```

### 6.2 附件状态处理

| 状态 | 前端展示 |
|------|---------|
| `pending` | ⏳ 上传中... |
| `processing` | ⚙️ 解析中... |
| `completed` | ✅ 就绪 |
| `failed` | ❌ 解析失败 |

### 6.3 气泡文件名展示

```vue
<div v-if="message.attachments?.length" class="attachments">
  <UBadge v-for="att in message.attachments" :key="att.id" variant="subtle">
    <UIcon name="i-heroicons-document" />
    {{ att.name }}
    <UBadge v-if="att.status !== 'completed'" :color="statusColor(att.status)">
      {{ att.status }}
    </UBadge>
  </UBadge>
</div>
```

---

## 7. 任务拆分

| 序号 | 功能点 | 预计 | 难度 | 状态 | 说明 |
|-----|-------|------|------|------|------|
| 13.0 | RAG 门槛验收 | 1天 | ⭐⭐ | 🔴 | 本期第一步 |
| 13.1 | ChatMessage + Attachments DTO | 0.5天 | ⭐ | 🔴 | 数据模型变更 |
| 13.2 | ChatService RAG 检索集成 | 2天 | ⭐⭐⭐ | 🔴 | 核心逻辑 |
| 13.3 | 前端发送框附件上传 | 1.5天 | ⭐⭐ | 🔴 | 复用 KB 上传 |
| 13.4 | 前端消息气泡附件展示 | 0.5天 | ⭐ | 🔴 | 图标+文件名 |
| 13.5 | RAG Top-K 配置 | 0.5天 | ⭐ | 🔴 | 可配置片段数 |
| 13.6 | 端到端联调测试 | 1天 | ⭐⭐ | 🔴 | 全文验收 |

---

## 8. 明确不做

- 把 PDF 当图片发给 vision
- 把整份文档 base64 塞进 messages
- 在 RAG 检索失败时阻塞对话（降级为无 RAG）
- 附件检索结果缓存（每次重新检索，保证新鲜）

---

## 9. 配置项

| 配置 | 默认值 | 说明 |
|------|--------|------|
| `RAG_TOP_K` | 3 | 检索片段数 |
| `RAG_MIN_SCORE` | 0.5 | 最低相似度 |
| `RAG_MAX_TOKENS` | 2000 | 片段最大 Token |

---

## 10. 验收测试

### 人工验收

1. **RAG 检索验证**
   - 知识库上传一份 PDF（产品说明书）
   - 发消息问 PDF 里的内容
   - 回答能对 PDF 内容作答，不是瞎编

2. **附件状态验证**
   - 上传大文件，显示 "解析中"
   - 解析完成，显示 ✅

3. **降级验证**
   - RAG 服务挂了
   - 对话仍能正常回答（只是没有 RAG 增强）
