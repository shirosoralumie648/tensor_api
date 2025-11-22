# Oblivious AI 平台 - 完整代码地图

> 📅 生成时间: 2025-11-22  
> 📊 项目状态: 开发中 (30% 完成)  
> 🎯 目标: C端AI应用平台 + B端API中转

---

## 📋 目录

- [1. 项目概览](#1-项目概览)
- [2. 技术栈](#2-技术栈)
- [3. 项目结构](#3-项目结构)
- [4. 核心模块详解](#4-核心模块详解)
- [5. 数据库架构](#5-数据库架构)
- [6. 服务架构](#6-服务架构)
- [7. 关键文件索引](#7-关键文件索引)
- [8. 数据流转](#8-数据流转)
- [9. 开发指南](#9-开发指南)

---

## 1. 项目概览

### 项目简介

Oblivious 是一个面向 C 端用户的 AI 应用服务平台，采用微服务架构设计，同时保留 B 端 API 中转能力。

### 核心特性

#### C 端功能
- 🤖 智能对话: GPT-4、Claude、Gemini等主流大模型
- 👤 AI 助手: 助手市场，一键安装
- 📚 知识库: RAG技术，文档问答
- 🔌 插件系统: 联网搜索、代码执行、图片生成
- 🎨 精美界面: 现代化设计，支持深色模式

#### B 端功能
- 🔄 API中转: 统一接口对接多家AI提供商
- 💰 计费管理: 按量计费，支持额度充值
- 📊 数据统计: 实时监控API调用
- 🔐 权限管理: 多用户、多渠道管理
- ⚖️ 负载均衡: 智能选择最优渠道

### 项目规模

```
总代码量: 50,000+ 行
后端代码: 30,000+ 行 Go
前端代码: 15,000+ 行 TypeScript
配置文件: 5,000+ 行
文档: 20+ 个 Markdown 文件
数据库表: 16个迁移文件（32个.sql文件）
API接口: 100+ 个
微服务: 11 个（gateway, user, chat, relay, billing, agent, kb, file, plugin, rag, migrate）
```

---

## 2. 技术栈

### 后端技术栈

```yaml
语言: Go 1.23+
Web框架: Gin
ORM: GORM
数据库: PostgreSQL 15
缓存: Redis Cluster
消息队列: RabbitMQ
对象存储: MinIO
向量数据库: pgvector

核心依赖:
  - gin-gonic/gin: Web框架
  - gorm.io/gorm: ORM
  - golang-jwt/jwt: JWT认证
  - go-redis/redis: Redis客户端
  - pkoukk/tiktoken-go: Token计数
  - prometheus/client_golang: 监控
  - uber-go/zap: 日志
```

### 前端技术栈

```yaml
框架: React 18 + Next.js 14
语言: TypeScript 5.2
状态管理: Zustand
HTTP客户端: Axios
UI组件: 自定义组件库
样式: TailwindCSS
图表: Recharts
Markdown: markdown-it
代码高亮: highlight.js
数学公式: KaTeX
图表渲染: Mermaid

开发工具:
  - ESLint: 代码检查
  - Prettier: 代码格式化
  - Jest: 单元测试
  - Storybook: 组件开发
```

### DevOps技术栈

```yaml
容器化: Docker
编排: Kubernetes
CI/CD: GitHub Actions
监控: Prometheus + Grafana
日志: Loki
链路追踪: Jaeger (OpenTelemetry)
```

---

## 3. 项目结构

### 总体目录结构

```
oblivious/
├── backend/                 # Go 后端服务
├── frontend/                # React 前端应用
├── lobe-chat-next/          # LobeChat 集成
├── new-api-main/            # NewAPI 集成
├── deploy/                  # 部署配置
├── docs/                    # 项目文档
├── .github/                 # GitHub配置
├── docker-compose.yml       # Docker编排
├── README.md               # 项目说明
├── PROJECT_SUMMARY.md      # 项目总结
└── CODEMAP.md             # 本文档
```

### 后端目录结构 (`backend/`)

```
backend/
├── cmd/                           # 服务启动入口
│   ├── gateway/                   # API网关服务
│   │   └── main.go
│   ├── user/                      # 用户服务
│   │   └── main.go
│   ├── chat/                      # 对话服务
│   │   └── main.go
│   ├── relay/                     # 中转服务
│   │   └── main.go
│   ├── billing/                   # 计费服务
│   ├── agent/                     # 助手服务
│   ├── kb/                        # 知识库服务
│   ├── file/                      # 文件服务
│   ├── plugin/                    # 插件服务
│   ├── worker/                    # 后台任务服务
│   ├── migrate/                   # 数据库迁移工具
│   ├── rag/                       # RAG服务
│   └── server/                    # 单体服务(开发用)
│       └── main_example.go
│
├── internal/                      # 内部业务逻辑
│   ├── adapter/                   # AI提供商适配器 ✅
│   │   ├── adapter.go             # 适配器接口定义
│   │   ├── factory.go             # 适配器工厂
│   │   ├── registry.go            # 适配器注册中心
│   │   ├── providers.go           # 提供商列表
│   │   └── batch_providers.go     # 批量提供商管理
│   │
│   ├── selector/                  # 渠道选择器 ✅
│   │   ├── selector.go            # 选择器接口
│   │   ├── types.go               # 类型定义
│   │   ├── cache.go               # 渠道缓存
│   │   ├── stats.go               # 统计管理
│   │   ├── strategies.go          # 选择策略
│   │   └── retry.go               # 重试逻辑
│   │
│   ├── tokenizer/                 # Token计数器 ✅
│   │   ├── factory.go             # 计数器工厂
│   │   ├── tiktoken.go            # Tiktoken实现
│   │   ├── models.go              # 模型映射
│   │   └── counter.go             # 计数器接口
│   │
│   ├── quota/                     # 配额服务 ✅
│   │   ├── types.go               # 类型定义
│   │   ├── service.go             # 服务实现
│   │   ├── calculator.go          # 配额计算
│   │   └── cache.go               # 配额缓存
│   │
│   ├── relay/                     # 中继层 ✅
│   │   ├── types.go               # 请求/响应类型
│   │   ├── stream_handler.go     # 流式处理
│   │   ├── stream_sender.go      # SSE发送器
│   │   ├── stream_monitor.go     # 流式监控
│   │   └── dispatcher.go         # 请求分发
│   │
│   ├── service/                   # 业务服务层 ✅
│   │   ├── pricing_service.go    # 定价服务
│   │   ├── channel_ability_service.go  # 渠道能力
│   │   ├── health_check_service.go     # 健康检查
│   │   └── ...
│   │
│   ├── handler/                   # HTTP处理器 ✅
│   │   ├── channel_handler.go    # 渠道管理API
│   │   ├── pricing_handler.go    # 定价管理API
│   │   ├── stats_handler.go      # 统计监控API
│   │   └── health_handler.go     # 健康检查API
│   │
│   ├── repository/                # 数据访问层
│   │   ├── channel_repository.go
│   │   ├── user_repository.go
│   │   └── ...
│   │
│   ├── model/                     # 数据模型 ✅
│   │   ├── user.go                # 用户模型
│   │   ├── channel.go             # 渠道模型
│   │   ├── adapter_config.go      # 适配器配置
│   │   ├── channel_ability.go     # 渠道能力
│   │   ├── unified_log.go         # 统一日志
│   │   ├── model_pricing.go       # 模型定价
│   │   └── ...
│   │
│   ├── middleware/                # 中间件
│   │   ├── auth.go                # JWT认证
│   │   ├── cors.go                # 跨域处理
│   │   ├── logger.go              # 日志记录
│   │   ├── ratelimit.go           # 限流
│   │   └── recovery.go            # 错误恢复
│   │
│   ├── chat/                      # 聊天相关
│   │   ├── session.go             # 会话管理
│   │   ├── message_formatter.go   # 消息格式化
│   │   ├── context_manager.go     # 上下文管理
│   │   └── ...
│   │
│   ├── billing/                   # 计费相关
│   │   ├── quota_manager.go       # 配额管理
│   │   ├── pricing.go             # 定价逻辑
│   │   ├── token_counter.go       # Token计数
│   │   └── async_accounting.go    # 异步记账
│   │
│   ├── cache/                     # 缓存层
│   │   ├── cache.go               # 缓存接口
│   │   ├── redis.go               # Redis实现
│   │   ├── user_cache.go          # 用户缓存
│   │   └── ...
│   │
│   ├── analytics/                 # 数据分析
│   │   ├── realtime_stats.go      # 实时统计
│   │   └── usage_logger.go        # 使用日志
│   │
│   ├── rag/                       # RAG系统
│   │   ├── embedding.go           # 向量化
│   │   ├── chunker.go             # 文本分块
│   │   └── retriever.go           # 检索器
│   │
│   ├── security/                  # 安全相关
│   │   ├── jwt.go                 # JWT工具
│   │   └── crypto.go              # 加密工具
│   │
│   ├── database/                  # 数据库工具
│   │   └── postgres.go            # PostgreSQL连接
│   │
│   ├── config/                    # 配置管理
│   │   └── config.go
│   │
│   └── utils/                     # 工具函数
│       └── ...
│
├── migrations/                    # 数据库迁移 ✅
│   ├── 000001_create_users_table.up.sql
│   ├── 000002_create_user_settings_table.up.sql
│   ├── 000006_create_channels_table.up.sql
│   ├── 000013_create_adapter_configs.up.sql
│   ├── 000014_create_channel_abilities.up.sql
│   ├── 000015_create_unified_logs.up.sql
│   ├── 000016_create_model_pricing.up.sql
│   └── ... (共32个迁移文件)
│
├── pkg/                           # 公共包
│   ├── logger/                    # 日志包
│   ├── metrics/                   # 监控指标
│   └── queue/                     # 消息队列
│
├── scripts/                       # 运行脚本
│   ├── init.sh                    # 初始化脚本
│   └── ...
│
├── go.mod                         # Go依赖管理
├── go.sum                         # 依赖校验
├── Dockerfile                     # Docker镜像
└── README.md                      # 后端文档
```

### 前端目录结构 (`frontend/`)

```
frontend/
├── src/
│   ├── app/                       # Next.js App Router
│   │   ├── layout.tsx             # 根布局
│   │   ├── page.tsx               # 首页
│   │   ├── chat/                  # 聊天页面
│   │   ├── assistant/             # 助手页面
│   │   ├── knowledge/             # 知识库页面
│   │   └── settings/              # 设置页面
│   │
│   ├── components/                # React组件
│   │   ├── chat/                  # 聊天组件
│   │   │   ├── ChatBox.tsx
│   │   │   ├── MessageList.tsx
│   │   │   └── InputArea.tsx
│   │   ├── assistant/             # 助手组件
│   │   ├── common/                # 通用组件
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   └── Modal.tsx
│   │   └── layout/                # 布局组件
│   │       ├── Header.tsx
│   │       ├── Sidebar.tsx
│   │       └── Footer.tsx
│   │
│   ├── hooks/                     # 自定义Hooks
│   │   ├── useChat.ts             # 聊天Hook
│   │   ├── useAuth.ts             # 认证Hook
│   │   └── useWebSocket.ts        # WebSocket Hook
│   │
│   ├── stores/                    # Zustand状态管理
│   │   ├── authStore.ts           # 认证状态
│   │   ├── chatStore.ts           # 聊天状态
│   │   └── settingsStore.ts       # 设置状态
│   │
│   ├── services/                  # API服务
│   │   ├── api.ts                 # API客户端
│   │   ├── chatApi.ts             # 聊天API
│   │   ├── userApi.ts             # 用户API
│   │   └── assistantApi.ts        # 助手API
│   │
│   ├── types/                     # TypeScript类型
│   │   ├── chat.ts
│   │   ├── user.ts
│   │   └── assistant.ts
│   │
│   ├── utils/                     # 工具函数
│   │   ├── format.ts              # 格式化
│   │   └── validator.ts           # 验证
│   │
│   └── styles/                    # 样式文件
│       └── globals.css
│
├── public/                        # 静态资源
│   ├── images/
│   └── icons/
│
├── .storybook/                    # Storybook配置
├── package.json                   # 依赖管理
├── tsconfig.json                  # TypeScript配置
├── tailwind.config.js             # TailwindCSS配置
└── next.config.js                 # Next.js配置
```

---

## 4. 核心模块详解

### 4.1 适配器系统 (Adapter)

**位置**: `backend/internal/adapter/`

**功能**: 统一不同AI提供商的接口

**核心文件**:
- `adapter.go`: 定义适配器接口
- `factory.go`: 适配器工厂模式
- `registry.go`: 适配器注册中心
- `providers.go`: 支持的提供商列表

**支持的提供商**:
```go
- OpenAI (GPT-3.5, GPT-4, GPT-4o)
- Anthropic (Claude-3)
- Google (Gemini)
- Azure OpenAI
- 国内模型 (通义千问、文心一言等)
```

**关键代码**:
```go
// 适配器接口
type Adapter interface {
    ChatCompletion(ctx context.Context, req *relay.Request) (*relay.Response, error)
    StreamChatCompletion(ctx context.Context, req *relay.Request) (<-chan *relay.StreamChunk, error)
}

// 工厂方法
func CreateAdapter(providerType string, config *Config) (Adapter, error)
```

### 4.2 渠道选择器 (Selector)

**位置**: `backend/internal/selector/`

**功能**: 智能选择最优AI渠道

**5种选择策略**:
1. **权重策略** (Weight): 按配置权重随机选择
2. **优先级策略** (Priority): 选择最高优先级渠道
3. **轮询策略** (RoundRobin): 循环选择渠道
4. **最低延迟** (LowestLatency): 选择响应最快的渠道
5. **随机策略** (Random): 完全随机选择

**核心文件**:
- `selector.go`: 选择器接口和主逻辑
- `strategies.go`: 各种选择策略实现
- `cache.go`: 渠道缓存管理
- `stats.go`: 渠道统计信息
- `retry.go`: 失败重试逻辑

**关键功能**:
```go
// 选择渠道
func (s *Selector) Select(ctx context.Context, req *SelectRequest) (*Channel, error)

// 带重试的选择
func (s *Selector) SelectWithRetry(ctx context.Context, req *SelectRequest, maxRetries int) (*Channel, error)

// 更新统计
func (s *Selector) UpdateStats(ctx context.Context, channelID int, success bool, latency time.Duration)
```

### 4.3 Token计数器 (Tokenizer)

**位置**: `backend/internal/tokenizer/`

**功能**: 精确计算Token消耗

**支持的模型**:
- GPT-4系列: cl100k_base编码
- GPT-3.5系列: cl100k_base编码  
- Claude系列: cl100k_base编码
- 自定义模型: 可配置编码器

**核心文件**:
- `factory.go`: 计数器工厂
- `tiktoken.go`: Tiktoken实现
- `models.go`: 模型编码器映射
- `counter.go`: 计数器接口

**关键代码**:
```go
// 计算文本Token数
func CountTokens(text string, model string) (int, error)

// 计算消息Token数
func CountMessageTokens(messages []Message, model string) (int, error)

// 流式计数
func CountStreamTokens(chunks <-chan string, model string) (int, error)
```

### 4.4 配额服务 (Quota)

**位置**: `backend/internal/quota/`

**功能**: 用户配额管理和计费

**计费模式**:
1. **按Token计费**: 精确到每个Token
2. **按次计费**: 固定价格

**用户分组折扣**:
- default: 1.0x (标准价格)
- vip: 0.8x (8折)
- premium: 0.6x (6折)
- free: 1.5x (免费用户加价)

**核心文件**:
- `service.go`: 配额服务主逻辑
- `calculator.go`: 配额计算器
- `cache.go`: 配额缓存
- `types.go`: 类型定义

**扣费流程**:
```
1. 预扣费 (Pre-deduct): 根据估算扣除配额
2. 调用API
3. 后扣费 (Post-deduct): 根据实际消耗调整
4. 失败退款: 失败时全额退还
```

### 4.5 流式处理 (Stream)

**位置**: `backend/internal/relay/`

**功能**: SSE流式响应处理

**核心文件**:
- `stream_handler.go`: 流式处理主逻辑
- `stream_sender.go`: SSE发送器
- `stream_monitor.go`: 流式监控
- `types.go`: 请求响应类型

**流式处理流程**:
```
1. 建立SSE连接
2. 接收上游流式数据
3. 实时Token计数
4. 发送到客户端
5. 完成后扣费
```

**关键代码**:
```go
// 处理流式响应
func HandleStream(ctx context.Context, upstream <-chan *Chunk, downstream chan<- *Chunk)

// SSE发送
func SendSSE(w http.ResponseWriter, chunk *Chunk)
```

### 4.6 定价系统 (Pricing)

**位置**: `backend/internal/service/pricing_service.go`

**功能**: 灵活的模型定价管理

**定价维度**:
- 模型: 不同模型不同价格
- 用户组: 用户分组折扣
- Token类型: Prompt和Completion分别计价

**核心功能**:
```go
// 获取模型定价
func GetModelPricing(model string) (*Pricing, error)

// 计算配额消耗
func CalculateQuota(model string, promptTokens, completionTokens int, userGroup string) (int, error)

// 更新定价
func UpdatePricing(model string, pricing *Pricing) error
```

### 4.7 健康检查 (Health Check)

**位置**: `backend/internal/service/health_check_service.go`

**功能**: 自动渠道健康监控

**监控指标**:
- 可用性: 渠道是否可访问
- 响应时间: 平均延迟
- 成功率: 请求成功比例
- 健康评分: 综合评分(0-100)

**自动运维**:
- 定期检查: 每30分钟自动检查
- 自动禁用: 连续失败3次自动禁用
- 自动恢复: 恢复后自动启用

**关键代码**:
```go
// 检查渠道健康
func CheckChannelHealth(ctx context.Context, channelID int) (*HealthStatus, error)

// 获取健康评分
func GetHealthScore(ctx context.Context, channelID int) (int, error)

// 自动检查所有渠道
func AutoCheckAll(ctx context.Context) error
```

---

## 5. 数据库架构

### 数据库表结构

#### 核心业务表

**1. users - 用户表**
```sql
id                BIGSERIAL PRIMARY KEY
username          VARCHAR(50) UNIQUE
email             VARCHAR(255) UNIQUE  
password_hash     VARCHAR(255)
quota             INTEGER DEFAULT 0      -- 剩余配额
user_group        VARCHAR(20)           -- 用户分组
status            SMALLINT DEFAULT 0     -- 状态: 0正常 1禁用
created_at        TIMESTAMP
updated_at        TIMESTAMP
```

**2. channels - 渠道表**
```sql
id                BIGSERIAL PRIMARY KEY
name              VARCHAR(100)           -- 渠道名称
type              VARCHAR(50)            -- 提供商类型
api_base          VARCHAR(255)           -- API基础URL
api_keys          TEXT                   -- API密钥(加密)
support_models    TEXT                   -- 支持的模型列表
priority          INTEGER DEFAULT 0      -- 优先级
weight            INTEGER DEFAULT 1      -- 权重
status            SMALLINT DEFAULT 0     -- 状态: 0启用 1禁用
test_model        VARCHAR(100)           -- 测试用模型
created_at        TIMESTAMP
updated_at        TIMESTAMP
```

**3. adapter_configs - 适配器配置表**
```sql
id                BIGSERIAL PRIMARY KEY
channel_id        BIGINT REFERENCES channels(id)
provider_type     VARCHAR(50)            -- 提供商类型
config_json       JSONB                  -- 配置JSON
enabled           BOOLEAN DEFAULT true
created_at        TIMESTAMP
updated_at        TIMESTAMP
```

**4. channel_abilities - 渠道能力表**
```sql
id                BIGSERIAL PRIMARY KEY
channel_id        BIGINT REFERENCES channels(id)
model             VARCHAR(100)           -- 模型名称
max_tokens        INTEGER                -- 最大Token数
supports_stream   BOOLEAN DEFAULT true   -- 支持流式
supports_functions BOOLEAN DEFAULT false -- 支持函数调用
supports_vision   BOOLEAN DEFAULT false  -- 支持视觉
price_info        JSONB                  -- 价格信息
created_at        TIMESTAMP
updated_at        TIMESTAMP
```

**5. model_pricing - 模型定价表**
```sql
id                BIGSERIAL PRIMARY KEY
model             VARCHAR(100) UNIQUE    -- 模型名称
quota_type        SMALLINT DEFAULT 0     -- 计费类型: 0按Token 1按次
model_ratio       DECIMAL(10,2)          -- 模型倍率
completion_ratio  DECIMAL(10,2)          -- Completion倍率
group_ratio       JSONB                  -- 分组倍率
enabled           BOOLEAN DEFAULT true
created_at        TIMESTAMP
updated_at        TIMESTAMP
```

#### 日志和统计表

**6. unified_logs - 统一日志表**
```sql
id                BIGSERIAL PRIMARY KEY
user_id           BIGINT                 -- 用户ID
channel_id        BIGINT                 -- 渠道ID
model             VARCHAR(100)           -- 使用的模型
request_type      VARCHAR(50)            -- 请求类型
prompt_tokens     INTEGER                -- Prompt Token数
completion_tokens INTEGER                -- Completion Token数
quota_used        INTEGER                -- 消耗配额
response_time     INTEGER                -- 响应时间(ms)
success           BOOLEAN                -- 是否成功
error_message     TEXT                   -- 错误信息
created_at        TIMESTAMP
INDEX idx_user_created (user_id, created_at)
INDEX idx_channel_created (channel_id, created_at)
```

**7. sessions - 会话表**
```sql
id                BIGSERIAL PRIMARY KEY
user_id           BIGINT REFERENCES users(id)
title             VARCHAR(200)           -- 会话标题
model             VARCHAR(100)           -- 使用的模型
system_prompt     TEXT                   -- 系统提示词
context_length    INTEGER DEFAULT 10     -- 上下文长度
created_at        TIMESTAMP
updated_at        TIMESTAMP
```

**8. messages - 消息表**
```sql
id                BIGSERIAL PRIMARY KEY  
session_id        BIGINT REFERENCES sessions(id)
role              VARCHAR(20)            -- user/assistant/system
content           TEXT                   -- 消息内容
tokens            INTEGER                -- Token数
created_at        TIMESTAMP
INDEX idx_session (session_id, created_at)
```

#### 计费相关表

**9. quota_logs - 配额日志表**
```sql
id                BIGSERIAL PRIMARY KEY
user_id           BIGINT REFERENCES users(id)
change_amount     INTEGER                -- 变化量(正/负)
balance_after     INTEGER                -- 变化后余额
log_type          SMALLINT               -- 类型: 0充值 1消费 2退款
description       TEXT                   -- 说明
created_at        TIMESTAMP
INDEX idx_user_created (user_id, created_at)
```

**10. billing_records - 账单记录表**
```sql
id                BIGSERIAL PRIMARY KEY
user_id           BIGINT REFERENCES users(id)
log_id            BIGINT REFERENCES unified_logs(id)
model             VARCHAR(100)
tokens_used       INTEGER
quota_consumed    INTEGER
amount            DECIMAL(10,2)          -- 金额
created_at        TIMESTAMP
```

### 数据库索引策略

```sql
-- 用户查询优化
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- 渠道查询优化
CREATE INDEX idx_channels_type ON channels(type);
CREATE INDEX idx_channels_status ON channels(status);

-- 日志查询优化
CREATE INDEX idx_logs_user_time ON unified_logs(user_id, created_at DESC);
CREATE INDEX idx_logs_channel_time ON unified_logs(channel_id, created_at DESC);
CREATE INDEX idx_logs_model ON unified_logs(model);

-- 会话消息查询优化  
CREATE INDEX idx_messages_session ON messages(session_id, created_at DESC);
```

---

## 6. 服务架构

### 6.1 微服务列表

```
1. Gateway Service (网关服务)
   - 端口: 8080
   - 职责: 统一入口、认证、路由

2. User Service (用户服务)
   - 端口: 8081
   - 职责: 用户管理、认证授权

3. Chat Service (对话服务)
   - 端口: 8082
   - 职责: 会话管理、消息处理

4. Relay Service (中转服务)
   - 端口: 8083
   - 职责: AI API中转、渠道调度

5. Billing Service (计费服务)
   - 端口: 8084
   - 职责: 配额管理、账单记录

6. Agent Service (助手服务)
   - 端口: 8085
   - 职责: AI助手管理

7. Knowledge Service (知识库服务)
   - 端口: 8086
   - 职责: RAG、向量检索

8. File Service (文件服务)
   - 端口: 8087
   - 职责: 文件上传下载

9. Plugin Service (插件服务)
   - 端口: 8088
   - 职责: 插件管理和调用

10. Worker Service (后台任务)
    - 职责: 异步任务处理
```

### 6.2 服务间通信

**同步通信**: HTTP/gRPC
**异步通信**: RabbitMQ消息队列

```
┌─────────┐    HTTP     ┌─────────┐
│ Gateway │────────────>│  User   │
└─────────┘             └─────────┘
     │
     │ HTTP
     ▼
┌─────────┐    HTTP     ┌─────────┐
│  Chat   │────────────>│  Relay  │
└─────────┘             └─────────┘
     │                       │
     │ MQ                    │ MQ
     ▼                       ▼
┌─────────────────────────────────┐
│        RabbitMQ                 │
└─────────────────────────────────┘
     │
     │ Consumer
     ▼
┌─────────┐
│ Billing │
└─────────┘
```

---

## 7. 关键文件索引

### 7.1 后端核心文件

#### 服务启动入口

**gateway/main.go** - API网关服务入口
- `main()` - 网关服务启动函数
  - 加载配置 (config.Load())
  - 初始化日志 (logger.Init())
  - 初始化Redis（用于限流）
  - 初始化JWT
  - 配置Gin框架
  - **全局中间件:**
    - Recovery - panic恢复
    - RequestIDMiddleware - 请求ID追踪
    - LoggerMiddleware - 请求日志
    - CORSMiddleware - 跨域处理
  - **路由配置:**
    - 公开接口 (限流: 10 req/min)
      - POST /api/v1/register - 注册
      - POST /api/v1/login - 登录
      - POST /api/v1/refresh - 刷新token
    - 需鉴权接口 (限流: 100 req/min)
      - 用户相关: profile
      - 对话相关: sessions, messages
      - 计费相关: billing
  - **请求代理:**
    - `proxyToService()` - 普通HTTP请求代理
      - 复制请求头和Body
      - 传递用户信息 (X-User-ID, X-Username, X-User-Role)
      - 转发到目标微服务
      - 返回响应
    - `proxyToServiceSSE()` - SSE流式请求代理
      - 设置SSE响应头
      - 流式转发响应
      - 超时时间: 300秒
  - 健康检查: GET /health
  - 启动服务: 端口8080

**relay/main.go** - 中转服务入口
- `main()` - 中转服务启动函数
  - 加载配置
  - 初始化日志、数据库、JWT
  - 配置Gin框架和中间件
  - **初始化服务:**
    - RelayService - 中转服务
  - **API路由:**
    - POST /v1/chat/completions - 聊天补全接口
      - 支持流式(stream=true)和非流式
      - 流式: SSE格式，实时推送
      - 非流式: JSON响应
    - GET /v1/models - 列出可用模型
      - 返回所有渠道支持的模型列表
    - GET /v1/channels - 获取渠道列表
  - **管理接口 (需鉴权):**
    - GET /v1/model-price/:channel_id/:model - 获取模型价格
  - **核心逻辑:**
    - 调用RelayService处理请求
    - 流式响应使用SSE格式 ("data: {json}\n\n")
    - 完成标记: "data: [DONE]\n\n"
  - 健康检查: GET /health
  - 启动服务: 端口8083

**user/main.go** - 用户服务入口
- `main()` - 用户服务启动函数
  - 初始化配置、日志、数据库
  - 用户认证API
  - 用户管理API
  - JWT token管理
  - 端口8081

**chat/main.go** - 对话服务入口
- `main()` - 对话服务启动函数
  - 初始化配置、日志、数据库
  - 会话管理API
  - 消息管理API
  - 流式对话支持
  - 端口8082

**kb/main.go** - 知识库服务入口
- `main()` - 知识库服务启动函数
  - 初始化配置、日志、数据库
  - 知识库管理API
  - RAG检索API
  - 向量化服务
  - 端口8086

**agent/main.go** - 助手服务入口
- `main()` - 助手服务启动函数
  - 初始化配置、日志、数据库
  - AI助手管理API
  - 助手市场API
  - 端口8085

**server/main_example.go** - 单体服务入口(开发环境)
- 集成所有微服务功能
- 适用于本地开发调试
- 单一端口启动所有服务

#### 适配器系统

**adapter.go** - 适配器接口和基础实现
- `type Adapter interface` - 适配器统一接口
  - `Name()` - 获取适配器名称
  - `GetSupportedModels()` - 获取支持的模型列表
  - `ConvertRequest()` - 转换请求格式为提供商格式
  - `DoRequest()` - 发送HTTP请求到提供商
  - `ParseResponse()` - 解析提供商响应为标准格式
  - `ParseStreamResponse()` - 解析流式响应
  - `ExtractUsage()` - 提取Token使用量
  - `GetError()` - 获取错误信息
  - `HealthCheck()` - 健康检查
- `type BaseAdapter struct` - 基础适配器实现
  - `NewBaseAdapter()` - 创建基础适配器
  - `NewRequest()` - 创建HTTP请求
  - `DoHTTPRequest()` - 执行HTTP请求
  - `addAuthHeader()` - 添加认证头
- 数据结构: OpenAIRequest, OpenAIResponse, StreamChunk, Message, Usage

**factory.go** - 适配器配置管理器
- `type ConfigManager struct` - 从数据库加载适配器配置
  - `NewConfigManager()` - 创建配置管理器
  - `Initialize()` - 从数据库初始化所有配置
  - `GetAdapter()` - 动态创建适配器实例
  - `ListAdapters()` - 列出所有可用适配器
  - `ReloadConfig()` - 热更新单个配置
  - `ReloadAllConfigs()` - 重新加载所有配置
  - `GetConfig()` - 获取适配器配置
  - `IsInitialized()` - 检查是否已初始化

**registry.go** - 适配器注册表
- `type AdapterRegistry struct` - 适配器注册中心
  - `NewAdapterRegistry()` - 创建注册表
  - `Register()` - 注册适配器工厂函数
  - `Unregister()` - 卸载适配器
  - `Update()` - 热更新适配器（支持热插拔）
  - `Create()` - 创建适配器实例
  - `GetVersion()` - 获取适配器版本
  - `List()` - 列出所有已注册适配器
- 全局函数:
  - `GetGlobalRegistry()` - 获取全局注册表单例
  - `CreateAdapter()` - 使用全局注册表创建适配器
  - `RegisterAdapter()` - 向全局注册表注册
  - `registerCoreAdapters()` - 注册核心提供商(OpenAI/Claude/Gemini/Baidu/Qwen)
  - `registerBatchAdapters()` - 注册批量提供商(DeepSeek/Moonshot/MiniMax)

**providers.go** - AI提供商具体实现
- `type OpenAIAdapter` - OpenAI适配器
  - `NewOpenAIAdapter()` - 创建实例，支持gpt-4/gpt-3.5系列
  - `ConvertRequest()` - 请求格式转换（无转换，直接使用）
  - `DoRequest()` - 调用OpenAI API
  - `ParseResponse()` - 解析JSON响应
  - `ParseStreamResponse()` - 解析SSE流式响应
  - `ExtractUsage()` - 提取Token使用量
- `type ClaudeAdapter` - Anthropic Claude适配器
  - `NewClaudeAdapter()` - 创建实例，支持claude-3系列
  - `ConvertRequest()` - 转换为Claude Messages格式
  - `DoRequest()` - 调用Claude API
  - `ParseResponse()` - 解析并转换为OpenAI格式
  - `ExtractUsage()` - 从usage字段提取input_tokens/output_tokens
- `type GeminiAdapter` - Google Gemini适配器
  - `NewGeminiAdapter()` - 创建实例，支持gemini-pro系列
  - `ConvertRequest()` - 转换为Gemini格式（contents和generation_config）
  - `DoRequest()` - 调用Gemini API
  - `ParseResponse()` - 解析并转换响应格式
- `type BaiduAdapter` - 百度文心一言适配器
  - `NewBaiduAdapter()` - 创建实例，支持eb系列模型
  - `ConvertRequest()` - 转换请求格式
- `type QwenAdapter` - 阿里通义千问适配器
  - `NewQwenAdapter()` - 创建实例，支持qwen系列模型
  - `ConvertRequest()` - 转换请求格式

**batch_providers.go** - 批量提供商实现
- `type DeepSeekAdapter` - DeepSeek适配器（代码模型）
  - `NewDeepSeekAdapter()` - 支持deepseek-coder/deepseek-chat
- `type MoonshotAdapter` - Moonshot(月之暗面)适配器
  - `NewMoonshotAdapter()` - 支持moonshot-v1系列(8k/32k/128k)
- `type MinimaxAdapter` - MiniMax适配器
  - `NewMinimaxAdapter()` - 支持abab6.5系列模型
- `type GenericAdapter` - 通用适配器（可配置映射）
  - `NewGenericAdapter()` - 创建可自定义转换的通用适配器
  - 支持自定义字段映射和转换函数

#### 渠道选择器

**selector.go** - 渠道选择器核心逻辑
- `type DefaultChannelSelector struct` - 默认选择器实现
  - `NewDefaultChannelSelector()` - 创建选择器实例
  - `registerStrategies()` - 注册所有选择策略
  - `Select()` - 选择最优渠道（单次）
    - 从缓存获取可用渠道列表
    - 过滤排除的渠道
    - 应用选择策略
  - `SelectWithRetry()` - 带重试的选择（最多3次）
    - 失败后自动排除故障渠道
    - 累计失败计数
  - `UpdateStats()` - 更新渠道统计信息
  - `GetStats()` - 获取渠道统计
  - `MarkChannelFailed()` - 标记渠道失败
    - 记录失败次数
    - 失败率>50%自动禁用渠道
  - `RefreshCache()` - 刷新渠道缓存
  - `filterExcludedChannels()` - 过滤排除的渠道
  - `disableChannel()` - 禁用故障渠道

**strategies.go** - 5种选择策略实现
- `selectByWeight()` - 权重策略
  - 计算总权重
  - 根据权重随机选择（加权随机）
- `selectByPriority()` - 优先级策略
  - 选择最高优先级的渠道
  - 相同优先级时使用权重策略
- `selectByRoundRobin()` - 轮询策略
  - 使用原子计数器循环选择
  - 保证负载均衡
- `selectByLowestLatency()` - 最低延迟策略
  - 从统计信息获取平均响应时间
  - 选择延迟最低的渠道
  - 无统计数据时退化为权重策略
- `selectByRandom()` - 随机策略
  - 完全随机选择，不考虑权重

**types.go** - 类型定义
- `type SelectRequest` - 选择请求
  - Model: 请求的模型名称
  - Strategy: 选择策略
  - ExcludeIDs: 排除的渠道ID列表
- `type SelectResult` - 选择结果
  - Channel: 选中的渠道
  - TotalAttempts: 总尝试次数
  - FailedCount: 失败次数
- `type SelectStrategy` - 策略枚举（weight/priority/round_robin/lowest_latency/random）

**cache.go** - 渠道缓存管理
- `type ChannelCache` - 渠道缓存接口
  - `GetAvailableChannels()` - 获取指定模型的可用渠道
  - `Refresh()` - 刷新缓存
  - `Invalidate()` - 失效缓存

**stats.go** - 统计管理
- `type StatsManager` - 统计管理器
  - `UpdateStats()` - 更新渠道统计（成功/失败、响应时间）
  - `GetStats()` - 获取渠道统计信息
  - `RecordFailure()` - 记录失败
  - 统计指标: 总请求数、成功数、失败数、平均响应时间

**retry.go** - 重试机制
- 重试配置和逻辑
- 指数退避策略
- 可重试错误判断

#### Token计数器

**factory.go** - Token计数器工厂
- `type TokenizerFactory struct` - 计数器工厂
  - `NewTokenizerFactory()` - 创建工厂实例
  - `GetTokenizer()` - 获取指定模型的计数器
    - OpenAI模型使用Tiktoken
    - 其他模型使用通用计数器
  - `CreateStreamCounter()` - 创建流式Token计数器
  - `CreateBatchStreamCounter()` - 创建批量流式计数器
  - `isOpenAIModel()` - 判断是否为OpenAI模型
  - `getGenericTokenizer()` - 获取通用计数器（带缓存）
  - `Close()` - 释放资源
- 全局函数:
  - `GetGlobalFactory()` - 获取全局工厂单例
  - `CountTokensQuick()` - 快速Token计数（使用全局工厂）

**tiktoken.go** - Tiktoken实现
- `type TiktokenTokenizer` - 基于tiktoken的精确计数
  - `NewTiktokenTokenizer()` - 创建实例
  - `CountTokens()` - 计算文本Token数
  - `CountMessages()` - 计算消息列表Token数
  - `GetEncoding()` - 获取编码器（cl100k_base/p50k_base等）
  - 支持GPT-4、GPT-3.5等OpenAI模型

**models.go** - 模型编码器映射
- OpenAI模型到编码器的映射表
  - gpt-4系列 → cl100k_base
  - gpt-3.5系列 → cl100k_base
  - text-davinci-003 → p50k_base
  - 默认编码器配置

**counter.go** - 计数器接口
- `type Tokenizer interface` - 计数器统一接口
  - `CountTokens(text string) int` - 计算文本Token数
  - `CountMessages(messages []Message) int` - 计算消息Token数
- `type StreamTokenCounter interface` - 流式计数器接口
  - `AddChunk(chunk string)` - 添加流式数据块
  - `GetCurrentCount() int` - 获取当前计数
  - `Finalize() int` - 完成计数
- `type GenericTokenizer` - 通用计数器（按字符估算）
  - 中文: 1字符 ≈ 1.5 tokens
  - 英文: 4字符 ≈ 1 token

#### 配额和计费

**quota/service.go** - 配额服务核心实现
- `type DefaultQuotaService struct` - 默认配额服务
  - `NewDefaultQuotaService()` - 创建服务实例
  - **预扣费流程:**
    - `PreConsumeQuota()` - 预扣除配额
      - 检查用户余额是否充足
      - 信任优化: 余额充足时不实际扣费
      - 记录预扣费到Redis缓存（15分钟过期）
      - 返回预扣状态和剩余余额
  - **退款流程:**
    - `ReturnPreConsumedQuota()` - 归还预扣费（请求失败时）
      - 从缓存获取预扣记录
      - 退还配额到用户账户
      - 删除预扣记录
  - **后扣费流程:**
    - `PostConsumeQuota()` - 实际消费扣费
      - 计算实际消耗与预扣的差额
      - 补扣或退还差额
      - 记录消费日志到unified_logs表
      - 失效用户余额缓存
  - `RefundQuota()` - 主动退款
  - `GetUserBalance()` - 获取用户余额（带缓存）
  - `GetPreConsumedRecord()` - 获取预扣费记录
  - `deductQuota()` - 扣除配额（原子操作，使用WHERE条件防止超扣）
  - `refundQuota()` - 退还配额

**quota/calculator.go** - 配额计算器
- `type QuotaCalculator interface` - 计算器接口
  - `Calculate()` - 计算配额消耗
  - 计算公式: quota = (prompt_tokens * model_ratio + completion_tokens * model_ratio * completion_ratio) * group_ratio
- 支持按Token和按次两种计费模式

**quota/cache.go** - 配额缓存
- `type QuotaCache interface` - 缓存接口
  - `GetUserBalance()` - 获取用户余额缓存
  - `SetUserBalance()` - 设置用户余额缓存
  - `InvalidateUserBalance()` - 失效用户余额缓存
  - `SetPreConsumed()` - 设置预扣费记录
  - `GetPreConsumed()` - 获取预扣费记录  
  - `DeletePreConsumed()` - 删除预扣费记录
- 使用Redis实现，TTL 15分钟

**quota/types.go** - 类型定义
- `type PreConsumeRequest` - 预扣费请求
- `type PreConsumeResponse` - 预扣费响应
- `type PostConsumeRequest` - 后扣费请求
- `type RefundRequest` - 退款请求
- `type PreConsumedRecord` - 预扣费记录

#### 流式处理

**stream_handler.go** - 流式请求处理器
- `type StreamHandler struct` - 流式处理器
  - `NewStreamHandler()` - 创建处理器实例
  - **核心方法:**
    - `HandleStreamResponse()` - 处理流式响应
      - 设置总超时（默认5分钟）和空闲超时（默认30秒）
      - 实时Token计数（逐块累加）
      - 发送SSE数据块到客户端
      - defer执行后扣费（确保计费）
      - 错误处理和超时控制
    - `HandleStreamWithRetry()` - 带重试的流式处理
      - 支持最多N次重试
      - 指数退避策略
      - 可重试错误判断
  - `isRetryableError()` - 判断错误是否可重试
    - 超时、连接重置、503/502/504等
- `type StreamOptions` - 流式处理选项
  - RequestID, UserID, ChannelID
  - Model, PromptTokens, MaxTokens
  - TotalTimeout, IdleTimeout
- `type StreamResult` - 流式处理结果
  - PromptTokens, CompletionTokens, TotalTokens
  - Duration, ChunkCount

**stream_sender.go** - SSE发送器
- `type StreamSender struct` - SSE数据发送器
  - `NewStreamSender()` - 创建发送器
  - `Send()` - 发送数据块
    - 格式化为SSE格式: "data: {json}\n\n"
    - 调用Flush()立即推送
  - `SendError()` - 发送错误信息
  - `SendDone()` - 发送完成标记 "data: [DONE]\n\n"
  - `SetHeaders()` - 设置SSE响应头
    - Content-Type: text/event-stream
    - Cache-Control: no-cache
    - Connection: keep-alive

**stream_monitor.go** - 流式监控
- `type StreamMonitor` - 流式性能监控
  - `RecordChunk()` - 记录数据块
  - `GetMetrics()` - 获取监控指标
  - 监控指标: 总块数、总字节数、平均块大小、吞吐量

**types.go** - 请求响应类型
- `type ChatCompletionRequest` - 聊天补全请求
  - Model, Messages, Temperature, MaxTokens
  - TopP, FrequencyPenalty, PresencePenalty
  - Stream, Tools, User
- `type ChatCompletionResponse` - 聊天补全响应
  - ID, Object, Created, Model
  - Choices, Usage
- `type ChatMessage` - 聊天消息
  - Role (system/user/assistant)
  - Content
- `type StreamChunk` - 流式数据块
  - Delta增量内容

#### 业务服务

**pricing_service.go** - 模型定价服务
- `type PricingService interface` - 定价服务接口
- `type DefaultPricingService struct` - 默认实现
  - `NewPricingService()` - 创建服务实例
  - **CRUD操作:**
    - `GetPricing()` - 获取模型定价
      - 支持按模型和用户分组查询
      - 二级缓存（内存缓存，5分钟TTL）
      - 找不到时fallback到default分组
    - `ListPricing()` - 列出所有定价
      - 支持按enabled状态过滤
      - 按模型名和分组排序
    - `CreatePricing()` - 创建定价配置
      - 检查重复
      - 自动刷新缓存
    - `UpdatePricing()` - 更新定价
    - `DeletePricing()` - 软删除定价
  - **计费计算:**
    - `CalculateQuota()` - 计算配额消耗
      - 获取模型定价
      - 应用分组倍率
      - 公式: (prompt_tokens + completion_tokens * completion_ratio) * model_ratio * group_ratio
  - **缓存管理:**
    - `RefreshCache()` - 刷新定价缓存
      - 从数据库重新加载所有启用的定价
      - 重建内存缓存
  - **分组倍率:**
    - `SetGroupRatio()` - 设置用户分组倍率
    - `GetGroupRatio()` - 获取分组倍率
    - 预设倍率: default(1.0), vip(0.8), premium(0.6), free(1.5)

**channel_ability_service.go** - 渠道能力服务
- `type ChannelAbilityService interface` - 能力服务接口
- `type DefaultChannelAbilityService struct` - 默认实现
  - `NewChannelAbilityService()` - 创建服务实例
  - **能力同步:**
    - `SyncFromChannel()` - 从渠道同步能力
      - 解析支持的模型列表
      - 构建能力记录（ChannelAbility）
      - 批量更新数据库
      - 清空缓存
  - **查询方法:**
    - `FindByModelAndGroup()` - 查找指定模型和分组的渠道能力
      - 带缓存（5分钟TTL）
      - 缓存Key: {model}_{group}
    - `GetAvailableChannelsForModel()` - 获取可用渠道列表
      - 只返回enabled=true的渠道
      - 按优先级和权重排序
  - **管理操作:**
    - `DeleteByChannel()` - 删除渠道的所有能力记录
    - `invalidateCache()` - 清空所有缓存
- 辅助函数:
  - `ParseSupportedModels()` - 解析逗号分隔的模型列表

**health_check_service.go** - 健康检查服务
- `type HealthCheckService interface` - 健康检查接口
- `type DefaultHealthCheckService struct` - 默认实现
  - `NewHealthCheckService()` - 创建服务实例
  - **定期检查:**
    - `StartPeriodicCheck()` - 启动定期健康检查
      - 每30分钟自动检查一次
      - 并发检查多个渠道（最多5个并发）
    - `checkAllChannels()` - 检查所有启用的渠道
  - **单渠道检查:**
    - `CheckChannel()` - 检查单个渠道
      - 创建测试请求（"Hi"，最多5 tokens）
      - 调用渠道适配器
      - 记录响应时间
      - 返回成功/失败状态
    - `handleCheckResult()` - 处理检查结果
      - 成功时重置失败计数
      - 失败时累加计数
      - 连续失败3次自动禁用渠道
      - 恢复后自动启用
  - **健康评分:**
    - `CalculateHealthScore()` - 计算健康度评分（0-100分）
      - 成功率权重70%
      - 响应速度权重30%
      - 基于最近48次检查记录
    - `GetHealthStatus()` - 获取健康状态
      - 状态: healthy/degraded/unhealthy
      - 包含失败计数和最后检查时间
  - **内部方法:**
    - `incrementFailureCount()` - 增加失败计数
    - `resetFailureCount()` - 重置失败计数
    - `saveCheckResult()` - 保存检查记录（最多保留100条）
- 数据结构:
  - `HealthCheckResult` - 检查结果
  - `HealthScore` - 健康评分
  - `HealthStatus` - 健康状态

#### HTTP处理器

**channel_handler.go** - 渠道管理API处理器
- `type ChannelHandler struct` - 渠道处理器
  - `NewChannelHandler()` - 创建处理器实例
  - **渠道CRUD:**
    - `ListChannels()` - 分页查询渠道 [GET /api/admin/channels]
      - 查询参数: page, page_size, type, group, status, enabled
      - 返回分页结果
    - `CreateChannel()` - 创建渠道 [POST /api/admin/channels]
      - 验证必填字段: name, type, api_keys, support_models
      - 设置默认值: group(default), priority(100), weight(10)
      - 自动同步渠道能力
    - `UpdateChannel()` - 更新渠道 [PUT /api/admin/channels/:id]
      - 支持部分更新
      - support_models变更时重新同步能力
    - `DeleteChannel()` - 删除渠道 [DELETE /api/admin/channels/:id]
      - 软删除
      - 自动删除关联的能力记录
  - **渠道测试:**
    - `TestChannel()` - 测试渠道连接 [POST /api/admin/channels/:id/test]
      - 调用健康检查服务
      - 返回延迟和状态
  - **批量操作:**
    - `BatchOperation()` - 批量操作 [POST /api/admin/channels/batch]
      - 操作类型: enable, disable, delete
      - 返回成功/失败统计
  - `RegisterRoutes()` - 注册路由
- 请求/响应结构:
  - ListChannelsRequest, ListChannelsResponse
  - CreateChannelRequest, UpdateChannelRequest
  - BatchOperationRequest

**pricing_handler.go** - 定价管理API处理器
- `type PricingHandler struct` - 定价处理器
  - `NewPricingHandler()` - 创建处理器实例
  - **定价CRUD:**
    - `ListPricing()` - 列出所有定价 [GET /api/v1/pricing]
      - 查询参数: enabled
    - `GetPricing()` - 获取模型定价 [GET /api/v1/pricing/:model]
      - 查询参数: group (默认default)
    - `CreatePricing()` - 创建定价 [POST /api/v1/pricing]
      - 字段: model, group, quota_type, model_ratio, completion_ratio, group_ratio
      - 默认值: group(default), completion_ratio(1.0), group_ratio(1.0)
    - `UpdatePricing()` - 更新定价 [PUT /api/v1/pricing/:id]
      - 支持部分更新
    - `DeletePricing()` - 删除定价 [DELETE /api/v1/pricing/:id]
      - 软删除
  - **配额计算:**
    - `CalculateQuota()` - 计算配额 [POST /api/v1/pricing/calculate]
      - 请求: model, group, prompt_tokens, completion_tokens
      - 响应: quota, group_ratio
  - **缓存管理:**
    - `RefreshCache()` - 刷新缓存 [POST /api/v1/pricing/refresh]
  - `RegisterRoutes()` - 注册路由
- 请求/响应结构:
  - CreatePricingRequest, UpdatePricingRequest
  - CalculateQuotaRequest, CalculateQuotaResponse

**stats_handler.go** - 统计监控API
- `type StatsHandler struct` - 统计处理器
  - `GetOverview()` - 获取总览统计 [GET /api/admin/stats/overview]
    - 总用户数、总会话数、总消息数
    - 今日请求数、Token消耗
  - `GetChannelStats()` - 渠道统计 [GET /api/admin/stats/channels]
    - 按渠道分组统计
    - 请求数、成功率、平均延迟
    - 时间范围: days参数
  - `GetModelStats()` - 模型统计 [GET /api/admin/stats/models]
    - 按模型分组统计
    - 使用频率、Token消耗
  - `GetUserStats()` - 用户统计 [GET /api/admin/stats/users]
    - Top用户排行
    - 消费统计

**health_handler.go** - 健康检查API
- `type HealthHandler struct` - 健康检查处理器
  - `CheckChannel()` - 检查渠道健康 [POST /api/admin/health/channels/:id]
    - 立即执行健康检查
  - `GetChannelHealth()` - 获取渠道健康状态 [GET /api/admin/health/channels/:id/status]
    - 返回最后检查结果
  - `GetHealthScore()` - 获取健康评分 [GET /api/admin/health/channels/:id/score]
    - 返回0-100评分
  - `CheckAllChannels()` - 检查所有渠道 [POST /api/admin/health/check-all]
    - 并发检查所有启用的渠道

#### 数据模型
```
backend/internal/model/user.go            - 用户模型
backend/internal/model/channel.go         - 渠道模型
backend/internal/model/adapter_config.go  - 适配器配置
backend/internal/model/channel_ability.go - 渠道能力
backend/internal/model/unified_log.go     - 统一日志
backend/internal/model/model_pricing.go   - 模型定价
backend/internal/model/session.go         - 会话模型
backend/internal/model/message.go         - 消息模型
```

#### 中间件
```
backend/internal/middleware/auth.go       - JWT认证
backend/internal/middleware/cors.go       - 跨域处理
backend/internal/middleware/logger.go     - 日志记录
backend/internal/middleware/ratelimit.go  - 限流
backend/internal/middleware/recovery.go   - 错误恢复
```

### 7.2 前端核心文件

#### 页面组件 (Next.js App Router)

**layout.tsx** - 根布局组件
- 全局布局结构
- 包含头部导航、侧边栏、主内容区
- 集成Provider (状态管理、主题)
- 设置元数据和字体

**page.tsx** - 首页
- 欢迎页面
- 功能介绍
- 快速开始引导
- CTA按钮（开始聊天、浏览助手）

**chat/page.tsx** - 聊天页面
- 对话界面主页面
- 集成ChatBox组件
- 会话列表（SessionSidebar）
- 消息渲染（MessageRenderer）
- 支持流式响应
- 实时Token统计

**admin/page.tsx** - 管理后台
- 渠道管理
- 定价配置
- 用户管理
- 统计报表
- 系统设置

**developer/page.tsx** - 开发者控制台
- API密钥管理（TokenManagementTab）
- 使用统计（UsageStatsTab）
- 开发者工具（DeveloperConsole）
- API文档链接

**login/page.tsx** - 登录页面
- 用户登录表单
- JWT认证
- 第三方登录集成

**register/page.tsx** - 注册页面
- 用户注册表单
- 邮箱验证
- 用户协议

**user/page.tsx** - 用户中心
- 个人信息
- 配额管理
- 账单历史
- 设置偏好

#### 核心组件

**ChatBox.tsx** - 聊天框主组件
- 完整的聊天界面容器
- 集成消息列表和输入区
- 管理会话状态
- 处理消息发送和接收
- 支持Markdown和代码高亮
- 支持文件上传

**SessionSidebar.tsx** - 会话侧边栏
- 会话列表显示
- 新建会话
- 切换会话
- 删除会话
- 搜索会话
- 会话分组

**MessageRenderer.tsx** - 消息渲染器
- Markdown渲染
- 代码块语法高亮
- 数学公式渲染(KaTeX)
- 表格渲染
- 图表渲染(Mermaid)
- 消息操作（复制、编辑、删除、重新生成）

**DeveloperConsole.tsx** - 开发者控制台
- API密钥管理
- 请求日志查看
- 调试工具
- 性能监控

**TokenManagementTab.tsx** - Token管理标签页
- 显示API密钥列表
- 创建新密钥
- 删除/禁用密钥
- 使用统计

**UsageStatsTab.tsx** - 使用统计标签页
- 请求量统计
- Token消耗统计
- 费用统计
- 图表可视化

**MessageExport.tsx** - 消息导出
- 导出为Markdown
- 导出为PDF
- 导出为JSON
- 选择性导出

**SessionShare.tsx** - 会话分享
- 生成分享链接
- 权限控制
- 过期时间设置

#### UI基础组件 (components/ui/)

**Button.tsx** - 按钮组件
- 多种尺寸: sm, md, lg
- 多种变体: primary, secondary, outline, ghost
- 加载状态
- 禁用状态
- 图标支持

**Input.tsx** - 输入框组件
- 文本输入
- 密码输入
- 验证状态
- 错误提示
- 前缀/后缀支持

**Modal.tsx** - 模态框组件
- 可定制标题和内容
- 确认/取消按钮
- 关闭按钮
- 遮罩层
- 动画效果

**Card.tsx** - 卡片组件
- 标题和内容区
- 可定制样式
- 阴影效果

**Table.tsx** - 表格组件
- 数据表格
- 排序支持
- 分页支持
- 行选择

**Select.tsx** - 下拉选择组件
- 单选/多选
- 搜索过滤
- 自定义选项渲染

**Spinner.tsx** - 加载动画
- 多种尺寸
- 自定义颜色

**Progress.tsx** - 进度条
- 线性进度条
- 环形进度条
- 自定义颜色

**CodeBlock.tsx** - 代码块组件
- 语法高亮(highlight.js)
- 复制按钮
- 行号显示
- 主题切换

**FileUpload.tsx** - 文件上传组件
- 拖拽上传
- 多文件上传
- 进度显示
- 文件类型限制

**StatCard.tsx** - 统计卡片
- 显示统计数字
- 趋势图标
- 百分比变化

**Alert.tsx** - 警告提示
- 多种类型: success, warning, error, info
- 可关闭
- 图标支持

**Pagination.tsx** - 分页组件
- 页码显示
- 上一页/下一页
- 跳转到指定页

#### 状态管理 (Zustand)

**store/authStore.ts** - 认证状态管理
- State:
  - `user` - 当前用户信息
  - `token` - JWT token
  - `isAuthenticated` - 是否已认证
  - `isLoading` - 加载状态
- Actions:
  - `login()` - 登录
    - 调用登录API
    - 保存token到localStorage
    - 更新用户状态
  - `logout()` - 登出
    - 清除token
    - 清空用户信息
    - 跳转到登录页
  - `register()` - 注册
  - `refreshToken()` - 刷新token
  - `updateProfile()` - 更新用户信息
  - `checkAuth()` - 检查认证状态

**stores/chatStore.ts** - 聊天状态管理
- State:
  - `sessions` - 会话列表
  - `currentSession` - 当前会话
  - `messages` - 当前会话消息列表
  - `isStreaming` - 是否正在流式响应
  - `streamingMessage` - 流式消息累积
- Actions:
  - **会话管理:**
    - `fetchSessions()` - 获取会话列表
    - `createSession()` - 创建新会话
    - `selectSession()` - 切换会话
    - `updateSession()` - 更新会话信息
    - `deleteSession()` - 删除会话
  - **消息管理:**
    - `sendMessage()` - 发送消息
      - 添加用户消息到列表
      - 调用聊天API
      - 处理流式响应
    - `sendStreamingMessage()` - 发送流式消息
      - 建立SSE连接
      - 实时更新streamingMessage
      - 完成后添加到messages
    - `regenerateMessage()` - 重新生成消息
    - `editMessage()` - 编辑消息
    - `deleteMessage()` - 删除消息
  - **流式处理:**
    - `startStreaming()` - 开始流式响应
    - `appendStreamChunk()` - 追加流式数据块
    - `completeStreaming()` - 完成流式响应
    - `cancelStreaming()` - 取消流式响应

**stores/settingsStore.ts** - 设置状态管理
- State:
  - `theme` - 主题 (light/dark/system)
  - `language` - 语言
  - `model` - 默认模型
  - `temperature` - 温度参数
  - `maxTokens` - 最大Token数
  - `systemPrompt` - 系统提示词
  - `contextLength` - 上下文长度
- Actions:
  - `updateTheme()` - 更新主题
  - `updateLanguage()` - 更新语言
  - `updateModelSettings()` - 更新模型设置
  - `resetSettings()` - 重置为默认设置
  - `loadSettings()` - 从localStorage加载设置
  - `saveSettings()` - 保存设置到localStorage

**stores/assistantStore.ts** - 助手状态管理
- State:
  - `assistants` - 助手列表
  - `currentAssistant` - 当前选中的助手
  - `categories` - 助手分类
- Actions:
  - `fetchAssistants()` - 获取助手列表
  - `selectAssistant()` - 选择助手
  - `installAssistant()` - 安装助手
  - `uninstallAssistant()` - 卸载助手
  - `createCustomAssistant()` - 创建自定义助手

**stores/knowledgeStore.ts** - 知识库状态管理
- State:
  - `knowledgeBases` - 知识库列表
  - `currentKB` - 当前知识库
  - `documents` - 文档列表
- Actions:
  - `fetchKnowledgeBases()` - 获取知识库列表
  - `createKnowledgeBase()` - 创建知识库
  - `uploadDocument()` - 上传文档
  - `searchDocuments()` - 搜索文档
  - `deleteDocument()` - 删除文档

#### API服务

**api.ts** - API客户端基础
- `apiClient` - Axios实例配置
  - BaseURL配置
  - 请求拦截器（添加JWT token）
  - 响应拦截器（处理错误、刷新token）
  - 超时配置
- 通用API方法:
  - `get()` - GET请求
  - `post()` - POST请求
  - `put()` - PUT请求
  - `delete()` - DELETE请求
- 错误处理:
  - 401 自动跳转登录
  - 403 权限不足提示
  - 500 服务器错误提示
  - 网络错误重试

**sse.ts** - SSE流式通信
- `SSEClient` - SSE客户端类
  - `connect()` - 建立SSE连接
  - `onMessage()` - 消息回调
  - `onError()` - 错误回调
  - `onComplete()` - 完成回调
  - `close()` - 关闭连接
- 自动重连机制
- 心跳检测
- 断线重连

**streaming.ts** - 流式响应处理
- `StreamHandler` - 流式数据处理器
  - `handleChunk()` - 处理数据块
  - `parseSSE()` - 解析SSE格式
  - `accumulate()` - 累积完整消息
- Token实时计数
- 进度回调

**websocket.ts** - WebSocket通信
- `WebSocketClient` - WebSocket客户端
  - `connect()` - 建立连接
  - `send()` - 发送消息
  - `onMessage()` - 消息监听
  - `close()` - 关闭连接
- 心跳保活
- 自动重连
- 消息队列

**upload.ts** - 文件上传服务
- `uploadFile()` - 上传单个文件
  - 支持FormData
  - 进度回调
  - 断点续传
- `uploadMultiple()` - 批量上传
- `uploadChunk()` - 分块上传（大文件）
- `cancelUpload()` - 取消上传

### 7.3 配置文件

#### 后端配置
```
backend/go.mod                - Go依赖管理
backend/go.sum                - 依赖校验和
backend/Dockerfile            - Docker镜像构建
backend/.env.example          - 环境变量示例
```

#### 前端配置
```
frontend/package.json         - NPM依赖管理
frontend/tsconfig.json        - TypeScript配置
frontend/tailwind.config.js   - TailwindCSS配置
frontend/next.config.js       - Next.js配置
frontend/.eslintrc.json       - ESLint配置
```

#### 部署配置
```
docker-compose.yml                    - Docker编排
deploy/kubernetes/backend.yaml        - 后端K8s配置
deploy/kubernetes/frontend.yaml       - 前端K8s配置
deploy/kubernetes/ingress.yaml        - Ingress配置
deploy/helm/values.yaml               - Helm配置
```

### 7.4 数据库迁移

```
backend/migrations/000001_create_users_table.up.sql           - 用户表
backend/migrations/000002_create_user_settings_table.up.sql   - 用户设置表
backend/migrations/000003_create_quota_logs_table.up.sql      - 配额日志表
backend/migrations/000004_create_sessions_table.up.sql        - 会话表
backend/migrations/000005_create_messages_table.up.sql        - 消息表
backend/migrations/000006_create_channels_table.up.sql        - 渠道表
backend/migrations/000007_create_billing_tables.up.sql        - 计费表
backend/migrations/000013_create_adapter_configs.up.sql       - 适配器配置表
backend/migrations/000014_create_channel_abilities.up.sql     - 渠道能力表
backend/migrations/000015_create_unified_logs.up.sql          - 统一日志表
backend/migrations/000016_create_model_pricing.up.sql         - 模型定价表
```

---

## 8. 数据流转

### 8.1 用户对话流程

```
┌─────────┐
│  用户   │
└────┬────┘
     │ 1. 发送消息
     ▼
┌─────────────┐
│   前端      │
│  (React)    │
└────┬────────┘
     │ 2. HTTP POST /api/chat/completions
     ▼
┌─────────────┐
│ API Gateway │
│  - JWT认证  │
│  - 限流检查 │
└────┬────────┘
     │ 3. 转发到Chat Service
     ▼
┌─────────────┐
│ Chat Service│
│  - 创建会话 │
│  - 保存消息 │
└────┬────────┘
     │ 4. 调用Billing检查配额
     ▼
┌──────────────┐
│Billing Service│
│  - 检查余额  │
│  - 预扣费    │
└────┬─────────┘
     │ 5. 配额充足,继续
     ▼
┌─────────────┐
│Relay Service│
│  - 选择渠道 │
│  - 调用适配器│
└────┬────────┘
     │ 6. 调用上游AI API
     ▼
┌─────────────┐
│  OpenAI/    │
│  Claude等   │
└────┬────────┘
     │ 7. 流式返回
     ▼
┌─────────────┐
│Stream Handler│
│  - 实时计数 │
│  - SSE推送  │
└────┬────────┘
     │ 8. 推送到客户端
     ▼
┌─────────────┐
│   前端      │
│  (显示响应) │
└─────────────┘
     │
     ▼
┌──────────────┐
│Billing Service│
│  - 后扣费    │
│  - 记录日志  │
└──────────────┘
     │
     ▼
┌─────────────┐
│  RabbitMQ   │
│ (异步账单)  │
└─────────────┘
```

### 8.2 渠道选择流程

```
1. 接收请求
   └─> 提取模型信息 (如: gpt-4o)

2. 查询可用渠道
   └─> 从缓存获取支持该模型的渠道列表
   └─> 过滤禁用渠道

3. 应用选择策略
   ├─> 权重策略: 按权重随机选择
   ├─> 优先级策略: 选择最高优先级
   ├─> 轮询策略: 循环选择
   ├─> 最低延迟: 选择响应最快的
   └─> 随机策略: 完全随机

4. 调用选中渠道
   └─> 通过适配器转换请求格式
   └─> 发送HTTP请求

5. 处理响应
   ├─> 成功: 更新统计信息
   └─> 失败: 
       ├─> 标记失败
       ├─> 排除该渠道
       └─> 重试其他渠道 (最多3次)

6. 返回结果
```

### 8.3 计费流程

```
预扣费阶段:
┌────────────────────────────────────┐
│ 1. 估算Token数                      │
│    - 根据历史数据估算              │
│    - 或使用固定值(如1000 tokens)   │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 2. 计算预估配额                    │
│    quota = tokens * model_ratio    │
│           * completion_ratio       │
│           * group_ratio            │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 3. 检查用户余额                    │
│    if balance < quota:             │
│       return error                 │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 4. 预扣除配额                      │
│    balance -= quota                │
│    记录预扣费日志                  │
└──────────────┬─────────────────────┘
               ▼
         【调用AI API】
               ▼
后扣费阶段:
┌────────────────────────────────────┐
│ 5. 精确计算实际消耗                │
│    actual_tokens = prompt_tokens   │
│                  + completion_tokens│
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 6. 计算实际配额                    │
│    actual_quota = actual_tokens    │
│                 * model_ratio      │
│                 * ...              │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 7. 调整配额                        │
│    diff = actual_quota - quota     │
│    balance -= diff                 │
│    记录后扣费日志                  │
└────────────────────────────────────┘
               ▼
失败退款:
┌────────────────────────────────────┐
│ 8. 如果API调用失败                 │
│    balance += quota                │
│    记录退款日志                    │
└────────────────────────────────────┘
```

### 8.4 健康检查流程

```
定时任务 (每30分钟):
┌────────────────────────────────────┐
│ 1. 获取所有启用的渠道              │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 2. 并发检查渠道健康                │
│    for each channel:               │
│      - 调用测试API                 │
│      - 记录响应时间                │
│      - 记录成功/失败               │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 3. 更新统计信息                    │
│    - 总请求数                      │
│    - 成功次数                      │
│    - 失败次数                      │
│    - 平均延迟                      │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 4. 计算健康评分                    │
│    score = success_rate * 0.7      │
│          + latency_score * 0.3     │
└──────────────┬─────────────────────┘
               ▼
┌────────────────────────────────────┐
│ 5. 自动运维决策                    │
│    if consecutive_failures >= 3:   │
│       disable_channel()            │
│    if recovered:                   │
│       enable_channel()             │
└────────────────────────────────────┘
```

---

## 9. 开发指南

### 9.1 本地开发环境搭建

#### 前置要求
```bash
- Go 1.23+
- Node.js 20+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose
```

#### 启动步骤

**1. 克隆项目**
```bash
git clone https://github.com/your-org/oblivious.git
cd oblivious
```

**2. 启动基础设施**
```bash
cd deploy
docker-compose up -d postgres redis minio rabbitmq
```

**3. 初始化数据库**
```bash
cd backend
# 执行所有迁移
for f in migrations/*.up.sql; do
  psql $DATABASE_URL < $f
done
```

**4. 启动后端服务**
```bash
# 方式1: 单体服务(开发)
cd backend
go run cmd/server/main_example.go

# 方式2: 微服务
cd backend/cmd/gateway && go run main.go &
cd backend/cmd/user && go run main.go &
cd backend/cmd/chat && go run main.go &
cd backend/cmd/relay && go run main.go &
```

**5. 启动前端**
```bash
cd frontend
npm install
npm run dev
```

**6. 访问应用**
- 前端: http://localhost:3000
- API网关: http://localhost:8080
- API文档: http://localhost:8080/swagger

### 9.2 添加新的AI提供商

**步骤1: 实现适配器**
```go
// backend/internal/adapter/my_provider.go
package adapter

type MyProviderAdapter struct {
    config *AdapterConfig
}

func (a *MyProviderAdapter) ChatCompletion(
    ctx context.Context,
    req *relay.ChatCompletionRequest,
) (*relay.ChatCompletionResponse, error) {
    // 1. 转换请求格式
    providerReq := convertRequest(req)
    
    // 2. 调用提供商API
    resp, err := callProviderAPI(providerReq)
    if err != nil {
        return nil, err
    }
    
    // 3. 转换响应格式
    return convertResponse(resp), nil
}

func (a *MyProviderAdapter) StreamChatCompletion(
    ctx context.Context,
    req *relay.ChatCompletionRequest,
) (<-chan *relay.StreamChunk, error) {
    // 实现流式响应
}
```

**步骤2: 注册适配器**
```go
// backend/internal/adapter/registry.go
func init() {
    registry.Register("my_provider", func(config *AdapterConfig) (Adapter, error) {
        return &MyProviderAdapter{config: config}, nil
    })
}
```

**步骤3: 添加到提供商列表**
```go
// backend/internal/adapter/providers.go
var SupportedProviders = []string{
    "openai",
    "anthropic",
    "google",
    "my_provider", // 新增
}
```

**步骤4: 创建渠道**
```bash
curl -X POST http://localhost:8080/api/admin/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的提供商渠道",
    "type": "my_provider",
    "api_base": "https://api.myprovider.com",
    "api_keys": "sk-xxx",
    "support_models": "my-model-1,my-model-2"
  }'
```

### 9.3 添加新的选择策略

**步骤1: 实现策略函数**
```go
// backend/internal/selector/strategies.go
func (s *DefaultChannelSelector) selectByCustom(
    ctx context.Context,
    channels []*model.Channel,
    req *SelectRequest,
) (*model.Channel, error) {
    // 实现自定义选择逻辑
    // 例如: 基于成本、基于地理位置等
    
    return selectedChannel, nil
}
```

**步骤2: 注册策略**
```go
// backend/internal/selector/selector.go
const (
    StrategyWeight        SelectStrategy = "weight"
    StrategyPriority      SelectStrategy = "priority"
    StrategyRoundRobin    SelectStrategy = "round_robin"
    StrategyLowestLatency SelectStrategy = "lowest_latency"
    StrategyRandom        SelectStrategy = "random"
    StrategyCustom        SelectStrategy = "custom" // 新增
)

func (s *DefaultChannelSelector) registerStrategies() {
    s.strategies[StrategyWeight] = s.selectByWeight
    s.strategies[StrategyPriority] = s.selectByPriority
    s.strategies[StrategyRoundRobin] = s.selectByRoundRobin
    s.strategies[StrategyLowestLatency] = s.selectByLowestLatency
    s.strategies[StrategyRandom] = s.selectByRandom
    s.strategies[StrategyCustom] = s.selectByCustom // 新增
}
```

### 9.4 常用命令

#### 后端命令
```bash
# 运行测试
go test ./...

# 测试覆盖率
go test ./... -cover

# 代码格式化
go fmt ./...

# 静态检查
go vet ./...

# 构建二进制
go build -o bin/oblivious cmd/server/main_example.go

# 运行迁移
psql $DATABASE_URL < migrations/xxx.up.sql

# 回滚迁移
psql $DATABASE_URL < migrations/xxx.down.sql
```

#### 前端命令
```bash
# 开发模式
npm run dev

# 生产构建
npm run build

# 启动生产服务器
npm run start

# 代码检查
npm run lint

# 类型检查
npm run type-check

# 运行测试
npm run test

# Storybook
npm run storybook
```

#### Docker命令
```bash
# 构建镜像
docker build -t oblivious-backend:latest -f backend/Dockerfile .
docker build -t oblivious-frontend:latest -f frontend/Dockerfile .

# 运行容器
docker run -p 8080:8080 oblivious-backend
docker run -p 3000:3000 oblivious-frontend

# Docker Compose
docker-compose up -d
docker-compose logs -f
docker-compose down
```

### 9.5 API测试示例

#### 创建渠道
```bash
curl -X POST http://localhost:8080/api/admin/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpenAI主渠道",
    "type": "openai",
    "api_keys": "sk-xxx",
    "support_models": "gpt-4o,gpt-3.5-turbo",
    "priority": 100,
    "weight": 10
  }'
```

#### 配置定价
```bash
curl -X POST http://localhost:8080/api/admin/pricing \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "quota_type": 0,
    "model_ratio": 15.0,
    "completion_ratio": 2.0
  }'
```

#### 查看统计
```bash
# 总览
curl http://localhost:8080/api/admin/stats/overview

# 渠道统计
curl http://localhost:8080/api/admin/stats/channels?days=7

# 模型统计
curl http://localhost:8080/api/admin/stats/models?days=7
```

#### 健康检查
```bash
# 检查单个渠道
curl -X POST http://localhost:8080/api/admin/health/channels/1

# 获取健康状态
curl http://localhost:8080/api/admin/health/channels/1/status

# 获取健康评分
curl http://localhost:8080/api/admin/health/channels/1/score
```

---

## 10. 附录

### 10.1 相关文档

- **README.md**: 项目总体介绍
- **PROJECT_SUMMARY.md**: 项目总结和进度
- **QUICK_START.md**: 快速启动指南
- **REFACTOR_PROGRESS.md**: 重构进度跟踪
- **backend/README.md**: 后端详细文档
- **docs/ARCHITECTURE.md**: 架构设计文档
- **docs/API_REFERENCE.md**: API参考文档
- **docs/DATABASE_DESIGN.md**: 数据库设计文档

### 10.2 技术决策

**为什么选择 Go?**
- 高性能并发处理
- 简洁的语法
- 丰富的生态
- 云原生友好

**为什么选择微服务?**
- 独立扩展
- 故障隔离
- 技术异构
- 团队协作

**为什么选择 PostgreSQL?**
- 功能强大
- pgvector支持
- 成熟稳定
- 开源免费

**为什么选择 Next.js?**
- SSR/SSG支持
- 优秀的开发体验
- React生态
- SEO友好

### 10.3 性能指标

**目标性能**:
- QPS: 10,000+
- 延迟: <100ms (p99)
- 可用性: 99.9%
- Token计数精度: >99%

**当前性能**:
- 单实例QPS: 1,000+
- 平均延迟: 50-100ms
- Token计数精度: >99%

### 10.4 路线图

**Phase 1: 核心功能** (已完成 30%)
- [x] 数据库架构
- [x] 适配器系统
- [x] 渠道选择器
- [x] Token计数
- [x] 配额服务
- [x] 流式处理
- [x] 定价系统
- [x] 健康检查
- [ ] 完整测试

**Phase 2: 前端开发** (进行中)
- [ ] 用户认证界面
- [ ] 聊天界面
- [ ] 助手市场
- [ ] 知识库管理
- [ ] 管理后台

**Phase 3: 企业级特性** (规划中)
- [ ] 限流系统
- [ ] 告警通知
- [ ] 日志聚合
- [ ] 链路追踪
- [ ] 性能优化

**Phase 4: 扩展功能** (未来)
- [ ] 多租户支持
- [ ] K8s自动扩展
- [ ] 插件市场
- [ ] 移动端应用

---

## 10. 完整函数索引

### 后端核心函数速查

#### Adapter模块
```go
// adapter.go
type Adapter interface {
    Name() string
    GetSupportedModels() []string
    ConvertRequest(*OpenAIRequest) (interface{}, error)
    DoRequest(context.Context, interface{}) (*http.Response, error)
    ParseResponse(*http.Response) (*OpenAIResponse, error)
    ParseStreamResponse(*http.Response) (<-chan *StreamChunk, error)
    ExtractUsage(interface{}) (*Usage, error)
    GetError(*http.Response) error
    HealthCheck(context.Context) error
}

type BaseAdapter struct {
    NewBaseAdapter(*AdapterConfig) *BaseAdapter
    Name() string
    GetSupportedModels() []string
    SetSupportedModels([]string)
    NewRequest(context.Context, string, string, interface{}) (*http.Request, error)
    DoHTTPRequest(context.Context, string, string, interface{}) (*http.Response, error)
    addAuthHeader(*http.Request)
}

// registry.go
type AdapterRegistry struct {
    Register(name, factory, version string) error
    Unregister(name string) error
    Update(name, factory, version string) error
    Create(name string, config *AdapterConfig) (Adapter, error)
    GetVersion(name string) (string, error)
    List() map[string]string
}

// factory.go
type ConfigManager struct {
    Initialize(context.Context) error
    GetAdapter(name string, config *AdapterConfig) (Adapter, error)
    ListAdapters() []string
    ReloadConfig(context.Context, string) error
    ReloadAllConfigs(context.Context) error
    GetConfig(name string) (*model.AdapterConfig, error)
    IsInitialized() bool
}
```

#### Selector模块
```go
// selector.go
type DefaultChannelSelector struct {
    Select(context.Context, *SelectRequest) (*SelectResult, error)
    SelectWithRetry(context.Context, *SelectRequest, int) (*SelectResult, error)
    UpdateStats(context.Context, int, bool, time.Duration) error
    GetStats(context.Context, int) (*ChannelStats, error)
    MarkChannelFailed(context.Context, int, string) error
    RefreshCache(context.Context) error
    registerStrategies()
    filterExcludedChannels([]*model.Channel, []int) []*model.Channel
    disableChannel(context.Context, int, string) error
}

// strategies.go
selectByWeight(context.Context, []*model.Channel, *SelectRequest) (*model.Channel, error)
selectByPriority(context.Context, []*model.Channel, *SelectRequest) (*model.Channel, error)
selectByRoundRobin(context.Context, []*model.Channel, *SelectRequest) (*model.Channel, error)
selectByLowestLatency(context.Context, []*model.Channel, *SelectRequest) (*model.Channel, error)
selectByRandom(context.Context, []*model.Channel, *SelectRequest) (*model.Channel, error)
```

#### Tokenizer模块
```go
// factory.go
type TokenizerFactory struct {
    NewTokenizerFactory() (*TokenizerFactory, error)
    GetTokenizer(model string) (Tokenizer, error)
    CreateStreamCounter(model string) (StreamTokenCounter, error)
    CreateBatchStreamCounter() *BatchStreamTokenCounter
    isOpenAIModel(model string) bool
    getGenericTokenizer(model string) *GenericTokenizer
    Close() error
}

GetGlobalFactory() (*TokenizerFactory, error)
CountTokensQuick(model string, messages []Message) (int, error)

// counter.go
type Tokenizer interface {
    CountTokens(text string) int
    CountMessages(messages []Message) int
}

type StreamTokenCounter interface {
    AddChunk(chunk string)
    GetCurrentCount() int
    Finalize() int
}
```

#### Quota模块
```go
// service.go
type DefaultQuotaService struct {
    PreConsumeQuota(*PreConsumeRequest) (*PreConsumeResponse, error)
    ReturnPreConsumedQuota(requestID string, userID int) error
    PostConsumeQuota(*PostConsumeRequest) error
    RefundQuota(*RefundRequest) error
    GetUserBalance(userID int) (float64, error)
    GetPreConsumedRecord(requestID string) (*PreConsumedRecord, error)
    deductQuota(userID int, quota float64) error
    refundQuota(userID int, quota float64) error
}

// cache.go
type QuotaCache interface {
    GetUserBalance(userID int) (float64, bool, error)
    SetUserBalance(userID int, balance float64) error
    InvalidateUserBalance(userID int) error
    SetPreConsumed(*PreConsumedRecord) error
    GetPreConsumed(requestID string) (*PreConsumedRecord, error)
    DeletePreConsumed(requestID string) error
}
```

#### Relay模块
```go
// stream_handler.go
type StreamHandler struct {
    NewStreamHandler(QuotaService, *TokenizerFactory) *StreamHandler
    HandleStreamResponse(context.Context, *StreamSender, <-chan *StreamChunk, <-chan error, *StreamOptions) (*StreamResult, error)
    HandleStreamWithRetry(context.Context, *StreamSender, func(), *StreamOptions, int) (*StreamResult, error)
}

// stream_sender.go
type StreamSender struct {
    NewStreamSender(http.ResponseWriter) *StreamSender
    Send(*StreamChunk) error
    SendError(string) error
    SendDone() error
    SetHeaders()
}
```

#### Service模块
```go
// pricing_service.go
type PricingService interface {
    GetPricing(context.Context, string, string) (*model.ModelPricing, error)
    ListPricing(context.Context, *bool) ([]*model.ModelPricing, error)
    CreatePricing(context.Context, *model.ModelPricing) error
    UpdatePricing(context.Context, int, *model.ModelPricing) error
    DeletePricing(context.Context, int) error
    CalculateQuota(context.Context, string, string, int, int) (int, error)
    RefreshCache(context.Context) error
}

type DefaultPricingService struct {
    GetGroupRatio(group string) float64
    SetGroupRatio(group string, ratio float64)
}

// channel_ability_service.go
type ChannelAbilityService interface {
    SyncFromChannel(context.Context, *model.Channel) error
    FindByModelAndGroup(context.Context, string, string) ([]*model.ChannelAbility, error)
    GetAvailableChannelsForModel(context.Context, string) ([]*model.ChannelAbility, error)
    DeleteByChannel(context.Context, int) error
}

// health_check_service.go
type HealthCheckService interface {
    StartPeriodicCheck(context.Context)
    CheckChannel(context.Context, int) (*HealthCheckResult, error)
    CalculateHealthScore(context.Context, int) (*HealthScore, error)
    GetHealthStatus(context.Context, int) (*HealthStatus, error)
}

type DefaultHealthCheckService struct {
    checkAllChannels(context.Context)
    handleCheckResult(context.Context, *model.Channel, *HealthCheckResult)
    incrementFailureCount(int) int
    resetFailureCount(int)
    getFailureCount(int) int
    saveCheckResult(*HealthCheckResult)
}
```

#### Handler模块
```go
// channel_handler.go
type ChannelHandler struct {
    ListChannels(*gin.Context)        // GET /api/admin/channels
    CreateChannel(*gin.Context)       // POST /api/admin/channels
    UpdateChannel(*gin.Context)       // PUT /api/admin/channels/:id
    DeleteChannel(*gin.Context)       // DELETE /api/admin/channels/:id
    TestChannel(*gin.Context)         // POST /api/admin/channels/:id/test
    BatchOperation(*gin.Context)      // POST /api/admin/channels/batch
    RegisterRoutes(*gin.RouterGroup)
}

// pricing_handler.go
type PricingHandler struct {
    ListPricing(*gin.Context)         // GET /api/v1/pricing
    GetPricing(*gin.Context)          // GET /api/v1/pricing/:model
    CreatePricing(*gin.Context)       // POST /api/v1/pricing
    UpdatePricing(*gin.Context)       // PUT /api/v1/pricing/:id
    DeletePricing(*gin.Context)       // DELETE /api/v1/pricing/:id
    CalculateQuota(*gin.Context)      // POST /api/v1/pricing/calculate
    RefreshCache(*gin.Context)        // POST /api/v1/pricing/refresh
    RegisterRoutes(*gin.RouterGroup)
}
```

### 前端核心函数速查

#### API服务
```typescript
// api.ts
apiClient.get<T>(url: string, config?: AxiosRequestConfig): Promise<T>
apiClient.post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
apiClient.put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
apiClient.delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>

// sse.ts
class SSEClient {
    connect(url: string, options?: SSEOptions): void
    onMessage(callback: (data: any) => void): void
    onError(callback: (error: Error) => void): void
    onComplete(callback: () => void): void
    close(): void
}

// upload.ts
uploadFile(file: File, onProgress?: (progress: number) => void): Promise<string>
uploadMultiple(files: File[]): Promise<string[]>
uploadChunk(file: File, chunkIndex: number): Promise<void>
cancelUpload(uploadId: string): void
```

#### 状态管理
```typescript
// authStore.ts
interface AuthStore {
    user: User | null
    token: string | null
    isAuthenticated: boolean
    
    login(credentials: LoginCredentials): Promise<void>
    logout(): void
    register(userData: RegisterData): Promise<void>
    refreshToken(): Promise<void>
    updateProfile(data: ProfileData): Promise<void>
    checkAuth(): Promise<boolean>
}

// chatStore.ts
interface ChatStore {
    sessions: Session[]
    currentSession: Session | null
    messages: Message[]
    isStreaming: boolean
    streamingMessage: string
    
    fetchSessions(): Promise<void>
    createSession(title?: string): Promise<Session>
    selectSession(sessionId: string): void
    updateSession(sessionId: string, data: Partial<Session>): Promise<void>
    deleteSession(sessionId: string): Promise<void>
    
    sendMessage(content: string): Promise<void>
    sendStreamingMessage(content: string): Promise<void>
    regenerateMessage(messageId: string): Promise<void>
    editMessage(messageId: string, content: string): Promise<void>
    deleteMessage(messageId: string): Promise<void>
    
    startStreaming(): void
    appendStreamChunk(chunk: string): void
    completeStreaming(): void
    cancelStreaming(): void
}

// settingsStore.ts
interface SettingsStore {
    theme: 'light' | 'dark' | 'system'
    language: string
    model: string
    temperature: number
    maxTokens: number
    systemPrompt: string
    contextLength: number
    
    updateTheme(theme: string): void
    updateLanguage(language: string): void
    updateModelSettings(settings: ModelSettings): void
    resetSettings(): void
    loadSettings(): void
    saveSettings(): void
}
```

---

## 结语

这份 Codemap 为 Oblivious AI 平台提供了全面的代码导航。通过这份文档,您可以:

✅ **快速定位**: 根据功能找到对应的文件和模块  
✅ **理解架构**: 掌握整体设计和数据流转  
✅ **参与开发**: 按照开发指南添加新功能  
✅ **问题排查**: 通过关键文件索引定位问题  

**保持更新**: 本文档会随着项目发展持续更新。

**贡献**: 欢迎提交 PR 完善文档内容。

---

**📅 最后更新**: 2025-11-22  
**📝 文档版本**: v1.0.0  
**👥 维护者**: Oblivious Team

