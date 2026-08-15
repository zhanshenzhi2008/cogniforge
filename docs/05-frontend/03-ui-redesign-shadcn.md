# CogniForge 产品 UI 重设计（shadcn-vue）

## [变更记录]

| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|----------|--------|
| 2026-08-15 | v1.31 | Playground：兼容非 SSE 完整 JSON，避免发送后空白气泡 | orjrs |
| 2026-08-15 | v1.30 | P8：中英文界面切换（无路由前缀；设置/顶栏/登录可切） | orjrs |
| 2026-08-15 | v1.29 | P7.8：注册页英文对齐登录；测试弹窗去掉多余关闭；文档补 P7.7 | orjrs |
| 2026-08-15 | v1.28 | P7.6：工作流画布壳 + 登录/注册提交迁 `CfButton`；业务页改造收口 | orjrs |
| 2026-08-15 | v1.27 | P7.5：Settings / Dashboard / Playground 顶栏迁 `CfButton` | orjrs |
| 2026-08-15 | v1.26 | P7.4：Flows / Admin / Monitor 迁 `CfButton` | orjrs |
| 2026-08-15 | v1.25 | P7.3：Keys / Knowledge 迁 `CfButton`（大按钮带图标；行内 icon-only） | orjrs |
| 2026-08-15 | v1.24 | 个人标定稿 Solitary √3；技能 `orjrs-brand` 跨项目复用 | orjrs |
| 2026-08-15 | v1.23 | 品牌分层：个人标 √3/orjrs 跨项目复用；产品名仅作字标 | orjrs |
| 2026-08-15 | v1.22 | Favicon 改为 √3 + orjrs 签名（sqrt3lab DNA），降低撞标风险 | orjrs |
| 2026-08-15 | v1.21 | 站点 favicon / 品牌标：青底 C+火花；顶栏与登录同用 | orjrs |
| 2026-08-15 | v1.20 | CfButton：大按钮必带图标；文字可并排或 label-mode=tip | orjrs |
| 2026-08-15 | v1.19 | 统一 `CfButton`：全员有底；卡片 icon-only；primary=保存 soft | orjrs |
| 2026-08-15 | v1.18 | 按钮观感：保存/添加改 soft，编辑改 primary ghost；禁用实心青绿块 | orjrs |
| 2026-08-15 | v1.17 | 清除残留 naive 配置（`.npmrc` hoist）与历史文档示例中的 naive 引用 | orjrs |
| 2026-08-15 | v1.16 | 常用按钮规范；Models 卡片/列表切换 + 详情只读同弹窗；校正字体/列表实现描述 | orjrs |
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

## [变更] Playground 流式空白兜底（2026-08-15）

- **变更原因**：`/chat/stream` 在有 Provider 时上游曾误发非流式 JSON，页面只认 SSE `delta`，助手气泡为空
- **包含代码**：`pages/playground.vue`（解析兜底）；后端见 `internal/chat/service.go`
- **变更后**：SSE `delta.content` 逐字追加；若整包是 `choices[0].message.content` 也显示；空 assistant 不再塞进下一轮上下文
- **未改**：UChat 布局、API 路径

## [变更] P8 中英文界面（2026-08-15）

- **变更原因**：设置里已有语言选项但不换界面；页面中英混排
- **包含代码**：`cogniforge-web/i18n/messages.ts`、`composables/useLocale.ts`、`components/LanguageSwitch.vue`；业务页文案走 `t()` / `d()`
- **行为**：仅 `zh-CN` / `en-US`；**不改路由**（无 `/en` 前缀，IA 冻结）；本地 `localStorage` `cf-locale`；登录后 `PUT /api/v1/settings` `{ language }`（现有接口，入参不变）
- **入口**：顶栏地球图标、登录/注册右上角、设置 → Preferences
- **默认**：有本地记录用本地；否则浏览器 `zh*` → 中文，其余英文
- **未改**：API 契约、Vue Flow 逻辑、产品名 CogniForge

## [变更] P7.8 注册英文 + 收口残留（2026-08-15）

- **变更原因**：登录已是英文，注册仍中文；Models 测试结果弹窗底栏「关闭」与右上角 X 重复；设计文档未记 P7.7
- **包含代码**：`pages/register.vue`、`pages/models.vue`、`layouts/default.vue`（汉堡菜单改 `CfButton`）
- **已完成未单独成章**：Users/Roles 双击行编辑；Monitor 双击日志详情（无底栏 Close）；Keys 显示/复制
- **刻意保留 `UButton`**：顶栏主题下拉触发器（`UDropdownMenu` 槽）；登录/注册文字 link；`CfButton` 内部封装

## [变更] P7.6 画布壳 + 登录提交迁 CfButton（2026-08-15）

- **变更原因**：工作流编辑器顶栏/抽屉、登录注册主按钮尚未对齐 §2.8
- **包含代码**：`pages/workflows/[id].vue`、`pages/login.vue`、`pages/register.vue`
- **硬约束**：Vue Flow 画布逻辑未改
- **刻意保留 `UButton`**：顶栏主题/菜单图标（`layouts/default.vue`）；登录注册文字 link；`CfButton` 内部封装

## [变更] P7.5 Settings / Dashboard / Playground 迁 CfButton（2026-08-15）

- **变更原因**：设置与对话顶栏仍混用 ghost `UButton`
- **包含代码**：`ProfileSection.vue`、`SecuritySection.vue`、`SessionsSection.vue`、`pages/index.vue`、`pages/playground.vue`
- **影响范围**：资料保存、改密、会话撤销、Dashboard CTA、Playground Params/清空/建议词
- **未改**：工作流画布壳、登录/注册 link、顶栏主题/菜单图标（壳层控件）

## [变更] P7.4 Flows / Admin / Monitor 迁 CfButton（2026-08-15）

- **变更原因**：列表页按钮尚未对齐 §2.8
- **包含代码**：`pages/workflows/index.vue`、`pages/admin/users.vue`、`pages/admin/roles.vue`、`pages/monitor.vue`
- **影响范围**：工作流列表、用户/角色、监控日志；画布编辑器与 Settings 未改
- **变更前**：ghost 文字按钮、分页 outline、日志详情多余「关闭」
- **变更后**：大按钮带图标；行内 icon-only；分页 secondary；日志详情只靠右上角 X

## [变更] P7.3 Keys / Knowledge 迁 CfButton（2026-08-15）

- **变更原因**：Models / Agents 已统一 `CfButton`；Keys / Knowledge 仍混用 ghost `UButton`
- **包含代码**：`pages/keys.vue`、`pages/knowledge.vue`
- **影响范围**：密钥列表/创建/撤销；知识库卡片菜单、详情弹窗、文档抽屉、上传与删除确认
- **变更前**：页头/底栏/行内多为 `UButton` ghost 或无图标
- **变更后**：大按钮必带图标；行内仅图标 + tip；危险操作用 `tone="danger" strong`

## [变更] Solitary √3 个人标（2026-08-15）

- **中文意图**：孤独的根号3
- **英文定名**：**Solitary √3**（the lonely √3）
- **资产与规则**：个人技能 `~/.cursor/skills/orjrs-brand/`；图标只含 √3 + orjrs；产品名（CogniForge）只做旁边字标

## [变更] 站点品牌 Favicon（2026-08-15）

- **变更原因**：浏览器标签页此前指向不存在的 `/favicon.ico`，显示默认地球图标
- **包含代码**：`cogniforge-web/public/favicon.svg` 等；`nuxt.config.ts`；`BrandMark.vue`；登录/注册页
- **影响范围**：标签页图标、顶栏品牌、登录/注册品牌区
- **变更前**：无真实 favicon → 曾用 C+火花 → 现为 **Solitary √3 + orjrs**
- **变更后**：个人标（孤独的根号3）跨项目复用；本仓库产品字标为 CogniForge

## [变更] CfButton 大按钮必带图标（2026-08-15）

- **变更原因**：大按钮只有文字显得空、不好认；图标统一后更整齐好看
- **包含代码**：`CfButton.vue`、`main.css`、`pages/models.vue`、`pages/agents.vue`
- **影响范围**：Models / Agents 弹窗与页头；其余模块后续继续迁到 `CfButton`
- **变更前**：大按钮可无图标；文字只有并排一种
- **变更后**：`primary` / `secondary` / `danger` 必有图标；默认图标+文字；`label-mode="tip"` 时只显示图标、文字进悬浮
- **关键差异**：新增 `labelMode`；默认图标表；大按钮间距/图标尺寸微调

## [变更] 清除 naive 残留引用（2026-08-15）

### 变更原因
业务代码已卸 naive-ui，但仍残留 `.npmrc` hoist，以及 `01`/`02` 设计文档示例中的 `naive-ui` / `n-*` 片段，容易误导后续实现。

### 变更后
- 删除 `cogniforge-web/.npmrc`（原 `@bg-dev/nuxt-naiveui` / `naive-ui` hoist）
- `01-frontend-design.md` / `02-settings-module-design.md` 示例改为 Nuxt UI 写法；产品实现仍以本文为准
- 变更记录中「曾卸载 naive」的历史摘要保留，不作代码依赖

---

## [变更] 按钮规范 + Models 详情体验（2026-08-15）

### 变更原因
全站保存 / 编辑 / 删除等按钮用法不统一，弹窗底栏观感差；Models 卡片过高且缺少详情只读入口。需在设计文档固化约定，并与实现对齐。

### 包含代码
- `cogniforge-web/pages/models.vue`：卡片默认 + 列表切换、双击只读详情、同弹窗编辑、底栏左右分区
- `docs/05-frontend/03-ui-redesign-shadcn.md`：§2.8 常用操作按钮规范；校正字体与列表实现描述

### 变更后
- 见 **§2.8**：主操作 / 次操作 / 危险操作 / 行内图标按钮约定
- Models：默认紧凑卡片（约 3–5 列），可切列表；双击进只读详情，详情与编辑同一 `UModal`
- 资源列表优先 **editorial 原生表**（`.cf-data-table`），不强制 `UTable`

### 变更前 vs 变更后（Models）

| 项 | 变更前 | 变更后 |
|----|--------|--------|
| 列表形态 | 高大双列卡片 | 紧凑多列卡片（默认）+ 列表切换 |
| 详情 | 无 / 直接编辑 | 双击只读详情 → 可切编辑 |
| 弹窗底栏 | 按钮挤在左侧 | 左关闭/返回，右主操作 |

---

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
| 品牌 / 大标题（`--font-display`） | **Source Serif 4**（回退 Syne / IBM Plex Sans） | 安静编辑室气质；与 Dashboard 标题一致 |
| 正文 / UI | **IBM Plex Sans**（或系统无衬线回退） | 可读、偏工程 |
| 代码 / Trace / Key / 模型名 | **ui-monospace / JetBrains Mono** | 技术风锚点 |

加载方式：在 `nuxt.config` / CSS 引入；标题用 `font-display` / `.cf-page-title`，正文走 Tailwind `font-sans`。

### 2.5 造型与密度

- 圆角：`6px` / `10px`（组件），**不用** `rounded-full` 做主按钮
- 边框：1px `--cf-line`，靠线条分区，不靠厚卡片阴影
- 间距：页面水平桌面 `20–28px`、手机 `16px`；内容宽默认 `~1120px`（Models 等可放宽至 `~1360px`）
- 页面壳：统一 `.cf-page` / `.cf-page-header` / `.cf-panel` / `.cf-data-table`（见 P7）
- 卡片：**默认无卡**；资源总览可用**紧凑卡片网格**（可交互容器）；列表型数据优先细边框表格
- 背景氛围：由主题 `--cf-bg-aura` 提供，可绚丽；正文区保持清晰

### 2.6 动效（至少 3 处有意为之）

1. 路由切换：主内容区短 fade（150–200ms）
2. 顶栏当前项：下划线/色条滑动（editorial）或胶囊（island）
3. 主 CTA / 列表行 / 供应商卡片：hover 时 `translateY(-1px)` + 边框色微变（触摸设备以 `:active` / 透明度反馈代替 hover）  
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
1. 壳层：桌面横排导航；`<lg` 收成 `USlideover` 菜单
2. 栅格：统计/双栏用 `grid` + `md:`/`lg:` 断点，禁止写死只适配 1440
3. 表格：桌面 `.cf-data-table`（或 `UTable`）；窄屏优先横滑（`overflow-x-auto`）；Models 等可提供卡片/列表切换
4. 触摸：可点区域 ≥ 44px；主按钮全宽仅在 `<md`
5. 安全区：底部固定条预留 `env(safe-area-inset-*)`（P1/P2）

**按模块的端能力：**

| 模块 | Desktop | Mobile（P2） |
|------|---------|--------------|
| 登录 / 注册 | 完整 | 完整 |
| Dashboard | 完整 | 完整（单列；活动区可隐藏） |
| 列表页（Agent/模型/知识库/密钥） | 完整 | 浏览 + 简易新建；复杂筛选可收纳 |
| Models | 卡片默认 + 列表切换；双击详情 | 单列卡片；详情弹窗 |
| Playground | 三栏 | 单栏：对话为主，参数进 Slideover |
| 工作流画布（Vue Flow） | **完整编辑** | **只读/提示用电脑编辑**（不阻塞 Web 交付） |
| 监控 / 管理表 | 完整 | 横滑查看；批量操作可隐藏 |

实施顺序仍以 Desktop 验收为准：某一页 Desktop 未过，不开始该页 Mobile 精细打磨。

### 2.8 按钮系统（`CfButton` 统一调用）

> **实现入口**：`cogniforge-web/components/CfButton.vue`  
> 全站业务按钮优先用 `CfButton`，不要再手写一堆不同的 `UButton` color/variant。  
> 底层仍是 Nuxt UI `UButton`，但对外只暴露 **场景 tone**。

#### 2.8.1 设计原则

1. **都有背景**：一律 soft 浅色填充
2. **大按钮必有图标**：`primary` / `secondary` / `danger` 必须带 `icon`（未传有默认）
3. **文字两种展示**：默认图标+文字并排；`label-mode="tip"` 时只显示图标，文字进悬浮
4. **小空间**：`icon*` 永远仅图标 + `tip`
5. **保存感**：`tone="primary"` = 浅青 soft
6. **危险**：确认删除 `tone="danger" strong`

#### 2.8.2 Tone 一览

| `tone` | 场景 | 外观 | 文案 |
|--------|------|------|------|
| `primary` | 保存 / 添加 / 页头 New / 详情里「编辑」 | 浅青底 | **必有图标** + 文字（或 `label-mode="tip"`） |
| `secondary` | 取消 / 测试连接 / 返回详情 / 关闭 | 中性浅底 | **必有图标** + 文字（或 tip） |
| `danger` | 确认框删除 | 浅红底；`strong` 时实心红 | **必有图标** + 文字 |
| `icon` | 卡片/行内次要工具（测试、星标…） | 中性浅底小方钮 | **仅图标** + `tip` |
| `icon-accent` | 卡片/行内编辑 | 浅青小方钮 | **仅图标** + `tip` |
| `icon-danger` | 卡片/行内删除 | 浅红小方钮 | **仅图标** + `tip` |

#### 2.8.3 调用示例

```vue
<!-- 页头新建：图标 + 文字并排（默认好看） -->
<CfButton tone="primary" icon="i-lucide-plus" @click="create">Add provider</CfButton>

<!-- 弹窗底栏：大按钮都带图标 -->
<CfButton tone="secondary" icon="i-lucide-x" @click="close">取消</CfButton>
<CfButton tone="primary" icon="i-lucide-check" type="submit" form="f" :loading="saving">保存</CfButton>

<!-- 大按钮只显示图标，文字进悬浮 -->
<CfButton tone="secondary" icon="i-lucide-x" tip="关闭" label-mode="tip" @click="close" />

<!-- 卡片操作（无文字） -->
<CfButton tone="icon-accent" icon="i-lucide-pencil" tip="编辑" @click="edit" />
<CfButton tone="icon" icon="i-lucide-refresh-cw" tip="测试连接" @click="test" />
<CfButton tone="icon-danger" icon="i-lucide-trash-2" tip="删除" @click="askDelete" />

<!-- 危险确认 -->
<CfButton tone="danger" strong icon="i-lucide-trash-2" :loading="deleting" @click="confirm">删除</CfButton>
```

#### 2.8.4 弹窗底栏布局

```
┌─────────────────────────────────────────────┐
│  [secondary 详情]     [secondary] [primary] │
└─────────────────────────────────────────────┘
```

- 只读详情：不放「关闭」文案（右上角 X）；右下 `secondary` 测试 + `primary` 编辑
- 编辑/新建：右下取消 + 保存/添加
- 用 `.modal-actions` / `__left` / `__right`

#### 2.8.5 反例（禁止）

- 裸 `ghost` 无背景文字按钮当主操作
- **大按钮只有文字、没有图标**
- 卡片里「编辑」「删除」带长文字挤布局
- 同页混用未封装的各色 `UButton` 冒充规范
- `rounded-full` 胶囊主按钮、紫渐变

> 参考落地页：`pages/models.vue`、`pages/agents.vue`、`pages/keys.vue`、`pages/knowledge.vue`、`pages/workflows/index.vue`、`pages/admin/*`、`pages/monitor.vue`。


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
2. **主按钮全站优先 `CfButton`**（底层仍是 `UButton`），避免同页混用 shadcn `Button`。**保存 / 编辑 / 删除等语义见 §2.8**。
3. **Toast 只用 Nuxt UI**（`UToast` / `useToast`）。
4. **表单**：复杂表单 `UForm`；极简登录也可用 `UInput` + `UButton` 以保持一套。
5. **圆角 / 主色 / 字号**走主题 Token；`app.config.ts` 把 Nuxt UI `primary` 映射 `--cf-accent`。
6. **营销向 Page\*** 默认不用。
7. Chat 做适配层，**不改后端接口**。
8. shadcn：**能不用则不用**；若引入，禁止与 Nuxt UI 职责重叠（尤其 Button/Nav）。
9. **资源列表**：优先自研 editorial 表（`.cf-data-table`）+ `UBadge`/`CfButton`；复杂排序筛选再考虑 `UTable`。
10. **弹窗底栏**：左次操作、右主操作（§2.8.2）。
---

## 4. 信息架构与壳层

### 4.1 路由与导航 IA（冻结，与现网一致）

> 来源：现网 `layouts/default.vue`。UI 重设计**只换组件实现**，不改信息架构。

#### 顶栏主模块（按角色过滤）

| 顺序 | 显示名 | key | 路由 | 可见角色 |
|------|--------|-----|------|----------|
| 1 | Dashboard | `dashboard` | `/` | admin, user |
| 2 | Play | `playground` | `/playground` | admin, user |
| 3 | Agents | `agents` | `/agents` | admin, user |
| 4 | Models | `models` | `/models` | admin, user |
| 5 | Flows | `workflows` | `/workflows` | admin, user |
| 6 | Knowledge | `knowledge` | `/knowledge` | admin, user |
| 7 | Keys | `keys` | `/keys` | admin, user |
| 8 | Monitor | `monitor` | `/monitor` | **仅 admin** |

激活态规则（保持）：
- `/` 仅精确匹配 Dashboard
- 其余：`path === to` 或 `path.startsWith(to + '/')`（如 `/workflows/:id` 仍高亮 Flows）

#### 用户下拉菜单（保持）

| 文案 | 行为 | 可见 |
|------|------|------|
| 个人设置 | 跳转 `/settings` | 全部登录用户 |
| 用户管理 | 跳转 `/admin/users` | **仅 admin** |
| 角色权限 | 跳转 `/admin/roles` | **仅 admin** |
| 退出登录 | 调现有 logout API → `clearAuth` → `/login` | 全部登录用户 |

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
6. **参数**：左侧窄轨 / 顶部模型 `USelectMenu`；高级参数默认折叠
7. **Token / 延迟**：角落轻量元数据，不做底栏大字报
8. **流式**：`UChatShimmer`；Markdown 继续现有渲染路径
9. **适配层**：现有 Playground composable → `mapToUIMessage()`（AI SDK `parts` 形）→ Nuxt UI；**不改后端出参**
10. **禁止**：等宽 role 标签墙、API Playground 时间戳列、微信式双绿泡

手机（P2）：单列仍保持左右分侧；参数进 Slideover。

### 5.4 Agent / 模型 / 知识库 / 密钥
- 页面壳：`.cf-page` + 英文标题（与导航一致）
- 列表：优先 **`.cf-data-table`（原生表）+ `UBadge` + 行内 `CfButton`**；分页可用 `UPagination` 或简易 pager
- 筛选/模型选择：优先 **`USelect` / `USelectMenu` / `UInput`**
- 新建/编辑：统一 **`UModal`**；按钮语义见 **§2.8**
- **Models 专项**：
  - 默认**紧凑卡片网格**（约 3–5 列），可切换列表；偏好存 `localStorage`（`cf-models-view`）
  - **双击**卡片/行 → 只读详情（同弹窗）；详情可切编辑；API Key 脱敏展示
  - 页头 Active 条保持单行，避免大块横幅占高
- **Agents / Knowledge**：同样支持双击只读详情 → 同弹窗编辑；Knowledge 单击仍打开文档抽屉（与双击防抖区分）
- 知识库上传：**`UFileUpload`**
- 知识库也可卡片网格（业务需要时）；密钥/Agent 以表为主（Agent 表 + 双击详情）
- **Keys**：行内显示/隐藏 + 复制（未揭示时复制脱敏并 toast 提示）；无独立详情弹窗

### 5.5 工作流（**Vue Flow 先不动引擎**）
- 列表页：`.cf-data-table` + §2.8 按钮
- 编辑器页（Desktop）：
  - 顶栏工具换皮：按钮体系与全站一致（§2.8 / §6）
  - 左侧节点面板、右侧属性面板换皮
  - **画布核心**：继续 `WorkflowCanvas.vue` + `@vue-flow/*`
  - 节点外壳允许换成 Tailwind class；**节点 type、handle id、events、存盘 JSON 结构不变**
- 编辑器页（Mobile，P2）：只读提示「请用电脑」
- 本阶段明确：**不升级 Vue Flow 大版本、不换库、不改编排协议**

### 5.6 监控中心
- 指标卡 + **`.cf-data-table`** 日志；筛选用 `USelect` / `UInput`
- 状态码、耗时：mono
- **双击行**打开日志详情；详情无底栏 Close，用右上角 X

### 5.7 设置 / 管理
- 设置侧栏英文分区标题（Profile / Security / Preferences / Sessions）+ `.cf-section-title`
- 复杂表单：优先 **`UForm` + `UFormField`** 或 `UInput`/`CfButton`（§2.8）
- 管理 CRUD：`.cf-data-table` + `UModal`；**Users / Roles 双击行进入编辑**

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
| Button / Input | `UButton` `UInput` | 有 | **`CfButton` + `UInput`** | 业务按钮走 §2.8；底层仍 Nuxt UI |
| Select（可搜索） | `USelectMenu` | 基础 Select | **Nuxt UI** | 长列表 |
| Form | `UForm` | 自搭 | **Nuxt UI** | |
| Table / Pagination | `UTable` | 基础 Table | **优先 `.cf-data-table`；复杂表再用 UTable** | editorial 列表已落地原生表 |
| Overlay / Toast | `UModal` `UToast`… | Dialog/sonner | **Nuxt UI** | 底栏布局见 §2.8.2 |
| AI Chat | `UChat*` | 无对等 | **Nuxt UI** | 已定，左右分侧 |
| FileUpload | `UFileUpload` | 自研 | **Nuxt UI** | |
| Command palette | `UCommandPalette` | 需拼 | **Nuxt UI** | |
| shadcn 整包 | — | — | **默认不作为主栈** | 减少双库；确有缺口再单点引入 |
| 工作流画布 | 无 | 无 | **Vue Flow** | |

### 6.3 按页面落库（实施对照）

| 页面 | Nuxt UI | 自研 / 备注 |
|------|---------|-------------|
| 全局壳 | Header（editorial/island）、Slideover、Toast、UApp | BrandMark / 自研导航链 |
| Login / Register | Input、`CfButton` 提交、文字 link 仍 `UButton` | 大标题品牌区自研；文案英文 |
| Dashboard | Button、统计区 | `.cf-page` 欢迎 + next steps |
| Playground | **Chat* 左右分侧**、Select、Slider | 元数据小条；空态英文 |
| Agents / Keys / Flows / Admin | `.cf-data-table`、Modal、Badge、行内 `CfButton` | §2.8 |
| Models | 同上 + **卡片/列表切换**、详情只读同弹窗 | `cf-models-view` |
| Knowledge | 卡片网格 / Modal、FileUpload | |
| Workflows 编辑器 | Toast、Slideover；画布 Vue Flow | 节点壳换皮 |
| Settings | Form 控件、章节 `.cf-section-title` | 侧栏 Settings |

### 6.4 从 Naive 迁移时的替换方向

| 现用（Naive） | 目标 |
|---------------|------|
| `n-layout` / `n-menu` 顶栏 | `UHeader` + `UNavigationMenu` |
| `n-button` / `n-input` | `CfButton` / `UInput` |
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
- 登录注册提交用 `CfButton`；页脚文字 link 仍可用 `UButton variant="link"`
- 默认布局切换到 Nuxt UI Header 导航

### Phase 2 — 列表型页面（Desktop 优先）
- Dashboard / Agents / Models / Knowledge / Keys / Monitor / Admin
- 列表统一 `.cf-data-table`（复杂场景可再用 `UTable`）；上传用 `UFileUpload`
- 按钮语义统一 **§2.8**

### Phase 3 — 复杂页（Desktop 优先）
- Playground（**Nuxt UI Chat* + 适配层**）
- Settings（`UForm`）
- Workflows 壳层（画布逻辑冻结）
- 可选：`UCommandPalette`（⌘K）

### Phase 4 — 清理
- 删除 Naive/Element 依赖与 `plugins/naive-ui.client.ts`
- 更新 `cogniforge-web/README.md` 技术栈说明
- 回归 composable 单测

### Phase 7 — editorial 对齐 + Models（已完成）
- 共享 `.cf-page` chrome；英文模块标题
- Models：卡片默认、列表切换、双击只读详情、同弹窗编辑；按钮规范入库 §2.8
- 业务页 `CfButton`；Users/Roles/Monitor 双击；Keys 显示/复制
- 登录/注册提交 `CfButton`、文案英文；测试弹窗无多余 Close

### Phase 8 — 中英文（已完成）
- `useLocale` + `i18n/messages.ts`；仅 zh-CN / en-US
- 顶栏 / 登录 / Preferences 切换；`localStorage` + 现有 settings `language`
- **不改路由前缀**

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
- 不重做产品信息架构（**导航模块/路由/角色可见性冻结，见 §4.1**）
- 不为「更炫」加 3D/粒子/大英雄营销首屏
- **不做**独立原生 App；手机以响应式 Web 为准
- **不承诺**手机端完整工作流可视化编排（P2 降级即可）

---

## 9. 验收标准（给不懂代码的同学）

### 9.1 Web 桌面（必须过，否则不算完成）
1. 打开登录页，第一眼看到的是 **CogniForge** 品牌，而不是普通灰表单。
2. 登录后顶栏像产品，但**以前的模块都能找到**（控制台、Playground、Agent、模型、工作流、知识库、密钥；admin 另有监控/用户/角色）。
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
