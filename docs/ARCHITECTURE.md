# Oblivious 项目架构设计

## 项目概述

**Oblivious** 是一个面向 C 端用户的 AI 应用服务平台，同时保留面向 B 端的 API 中转分发能力。项目基于 NewAPI 进行后端重构，借鉴 LobeChat 的前端实现，实现完全的前后端分离和微服务化架构，支持 Kubernetes 水平扩容。

## 核心设计原则

1. **前后端完全分离**：前端静态资源独立部署，后端提供纯 RESTful/gRPC API
2. **微服务化**：将后端按业务功能拆分为独立服务，各服务可独立扩容
3. **高可用性**：所有服务无状态设计，依赖中间件（Redis/消息队列）进行状态共享
4. **可观测性**：统一的日志、监控、链路追踪体系
5. **安全第一**：API 网关统一鉴权，服务间通信加密

---

## 系统架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                          客户端层 (Client Layer)                      │
├─────────────────────────────────────────────────────────────────────┤
│  Web 浏览器          │    移动端 APP (未来)   │   桌面应用 (未来)      │
└──────────────┬──────────────────────────────────────────────────────┘
               │ HTTPS
               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Ingress / Load Balancer                       │
│                        (Nginx/Traefik/Envoy)                         │
└──────────────┬──────────────────────────────────────────────────────┘
               │
       ┌───────┴────────┐
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────────────────────────────────────────┐
│             │  │         API 网关 (Gateway Service)                │
│   前端静态  │  │  - JWT 鉴权                                       │
│   资源服务  │  │  - 限流/熔断                                      │
│   (Nginx)   │  │  - 路由转发                                       │
│             │  │  - 日志记录                                       │
└─────────────┘  └──────────────┬───────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
        ┌──────────────┐ ┌─────────────┐ ┌──────────────┐
        │   用户服务   │ │  对话服务   │ │  中转服务    │
        │(User Service)│ │(Chat Service)│ │(Relay Service│
        └──────┬───────┘ └──────┬──────┘ └──────┬───────┘
               │                │               │
        ┌──────────────┐ ┌─────────────┐ ┌──────────────┐
        │  助手服务    │ │ 知识库服务  │ │  文件服务    │
        │(Agent Svc)   │ │ (RAG Svc)   │ │ (File Svc)   │
        └──────┬───────┘ └──────┬──────┘ └──────┬───────┘
               │                │               │
        ┌──────────────┐ ┌─────────────┐ ┌──────────────┐
        │  插件服务    │ │  计费服务   │ │  监控服务    │
        │(Plugin Svc)  │ │(Billing Svc)│ │(Monitor Svc) │
        └──────────────┘ └─────────────┘ └──────────────┘
                │                │               │
        ┌───────┴────────────────┴───────────────┴──────┐
        │                                                │
        ▼                                                ▼
┌──────────────┐                                ┌──────────────┐
│   PostgreSQL │                                │     Redis    │
│   (主数据库) │                                │   (缓存/会话) │
└──────────────┘                                └──────────────┘
        │                                                │
        ▼                                                ▼
┌──────────────┐                                ┌──────────────┐
│   对象存储   │                                │   消息队列   │
│    (MinIO)   │                                │   (RabbitMQ) │
└──────────────┘                                └──────────────┘
```

---

## 核心服务详解

### 1. **前端服务 (Frontend Service)**

**职责**：
- 提供用户界面（借鉴 LobeChat 的 UI/UX）
- 纯静态资源（HTML/CSS/JS），由 Nginx/Caddy 托管
- 通过 API 网关与后端通信

**技术栈**：
- React 19 + Next.js 15 (Static Export 模式)
- TypeScript
- Zustand (状态管理)
- TailwindCSS / antd-style

**部署**：
- Docker 镜像：基于 `nginx:alpine`
- 资源：CPU 0.1 core, Memory 128MB
- 副本数：2-10 (根据流量自动扩缩容)

---

### 2. **API 网关 (Gateway Service)**

**职责**：
- 统一入口，路由请求到各微服务
- JWT 鉴权和权限校验
- 限流、熔断、降级
- 日志聚合和链路追踪
- 跨域处理 (CORS)

**技术栈**：
- Golang (基于 NewAPI 的 Gin 框架)
- go-jwt (JWT 处理)
- go-redis (分布式限流)
- OpenTelemetry (链路追踪)

**API 设计**：
```
POST   /api/v1/auth/register       # 用户注册
POST   /api/v1/auth/login          # 用户登录
POST   /api/v1/auth/logout         # 登出
GET    /api/v1/auth/refresh        # 刷新 Token

GET    /api/v1/user/profile        # 获取用户信息
PUT    /api/v1/user/profile        # 更新用户信息
GET    /api/v1/user/quota          # 查询额度

POST   /api/v1/chat/completions    # 对话接口 (流式/非流式)
GET    /api/v1/chat/sessions       # 会话列表
POST   /api/v1/chat/sessions       # 创建会话
GET    /api/v1/chat/messages       # 消息历史

GET    /api/v1/agents              # 助手列表
POST   /api/v1/agents              # 创建助手
GET    /api/v1/agents/:id          # 获取助手详情

POST   /api/v1/files/upload        # 文件上传
GET    /api/v1/files/:id           # 文件下载

GET    /api/v1/models              # 可用模型列表
GET    /api/v1/plugins             # 插件列表

# B 端 API（保留 NewAPI 原有功能）
POST   /api/v1/admin/channels      # 渠道管理
GET    /api/v1/admin/logs          # 日志查询
GET    /api/v1/admin/stats         # 数据统计
```

**部署**：
- 资源：CPU 0.5 core, Memory 512MB
- 副本数：3-20 (网关是流量入口，需要足够副本)

---

### 3. **用户服务 (User Service)**

**职责**：
- 用户注册、登录、鉴权
- 用户信息管理（头像、昵称、偏好设置）
- 额度管理（充值、消费、查询）
- 2FA/Passkey 等安全功能

**数据模型**：
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    role INT DEFAULT 1,  -- 1: 普通用户, 10: VIP, 100: 管理员
    quota BIGINT DEFAULT 0,  -- 剩余额度（单位：分）
    total_quota BIGINT DEFAULT 0,  -- 总充值额度
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE TABLE user_settings (
    user_id INT PRIMARY KEY REFERENCES users(id),
    language VARCHAR(10) DEFAULT 'zh-CN',
    theme VARCHAR(20) DEFAULT 'auto',
    tts_voice VARCHAR(50),
    custom_config JSONB,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE quota_logs (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    type INT,  -- 1: 充值, 2: 消费, 3: 退款
    amount BIGINT,
    balance BIGINT,  -- 变更后余额
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**技术栈**：
- Golang + Gin
- GORM (ORM)
- bcrypt (密码加密)
- go-webauthn (Passkey 支持)

**部署**：
- 资源：CPU 0.2 core, Memory 256MB
- 副本数：2-10

---

### 4. **对话服务 (Chat Service)**

**职责**：
- 处理用户的聊天请求（核心服务）
- 管理会话 (Session) 和话题 (Topic)
- 维护消息历史
- 上下文管理和 Token 计数

**数据模型**：
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INT NOT NULL REFERENCES users(id),
    agent_id INT REFERENCES agents(id),  -- 关联的助手
    title VARCHAR(200),
    pinned BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    topic_id UUID REFERENCES topics(id),
    role VARCHAR(20) NOT NULL,  -- 'user' | 'assistant' | 'system'
    content TEXT NOT NULL,
    model VARCHAR(100),
    tokens INT,  -- Token 数量
    metadata JSONB,  -- 扩展字段（插件调用结果、工具调用等）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_session ON messages(session_id);
CREATE INDEX idx_messages_created ON messages(created_at DESC);

CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id),
    title VARCHAR(200),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**技术栈**：
- Golang + Gin
- SSE (Server-Sent Events) 流式响应
- 调用 **中转服务** 获取 AI 响应

**部署**：
- 资源：CPU 0.5 core, Memory 512MB
- 副本数：5-50 (对话服务是核心，流量最大)

---

### 5. **中转服务 (Relay Service)**

**职责**：
- 对接上游 AI 提供商（OpenAI/Claude/Gemini/国产大模型等）
- 渠道管理和负载均衡
- 请求重试和容错
- 计费和额度扣减（调用计费服务）

**复用 NewAPI 的核心功能**：
- `relay/` 模块：适配各家 AI 提供商的协议
- `model/channel.go`：渠道配置和缓存
- `service/channel_select.go`：智能选择渠道

**数据模型**：
```sql
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type INT NOT NULL,  -- 1: OpenAI, 2: Claude, 3: Gemini, ...
    base_url TEXT,
    key TEXT NOT NULL,  -- 加密存储的 API Key
    models TEXT[],  -- 支持的模型列表
    priority INT DEFAULT 0,
    weight INT DEFAULT 100,  -- 负载均衡权重
    status INT DEFAULT 1,  -- 1: 启用, 2: 禁用, 3: 维护中
    test_time TIMESTAMP,  -- 上次测试时间
    response_time INT,  -- 平均响应时间(ms)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**技术栈**：
- Golang + Gin
- go-resty (HTTP 客户端)
- 复用 NewAPI 的 relay 适配层

**部署**：
- 资源：CPU 1 core, Memory 1GB (需要处理大量并发请求)
- 副本数：5-30

---

### 6. **助手服务 (Agent Service)**

**职责**：
- 管理用户创建的助手（Agent/角色）
- 存储助手的系统提示词、工具配置
- 提供助手市场（用户分享的助手）

**数据模型**：
```sql
CREATE TABLE agents (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),  -- NULL 表示系统内置
    name VARCHAR(100) NOT NULL,
    avatar VARCHAR(255),
    description TEXT,
    system_role TEXT,  -- System Prompt
    model VARCHAR(100),
    temperature FLOAT DEFAULT 0.7,
    tools JSONB,  -- 启用的工具列表
    plugins JSONB,  -- 启用的插件
    is_public BOOLEAN DEFAULT FALSE,  -- 是否在市场公开
    likes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**技术栈**：
- Golang + Gin
- PostgreSQL (存储)

**部署**：
- 资源：CPU 0.2 core, Memory 256MB
- 副本数：2-5

---

### 7. **知识库服务 (RAG Service)**

**职责**：
- 文档上传和解析（PDF/Word/Markdown/网页）
- 文本切片 (Chunking)
- 向量化和索引
- 检索增强生成 (RAG)

**数据模型**：
```sql
CREATE TABLE knowledge_bases (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-3-small',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id INT NOT NULL REFERENCES knowledge_bases(id),
    title VARCHAR(200),
    file_url TEXT,
    file_type VARCHAR(50),
    status INT DEFAULT 1,  -- 1: 处理中, 2: 已完成, 3: 失败
    chunk_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id),
    content TEXT NOT NULL,
    embedding VECTOR(1536),  -- 使用 pgvector 扩展
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chunks_embedding ON chunks USING ivfflat (embedding vector_cosine_ops);
```

**技术栈**：
- Golang + Gin
- pgvector (PostgreSQL 向量扩展)
- Unstructured (文档解析)
- 调用 **中转服务** 获取 Embedding

**部署**：
- 资源：CPU 0.5 core, Memory 512MB
- 副本数：2-10

---

### 8. **文件服务 (File Service)**

**职责**：
- 文件上传（用户头像、对话图片、文档）
- 文件存储（对接 MinIO/S3）
- 文件访问授权和临时 URL 签名

**技术栈**：
- Golang + Gin
- MinIO SDK

**部署**：
- 资源：CPU 0.2 core, Memory 256MB
- 副本数：2-5

---

### 9. **插件服务 (Plugin Service)**

**职责**：
- 插件市场（浏览、搜索插件）
- 插件安装和配置
- 插件调用代理（安全沙箱执行）

**数据模型**：
```sql
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    identifier VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    author VARCHAR(100),
    version VARCHAR(20),
    manifest JSONB,  -- 插件的功能定义
    api_endpoint TEXT,
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_plugins (
    user_id INT REFERENCES users(id),
    plugin_id INT REFERENCES plugins(id),
    config JSONB,  -- 用户的个性化配置（如 API Key）
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, plugin_id)
);
```

**技术栈**：
- Golang + Gin
- HTTP Proxy (转发插件请求)

**部署**：
- 资源：CPU 0.2 core, Memory 256MB
- 副本数：2-5

---

### 10. **计费服务 (Billing Service)**

**职责**：
- 按量计费（Token 消耗）
- 计费规则管理（不同模型的价格）
- 充值和余额变动记录
- 对接支付网关（支付宝/微信/Stripe）

**数据模型**：
```sql
CREATE TABLE pricing (
    id SERIAL PRIMARY KEY,
    model VARCHAR(100) UNIQUE NOT NULL,
    input_price DECIMAL(10, 6),  -- 单位：元/1K tokens
    output_price DECIMAL(10, 6),
    currency VARCHAR(10) DEFAULT 'CNY',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE billing_logs (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    session_id UUID REFERENCES sessions(id),
    model VARCHAR(100),
    input_tokens INT,
    output_tokens INT,
    cost BIGINT,  -- 花费（单位：分）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_billing_user ON billing_logs(user_id, created_at DESC);
```

**技术栈**：
- Golang + Gin
- PostgreSQL (记录账单)
- Redis (缓存计费规则)

**部署**：
- 资源：CPU 0.3 core, Memory 256MB
- 副本数：2-10

---

### 11. **监控服务 (Monitor Service)**

**职责**：
- 收集系统指标（QPS、延迟、错误率）
- 健康检查和告警
- 日志聚合和查询

**技术栈**：
- Prometheus (指标采集)
- Grafana (可视化)
- Loki (日志聚合)
- OpenTelemetry (链路追踪)

**部署**：
- 独立部署，不需要水平扩容

---

## 中间件和基础设施

### 1. **数据库 (PostgreSQL)**
- **主数据库**：存储用户、会话、消息、助手等核心数据
- **副本读写分离**：主库写入，从库读取（降低主库压力）
- **分片策略**：按 `user_id` 分片（当单库容量 > 500GB 时启用）

### 2. **缓存 (Redis)**
- **用途**：
  - 会话状态缓存
  - JWT Token 黑名单
  - 限流计数器
  - 热点数据缓存（用户信息、模型列表）
- **部署模式**：Redis Cluster (3 主 3 从)

### 3. **对象存储 (MinIO/S3)**
- **用途**：
  - 用户上传的文件
  - 语音文件
  - 导出的对话记录
- **生命周期管理**：30 天后自动迁移到冷存储

### 4. **消息队列 (RabbitMQ/Kafka)**
- **用途**：
  - 异步任务（文档解析、邮件发送）
  - 解耦服务（计费日志异步写入）
  - 事件驱动架构

### 5. **服务网格 (Istio/Linkerd)**
- **可选**：当服务数量 > 10 时引入
- **功能**：
  - 服务间通信加密 (mTLS)
  - 流量管理和灰度发布
  - 可观测性增强

---

## 数据流示例

### 用户发起聊天请求

1. **前端** 发送 POST `/api/v1/chat/completions`，携带 JWT Token
2. **API 网关** 验证 Token，从 Redis 获取用户信息
3. **网关** 检查用户额度是否充足（调用 **用户服务**）
4. **网关** 路由到 **对话服务**
5. **对话服务** 从数据库加载历史消息，构建上下文
6. **对话服务** 调用 **中转服务** 获取 AI 响应（SSE 流式）
7. **中转服务** 选择最优渠道，转发请求到上游 AI
8. 响应通过 SSE 流式返回给前端
9. **对话服务** 异步调用 **计费服务** 记录消费
10. **计费服务** 扣减用户额度，写入计费日志

---

## 部署架构

### Kubernetes 部署清单

```yaml
# Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: oblivious

---
# 前端服务
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: oblivious
spec:
  replicas: 3
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: nginx
        image: oblivious-frontend:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: oblivious
spec:
  selector:
    app: frontend
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP

---
# API 网关
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
  namespace: oblivious
spec:
  replicas: 5
  selector:
    matchLabels:
      app: gateway
  template:
    metadata:
      labels:
        app: gateway
    spec:
      containers:
      - name: gateway
        image: oblivious-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: url
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: gateway-service
  namespace: oblivious
spec:
  selector:
    app: gateway
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP

---
# 其他服务（用户服务、对话服务等）
# 结构类似，根据资源需求调整 replicas 和 resources
```

### HPA (水平自动扩缩容)

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: chat-service-hpa
  namespace: oblivious
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: chat-service
  minReplicas: 5
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 开发计划

### 阶段 1：基础架构搭建 (2 周)
- [ ] 创建项目目录结构
- [ ] 实现 API 网关和基础鉴权
- [ ] 实现用户服务（注册/登录）
- [ ] 搭建数据库和 Redis

### 阶段 2：核心功能开发 (4 周)
- [ ] 实现对话服务
- [ ] 实现中转服务（复用 NewAPI 逻辑）
- [ ] 实现计费服务
- [ ] 前端基础页面（登录、对话界面）

### 阶段 3：高级功能 (3 周)
- [ ] 实现助手服务
- [ ] 实现知识库服务 (RAG)
- [ ] 实现插件系统
- [ ] 文件上传和管理

### 阶段 4：优化与上线 (2 周)
- [ ] 性能测试和优化
- [ ] 监控和告警配置
- [ ] 文档编写
- [ ] 生产环境部署

---

## 技术栈总结

| 层级       | 技术选型                                  |
|------------|-------------------------------------------|
| 前端       | React 19 + Next.js 15 + TypeScript        |
| 网关       | Golang + Gin + JWT + OpenTelemetry        |
| 后端服务   | Golang + Gin + GORM                       |
| 数据库     | PostgreSQL 15 + pgvector                  |
| 缓存       | Redis Cluster                             |
| 消息队列   | RabbitMQ                                  |
| 对象存储   | MinIO                                     |
| 容器编排   | Kubernetes + Docker                       |
| CI/CD      | GitHub Actions                            |
| 监控       | Prometheus + Grafana + Loki               |
| 链路追踪   | OpenTelemetry + Jaeger                    |

---

## 安全策略

1. **身份验证**：JWT Token + Refresh Token 机制
2. **API 限流**：基于 Redis 的令牌桶算法
3. **数据加密**：敏感字段（API Key、密码）使用 AES-256 加密
4. **HTTPS**：所有外部通信强制 HTTPS
5. **SQL 注入防护**：使用 GORM 参数化查询
6. **XSS 防护**：前端输出转义

---

## 未来扩展

1. **多租户支持**：为企业客户提供独立部署
2. **移动端 APP**：React Native 开发
3. **桌面端应用**：Electron 封装
4. **语音对话**：实时 STT/TTS
5. **多模态**：图片理解、视频分析
6. **联邦学习**：私有数据训练

---

## 总结

**Oblivious** 的架构设计遵循现代云原生应用的最佳实践，通过微服务化实现了灵活的扩展能力，同时兼顾了 C 端用户体验和 B 端 API 服务的需求。整体架构清晰、可维护性强，能够支撑百万级用户并发访问。

