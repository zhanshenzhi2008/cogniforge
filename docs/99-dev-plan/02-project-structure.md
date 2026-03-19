# CogniForge 项目结构规划

## 1. 整体架构

```
cogniforge/
├── web/                        # 前端 Web 应用 (Vue + Nuxt)
├── gateway/                    # API 网关 (Go)
├── user/                       # 用户中心服务 (Java)
├── agent/                      # Agent 引擎服务 (Go)
├── model/                      # 模型网关服务 (Go)
├── workflow/                   # 工作流编排服务 (Go)
├── knowledge/                  # 知识库服务 (Go + Python)
├── billing/                    # 计费中心服务 (Java)
├── monitor/                    # 监控服务 (Go)
├── ml/                         # ML 处理服务 (Python)
├── deploy/                     # 部署配置
└── docs/                       # 文档
```

---

## 2. 服务命名规范

### 2.1 服务名格式
```
功能名
```

| 服务名 | 功能 | 语言 | 端口 |
|-------|------|------|------|
| web | 前端 Web | Vue/Nuxt | 3000 |
| gateway | API 网关 | Go | 8080 |
| user | 用户中心 | Java | 8085 |
| agent | Agent 引擎 | Go | 8082 |
| model | 模型网关 | Go | 8081 |
| workflow | 工作流编排 | Go | 8083 |
| knowledge | 知识库 | Go | 8084 |
| billing | 计费中心 | Java | 8086 |
| monitor | 监控服务 | Go | 8087 |
| ml | ML 处理 | Python | 8088 |

### 2.2 Java 包名格式
```
com.orjrs.服务名简写.模块名
```

| 服务 | 包名示例 |
|------|---------|
| user | com.orjrs.usr.user, com.orjrs.usr.auth |
| billing | com.orjrs.bil.billing, com.orjrs.bil.invoice |

**服务名简写对照表**：

| 服务 | 简写 |
|------|------|
| user | usr |
| agent | agt |
| model | mdl |
| workflow | wf |
| knowledge | kb |
| billing | bil |
| monitor | mon |
| ml | ml |

### 2.3 Go 包名格式
```
com/orjrs/服务名简写/模块名
```

| 服务 | 包名示例 |
|------|---------|
| gateway | com/orjrs/gw/handler, com/orjrs/gw/middleware |
| agent | com/orjrs/agt/handler, com/orjrs/agt/engine |

---

## 3. 各服务项目结构

### 3.1 Go 服务结构 (推荐)

```
user/
├── cmd/                          # 入口目录
│   └── server/
│       └── main.go              # 程序入口
├── pkg/                          # 业务代码
│   └── orjrs/
│       └── usr/                 # 用户服务
│           ├── handler/         # HTTP 处理层
│           ├── service/         # 业务逻辑层
│           ├── repo/             # 数据访问层
│           ├── model/            # 数据模型
│           ├── dto/              # 数据传输对象
│           └── middleware/       # 中间件
├── configs/                      # 配置文件
├── migrations/                   # 数据库迁移
├── api/                         # API proto 定义
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```

### 3.2 Java 服务结构 (Spring Boot)

```
billing/
├── src/main/java/com/orjrs/bil/
│   ├── BillingApplication.java
│   ├── controller/              # 控制器
│   ├── service/                 # 业务逻辑
│   ├── repository/              # 数据访问
│   ├── model/                   # 实体类
│   ├── dto/                     # 数据传输对象
│   ├── config/                  # 配置
│   └── security/                # 安全相关
├── src/main/resources/
│   ├── application.yml
│   └── mapper/
├── src/test/java/
├── pom.xml
├── Dockerfile
├── Makefile
└── README.md
```

### 3.3 Python 服务结构 (FastAPI)

```
ml/
├── src/
│   └── orjrs/
│       └── ml/
│           ├── api/              # API 路由
│           ├── core/             # 核心逻辑
│           ├── models/            # 数据模型
│           ├── services/         # 业务服务
│           └── utils/            # 工具函数
├── configs/                      # 配置
├── tests/                        # 测试
├── pyproject.toml
├── Dockerfile
├── Makefile
└── README.md
```

### 3.4 前端结构 (Vue + Nuxt)

```
web/
├── src/
│   ├── assets/                   # 资源文件
│   ├── components/               # 组件
│   │   ├── ui/                  # 基础 UI
│   │   └── features/            # 功能组件
│   ├── composables/             # 组合式函数
│   ├── layouts/                 # 布局
│   ├── pages/                   # 页面
│   ├── stores/                  # Pinia 状态
│   ├── types/                   # TypeScript 类型
│   └── utils/                   # 工具函数
├── server/                      # Nuxt 服务端
│   └── api/                     # API 代理
├── public/                      # 静态资源
├── nuxt.config.ts
├── tailwind.config.ts
├── package.json
├── Dockerfile
├── Makefile
└── README.md
```

---

## 4. 打包构建配置

### 4.1 Makefile 模板 (Go)

```makefile
# Makefile for user service

.PHONY: build run test docker-build docker-run clean lint fmt

BINARY_NAME=user
DOCKER_IMAGE=orjrs/user
VERSION=$(shell git describe --tags --always --dirty)

# 构建
build:
	go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/${BINARY_NAME} ./cmd/server

# 运行
run: build
	./bin/${BINARY_NAME}

# 测试
test:
	go test -v -coverprofile coverage.out ./...

# 代码检查
lint:
	golangci-lint run

# 代码格式化
fmt:
	go fmt ./...
	go mod tidy

# Docker 构建
docker-build:
	docker build -t ${DOCKER_IMAGE}:${VERSION} .
	docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest

# Docker 运行
docker-run:
	docker run -p 8081:8081 ${DOCKER_IMAGE}:latest

# 清理
clean:
	rm -rf bin/
	rm -f coverage.out

# 帮助
help:
	@echo "Available targets: build, run, test, lint, fmt, docker-build, docker-run, clean"
```

### 4.2 Makefile 模板 (Java)

```makefile
# Makefile for billing service

.PHONY: build test docker-build docker-run clean package

PROJECT_VERSION=1.0.0
DOCKER_IMAGE=orjrs/billing

# Maven 包装器
MW=./mvnw

# 构建
build:
	${MW} clean package -DskipTests

# 测试
test:
	${MW} test

# Docker 构建
docker-build:
	docker build -t ${DOCKER_IMAGE}:${PROJECT_VERSION} .
	docker tag ${DOCKER_IMAGE}:${PROJECT_VERSION} ${DOCKER_IMAGE}:latest

# Docker 运行
docker-run:
	docker run -p 8086:8086 ${DOCKER_IMAGE}:latest

# 清理
clean:
	${MW} clean

# 帮助
help:
	@echo "Available targets: build, test, docker-build, docker-run, clean"
```

### 4.3 Makefile 模板 (Python)

```makefile
# Makefile for ml service

.PHONY: install test docker-build docker-run clean lint format

DOCKER_IMAGE=orjrs/ml
VERSION=$(shell git describe --tags --always --dirty)

# 安装依赖
install:
	poetry install

# 测试
test:
	poetry run pytest -v --cov=src

# 代码检查
lint:
	poetry run ruff check src/

# 代码格式化
format:
	poetry run ruff format src/

# Docker 构建
docker-build:
	docker build -t ${DOCKER_IMAGE}:${VERSION} .
	docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest

# Docker 运行
docker-run:
	docker run -p 8088:8088 ${DOCKER_IMAGE}:latest

# 清理
clean:
	rm -rf .pytest_cache
	rm -rf htmlcov
	rm -rf dist/

# 帮助
help:
	@echo "Available targets: install, test, lint, format, docker-build, docker-run, clean"
```

### 4.4 Makefile 模板 (前端)

```makefile
# Makefile for web

.PHONY: install dev build test lint format docker-build docker-run clean

DOCKER_IMAGE=orjrs/web

# 安装依赖
install:
	pnpm install

# 开发
dev:
	pnpm dev

# 构建
build:
	pnpm build

# 测试
test:
	pnpm test

# 代码检查
lint:
	pnpm lint

# 代码格式化
format:
	pnpm format

# Docker 构建
docker-build:
	docker build -t ${DOCKER_IMAGE} .

# Docker 运行
docker-run:
	docker run -p 3000:3000 ${DOCKER_IMAGE}

# 清理
clean:
	rm -rf .nuxt
	rm -rf output
	rm -rf node_modules

# 帮助
help:
	@echo "Available targets: install, dev, build, test, lint, format, docker-build, docker-run, clean"
```

---

## 5. 统一脚本入口

### 5.1 根目录 Makefile

```makefile
# CogniForge 统一构建脚本

.PHONY: help install test docker-build docker-up docker-down clean

# 显示帮助
help:
	@echo "CogniForge Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Global targets:"
	@echo "  install       - Install all dependencies"
	@echo "  test          - Run all tests"
	@echo "  docker-build  - Build all Docker images"
	@echo "  docker-up     - Start all services"
	@echo "  docker-down   - Stop all services"
	@echo "  clean         - Clean all build artifacts"
	@echo ""
	@echo "Service targets (run from service directory):"
	@echo "  cd user && make build"
	@echo "  cd agent && make build"
	@echo "  cd web && make build"

# 安装所有依赖
install:
	cd web && pnpm install
	cd user && go mod download
	cd agent && go mod download

# Docker Compose 构建
docker-build:
	docker-compose build

# 启动所有服务
docker-up:
	docker-compose up -d

# 停止所有服务
docker-down:
	docker-compose down

# 清理
clean:
	rm -rf web/.nuxt web/node_modules
	rm -rf user/bin
	rm -rf agent/bin
```

---

## 6. Docker Compose 本地开发

```yaml
# docker-compose.yml

services:
  # 前端
  web:
    build: ./web
    ports:
      - "3000:3000"
    volumes:
      - ./web/src:/app/src
    environment:
      - API_BASE=http://gateway:8080
    command: pnpm dev

  # API 网关
  gateway:
    build: ./gateway
    ports:
      - "8080:8080"
    environment:
      - USER_SERVICE=user:8085
      - AGENT_SERVICE=agent:8082

  # 用户服务 (Java)
  user:
    build: ./user
    ports:
      - "8085:8085"
    environment:
      - DATABASE_URL=postgres://postgres:5432/cogniforge
      - REDIS_URL=redis://redis:6379

  # Agent 服务 (Go)
  agent:
    build: ./agent
    ports:
      - "8082:8082"

  # 数据库
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: cogniforge
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # 缓存
  redis:
    image: redis:7
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

---

## 7. 目录结构总览

```
cogniforge/
├── Makefile                      # 统一入口
├── docker-compose.yml            # 本地开发
├── .env.example                  # 环境变量示例
│
├── web/                        # 前端 (Nuxt)
│   ├── src/
│   ├── Makefile
│   ├── Dockerfile
│   └── package.json
│
├── gateway/                    # API 网关 (Go)
│   ├── cmd/
│   ├── pkg/orjrs/gw/
│   ├── Makefile
│   ├── Dockerfile
│   └── go.mod
│
├── user/                       # 用户服务 (Java Spring Boot)
│   ├── src/main/java/com/orjrs/usr/
│   ├── src/main/resources/
│   ├── Makefile
│   ├── Dockerfile
│   └── pom.xml
│
├── agent/                      # Agent 服务 (Go)
│   ├── cmd/
│   ├── pkg/orjrs/agt/
│   ├── Makefile
│   ├── Dockerfile
│   └── go.mod
│
├── model/                      # 模型网关 (Go)
├── workflow/                   # 工作流编排 (Go)
├── knowledge/                   # 知识库 (Go+Python)
├── billing/                    # 计费服务 (Java Spring Boot)
│   ├── src/main/java/com/orjrs/bil/
│   ├── src/main/resources/
│   ├── Makefile
│   ├── Dockerfile
│   └── pom.xml
│
├── monitor/                    # 监控服务 (Go)
├── ml/                         # ML 服务 (Python)
│   ├── src/orjrs/ml/
│   ├── Makefile
│   ├── Dockerfile
│   └── pyproject.toml
│
└── deploy/                      # 部署配置
    ├── kubernetes/
    └── helm/
```

---

## 8. 开发流程

### 8.1 本地开发
```bash
# 1. 克隆项目
git clone https://github.com/orjrs/cogniforge.git
cd cogniforge

# 2. 安装依赖
make install

# 3. 启动服务
make docker-up

# 4. 访问
# 前端: http://localhost:3000
# API: http://localhost:8080
```

### 8.2 单服务开发
```bash
# 以用户服务为例
cd user

# 运行测试
make test

# 构建
make build

# Docker 构建
make docker-build

# 运行
make docker-run
```

---

**项目结构规划完成，是否需要我开始创建具体的项目代码？**
