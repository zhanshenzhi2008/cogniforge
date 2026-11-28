# Mission

## 一句话

CogniForge 是面向开发者与企业的 **AI 应用开发与运营控制台**：统一接入模型、搭 Agent / 小队、编排工作流、管知识库与配额，并在现网可观测、可鉴权。

## 为谁

| 角色 | 要什么 |
|------|--------|
| AI / 应用开发者 | Playground、Agent 工作室、工作流画布、模型与 MCP 配置 |
| 平台 / 运维 | 部署可跑、用量与监控、密钥与审计 |
| 管理员 | 用户 / 角色 / RBAC；可选知识密级与 clearance（金融扩展开关） |

## 做什么（当前产品面，以代码与 Web 导航为准）

- **对话**：Playground 流式聊天、排队/插入、附件 + RAG、短期窗口与长期记忆
- **智能体**：Agent CRUD、工作室预览、知识库绑定、版本回滚、发布渠道文案（Web/API/嵌入）
- **小队（Team）**：Supervisor / Handoff / Pipeline；预览 chat、任务 `/run`、画布 `team` 节点、一键成图（含并行扇出）
- **工作流**：Vue Flow 画布 + Go 引擎（条件/循环/HTTP/代码/检索/子流程/等待审批、调试与 checkpoint）
- **知识库**：入库、分段、pgvector 检索；文档生命周期与有效期；可选密级强制检索
- **配置**：模型供应商（本仓唯一 Key 出口）、API Key、MCP/内置工具、提示词、评测
- **治理**：登录 JWT、RBAC、Playground 配额、用量页、监控与密级审计 Tab

## 非目标（Non-goals）

- 不以 LangGraph / CrewAI / AutoGen / Temporal / n8n 作为运行时依赖
- 不把「多智能体」做成与工作流无关的第二产品线
- 不启用 `services/*` 下 Java 微服务存根作现网路径（用户/认证/计费等仍由 Go 承担）
- 不做 Agent 自由改系统权限或跳过审批
- 不做无上限并行工人
- README 里曾写的 **微调训练、插件市场、Kafka 分析栈、Kubernetes 必选** 等 — 当前代码未作为交付承诺，标为 Later / `_TBD_`，不写入本期 MVP

## MVP（已基本落地的最小可用面）

1. 登录后可用 Playground（配额闸门）+ 配置至少一个对话模型
2. 可创建 Agent 并在工作室/Playground 对话（含可选 RAG）
3. 可建知识库、入库并语义检索（pgvector；embedding 经 Go）
4. 可画工作流并试跑（含等待/审批路径）
5. 可建 Team 预览与 `/run`，可选展开为工作流图

## 成功怎么看

- 上述 MVP 在本地与部署环境可冒烟通过（清单见运维文档）
- 模型 Key 只经 Go；Python 不落明文供应商 Key
- 跨仓 API 行为有 validation / 冒烟可勾，而不是「文档写了但线上未验」
