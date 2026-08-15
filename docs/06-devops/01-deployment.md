# CogniForge 部署说明

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
- **不要从** `/opt/project/cogniforge-ai/.env` **删** `OPENAI_API_KEY`（知识库 embedding 仍可能用）

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
