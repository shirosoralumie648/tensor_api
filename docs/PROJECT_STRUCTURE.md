# Oblivious 项目目录结构

```
oblivious/
├── frontend/                           # 前端服务（借鉴 LobeChat）
│   ├── public/
│   │   ├── favicon.ico
│   │   ├── logo.png
│   │   └── manifest.json
│   ├── src/
│   │   ├── app/                        # Next.js App Router
│   │   │   ├── (auth)/
│   │   │   │   ├── login/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── register/
│   │   │   │       └── page.tsx
│   │   │   ├── (main)/
│   │   │   │   ├── chat/
│   │   │   │   │   ├── [[...slug]]/
│   │   │   │   │   │   └── page.tsx
│   │   │   │   │   └── layout.tsx
│   │   │   │   ├── agents/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── knowledge/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── settings/
│   │   │   │       └── page.tsx
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── components/                 # 通用组件
│   │   │   ├── Chat/
│   │   │   │   ├── ChatInput/
│   │   │   │   ├── ChatMessage/
│   │   │   │   └── SessionList/
│   │   │   ├── Agent/
│   │   │   │   ├── AgentCard/
│   │   │   │   └── AgentMarket/
│   │   │   └── Layout/
│   │   │       ├── Header/
│   │   │       └── Sidebar/
│   │   ├── features/                   # 功能模块
│   │   │   ├── ChatInput/
│   │   │   ├── FileUpload/
│   │   │   └── PluginMarket/
│   │   ├── services/                   # API 客户端
│   │   │   ├── api.ts                  # 统一 API 客户端
│   │   │   ├── auth.ts
│   │   │   ├── chat.ts
│   │   │   ├── agent.ts
│   │   │   └── file.ts
│   │   ├── store/                      # Zustand 状态管理
│   │   │   ├── auth.ts
│   │   │   ├── chat.ts
│   │   │   ├── agent.ts
│   │   │   └── user.ts
│   │   ├── types/                      # TypeScript 类型定义
│   │   │   ├── message.ts
│   │   │   ├── session.ts
│   │   │   └── agent.ts
│   │   ├── utils/                      # 工具函数
│   │   │   ├── request.ts
│   │   │   └── format.ts
│   │   └── styles/                     # 全局样式
│   │       └── globals.css
│   ├── .env.example
│   ├── .eslintrc.json
│   ├── Dockerfile
│   ├── next.config.js
│   ├── package.json
│   ├── tsconfig.json
│   └── README.md
│
├── backend/                            # 后端服务（Go 微服务）
│   ├── cmd/                            # 各服务入口
│   │   ├── gateway/                    # API 网关
│   │   │   └── main.go
│   │   ├── user/                       # 用户服务
│   │   │   └── main.go
│   │   ├── chat/                       # 对话服务
│   │   │   └── main.go
│   │   ├── relay/                      # 中转服务
│   │   │   └── main.go
│   │   ├── agent/                      # 助手服务
│   │   │   └── main.go
│   │   ├── rag/                        # 知识库服务
│   │   │   └── main.go
│   │   ├── file/                       # 文件服务
│   │   │   └── main.go
│   │   ├── plugin/                     # 插件服务
│   │   │   └── main.go
│   │   ├── billing/                    # 计费服务
│   │   │   └── main.go
│   │   └── worker/                     # 异步任务处理
│   │       └── main.go
│   │
│   ├── internal/                       # 内部包（各服务共享）
│   │   ├── config/                     # 配置管理
│   │   │   └── config.go
│   │   ├── database/                   # 数据库连接
│   │   │   ├── postgres.go
│   │   │   └── redis.go
│   │   ├── middleware/                 # 中间件
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   └── rate_limit.go
│   │   ├── model/                      # 数据模型
│   │   │   ├── user.go
│   │   │   ├── session.go
│   │   │   ├── message.go
│   │   │   ├── agent.go
│   │   │   └── channel.go
│   │   ├── repository/                 # 数据访问层
│   │   │   ├── user_repo.go
│   │   │   ├── session_repo.go
│   │   │   └── message_repo.go
│   │   ├── service/                    # 业务逻辑层
│   │   │   ├── auth_service.go
│   │   │   ├── chat_service.go
│   │   │   └── billing_service.go
│   │   ├── utils/                      # 工具函数
│   │   │   ├── crypto.go               # 加密解密
│   │   │   ├── jwt.go                  # JWT 处理
│   │   │   ├── validation.go           # 参数验证
│   │   │   └── response.go             # 统一响应格式
│   │   └── constants/                  # 常量定义
│   │       └── errors.go
│   │
│   ├── pkg/                            # 公共库（可独立引用）
│   │   ├── logger/                     # 日志库
│   │   │   └── logger.go
│   │   ├── tracing/                    # 链路追踪
│   │   │   └── tracer.go
│   │   ├── metrics/                    # 监控指标
│   │   │   └── prometheus.go
│   │   └── queue/                      # 消息队列
│   │       ├── rabbitmq.go
│   │       └── producer.go
│   │
│   ├── relay/                          # 中转服务专用（复用 NewAPI）
│   │   ├── adapters/                   # 各 AI 提供商适配器
│   │   │   ├── openai/
│   │   │   │   └── adapter.go
│   │   │   ├── claude/
│   │   │   │   └── adapter.go
│   │   │   └── gemini/
│   │   │       └── adapter.go
│   │   ├── channel/                    # 渠道管理
│   │   │   ├── selector.go             # 渠道选择器
│   │   │   └── balancer.go             # 负载均衡
│   │   └── proxy/                      # 请求代理
│   │       └── proxy.go
│   │
│   ├── migrations/                     # 数据库迁移文件
│   │   ├── 000001_create_users_table.up.sql
│   │   ├── 000001_create_users_table.down.sql
│   │   └── ...
│   │
│   ├── scripts/                        # 脚本工具
│   │   ├── build.sh                    # 构建脚本
│   │   ├── deploy.sh                   # 部署脚本
│   │   └── migrate.sh                  # 数据库迁移
│   │
│   ├── tests/                          # 测试文件
│   │   ├── unit/
│   │   └── integration/
│   │
│   ├── go.mod
│   ├── go.sum
│   ├── .env.example
│   ├── Makefile
│   └── README.md
│
├── deploy/                             # 部署配置
│   ├── docker/                         # Docker 配置
│   │   ├── frontend.Dockerfile
│   │   ├── gateway.Dockerfile
│   │   ├── user.Dockerfile
│   │   ├── chat.Dockerfile
│   │   └── ...
│   ├── k8s/                            # Kubernetes 配置
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secrets.yaml
│   │   ├── deployments/
│   │   │   ├── frontend.yaml
│   │   │   ├── gateway.yaml
│   │   │   ├── user.yaml
│   │   │   ├── chat.yaml
│   │   │   └── ...
│   │   ├── services/
│   │   │   ├── frontend-service.yaml
│   │   │   ├── gateway-service.yaml
│   │   │   └── ...
│   │   ├── ingress.yaml
│   │   └── hpa/                        # 自动扩缩容
│   │       ├── chat-hpa.yaml
│   │       └── relay-hpa.yaml
│   ├── helm/                           # Helm Charts（可选）
│   │   └── oblivious/
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       └── templates/
│   └── docker-compose.yml              # 本地开发用
│
├── infra/                              # 基础设施配置
│   ├── terraform/                      # IaC 工具（可选）
│   │   ├── aws/
│   │   └── aliyun/
│   ├── monitoring/                     # 监控配置
│   │   ├── prometheus/
│   │   │   └── prometheus.yml
│   │   ├── grafana/
│   │   │   └── dashboards/
│   │   └── loki/
│   │       └── loki.yml
│   └── minio/                          # 对象存储配置
│       └── docker-compose.yml
│
├── docs/                               # 项目文档
│   ├── ARCHITECTURE.md                 # 架构设计
│   ├── API_GATEWAY_DESIGN.md           # 网关设计
│   ├── DATABASE_DESIGN.md              # 数据库设计
│   ├── API_REFERENCE.md                # API 文档
│   ├── DEPLOYMENT.md                   # 部署指南
│   └── CONTRIBUTING.md                 # 贡献指南
│
├── .github/                            # GitHub Actions CI/CD
│   └── workflows/
│       ├── frontend-ci.yml
│       ├── backend-ci.yml
│       └── deploy.yml
│
├── .gitignore
├── LICENSE
├── README.md
└── PROJECT_STRUCTURE.md                # 本文件
```

---

## 目录说明

### 1. `frontend/`

前端采用 **Next.js 15 (App Router)** 构建，仅作为静态资源提供 UI 界面。所有业务逻辑通过 API 网关与后端通信。

**关键文件**：
- `src/services/api.ts`: 封装所有 API 调用，统一处理鉴权、错误
- `src/store/`: Zustand 状态管理，保持客户端状态
- `Dockerfile`: 多阶段构建，最终基于 `nginx:alpine` 镜像

### 2. `backend/`

后端采用 **Go + Gin** 构建微服务架构。每个服务都是一个独立的 Go 程序。

**目录规范**：
- `cmd/`: 各服务的 `main.go` 入口
- `internal/`: 内部包（不对外暴露），包含共享的数据模型、中间件、工具等
- `pkg/`: 公共库（可独立引用），如日志、监控、消息队列封装
- `relay/`: 中转服务的专用逻辑（复用 NewAPI 的适配器层）

### 3. `deploy/`

包含所有部署相关的配置文件：

- `docker/`: 每个服务的 Dockerfile
- `k8s/`: Kubernetes YAML 配置（Deployment, Service, Ingress, HPA）
- `helm/`: Helm Charts（可选，用于批量部署）
- `docker-compose.yml`: 本地开发环境一键启动

### 4. `infra/`

基础设施配置：

- `terraform/`: 使用 Terraform 管理云资源（可选）
- `monitoring/`: Prometheus、Grafana、Loki 配置
- `minio/`: 对象存储配置

### 5. `docs/`

项目文档集合，包含架构设计、API 文档、部署指南等。

---

## 服务依赖关系

```
            ┌─────────────┐
            │   Frontend  │
            └──────┬──────┘
                   │ HTTPS
                   ▼
            ┌─────────────┐
            │   Gateway   │ ← 统一入口
            └──────┬──────┘
                   │
      ┌────────────┼────────────┐
      ▼            ▼            ▼
┌──────────┐ ┌──────────┐ ┌──────────┐
│   User   │ │   Chat   │ │  Relay   │
│ Service  │ │ Service  │ │ Service  │
└────┬─────┘ └────┬─────┘ └────┬─────┘
     │            │            │
     └────────────┼────────────┘
                  ▼
           ┌──────────────┐
           │  PostgreSQL  │
           └──────────────┘
```

---

## 服务端口分配

| 服务              | 内部端口 | 外部端口  | 说明              |
|-------------------|---------|----------|-------------------|
| Frontend (Nginx)  | 80      | 80/443   | 静态资源          |
| Gateway           | 8080    | 8080     | API 网关          |
| User Service      | 8081    | -        | 内部服务          |
| Chat Service      | 8082    | -        | 内部服务          |
| Relay Service     | 8083    | -        | 内部服务          |
| Agent Service     | 8084    | -        | 内部服务          |
| RAG Service       | 8085    | -        | 内部服务          |
| File Service      | 8086    | -        | 内部服务          |
| Plugin Service    | 8087    | -        | 内部服务          |
| Billing Service   | 8088    | -        | 内部服务          |
| Worker            | 8089    | -        | 异步任务          |
| Prometheus        | 9090    | 9090     | 监控指标          |

---

## 开发流程

### 本地开发

1. **启动基础设施**（PostgreSQL, Redis, MinIO, RabbitMQ）：
   ```bash
   cd deploy
   docker-compose up -d postgres redis minio rabbitmq
   ```

2. **运行数据库迁移**：
   ```bash
   cd backend
   make migrate-up
   ```

3. **启动后端服务**（以网关为例）：
   ```bash
   cd backend/cmd/gateway
   go run main.go
   ```

4. **启动前端**：
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

5. 访问 `http://localhost:3000`

### 构建 Docker 镜像

```bash
# 构建前端
docker build -f deploy/docker/frontend.Dockerfile -t oblivious-frontend:latest frontend/

# 构建网关
docker build -f deploy/docker/gateway.Dockerfile -t oblivious-gateway:latest backend/
```

### 部署到 Kubernetes

```bash
# 应用配置
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secrets.yaml

# 部署服务
kubectl apply -f deploy/k8s/deployments/
kubectl apply -f deploy/k8s/services/
kubectl apply -f deploy/k8s/ingress.yaml
kubectl apply -f deploy/k8s/hpa/
```

---

## 代码规范

### Go 代码规范

- 遵循 [Effective Go](https://go.dev/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查
- 包名小写，单词之间不使用下划线
- 接口命名以 `er` 结尾（如 `UserRepository`, `ChatService`）

### TypeScript 代码规范

- 遵循 Airbnb Style Guide
- 使用 ESLint + Prettier 格式化
- 优先使用 `interface` 而非 `type`
- 组件使用 PascalCase，文件名与组件名一致

### Git 提交规范

```
feat: 添加用户登录功能
fix: 修复消息发送失败的问题
docs: 更新 API 文档
style: 代码格式化
refactor: 重构中转服务的渠道选择逻辑
perf: 优化消息查询性能
test: 添加单元测试
chore: 更新依赖包版本
```

---

## 测试策略

1. **单元测试**：每个函数/方法覆盖率 > 80%
2. **集成测试**：测试服务间通信
3. **端到端测试**：使用 Playwright 测试前端关键流程
4. **压力测试**：使用 k6 模拟高并发场景

---

## 未来扩展

1. **gRPC 服务间通信**：提升性能
2. **服务网格 (Istio)**：流量管理和可观测性
3. **GraphQL API**：为前端提供更灵活的查询
4. **桌面端和移动端**：Electron + React Native

---

## 总结

整个项目采用清晰的模块化设计，前后端完全分离，后端微服务化，便于独立开发、测试和部署。通过合理的目录结构和规范，可以支撑大规模团队协作。

