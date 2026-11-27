# CogniForge 产品 UI 重设计（shadcn-vue）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-11-27 | v1.159 | 提示词标签筛选「全部」改哨兵值，避免 USelect 空串 500 | orjrs |
| 2026-11-27 | v1.158 | Playground 插入发送改用 CfConfirmModal | orjrs |
| 2026-11-27 | v1.157 | 知识库弃编辑改用 CfConfirmModal | orjrs |
| 2026-11-27 | v1.156 | 工作室版本回滚改用 CfConfirmModal | orjrs |
| 2026-11-27 | v1.155 | Playground 删对话改用 CfConfirmModal | orjrs |
| 2026-11-27 | v1.154 | 评估页删除套件/用例改用 CfConfirmModal | orjrs |
| 2026-11-27 | v1.153 | Cron 六段校验；调度/回滚改用 CfConfirmModal | orjrs |
| 2026-11-27 | v1.152 | Webhook 地址校验；HTTP 副文案显示非默认超时 | orjrs |
| 2026-11-27 | v1.151 | Playwright 默认不复用端口，避免 e2e 404 | orjrs |
| 2026-11-27 | v1.150 | 画布副文案补充：LLM 流式 / 检索 topK / 等待超时 | orjrs |
| 2026-11-27 | v1.149 | 循环节点 for_each 配置与画布副文案 | orjrs |
| 2026-11-27 | v1.148 | 提示词版本归档 / 回滚发版（17.17） | orjrs |
| 2026-11-27 | v1.147 | 开始节点画布副文案（入参变量） | orjrs |
| 2026-11-27 | v1.146 | 循环节点画布副文案（次数/条件） | orjrs |
| 2026-11-27 | v1.145 | 延时/等待节点画布副文案（时长与原因） | orjrs |
| 2026-11-27 | v1.144 | 节点必填绑定 isWorkflowBindUnbound 单测 | orjrs |
| 2026-11-27 | v1.143 | LLM 节点模型空态提示与刷新 | orjrs |
| 2026-11-27 | v1.142 | 条件节点未填字段醒目标记与校验 | orjrs |
| 2026-11-27 | v1.141 | 代码节点未填代码醒目标记与校验 | orjrs |
| 2026-11-27 | v1.140 | HTTP 节点未填 URL 醒目标记与校验 | orjrs |
| 2026-11-27 | v1.139 | LLM 节点未选模型醒目标记与校验 | orjrs |
| 2026-11-27 | v1.138 | 提示词版本行级 diff（17.15） | orjrs |
| 2026-11-27 | v1.137 | 提示词 Tags 编辑 + 列表筛选（17.16） | orjrs |
| 2026-11-27 | v1.136 | 角色管理列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.135 | 用户管理列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.134 | 评估列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.133 | 画布已绑定节点显示资源名副文案 | orjrs |
| 2026-11-27 | v1.132 | 画布顶栏未绑定数量芯片（可点开首个） | orjrs |
| 2026-11-27 | v1.131 | 画布未绑定节点虚线框 + ! 角标；补 subflow 节点模板 | orjrs |
| 2026-11-27 | v1.130 | 执行/调试前同样校验节点必填绑定 | orjrs |
| 2026-11-27 | v1.129 | 保存时校验节点必填绑定；子流程写入 config | orjrs |
| 2026-11-27 | v1.128 | 画布检索节点知识库空态与刷新 | orjrs |
| 2026-11-27 | v1.127 | 画布子流程节点空态提示与刷新 | orjrs |
| 2026-11-27 | v1.126 | 画布 Agent 节点空态提示与刷新 | orjrs |
| 2026-11-27 | v1.125 | MCP 列表卡片改用 CfResourceCard | orjrs |
| 2026-11-27 | v1.124 | 画布多智能体节点：空态提示 + 文案对齐竞品 | orjrs |
| 2026-11-27 | v1.123 | 小队 max_depth/max_steps 限额可配 | orjrs |
| 2026-11-27 | v1.122 | 小队 worker_retry_once 创建/列表开关 | orjrs |
| 2026-11-27 | v1.121 | API Key 列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.120 | 提示词列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.119 | 工作流列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.118 | 小队列表 CfResourceCard + CfViewModeToggle | orjrs |
| 2026-11-27 | v1.117 | 工作室虚拟机/IM 即将推出占位 | orjrs |
| 2026-11-27 | v1.116 | MCP 列表 CfViewModeToggle 卡片/表格 | orjrs |
| 2026-11-27 | v1.115 | 对话输入浏览器语音听写（STT） | orjrs |
| 2026-11-27 | v1.114 | 对话助手回复浏览器朗读（TTS） | orjrs |
| 2026-11-27 | v1.113 | 知识库列表 CfViewModeToggle 卡片/表格 | orjrs |
| 2026-11-27 | v1.112 | CfResourceCard：agents/knowledge 卡片壳 | orjrs |
| 2026-11-27 | v1.111 | 猜你想问：空态推荐与一键填入 | orjrs |
| 2026-11-27 | v1.110 | 工作室发布渠道：网页/API/嵌入 | orjrs |
| 2026-11-27 | v1.109 | 工作室配置版本历史侧栏与回滚 | orjrs |
| 2026-11-27 | v1.108 | 工作室提示词 / 插变量、@ 绑定技能工具知识库 | orjrs |
| 2026-11-27 | v1.107 | 「了解更多」链到 /help/agent-studio | orjrs |
| 2026-11-27 | v1.106 | 工作室顶栏「对话日志」跳转 Playground | orjrs |
| 2026-11-27 | v1.105 | 工作室预览附件上传（allow_attachments） | orjrs |
| 2026-11-27 | v1.104 | 工作室可多选绑定知识库（rag_kb_ids） | orjrs |
| 2026-11-27 | v1.103 | 工作室开场白 + 预设问题；Playground 空态展示 | orjrs |
| 2026-11-27 | v1.102 | Playground 对话改走 useAgentChat / useChatStream | orjrs |
| 2026-11-27 | v1.101 | 共用 CfFieldHelp：表单旁帮助图标 | orjrs |
| 2026-11-27 | v1.100 | 共用 CfEmptyState / CfLoadingState：列表空态与加载 | orjrs |
| 2026-11-27 | v1.99 | 共用 CfConfirmModal：列表删除确认 | orjrs |
| 2026-11-27 | v1.98 | 通用版默认隐藏密级相关 UI | orjrs |
| 2026-11-27 | v1.97 | 监控页密级审计 Tab | orjrs |
| 2026-11-27 | v1.96 | 调试流式：画布节点 hint + 侧栏 Live stream | orjrs |
| 2026-11-24 | v1.95 | 工作流「整理布局」：按连线自动左→右排版 | orjrs |
| 2026-11-24 | v1.94 | 工作流连线加终点箭头，标明执行方向 | orjrs |
| 2026-11-24 | v1.93 | 小队页一键生成流水线；画布 Agent 可绑定智能体 | orjrs |
| 2026-11-23 | v1.92 | 工作流画布增加小队（team）节点配置 | orjrs |
| 2026-11-23 | v1.91 | 小队预览：续聊 conversation_id、拉历史、工人失败提示 | orjrs |
| 2026-11-23 | v1.90 | 顶栏「小队」+/teams：列表新建预览（多 Agent） | orjrs |
| 2026-11-23 | v1.89 | 构建做不到的能力固定拒答；清单外工具提示去 MCP | orjrs |
| 2026-11-23 | v1.88 | 工作室构建：对话写入提示词+工具，保存固化 | orjrs |
| 2026-11-20 | v1.87 | 共用 CfChatPanel；跨模块重复代码 backlog | orjrs |
| 2026-11-20 | v1.86 | 智能体工作室未做能力列入后续 TODO | orjrs |
| 2026-11-20 | v1.85 | 智能体工作室左配置右构建/预览；列表卡片+公共视图切换 | orjrs |
| 2026-11-20 | v1.84 | 新建智能体改为独立页 `/agents/new`，不再弹窗 | orjrs |
| 2026-11-19 | v1.83 | 模型页可勾 Rerank；知识库开启 Rerank 后选已配置模型 | orjrs |
| 2026-11-19 | v1.82 | 三种检索卡内 TopK/最低分；混合加语义↔关键词权重滑条 | orjrs |
| 2026-11-16 | v1.81 | 模型页可配 Jina/Firecrawl；网页入库可选抓取引擎 | orjrs |
| 2026-11-16 | v1.80 | 上传三步：数据源（本地/网页/文本）→ 分段 → 确认表 | orjrs |
| 2026-11-16 | v1.79 | 上传弹窗去掉重置；操作进页脚分层 | orjrs |
| 2026-11-16 | v1.77 | 检索测试：方式芯片 + 右侧栏切换 Top-K | orjrs |
| 2026-11-16 | v1.76 | 「设置」第三标签；检索方式 + 分模式 Rerank | orjrs |
| 2026-11-16 | v1.75 | 知识库文档改独立页；添加文件用大弹窗步骤 | orjrs |
| 2026-11-16 | v1.74 | 「设置默认」挪到文档/检索标签栏右侧 | orjrs |
| 2026-11-16 | v1.73 | 上传改成可点步骤条（完成打勾 / 连线 / 可回退） | orjrs |
| 2026-11-16 | v1.72 | 上传第二步「返回」靠左，重置和上传靠右 | orjrs |
| 2026-11-16 | v1.71 | 预览分块：超限文件在步骤内说明，不再空白 | orjrs |
| 2026-11-16 | v1.70 | 文档抽屉上传区常驻；标题栏收矮并去掉文档数 | orjrs |
| 2026-11-16 | v1.69 | 文档列表分页；上传改抽屉内分步；默认设置改弹窗 | orjrs |
| 2026-11-16 | v1.68 | 知识库列表上方关键字搜索 | orjrs |
| 2026-11-16 | v1.67 | 知识库首页去掉目录分组，只留库名 | orjrs |
| 2026-11-16 | v1.66 | 向量模型创建后锁定；目录「新建」与选项分层 | orjrs |
| 2026-11-15 | v1.65 | 知识库文档预览工具栏加最大化（同 CfModal 图标） | orjrs |
| 2026-11-15 | v1.64 | 知识库检索测试：多行问句框 + 文字按钮 | orjrs |
| 2026-11-15 | v1.63 | 知识库分段/检索字段加帮助图标悬浮说明 | orjrs |
| 2026-11-15 | v1.62 | 模型页：Scope 徽章、公开开关、只读与个人/全站默认 | orjrs |
| 2026-11-15 | v1.61 | 弹窗内 Select portal=false，修分段下拉点不开 | orjrs |
| 2026-11-15 | v1.60 | 知识库弹窗「取消」去掉 X 图标 | orjrs |
| 2026-11-15 | v1.58 | 知识库卡片操作图标悬停显示；删除须弹窗确认 | orjrs |
| 2026-11-15 | v1.57 | 知识库卡片：⋯ 菜单改为文档/编辑/删除图标按钮 | orjrs |
| 2026-11-15 | v1.56 | CfModal 内 Select 下拉抬到 z-300，避免被弹窗挡住 | orjrs |
| 2026-11-15 | v1.55 | 知识库文档预览区：预览 / Markdown 编辑切换并保存 | orjrs |
| 2026-11-15 | v1.54 | 知识库抽屉「设置」：默认分段/检索 + 权限占位 | orjrs |
| 2026-11-15 | v1.53 | 检索测试可调 Top-K / 最低相似度 | orjrs |
| 2026-11-15 | v1.52 | 知识库上传两步：分段设置 + 分块预览 | orjrs |
| 2026-11-15 | v1.51 | CfModal：叠在 USlideover 上可点击（pointer-events + z-index） | orjrs |
| 2026-11-15 | v1.50 | 知识库：目录下拉 +「+」新建；列表按目录分段 | orjrs |
| 2026-11-15 | v1.49 | 知识库抽屉：点文档名预览正文（列表+右侧阅读区） | orjrs |
| 2026-11-11 | v1.48 | 前端 SSE：共用读取，缓冲/取消/连接重试/断线提示 | orjrs |
| 2026-11-09 | v1.47 | 评估页：聊天模型下拉，免手抄模型 ID | orjrs |
| 2026-11-08 | v1.46 | 模型页：厂商决定卡片小字，新增不再默认 OpenAI | orjrs |
| 2026-11-08 | v1.45 | 模型页：类型改为可选快捷模板，模型名可手填 | orjrs |
| 2026-11-08 | v1.44 | 模型页增加智谱 BigModel；免费 Flash 优先 | orjrs |
| 2026-11-07 | v1.43 | Playground：绑定 Prompt 的 Agent 展示变量表单（17.14） | orjrs |
| 2026-10-24 | v1.42 | 提示词试跑；评估变量表单；顶栏展示邮箱 | orjrs |
| 2026-10-18 | v1.41 | 共用 CfModal：拖动 / 缩放 / 最大化，替换全站 UModal | orjrs |
| 2026-10-18 | v1.40 | MCP 页：卡片网格 + 筛选/开关/工具数（对标 Cursor） | orjrs |
| 2026-10-18 | v1.39 | MCP 页：点行看工具明细；创建表单强化飞书路径说明 | orjrs |
| 2026-10-15 | v1.38 | 配置下增加「提示词」/prompts；评估页支持 prompt 目标 | orjrs |
| 2026-08-25 | v1.37 | 模型页用途：对话 / 向量，两套默认 | orjrs |
| 2026-08-22 | v1.36 | 顶栏改为：控制台 对话 智能体 工作流 知识库 配置；用量/监控进头像 | orjrs |
| 2026-08-22 | v1.35 | 通用文件附件排到阶段十三（等 RAG 稳） | orjrs |
| 2026-08-22 | v1.34 | 发送框支持从访达复制文件粘贴（图片） | orjrs |
| 2026-08-22 | v1.33 | 排队编辑改到发送框；行上「编辑中 ×」取消 | orjrs |
| 2026-08-21 | v1.32 | Playground 排队（可编辑删除）+ 插入硬截断 | orjrs |
| 2026-08-21 | v1.31 | Playground 历史对话支持按标题搜索 | orjrs |
| 2026-08-21 | v1.30 | Playground 支持上传/粘贴图片发图（多模态 vision） | orjrs |
| 2026-08-21 | v1.29 | Playground 聊天支持 Markdown/链接图片展示 | orjrs |
| 2026-08-21 | v1.28 | Playground 历史对话置顶（pin / unpin） | orjrs |
| 2026-08-20 | v1.27 | Resend 邮件忘记密码：发信表单 + 重置密码页 | orjrs |
| 2026-08-20 | v1.26 | 登录「忘记密码」+ 管理员重置临时密码（无邮件版） | orjrs |
| 2026-08-19 | v1.25 | 全屏径向渐变背景 + 内容居中；卡片半透明融入，不再像贴上去的白块 | orjrs |
| 2026-08-19 | v1.24 | ~~控制台铺满窗口~~（同日改为渐变氛围 + 内容居中） | orjrs |
| 2026-08-19 | v1.23 | Usage 卡片 0 用量时进度条静止，不再当成加载动画 | orjrs |
| 2026-08-19 | v1.22 | 按钮语义：能点用主题实心色，不能点才灰色；补 Playwright e2e | orjrs |
| 2026-08-18 | v1.21 | 配额：导航加 Usage；Playground 额度条；用量图表页 | orjrs |
| 2026-08-16 | v1.20 | 未配置默认模型时对话 Toast 提示去「模型」页，不再显示 mock | orjrs |
| 2026-08-16 | v1.19 | 对话/模型页 DeepSeek 增加 V4，同时保留 chat / reasoner | orjrs |
| 2026-08-16 | v1.18 | Playground 左侧历史可折叠（桌面顶栏图标；状态记在浏览器） | orjrs |
| 2026-08-16 | v1.17 | Playground：左侧历史对话；参数改为右上角滑层（桌面也进滑层） | orjrs |
| 2026-08-15 | v1.16 | Agents/Workflows 模型下拉改为 GET /api/v1/models（数据库配置） | orjrs |
| 2026-08-15 | v1.15 | P7：共享 editorial chrome（cf-page 等）；各模块标题/面板/表格对齐 Dashboard | orjrs |
| 2026-08-15 | v1.14 | Console→Dashboard；导航英文短标签；登录/Playground 向示意稿构图靠拢（主题 token 不变） | orjrs |
| 2026-08-13 | v1.13 | P6 手机兼容：导航账户区、Playground 参数 USlideover、工作流「请用电脑」提示 | orjrs |
| 2026-08-13 | v1.12 | P4/P5 完成：Workflows 壳层 Nuxt UI；卸载 naive-ui / element-plus / @vicons；app 仅 UApp | orjrs |
| 2026-08-12 | v1.11 | Playground 落地：UChatMessages/Message/Prompt/Submit/Shimmer；本地消息 → parts 适配层；侧栏 Nuxt UI；去尽 naive-ui | orjrs |
| 2026-08-12 | v1.10 | 框架改为 Nuxt 4.5 + Nuxt UI 4（解锁 AI Chat）；P0 壳与四主题已落地 | orjrs |
| 2026-08-12 | v1.9 | 确认四套主题全部采纳（Aurora 默认；Ink Night / Citrus / Glass 可切换） | orjrs |
| 2026-08-12 | v1.8 | 增加 4 套可切换视觉主题（Aurora / Ink Night / Citrus / Glass）；默认更绚丽 | orjrs |
| 2026-08-12 | v1.7 | 明确导航 IA 冻结：模块/路由/角色可见性与现网一致，仅换 Nuxt UI 壳 | orjrs |
| 2026-08-12 | v1.6 | 壳层/导航改以 Nuxt UI 为主（好看且好写）；shadcn 收缩为可深度改皮的补充件 | orjrs |
| 2026-08-12 | v1.5 | Chat 明确左右分侧（助手左 / 用户右）；正式采用 Nuxt UI AI Chat 组件套件 | orjrs |
| 2026-08-12 | v1.4 | Playground Chat 交互改为 Claude×Cursor 阅读流；Nuxt UI Chat 作实现底座；弃用 API 调试台式示意 | orjrs |
| 2026-08-12 | v1.3 | 增加 Nuxt UI vs shadcn-vue 组件选型：混用策略与冲突规则 | orjrs |
| 2026-08-12 | v1.2 | 补充关键屏 UI 示意稿（登录 / 控制台 / Playground / 手机） | orjrs |
| 2026-08-12 | v1.1 | 补充响应式策略：Web 优先验收，手机端预留兼容但不阻塞主路径 | orjrs |
| 2026-08-12 | v1.0 | 全新 UI 方向：Vue3 + Nuxt3 + Tailwind4 + shadcn-vue；保留 Vue Flow；接口与 CI/CD 不变 | orjrs |

> 注：任务状态变更直接在下方任务表格中更新即可，无需额外记录。

## [变更] 模型页 Jina / Firecrawl + 网页引擎（2026-11-16）

- **变更原因**：网页入库要用外部 Reader，Key 进模型页
- **变更前 vs 变更后**：~~仅对话/向量用途~~ → 用途增加「网页抓取」；厂商增加 Jina Reader / Firecrawl；知识库网页源可选 auto/jina/firecrawl/basic
- **包含代码**：`cogniforge-web/pages/models.vue`、`useProviders.ts`、`pages/knowledge/[id].vue`、i18n
- **后续 TODO（约一周后，目标 2026-11-23）**：Crawl4AI 独立服务 UI/部署；本期不做

## [变更] 上传三步数据源向导（2026-11-16）

- **变更原因**：入库不只本地文件；需要网页与自定义文本，并在上传前核对列表
- **变更前 vs 变更后**：~~两步：选择文件 → 本次分段~~ → 三步：选择数据源（本地多文件可删 / 网页 URL / 自定义 md|txt）→ 本次分段预览 → 确认上传表（名称 / 状态 / 删除）
- **包含代码**：`cogniforge-web/pages/knowledge/[id].vue`、`composables/useKnowledgeBases.ts`、i18n；后端 `documents/text`、`documents/url`

## [变更] 上传弹窗页脚分层（2026-11-16）

- **变更原因**：按钮和内容挤在一起；重置多余
- **变更前 vs 变更后**：~~内容区底部一排返回/重置/上传~~ → 去掉重置；返回与下一步/上传进 CfModal 页脚，顶部分隔线
- **包含代码**：`cogniforge-web/pages/knowledge/[id].vue`

## [变更] 通用版隐藏密级 UI（2026-11-27）

- **原因**：密级是金融扩展，通用版不展示
- **落地**：`NUXT_PUBLIC_SHOW_SECURITY_LEVEL`；`useSecurityFeatures`
- **含代码**：跨仓 knowledge / users / roles / monitor

## [变更] 监控密级审计 Tab（2026-11-27）

- **原因**：金融扩展审计要可看
- **落地**：监控页第二 Tab 拉 `/admin/security-audits`
- **含代码**：跨仓 `pages/monitor.vue`、`useMonitor`

## [变更] 调试流式进度 UI（2026-11-27）

- **原因**：LLM 流式开启后用户要在编辑器看见「正在生成」
- **落地**：节点 `hint`/`streamProgress`；调试侧栏「流式进度」块；时间线 note 跟 `detail.progress`
- **含代码**：跨仓 `WorkflowFlowNode`、`WorkflowCanvas`、`workflows/[id].vue`

## [变更] 工作流整理布局（2026-11-24）

- **变更原因**：节点拖乱、横向拉太开时难看；需要一键按执行方向排整齐
- **变更后**：顶栏「整理布局」按连线分层左→右排版，同列上下对齐，并 fitView
- **包含代码**：`utils/workflowLayout.ts`、`pages/workflows/[id].vue`、`WorkflowCanvas.vue`

## [变更] 工作流连线箭头（2026-11-24）

- **变更原因**：无箭头时弯线看不清谁先谁后；n8n 等产品均标方向
- **变更后**：边默认 `markerEnd: ArrowClosed`；加载/新建连线统一带箭头
- **包含代码**：`WorkflowCanvas.vue`、`pages/workflows/[id].vue`

## [变更] 小队一键流水线入口（2026-11-24）

- **变更原因**：Pipeline 产品化：从小队生成可编辑线性图
- **变更后**：小队卡「生成/打开流水线」；画布 Agent 节点可选绑定智能体与提示词
- **包含代码**：`pages/teams/index.vue`、`workflows/[id].vue`、i18n

## [变更] 画布小队节点（2026-11-23）

- **变更原因**：可选精细控制：工作流里绑定一支小队跑任务
- **变更后**：节点类型 `team`；抽屉选小队 + 任务说明；执行走引擎 `TeamNodeExecutor`
- **包含代码**：`cogniforge-web/pages/workflows/[id].vue`、`useTeams`、i18n

## [变更] 小队预览续聊与失败提示（2026-11-23）

- **变更原因**：对齐 Go §12：预览落 chat、软失败可感知
- **变更后**：打开预览拉该小队最近会话；chat 带 `conversation_id`；工人失败 toast + 行内提示；可「新对话」
- **包含代码**：`pages/teams/index.vue`、`useTeams`、`useConversations`、i18n

## [变更] 智能体小队页（2026-11-23）

- **变更原因**：多 Agent 后端已有 `/teams`，需要产品入口
- **变更后**：顶栏「小队」→ `/teams`；卡片列表、新建（模式/主管/工人）、预览对话展示 `active_agent_id`
- **包含代码**：`cogniforge-web/pages/teams/index.vue`、`composables/useTeams.ts`、`constants/nav.ts`

## [变更] 构建做不到时明确拒答（2026-11-23）

- **变更原因**：避免模型嘴硬；清单外工具静默丢弃无感
- **变更后**：虚拟机/语音/发布/装新工具本地固定回复；清单外工具 Toast + 文内说明引导 MCP 页
- **包含代码**：`cogniforge-web/utils/agentBuildConfig.ts`、`AgentStudio.vue`

## [变更] 工作室构建固化配置（2026-11-23）

- **变更原因**：构建需对齐「目标 + 工具 + 示例 → 固化到 Agent」，不能只写提示词
- **变更后**：构建助手输出 `agent-config`；应用到左侧或保存时自动写入；工具仅限已接入 MCP
- **包含代码**：`cogniforge-web/utils/agentBuildConfig.ts`、`AgentStudio.vue`

## [变更] 共用删除确认 CfConfirmModal（2026-11-27）

- **变更原因**：各列表页手写同一套 CfModal 删除确认，文案/按钮重复
- **变更后**：`CfConfirmModal`（窄窗、关缩放/最大化、取消+危险确认）；已接 agents / teams / models / keys / workflows / knowledge / prompts / mcp / users / roles
- **未迁**：知识库文档删除、用户重置密码等非「列表删资源」确认仍可保留原 CfModal
- **包含代码**：`cogniforge-web/components/CfConfirmModal.vue`；上述列表页

## [变更] 共用空态/加载 CfEmptyState · CfLoadingState（2026-11-27）

- **变更原因**：各列表手写 `cf-state` 转圈与空图标文案
- **变更后**：`CfLoadingState`、`CfEmptyState`（可选 icon / message / #action）；已接 agents / teams / models / mcp / prompts / kb / workflows / keys / users / roles / monitor / evals
- **未迁**：详情页紧凑态、Playground 对话空态、用量图等特殊布局仍可自写
- **包含代码**：`cogniforge-web/components/CfEmptyState.vue`、`CfLoadingState.vue`

## [变更] 共用表单帮助 CfFieldHelp（2026-11-27）

- **变更原因**：知识库检索/分段旁反复手写 UTooltip + help 按钮
- **变更后**：`CfFieldHelp`（`text` + 可选 `icon`）；已接知识库详情检索查询与分段参数标签
- **包含代码**：`cogniforge-web/components/CfFieldHelp.vue`、`pages/knowledge/[id].vue`

## [变更] 工作室对话日志（2026-11-27）

- **变更原因**：工作室 backlog；查看该 Agent 在对话页的历史
- **变更后**：顶栏「对话日志」侧栏列表（`GET /conversations?agent_id=`）；点条目打开 `/playground?agent=&c=`
- **包含代码**：`AgentStudio.vue`、`useConversations.ts`；主仓 API v1.98

## [变更] 工作室预览附件上传（2026-11-27）

- **变更原因**：工作室 backlog；对齐 Playground attachments
- **变更后**：左侧「允许上传附件」；预览区回形针上传文档并随 `streamAgentChat` 传 `attachments`；Playground 按开关显隐文件钮
- **包含代码**：`AgentStudio.vue`、`playground.vue`；主仓 API v1.97 / DB v1.42

## [变更] 工作室绑定知识库（2026-11-27）

- **变更原因**：工作室 backlog；对话需自动检索所选知识库
- **变更后**：左侧多选 `rag_kb_ids`（≤8）；保存到 Agent；预览/Playground 选该 Agent 对话时后端自动 RAG
- **包含代码**：`AgentStudio.vue`、`useAgents.ts`；主仓 API v1.96 / DB v1.41

## [变更] 工作室开场白与预设问题（2026-11-27）

- **变更原因**：竞品常见首屏欢迎语 + 一点即发预设问
- **变更后**：左侧可配 `opening_message` / `suggested_prompts`（≤6）；预览空态与 Playground 空态展示；不发给模型
- **包含代码**：`AgentStudio.vue`、`playground.vue`、`useAgents.ts`；主仓 API v1.95 / DB v1.40

## [变更] Playground 改走 useAgentChat / useChatStream（2026-11-27）

- **变更原因**：对话页仍内联 SSE，与工作室共用 composable 重复
- **变更后**：有 Agent → `useAgentChat.streamChat`；无 Agent → `useChatStream`（含 web_search mcp）；去掉页面内 fetchSse
- **包含代码**：`playground.vue`、`useAgentChat.ts`、`useChatStream.ts`

## [变更] 共用聊天面板 CfChatPanel（2026-11-20）

- **变更原因**：工作室构建/预览与对话页聊天 UI、SSE 发送重复
- **变更后**：`CfChatPanel` 共用消息区+输入框；`useChatStream` 管 `/chat/stream`；`useAgentChat` 管 Agent 对话；错误码 `mapChatHttpError`
- **后续去重**：~~`CfConfirmModal`~~ / ~~`CfEmptyState`~~ / ~~`CfFieldHelp`~~ / ~~Playground 内联 SSE~~ / ~~`CfResourceCard`~~（2026-11-27）
- **包含代码**：`cogniforge-web/components/CfChatPanel.vue`、`useChatStream.ts`、`AgentStudio.vue`、`playground.vue`、`prompts.vue`

## [变更] CfResourceCard 共用卡片壳（2026-11-27）

- **变更原因**：agents/knowledge 卡片头图/操作/描述重复
- **变更后**：`CfResourceCard` 提供图标、标题、徽章、操作、描述、meta 槽；两页列表改用
- **包含代码**：`CfResourceCard.vue`、`pages/agents/index.vue`、`pages/knowledge/index.vue`

## [变更] 知识库列表视图切换（2026-11-27）

- **变更原因**：对齐智能体列表；去重 backlog 中的 ViewModeToggle
- **变更后**：知识库顶栏可切卡片/表格；偏好记 `cf-kb-view`
- **包含代码**：`pages/knowledge/index.vue`、`CfViewModeToggle`

## [变更] 智能体工作室后续 TODO（2026-11-20）

- **变更原因**：本期只做工作室骨架与列表卡片；截图其余能力需明确 backlog
- **后续 TODO**：
  - 优先：绑定知识库、开场白/预设问、预览附件上传、~~`/` `@` 提示词增强~~、~~「了解更多」文档链~~、对话日志、~~配置版本历史~~
  - 中期：~~发布渠道~~（Web/API/嵌入已做；IM 仍待）、~~猜你想问~~
  - 远期：~~虚拟机~~（UI 占位；运行时仍待）、~~语音播放~~（浏览器 TTS）、~~语音输入~~（浏览器 STT；云 ASR 另排）
- **详见**：`cogniforge-web/docs/99-dev-plan/01-development-plan.md` v1.67；主仓开发计划「后续待办」

## [变更] 提示词标签筛选空串 500（2026-11-27）

- **变更原因**：USelect 禁止 `SelectItem` value 为空串，「全部」用 `''` 会整页 500
- **变更后**：哨兵 `__all__`；列表加载对非数组 data 容错
- **包含代码**：`pages/prompts.vue`

## [变更] Playground 插入发送确认框（2026-11-27）

- **变更原因**：中间插入会删后续消息时仍用浏览器 `confirm`
- **变更后**：改用 `CfConfirmModal`
- **包含代码**：`pages/playground.vue`

## [变更] 知识库弃编辑确认框（2026-11-27）

- **变更原因**：未保存文档编辑离开时仍用浏览器 `confirm`
- **变更后**：切 Tab / 换文档 / 关闭预览统一改用 `CfConfirmModal`
- **包含代码**：`pages/knowledge/[id].vue`

## [变更] 工作室版本回滚确认框（2026-11-27）

- **变更原因**：配置版本回滚仍用浏览器 `confirm`
- **变更后**：改用 `CfConfirmModal`
- **包含代码**：`components/AgentStudio.vue`

## [变更] Playground 删对话确认框（2026-11-27）

- **变更原因**：删对话仍用浏览器 `confirm`
- **变更后**：改用 `CfConfirmModal`
- **包含代码**：`pages/playground.vue`

## [变更] 评估页删除确认框（2026-11-27）

- **变更原因**：套件删除仍用浏览器 `confirm`；用例删除无确认
- **变更后**：套件与用例删除均走 `CfConfirmModal`
- **包含代码**：`pages/evals.vue`、i18n

## [变更] Cron 校验与确认框（2026-11-27）

- **变更原因**：5 段 Cron 要到后端才报错；删除调度/回滚仍用浏览器 confirm
- **变更后**：前端先校验 6 段；调度删除与版本回滚改用 `CfConfirmModal`
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] Webhook 校验与 HTTP 超时副文案（2026-11-27）

- **变更原因**：回调可填脏地址；HTTP 节点看不出非默认超时
- **变更后**：保存基础信息时校验 http(s)；HTTP 副文案在超时≠30s 时显示 `· Ns`
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] Playwright 默认不复用端口（2026-11-27）

- **变更原因**：本地占用 3010 的其它服务被误复用，e2e 访问 `/prompts` 得 404
- **变更后**：默认自起 Nuxt；探活 `/login`；`PW_REUSE=1` 才复用
- **包含代码**：`cogniforge-web/playwright.config.ts`

## [变更] 画布副文案补充（2026-11-27）

- **变更原因**：LLM/检索/等待已绑定后仍看不出流式、条数、超时
- **变更后**：LLM 显示「模型 · 流式」；检索「库名 · mode · top N」；等待附带 `Ns` 超时
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 循环节点 for_each（2026-11-27）

- **变更原因**：引擎已支持遍历数组，画布只有 times/while
- **变更后**：循环方式可选「遍历数组」；填数组/元素/下标变量；未填数组标未绑定；画布副文案 `数组 → 元素`
- **包含代码**：`pages/workflows/[id].vue`、`WorkflowCanvas.vue`、`workflowBind.ts`、i18n；Go `vueDataToConfig`

## [变更] 提示词版本归档 / 回滚发版（2026-11-27）

- **变更原因**：17.17 运维需废弃旧版与按正文回滚
- **变更后**：版本卡可归档（非当前 published）；回滚发版按旧正文发新 published；徽章区分 archived
- **包含代码**：`pages/prompts.vue`、`usePrompts.ts`、i18n

## [变更] 开始节点画布副文案（2026-11-27）

- **变更原因**：开始节点在画布上看不出入参变量
- **变更后**：无入参显示「无入参」；≤2 个显示变量名；更多显示「a, b 等 N 个」
- **包含代码**：`WorkflowCanvas.vue`、`pages/workflows/[id].vue`、i18n

## [变更] 循环节点画布副文案（2026-11-27）

- **变更原因**：循环节点在画布上看不出次数或条件
- **变更后**：固定次数显示「固定次数 ×N」；条件循环显示条件摘要或「≤最大次数」；新建默认 times×3
- **包含代码**：`WorkflowCanvas.vue`、`pages/workflows/[id].vue`

## [变更] 延时/等待节点画布副文案（2026-11-27）

- **变更原因**：延时与等待在画布上看不出配置摘要
- **变更后**：延时显示「N 秒/分/时」；等待显示原因（及消息摘要）；新建带默认值
- **包含代码**：`WorkflowCanvas.vue`、`pages/workflows/[id].vue`

## [变更] 节点必填绑定单测（2026-11-27）

- **变更原因**：近期补了多种未绑定拦截，缺回归保护
- **变更后**：抽出 `workflowBind.ts`；`isWorkflowBindUnbound` vitest 覆盖 llm/http/code/condition/team/agent/search/subflow
- **包含代码**：`utils/workflowBind.ts`、`utils/workflowBind.test.ts`、`workflowNodeMeta.ts` 再导出

## [变更] LLM 节点模型空态提示（2026-11-27）

- **变更原因**：无模型时下拉空且无引导；点开不刷新列表
- **变更后**：空态提示 + `/models`；点开 LLM 节点刷新模型列表
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 条件节点未填字段醒目标记（2026-11-27）

- **变更原因**：条件节点无判断字段仍可保存/跑
- **变更后**：`isWorkflowBindUnbound('condition')`；虚线框 + 顶栏计数；填好后显示 `字段 运算符 值`
- **包含代码**：`workflowNodeMeta.ts`、`WorkflowCanvas.vue`、`pages/workflows/[id].vue`、i18n

## [变更] 代码节点未填代码醒目标记（2026-11-27）

- **变更原因**：代码节点空源码仍可保存/跑
- **变更后**：`isWorkflowBindUnbound('code')`；虚线框 + 顶栏计数；填好后显示 `语言: 首行摘要`
- **包含代码**：`workflowNodeMeta.ts`、`WorkflowCanvas.vue`、`pages/workflows/[id].vue`、i18n

## [变更] HTTP 节点未填 URL 醒目标记（2026-11-27）

- **变更原因**：HTTP 节点空地址仍可保存/跑
- **变更后**：`isWorkflowBindUnbound('http')`；虚线框 + 顶栏计数；填好后显示 `METHOD url` 副文案
- **包含代码**：`workflowNodeMeta.ts`、`WorkflowCanvas.vue`、`pages/workflows/[id].vue`、i18n

## [变更] LLM 节点未选模型醒目标记（2026-11-27）

- **变更原因**：LLM 节点无模型也可保存/跑，与其它绑定节点不一致
- **变更后**：`isWorkflowBindUnbound('llm')`；虚线框 + 顶栏计数；选中后显示模型名副文案
- **包含代码**：`workflowNodeMeta.ts`、`WorkflowCanvas.vue`、`pages/workflows/[id].vue`、i18n

## [变更] 提示词版本行级 diff（2026-11-27）

- **变更原因**：P2 17.15；并排原文难定位改动
- **变更后**：对比弹窗默认行级差异（可切并排）；`lineDiff` LCS
- **包含代码**：`utils/lineDiff.ts`、`pages/prompts.vue`、i18n

## [变更] 提示词 Tags 编辑与筛选（2026-11-27）

- **变更原因**：P2 17.16；后端已有 `tags`，前端未做
- **变更后**：新建/详情编辑标签（逗号分隔）；列表按标签筛选；卡片徽章展示
- **包含代码**：`pages/prompts.vue`、i18n

## [变更] 角色管理列表卡片视图（2026-11-27）

- **变更原因**：跨模块列表去重；管理端角色页仍只有表格
- **变更后**：默认卡片 + `CfViewModeToggle`（`cf-roles-view`）；系统/默认徽章；表格保留；分页共用
- **包含代码**：`pages/admin/roles.vue`

## [变更] 用户管理列表卡片视图（2026-11-27）

- **变更原因**：跨模块列表去重；管理端用户页仍只有表格
- **变更后**：默认卡片 + `CfViewModeToggle`（`cf-users-view`）；角色/状态徽章；表格保留；搜索/分页共用
- **包含代码**：`pages/admin/users.vue`

## [变更] 评估列表卡片视图（2026-11-27）

- **变更原因**：跨模块列表去重；评估套件页仍只有表格
- **变更后**：默认卡片 + `CfViewModeToggle`（`cf-evals-view`）；表格视图保留
- **包含代码**：`pages/evals.vue`

## [变更] 画布已绑定节点显示资源名（2026-11-27）

- **变更原因**：绑好后画布仍只显示节点类型名，看不出绑的是谁
- **变更后**：`bindCaption` 副文案（小队/智能体/知识库/子流程名）；保存与选项刷新时同步；流式进度优先
- **包含代码**：`WorkflowCanvas.vue`、`pages/workflows/[id].vue`

## [变更] 画布顶栏未绑定数量芯片（2026-11-27）

- **变更原因**：节点多时难扫全图；需要顶栏汇总
- **变更后**：有未绑定节点时顶栏显示「N 个未绑定」芯片；点击打开第一个未配置节点；与校验共用 `focusFirstUnboundNode`
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 画布未绑定节点醒目标记（2026-11-27）

- **变更原因**：保存/执行拦截前，画布上看不出哪些节点还没绑资源
- **变更后**：agent/team/search/subflow 未绑定时虚线琥珀边框 + `!` 角标 + 副文案；`isWorkflowBindUnbound` 与校验共用；补齐 `#node-subflow` 模板
- **包含代码**：`WorkflowFlowNode.vue`、`WorkflowCanvas.vue`、`workflowNodeMeta.ts`、`pages/workflows/[id].vue`、i18n

## [变更] 执行/调试校验节点绑定（2026-11-27）

- **变更原因**：仅保存拦截不够，正式运行/调试仍可能带着空绑定出发
- **变更后**：`ensureBindConfigured` 共用；保存 / 执行 / 调试均拦截
- **包含代码**：`pages/workflows/[id].vue`

## [变更] 保存校验节点绑定（2026-11-27）

- **变更原因**：未选团队/智能体/知识库/子流程仍可保存，运行才报错
- **变更后**：保存前拦截并打开第一个未配置节点；子流程 `workflow_id` 写入 definition.config
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 画布检索节点知识库空态（2026-11-27）

- **变更原因**：无知识库时下拉空且无引导
- **变更后**：空态提示 + `/knowledge`；点开节点刷新列表
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 画布子流程节点空态（2026-11-27）

- **变更原因**：无可嵌套工作流时下拉空且无说明；当前图会被排除
- **变更后**：空态提示 + `/workflows`；点开节点刷新列表
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 画布 Agent 节点空态（2026-11-27）

- **变更原因**：与多智能体节点一致；无 Agent 时下拉空且无引导
- **变更后**：空态提示 + `/agents/new`；点开节点刷新列表
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] MCP 列表 CfResourceCard（2026-11-27）

- **变更原因**：MCP 页自绘卡片；对齐其它资源列表共用壳
- **变更后**：卡片视图改用 `CfResourceCard`；开关/工具数/操作保留
- **包含代码**：`pages/mcp.vue`

## [变更] 画布多智能体节点文案与空态（2026-11-27）

- **变更原因**：绑定下拉常为空且无说明；「小队节点」叫法偏内部
- **变更后**：节点名「多智能体」；字段「选择团队」；无团队时提示并链到 `/teams`；点开节点时刷新列表
- **包含代码**：`pages/workflows/[id].vue`、i18n

## [变更] 小队委派深度/步数限额（2026-11-27）

- **变更原因**：`config.max_depth` / `max_steps` 仅后端默认，产品不可改
- **变更后**：新建可填；列表「限额」弹窗改深度/步数/失败重试；卡片 meta 展示
- **包含代码**：`pages/teams/index.vue`、i18n

## [变更] 小队工人失败重试开关（2026-11-27）

- **变更原因**：后端已有 `config.worker_retry_once`，前端无入口
- **变更后**：新建勾选；列表/卡片一键开关；徽章「失败重试」
- **包含代码**：`pages/teams/index.vue`、i18n

## [变更] API Key 列表卡片壳（2026-11-27）

- **变更原因**：Keys 页仅表格；对齐其它资源列表
- **变更后**：默认卡片 + `CfViewModeToggle`（`cf-keys-view`）；密钥展示/复制/吊销；表格保留
- **包含代码**：`pages/keys.vue`

## [变更] 提示词列表卡片壳（2026-11-27）

- **变更原因**：提示词页仅表格；对齐其它资源列表
- **变更后**：默认卡片 + `CfViewModeToggle`（`cf-prompts-view`）；表格保留
- **包含代码**：`pages/prompts.vue`

## [变更] 工作流列表卡片壳（2026-11-27）

- **变更原因**：工作流页仅表格；对齐 agents/knowledge/teams
- **变更后**：默认卡片 + `CfViewModeToggle`（`cf-workflows-view`）；表格保留
- **包含代码**：`pages/workflows/index.vue`

## [变更] 小队列表卡片壳（2026-11-27）

- **变更原因**：小队页自绘卡片；对齐 agents/knowledge 共用壳
- **变更后**：`CfResourceCard` + `CfViewModeToggle`（`cf-teams-view`）；支持表格视图
- **包含代码**：`pages/teams/index.vue`

## [变更] 工作室虚拟机/IM 占位（2026-11-27）

- **变更原因**：远期能力未落地，避免用户以为已开通
- **变更后**：发布侧栏展示 IM/虚拟机「即将推出」；左侧虚拟机开关禁用；帮助页补充说明
- **包含代码**：`AgentStudio.vue`、`pages/help/agent-studio.vue`、i18n

## [变更] 浏览器语音听写（2026-11-27）

- **变更原因**：远期语音输入；先用不依赖云 ASR 的浏览器 SpeechRecognition
- **变更后**：`CfChatPanel` 输入栏麦克风听写写入 composer；需麦克风权限
- **包含代码**：`utils/speechStt.ts`、`CfChatPanel.vue`

## [变更] MCP 列表视图切换（2026-11-27）

- **变更原因**：对齐 agents/knowledge；去重 backlog
- **变更后**：MCP 顶栏可切卡片/表格；偏好 `cf-mcp-view`
- **包含代码**：`pages/mcp.vue`、`CfViewModeToggle`

## [变更] 浏览器朗读助手回复（2026-11-27）

- **变更原因**：远期语音播放；先用不依赖云服务的浏览器 TTS
- **变更后**：`CfChatPanel` 助手消息旁朗读/停止；Playground 与工作室共用
- **包含代码**：`utils/speechTts.ts`、`CfChatPanel.vue`

## [变更] 猜你想问（2026-11-27）

- **变更原因**：无预设问题时空态太冷；与 `suggested_prompts` 合并设计
- **变更后**：无预设时按名称/描述启发式生成芯片；工作室可「填入猜你想问」写入后再保存
- **包含代码**：`utils/guessSuggestedPrompts.ts`、`AgentStudio.vue`、`playground.vue`

## [变更] 工作室发布渠道（2026-11-27）

- **变更原因**：中期 backlog；先给可复制接入
- **变更后**：顶栏「发布」侧栏：对话页链接（`?agent=`）、API curl、iframe 嵌入；第三方 IM 未做
- **包含代码**：`AgentStudio.vue`、i18n

## [变更] 工作室配置版本历史（2026-11-27）

- **变更原因**：竞品时钟回滚；对齐工作流版本弹窗
- **变更后**：顶栏「版本」侧栏列 `agent_versions`；可回滚；显示当前 `vN`
- **包含代码**：`AgentStudio.vue`、`useAgents.ts`；主仓 API v1.99 / DB v1.43

## [变更] 工作室提示词 / @ 插入（2026-11-27）

- **变更原因**：系统提示编辑要能快速插变量与引用已有能力
- **变更后**：未绑定提示词资产时可 `/` 插入 `{{var}}`；`@` 勾选技能/工具/知识库并写入 `[@类型: 名称]` 标注（后端不解析）
- **包含代码**：`PromptMentionMenu.vue`、`utils/promptInsert.ts`、`AgentStudio.vue`

## [变更] 工作室「了解更多」说明页（2026-11-27）

- **变更原因**：构建/预览帮助气泡「了解更多」无跳转目标
- **变更后**：`/help/agent-studio` 说明页；构建与预览 tip 均链到该页
- **包含代码**：`cogniforge-web/pages/help/agent-studio.vue`、`AgentStudio.vue`、i18n

## [变更] 智能体工作室与卡片列表（2026-11-20）

- **变更原因**：对齐常见 Agent 配置体验（左配置 / 右构建预览）；列表默认卡片并支持切换
- **变更前 vs 变更后**：
  - ~~弹窗编辑 / 简单新建页~~ → `/agents/new` 与 `/agents/:id` 共用 `AgentStudio`：左侧模型/提示词/技能/工具/状态，右侧「构建」「预览」+ 悬浮说明
  - ~~仅表格列表~~ → 默认卡片；公共组件 `CfViewModeToggle`（模型页复用）可切列表
- **能力取舍**：虚拟机/语音/发布渠道等未做，见 v1.86 TODO；构建对话辅助写提示词，预览需先保存后试聊
- **包含代码**：`components/AgentStudio.vue`、`CfViewModeToggle.vue`、`pages/agents/*`、`pages/models.vue`

## [变更] 新建智能体改独立页（2026-11-20）

- **变更原因**：新建表单字段多，弹窗空间不够，改为完整页面
- **变更前 vs 变更后**：~~列表页弹窗创建~~ → `/agents/new` 独立页；编辑进 `/agents/:id`（`AgentStudio`）；路由结构为 `pages/agents/index.vue` + `new.vue` + `[id].vue`
- **包含代码**：`cogniforge-web/pages/agents/index.vue`、`pages/agents/new.vue`、`pages/agents/[id].vue`

## [变更] Rerank 可选已配置模型（2026-11-19）

- **变更原因**：开启 Rerank 不能只开关，必须指定 rerank 类型模型
- **变更前 vs 变更后**：~~仅 Rerank 开关~~ → 模型页用途可勾「Rerank」并填模型名；知识库设置/检索测试在开启 Rerank 后下拉选择已启用且带 `rerank` 能力的配置
- **包含代码**：`cogniforge-web/pages/models.vue`、`pages/knowledge/[id].vue`、`composables/useProviders.ts`、`useKnowledgeBases.ts`、`i18n/messages.ts`

## [变更] 分模式 TopK / 阈值与混合权重 UI（2026-11-19）

- **变更原因**：混合检索要有权重；三种检索各自要有 Top-K 与 score 阈值
- **变更前 vs 变更后**：~~设置页下方单独一组 Top-K / 最低分；混合无权重~~ → 每种检索卡片内各自 Top-K / 最低分；混合卡片增加蓝/青双色权重滑条（语义 + 关键词 = 1）；检索测试右侧栏选中某方式后同样展示
- **包含代码**：`cogniforge-web/pages/knowledge/[id].vue`、`composables/useKnowledgeBases.ts`、`i18n/messages.ts`

## [变更] 检索测试切换检索方式（2026-11-16）

- **变更原因**：测试要能临时换检索方式，参考常见产品问句框上的方式芯片
- **变更前 vs 变更后**：~~Top-K / 最低分平铺在问句下~~ → 芯片显示当前方式；右侧栏选语义/关键字/混合、Rerank、Top-K 滑条；只影响本次测试
- **包含代码**：`cogniforge-web/pages/knowledge/[id].vue`

## [变更] 设置第三标签与检索方式（2026-11-16）

- **变更原因**：按钮和标签不一致；设置应管检索方式而不是「默认分段」弹窗
- **变更前 vs 变更后**：~~标签栏右侧「设置默认」按钮 + 默认分段弹窗~~ → 第三标签「设置」：语义 / 关键字 / 混合 + 各自 Rerank；Top-K / 最低分
- **包含代码**：`cogniforge-web/pages/knowledge/[id].vue`

## [变更] 知识库文档独立页（2026-11-16）

- **变更原因**：侧栏里叠上传和列表太挤
- **变更前 vs 变更后**：~~USlideover 管文档~~ → `/knowledge/:id` 独立页（返回 + 列表 + 添加文件）；上传两步进大弹窗；设置默认仍在标签栏
- **包含代码**：`cogniforge-web/pages/knowledge/index.vue`、`pages/knowledge/[id].vue`

## [变更] 「设置默认」进标签栏（2026-11-16）

- **变更原因**：库默认分段 / 检索 / 权限不属于上传步骤
- **变更前 vs 变更后**：~~按钮在上传步骤条右侧~~ → 与「文档管理 / 检索测试」同一行靠右；上传区只留步骤
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 上传可点步骤条（2026-11-16）

- **变更原因**：步骤文字不能点，也不像向导
- **变更前 vs 变更后**：~~两个静态标签~~ → 步骤条：当前高亮、完成打勾、中间连线；可点回退；无文件时第 2 步禁用
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`i18n/messages.ts`

## [变更] 上传「返回」靠左（2026-11-16）

- **变更原因**：返回和重置、上传挤在同一侧，不像退回上一步
- **变更前 vs 变更后**：~~三个按钮都靠右~~ → 「返回」在操作行最左，重置和上传仍靠右
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 预览分块超限提示（2026-11-16）

- **变更原因**：点「预览分块」看不到内容。大文件被网关拒绝后，抽屉里没有任何说明
- **变更前 vs 变更后**：~~失败只可能在角落弹出 HTTP 413~~ → 第二步列出文件名和大小；超过 50MB 时红字说明，预览和上传按钮不可点
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`nginx.conf`（单次请求上限 60MB）

## [变更] 文档抽屉标题与上传区（2026-11-16）

- **变更原因**：上传步骤应和列表在同一页；标题说明与标签数量重复
- **交互**：去掉上传按钮和「N 个文档」；标题栏收矮；上传两步常驻在文档列表上方
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 知识库文档抽屉改版（2026-11-16）

- **变更原因**：上传弹窗被抽屉挡住；设置标签和上传步骤混在一起
- **交互**：列表分页；点上传在列表上方展开两步；「设置默认」弹窗编辑库默认值。CfModal 提到 z-index 1100
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`CfModal.vue`、i18n

## [变更] 知识库列表搜索（2026-11-16）

- **变更原因**：首页卡片多了之后缺少按关键字过滤
- **交互**：列表上方搜索框，不写标签；占位符说明搜名称或说明；输入框自带搜索图标
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`i18n/messages.ts`

## [变更] 知识库去掉目录层（2026-11-16）

- **变更原因**：库名已经是检索单元的名字，目录分组多余
- **交互**：首页平铺知识库卡片；创建、编辑、详情不再出现目录
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 知识库层级与目录操作（2026-11-16）

- **变更原因**：向量必须库级锁定；分段默认 vs 本次上传要能分清；「新建目录」不能看起来像一个目录名
- **交互**：编辑知识库时向量库/模型不可改；设置与上传文案分开；下拉底部分隔线 + 小加号 + 强调色
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`i18n/messages.ts`

## [变更] 知识库预览最大化（2026-11-15）

- **变更原因**：预览区偏窄，编辑长文不便；弹窗已有最大化图标，预览应对齐
- **交互**：预览工具栏「最大化 / 还原」；最大化时隐藏左侧文档列表并加高阅读区
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 知识库检索测试多行输入（2026-11-15）

- **变更原因**：单行小框 + 纯图标按钮不协调，问句常需多行
- **交互**：改为 Textarea；Enter 检索、Shift+Enter 换行；下方 Top-K/最低分 +「清空/检索」文字按钮
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`i18n/messages.ts`

## [变更] 知识库分段字段帮助提示（2026-11-15）

- **变更原因**：分段方式/块大小/重叠等术语不易懂，需要悬停说明
- **交互**：设置页与上传「分段设置」标签旁加 `?` 图标，悬停解释智能递归、重叠、Top-K 等
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`i18n/messages.ts`

## [变更] 模型页 Scope + 权限（2026-11-15）

- **变更原因**：一期供应商可见范围与写权限；公开配置全站共用，私有仅本人可改
- **交互**：徽章区分公开/私有/全站默认/我的默认；admin 可开「公开」；无写权限只读；设默认：管理员写全站，普通用户写个人偏好
- **包含代码**：Web `pages/models.vue`、`useProviders.ts`、i18n；后端见 API v1.58

## [变更] 弹窗内下拉可用（2026-11-15）

- **变更原因**：上传「分段方式」等 USelect 在 CfModal 里点不开
- **交互**：弹窗内 Select `portal=false`；面板允许溢出；body 单独滚动
- **包含代码**：`CfModal.vue`、`knowledge.vue`、`main.css`

## [变更] 弹窗取消去 X 图标（2026-11-15）

- **变更原因**：页脚「取消」与 CfModal 右上角关闭重复
- **交互**：知识库各弹窗取消只留文字
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 卡片悬停显示操作 + 删除确认（2026-11-15）

- **变更原因**：图标常显吵；删除要人工确认防误触
- **交互**：悬停/焦点才显示操作图标（触屏常显）；删库、删文档均弹窗确认后才执行
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、i18n

## [变更] 知识库卡片图标操作（2026-11-15）

- **变更原因**：卡片只有「⋯」，操作不好找；对齐 MCP / 智能体列表的图标按钮
- **交互**：右上「文档 / 编辑 / 删除」；单击卡片仍开抽屉，双击仍看详情
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] CfModal 内下拉不被挡住（2026-11-15）

- **变更原因**：弹窗里点「目录」等 USelect，aria 已展开但看不到列表（内容层无 z-index，被 CfModal z-200 盖住），体感像卡顿/点不动
- **处理**：`main.css` 把打开态 content/listbox 提到 `z-index:300`；知识库表单 Select 再加 `ui.content`；CfModal 根节点 `pointer-events:none`，遮罩/面板再开回
- **包含代码**：`assets/css/main.css`、`components/CfModal.vue`、`pages/knowledge.vue`

## [变更] 文档预览 / Markdown 编辑（2026-11-15）

- **变更原因**：右侧预览只能看，无法改演示 Markdown
- **交互**：可编辑文档工具栏「预览 | 编辑」；保存后自动重新解析；未保存切换会确认
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`useKnowledgeBases.ts`；Go `PUT .../documents/:docId/content`

## [变更] 知识库设置页（2026-11-15）

- **交互**：抽屉第三页「设置」——默认分段、默认检索、可见范围（仅自己）；组织共享文案占位
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、i18n

## [变更] 检索测试 Top-K / 最低分（2026-11-15）

- **交互**：检索测试面板增加 Top-K、最低相似度；打开知识库时用库默认填充
- **包含代码**：`cogniforge-web/pages/knowledge.vue`

## [变更] 知识库上传分段预览（2026-11-15）

- **交互**：上传弹窗两步——选文件 → 分段方式/块大小/重叠 +「预览分块」→ 确认上传
- **包含代码**：`cogniforge-web/pages/knowledge.vue`、`useKnowledgeBases.ts`（`previewChunks` / `uploadDoc` 带 parse options）

## [变更] 知识库目录（2026-11-15）

- **交互**：创建/编辑表单「目录」下拉；选项列表末尾「+ 新建目录」弹窗新建（不在下拉框右侧）；首页按目录分段展示
- **包含代码**：`pages/knowledge.vue`、`useKnowledgeBases.ts`

## [变更] 知识库文档预览（2026-11-15）

- **变更原因**：文档名不可点，用户无法查看正文
- **交互**：抽屉加宽；点文档名打开右侧预览（Markdown 渲染）；对标 Dify/FastGPT「列表 + 阅读区」
- **包含代码**：`cogniforge-web/pages/knowledge.vue`；依赖 Go `GET .../documents/:docId`

## [变更] 前端 SSE 流式加固（2026-11-11）

- **变更原因**：多处手写 `getReader`，半截事件/取消/断线处理不一致
- **交互**：仍用 `fetch` + 流读取（不用 EventSource / fetch-event-source）；连接失败自动再试 1 次；传一半断线则保留已显示文字并提示重发
- **包含代码**：`cogniforge-web` `utils/sseStream.ts`、Playground / Agent / 提示词试跑

## [变更] 评估页聊天模型下拉（2026-11-09）

- **变更原因**：跑评估时目标为 model / prompt，聊天模型要手抄 ID，还得去模型页看
- **交互**：改为 `GET /api/v1/models` 下拉；默认选第一项；列表为空时提示去「配置 → 模型」
- **包含代码**：`cogniforge-web` `pages/evals.vue`、i18n

## [变更] 模型页厂商=卡片小字（2026-11-08）

- **变更原因**：卡片第二行来自 `provider`。新增默认 `openai`，格子文案又像「点了也不算数」，所以选了智谱或只填了智谱地址，小字仍是 OpenAI
- **交互**：格子改称「厂商」。点中后下方预览卡片（大字名称、小字厂商）。新增不预选。地址像智谱等已知厂商时，未选则自动对上；已选不一致时出现「改成智谱 BigModel」，只改小字、保留已填地址和模型名。「自定义兼容」的小字固定为「自定义兼容」
- **包含代码**：`pages/models.vue`、`useProviders.ts`；Go 列表时修正误标的 `openai`

## [变更] Playground Agent 提示词变量（2026-11-07）

- **变更原因**：17.14；对话侧补齐与试跑/评估一致的变量表单
- **包含代码**：`playground.vue`、`PlaygroundConfigPanel.vue`、`useAgentChat.ts`
- **交互**：选中绑定 Prompt 的 Agent → 参数滑层出变量表单；发送带 `variables`；必填未填拦截

## [变更] 智谱 BigModel 供应商（2026-11-08）

- **变更原因**：用户有智谱资源包/免费额度，模型页需可选
- **包含代码**：`useProviders.ts` `PROVIDER_META.zhipu`；`pages/models.vue` 提示；Go `CatalogModels` + `/v4` 路径
- **交互**：类型选「智谱」→ 默认 `glm-4.7-flash`；下拉免费 Flash 在前，资源包模型靠后

## [变更] 模型页快捷模板（不锁死）（2026-11-08）

- **变更原因**：供应商类型格子固定，且模型只能下拉，列表不全时用不了
- **交互**：格子改为「快捷模板」预填；Base URL / 模型始终可手填；推荐改成点选芯片；编辑也可换模板；增加「自定义兼容」
- **包含代码**：`pages/models.vue`、`useProviders.ts`；Go Update 可改 `provider`

## [变更] 提示词试跑 / 评估变量表单 / 顶栏邮箱（2026-10-24）

- **变更原因**：阶段十七 P1；验人设、评估填变量、同名用户区分
- **包含代码**：`cogniforge-web` `pages/prompts.vue`、`pages/evals.vue`、`layouts/default.vue`、`PromptVarForm.vue`
- **交互**：提示词详情「试跑」；评估选 prompt 时变量为表单；顶栏姓名下显示邮箱

## [变更] CfModal 叠在侧栏上可点（2026-11-15）

- **变更原因**：知识库抽屉（`USlideover`/Dialog）打开时 body 为 `pointer-events:none`，Teleport 出的 CfModal 看得见但点选/上传无反应
- **包含代码**：`cogniforge-web/components/CfModal.vue`（`pointer-events: auto` + `z-index: 200`）；`knowledge.vue` 未选文件时 toast
- **变更前 vs 变更后**：~~叠在 Slideover 上点击无效~~（2026-11-15）→ 可选文件、点上传正常

## [变更] 共用 CfModal（拖动/缩放/最大化）（2026-10-18）

- **变更原因**：Nuxt UI `UModal` 仅有 `fullscreen`，不支持拖动与改尺寸；弹窗需统一体验
- **包含代码**：`cogniforge-web/components/CfModal.vue`；全站 `UModal` → `CfModal`
- **能力**：标题栏拖动；四边/角缩放；最大化/还原（双击标题栏）；Esc/遮罩关闭；API 兼容 `#body`/`#footer`/`#content` + `ui.content` 宽度
- **未改**：`USlideover`（侧栏，非居中弹窗）

## [变更] MCP 卡片式管理 UI（2026-10-18）

- **变更原因**：业界（Cursor / Claude / Comis）以卡片+开关+工具数为主，表格密度过高
- **包含代码**：`cogniforge-web/pages/mcp.vue`；Go `tool_count`、`POST .../tools/refresh`
- **交互**：全部/内置/自定义筛选；搜索；卡片开关；点开看工具；HTTP「刷新工具」

## [变更] MCP 页工具明细与创建表单（2026-10-18）

- **变更原因**：内置三项只有标签无工具说明；新建无法指导飞书 HTTP 配置
- **包含代码**：`cogniforge-web/pages/mcp.vue`、`useMcpServers.ts`、i18n `mcp.*`
- **交互**：点行 → 滑层展示工具名/说明/参数；创建仅 HTTP（名称、说明、URL、鉴权、启用）+ 飞书路径提示
- **文案（同日更新）**：已支持 Streamable HTTP；stdio 需自建网关

## [变更] 模型页对话/向量用途（2026-08-25）

- **变更原因**：知识库不能用 deepseek-chat 做向量
- **包含代码**：Web `pages/models.vue`、`useProviders.ts`
- **变更前 vs 变更后**：~~一颗星统管全部~~（2026-08-25）→ 对话默认（星）与向量默认（数据库图标）分开

## [变更] 顶栏菜单整理（2026-08-22）

- **变更原因**：9 项平铺挤；模型/密钥是配置，用量不是设置
- **包含代码**：`cogniforge-web/constants/nav.ts`、`layouts/default.vue`、`i18n/messages.ts`
- **变更后** 顶栏：控制台 · 对话 · 智能体 · 工作流 · 知识库 · **配置 ▾**（模型、密钥）
- **用量**、**监控**（仅 admin）进头像菜单，不进个人设置页
- **~~顶栏平铺 Models / Keys / Usage / Monitor~~**（2026-08-22）

## [变更] 通用文件等 RAG 再做（2026-08-22）

- **变更原因**：发送框会遇到 PDF 等，但不能现在硬上
- **排期**：阶段十三，需求 `docs/01-requirements/04-playground-files-rag.md`
- **现在**：只图片；**~~发送框通用文件附件~~**（2026-08-22 排到 RAG 就绪后）

## [变更] 发送框粘贴访达文件（2026-08-22）

- **变更原因**：从访达复制图片再粘贴，剪贴板不带 `image/*`，原先进不了发送框
- **包含代码**：`cogniforge-web/utils/chatImages.ts`、`pages/playground.vue`
- **变更后**：粘贴 / 拖入同时读 `clipboardData.files` 与 `items`；空 MIME 按扩展名认 JPG/PNG/GIF/WebP；非图片提示只支持这些格式
- **不做**：PDF / 文档当附件（**阶段十三**，等 RAG 完全可用，见 `04-playground-files-rag.md`）

## [变更] 排队编辑走发送框（2026-08-22）

- **变更原因**：不要点文字/双击弹层；取消也不要做成发送区按钮
- **包含代码**：`PlaygroundQueueBar.vue`、`pages/playground.vue`
- **变更后**：铅笔进入编辑；发送框改内容；回车保存回排队；该行「编辑中 ×」取消；正在编辑的条目不自动发出
- **~~点文字弹出行内编辑框~~**（2026-08-22）

## [变更] Playground 排队 + 插入（2026-08-21）

- **变更原因**：流式中要连问；改中间一问要可控截断
- **包含代码**：`PlaygroundQueueBar.vue`、`pages/playground.vue`、`useConversations.ts`、`i18n/messages.ts`；需求 `docs/01-requirements/03-chat-queue-insert.md`
- **变更后**：
  - 流式中 Enter / 提交 → **加入排队**（最多 5）；队内点**编辑**进发送框改、删、↑↓；发送中锁定
  - 用户气泡 hover「在此后插入」→ 确认后 **硬截断**后续消息并发送；同时清空排队
  - 排队持久化 `message_queue`，刷新不丢
- **不做（本期）**：‹ › 分支树；分叉新会话；Cmd+Enter 停流并立刻发

## [变更] Playground 历史搜索（2026-08-21）

- **变更原因**：历史多了以后不好找
- **包含代码**：`cogniforge-web/components/PlaygroundHistoryPanel.vue`、`i18n/messages.ts`
- **变更后**：侧栏/滑层顶部搜索框，按标题本地过滤（含置顶段）；无匹配时提示「没有匹配的对话」
- **不做（本期）**：服务端全文检索消息正文

## [变更] Playground 上传发图 / vision（2026-08-21）

- **变更原因**：用户要能在对话里发图片给识图模型看
- **包含代码**：`cogniforge-web/pages/playground.vue`、`utils/chatImages.ts`；Go `ChatMessage.Content` 改为透传 string|parts；`ConversationMessage.images`
- **变更后**：
  - 输入框旁「添加图片」、可粘贴 / **拖入**；最多 4 张，单张 ≤4MB（客户端压缩）
  - 请求体 OpenAI 兼容 `image_url`；历史 JSON 存 `images`
  - 气泡展示用户附图；回复里的 Markdown 图仍走既有渲染
- **注意**：模型本身须支持 vision（如 GPT-4o）；纯文本模型会由上游报错

## [变更] Playground 聊天图片展示（2026-08-21）

- **变更原因**：模型回复里的 Markdown 图片 / 裸图片链接只显示文字，看不到图
- **包含代码**：`cogniforge-web/utils/chatMarkdown.ts`、`pages/playground.vue`
- **变更后**：
  - `![alt](url)` 与裸 `https://…png|jpg|gif|webp`、`data:image/…;base64,…` 渲染为图片
  - 图片限宽、圆角；点击新标签打开原图
  - 拒绝 `javascript:` / `svg+xml` 等不安全源
- **不做（本期）**：用户上传发图、多模态 vision 请求

## [变更] Playground 历史置顶（2026-08-21）

- **变更原因**：常用对话要能固定在历史列表顶部，方便反复打开
- **包含代码**：`cogniforge-web/components/PlaygroundHistoryPanel.vue`、`pages/playground.vue`、`composables/useConversations.ts`、`i18n/messages.ts`；Go `pinned` 字段
- **变更后**：
  - 历史侧栏分「置顶」与「历史」两段
  - 每条可 pin / unpin；置顶项带图钉标记
  - `PUT /api/v1/conversations/:id` 传 `{ pinned: true|false }`

## [变更] Resend 邮件忘记密码 UI（2026-08-20）

- **变更原因**：开通 Resend 后用户可自助重置
- **包含代码**：`forgot-password.vue`、`reset-password.vue`、`auth.global.ts`、i18n
- **变更前 vs 变更后**：~~仅管理员指引~~（2026-08-20）→ 若 `email_enabled` 显示邮箱表单；邮件链接打开 `/reset-password?token=`；未配置时仍显示管理员步骤

## [变更] 忘记密码入口（2026-08-20）

- **变更原因**：登录页缺少忘记密码
- **包含代码**：`cogniforge-web/pages/login.vue`、`pages/forgot-password.vue`、`pages/admin/users.vue`、`i18n/messages.ts`
- **变更前 vs 变更后**：~~登录只有注册链接~~（2026-08-20）→ 密码旁「忘记密码？」；说明页引导找管理员；用户管理可一键重置并复制临时密码
- **不做**：邮箱自助重置（待发信能力）

## [变更] 渐变氛围壳（2026-08-19）

- **变更原因**：宽屏两侧空白看起来像「没收拾完」；卡片白底硬边，和背景是两层皮。参考 JetBrains 官网：背景铺满渐变，内容居中，界面浮在光斑上
- **包含代码**：`cogniforge-web/assets/css/main.css`、`layouts/default.vue`、`layouts/auth.vue`、`pages/index.vue`
- **影响范围**：四套主题的 `--cf-bg-aura`；Dashboard 卡片；顶栏半透明
- **变更前 vs 变更后**：~~铺满窗口拉内容 / 或 1120 窄栏加硬卡片~~（2026-08-19）→ **背景满屏径向渐变**，内容 `max-width: 1200px` 居中；卡片半透明 + 轻光晕，顶栏透出背景

## [变更] 控制台铺满窗口（2026-08-19）

- **变更原因**：Dashboard 限宽 1120px，宽屏左右大块空白；下方两列还不等宽，看起来和上面的额度条对不齐
- **包含代码**：`cogniforge-web/assets/css/main.css`、`layouts/default.vue`、`pages/index.vue`、`pages/models.vue`
- **影响范围**：所有 `.cf-page` 列表/管理页 + Dashboard + 顶栏；登录/设置表单仍保持窄栏方便阅读
- **变更前 vs 变更后**：~~内容 max-width 1120px，顶栏 1200px~~（2026-08-19）→ ~~铺满窗口~~（同日改回居中 + 渐变背景，见上条）

## [变更] Usage 空用量进度条（2026-08-19）

- **变更原因**：`UProgress` 在值为 0 时会播放不确定进度动画，看起来像还在加载
- **包含代码**：`cogniforge-web/pages/usage.vue`
- **变更前 vs 变更后**：~~0 也在滑动的绿条~~（2026-08-19）→ 静止空槽；趋势图当天为 0 不再画一条矮绿柱

## [变更] 按钮能点用主题实心色（2026-08-19）

- **变更原因**：主按钮用浅青绿 `soft` 底，看起来像灰掉不能点；禁用态又残留绿色，像还能点
- **包含代码**：`cogniforge-web/components/CfButton.vue`、`assets/css/main.css`、`pages/playground.vue`
- **测试**：`utils/__tests__/cfButtonTone.spec.ts`；Playwright `e2e/login.spec.ts`、`e2e/playground.spec.ts`（`pnpm test:e2e`）
- **变更前 vs 变更后**：~~全部 soft 淡底~~（2026-08-19）→ 登录/保存/确认实心 `--cf-accent`；取消描边；禁用只灰色
- **不改**：API、主题切换入口、按钮文案与图标约定

## [变更] 用量配额 UI（2026-08-18）

- **变更原因**：Playground 无限刷；用户和管理员都看不见 Token
- **详细设计**：`docs/01-requirements/02-quota-design.md`（含柱状图 / 饼图 / 排行示意）
- **导航**：Keys 与 Monitor 之间新增 Usage `/usage`（admin 与 user 都可见）；配额策略 `/admin/quota` 仅 admin
- **Playground**：输入框上方轻量额度条；用尽禁用发送；文案走 i18n
- **Dashboard**：加一张「今日剩余条数」卡，点击进 `/usage`
- **图表**：全站只选一种库（建议 ECharts）；四张图规格见配额文档 §5.3
- **不改**：Monitor 仍只做 HTTP 日志，不和用量抢页

## [变更] 未配置模型时 Toast 提示（2026-08-16）

- **变更原因**：缺默认模型 / API Key 时，对话区会把 mock 句子当成模型回复
- **包含代码**：`pages/playground.vue`、`i18n/messages.ts`、`composables/useAgentChat.ts`
- **变更后**：HTTP 503 / `code=4010` 时 Toast：「请先到「模型」页填写 API Key，并设置一个默认模型后再对话。」不写入助手气泡

## [变更] DeepSeek V4 模型下拉（2026-08-16）

- **变更原因**：要能选 V4，但日常仍用更便宜的 `deepseek-chat`
- **包含代码**：`pages/models.vue`、`composables/useProviders.ts`；后端 `CatalogModels`
- **变更后**：DeepSeek 可选 chat / reasoner / v4-flash / v4-pro，**不删**旧选项

## [变更] Playground 历史侧栏可折叠（2026-08-16）

- **变更原因**：历史列表常开会占掉对话宽度
- **包含代码**：`cogniforge-web/pages/playground.vue`
- **变更后**：桌面标题左侧图标收起/展开历史；选择会记在浏览器里。手机仍用「历史对话」滑层

## [变更] Playground 历史对话 + 参数挪位（2026-08-16）

- **变更原因**：对话要能找回；左侧再放 Agent/滑条会挤掉历史
- **包含代码**：`cogniforge-web/pages/playground.vue`、`components/PlaygroundHistoryPanel.vue`、`composables/useConversations.ts`；Go `/api/v1/conversations`
- **影响范围**：Playground 布局与存档

### 变更前 vs 变更后

- **变更前**：桌面左侧 = Agent/模型/Temperature；消息只在内存，刷新就没了；「参数」只在窄屏出现
- **变更后**：
  - 桌面左侧 = 新对话 + 历史列表
  - Agent / 模型 / 滑条 **一律** 右上角「参数」滑层（桌面、手机相同）
  - 每轮结束后写入当前用户的 `chat_conversations`
  - 窄屏左侧隐藏，顶栏多一个「历史对话」滑层

## [变更] 模型下拉改为数据库配置（2026-08-15）

- **变更原因**：Agents / 工作流 LLM 节点写死 GPT/Claude，与「模型」页无关
- **包含代码**：`pages/agents.vue`、`pages/workflows/[id].vue`；后端 `GET /api/v1/models`
- **变更后**：下拉只显示已启用供应商的 `default_model`

## [变更] P7 各模块 editorial 页面壳（2026-08-15）

### 变更原因
Dashboard 已定「安静编辑室」视觉语言后，各业务模块仍有各自一套标题/面板/表格间距，需要统一。

### 包含代码
- `cogniforge-web/assets/css/main.css`：`.cf-page` / `.cf-page-header` / `.cf-page-title` / `.cf-page-sub` / `.cf-panel` / `.cf-state` / `.cf-data-table` / `.cf-section-title`
- 列表与管理页：`agents` / `workflows` / `knowledge` / `keys` / `models` / `monitor` / `admin/users` / `admin/roles`
- Settings 侧栏与 Profile/Security/Preferences/Sessions 章节标题；Playground 顶栏与空状态

### 变更后
- 页面标题与导航英文短标签一致（Agents / Flows / Knowledge / Keys / Models / Monitor / Users / Roles / Settings）
- 主内容区统一 max-width、serif 大标题、细边框面板、大写表头
- **不改** API、路由、Vue Flow、角色可见性

---

## [变更] Playground 迁 Nuxt UI Chat（2026-08-12）

### 变更原因
P3 落地：Playground 从 Naive UI 迁到 Nuxt UI 4 Chat 套件，实现助手左 / 用户右，并去掉本页全部 naive-ui / @vicons。

### 包含代码
- `cogniforge-web/pages/playground.vue`

### 变更后
- 对话区：`UChatMessages` + `UChatPrompt` + `UChatPromptSubmit` + `UChatShimmer`
- 适配层：本地 `{ role, content }` → AI SDK 形 `{ id, role, parts:[{type:'text',text}] }`（`mapToUIMessage`）；**不改后端 SSE / chat 契约**
- 侧栏：`USelect` / `USlider` / `UInput` / `UBadge` / `UIcon` / `UButton`；主题用 `--cf-*` + `cf-surface`（电青 accent）

## [变更] Nuxt UI + shadcn-vue 混用选型（2026-08-12）

### 变更原因
Nuxt UI 在 Chat、CommandPalette、FileUpload、可搜索 Select、Toast、Table 等「产品级」能力明显更强；shadcn-vue 更适合品牌壳层与可拷贝改造的基础件。二者均基于 **Reka UI + Tailwind**，可混用，但必须统一主题与冲突规则。

### 变更后策略（一句话）
**品牌壳 + 基础控件用 shadcn；高阶场景用 Nuxt UI；同一职责只保留一套。**

## [变更] 增加 UI 示意稿（2026-08-12）

### 变更原因
用可视化稿对齐「个人品牌技术风」，减少只看文字产生的理解偏差。

### 示意稿位置
`docs/05-frontend/assets/`

| 文件 | 画面 |
|------|------|
| `cogniforge-ui-login.png` | 登录 |
| `cogniforge-ui-console.png` | Dashboard（Desktop，原 Console） |
| `cogniforge-ui-playground.png` | Playground（旧：API 调试台，不推荐） |
| `cogniforge-ui-chat-v2.png` | Chat v2（阅读流、左右弱） |
| `cogniforge-ui-chat-v3-lr.png` | Chat v3（**推荐**：助手左 / 用户右 + Nuxt UI AI） |
| `cogniforge-ui-mobile.png` | Dashboard（手机兼容示意） |

> 示意稿表达气质与布局，正式实现以本文 Token / 组件约定为准；细节文案可微调。

## [变更] 响应式：Web 优先，手机可兼容（2026-08-12）

### 变更原因
产品需要能在手机打开，但核心锻造场景（工作流画布、复杂表格、Playground 多栏）以桌面为主。避免一上来双端并行导致延期。

### 变更后策略
- **P0**：桌面 Web（≥1024px）完整可用、视觉达标
- **P1**：平板（768–1023）不崩、可完成主流程
- **P2**：手机（<768）可读可登录可浏览；重编辑能力降级或引导「请用电脑」

### 影响范围
仅设计与前端布局；接口 / Vue Flow 引擎 / CI 不变。

## [变更] 产品 UI 从 Naive UI 企业后台风迁移到个人品牌技术风（2026-08-12）

### 变更原因
现有界面偏传统企业后台（顶栏菜单 + 白底灰区 + Naive 组件堆叠），品牌感弱，不像独立产品。目标改为流行产品感、个人品牌、技术风，同时降低 UI 库锁定成本。

### 包含代码（实施阶段）
- `cogniforge-web/` 全站页面与布局（`layouts/`、`pages/`、`components/`）
- 新增 `components/ui/`（shadcn-vue）
- 保留 `components/WorkflowCanvas.vue` 与 `@vue-flow/*` 图编辑逻辑
- **不改** `composables/*` 的请求路径、入参、出参类型
- **尽量不改** `Dockerfile` / `nginx.conf` / `.github/workflows/*` / `docker-compose-web.yml`

### 影响范围
| 范围 | 是否变更 |
|------|----------|
| 视觉与交互 | 是 |
| 路由路径 | 否（保持现有 pages） |
| API 入参/出参 | 否 |
| 工作流图引擎 | 否（继续 Vue Flow） |
| 打包产物路径 | 否（仍 `pnpm build` → `.output/public`） |
| CI/CD 流程 | 否（脚本名与镜像阶段保持） |

### 变更前 vs 变更后

| 维度 | 变更前 | 变更后 |
|------|--------|--------|
| UI 库 | Naive UI（+ 残留 Element Plus） | shadcn-vue（Reka UI + Tailwind） |
| 样式 | 组件库主题 + 局部 CSS | Tailwind CSS v4 + CSS 变量设计系统 |
| 图标 | `@vicons/ionicons5` | `lucide-vue-next` |
| 气质 | 企业后台 | 个人品牌 · 技术产品 |
| 工作流 | Vue Flow + Naive 节点壳 | Vue Flow 保留；仅换节点外壳样式 |

---

## 1. 产品气质（必须先对齐）

### 1.1 一句话定位
**CogniForge = 个人/小团队的 AI Agent 锻造台**：像工作室工具，不像公司 OA。

### 1.2 要像什么 / 不要像什么

| 要像 | 不要像 |
|------|--------|
| Linear / Cursor / Vercel Dashboard 的克制产品感 | 传统 Admin Template |
| 品牌名一眼可见 | 只有左上角小字 Logo |
| 技术感来自排版、网格、等宽标签 | 发光、赛博霓虹、紫渐变堆料 |
| 每个区块只干一件事 | 密密麻麻统计卡 + 多卡片仪表盘 |

### 1.3 品牌测试（第一屏）
去掉顶栏后，页面仍应能认出这是 **CogniForge**，而不是随便一个后台。  
登录页与控制台首页必须把 **CogniForge** 做成主视觉信号（字重/字号/位置优先于功能标题）。

---

## 2. 视觉设计系统

### 2.1 主题切换（更绚丽，可挑选）

产品壳与组件不变，**换的是背景氛围 + 色板 Token**。实现上用 `data-theme="aurora|inknight|citrus|glass"` 挂在 `html`，配合 Nuxt UI `primary` 映射。

| ID | 名称 | 气质 | 背景 | 主强调色 | 适合 |
|----|------|------|------|----------|------|
| `aurora` | **极光薄雾**（默认推荐） | 亮、通透、有光感 | 青→天蓝柔和晕染 + 轻纹理 | 电青 `#0f9f8a` | 大多数用户、品牌首页感 |
| `inknight` | **墨夜信号** | 深色科技、冷静锐利 | 石墨底 + 淡青网格/角光 | 青霓 `#2ee6c7` | 夜间、偏开发者 |
| `citrus` | **青柠锻场** | 明快、活泼一点 | 薄荷lime与冰蓝交叠 | 青绿 `#0d9f6e` + 点缀琥珀 | 想更「年轻产品」 |
| `glass` | **玻璃工坊** | 高级、层次多 | 浅底 + 玻璃拟态面板 + 轻彩光斑 | 电青 + 极淡副彩 | 想更「华丽 SaaS」 |

**默认：`aurora`。** 原极简纸色 Forge Ink 可保留为内部第五档 `ink`（可选，不作为默认）。

#### 示意稿

| 主题 | 文件 |
|------|------|
| Aurora | `assets/cf-theme-aurora.png` |
| Ink Night | `assets/cf-theme-inknight.png` |
| Citrus | `assets/cf-theme-citrus.png` |
| Glass | `assets/cf-theme-glass.png` |

#### Token 示例（概念）

```css
html[data-theme="aurora"] {
  --cf-bg: #eef7f8;
  --cf-bg-elevated: rgb(255 255 255 / 0.72);
  --cf-ink: #12151a;
  --cf-accent: #0f9f8a;
  --cf-bg-aura: radial-gradient(1200px 600px at 10% -10%, #b8f1e8 0%, transparent 55%),
                radial-gradient(900px 500px at 90% 0%, #cfe8ff 0%, transparent 50%),
                #eef7f8;
}

html[data-theme="inknight"] {
  --cf-bg: #0e1116;
  --cf-bg-elevated: #161b22;
  --cf-ink: #e8eef5;
  --cf-accent: #2ee6c7;
  --cf-bg-aura: radial-gradient(800px 400px at 100% 0%, #12352f 0%, transparent 60%),
                #0e1116;
}

html[data-theme="citrus"] {
  --cf-bg: #f3faf3;
  --cf-bg-elevated: #ffffff;
  --cf-ink: #14201a;
  --cf-accent: #0d9f6e;
  --cf-accent-2: #e6a23c; /* 仅小点缀 */
  --cf-bg-aura: radial-gradient(900px 500px at 0% 0%, #d4f5c8 0%, transparent 55%),
                radial-gradient(800px 480px at 100% 0%, #d9f0ff 0%, transparent 50%),
                #f3faf3;
}

html[data-theme="glass"] {
  --cf-bg: #f4f7fb;
  --cf-bg-elevated: rgb(255 255 255 / 0.55);
  --cf-ink: #12151a;
  --cf-accent: #0f9f8a;
  --cf-bg-aura: radial-gradient(600px 400px at 20% 10%, rgb(15 159 138 / 0.18), transparent),
                radial-gradient(500px 360px at 80% 0%, rgb(100 160 255 / 0.14), transparent),
                #f4f7fb;
  --cf-blur: blur(16px);
}
```

页面根背景使用 `background: var(--cf-bg-aura)`；卡片/顶栏用 `var(--cf-bg-elevated)`（Glass / Aurora 可加 `backdrop-filter`）。

#### 切换入口
- 设置页「外观」：四选一卡片预览（小色板）
- 可选：顶栏太阳/主题快捷切换（次要）
- 持久化：`localStorage` + 若有用户偏好 API 则写入（**无接口则先本地**，不强改后端）

#### 绚丽但不踩坑
- 背景可以花，**正文对比度仍要够**（WCAG 思路）
- 不做大面积紫靛渐变、不做闪烁霓虹、不把主按钮做成全圆胶囊
- Chat 左右分侧、导航模块在所有主题下行为一致

### 2.2 方向名与默认

默认主题名：**Aurora Mist（极光薄雾）**。  
品牌测试仍成立：第一屏品牌字 + 氛围底，去掉顶栏仍能认出 CogniForge。

### 2.3 色彩 Token

各主题共用同一套语义名（`--cf-bg` / `--cf-ink` / `--cf-accent` / `--cf-line` / `--cf-ok` / `--cf-warn` / `--cf-danger`），具体色值随 `data-theme` 变。  
Nuxt UI：`app.config` / runtime 把 `primary` 绑到 `--cf-accent`。

### 2.4 字体

| 用途 | 字体 | 理由 |
|------|------|------|
| 品牌 / 大标题 | **Syne**（或 Space Grotesk） | 有个性，不像系统后台 |
| 正文 / UI | **IBM Plex Sans** | 可读、偏工程 |
| 代码 / Trace / Key | **JetBrains Mono** | 技术风锚点 |

加载方式：Google Fonts 或自托管 woff2；在 `nuxt.config` `app.head` 引入，正文用 Tailwind `font-sans` / `font-display` / `font-mono` 映射。

### 2.5 造型与密度

- 圆角：`6px` / `10px`（组件），**不用** `rounded-full` 做主按钮
- 边框：1px `--cf-line`，靠线条分区，不靠厚卡片阴影
- 间距：页面水平桌面 `--cf-page-pad` 32px、平板 24px、手机 16px；**背景满屏渐变**，主内容 `--cf-content-max: 1200px` 居中（和顶栏对齐）。登录/注册卡片、设置表单仍可更窄（约 400 / 800px）
- 卡片：**默认无卡**；仅在「可交互容器」或「需要隔离的编辑区」使用表面块（Glass 主题允许更多半透明面）
- 背景氛围：由主题 `--cf-bg-aura` 提供，可绚丽；正文区保持清晰

### 2.6 动效（至少 3 处有意为之）

1. 路由切换：主内容区短 fade（150–200ms）
2. 顶栏当前项：下划线/色条滑动
3. 主 CTA / 列表行：hover 时 `translateY(-1px)` + 边框色微变（触摸设备以 `:active` / 透明度反馈代替 hover）  
禁止装饰性无限闪烁与大段视差。  
`prefers-reduced-motion: reduce` 时关闭非必要动画。

### 2.7 响应式策略（Web 优先）

**总原则：先保证桌面 Web 完美，再让手机「能用且不丑」，不为手机重做产品。**

| 断点 | Tailwind | 目标 |
|------|----------|------|
| Desktop | `lg` ≥1024px | **验收主场**：完整导航、多栏布局、表格、画布 |
| Tablet | `md` 768–1023 | 顶栏可折叠；双栏改单栏堆叠 |
| Mobile | 默认 <768 | 汉堡菜单 + Sheet；列表可浏览；重操作降级 |

**布局约定（写代码时就要带上，避免事后返工）：**
1. 壳层：桌面横排导航；`<lg` 收成 `Sheet`/`Drawer` 菜单（shadcn `Sheet`）
2. 栅格：统计/双栏用 `grid` + `md:`/`lg:` 断点，禁止写死只适配 1440
3. 表格：桌面 `Table`；窄屏优先横滑（`overflow-x-auto`），P2 再考虑卡片列表
4. 触摸：可点区域 ≥ 44px；主按钮全宽仅在 `<md`
5. 安全区：底部固定条预留 `env(safe-area-inset-*)`（P1/P2）

**按模块的端能力：**

| 模块 | Desktop | Mobile（P2） |
|------|---------|--------------|
| 登录 / 注册 | 完整 | 完整 |
| 控制台 | 完整 | 完整（单列） |
| 列表页（Agent/模型/知识库/密钥） | 完整 | 浏览 + 简易新建；复杂筛选可收纳 |
| Playground | 左历史 + 中对话 | 单栏：对话为主；历史/参数都进 Slideover |
| 工作流画布（Vue Flow） | **完整编辑** | **只读/提示用电脑编辑**（不阻塞 Web 交付） |
| 监控 / 管理表 | 完整 | 横滑查看；批量操作可隐藏 |

实施顺序仍以 Desktop 验收为准：某一页 Desktop 未过，不开始该页 Mobile 精细打磨。

### 2.8 按钮语义（能点 / 不能点）

浅青绿软底看起来像灰掉，用户分不清能不能点。全站 `CfButton` 与对话发送钮统一如下。色值跟当前主题 `--cf-accent`，不要另做一套高饱和按钮。

| 场景 | tone | 能操作时 | 不能操作时 |
|------|------|----------|------------|
| 登录 / 保存 / 确认 / 新建 / 发送 | `primary` | 实心 `--cf-accent`，字用 `--cf-on-accent` | 灰底灰字，不留绿色 |
| 取消 / 返回 / 重置 | `secondary` | 细边框 + 正文色，浅底 | 同上灰色 |
| 删除确认 | `danger` | 实心 `--cf-danger`，白字 | 同上灰色 |
| 表格行图标 | `icon` / `icon-accent` / `icon-danger` | 透明底 + 对应色图标 | 同上灰色 |

规则：

- **灰色只表示不能点**。禁用态禁止再用浅 accent / 半透明主色。
- **能点的主操作必须是实心主题色**，不要 `soft` 淡底。
- 圆角 8px，不用全圆胶囊；不要霓虹描边。
- Playground 发送钮同样走这套：有输入且额度未用尽 = 实心主题色；空输入或额度用尽 = 灰。

包含代码：`cogniforge-web/components/CfButton.vue`、`assets/css/main.css`、`pages/playground.vue`

---

## 3. 技术栈（最新且兼容 Nuxt 3）

### 3.1 目标组合

| 层 | 选型 | 版本策略 | 说明 |
|----|------|----------|------|
| 框架 | Nuxt **4**（由 3 升级） | 跟随 `@nuxt/ui` peer | UI 4.10 要求 Nuxt ≥4.1；Nuxt 3 已 EOL |
| 运行时 | Vue 3 | 跟随 Nuxt peer | 已是 3.5.x |
| 样式 | Tailwind CSS **v4** | 与 `@nuxt/ui` 官方一致 | 单一 Tailwind 实例 |
| UI A | **shadcn-vue** + `shadcn-nuxt` | 按需 CLI 拉取 | 品牌壳、基础件、可深度改源码 |
| UI B | **`@nuxt/ui`**（Nuxt UI） | 最新兼容 Nuxt 3 的稳定版 | Chat / 命令面板 / 上传 / 可搜索选择 / Toast 等 |
| 原语 | 两者均基于 `reka-ui` | 随库版本 | 无障碍与交互模型一致，利于混用 |
| 图标 | Nuxt UI 自带 `@nuxt/icon`；品牌区可用 lucide | 最新稳定 | 避免再引 ionicons |
| 状态 | Pinia + `@pinia/nuxt` | 保持 | 不改业务状态模型 |
| 工作流画布 | **@vue-flow/core** 等 | **保持现版本线** | 先不动引擎与数据结构 |
| 测试 | Vitest | 保持 | composable 测试优先保留 |

> 完整「谁用哪个」见 **§6 组件选型矩阵**。

### 3.2 移除依赖（实施时）

- `naive-ui`、`@bg-dev/nuxt-naiveui`
- `element-plus`、`@element-plus/icons-vue`（已无业务引用）
- `@vicons/ionicons5`（图标迁到 Nuxt Icon / lucide）

### 3.3 兼容性约束

```
Nuxt 3.21.x
  └─ Vite 7（已随 Nuxt）
       ├─ Tailwind CSS v4
       ├─ @nuxt/ui          → U* 高阶组件（主题对齐 Forge Ink）
       ├─ shadcn-nuxt       → components/ui/*（品牌与基础件）
       └─ @vue-flow/*       （client-only 画布；逻辑冻结）
```

- 当前产品默认 `ssr: false`；迁移后继续 **SPA/CSR 为主**。
- `app.vue` 需包一层 Nuxt UI 的 `UApp`（Toast / Tooltip / 程序化 Overlay 依赖）。
- 主题：**以 Forge Ink CSS 变量为真相**；把 Nuxt UI `primary` 映射到电青 `#0f9f8a`，避免两套色打架。
- `Dockerfile` 仍：`pnpm install --frozen-lockfile` → `pnpm build` → Nginx 托管 `.output/public`。
- GitHub Actions 仍调用既有 `pnpm` script，不改 job 结构。

### 3.4 初始化约定

```bash
# Tailwind + Nuxt UI
pnpm add @nuxt/ui tailwindcss

# shadcn（品牌/基础件）
pnpm dlx nuxi@latest module add shadcn-nuxt
pnpm dlx shadcn-vue@latest init
pnpm dlx shadcn-vue@latest add button input label textarea dialog \
  dropdown-menu sheet separator badge avatar tabs skeleton
```

`assets/css/main.css` 概念结构：

```css
@import "tailwindcss";
@import "@nuxt/ui";
/* Forge Ink tokens + shadcn 变量映射写在其后 */
```

目录：

```
components/
  ui/                 # shadcn 生成，可改（品牌基础件）
  layout/             # AppShell / AppNav / BrandMark（偏自研 + shadcn）
  chat/               # Playground 适配层（包一层 Nuxt UI Chat*）
  workflow/           # 仅样式壳；引擎仍走 WorkflowCanvas
  WorkflowCanvas.vue  # Vue Flow：逻辑冻结
```

工具函数：`utils/cn.ts`（`clsx` + `tailwind-merge`）。

### 3.5 混用冲突规则（必须遵守）

1. **壳层与导航统一 Nuxt UI**，不要再用 shadcn 另做一套顶栏。
2. **主按钮全站优先 `UButton`**，避免同页混用 shadcn `Button`。
3. **Toast 只用 Nuxt UI**（`UToast` / `useToast`）。
4. **表单**：复杂表单 `UForm`；极简登录也可用 `UInput` + `UButton` 以保持一套。
5. **圆角 / 主色 / 字号**走 Forge Ink；`app.config.ts` 把 Nuxt UI `primary` 映射电青。
6. **营销向 Page\*** 默认不用。
7. Chat 做适配层，**不改后端接口**。
8. shadcn：**能不用则不用**；若引入，禁止与 Nuxt UI 职责重叠（尤其 Button/Nav）。

---

## 4. 信息架构与壳层

### 4.1 路由与导航 IA（2026-08-22 整理；2026-10-15 增补提示词/评估）

> 路由和权限不变，只改顶栏分组。~~2026-08-12 冻结的 9 项平铺~~（2026-08-22）

## [变更] 配置菜单：提示词 + 评估联动（2026-10-15）

- **变更原因**：阶段十七提示词工程；评估页要能选 prompt / 临时 system
- **包含代码**：`cogniforge-web` `constants/nav.ts`、`pages/prompts.vue`、`pages/evals.vue`、`i18n/messages.ts`
- **变更前 vs 变更后**：~~配置无提示词入口~~（2026-10-15）→ `/prompts`；`/evals` 运行可选 model|agent|prompt + system_prompt

#### 顶栏主模块

| 顺序 | 显示名 | key | 路由 | 可见角色 |
|------|--------|-----|------|----------|
| 1 | 控制台 | `dashboard` | `/` | admin, user |
| 2 | 对话 | `playground` | `/playground` | admin, user |
| 3 | 智能体 | `agents` | `/agents` | admin, user |
| 4 | 工作流 | `workflows` | `/workflows` | admin, user |
| 5 | 知识库 | `knowledge` | `/knowledge` | admin, user |
| 6 | 配置 ▾ | `config` | （下拉，无独立页） | admin, user |
| 6a | 模型 | `models` | `/models` | admin, user |
| 6b | 密钥 | `keys` | `/keys` | admin, user |
| 6c | MCP | `mcp` | `/mcp` | admin, user |
| 6d | 提示词 | `prompts` | `/prompts` | admin, user |
| 6e | 评估 | `evals` | `/evals` | admin, user |

激活态：`/models`、`/keys`、`/mcp`、`/prompts`、`/evals` 时「配置」高亮。`/` 仅精确匹配控制台。

#### 用户下拉菜单

| 文案 | 行为 | 可见 |
|------|------|------|
| 个人设置 | `/settings` | 全部登录用户 |
| 用量 | `/usage` | 全部登录用户 |
| 监控 | `/monitor` | **仅 admin** |
| 用户管理 | `/admin/users` | **仅 admin** |
| 角色权限 | `/admin/roles` | **仅 admin** |
| 配额策略 | `/admin/quota` | **仅 admin** |
| 退出登录 | logout API | 全部登录用户 |

#### 认证相关路由（保持）

| 路由 | 说明 |
|------|------|
| `/login` | 登录 |
| `/register` | 注册 |

#### 实施时（Nuxt UI）
- `UNavigationMenu` / `UHeader` 的 items **必须按上表生成**，并沿用 `roles` 过滤
- 窄屏 `USlideover` 内菜单与桌面**同一数据源**，禁止桌面有、手机无
- 允许视觉缩短显示名（如「Agent」），但 **tooltip / 无障碍名仍用完整模块名**，且 `to` 不变

### 4.2 App Shell（登录后）

**Desktop（主验收）— 用 Nuxt UI 壳**

```
┌──────────────────────────────────────────────────────────────┐
│  [BrandMark]     UNavigationMenu …                  UUser/菜单│  ← UHeader
├──────────────────────────────────────────────────────────────┤
│  UMain：页面标题 + 内容                                        │
└──────────────────────────────────────────────────────────────┘
```

推荐组件：
- `UApp`（根）
- `UHeader` + leading 插 `BrandMark`（大字 CogniForge）
- `UNavigationMenu`（横导航，绑定 Nuxt 路由）
- 用户区：`UDropdownMenu` / `UAvatar`
- `<lg`：`USlideover` 或 Dashboard 自带的 mobile toggle

**Mobile / Tablet**

```
┌─────────────────────────────┐
│  CogniForge          ☰  User│  ← Header + toggle
├─────────────────────────────┤
│  单列主内容                   │
└─────────────────────────────┘
        ☰ → USlideover 导航
```

规则：
- 品牌名在 Header leading 里要足够大（自研 BrandMark）
- **模块清单、路由、角色过滤见 §4.1，必须完整保留**
- 路由仍用 Nuxt `pages/`，与用哪套 UI 无关

### 4.3 认证壳

登录/注册：全屏纸色底 + 淡网格；**上方巨型品牌字**（窄屏改纵向堆叠，避免左右分栏在手机上挤扁），表单用细线面板，最大宽约 `400px` 居中。

---

## 5. 分页面 UI 设计

> 原则：接口字段与操作流程不变，只换呈现。

### 5.1 登录 / 注册
- 主视觉：`CogniForge` + 一行副文案「Forge your agents.」
- 表单：优先 **shadcn** `Input` / `Button` / `Label`（品牌可控）
- 错误：行下轻提示；成功/失败可用 **Nuxt UI Toast**
- 注册密码强度：保留现有逻辑组件，换皮

### 5.2 控制台 `/`
- 第一屏只保留：品牌问候一句 + 4 个关键指标（轻表面，非统计大盘）+ 「下一步」列表
- 「最近活动」空状态：自研 `EmptyState` 或 `UEmpty`（二选一，全站统一）
- **不要**把监控图表、公告、多营销块塞进首页

## [变更] 升级 Nuxt 4 以使用 Nuxt UI 4（2026-08-12）

### 变更原因
`@nuxt/ui@4`（含 AI Chat）要求 Nuxt ≥4.1；Nuxt 3 已 EOL。为落地 Header/Chat 等组件，将 `cogniforge-web` 升级到 Nuxt 4.5。

### 变更后
- `nuxt` 4.5.x + `@nuxt/ui` 4.10 + Tailwind 4.3
- P0：四主题 CSS、`UHeader` 导航壳（IA 冻结）、顶栏主题快捷切换
- 业务页仍暂时包着 Naive（过渡）；API / Vue Flow / Docker 流程不变
- `pnpm build` 已通过

## [变更] 四套主题确认采纳（2026-08-12）

### 变更原因
产品确认 Aurora / Ink Night / Citrus / Glass 四套风格均喜欢，作为正式可切换主题落地。

### 变更后
- **四套全部保留**，设置页可切换
- **默认主题：Aurora Mist（极光薄雾）**
- 导航模块、路由、接口、Vue Flow 约束不变

## [变更] 可切换绚丽主题（2026-08-12）

### 变更原因
原 Forge Ink 偏克制素净；期望背景与整体气质更绚丽，并支持多套风格切换（设置里可选，默认选一套更亮的）。

### 变更后
- 提供 **4 套主题预设**，同一套组件/导航，只换 CSS 变量与背景层
- **默认推荐：Aurora Mist（极光薄雾）** — 亮色、有氛围、仍好读
- 用户可在「个人设置 → 外观」切换；未登录页跟随系统默认主题
- 示意：`assets/cf-theme-aurora.png` / `cf-theme-inknight.png` / `cf-theme-citrus.png` / `cf-theme-glass.png`

## [变更] 导航 IA 冻结：只换皮不换模块（2026-08-12）

### 变更原因
换 Nuxt UI 壳时，必须保留现网全部导航模块、路由与角色可见性，避免「好看但少功能」。

### 硬约束
- **不得删除、合并、改路径**现有模块入口
- **不得改变** admin / user 可见范围
- 用户菜单项与退出登录流程保持（可换组件，不可少能力）
- 中文标签可微调用词，但模块含义必须一一对应

## [变更] 壳层与导航改用 Nuxt UI（2026-08-12）

### 变更原因
Nuxt UI 的 Header / NavigationMenu / Dashboard / Slideover 开箱即好看、和路由集成省事；shadcn 同等效果需要自己拼布局与交互。Chat 已锁定 Nuxt UI，壳层同族可减少两套导航心智。

### 变更后策略（更新）
**产品壳 + 导航 + Chat + 表格等高阶 → Nuxt UI；shadcn 仅保留「要改到骨子里」的少数基础件或暂不引入重复件。**

| 层 | 用谁 | 说明 |
|----|------|------|
| 顶栏 / 导航 / 移动菜单 | **Nuxt UI** | `UHeader`、`UNavigationMenu`、`USlideover` / `UDashboardNavbar` |
| 路由本身 | **Nuxt 文件路由** | 库不负责路由；两边都一样 |
| Chat | **Nuxt UI** | 已定 |
| Table / Upload / Toast / SelectMenu | **Nuxt UI** | 已定 |
| 品牌字标 BrandMark | **自研一小块** | 只要把「CogniForge」做得够大、够认，可插在 `UHeader` 的 leading 槽 |
| shadcn | **可选补充** | 仅当某原子控件 Nuxt UI 不够改时再加；避免 Button 两套混用 |

### 变更前 vs 变更后（壳层）

| | 变更前（v1.3） | 变更后（v1.6） |
|--|----------------|----------------|
| AppShell | 自研 + shadcn | **Nuxt UI Header/Nav 为主** |
| 品牌 | 自研优先 | 自研 BrandMark 插入 Nuxt UI 槽位 |
| 基础 Button | shadcn 为主 | **与壳统一：优先 `UButton`**（全站一套） |

### 变更原因
需要明确区分说话方（左右），且认可 Nuxt UI AI Chat 套件的交互完成度；在「阅读流」基础上补回左右布局。

### 变更后
- **助手 → 左侧**（`side=left`，`variant=naked` 或轻 `soft`）
- **用户 → 右侧**（`side=right`，`variant=soft`，宽度约 ≤75%）
- 实现：**必须用 Nuxt UI** `UChatMessages` / `UChatMessage` / `UChatPrompt` / `UChatPromptSubmit` / `UChatShimmer`（可按需 `UChatReasoning` / `UChatTool`）
- 示意：`assets/cogniforge-ui-chat-v3-lr.png`

### 5.3 Playground（Chat 专项）

#### 市面参考（学交互，不抄皮）

| 产品 | 值得学 | 不适合照搬到 CogniForge |
|------|--------|-------------------------|
| **Claude.ai** | 居中阅读柱、助手像文档、底部大输入框 | 完全无左右时辨识弱（我们要左右） |
| **ChatGPT / 豆包** | 用户右、助手左的对话方位清晰 | 过消费化圆泡 |
| **Cursor Chat** | 紧凑、代码友好、操作靠近消息 | IDE 深色嵌套感过强 |
| **Nuxt UI Chat 模板** | 组件齐全：分侧、actions、流式、Prompt | 默认主题需换成 Forge Ink |
| **Kimi / 通义** | 建议 Chip、上传入口清楚 | 过运营感 |

#### 结论：用 Nuxt UI AI 组件
**正式锁定 Nuxt UI Chat 套件为 Playground 对话实现。**  
左右分侧直接用组件自带的 `side`：`UChatMessages` 默认助手左、用户右。  
主题映射 Forge Ink（电青 primary），避免官方默认绿/模板脸。

#### Forge Chat 交互规格（Desktop 优先）

1. **左右**：助手左 + 可选小头像/图标；用户右 + `soft` 浅底气泡（电青淡底，非微信绿）
2. **主舞台**：居中对话柱约 `720–800px`，仍保持产品阅读感，不是调试台
3. **Composer**：`UChatPrompt` + `UChatPromptSubmit` 贴底；`⌘/Ctrl + Enter` 发送；支持停止生成
4. **消息操作**（hover `actions`）：复制、重新生成
5. **空态**：品牌问候 + 3 个建议 Chip
6. **参数**：右上角「参数」滑层（Agent / 模型 / Temperature / Max Tokens / Top P）；**不再**放左侧窄轨
7. **历史**：桌面左侧列表，可用顶栏图标折叠；手机顶栏「历史对话」滑层；数据走 `GET/POST/PUT/DELETE /api/v1/conversations`；支持 **置顶**（`pinned`，列表上方单独「置顶」段）；支持 **按标题搜索**（本地过滤）；支持 **排队**（`message_queue`）与 **插入截断**
7b. **图片展示**（2026-08-21）：消息 Markdown 图片与裸图片 URL / base64 渲染为 `<img>`；点击打开原图
7c. **上传发图**（2026-08-21）：作曲区附图 / 粘贴 / 拖入；OpenAI 兼容多模态 `content` parts；历史存 `images[]`；需 vision 模型
8. **Token / 延迟**：角落轻量元数据，不做底栏大字报
8b. **额度条**（2026-08-18）：输入框上方显示今日次数/Token 进度；≥80% 黄灯；用尽禁用发送并 Toast「明天 0 点（北京时间）恢复」
9. **流式**：`UChatShimmer`；Markdown 继续现有渲染路径
10. **适配层**：现有 Playground composable → `mapToUIMessage()`（AI SDK `parts` 形）→ Nuxt UI；聊天补全出参不变
11. **禁止**：等宽 role 标签墙、API Playground 时间戳列、微信式双绿泡

手机（P2）：单列仍保持左右分侧；历史与参数都进 Slideover。

### 5.4 Agent / 模型 / 知识库 / 密钥
- 列表：优先 **`UTable` + `UPagination` + `UBadge`**
- 筛选/模型选择：优先 **`USelectMenu` / `UInputMenu`**（可搜索，比裸 Select 强）
- 新建/编辑：`UModal` 或 `USlideover`；简单确认也可用 shadcn `Dialog`（全站弹层建议统一 Nuxt UI Overlay）
- 知识库上传：**`UFileUpload`**

### 5.5 工作流（**Vue Flow 先不动引擎**）
- 列表页：同其他资源列表（`UTable`）
- 编辑器页（Desktop）：
  - 顶栏工具换皮：按钮体系与全站一致（见 §6，基础按钮走选定的那一套）
  - 左侧节点面板、右侧属性面板换皮
  - **画布核心**：继续 `WorkflowCanvas.vue` + `@vue-flow/*`
  - 节点外壳允许换成 Tailwind class；**节点 type、handle id、events、存盘 JSON 结构不变**
- 编辑器页（Mobile，P2）：只读提示「请用电脑」
- 本阶段明确：**不升级 Vue Flow 大版本、不换库、不改编排协议**

### 5.6 监控中心
- 指标 + **`UTable`** 日志；筛选用 `USelectMenu` / `UInput`
- 状态码、耗时：mono
- **不承担** Token 配额图表（那是 `/usage`）

### 5.6b 用量 Usage（2026-08-18）
- 路由 `/usage`，页面壳 `cf-page`
- 页头：今日次数、今日 Token、本月 Token 三张 `stat-card` + 两个 `UProgress`
- 主图：近 7/30 天柱状图；旁侧模型占比饼图
- admin 可切「全站」并看到 Top 用户表（`UTable`）
- 图表库全站只留一种，建议 ECharts；组件：`pages/usage.vue`、`composables/useQuota.ts`、`components/QuotaBar.vue`

### 5.7 设置 / 管理
- 设置分区：可用 shadcn `Tabs` 或 `UTabs`（全站选一种）
- 复杂表单：优先 **`UForm` + `UFormField`**
- 管理 CRUD：`UTable` + `UModal`

---

## 6. 组件选型矩阵（Nuxt UI vs shadcn）

### 6.1 怎么判断（简单规则）

| 情况 | 选谁 |
|------|------|
| 顶栏、导航、移动菜单、Dashboard 分栏 | **Nuxt UI**（好写好看） |
| Chat / Table / Upload / Toast / 可搜索 Select / Form | **Nuxt UI** |
| 品牌字标、极个别要改到源码的原子 | **自研 / 可选 shadcn** |
| 路由 | **Nuxt 自带**（与 UI 库无关） |
| 工作流画布 | **Vue Flow** |
| 营销 Page\* | **不用** |

> **直答你的问题**：导航和框架壳用 **Nuxt UI 更好写也更好看**。shadcn 更像零件箱，同等导航要自己设计拼装；我们不再以 shadcn 做壳。

### 6.2 对照表（按 CogniForge 实际用到的）

| 能力 | Nuxt UI | shadcn-vue | **本项目选用** | 理由 |
|------|---------|------------|----------------|------|
| 顶栏 / 导航 / 移动菜单 | **`UHeader` `UNavigationMenu` `USlideover`** | 需自拼 | **Nuxt UI** | 好看、省事、响应式现成 |
| Dashboard 分栏 | `UDashboard*` | 自研 | **按需 Nuxt UI** | Playground 三栏可借力 |
| BrandMark | 槽位插入 | 自研 | **自研插入 Header** | 保证品牌测试通过 |
| Button / Input | `UButton` `UInput` | 有 | **Nuxt UI** | 与壳一套，避免混用 |
| Select（可搜索） | `USelectMenu` | 基础 Select | **Nuxt UI** | 长列表 |
| Form | `UForm` | 自搭 | **Nuxt UI** | |
| Table / Pagination | `UTable` | 基础 Table | **Nuxt UI** | |
| Overlay / Toast | `UModal` `UToast`… | Dialog/sonner | **Nuxt UI** | |
| AI Chat | `UChat*` | 无对等 | **Nuxt UI** | 已定，左右分侧 |
| FileUpload | `UFileUpload` | 自研 | **Nuxt UI** | |
| Command palette | `UCommandPalette` | 需拼 | **Nuxt UI** | |
| shadcn 整包 | — | — | **默认不作为主栈** | 减少双库；确有缺口再单点引入 |
| 工作流画布 | 无 | 无 | **Vue Flow** | |

### 6.3 按页面落库（实施对照）

| 页面 | Nuxt UI | 自研 / 备注 |
|------|---------|-------------|
| 全局壳 | Header、NavigationMenu、Slideover、Toast、UApp | BrandMark 插入 leading |
| Login / Register | Input、Button、Form、Toast | 大标题品牌区自研 |
| Console | Button、可选 Empty | 指标块可用简单 section |
| Playground | **Chat* 左右分侧**、SelectMenu、Slider、可选 DashboardPanel | 元数据小条 |
| Agents / Models / Keys / Admin | Table、Pagination、Modal/Slideover、SelectMenu | PageHeader 可用 UPageHeader 或自研一句 |
| Knowledge | Table、FileUpload、Modal | |
| Workflows | Table、Toast、Slideover；画布 Vue Flow | 节点壳换皮 |
| Settings | Form、FormField、Tabs、Toast | |

### 6.4 从 Naive 迁移时的替换方向

| 现用（Naive） | 目标 |
|---------------|------|
| `n-layout` / `n-menu` 顶栏 | `UHeader` + `UNavigationMenu` |
| `n-button` / `n-input` | `UButton` / `UInput` |
| `n-select`（长列表） | `USelectMenu` |
| `n-data-table` | `UTable` |
| `n-modal` / `n-drawer` | `UModal` / `USlideover` |
| `n-dropdown` | `UDropdownMenu` |
| `useMessage` | `useToast` |
| Playground 对话区 | `UChatMessages` + 左右 `side` |
| Vue Flow 节点内图标 | Nuxt Icon |

---

## 7. 实施分期（降低风险）

### Phase 0 — 基建（不改业务页）
1. 接入 Tailwind v4 + Forge Ink + `@nuxt/ui`（`UApp`，primary=电青）
2. 搭 `UHeader` + `UNavigationMenu` 壳（BrandMark 插入）
3. shadcn：**默认跳过**；有明确缺口再加
4. `pnpm build` / Docker 回归

### Phase 1 — 认证 + Shell（**Desktop 验收**）
- 登录注册用 `UForm`/`UInput`/`UButton`
- 默认布局切换到 Nuxt UI Header 导航

### Phase 2 — 列表型页面（Desktop 优先）
- Console / Agents / Models / Knowledge / Keys / Monitor / Admin
- 列表统一 `UTable`；上传用 `UFileUpload`

### Phase 3 — 复杂页（Desktop 优先）
- Playground（**Nuxt UI Chat* + 适配层**）
- Settings（`UForm`）
- Workflows 壳层（画布逻辑冻结）
- 可选：`UCommandPalette`（⌘K）

### Phase 4 — 清理
- 删除 Naive/Element 依赖与 `plugins/naive-ui.client.ts`
- 更新 `cogniforge-web/README.md` 技术栈说明
- 回归 composable 单测

### Phase 5 — 手机兼容打磨（不阻塞上线）
- 导航滑层、Playground 参数滑层、表格横滑、工作流只读提示
- 真机抽测登录与浏览

每期合并标准：**接口契约不变、关键 Desktop 路径可用、构建产物路径不变**。  
Mobile 问题记入 backlog，**不阻断** Web 主路径合并。

---

## 8. 明确不在本次范围

- 不改后端 API、不改字段名
- 不换 Vue Flow 为其它流程图库
- 不强制升级 Nuxt 4（可另开分支评估）
- 不重做产品信息架构（**旧模块/路由/角色可见性冻结，见 §4.1**；2026-08-18 允许**新增** Usage / 配额策略，不删旧项）
- 不为「更炫」加 3D/粒子/大英雄营销首屏
- **不做**独立原生 App；手机以响应式 Web 为准
- **不承诺**手机端完整工作流可视化编排（P2 降级即可）

---

## 9. 验收标准（给不懂代码的同学）

### 9.1 Web 桌面（必须过，否则不算完成）
1. 打开登录页，第一眼看到的是 **CogniForge** 品牌，而不是普通灰表单。
2. 登录后顶栏像产品，但**以前的模块都能找到**（控制台、Playground、Agent、模型、工作流、知识库、密钥、用量；admin 另有监控/用户/角色/配额策略）。
3. 点每个导航，进入的还是原来的页面路径。
4. 原来会用的功能都在，按钮能点，数据还能出来。
5. 工作流画布还能拖节点、连线、保存（和现在一样）。
6. 服务器部署方式不用重学：还是原来的镜像和发布流程。

### 9.2 手机（加分 / 后续，不挡 Web）
1. 能打开登录页并登录。
2. 能进控制台、看列表，不出现左右严重错位或点不到的按钮。
3. 工作流编辑若不好用，有清楚提示「请用电脑」，而不是白屏或坏掉。

---

## 10. 文档关系

| 文档 | 角色 |
|------|------|
| 本文 `03-ui-redesign-shadcn.md` | **UI 重设计唯一真相**（视觉 + 技术迁移） |
| `01-frontend-design.md` | 历史前端设计；涉及冲突时以本文为准 |
| `02-settings-module-design.md` | 设置功能逻辑仍有效；控件换成 shadcn |
| `99-dev-plan/01-development-plan.md` | 记录本里程碑进度 |

---

**文档版本**: v1.10  
**最后更新**: 2026-08-12  
**维护**: CogniForge 前端  
**分支建议**: `feat/nuxt-shadcn-vue`
