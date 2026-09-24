# Chat 短期记忆 / 长期记忆方案

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-09-02 | v1.1 | 明确 Python/Go 分工：Python=LLM 层（RAG、抽取、摘要），Go=应用层（组装、路由、写入） | orjrs |
| 2026-09-02 | v1.2 | Python 直调上游：JWT 临时凭证 + Redis 缓存（TTL 5min），屏障降级走 Go 代理 | orjrs |
| 2026-08-27 | v1.0 | 立项：Chat 当前无记忆；给出短期/长期记忆方案（场景 + 技术） | orjrs |

## [变更] 立项（2026-08-27）

- **变更原因**：Playground / Agent 对话现在只有「聊天记录」，没有真正的记忆。产品需求 MEM-001/002/003 写了但没落地；Agent 表上的 `memory_turns` 也没被用到。
- **包含代码**：无（本期只定方案，未改代码）
- **影响范围**：Go `internal/chat`、`internal/agent`；Python `memory/`；Web Playground / Agent 设置；PostgreSQL 新表；Redis 新键；配额（摘要/抽取会多打模型）
- **变更前 vs 变更后**：~~前端把整段对话原文塞给模型，换会话即失忆~~（2026-08-27）→ 服务端组装短期窗口 + 可选长期检索

---

## 1. 一句话结论

**现在不支持记忆。**  
现在有的是「历史存档」：同一窗口里把旧消息原样再发给模型；换一个新对话，模型什么都不记得。

要用大白话理解：

| | 现在实际在做什么 | 真正的记忆应该做什么 |
|--|--|--|
| 短期 | 把这次聊天的**全部原文**塞进下一次请求 | 只带**最近几轮** + 更早内容的**摘要**，保证不撑爆 Token |
| 长期 | 没有 | 跨对话记住「你是谁、偏好、约定」，提问时**检索相关几条**再塞进提示 |

---

## 2. 现在代码真实情况（对照）

| 位置 | 现状 | 算不算记忆 |
|--|--|--|
| Web `playground.vue` | 每次发送把当前窗口 **全部 messages** 交给 `/chat/stream` 或 `/agents/:id/chat` | 否。是客户端甩全量上下文，没有裁剪 |
| Go `internal/chat` | 无状态转发上游，不读历史、不裁窗口 | 否 |
| Go `chat_conversations` | PostgreSQL JSONB 存整段对话，用于左侧历史列表回显 | **存档**，不参与组装上下文 |
| Go `agents.memory_type` / `memory_turns` | 创建时写死 `short_term` / `10` | **字段空转**，Agent 对话不用它 |
| Go `AgentChat` | 只在最前面拼 Agent 的 `system_prompt` | 角色设定，不是记忆 |
| Python `memory/manager.py` | 进程内 `dict`，重启丢失；最多 50 条；关键词搜索 | 原型，**Go 对话路径没调用** |
| 知识库 RAG | 文档向量检索（阶段七/十三） | 是「查资料」，不是「记用户」 |

产品需求里 MEM-001 短期 / MEM-002 长期 / MEM-003 配置仍是 P0，**未实现**。

---

## 3. 两个概念（给非技术同学）

把对话想成「跟一个新来的同事说话」：

**短期记忆 = 这一通电话还没挂**

- 场景：你刚说「帮我写一段配额说明」，下一句「改成更口语」——它必须知道「改」的是刚才那段。
- 特点：只在**当前这一段对话**里有效；关掉这段、新建对话，短期记忆清空。
- 限制：电话越长，记不住全部原文，必须「看最近几句 + 前面用一张便签概括」。

**长期记忆 = 这个同事的笔记本**

- 场景：上周你说过「我叫小明，后端用 Go，回复要短」。今天开一个**新对话**，它仍应用这些。
- 特点：跨对话、跨天数；只记**稳定有用的事实**，不把每句闲聊都记下来。
- 限制：笔记本会变厚，提问时只抽出**跟当前问题相关**的几条，不能整本贴上去。

**不要和知识库搞混：**

| | 记什么 | 谁的 |
|--|--|--|
| 知识库 RAG | 公司 PDF / 说明书 | 文档 |
| 长期记忆 | 用户偏好、约定、项目上下文 | 这个用户（可选：这个 Agent） |

---

## 4. 必须覆盖的场景

### 场景 A — 同一窗口连续追问（短期，P0）

- **用户动作**：Playground 里连问 3 句：「写一封请假邮件」→「改成更正式」→「再加明天上午」。
- **期望**：第三句能生成「明天上午、语气正式」的请假邮件。
- **现在**：碰巧能成，因为前端把三轮原文都发了。对话变长后会失败或变贵。
- **方案**：服务端取本会话最近 N 轮（默认 10，即现有 `memory_turns`），再发给模型。

### 场景 B — 聊了很久，上下文爆了（短期 + 摘要，P0）

- **用户动作**：同一会话聊了 80 轮，继续问「刚才那个接口路径是什么」。
- **期望**：不报错、不把 80 轮全文都计费；仍能答出路径。
- **现在**：全文塞进去，Token 爆炸或被上游截断，早期内容丢失且无摘要。
- **方案**：
  1. 最近 K 轮原文保留（滑动窗口）
  2. 窗口外的内容压成一段「对话摘要」放在 system 附近
  3. 组装时卡 **Token 预算**（例如输入最多占模型窗口的 50%）

### 场景 C — 换新对话仍认识你（长期，P1）

- **用户动作**：对话 1 说「我叫小明，用 Go，别写长文」。关掉，点「新对话」，问「帮我设计一个 Redis 键」。
- **期望**：回答用 Go 示例、短文、可称呼小明；**不会**把对话 1 的整段闲聊贴进来。
- **现在**：完全失忆。
- **方案**：对话结束后（或每 N 轮）抽取「事实」写入用户长期记忆；新对话检索 Top-K 条注入。

### 场景 D — 客服 Agent 记住工单约定（长期 + Agent 作用域，P1）

- **用户动作**：绑定「客服助手」Agent。上周说「订单用顺丰、不要周末电联」。今天又进同一 Agent 问物流。
- **期望**：Agent 用上顺丰、不周末电联；**换另一个 Agent**（例如「代码助手」）不该带上快递偏好。
- **方案**：长期记忆带 `scope`：`user`（全局）或 `agent:{id}`。检索时：全局 ∪ 当前 Agent，不含其他 Agent。

### 场景 E — 用户主动改口 / 忘记（长期可改，P1）

- **用户动作**：「以后改叫我 Alex」或「忘掉我的快递偏好」或设置页删除某条记忆。
- **期望**：旧事实失效，新事实生效；删除后不再被检索到。
- **方案**：每条记忆有 id；同类事实覆盖；提供列表 / 删除 / 「本会话不用长期记忆」开关。

### 场景 F — 不该记住的东西（安全，P0 起就约束）

- **不写入长期记忆**：密码、API Key、身份证、完整聊天原文、一次性验证码。
- **不写入短期摘要**：密钥类原文；摘要里只保留「用户提到了一份密钥，已忽略」。
- **隔离**：记忆按 `user_id` 过滤，禁止跨用户召回。

### 场景 G — 配额（和阶段十二对齐）

- 短期裁剪 / 摘要 / 长期抽取都会**额外打模型**。
- **约定**：
  - 用户可见的对话回合：计入该用户 Playground/Agent 配额（现状）
  - 后台「滚动摘要」「事实抽取」：走系统用量，**默认不计入**用户对话额度（与知识库 embedding 不计个人额度同一原则）
  - 抽取失败：不影响本轮回复，只记日志，下轮再试

---

## 5. 总体架构

核心原则：**Go 是应用层，对话入口、上下文组装、数据写入；Python 是 LLM 层，RAG、抽取、摘要、向量检索。**

```
用户发一句
    │
    ▼
Go  /chat/stream 或 /agents/:id/chat
    │
    ├─ 1. 鉴权 + 配额闸门（已有）
    ├─ 2. 读短期：本 conversation 最近窗口 + 滚动摘要
    ├─ 3. 若开启长期：
    │      └─ Go 调 Python：POST /api/memory/search
    │           └─ Python 查 pgvector，返回 Top-K 片段
    │      └─ Go 拿到片段后注入 messages
    ├─ 4. 组装 messages（见 §6）
    ├─ 5a. 调上游 LLM（已有 ChatStream，流式经 Go）
    ├─ 5b. 给 Python 发临时 token（Redis 缓存，TTL 5min）
    ├─ 6. 流式回前端；落 chat_conversations（已有存档）
    └─ 7. 异步：Go 调 Python POST /api/memory/process
             └─ Python 拿 token 直调上游（抽取 + 摘要）
             └─ Go：写入 chat_memories / chat_conversations.summary
```

**Go 调 Python 只走两个接口：**

| 接口 | 方向 | 用途 |
|---|---|---|
| `POST /api/memory/search` | Go → Python | 长期记忆向量检索 |
| `POST /api/memory/process` | Go → Python | 抽取 + 摘要（异步） |

### 5.1 Python 直调上游：临时凭证 + Redis 缓存

Python 的 LLM 调用**直调上游**，不经过 Go 做代理。Key 仍然在 Go Python 无持久化 Key，通过**临时凭证**授信。

#### 凭证生命周期

```
用户发消息 → Go 鉴权通过
               │
               ├─ 查 Redis：cogniforge:llm_token:{user_id}
               │     ├─ 有且未过期 → 返回 token 给 Python
               │     └─ 无或已过期 → 生成新 token，写入 Redis（TTL 5 分钟）
               │
               └─ 返回临时 token
                           │
Python 拿到 token
    │
    ├─ 调上游 LLM：Authorization: Bearer <token>
    └─ 调 embedding：同上
```

#### 凭证设计

**存储**：Redis db0，键前缀 `cogniforge:llm_token:`，TTL 5 分钟。

```
cogniforge:llm_token:{user_id}
```

**值**：JWT，包含：
- `sub`: user_id
- `exp`: 5 分钟后过期
- `scope`: `llm_call`
- `sign`: HMAC-SHA256(Go secret key)，防伪造

Go 签发，Python 验证（Go/Python 共用同一个 HMAC key，存 docker-compose 共用的 `.env`，**复用已有 `JWT_SECRET`**）。

**屏障机制**（gap prevention）：
- 取不到 token 或 token 失效时，Python 同步阻塞请求 → Go 获取新 token → 重试一次
- 重试仍失败 → 降级：走 Go `/api/v1/chat/completions` 代理（兜底，不阻塞对话）
- 对话流式不走这个路径（实时性要求高，继续经 Go）

#### 各端职责

| | Go | Python |
|---|---|---|
| Key 持有 | ✅ 唯一 | ❌ 无 |
| 凭证签发 | ✅ 唯一 | ❌ |
| 凭证验证 | ✅ | ✅（JWT HMAC） |
| LLM 直调 | ❌ | ✅ 记忆/摘要/Agent |
| LLM 代理（兜底） | ✅ 流式对话 + 屏障降级 | ❌ |

#### 为什么不是「屏障」而是「双重保障」

屏障（semaphore/lock）是防并发。这里是**授信下发**：Go 主动给 Python 授权，让 Python 在 TTL 内自由调 LLM，不需要每次都问 Go。

- 如果 Redis 挂 → Python 降级走 Go 代理，不卡死
- 如果 Go 签发故障 → Python 用已有 token 撑完 TTL（最多 5 分钟无新 token）
- 如果 Python 拿到旧 token（TTL 临界）→ 上游会 401，Python 立刻去拿新的 + 重试

组装顺序（固定，避免「资料把人设冲掉」）：

```
[0] system：Agent 人设 / 默认助手（已有）
[1] 长期记忆块（可选，检索到才加）：编号列表，告诉模型「这是笔记本，不是当前对话」
[2] 对话摘要（可选）：窗口外压缩
[3] 最近 K 轮 user/assistant 原文
[4] 本轮用户消息
```

知识库 RAG（阶段十三）若同时开启，插在 **[1] 和 [2] 之间**，单独标明「来自知识库文档」，不要和长期记忆混成一块。

---

## 6. 短期记忆（MEM-001）— 技术

### 6.1 策略（由简到繁，必须按序做）

| 档 | 名字 | 做什么 | 何时用 |
|--|--|--|--|
| L1 | 滑动窗口 | 只保留最近 `memory_turns` 轮（一轮 = 1 user + 1 assistant） | **先做这个**，立刻能控 Token |
| L2 | Token 预算 | 从最新往旧加，加到 `max_input_tokens` 为止；超了丢掉更旧的 | 和 L1 一起做，防超长单条 |
| L3 | 滚动摘要 | 被裁掉的旧轮，用小模型压成 ≤ 500 字摘要，下次放在 [2] | 窗口经常被裁时再加 |

默认：Playground 无 Agent 时 `memory_turns = 10`（与 Agent 表默认值一致）。用户可在参数滑层改 4 / 10 / 20。上限 40，禁止「无限」。

### 6.2 数据放哪

**权威数据**：已有 `chat_conversations.messages`（PostgreSQL JSONB）。短期记忆**不另存一份全文**。

**热缓存（可选，L2 之后）** Redis，键必须 `cogniforge:` 前缀，只 db0：

```
cogniforge:memory:short:{conversation_id}     # JSON：最近窗口消息；TTL 24h
cogniforge:memory:summary:{conversation_id}   # 滚动摘要文本；TTL 7d
```

禁止把 API Key、密码写入这些键。对话删除时 `DEL` 对应键。

### 6.3 服务端裁剪算法（L1+L2）

输入：`messages[]`（本会话已落库 + 本轮新句）、`memory_turns`、`max_input_tokens`。

1. 去掉空内容。
2. 从尾部取最多 `memory_turns * 2` 条（user/assistant 成对优先；落单的 user 保留）。
3. 估算 Token（复用现有 `EstimateUsage` 口径即可，不求精确）。
4. 若仍超预算：继续从最旧一条删，直到进入预算；**本轮 user 消息永不删**。
5. 若开启 L3 且发生了删除：把删掉的部分送给摘要模型，覆盖 `summary`。
6. 输出：`summary?` + `window[]`。

前端改动（重要）：

- **可以继续**把当前窗口 messages 发上来（兼容现状）。
- Go **以服务端裁剪结果为准**，忽略前端多带的旧消息，避免两边各裁一次不一致。
- 长期做完后，请求可改为只带 `conversation_id` + 本轮一句，减少上传量；一期不强求。

### 6.4 Agent 字段终于要生效

| 字段 | 用法 |
|--|--|
| `memory_type` | `off` = 不带历史（每句独立）；`short_term` = 只用窗口/摘要；`full` = 短期 + 长期 |
| `memory_turns` | 窗口轮数，默认 10，范围 1–40 |

`AgentChat` 组装前读取这两个字段；Playground 未选 Agent 时用用户偏好或默认 `short_term` / 10。

工作流 Agent 节点已有 `memory_turns` 配置，真正跑 Agent 时走同一套组装，禁止再各写一套。

---

## 7. 长期记忆（MEM-002）— 技术

### 7.1 记什么（白名单，宁少勿滥）

只抽取下面几类，写成**短句事实**（建议 ≤ 80 字/条）：

| kind | 例子 |
|--|--|
| `profile` | 用户叫小明；职业后端；主要语言 Go |
| `preference` | 回复用中文、要短、代码要可复制 |
| `decision` | 配额不做充值；文件附件等 RAG 稳了再做 |
| `episode` | 2026-08-20 一起排过密码重置邮件方案（一句话，不贴全文） |

不抽：情绪发泄、整段代码、他人隐私、密钥、一次性问题（「这行报错什么意思」）。

### 7.2 何时写入（Python LLM 层）

对话流式**成功结束后**，Go 异步调 `POST /api/memory/process`：

1. Go 把本轮 user + assistant 文本 POST 给 Python
2. Python 调用 LLM（通过 Go Gateway）抽取事实
3. 返回 JSON 列表 `{kind, content, scope, confidence}`
4. `confidence < 0.7` 丢弃
5. 与已有记忆去重：同 `user_id` + 近义（向量相似度 > 0.9）→ **更新原文**，不堆重复
6. Python 返回处理结果；Go 写入 `chat_memories`

**Python LLM 路径**：Python → Go `/api/v1/chat/completions` → 上游（Key 持有者在 Go）。

抽取提示必须带：「不要记录密钥和证件号；不要复述对话全文；每条一个完整短句」。

### 7.3 何时读出（Python LLM 层）

每次用户发消息且 `memory_type = full`：

1. Go 调 `POST /api/memory/search`，带当前问句文本
2. Python 做 embedding（调 Go `/api/v1/embeddings`）
3. Python 查 pgvector，过滤 `user_id = 当前用户`，`scope IN ('user', 'agent:'+当前agent或空)`，`deleted_at IS NULL`
4. `top_k = 5`，`min_score = 0.4`（可配置）
5. 命中 0 条：Python 返回空；Go 不加长期块
6. 命中后 Python 返回 Top-K 片段；Go 注入 [1]；Python 异步更新 `last_accessed_at`

### 7.4 表设计（落地时写入 `docs/04-database`，现网表**不加** `cf_` 前缀）

```sql
CREATE TABLE chat_memories (
    id            VARCHAR(64) PRIMARY KEY,
    user_id       VARCHAR(64) NOT NULL,
    agent_id      VARCHAR(64),          -- 空 = 用户全局；有值 = 仅该 Agent
    conversation_id VARCHAR(64),        -- 来源会话，可空（手工创建）
    kind          VARCHAR(32) NOT NULL, -- profile / preference / decision / episode
    content       TEXT NOT NULL,
    vector        vector(1536),
    importance    SMALLINT NOT NULL DEFAULT 5,  -- 1-10，检索时可加权
    metadata      JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_chat_memories_user ON chat_memories(user_id);
CREATE INDEX idx_chat_memories_agent ON chat_memories(agent_id);
CREATE INDEX idx_chat_memories_vector ON chat_memories
    USING hnsw (vector vector_cosine_ops) WITH (m = 16, ef_construction = 64);
```

与知识库 `knowledge_chunks` **分表**，不要把用户隐私写进知识库。

滚动摘要**不入**此表，只放 Redis / `chat_conversations` 扩展字段 `summary TEXT`（落地时二选一：优先列上 `chat_conversations.summary`，重启不丢）。

### 7.5 Python memory/manager.py 处置

~~进程内 dict 当记忆~~（2026-08-27）：对话主路径废弃它。

- 改为：长期记忆检索（`/api/memory/search`）和抽取（`/api/memory/process`）的新路由
- 复用 Python RAG 管道（pgvector_store、GoEmbedder）
- **不再**在 Python 进程内存存任何生产数据
- `/api/memory/*` 路由：管理接口（列表/删除/创建长期记忆条目），走 Go 写的 `chat_memories` 表

---

## 8. 记忆配置（MEM-003）

### 8.1 Agent / Playground 可配项

| 配置 | 默认 | 说明 |
|--|--|--|
| `memory_type` | `short_term` | `off` / `short_term` / `full` |
| `memory_turns` | 10 | 1–40 |
| `memory_top_k` | 5 | 长期检索条数 |
| `memory_min_score` | 0.4 | 低于则不用 |

Playground 未选 Agent：用账号偏好；没有偏好则默认短期 10 轮、**先不默认打开长期**（长期要等抽取稳定）。

### 8.2 用户可见能力（设置页或 Playground 侧栏后期）

- 查看我的长期记忆列表（按 kind 分组）
- 删除一条 / 清空全部
- 本会话开关：「这次对话不要用长期记忆」（请求带 `use_long_term: false`）

一期可只做 API，UI 用最简单列表。

### 8.3 API 草案（落地时写入 `docs/03-api`）

```
GET    /api/v1/memories              当前用户长期记忆列表
DELETE /api/v1/memories/:id          删除一条（软删）
DELETE /api/v1/memories              清空（需确认参数）
POST   /api/v1/memories              手工补一条（可选）
```

对话请求**不新增必填字段**。`conversation_id` 已能从历史接口拿到；组装在服务端做。可选：`use_long_term: boolean`。

---

## 9. 注入给模型的格式（示例）

```
system: 你是客服助手……（Agent 原 system_prompt）

system: 【长期记忆·仅供参考，与当前问题无关则忽略】
1. [preference] 用户希望周末不要电联
2. [decision] 该用户订单默认顺丰

system: 【本会话更早内容的摘要】
用户在核对物流时效，已确认走顺丰，地址在市区。

user: 周末能派送吗？
```

明确写成独立 system/片段，**不要**伪装成用户说过的话。

---

## 10. 分步落地（简单可控，禁止一次做完）

### 第 1 步 — 短期窗口生效（约 3 天）⭐ 先做

- Go `Chat` / `AgentChat`：按 `memory_turns` + Token 预算裁剪
- Agent 的 `memory_turns` 真正参与
- Playground 参数滑层露出「记住最近 N 轮」
- 单测：40 轮历史只带 10 轮；本轮 user 不被裁掉

**验收（非技术）**：故意把 N 调成 2，第三句之前的细节模型会忘；调回 10 又能接上。证明裁剪真的发生在服务端。

### 第 2 步 — 滚动摘要（约 3 天）

- `chat_conversations.summary`
- 窗口外内容压缩；注入 [2]
- 摘要失败则退回「只窗口、无摘要」

**验收**：同一会话聊很久，问很早以前的一个结论，仍能大致说对，但请求 Token 明显低于全文。

### 第 3 步 — 长期记忆（约 1 周，依赖 embedding 可用）

- 表 `chat_memories` + AutoMigrate
- 异步抽取 + 去重 + 检索注入
- `memory_type=full` 才走
- 列表/删除 API

**开工门槛**（与阶段十三类似，缺一不可）：

1. 向量默认供应商能用（对话/向量已分开）
2. `/embeddings` 稳定
3. 短期窗口已上线

**验收**：对话 1 说偏好；新开对话 2（`full`）问相关问题能用上偏好；删除该条后不再出现。

### 第 4 步 — 配置 UI + 安全打磨

- Agent 表单露出记忆类型 / 轮数（现在表单没露）
- 记忆列表页
- 抽取过滤密钥的回归测试

---

## 11. 明确不做

- 不做「把全部历史对话向量化当搜索引擎」当一期长期记忆（太贵、易泄隐私）
- 不把长期记忆写入知识库表
- 不把记忆当无限上下文：再强的记忆也要 Token 预算
- 不在前端做唯一裁剪逻辑
- 不把 Python 进程内存当生产记忆
- 一期不做记忆自动过期删除（只留 `last_accessed_at`）；跨用户共享记忆不做

---

## 12. 风险

| 风险 | 处理 |
|--|--|
| 抽取把闲聊写成「事实」 | 白名单 kind + 置信度 + 用户可删 |
| 长期记忆污染当前任务 | 提示写明「无关则忽略」；默认 Playground 先只开短期 |
| 摘要幻觉 | 摘要只用于辅助；关键以窗口原文为准 |
| 多打模型烧钱 | 摘要/抽取不计用户对话额度，但要打系统监控；可每 5 轮抽一次而不是每轮 |
| RAG 与记忆抢上下文 | 分块标注来源；共同占用 Token 预算，超了优先保本轮问句 + 窗口 |

---

## 13. 和现有文档的关系

| 文档 | 关系 |
|--|--|
| `01-product-requirements.md` §2.4.3 MEM-001/002/003 | 本文件是详细方案 |
| `04-playground-files-rag.md` | 知识库附件 ≠ 记忆；阶段十三仍等 RAG |
| `02-quota-design.md` | 用户可见对话计配额；抽取/摘要默认不计个人额度 |
| `04-database` / `03-api` | **代码落地当周**再改这两份，本文件 §7.4 / §8.3 为草案 |
