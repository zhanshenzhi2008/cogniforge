# CogniForge AI Platform

<p align="center">
  <img src="https://via.placeholder.com/150x50?text=CogniForge" alt="CogniForge Logo" />
</p>

<p align="center">
  企业级AI应用开发与运营平台 | Enterprise AI Application Development & Operations Platform
</p>

<p align="center">
  <a href="#快速开始">快速开始</a> •
  <a href="#核心功能">核心功能</a> •
  <a href="#文档">文档</a> •
  <a href="#贡献">贡献</a> •
  <a href="#许可证">许可证</a>
</p>

---

## 简介

CogniForge 是一个企业级AI应用开发与运营平台，旨在为开发者提供一个统一、可扩展、安全的环境，用于构建、部署和管理AI应用。平台支持大语言模型、多模态模型、Agent工作流编排以及企业级治理能力。

### 核心特性

- **🤖 多模型支持** - 统一接入OpenAI、Anthropic、Google等主流AI供应商
- **🔧 Agent引擎** - 构建具有自主推理和工具使用能力的AI Agent
- **🔄 工作流编排** - 可视化拖拽设计复杂AI业务流程
- **📚 知识库服务** - 企业文档向量化存储与语义检索
- **📊 监控中心** - 全面的可观测性与成本分析
- **🔒 企业级安全** - RBAC访问控制、数据隔离、审计日志

---

## 快速开始

### 环境要求

| 组件 | 最低版本 |
|-----|---------|
| Go | 1.22+ |
| Java | 21+ |
| Python | 3.11+ |
| Node.js | 20+ |
| PostgreSQL | 15+ |
| Redis | 7+ |
| Docker | 24+ |
| Kubernetes | 1.28+ |

### 本地开发

```bash
# 克隆项目
git clone https://github.com/cogniforge/cogniforge.git
cd cogniforge

# 安装依赖
npm install

# 启动开发服务
npm run dev

# 运行测试
npm run test
```

### Docker Compose 启动

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

---

## 核心功能

### 1. 开发者控制台 (Developer Console)

统一的Web控制台，提供交互式Playground、API密钥管理、请求日志和用量统计。

### 2. 模型网关 (Model Gateway)

统一API接口，适配多供应商模型，支持负载均衡和故障转移。

### 3. Agent引擎 (Agent Engine)

构建具有推理能力、工具使用和多轮对话的AI Agent。

### 4. 工作流编排 (Workflow Orchestration)

可视化拖拽设计器，支持条件分支、并行执行、循环等高级控制流。

### 5. 知识库服务 (Knowledge Base)

企业文档向量化存储，自然语言语义检索，源引用追溯。

### 6. 微调训练 (Fine-tuning)

定制模型训练、评估和部署，构建企业专属模型。

### 7. 监控中心 (Monitoring Center)

实时监控、性能分析、成本优化、告警通知。

### 8. 安全与合规 (Security & Compliance)

RBAC权限管理、数据隔离、审计日志、合规报告。

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CogniForge AI Platform                       │
├─────────────────────────────────────────────────────────────────────┤
│  开发者控制台 ── API网关 ── 核心服务 ── 基础设施 ── 外部模型        │
└─────────────────────────────────────────────────────────────────────┘
```

详细架构设计请参阅 [技术架构设计](./docs/02-architecture/01-technical-architecture.md)。

---

## 文档

项目文档位于 `docs/` 目录下：

### 需求文档 (docs/01-requirements/)
- [产品需求文档](./docs/01-requirements/01-product-requirements.md) - 完整的产品功能需求

### 技术架构 (docs/02-architecture/)
- [技术架构设计](./docs/02-architecture/01-technical-architecture.md) - 多语言混合架构设计

### API设计 (docs/03-api/)
- [API接口设计](./docs/03-api/01-api-design.md) - REST API接口规范

### 数据库设计 (docs/04-database/)
- [数据库设计](./docs/04-database/01-database-design.md) - 数据库表结构设计

### 前端设计 (docs/05-frontend/)
- [前端设计文档](./docs/05-frontend/01-frontend-design.md) - 前端技术栈与组件设计

### DevOps (docs/06-devops/)
- 部署配置与CI/CD流程（待完善）

---

## 技术栈

采用多语言混合架构，根据服务特性选择最优技术栈：

| 服务类型 | 语言 | 框架 | 选型理由 |
|---------|------|------|---------|
| **API网关** | Go | Gin | 高性能、低延迟、高并发 |
| **模型网关** | Go | Gin+gRPC | AI请求处理、低延迟 |
| **Agent引擎** | Go | Gin+gRPC | 实时推理、低延迟 |
| **工作流编排** | Go | Gin+gRPC | 高并发、执行调度 |
| **知识库服务** | Go+Python | Gin+FastAPI | Go处理API, Python处理ML |
| **用户中心** | Java | Spring Boot | 成熟稳定、事务处理 |
| **计费中心** | Java | Spring Boot | 业务复杂、需要事务 |
| **监控服务** | Go | Gin | 高性能指标处理 |
| **ML处理** | Python | FastAPI | AI/ML事实标准 |

### 基础设施

| 组件 | 技术 |
|-----|------|
| 前端 | React 18, TypeScript, Next.js, Tailwind CSS |
| 数据库 | PostgreSQL 15+, Redis 7+ |
| 向量库 | Milvus/Qdrant |
| 消息队列 | Kafka 3.6+ |
| 容器编排 | Kubernetes 1.28+ |
| 监控 | Prometheus, Grafana, Loki |
| 对象存储 | S3/MinIO |
| 分析 | ClickHouse |

---

## 路线图

### v1.0 (当前版本)

- [x] 开发者控制台
- [x] 模型网关
- [x] Agent引擎核心
- [x] 监控中心
- [x] 安全基础功能

### v1.1 (规划中)

- [ ] 工作流编排可视化编辑器
- [ ] 知识库服务

### v1.2 (规划中)

- [ ] 微调训练功能
- [ ] 告警系统

### v2.0 (规划中)

- [ ] 多Agent协作
- [ ] 企业级合规功能
- [ ] 插件市场

---

## 贡献指南

我们欢迎社区贡献！请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解如何参与贡献。

```bash
# 创建特性分支
git checkout -b feature/amazing-feature

# 提交更改
git commit -m 'Add amazing feature'

# 推送分支
git push origin feature/amazing-feature

# 发起Pull Request
```

---

## 支持

- 📖 [文档](https://docs.cogniforge.ai)
- 💬 [社区论坛](https://community.cogniforge.ai)
- 🐛 [问题反馈](https://github.com/cogniforge/cogniforge/issues)
- 📧 [邮箱](mailto:support@cogniforge.ai)

---

## 许可证

本项目基于 MIT 许可证开源。详细信息请参阅 [LICENSE](./LICENSE)。

---

<p align="center">
  © 2026 CogniForge. All rights reserved.
</p>
