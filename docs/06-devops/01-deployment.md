# CogniForge 部署说明

## [变更记录]
| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-10-17 | v1.35 | 容器默认 TZ=Asia/Shanghai，修 Portainer 日志少 8 小时 | orjrs |
| 2026-10-16 | v1.34 | cogniforge-ai Dockerfile：按 pyproject 装依赖，修 pip/hatch 构建失败 | orjrs |
| 2026-10-16 | v1.33 | GHCR 镜像扁平化：与仓名对齐（同 ai-job-hunter） | orjrs |
| 2026-10-16 | v1.32 | CD 登录 GHCR 改用 GITHUB_TOKEN（不再依赖 IMAGE_REGISTRY_TOKEN） | orjrs |
| 2026-10-15 | v1.31 | 补充 code_execute / Tavily 运维开关说明 | orjrs |
| 2026-10-14 | v1.30 | compose 各服务加 mem_limit + cpus（小机防 OOM/CPU 打满） | orjrs |
| 2026-08-16 | v1.29 | 前端 Docker 钉 pnpm@9，避免 ERR_PNPM_IGNORED_BUILDS | orjrs |
| 2026-08-15 | v1.28 | 模型配置走 Redis；cogniforge-ai 增加 REDIS_HOST | orjrs |
| 2026-08-15 | v1.27 | Python LLM 回调 Go；cogniforge-ai 增加 COGNIFORGE_API_URL | orjrs |

## [变更] 容器时区 Asia/Shanghai（2026-10-17）

- **变更原因**：Portainer 看 Gin 日志比北京时间少 8 小时（容器默认 UTC）
- **约定**：**Dockerfile** 装 `tzdata` + `ENV TZ=Asia/Shanghai`；compose / `.env` 默认不写（注释保留，仅临时覆盖）
- **包含代码**：三仓 `Dockerfile`；compose 中 `TZ` 行已注释；deploy skill §七·七 同步
- **生效**：须重新 build/push 镜像后再 `compose pull && up -d`

## [变更] cogniforge-ai Docker 构建修 pip install（2026-10-16）

- **变更原因**：`pip install .` 只 COPY 了 `pyproject.toml`，hatchling 报缺 `README.md`，且无 `cogniforge_ai` 包目录
- **包含代码**：`cogniforge-ai/Dockerfile`、`.dockerignore`（`!README.md`）、`pyproject.toml` hatch packages
- **变更前 vs 变更后**：~~`pip install .`（构建 wheel）~~ → 从 pyproject 抽出 dependencies 再 `pip install -r`；源码 `COPY .` 后由 uvicorn 直接加载

## [变更] GHCR 镜像扁平化（2026-10-16）

- **变更原因**：三段路径 `owner/cogniforge/cogniforge` 由 PAT 建包后，`GITHUB_TOKEN` 无法自动绑仓库（`read_package`）；对齐 ai-job-hunter 的两段路径
- **包含代码**：三仓 `.github/workflows/cd.yml` + 各 `docker-compose*.yml`
- **变更前 vs 变更后**：
  - ~~`ghcr.io/<owner>/cogniforge/cogniforge`~~ → `ghcr.io/<owner>/cogniforge`
  - ~~`ghcr.io/<owner>/cogniforge/cogniforge-ai`~~ → `ghcr.io/<owner>/cogniforge-ai`
  - ~~`ghcr.io/<owner>/cogniforge/cogniforge-web`~~ → `ghcr.io/<owner>/cogniforge-web`
- **服务器须人工**（CD 不覆盖 compose）：改 `/opt/project/cogniforge/` 下三个 compose 的 `image:` 行，再等 CD 推新包或手工 `pull` + `up`
- **旧包**：`cogniforge/cogniforge*` 可保留作历史，不再使用；新包首次由 Actions 推送时自动挂仓库

## [变更] CD 镜像仓库认证改用 GITHUB_TOKEN（2026-10-16）

- **变更原因**：`IMAGE_REGISTRY_TOKEN`（PAT）失效/权限不足导致 `ghcr.io` 登录 `denied: denied`；Actions 自带 `GITHUB_TOKEN` 更稳
- **包含代码**：三仓 `.github/workflows/cd.yml`（cogniforge / cogniforge-ai / cogniforge-web）
- **变更前 vs 变更后**：~~`password: secrets.IMAGE_REGISTRY_TOKEN`~~（2026-10-16）→ `password: secrets.GITHUB_TOKEN`；`username: github.actor`；job 声明 `permissions: packages: write`
- **Secret**：可删仓库/Environment `xcy` 里的 `IMAGE_REGISTRY_TOKEN`（不再使用）；`DEPLOY_*` 仍要保留
- **注意**：服务器若需手工 `docker pull`，仍要在机器上用个人 PAT 做过一次 `docker login ghcr.io`（与 CI 无关）

## [变更] MCP 工具运维开关（2026-10-15）

- **code_execute**：Go 进程内本地子进程执行（需镜像/宿主机有 `python3` / `node` / `go`）；生产若暂不开放设 `CODE_EXECUTE_DISABLED=1`
- **web_search**：读 `TAVILY_API_KEY`（空则 keyless）；见 `.env.example`
- **包含配置**：`.env.example` 注释段；详细设计 `docs/01-requirements/06-mcp-skill-design.md`

## [变更] compose 内存 / CPU 硬顶（2026-10-14）

- **变更原因**：同机多项目（尤其 Java）曾把 4G 小机打满导致登录卡死
- **包含代码**：`docker-compose.yml`（Go 后端）；`cogniforge-web/docker-compose-web.yml`、`cogniforge-ai/docker-compose-ai.yml` 同步
- **约定**：每个服务写 `mem_limit` + `cpus`；Go/Python/Nginx **不需要** Dockerfile 堆参数（仅 Java 用 `JAVA_OPTS` MaxRAMPercentage）
- **默认**：cogniforge `256m`/`0.5`；web `128m`/`0.25`；ai `512m`/`1.0`
- **服务器须人工**（CD 不覆盖 compose）：对齐仓库后 `docker compose up -d --force-recreate`，用 `docker stats` 看 LIMIT

## [变更] 前端 Docker 钉 pnpm@9（2026-08-16）

- **变更原因**：`cogniforge-web` CD 构建镜像时无版本安装 pnpm，拿到 10/11 后拦截依赖构建脚本，安装以 `ERR_PNPM_IGNORED_BUILDS` 退出 1
- **包含代码**：`cogniforge-web/Dockerfile`：`npm install -g pnpm@9 && pnpm install --frozen-lockfile`
- **影响范围**：仅 web 镜像构建阶段；运行层仍是 Nginx 静态托管
- **变更前 vs 变更后**：~~`npm install -g pnpm`（最新）~~（2026-08-16）→ `npm install -g pnpm@9`（与 CI `pnpm/action-setup` version 9 一致）
- **不要**：在 Dockerfile 里再改回无版本 `pnpm`；本机可用 9 或更新，但镜像必须钉 9

## [变更] 模型配置 Redis 缓存（2026-08-15）

- **cogniforge-ai `.env` 需要**：`COGNIFORGE_API_URL`、`PGSQL_*`、`REDIS_HOST`/`REDIS_PORT`/`REDIS_PWD`（与 Go 同一 Redis）
- Redis 连不上时两边都降级：Go 查库，Python 不带默认模型名（仍回调 Go）

## [变更] Python 只回调 Go 拿模型（2026-08-15）

- **变更原因**：模型配置只保留「模型」页一套；Python 不再解密 `ai_providers`
- **包含代码**：`cogniforge-ai/docker-compose-ai.yml`、`.env.example`
- **cogniforge-ai `.env` 需要**：
  - `COGNIFORGE_API_URL=http://cogniforge:8080`（本地开发用 `http://localhost:8080`）
  - `PGSQL_*`（pgvector 仍直连数据库）
- **cogniforge-ai `.env` 请删除**：`ENCRYPTION_KEY`（不再解 LLM Key）、`OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `OPENROUTER_API_KEY` / `DEFAULT_MODEL`
- **Go `.env` 不要删**：`ENCRYPTION_KEY`（Go 自己解密 `ai_providers`）、`JWT_SECRET`、`PGSQL_*`、`REDIS_*`、`DOMAIN`、`AI_SERVICE_URL`

删完后分别重启 `cogniforge` 与 `cogniforge-ai` 容器。GoLand 本地调试需 Restart Debug 才会加载新的 `/embeddings`。

## [变更] 删除 Go 后端 LLM 环境变量（2026-08-15）

- **变更原因**：聊天密钥与默认模型已在「模型」页写入 `ai_providers`，`AI_*` 环境变量已从代码移除
- **包含代码**：`configs/config.yaml`、`.env.example`
- **远程需人工删除**（`/opt/project/cogniforge/.env`，以及 Portainer Stack 环境若有）：
  - `AI_PROVIDER`
  - `AI_BASE_URL`
  - `AI_API_KEY`
  - `AI_DEFAULT_MODEL`
  - `OPENROUTER_HTTP_REFERER`
  - `OPENROUTER_TITLE`
  - 若误写在此文件：`OPENAI_API_KEY` / `ANTHROPIC_API_KEY` / `OPENROUTER_API_KEY`
- **不要删**：`ENCRYPTION_KEY`、`JWT_SECRET`、`PGSQL_*`、`REDIS_*`、`DOMAIN`、`AI_SERVICE_URL`（Python 服务地址）
- ~~**cogniforge-ai** 同样读 `ai_providers`。需要与 Go 相同的 `ENCRYPTION_KEY`~~（2026-08-15，改为只回调 Go）

删完后重新 `docker compose up -d` 或重启 cogniforge 容器。

## [变更] 修复 cogniforge-web healthcheck 导致 Traefik 过滤路由（2026-08-11）

### 变更原因 / 包含代码 / 影响范围
- 变更原因：Traefik v3 Docker provider 会过滤 `unhealthy`/`starting` 容器，日志为 `Filtering unhealthy or starting container`；`cogniforge-web` 健康检查探测错误端口/路径时长期 unhealthy，公网 Host 路由不注册，表现为 404。
- 包含代码：
  - `cogniforge-web/docker-compose-web.yml`：healthcheck 改为 `http://127.0.0.1:80/health`，与 Nginx 监听端口及 `nginx.conf` 精确 location 对齐。
- 影响范围：前端容器健康状态与 Traefik 自动发现；不改变业务页面逻辑。

### 变更前 vs 变更后
- 变更前：healthcheck 指向错误端口（如 3000）或带尾斜杠的 `/health/`，容器 unhealthy/starting → Traefik 跳过 → 域名 404。
- 变更后：healthcheck 探测 Nginx `80/health`，容器 healthy 后出现 `cogniforge-web@docker` 路由。

### 关键差异（新增 / 移除 / 修改）
- 修改：`cogniforge-web` compose healthcheck 命令与间隔参数。

## [变更] 复盘 cogniforge 404 根因并加固 healthcheck / BusyBox wget（2026-08-11）

### 变更原因 / 包含代码 / 影响范围
- 变更原因：公网 404 叠多层问题（Traefik 版本过旧、web healthcheck 使用 BusyBox 不支持参数导致 unhealthy、后端 DB/configs）。
- 包含代码：
  - `cogniforge-web/docker-compose-web.yml`：BusyBox 兼容 healthcheck（无 `--bind-address`）。
  - `cogniforge/Dockerfile` + `docker-compose.yml`：镜像内置 configs，去掉宿主机 configs 挂载。
  - `agent-insight` 前端 Dockerfile 与 docker-traefik compose：同步 BusyBox 兼容 healthcheck。
- 影响范围：Traefik 自动发现、前端 healthy 状态、域名路由注册。

### 关键差异（新增 / 移除 / 修改）
- 修改：healthcheck 仅使用 BusyBox wget 支持的参数。
- 文档：本文件与 deploy skill 补充 404 复盘要点。

## [变更] CI 测试服务版本对齐 agent-insight 数据库栈（2026-08-11）

### 变更原因 / 包含代码 / 影响范围
- 变更原因：为避免跨仓库环境版本差异导致的测试偏差，CI 中 PostgreSQL 与 Redis 版本需与 `agent-insight/docker-traefik/databases/docker-compose.yml` 保持一致。
- 包含代码：
  - `.github/workflows/ci.yml`：升级 `services.postgres.image` 与 `services.redis.image`。
- 影响范围：仅影响 CI 测试容器版本，不影响生产部署镜像与业务代码。

### 变更前 vs 变更后
- 变更前：
  - `services.postgres.image: postgres:16-alpine`
  - `services.redis.image: redis:7-alpine`
- 变更后：
  - `services.postgres.image: pgvector/pgvector:0.8.6-pg18-bookworm`
  - `services.redis.image: redis:8.8.0-alpine`

### 关键差异（新增 / 移除 / 修改）
- 修改：CI PostgreSQL 版本由 16 升级为 18，并切换到 `pgvector` 一体镜像。
- 修改：CI Redis 版本由 7 升级为 8.8.0。

## [变更] 修复 CI services.image 对 env 引用报错（2026-08-11）

### 变更原因 / 包含代码 / 影响范围
- 变更原因：GitHub Actions 在 `jobs.<job>.services.<name>.image` 位置不识别顶层 `env` 上下文，导致 CI 解析失败并报 `Unrecognized named-value: 'env'`。
- 包含代码：
  - `.github/workflows/ci.yml`：将 `services.postgres.image` 与 `services.redis.image` 从 `${{ env.* }}` 改为直接镜像值。
- 影响范围：仅影响 CI 工作流解析与启动阶段，不影响运行时业务代码逻辑。

### 变更前 vs 变更后
- 变更前：
  - `services.postgres.image: ${{ env.POSTGRES_IMAGE }}`
  - `services.redis.image: ${{ env.REDIS_IMAGE }}`
  - 工作流解析失败，CI 无法启动。
- 变更后：
  - `services.postgres.image: postgres:16-alpine`
  - `services.redis.image: redis:7-alpine`
  - 工作流可正常解析并执行测试任务。

### 关键差异（新增 / 移除 / 修改）
- 移除：顶层 `env` 中 `POSTGRES_IMAGE`、`REDIS_IMAGE` 配置项。
- 修改：`services.image` 改为直接镜像字符串，避免不支持的表达式上下文。

## [变更] 移除后端 configs 宿主机挂载（2026-08-11）

### 变更原因 / 包含代码 / 影响范围
- 变更原因：生产环境出现 `./configs` 目录存在但 `config.yaml` 缺失时，挂载会覆盖容器内配置目录，导致数据库配置读取为空并持续重启。
- 包含代码：
  - `Dockerfile`：运行层新增 `COPY --from=builder /app/configs /app/configs`。
  - `docker-compose.yml`：移除 `./configs:/app/configs:ro` 挂载。
- 影响范围：`cogniforge` 后端容器启动配置加载逻辑；部署流程从“依赖宿主机 configs”切换为“镜像内置配置 + .env 覆盖”。

### 变更前 vs 变更后
- 变更前：
  - 运行时依赖宿主机 `./configs` 目录。
  - 若宿主机目录为空会覆盖容器目录，导致 `/app/configs/config.yaml` 缺失。
- 变更后：
  - 镜像自带 `/app/configs/config.yaml`。
  - 运行时仅通过 `.env` 覆盖配置变量，不再要求挂载 `./configs`。

### 关键差异（新增 / 移除 / 修改）
- 新增：镜像内置 `configs` 目录。
- 移除：`docker-compose.yml` 中 `./configs:/app/configs:ro`。
- 修改：部署约束从“外置配置目录必需”调整为“`.env` 必需，外置 configs 非必需”。

## [变更记录]
| 日期 | 版本 | 变更摘要 | 负责人 |
|------|------|---------|--------|
| 2026-08-11 | v1.4 | 复盘 404 多层根因；BusyBox healthcheck 加固（web + skill + agent-insight） | Codex |
| 2026-08-11 | v1.3 | 修复 web healthcheck，避免 Traefik 因 unhealthy 过滤路由导致 404 | Codex |
| 2026-08-11 | v1.2 | CI 测试服务版本对齐为 PostgreSQL 18（pgvector）与 Redis 8.8.0 | Codex |
| 2026-08-11 | v1.1 | 修复 CI 在 services.image 中引用 env 导致的解析报错 | Codex |
| 2026-08-11 | v1.0 | 后端镜像内置 configs，移除宿主机 configs 挂载，避免空目录覆盖导致启动失败 | Codex |
