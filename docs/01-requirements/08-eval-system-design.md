# 设计评估体系（阶段十六）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-10-15 | v1.3 | 16.4.2：target=agent 注入 Agent.tools 并走 ChatWithTools | orjrs |
| 2026-10-15 | v1.2 | Run 支持 target=prompt + system_prompt 覆盖；联动提示词工程 | orjrs |
| 2026-10-15 | v1.1 | 单测覆盖 Suite/Case/Run/Judge；JSONB Valuer 修 SQLite；Web e2e | orjrs |
| 2026-10-15 | v1.0 | 立项：模型/智能体评测集 + 运行 + 打分；对齐 FT-020 | orjrs |

## [变更] Agent 评估走 MCP 工具循环（2026-10-15）

- **变更原因**：16.4.2；Agent 考工具题时不能只当普通 Chat
- **包含代码**：`internal/eval/service.go`（`callCase` + InjectMCPTools）；`handler.SetMCPRegistry`；router 注入
- **变更前 vs 变更后**：~~agent 仅 system+Chat~~（2026-10-15）→ 读取 `agents.tools`，有工具则 `ChatWithTools`
- **限制**：仍同步整卷；不跑 RAG；与 Agent 页勾选的 MCP Server 列表一致

## [变更] 提示词目标 + system_prompt（2026-10-15）

- **变更原因**：同一套题要对比不同 system prompt；临时实验不必改 Agent
- **关系**：提示词资产见 `docs/01-requirements/09-prompt-engineering-design.md`
- **包含代码**：`internal/eval/{dto,service}.go`；`EvalRun.model` / `system_prompt`；Web `/evals` 运行表单
- **变更前 vs 变更后**：~~仅 model|agent~~（2026-10-15）→ 支持 `prompt`（version_id + model）与请求体 `system_prompt` 覆盖

## [变更] 单测 + e2e（2026-10-15）

- **变更原因**：评估模块要可回归；SQLite 下 JSONBMap Valuer 未满足 `driver.Valuer` 导致 Case 插入失败
- **包含代码**：`internal/eval/service_test.go`、`judge_test.go`；`internal/model/types.go`；Web `e2e/evals.spec.ts`
- **变更前 vs 变更后**：~~无自动化测试~~（2026-10-15）→ Go `go test ./internal/eval` 绿；Playwright 建集→加题→跑分绿

## [变更] 立项（2026-10-15）

- **变更原因**：产品需求 FT-020「模型评估」；对话/Agent/工具链路已齐，需要可重复的质量评估，而不是靠人工点点点
- **包含代码**：`internal/model/eval.go`、`internal/eval/*`；Web `/evals`；开发计划阶段十六
- **影响范围**：PostgreSQL 新表；Go API；顶栏配置下增加「评估」
- **变更前 vs 变更后**：~~无评测模块~~（2026-10-15）→ 有评测集 / 用例 / 运行记录

---

## 1. 一句话说明

**评估体系** = 准备一组「题目」（评测用例）→ 选模型或 Agent 跑一遍 → 按规则打分 → 看通过率和明细。

给非技术同学：像考试卷。卷子是评测集，每道题是用例，交卷是一次运行，分数是结果。

---

## 2. 核心概念

| 概念 | 英文 | 说明 |
|------|------|------|
| 评测集 | EvalSuite | 一组用例的集合（如「客服 FAQ v1」） |
| 用例 | EvalCase | 一条输入 + 期望（关键词/全文）+ 打分方式 |
| 运行 | EvalRun | 某次对某模型/Agent 的整卷作答 |
| 结果 | EvalResult | 单题得分、模型实际输出、耗时 |

### 打分方式（一期）

| 类型 | 含义 | 适用 |
|------|------|------|
| `contains` | 回答里包含期望关键词（逗号分隔任一命中即过，可配置需全部） | 事实问答、工具是否提到某点 |
| `exact` | 去空白后完全相等 | 短答案、枚举 |
| `llm_judge` | 二期：用裁判模型打 0–1 分 | 开放式质量 |

一期 **只实现 contains / exact**；`llm_judge` 存库可写，运行时返回 `skipped`。

---

## 3. 数据模型

```
eval_suites
  id, user_id, name, description, created_at, updated_at

eval_cases
  id, suite_id, user_id
  name, input（用户提问）
  expected（期望文本或关键词）
  judge_type: contains | exact | llm_judge
  judge_config JSONB  -- 如 {"match":"any"|"all"}
  weight float        -- 默认 1
  sort_order int
  created_at, updated_at

eval_runs
  id, suite_id, user_id
  target_type: model | agent | prompt
  target_id     -- model 名 / agent_id / prompt_version_id
  model         -- target=prompt 时的聊天模型；其它目标可作覆盖记录
  system_prompt -- 本次实际使用的 system（便于回看）
  status: pending | running | completed | failed
  pass_count, fail_count, skip_count
  score float       -- 加权通过率 0–1
  error_message
  started_at, finished_at, created_at

eval_results
  id, run_id, case_id
  status: pass | fail | skip | error
  score float
  actual_output text
  latency_ms int
  detail JSONB      -- 命中关键词等
  created_at
```

索引：`idx_eval_cases_suite_id`、`idx_eval_runs_suite_id`、`idx_eval_runs_user_id`、`idx_eval_results_run_id`。

---

## 4. API（一期）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/evals/suites` | 列表 / 创建评测集 |
| GET/PUT/DELETE | `/api/v1/evals/suites/:id` | 详情 / 更新 / 删除（级联用例） |
| GET/POST | `/api/v1/evals/suites/:id/cases` | 用例列表 / 新增 |
| PUT/DELETE | `/api/v1/evals/cases/:id` | 更新 / 删除用例 |
| POST | `/api/v1/evals/suites/:id/runs` | 发起运行 `{ target_type, target_id, model?, system_prompt? }` |
| GET | `/api/v1/evals/runs` | 运行列表（可按 suite 过滤） |
| GET | `/api/v1/evals/runs/:id` | 运行详情 + results |

认证：JWT。资源按 `user_id` 隔离。

---

## 5. 运行流程（一期同步）

```
POST runs
  → status=running
  → 解析 system / model：
       target=model  → target_id 即模型名；默认 system
       target=agent  → 读 Agent system_prompt + model（暂不跑 MCP 工具循环，二期对齐）
       target=prompt → target_id=prompt_version_id；必填 model；system=版本正文
       若请求带 system_prompt → 覆盖以上 system（临时实验）
  → for each case:
       ChatService.Chat(system+user)
       judge contains/exact → pass/fail
  → 汇总 score = Σ(weight×pass) / Σ(weight)
  → status=completed
```

超时：单题默认 60s；整卷建议用例 ≤ 50（前端提示）。

---

## 6. 前端

- 路由：`/evals`（配置下拉）
- 页：评测集列表 → 详情（用例表 +「运行」）→ 运行历史与明细
- 一期：CRUD + 发起运行 + 看结果；不做图表大屏

---

## 7. 分期

| 子项 | 内容 | 状态 |
|------|------|------|
| 16.1 | 数据模型 + Suite/Case CRUD API | 本期 |
| 16.2 | Run：模型同步评测 + contains/exact | 本期 |
| 16.3 | 前端评估页 | 本期 |
| 16.4 | Agent 目标 + 工具题（需 MCP） | ✅ target=agent + ChatWithTools（按 Agent.tools） |
| 16.4b | Prompt 目标 + system_prompt 覆盖 | 本期（联动阶段十七） |
| 16.5 | llm_judge 裁判模型 | 二期 |
| 16.6 | 异步队列 + 进度 SSE | 二期 |

**不在本期**：微调数据集复用、与 fine-tune 联动、跨用户共享公开榜、A/B 自动选模。

---

## 8. 验收（给非技术同学）

1. 新建评测集，加 2 道「包含关键词」题  
2. 对当前默认模型点「运行」  
3. 能看到每题通过/失败和模型原话  
4. 删掉评测集后用例与运行记录一并清掉（或软删一致）
