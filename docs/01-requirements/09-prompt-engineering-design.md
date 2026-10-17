# 提示词工程（阶段十七 / 工业落地）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-10-17 | v2.3 | 来源字段；同 Agent 再导入发新版本，不加 (2) | orjrs |
| 2026-10-17 | v2.2 | 无独立编码：同用户 name 唯一 | orjrs |
| 2026-10-17 | v2.1 | 补齐剩余 backlog 清单；与开发计划阶段十七任务表对齐 | orjrs |
| 2026-10-17 | v2.0 | 工业闭环：发布态、变量定义、渲染、Agent 绑定、导入、评估代入 | orjrs |
| 2026-10-17 | v1.3 | 版本删除 / 设当前 / 双版对比 / {{var}} 预览 | orjrs |
| 2026-10-15 | v1.2 | 评估页提示词/版本下拉，免手抄 version id | orjrs |
| 2026-10-15 | v1.1 | 落地 CRUD + Web /prompts；评估联动已接线 | orjrs |
| 2026-10-15 | v1.0 | 立项：提示词库 + 版本；与评估体系联动（target=prompt） | orjrs |

## [变更] 来源字段 + 同 Agent 再导入（2026-10-17）

- **变更原因**：同一 Agent 多次导入堆 `名称 (2)` 不合适；需要可追溯来源
- **设计**：
  - 字段：`source_type`（manual/agent/skill）、`source_id`、`source_name`（展示快照）
  - 约束：`uk_prompts_user_source (user_id, source_type, source_id)`；手建 `source_id` 为空（可多条）
  - **同一 Agent 再导入**：在原提示词上发新 published 版本，不新建资产
  - 名称仍按用户唯一；仅首次导入且与其它提示词撞名时才加 `(2)`
- **UI**：详情显示「来源：Agent「xxx」」；导入 toast 区分新建 / 同步新版

## [变更] 名称唯一、无独立编码（2026-10-17）

- **变更原因**：不需要 slug/编码字段；导入/新建若同名会撞车
- **设计**：
  - **对人只展示 `name`**（版本用 `v1/v2`，不展示 UUID）
  - **同一 `user_id` 下 `name` 唯一**（`uk_prompts_user_name`）
  - **新建/改名**：冲突 → 报错「该名称已存在」
  - **删除**：软删前改名并清来源，原名可再建
- **包含代码**：`model.Prompt` 联合唯一索引；`prompt.Service` Create/Update/Import/Delete；Web 去掉版本 UUID 展示

## [变更] 工业闭环（2026-10-17）

- **变更原因**：一期仅 CRUD，无法支撑生产：无发布门禁、无变量契约、Agent 仍内联散落、评估无法代入
- **包含代码**：`internal/model/prompt.go`、`internal/prompt/*`；`agents.prompt_id`；`eval` Variables；Web `/prompts` `/agents` `/evals`
- **能力**：
  1. 版本 `draft` / `published`；发布后才能设为当前；新建默认已发布；新版本默认可存草稿
  2. 资产级变量定义（name/default/required）+ `POST .../render` + 评估/Agent 运行时替换 `{{var}}`
  3. Agent 绑定 `prompt_id`（优先于内联 system_prompt / Skill）
  4. `POST /prompts/import` 从 Agent 导入；`GET .../usages` 看绑定关系
  5. 删版本 / 设当前 / 双版对比（延续 v1.3）
- **不做**：提示词市场、APE 自动优化、多轮模板编排器

---

## 1. 一句话

**工业提示词工程** = 可版本发布的 system prompt 资产 + 变量契约 + 绑定到 Agent/评估，用同一套题回归哪一版更好。

---

## 2. 概念

| 概念 | 说明 |
|------|------|
| Prompt | 资产：名称、变量定义、当前已发布版本指针 |
| PromptVersion | 正文 + changelog；`draft` 或 `published` |
| 当前版 | `current_version_id`，必须是已发布版；Agent/默认评估取它 |
| 变量 | `{{name}}`；定义在 Prompt；运行时用 defaults / 请求 Variables 替换 |
| Agent 绑定 | `agents.prompt_id`；对话优先解析绑定提示词当前已发布版 |

---

## 3. 数据模型

```
prompts
  id, user_id, name, description   -- 无 code/slug；uk_prompts_user_name(user_id,name)
  source_type, source_id, source_name  -- manual/agent/skill；uk_prompts_user_source
  current_version_id
  variables jsonb  -- [{name, description, default, required}]
  tags jsonb
  created_at, updated_at, deleted_at

prompt_versions
  id, prompt_id, user_id
  version int                      -- 对人展示 v1/v2，不展示 UUID
  status  draft|published
  content, changelog
  published_at, created_at

agents
  ... existing ...
  prompt_id  -- 可选绑定
```

---

## 4. API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/prompts` | 列表 / 创建 |
| POST | `/api/v1/prompts/import` | `{ agent_id }` 从 Agent 导入 |
| GET/PUT/DELETE | `/api/v1/prompts/:id` | 详情 / 元数据与变量 / 删除 |
| GET/POST | `/api/v1/prompts/:id/versions` | 列表 / 新建（`publish` 可选） |
| PUT | `/api/v1/prompts/:id/current-version` | 设当前（仅 published） |
| POST | `/api/v1/prompts/:id/render` | 渲染 `{ version_id?, variables }` |
| GET | `/api/v1/prompts/:id/usages` | 绑定了该提示词的 Agent |
| GET/DELETE | `/api/v1/prompts/versions/:id` | 详情 / 删版 |
| POST | `/api/v1/prompts/versions/:id/publish` | 发布并设当前 |

评估：`POST /evals/suites/:id/runs` 增加 `variables` map。

Agent：创建/更新可带 `prompt_id`。

---

## 5. 运行时优先级（Agent 对话）

1. 若 `prompt_id` 有值 → 当前已发布版 + 变量默认值渲染  
2. 否则若有 Skill → Skill 拼装  
3. 否则内联 `system_prompt`  
4. 再否则默认助手文案  

---

## 6. 分期状态

### 6.1 已完成（工业闭环，代码已写；须 push + 部署后线上生效）

| 子项 | 内容 | 状态 |
|------|------|------|
| 17.1–17.4 | CRUD + 评估选版 | ✅ |
| 17.5a | 正文 `{{var}}` 预览 | ✅ |
| 17.5b | 从 Agent 导入 | ✅ |
| 17.5c | 变量定义 + render + 评估/Agent 运行时渲染 | ✅ |
| 17.6 | 双版并排对比 | ✅ |
| 17.7 | 删版 / 设当前 / 详情删整条 | ✅ |
| 17.8 | draft / publish 门禁 | ✅ |
| 17.9 | Agent `prompt_id` 绑定 + usages | ✅ |

### 6.2 剩余待做（按优先级）

| 优先级 | 子项 | 内容 | 说明 |
|--------|------|------|------|
| P0 | 17.0 | **合入并部署**工业闭环 | 代码已 push；等/确认 CD 冒烟 |
| P1 | 17.10 | **试跑**：选版本 + 填变量 + 调模型看一轮回答 | 不必开评估集也能验人设；可挂 `/prompts` 或复用 Playground |
| P1 | 17.12 | **评估变量表单**：按 Prompt 变量定义生成输入框 | 现在评估页是手写 JSON，易错 |
| P1 | 17.13 | **顶栏显示邮箱/区分同名用户** | 已出现两个 `name=admin` 导致「同站看不到提示词」误判 |
| P2 | 17.11 | 从 **Skill** 导入提示词 | 与 17.5b Agent 导入对称 |
| P2 | 17.14 | Agent 对话请求可传 `variables` | 目前绑定 Prompt 只用定义里的 default |
| P2 | 17.15 | 版本 **行级 diff**（在并排之外） | 长文对比更清楚 |
| P2 | 17.16 | Tags 编辑 UI + 列表筛选 | 模型字段已有 `tags`，前端未做 |
| P3 | 17.17 | 归档 / 废弃版本；回滚=基于旧版发新 published | 运维友好 |
| P3 | 17.18 | 工业能力 e2e（发布/绑定/导入/变量评估） | 现有 e2e 偏一期 |

### 6.3 明确不做（本阶段）

提示词市场、APE 自动优化、多人协作审批流、跨租户共享库、多轮对话模板编排器。

---

## [变更] backlog 对齐（2026-10-17）

- **变更原因**：工业闭环代码已落，但开发计划表仍标 17.5/17.6 未开始；需写清「还要做什么」
- **影响文档**：本文件 v2.1；`docs/99-dev-plan/01-development-plan.md` 阶段十七；cogniforge-web 开发计划

---

## 7. 验收（非技术）

**已具备（部署后验）**：

1. 新建提示词，发草稿 → 发布 → 设为当前  
2. 定义 `{{brand}}`，评估填变量跑卷，回答里用到替换后的值  
3. Agent 绑定该提示词后对话，不再靠内联文案  
4. 从旧 Agent 一键导入成提示词资产  
5. 两版勾选对比；删错版可删（至少留一版）  

**剩余验收（做完 17.10/17.12/17.13 后再验）**：

6. 提示词页内可试跑一轮，变量用表单填  
7. 顶栏能分清同名账号（看到邮箱）  
