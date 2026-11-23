# 多 Agent 框架设计（阶段十九草案）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-23 | v0.4 | chat 回传 run_state；handoff 后按 active 换人设 | orjrs |
| 2026-11-23 | v0.3 | M2 起步：POST /teams/:id/chat + delegate/handoff 工具 | orjrs |
| 2026-11-23 | v0.2 | M1 起步：agent_teams 表 + Team CRUD + RunState/handoff 原语 | orjrs |
| 2026-11-23 | v0.1 | 立项：模式分层、与现有工作流关系、参考来源、分期不做清单 | orjrs |

## [变更] run_state 回传换人设（2026-11-23）

- **已落地**：请求可带上一轮 `run_state`；`ResumeRunState` 校验成员后按 `active_agent_id` 加载人设；Web 预览回传
- **未做**：服务端 Redis 持久化 run、画布 Team 节点
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

一期也可：**Team 不进画布**，只在「小队」页配置，执行时引擎内部展开成临时图——降低前端成本，但可观测性要靠 `node_runs` 打出子步骤。

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
| workflow_id | 可选：绑定一张展开后的工作流 |
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
| 单 Agent | 现状：对话填左栏 → 固化 | 现状：像用户试聊 |
| Team | 后期：对话生成成员角色与模式（可选） | 小队预览：能看到「当前是谁在答」 |

一期 Team 可先做表单配置，不做「对话建队」。

---

## 10. 分期

| 期 | 交付 | 验收 |
|----|------|------|
| **M0 设计** | 本文档 + 产品口径统一 | 本文 |
| **M1 Pipeline 产品化** | 文档/UI 明确「多 Agent 节点流水线」；可选一键从 Team 生成线性图 | 用户能说清与单 Agent 差别 |
| **M2 Supervisor** | `delegate_to_agent` + Team CRUD + 预览 SSE 带 agent 标签 | 主管能分活，工人结果可汇总；checkpoint 可恢复 |
| **M3 Handoff** | `active_agent_id` 切换 + 前端展示当前角色 | 分诊场景可演示 |
| **M4 并行工人** | 扇出/汇合（需引擎并行或子 run） | 可选 |
| **以后** | GroupChat、A2A、跨租户 Agent 市场 | backlog |

---

## 11. 明确不做（本期）

- 引入 LangGraph / CrewAI / AutoGen 作为运行时依赖  
- 嵌 Temporal / n8n  
- Agent 自由改系统权限或跳过审批  
- 无上限并行工人  
- 把「多智能体」做成与工作流无关的第二产品线  

---

## 12. 开放问题（实现前需拍板）

1. Team 是否必须绑定一张可见工作流，还是允许「隐形展开」？  
2. 预览小队时，会话历史存在 `chat_conversations` 还是仅 execution？  
3. `delegate_to_agent` 算工具调用（进 MCP 风格）还是编排原语（引擎内置）？**建议：编排原语**，工具列表里可对主管暴露同名能力。  
4. 工人失败：主管重试 / 换人 / 整 run 失败——默认策略？  

---

## 13. 总结

| 项 | 决定 |
|----|------|
| 定位 | 在现有 Agent + 工作流上加 **Team + Supervisor/Handoff** |
| 真相源 | 继续 `workflow_executions` + checkpoint |
| 一期模式 | Pipeline 说清楚 + Supervisor；Handoff 紧随 |
| 参考主轴 | LangGraph 状态图思想、OpenAI Handoff、Anthropic Orchestrator–Workers |
| 代码 | **先设计、后排期**；落地时再改库表/API/前端文档版本号 |
