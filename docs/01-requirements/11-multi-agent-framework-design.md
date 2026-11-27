# 多 Agent 框架设计（阶段十九草案）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-27 | v0.14 | LLM 亦写入 last_agent_content 串联下游 | orjrs |
| 2026-11-27 | v0.13 | Web：max_depth/max_steps 限额可配 | orjrs |
| 2026-11-27 | v0.12 | Web：worker_retry_once 创建/列表可配 | orjrs |
| 2026-11-27 | v0.11 | team 节点 OnProgress：delegate/handoff 子步骤 | orjrs |
| 2026-11-27 | v0.10 | M4 产品：expand-pipeline layout=parallel 扇出图 | orjrs |
| 2026-11-27 | v0.9 | M4 引擎：同层无依赖节点并行（ready-wave） | orjrs |
| 2026-11-24 | v0.8 | Pipeline 一键成图 expand-pipeline；Agent 串联吃上一步输出 | orjrs |
| 2026-11-23 | v0.7 | POST /teams/:id/run；画布 team 节点绑定小队 | orjrs |
| 2026-11-23 | v0.6 | §12 落地：软失败+可选重试、team_id 落 chat、Redis run_state、Web 续聊 | orjrs |
| 2026-11-23 | v0.5 | 拍板 §12 四问：隐形展开、预览落 chat、编排原语、工人失败回主管 | orjrs |
| 2026-11-23 | v0.4 | chat 回传 run_state；handoff 后按 active 换人设 | orjrs |
| 2026-11-23 | v0.3 | M2 起步：POST /teams/:id/chat + delegate/handoff 工具 | orjrs |
| 2026-11-23 | v0.2 | M1 起步：agent_teams 表 + Team CRUD + RunState/handoff 原语 | orjrs |
| 2026-11-23 | v0.1 | 立项：模式分层、与现有工作流关系、参考来源、分期不做清单 | orjrs |

## [变更] max_depth / max_steps Web 配置（2026-11-27）

- **变更原因**：门禁默认写进 config，产品侧原先改不了
- **已落地**：新建填深度/步数；「限额」弹窗改现有小队（含失败重试）
- **代码**：`cogniforge-web/pages/teams`；前端设计 v1.123

## [变更] worker_retry_once Web 配置（2026-11-27）

- **变更原因**：§12.4 可选同人重试仅后端可读 config，产品无开关
- **已落地**：跨仓小队新建勾选 + 列表/卡片切换；写 `config.worker_retry_once`
- **代码**：`cogniforge-web/pages/teams`；前端设计 v1.122

## [变更] team 节点子步骤进度（2026-11-27）

- **变更原因**：画布 team 跑完才出结果，中间委派/交接不可见
- **已落地**：`ChatDeps.OnProgress`；delegate/handoff 上报 → `node_runs.detail.progress`（复用 LLM 进度通道）
- **未做**：自动换人（默认关）；独立 SSE 事件流仍可后续加
- **代码**：`orchestration/team/chat.go`、`engine/nodes/team.go`；API v1.100

## [变更] M4 扇出成图（2026-11-27）

- **变更原因**：引擎已能并行；产品侧要一键生成 start→工人们→end
- **已落地**：`BuildParallelFanoutDefinition`；`expand-pipeline` body `layout=parallel`（最多 8 工人，剔除主管）
- **Web**：小队卡「生成并行图」
- **代码**：`orchestration/team/pipeline.go`、`ExpandOptions`；跨仓 `pages/teams`

## [变更] M4 引擎同层并行（2026-11-27）

- **变更原因**：扇出/汇合需要引擎能同时跑无依赖节点
- **已落地**：`WorkflowEngine.Execute` 按就绪波次调度；白名单类型同波并行；Output/`ForNode` 线程安全
- **未做**：~~Team 配置扇出工人产品化~~（2026-11-27 已做）；无上限并行仍禁止
- **代码**：`engine/core/workflow_engine.go`（EngineVersion 18.4）

## [变更] Pipeline 一键成图（2026-11-24）

- **变更原因**：M1「从 Team 生成线性图」；用户能说清流水线 vs 单 Agent
- **已落地**：
  - `POST /api/v1/teams/:id/expand-pipeline`：pipeline → start→agent…→end；其它模式 → start→team→end；回写 `workflow_id`
  - Agent 节点空 prompt 时接 `last_agent_content`（上一步输出）
  - ~~LLM 未写该变量~~（2026-11-27：LLM 成功亦写入；空 prompt 优先接上游）
  - Web 小队卡「生成/打开流水线」
- **未做**：~~并行工人~~（2026-11-27 引擎侧已做）→ Team 扇出 UI、自动换人
- **代码**：`orchestration/team/pipeline.go`、`service.ExpandToWorkflow`、`engine/nodes/agent.go`；Web `pages/teams`

## [变更] 任务型 /run + 画布 team 节点（2026-11-23）

- **变更原因**：预览之外需要正式任务入口与可选画布绑定（§6.1 / §8.2）
- **已落地**：
  - `POST /api/v1/teams/:id/run`：`task`/`goal` → `RunTask`；**不**落 `chat_conversations`
  - 工作流节点类型 `team`：绑 `team_id` + 任务说明；引擎内 `TeamNodeExecutor` → `RunTask`；输出含 `content`/`run_state` 进 `node_runs`
  - Web 画布节点面板可选小队、填任务
- **未做**：~~Pipeline 一键生成线性图~~（2026-11-24 已跟）→ ~~team 节点子步骤进度~~（2026-11-27 OnProgress）；自动换人仍待
- **代码**：`orchestration/team/run.go`、`engine/nodes/team.go`、`app/team/handler.go`；Web `workflows/[id].vue`

## [变更] §12 拍板项落地（2026-11-23）

- **变更原因**：§12 已定稿，补齐预览落库、Redis 续聊、工人软失败
- **已落地**：
  - 工人失败 → 工具结果字符串回传主管（不硬断）；`config.worker_retry_once` 可选同人再试 1 次
  - `chat_conversations.team_id`；`POST /teams/:id/chat` 写会话并回 `conversation_id`
  - Redis `cogniforge:team:run:{user}:{team}:{session}`（TTL 24h）；无 body `run_state` 时按 conversation 加载
  - Web 小队预览：带 `conversation_id` 续聊、拉历史、展示工人失败提示
- **未做**：~~画布 `team` 节点、`/teams/:id/run`~~（2026-11-23 已跟）→ 自动换人仍待
- **代码**：`orchestration/team/{chat,store}.go`、`app/team/handler.go`、`model/conversation.go`；Web `useTeams` / `pages/teams`

## [变更] 开放问题拍板（2026-11-23）

- **变更原因**：实现前四问需定稿，避免预览落库、工作流绑定、失败策略各写各的
- **影响范围**：小队预览存档、Team↔工作流关系、`delegate`/`handoff` 实现口径、工人失败默认行为；后续改 `chat_conversations` / Team chat / config
- **变更前 vs 变更后**：~~§12 四问待定~~ → 已拍板（见 §12）；§13 总结表补四行
- **代码**：~~本期仅文档；实现跟拍板另排期~~（2026-11-23）→ v0.6 已落地预览落 chat / 软失败 / Redis

## [变更] run_state 回传换人设（2026-11-23）

- **已落地**：请求可带上一轮 `run_state`；`ResumeRunState` 校验成员后按 `active_agent_id` 加载人设；Web 预览回传
- **未做**：~~服务端 Redis 持久化 run~~（2026-11-23 已跟）→ 画布 Team 节点仍待
- **代码**：`ResumeRunState`；handler/useTeams/teams 页

## [变更] M2 起步：小队 chat + 编排工具（2026-11-23）

- **已落地**：`POST /api/v1/teams/:id/chat`；主管注入 `delegate_to_agent` / `handoff_to_agent`；工人隔离上下文；返回 `run_state`
- **未做**：~~handoff 后自动换 system 续聊、Web 小队页~~（2026-11-23 已跟）→ 画布 Team 节点仍待
- **代码**：`orchestration/team/chat.go`、`tools.go`；`app/team` Chat

## [变更] M1 起步落地（2026-11-23）

- **变更原因**：从设计进入实现：先落小队说明书与共享状态原语
- **已落地**：
  - 表 `agent_teams`（`model.AgentTeam`，AutoMigrate）
  - `orchestration/team`：`RunState`、Handoff/Delegate 门禁、CRUD Service
  - `app/team`：`GET/POST/PUT/DELETE /api/v1/teams`
- **未做**：~~`/teams/:id/chat`、主管工具 `delegate_to_agent`~~（2026-11-23 已起步）→ 画布 Team 节点、前端小队页仍待做
- **包含代码**：`internal/model/team.go`、`internal/orchestration/team/*`、`internal/app/team/handler.go`、router 注册

## [变更] 多 Agent 框架立项（2026-11-23）

- **变更原因**：工作室已有「单 Agent 构建/预览」；工作流已有多个 Agent 节点串行；用户需要「一队人」协作，而不是再造一套编排引擎
- **范围**：~~设计先行，本期不写代码~~（2026-11-23）→ 已开始 M1 代码
- **影响**：后续改 `agents` / `workflows` / `orchestration` / Web 工作室；跨仓同步开发计划
- **不采用**：嵌 Temporal / n8n；独立第三套 FSM；Agent 互相改权限/审批；本期上 A2A 跨产品互联

---

## 1. 一句话（给非技术读者）

现在：一个智能体 = 一个人设 + 工具，自己跟用户聊。  
目标：可以组一支小队——有人管分活，有人查资料，有人写稿；小队怎么协作，规则写清楚，跑起来能暂停、能看进度、能接着跑。

**编排真相仍是「任务执行」**（沿用工作流 checkpoint），不是每个 Agent 各搞一套状态机。

---

## 2. 现状（我们已有什么）

| 能力 | 现状 | 算不算多 Agent |
|------|------|----------------|
| 单 Agent CRUD + 对话 + MCP/技能 | ✅ 工作室左配置 / 右构建预览 | 否（单体） |
| 工作流里多个 **Agent 节点** 串行 | ✅ Vue Flow + Go 引擎 + checkpoint / waiting | **弱多 Agent**：图上固定顺序，没有「谁接下棒」的运行时协商 |
| 条件 / 循环 / wait / 人审 | ✅ | 编排积木，不是角色协作 |
| Agent 调另一个 Agent 当工具 | ❌ | 缺口 |
| 主管分派 / 运行时交接（handoff） | ❌ | 缺口 |
| 并行辩论 / 投票 | ❌ | 后期可选 |

结论：**已有「图上排多个人」；缺「运行时像团队那样说话与交棒」。**  
设计原则：**加强现有 Agent + Workflow，不新开第三套编排产品。**  
（与 `10-workflow-checkpoint-design.md`「任务真相 = 工作流执行」一致。）

---

## 3. 参考来源（我们抄什么、不抄什么）

| 参考 | 学什么 | 不照搬 |
|------|--------|--------|
| **LangGraph**（LangChain） | `StateGraph`、共享状态、Supervisor 图、interrupt / resume | 不引入 Python LangGraph 依赖；状态落我们自己的 checkpoint |
| **OpenAI Swarm → Agents SDK** | **Handoff**：当前 Agent 把对话权交给另一个 Agent；轻量、好理解 | 不做无持久化的玩具 demo；我们要落库可恢复 |
| **Anthropic「Orchestrator–Workers」**（多 Agent 研究实践） | 主管拆任务 → 工人并行/串行 → 汇总；工人上下文隔离防污染 | 不做研究向无限并行；要配额与超时门 |
| **AutoGen / AG2 GroupChat** | 群聊选下一个说话者 | 默认不做自由群聊（难控、费 Token）；仅作后期实验模式 |
| **CrewAI** | 角色（Role）+ 任务（Task）+ 顺序/层级流程 | 不绑 CrewAI 运行时；角色模型可借鉴命名 |
| **Microsoft Magentic-One** | Orchestrator + 专用工人（Web/文件/代码） | 专用工人用我们已有 MCP/技能覆盖，不单独造一套 |
| **Google A2A** | Agent 间标准消息协议 | **远期**；本期队内用内部事件即可 |
| **Dify / Coze 工作流** | 画布上挂多个 Agent/LLM 节点 = 产品层「多 Agent」 | 我们已有此层；要补「Team / Handoff」语义 |
| **本仓已有文档** | `10-workflow-checkpoint-design.md`、`02-layering.md`（runtime vs orchestration） | — |

行业共识可收成三句话：

1. **共享一份任务状态**，不要每个 Agent 各维护真相。  
2. **交接要显式**（handoff / 边 / 工具调用），不要隐式抢麦。  
3. **人审与危险工具在编排层拦**，不交给模型改权限。

---

## 4. 概念模型

```text
Agent（已有）
  = 人设 + 模型 + 工具/技能 + 记忆范围

Team（新增，逻辑实体）
  = 一组 Agent 引用 + 协作模式 + 共享变量约定 + 入口 Agent
  ≈ 「一支小队的说明书」

Run / Execution（已有工作流执行，可挂 Team）
  = 一次任务跑动：游标、checkpoint、node_runs、配额 source
```

分层对应（对齐 `02-layering.md`）：

| 层 | 职责 |
|----|------|
| **runtime/agent** | 单个 Agent 怎么跑（拼 context、工具循环、SSE） |
| **orchestration** | Team / 图怎么走下一步、handoff、并行汇合、waiting |
| **app** | HTTP：Team CRUD、触发 run、预览 |
| **capability** | MCP / skill / knowledge / memory 积木复用 |

---

## 5. 三种协作模式（由简到繁）

### 模式 A — 流水线（Pipeline）【已基本具备】

```text
用户 → Agent甲 → Agent乙 → Agent丙 → 结束
```

- **实现**：现有工作流 Agent 节点 + 边；变量经 `ExecutionContext` 传递  
- **适用**：步骤固定（调研→撰写→校对）  
- **产品说法**：「按流程图依次调用多个智能体」

### 模式 B — 主管分派（Supervisor / Orchestrator–Workers）【一期重点】

```text
用户 → 主管 Agent
          ├─ 派给 工人A（可串行）
          ├─ 派给 工人B
          └─ 汇总后回复用户
```

- **实现选项（推荐）**：
  1. Team 配置里指定 `supervisor_agent_id` + `worker_agent_ids[]`
  2. 主管可用工具：`delegate_to_agent` / `ask_agent`（内部再调 runtime）
  3. 工人跑完结果写回 **共享 run state**，主管再读
- **适用**：任务类型多变，需主管判断找谁  
- **参考**：Anthropic Orchestrator–Workers；LangGraph supervisor

### 模式 C — 交接棒（Handoff）【一期可与 B 同做或紧随】

```text
用户 ↔ Agent甲  ——handoff——▶  Agent乙  ↔ 用户
```

- **语义**：对话权转移；用户仍觉得在同一会话，但系统提示与工具集换成乙  
- **实现**：SSE/消息里带 `agent_id` 切换事件；run state 记 `active_agent_id`  
- **适用**：客服分诊、专家接力  
- **参考**：OpenAI Swarm handoff

### 模式 D — 群聊 / 辩论【明确后期】

多 Agent 轮流发言、投票。默认 **不做**；费 Token、难调试，仅 backlog。

---

## 6. 推荐产品形态（贴合 Cogniforge）

不要让用户同时学「工作流」和「神秘多智能体 DSL」。建议：

| 入口 | 给谁 | 背后 |
|------|------|------|
| **智能体工作室** | 做一个「人」 | 单 Agent（现状） |
| **小队（Team）页 / 工作室扩展** | 组一支队 | 选模式 A/B/C + 成员；简单场景可自动生成一张工作流图 |
| **工作流画布** | 要精细控制 | 继续 Vue Flow；新增节点类型见下 |

### 6.1 画布增量（最小）

| 节点/边 | 含义 |
|---------|------|
| 现有 `agent` | 跑指定 Agent 一轮或直至工具结束 |
| 新增 `team`（或 `supervisor`） | 绑定 Team 配置，内部走 B/C |
| 可选边属性 `handoff` | 条件满足时切换 active agent（也可用工具触发，不必画边） |

**已拍板（默认）**：**Team 不强制进画布**（隐形展开）。只在「小队」页配置，执行时引擎内部展开；可观测性靠 `run_state` / 子步骤日志（挂 execution 后写 `node_runs`）。`workflow_id` 可选；要精细控制时再绑可见图或画布 `team` 节点。

### 6.2 共享状态（Run State）最小字段

```text
{
  "team_id": "...",
  "active_agent_id": "...",
  "goal": "用户原始目标",
  "scratchpad": { },          // 主管/工人可写的结构化草稿
  "artifacts": [ ],           // 产出物引用（文稿、检索摘要）
  "handoff_history": [ ]      // 谁→谁、原因、时间
}
```

持久化：塞进现有 **checkpoint payload** / `node_runs.detail`，不新造第四种全局 status。

---

## 7. 运行时约束（必须有，否则会上生产事故）

| 门禁 | 规则 |
|------|------|
| 深度 | 单次 run 内 `delegate` / `handoff` ≤ N（如 8） |
| 步数 | 总 LLM 调用 ≤ 配额与 `max_steps` |
| 循环 | 同一 `from→to` handoff 连续重复报警并打断 |
| 工具 | 工人只能用自己 Agent 上勾选的 MCP/技能；**不能**继承主管全部工具除非显式配置 |
| 权限 | 人审 / 危险工具仍走工作流 wait；Agent **不能**改 RBAC |
| 记忆 | 默认：工人只用本轮委派上下文 + 必要 artifacts；避免把主管全部历史灌给每个工人（防污染，Anthropic 实践） |
| 可观测 | 每步 `node_runs`：agent_id、phase、token、handoff 原因 |
| 配额 | `source=team` 或继续 `workflow`/`agent`，用量页可区分 |

---

## 8. 数据与 API（草案，未建表）

> 真正写迁移前再改 `docs/04-database` / `docs/03-api`；此处只定方向。

### 8.1 表（建议）

`agent_teams`（名可再议）：

| 字段 | 说明 |
|------|------|
| id, user_id, name, description | 基本信息 |
| mode | `pipeline` / `supervisor` / `handoff` |
| supervisor_agent_id | 可空（pipeline 不需要） |
| member_agent_ids | JSON 数组 |
| config | JSON：max_depth、是否允许并行、入口话术等 |
| workflow_id | 可选：绑定一张展开后的工作流（**非必须**；默认隐形展开，见 §12.1） |
| status | active / disabled |

### 8.2 API（示意）

```text
GET/POST   /api/v1/teams
GET/PUT/DELETE /api/v1/teams/:id
POST       /api/v1/teams/:id/chat      # 预览小队（SSE，事件含 agent 切换）
POST       /api/v1/teams/:id/run       # 非对话任务型（可选，可与 workflow execute 合并）
```

执行内核：优先在 **Go `orchestration`** 扩展；单 Agent 循环仍走 **runtime/agent**。Python `AgentExecutor` 不作为多 Agent 编排真相（避免第二套状态）。

---

## 9. 与「构建 / 预览」的关系

| 对象 | 构建 | 预览 |
|------|------|------|
| 单 Agent | 现状：对话填左栏 → 固化 | 现状：像用户试聊；历史 → `chat_conversations` |
| Team | 后期：对话生成成员角色与模式（可选） | 小队预览：能看到「当前是谁在答」；历史同样落 `chat_conversations`（标 `team_id`） |

一期 Team 可先做表单配置，不做「对话建队」。预览与正式 execution 职责见 §12.2。

---

## 10. 分期

| 期 | 交付 | 验收 |
|----|------|------|
| **M0 设计** | 本文档 + 产品口径统一 | 本文 |
| **M1 Pipeline 产品化** | 文档/UI 明确「多 Agent 节点流水线」；可选一键从 Team 生成线性图 | 用户能说清与单 Agent 差别 |
| **M2 Supervisor** | `delegate_to_agent` + Team CRUD + 预览 SSE 带 agent 标签 | 主管能分活，工人结果可汇总；checkpoint 可恢复 |
| **M3 Handoff** | `active_agent_id` 切换 + 前端展示当前角色 | 分诊场景可演示 |
| **M4 并行工人** | 扇出/汇合（需引擎并行或子 run） | 引擎并行 + `layout=parallel` 成图（2026-11-27） |
| **以后** | GroupChat、A2A、跨租户 Agent 市场 | backlog |

---

## 11. 明确不做（本期）

- 引入 LangGraph / CrewAI / AutoGen 作为运行时依赖  
- 嵌 Temporal / n8n  
- Agent 自由改系统权限或跳过审批  
- 无上限并行工人  
- 把「多智能体」做成与工作流无关的第二产品线  

---

## 12. 已拍板决策（原开放问题）

> ~~实现前需拍板~~（2026-11-23）→ 下列四条已定稿。

### 12.1 Team ↔ 可见工作流

| 决定 | 说明 |
|------|------|
| **默认隐形展开** | Team 不强制绑定可见工作流；小队页配置即可跑 |
| `workflow_id` | **可选**；Pipeline 可后期「一键生成线性图」；Supervisor 一期不必落成 Vue Flow 图 |
| 画布 | `team` 节点 / 精细控制为增强项，非预览门槛 |

**产品口径**：Team = 小队说明书；工作流 = 精细编排。

### 12.2 预览会话存哪

| 决定 | 说明 |
|------|------|
| **预览** | 落 `chat_conversations`（与单 Agent 预览一致：左侧历史、回显） |
| **标识** | 用 `team_id`（metadata 或新列）标小队会话；消息可带 `active_agent_id` / 委派摘要 |
| **任务真相** | 正式 run / 可恢复仍走 `workflow_executions` + checkpoint；**不**用聊天 JSONB 当编排 checkpoint |
| **禁止一期** | 预览与 execution 双写全文 |

### 12.3 `delegate_to_agent` / `handoff_to_agent`

| 决定 | 说明 |
|------|------|
| **引擎真相** | **编排原语**（`orchestration/team` 拦截）：深度门禁、成员校验、工人隔离、写 `run_state` |
| **对模型** | 工具列表暴露同名 function，便于主管决策（已有 `OrchestrationToolSpecs`） |
| **禁止** | 做成用户可配 MCP / 进通用工具市场（门禁可被绕过） |

### 12.4 工人失败默认策略

```text
工人失败 → ① 结果（含 error）回传主管（软失败，默认）
         → 可选 ② 同人自动重试 ≤1（仅瞬时类，config 可开）
         → 主管可再 delegate（计入 depth）
         → 触达 max_depth / max_steps → ④ 整 run failed
自动换人 = 默认关（可选 fallback 名单后期再开）
```

| 策略 | 默认 |
|------|------|
| 失败回传主管 | **开** |
| 同人瞬时重试 ≤1 | 可选，默认建议开 1 次或关（实现时写进 `config`） |
| 自动换人 | **关** |
| 超限额整 run 失败 | **开**（硬闸门） |

---

## 13. 总结

| 项 | 决定 |
|----|------|
| 定位 | 在现有 Agent + 工作流上加 **Team + Supervisor/Handoff** |
| 真相源 | 继续 `workflow_executions` + checkpoint（正式任务）；预览历史另见下 |
| Team↔工作流 | **默认隐形展开**；`workflow_id` 可选 |
| 预览存档 | **`chat_conversations`**（标 `team_id`）；不当 checkpoint |
| delegate/handoff | **编排原语** + 工具壳暴露给主管 |
| 工人失败 | **回传主管**；可选瞬时同人重试；超限额整 run 失败；自动换人默认关 |
| 一期模式 | Pipeline 说清楚 + Supervisor；Handoff 紧随 |
| 参考主轴 | LangGraph 状态图思想、OpenAI Handoff、Anthropic Orchestrator–Workers |
| 代码 | M1–M3 主干 + `/run` + 画布 `team` + 一键成图已落地；并行工人仍待 |
