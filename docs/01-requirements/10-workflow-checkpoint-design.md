# 工作流增强：真实 AI 节点 + Checkpoint（阶段十八）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-10-26 | v1.0 | 立项：加强现有 Vue Flow + Go 引擎；W0–W3 排期；不做第二套状态机 | orjrs |

## [变更] 立项（2026-10-26）

- **变更原因**：Agent 平台需要可恢复编排；现网图画得好但 LLM/Agent 节点仍为模拟，执行一口气跑完无法 waiting/resume
- **路线**：加强现有工作流（Vue Flow 画布 + Go `WorkflowEngine`），**不**引入 n8n / Temporal / Java 通用状态机引擎
- **包含文档**：本文件；`docs/04-database/01-database-design.md`；`docs/03-api/01-api-design.md`；开发计划阶段十八
- **不做**：LLM 决定跳过审批；Agent 全局 FSM 与任务 status 抢戏；嵌第三方编排产品

---

## 1. 一句话

**任务真相 = 工作流执行**；Checkpoint 挂在执行上；Agent/LLM 只是节点里的大脑（复用现有 `Chat` / `ChatWithTools`）。

---

## 2. 现状与缺口

| 已有 | 缺口 |
|------|------|
| Vue Flow 画布；节点类型 start/end/llm/agent/condition/loop/http/code/delay | LLM / Agent 节点返回模拟文案 |
| Go `WorkflowEngine` 拓扑执行；`workflow_executions` 三态 | 无 `waiting`；几乎不写 `current_node` |
| HTTP / condition / loop 等部分真执行 | 无 `node_runs` 表；无 checkpoint；无 resume API |
| 对话 Agent 已有 `ChatWithTools` | 工作流 Agent 节点未接线 |

---

## 3. 分层状态（独立字段，不是三套 FSM）

| 层次 | 存哪 | 取值 | 说明 |
|------|------|------|------|
| **任务** | `workflow_executions.status` | `pending` / `running` / `waiting` / `succeeded` / `failed` / `cancelled` | 唯一对外「跑到哪一档」 |
| **游标** | `current_node` | 节点 id | 当前/下一个关注点 |
| **节点尝试** | `workflow_node_runs` | `pending` / `running` / `succeeded` / `failed` / `skipped` | 每个节点每次尝试一行 |
| **检查点** | `workflow_checkpoints` | 快照行 | 恢复用，**不是**第四种 status |
| **Agent 回合**（可选） | `node_runs.detail.phase` | `thinking` / `tool_call` / `done` | 仅 agent 节点细节，不升全局 |

原则：

- 状态机规则写在 **引擎代码 + 图定义**；Agent 只提案，不能改权限/审批。
- ~~把工具失败写成任务 `failed` 且无法区分可重试~~ → 节点 `failed` + 任务可 `waiting`（重试策略）或任务终态 `failed`。

任务状态转移（目标）：

```text
pending → running ⇄ waiting
              ↓
     succeeded | failed | cancelled
```

- `waiting`：等人审、外部回调、或策略门（预算/危险工具）
- 可重试失败：节点失败后按策略再入 `running`（新 attempt），不是另造 `RETRY_WAITING` 全局态（可用 `waiting` + `wait_reason=retry`）

---

## 4. Checkpoint

### 4.1 存什么

| 字段 | 用途 |
|------|------|
| `execution_id` + `seq` | 有序快照 |
| `cursor_node_id` | resume 从谁开始（含「下一节点」语义，落地时写清） |
| `variables` | 执行变量 JSON |
| `completed_node_ids` | 已成功节点，避免重复副作用 |
| `definition_hash` / `engine_version` | 定义变了是否允许恢复 |
| `reason` | `node_ok` / `enter_waiting` / `manual` / `final` |

大消息（Agent 多轮）**不整段落表**：`detail` 摘要 + 可选外部引用；表保持小。

### 4.2 何时写

1. 每个**确定性节点**成功之后  
2. Agent/LLM 节点成功之后（整节点结束）  
3. 进入 `waiting` **之前**（必写）  
4. 终态 `succeeded` / `failed` / `cancelled`（最终点）

### 4.3 恢复

`POST /api/v1/executions/:id/resume`：

1. 校验 status ∈ {`waiting`}（或显式允许从中断的 `failed` 按策略恢复——一期可只做 waiting）  
2. 加载最新 checkpoint（或指定 `seq`）  
3. 重建 `ExecutionContext`（variables + 跳过已完成节点）  
4. status → `running`，继续拓扑执行  

幂等：已成功节点不重跑；工具侧后续用 idempotency key（W3）。

---

## 5. 分期（W0–W3）

| 阶段 | 内容 | 验收（非技术） | 预估 |
|------|------|----------------|------|
| **W0** | LLM / Agent 节点接真：走 Go Chat / `ChatWithTools`；去掉模拟 | 工作流里问一句能出真模型回答；勾了工具能查天气类 | 1.5–2 天 |
| **W1** | `workflow_node_runs` + 写 `current_node`；执行详情 API；画布节点高亮 | 跑的时候能看到亮到哪、哪失败 | 2 天 |
| **W2** | `workflow_checkpoints`；status 增 `waiting`/`cancelled`/`succeeded`；Resume/Cancel API | 能暂停在等待态，点继续接着跑 | 2–3 天 |
| **W3** | 审批节点/危险工具门；超时扫描；工具幂等 | 人审通过再继续；进程重启可捞 waiting | 3 天 |

**开工顺序**：必须先 W0，再 W1，再 W2。W3 可与产品 HITL 需求对齐后开。

---

## 6. W0 接线约定

### 6.1 LLM 节点

- 配置：`model`、`system_prompt`、`prompt`（支持 `{{var}}`）、temperature/max_tokens  
- 实现：组装 messages → `ChatService.Chat`（非流式即可；调试面板以后再 SSE）  
- 输出写入节点结果：`content`、`usage`、`model`  
- 配额：`source=workflow`（与现有配额设计对齐）

### 6.2 Agent 节点

- 配置：`agent_id`、可选覆盖 `prompt` / `variables`、`memory_turns`  
- 实现：加载 Agent（prompt_id / skill / system_prompt 优先级与对话一致）→ `InjectMCPTools` → 有工具则 `ChatWithTools`，否则 `Chat`  
- **禁止**再写一套人设解析；复用 `prompt.Service` / agent handler 逻辑（可抽共享函数）

### 6.3 明确不做（W0）

- 节点内流式推前端  
- Checkpoint / waiting  
- 子工作流、多 Agent 交接  

---

## 7. API（增量）

已有：`POST .../execute`、`GET .../executions`、`GET /executions/:id`

新增（W2 起强制；W1 可先 enrich GET）：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/executions/:id` | 返回 status、current_node、node_runs[]、latest_checkpoint_seq |
| GET | `/api/v1/executions/:id/nodes` | 节点尝试列表 |
| POST | `/api/v1/executions/:id/resume` | 从 checkpoint 继续 |
| POST | `/api/v1/executions/:id/cancel` | 取消 → `cancelled` |
| GET | `/api/v1/executions/:id/checkpoints` | 列表（排障） |

状态字：对外逐步用 `succeeded`（与现网 `completed` **兼容期双读**，写入新值前文档约定迁移：一期可读 completed 当 succeeded）。

---

## 8. 数据模型（摘要）

详见 `docs/04-database/01-database-design.md` §工作流增强。

- 扩展 `workflow_executions`：`wait_reason`、规范化 status  
- 新表 `workflow_node_runs`  
- 新表 `workflow_checkpoints`  

Redis：执行锁可选 `cogniforge:workflow:exec:lock:{execution_id}`（防双 resume）；先写库设计再写代码。

---

## 9. 前端（cogniforge-web）

| 阶段 | UI |
|------|-----|
| W0 | 执行结果展示真实 content（调试面板） |
| W1 | 画布节点按 node_runs 着色：running/success/failed |
| W2 | 「继续」「取消」按钮；waiting 提示文案 |
| W3 | 审批通过入口 |

画布继续 **Vue Flow**，不换 n8n。

---

## 10. 验收清单（产品）

1. 画一条 start → llm → end，执行后看到真实模型回答（非 Simulated）  
2. start → agent（带 web_search）→ end，工具题能答到实时信息  
3. 执行中/结束后能指出哪个节点失败  
4.（W2）人为进入 waiting 后 resume，不从第一节点重跑  
5. LLM **不能**单独把任务标成 succeeded 而跳过引擎校验  

---

## 11. 与其它方案边界

| 方案 | 本阶段态度 |
|------|------------|
| 自研加强（本设计） | ✅ 采用 |
| n8n 嵌入 | ❌ 不做（集成场景另议） |
| Temporal | ❌ 本期不引入；规模上来再评估 |
| LangGraph 当平台真相 | ❌；Agent 研究可另开沙盒 |
| GPT Java StateMachineEngine | ❌ 不平行造轮子 |
