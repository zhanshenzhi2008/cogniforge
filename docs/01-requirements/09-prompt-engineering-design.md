# 提示词工程（阶段十七）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-10-15 | v1.2 | 评估页提示词/版本下拉，免手抄 version id | orjrs |
| 2026-10-15 | v1.1 | 落地 CRUD + Web /prompts；评估联动已接线 | orjrs |
| 2026-10-15 | v1.0 | 立项：提示词库 + 版本；与评估体系联动（target=prompt） | orjrs |

## [变更] 评估页选提示词版本（2026-10-15）

- **变更原因**：17.4 体验补齐；手抄 ID 易错
- **包含代码**：cogniforge-web `pages/evals.vue`
- **变更前 vs 变更后**：~~粘贴 version id~~（2026-10-15）→ 下拉选提示词与版本

## [变更] 一期落地（2026-10-15）

- **变更原因**：立项后补齐可跑通路径
- **包含代码**：`internal/model/prompt.go`、`internal/prompt/*`、`cmd/server` AutoMigrate；Web `pages/prompts.vue`、`composables/usePrompts.ts`；评估 Run 扩展
- **变更前 vs 变更后**：~~仅设计文档~~（2026-10-15）→ API + 页面 + 评估可选 prompt / 临时 system_prompt
- **跨仓库**：cogniforge-web `/prompts`、导航「提示词」、`/evals` 运行表单

## [变更] 立项（2026-10-15）

- **变更原因**：评估只能考「模型/Agent」，无法系统化对比多版 system prompt；产品需要提示词资产与回归
- **关系**：评估体系负责「考试」；本模块负责「提示词试卷上的考生人设」
- **包含代码**：`internal/model/prompt.go`、`internal/prompt/*`；Web `/prompts`；评估 Run 支持 `system_prompt` / `target_type=prompt`
- **变更前 vs 变更后**：~~提示词散落在 Agent 字段~~（2026-10-15）→ 可版本化的提示词库，并可一键拿去跑评估

---

## 1. 一句话

**提示词工程** = 把 system prompt 当成可管理的产品资产（草稿、版本、说明），并能在评估里用「同一套题」对比不同版本。

---

## 2. 概念

| 概念 | 说明 |
|------|------|
| Prompt | 一条提示词资产（名称、描述、当前版本） |
| PromptVersion | 某次定稿内容（正文 + 变更说明 + 版本号） |
| 评估联动 | Run 时 `target_type=prompt` + `target_id=version_id`，另传 `model` 指定聊天模型 |

临时对比（不入库）：评估页填 `system_prompt` 覆盖，仍可用 `target_type=model`。

---

## 3. 数据模型

```
prompts
  id, user_id, name, description
  current_version_id
  created_at, updated_at, deleted_at

prompt_versions
  id, prompt_id, user_id
  version int          -- 从 1 递增
  content text         -- system prompt 正文
  changelog text
  created_at
```

索引：`idx_prompts_user_id`、`idx_prompt_versions_prompt_id`。

---

## 4. API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/v1/prompts` | 列表 / 创建（可带初版 content） |
| GET/PUT/DELETE | `/api/v1/prompts/:id` | 详情（含当前版）/ 更新元数据 / 删除 |
| GET/POST | `/api/v1/prompts/:id/versions` | 版本列表 / 新建版本（自动 version+1，并更新 current） |
| GET | `/api/v1/prompts/versions/:id` | 单版本详情 |

---

## 5. 与评估的约定

`POST /api/v1/evals/suites/:id/runs`：

```json
{
  "target_type": "prompt",
  "target_id": "<prompt_version_id>",
  "model": "deepseek-chat",
  "system_prompt": "可选；若传则覆盖版本正文（临时实验）"
}
```

或：

```json
{
  "target_type": "model",
  "target_id": "deepseek-chat",
  "system_prompt": "你是严格的客服，只答政策相关…"
}
```

---

## 6. 分期

| 子项 | 内容 | 状态 |
|------|------|------|
| 17.1 | Prompt / Version 表 + CRUD API | ✅ 完成 |
| 17.2 | 评估 Run：`system_prompt` 覆盖 + `target=prompt` | ✅ 完成 |
| 17.3 | 前端 `/prompts` 列表与版本 | ✅ 完成 |
| 17.4 | 评估页选提示词版本 / 临时 prompt | ✅ 完成 |
| 17.5 | 变量 `{{var}}`、从 Skill/Agent 导入 | 二期 |
| 17.6 | 双版本并排对比 UI | 二期 |

**不在本期**：提示词市场、自动优化（APE）、多轮对话模板编排器。

---

## 7. 验收（非技术）

1. 配置 → 提示词：新建一条，再发一版改文案  
2. 评估里选「提示词版本」+ 模型跑卷，或粘贴临时 system prompt 跑卷  
3. 两版分数能分开看，知道哪版更好  
