# Oblivious

> 一个面向 C 端用户的 AI 应用服务平台，同时保留 B 端 API 中转能力。

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24.10+-00ADD8?logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-15-000000?logo=next.js)](https://nextjs.org/)

## ✨ 特性

### C 端功能（面向普通用户）

- 🤖 **智能对话**：支持 GPT-4、Claude、Gemini 等主流大模型
- 👤 **AI 助手**：丰富的助手市场，一键安装使用
- 📚 **知识库**：RAG 技术，上传文档获得专属 AI 助理
- 🔌 **插件系统**：联网搜索、代码执行、图片生成等扩展功能
- 🎨 **精美界面**：现代化设计，支持深色模式
- 📱 **多端支持**：Web、桌面端、移动端（规划中）

### B 端功能（面向开发者）

- 🔄 **API 中转**：统一接口对接多家 AI 提供商
- 💰 **计费管理**：按量计费，支持额度充值
- 📊 **数据统计**：实时监控 API 调用情况
- 🔐 **权限管理**：多用户、多渠道管理
- ⚖️ **负载均衡**：智能选择最优渠道

### 技术特性

- 🏗️ **微服务架构**：服务解耦，独立扩容
- ☸️ **云原生部署**：支持 Kubernetes 水平扩展
- 🚀 **高性能**：Go 语言构建，支持万级 QPS
- 🔒 **安全可靠**：数据加密、限流熔断、日志追踪
- 📈 **可观测性**：完整的监控告警体系

---

## 📖 文档

- **快速开始**: [快速启动指南](docs/QUICK_START.md) - 本地开发环境搭建
- **系统设计**: 
  - [架构设计](docs/ARCHITECTURE.md) - 系统整体架构说明
  - [API 网关设计](docs/API_GATEWAY_DESIGN.md) - 网关详细设计
  - [数据库设计](docs/DATABASE_DESIGN.md) - 数据库表结构和优化策略
  - [项目结构](docs/PROJECT_STRUCTURE.md) - 目录结构说明
- **API 文档**: [API 参考](docs/API_REFERENCE.md) - RESTful API 接口文档
- **部署运维**:
  - [快速部署参考](docs/DEPLOYMENT_QUICK_REFERENCE.md) - 部署命令速查
  - [生产部署指南](docs/PRODUCTION_DEPLOYMENT_GUIDE.md) - 完整生产部署流程
- **开发指南**:
  - [贡献指南](docs/CONTRIBUTING.md) - 如何参与项目开发
  - [服务开发状态](docs/SERVICE_STATUS.md) - 各服务实现进度
- **专题文档**:
  - [计费系统指南](docs/BILLING_SYSTEM_GUIDE.md) - 高级计费功能说明
  - [AI 适配器设置](docs/AI_ADAPTER_SETUP.md) - AI 提供商集成

---

## 🚀 快速开始

### 前置条件

- Docker & Docker Compose (v2)
- Go 1.24.10 或更高版本
- Node.js 20+
- PostgreSQL 15+
- Redis 7+
- Make 工具（可选，用于简化命令）

### 本地开发

1. **克隆仓库**

```bash
git clone https://github.com/your-org/oblivious.git
cd oblivious
```

2. **启动基础设施**

```bash
cd deploy
docker-compose up -d postgres redis minio rabbitmq
```

3. **运行数据库迁移**

```bash
cd backend
make migrate-up
```

4. **启动后端服务**

```bash
# 启动 API 网关
cd backend/cmd/gateway
go run main.go

# 启动其他服务（新终端）
cd backend/cmd/user && go run main.go
cd backend/cmd/chat && go run main.go
cd backend/cmd/relay && go run main.go
```

5. **启动前端**

```bash
cd frontend
npm install
npm run dev
```

6. **访问应用**

打开浏览器访问 `http://localhost:3000`

---

## 🏗️ 架构概览

```
┌─────────────┐
│   浏览器    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────┐     ┌─────────────┐
│   前端静态  │     │  API 网关   │
│  (Next.js)  │────▶│  (Gateway)  │
└─────────────┘     └──────┬──────┘
                           │
      ┌────────────────────┼────────────────────┐
      ▼                    ▼                    ▼
┌──────────┐         ┌──────────┐        ┌──────────┐
│ 用户服务 │         │ 对话服务 │        │ 中转服务 │
└──────────┘         └──────────┘        └──────────┘
      │                    │                    │
      └────────────────────┼────────────────────┘
                           ▼
                    ┌──────────────┐
                    │  PostgreSQL  │
                    └──────────────┘
```

**核心服务**：

- **前端服务**：基于 Next.js 的静态资源，提供用户界面
- **API 网关**：统一入口，负责鉴权、限流、路由
- **用户服务**：用户管理、额度管理
- **对话服务**：聊天会话、消息存储
- **中转服务**：对接上游 AI 提供商
- **助手服务**：AI 助手管理
- **知识库服务**：RAG 功能
- **文件服务**：文件上传和存储
- **插件服务**：插件市场和调用
- **计费服务**：按量计费和账单

详见 [架构设计文档](docs/ARCHITECTURE.md)

---

## 🛠️ 技术栈

### 前端

- **框架**：React 19 + Next.js 15
- **语言**：TypeScript
- **状态管理**：Zustand
- **UI 组件**：antd / @lobehub/ui
- **样式**：TailwindCSS / antd-style
- **HTTP 客户端**：axios

### 后端

- **语言**：Go 1.24.10+
- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：PostgreSQL 15 (主库) + pgvector (向量检索)
- **缓存**：Redis Cluster
- **消息队列**：RabbitMQ
- **对象存储**：MinIO / S3

### DevOps

- **容器化**：Docker
- **编排**：Kubernetes
- **CI/CD**：GitHub Actions
- **监控**：Prometheus + Grafana
- **日志**：Loki
- **链路追踪**：Jaeger (OpenTelemetry)

---

## 📦 部署

### Docker Compose (本地/小规模)

```bash
cd deploy
docker-compose up -d
```

### Kubernetes (生产环境)

```bash
# 创建命名空间
kubectl apply -f deploy/k8s/namespace.yaml

# 应用配置
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secrets.yaml

# 部署服务
kubectl apply -f deploy/k8s/deployments/
kubectl apply -f deploy/k8s/services/
kubectl apply -f deploy/k8s/ingress.yaml

# 配置自动扩缩容
kubectl apply -f deploy/k8s/hpa/
```

详见 [快速部署参考](docs/DEPLOYMENT_QUICK_REFERENCE.md) 和 [生产部署指南](docs/PRODUCTION_DEPLOYMENT_GUIDE.md)

---

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

详见 [贡献指南](docs/CONTRIBUTING.md)

---

## 📄 开源协议

本项目基于 [MIT 协议](LICENSE) 开源。

---

## 🙏 致谢

本项目灵感来源于以下优秀开源项目：

- [NewAPI](https://github.com/Calcium-Ion/new-api) - API 中转和渠道管理
- [LobeChat](https://github.com/lobehub/lobe-chat) - 前端 UI/UX 设计参考

---

## 📧 联系方式

- 问题反馈：[GitHub Issues](https://github.com/your-org/oblivious/issues)
- 邮箱：support@oblivious.ai
- 社区讨论：[Discord](https://discord.gg/oblivious)

---

<p align="center">
  Made with ❤️ by the Oblivious Team
</p>

