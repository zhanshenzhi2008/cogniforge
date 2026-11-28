# Roadmap

来源：当前代码能力 + README 过时承诺的校正 + 旧 `docs/99-dev-plan` / 需求草案中的**未勾选项**。  
旧 `docs/` 只作归档参考；活优先级以本文件为准。

## Now（最多 3 条）

1. **线上冒烟勾选：工作流审批路径（W3）+ 智能体工作室（A1–A12）**  
   - **状态**：blocked  
   - **原因**：清单与修复已就绪（见 `docs/06-devops/01-deployment.md`），需在已部署环境由人工逐项勾选；非代码缺口  
   - **完成定义**：运维文档两处清单从「待线上勾选」改为通过，并回写本条 complete

2. **生产验收：Playground / Agent PDF 附件问答**  
   - **状态**：blocked（部分依赖运维）  
   - **原因**：依赖线上向量默认供应商、uploads 挂载与 RAG 健康；开发计划仍留「人工验收」未勾  
   - **完成定义**：固定样例 PDF 可问可答；失败原因记入 known-issues 或本 roadmap，而不是静默挂起

3. ~~**知识库组织共享（`visibility`：private → org）**~~  
   - **状态**：complete（2026-11-28）  
   - **口径**：不做部门/组织表；`org` = 本部署全体登录用户**可读**，仅所有者可写  
   - **规格**：[`specs/2026-11-28-kb-org-visibility/`](./2026-11-28-kb-org-visibility/)

## Next

- **Team 工人失败「自动换人」**（可选 fallback 名单；设计默认关）— 多 Agent 文档多次「未做」
- **Crawl4AI 独立爬虫服务**（跨 `cogniforge-ai`：`crawl/` + 第二容器；Go 再接线）— 目标日曾写 2026-11-23，仍未做
- **`code_execute` / 代码节点容器强隔离**（现为本地/同源沙箱；生产强隔离可选）
- **向量侧元数据下推**：密级 / 有效期在召回前过滤（少召回再丢弃）
- **导出 / 下载走密级与审计**（不只检索路径）
- **Team / LLM 真·逐字 SSE**（画布与预览；现有进度钩与轮询预览，非完整独立事件流）
- **库级整体生效窗**（按需；多数场景文档级已够）

## Later

- 知识库真·部门 / 组织成员 ACL（本轮 `org` 仅为全站登录用户可读）
- 工作室 **虚拟机运行时**（套餐门禁 + 沙箱；UI 占位已做）
- 云端 TTS / ASR（浏览器 TTS/STT 已做）
- 发布渠道：公开免登、IM 真接入（现为文案/占位）
- GroupChat、A2A、跨租户 Agent 市场
- Milvus 等独立向量库（可选；现网 pgvector）
- 微调训练、插件市场（README 曾宣传，代码未交付）
- 启用 Java `services/*` 微服务路径（当前明确不走）
- Kafka / ClickHouse 分析栈
- `_TBD_`：多区域 / 多租户计费产品化（配额已在 Go；充值不做）

## Recently complete（便于对照，不占 Now 名额）

- 知识库 `visibility=org` 本站共享（可读；仅所有者可写）（2026-11-28）
- 工作流节点大量配置支持 `{{变量}}`、下游 `last_*` 写出（2026-11-27～28）
- 工作流调试 UX、绑定校验、条件句柄分支、同层并行、子流程 MVP
- Team M1–M4 主干：CRUD、chat、run、画布节点、一键成图、并行扇出
- 知识生命周期 / 有效期过滤；密级 + clearance + 审计（开关控制）
- Chat 短期窗口 + 长期记忆；Playground 配额；MCP 内置工具热路径

## 归档说明

- `docs/99-dev-plan/01-development-plan.md` 变更流水极长，**不再**当作活 backlog；未完成项已筛入上表  
- `docs/99-dev-plan/03-known-issues.md` 可继续作排障笔记；修生产缺陷时优先挂本 roadmap 或开 feature 目录  
- 行为开发请开 `specs/YYYY-MM-DD-<slug>/`，不要往旧 PRD 树里「同步规格」
