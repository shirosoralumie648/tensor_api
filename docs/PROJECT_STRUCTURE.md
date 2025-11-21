# 项目结构文档

## 概述

Oblivious 采用 Monorepo 结构，包含前端、后端和部署配置。后端采用微服务架构，每个服务独立部署和扩展。

## 顶层目录结构

```
oblivious/
├── backend/              # 后端 Go 微服务
├── frontend/             # 前端 Next.js 应用
├── deploy/               # 部署配置和脚本
├── docs/                 # 项目文档
├── .github/              # GitHub Actions CI/CD
├── README.md             # 项目说明
└── LICENSE               # 开源协议
```

## 后端结构 (backend/)

### 目录布局

```
backend/
├── cmd/                  # 服务入口点（每个微服务一个目录）
│   ├── gateway/          # API 网关服务
│   │   └── main.go       # 网关启动入口
│   ├── user/             # 用户服务
│   │   └── main.go
│   ├── chat/             # 对话服务
│   │   └── main.go
│   ├── relay/            # AI 中转服务
│   │   └── main.go
│   ├── billing/          # 计费服务
│   │   └── main.go
│   ├── knowledge/        # 知识库服务
│   │   └── main.go
│   ├── file/             # 文件服务
│   │   └── main.go
│   ├── agent/            # 助手服务
│   │   └── main.go
│   └── plugin/           # 插件服务
│       └── main.go
│
├── internal/             # 内部包（不对外暴露）
│   ├── adapter/          # AI 提供商适配器
│   │   ├── openai/       # OpenAI 适配器
│   │   │   ├── client.go
│   │   │   └── types.go
│   │   ├── claude/       # Claude 适配器
│   │   └── gemini/       # Gemini 适配器
│   │
│   ├── gateway/          # 网关业务逻辑
│   │   ├── handler/      # HTTP 处理器
│   │   │   ├── proxy.go  # 代理处理
│   │   │   └── health.go # 健康检查
│   │   └── router/       # 路由配置
│   │       └── router.go
│   │
│   ├── user/             # 用户服务业务逻辑
│   │   ├── handler/      # HTTP 处理器
│   │   │   ├── auth.go   # 认证处理
│   │   │   └── profile.go
│   │   ├── service/      # 业务逻辑
│   │   │   └── user_service.go
│   │   └── repository/   # 数据访问层
│   │       └── user_repo.go
│   │
│   ├── chat/             # 对话服务业务逻辑
│   │   ├── handler/
│   │   │   ├── completion.go  # 对话接口
│   │   │   ├── session.go     # 会话管理
│   │   │   └── message.go     # 消息管理
│   │   ├── service/
│   │   │   ├── chat_service.go
│   │   │   └── stream.go      # 流式响应处理
│   │   └── repository/
│   │       ├── session_repo.go
│   │       └── message_repo.go
│   │
│   ├── relay/            # AI 中转服务
│   │   ├── service/
│   │   │   ├── dispatcher.go  # 渠道调度
│   │   │   └── balancer.go    # 负载均衡
│   │   └── handler/
│   │       └── relay.go
│   │
│   ├── billing/          # 计费服务
│   │   ├── handler/
│   │   │   └── billing.go
│   │   ├── service/
│   │   │   ├── deduct.go      # 扣费逻辑
│   │   │   └── record.go      # 账单记录
│   │   ├── consumer/          # 消息队列消费者
│   │   │   └── recorder.go
│   │   └── repository/
│   │       └── billing_repo.go
│   │
│   ├── knowledge/        # 知识库服务
│   │   ├── handler/
│   │   │   ├── kb.go          # 知识库管理
│   │   │   └── document.go    # 文档管理
│   │   ├── service/
│   │   │   ├── processor.go   # 文档处理
│   │   │   └── search.go      # 向量检索
│   │   └── repository/
│   │       └── vector.go      # 向量存储
│   │
│   ├── rag/              # RAG 相关工具
│   │   ├── embedding.go  # 向量化
│   │   ├── splitter.go   # 文本分割
│   │   └── retriever.go  # 检索器
│   │
│   ├── model/            # 数据模型（GORM）
│   │   ├── user.go
│   │   ├── session.go
│   │   ├── message.go
│   │   ├── billing.go
│   │   ├── channel.go
│   │   ├── agent.go
│   │   ├── knowledge.go
│   │   └── token.go
│   │
│   ├── middleware/       # 中间件
│   │   ├── auth.go       # JWT 认证
│   │   ├── cors.go       # 跨域处理
│   │   ├── logger.go     # 日志记录
│   │   ├── ratelimit.go  # 限流
│   │   └── recovery.go   # 错误恢复
│   │
│   ├── database/         # 数据库相关
│   │   ├── connection.go # 数据库连接
│   │   ├── migrations.go # 迁移管理
│   │   └── redis.go      # Redis 连接
│   │
│   ├── config/           # 配置管理
│   │   ├── config.go     # 配置结构
│   │   └── loader.go     # 配置加载
│   │
│   ├── utils/            # 工具函数
│   │   ├── jwt.go        # JWT 工具
│   │   ├── hash.go       # 哈希工具
│   │   ├── validator.go  # 验证工具
│   │   └── logger.go     # 日志工具
│   │
│   ├── ratelimit/        # 限流组件
│   │   └── quota.go
│   │
│   └── analytics/        # 数据分析
│       └── tracker.go
│
├── migrations/           # 数据库迁移文件
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_user_settings_table.up.sql
│   ├── 000002_create_user_settings_table.down.sql
│   └── ...
│
├── config/               # 配置文件
│   ├── config.example.yaml
│   └── config.yaml
│
├── go.mod                # Go 依赖管理
├── go.sum
├── Makefile              # 构建脚本
└── README.md
```

### 服务架构模式

每个微服务遵循三层架构：

```
Handler（HTTP 层）
    ↓
Service（业务逻辑层）
    ↓
Repository（数据访问层）
```

**示例：Chat 服务**

```go
// handler/completion.go - HTTP 处理器
package handler

func (h *ChatHandler) CreateCompletion(c *gin.Context) {
    // 1. 解析请求
    var req CreateCompletionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // 2. 调用 Service 层
    resp, err := h.chatService.CreateCompletion(c, &req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // 3. 返回响应
    c.JSON(200, resp)
}

// service/chat_service.go - 业务逻辑
package service

func (s *ChatService) CreateCompletion(ctx context.Context, req *CreateCompletionRequest) (*CompletionResponse, error) {
    // 业务逻辑处理
    // - 创建会话
    // - 调用 Relay 服务
    // - 保存消息
    // - 扣费
    return resp, nil
}

// repository/session_repo.go - 数据访问
package repository

func (r *SessionRepository) Create(session *model.Session) error {
    return r.db.Create(session).Error
}
```

## 前端结构 (frontend/)

### 目录布局

```
frontend/
├── src/
│   ├── app/              # Next.js 13+ App Router
│   │   ├── (auth)/       # 认证页面组
│   │   │   ├── login/
│   │   │   └── register/
│   │   ├── chat/         # 对话页面
│   │   │   └── page.tsx
│   │   ├── agents/       # 助手市场
│   │   ├── knowledge/    # 知识库管理
│   │   ├── settings/     # 设置页面
│   │   ├── layout.tsx    # 根布局
│   │   └── page.tsx      # 首页
│   │
│   ├── components/       # React 组件
│   │   ├── chat/
│   │   │   ├── ChatBox.tsx       # 聊天框
│   │   │   ├── MessageList.tsx   # 消息列表
│   │   │   ├── InputBox.tsx      # 输入框
│   │   │   └── SessionList.tsx   # 会话列表
│   │   ├── agent/
│   │   │   ├── AgentCard.tsx     # 助手卡片
│   │   │   └── AgentDetail.tsx
│   │   ├── knowledge/
│   │   │   ├── KBList.tsx        # 知识库列表
│   │   │   └── DocumentUpload.tsx
│   │   ├── common/
│   │   │   ├── Header.tsx        # 页头
│   │   │   ├── Sidebar.tsx       # 侧边栏
│   │   │   ├── Loading.tsx       # 加载组件
│   │   │   └── ErrorBoundary.tsx
│   │   └── ui/           # UI 基础组件（shadcn/ui）
│   │       ├── button.tsx
│   │       ├── input.tsx
│   │       ├── dialog.tsx
│   │       └── ...
│   │
│   ├── hooks/            # React Hooks
│   │   ├── useChat.ts    # 聊天功能
│   │   ├── useAuth.ts    # 认证
│   │   ├── useSession.ts # 会话管理
│   │   └── useKnowledge.ts
│   │
│   ├── services/         # API 服务层
│   │   ├── api.ts        # API 客户端配置
│   │   ├── chat.ts       # 对话 API
│   │   ├── auth.ts       # 认证 API
│   │   ├── agent.ts      # 助手 API
│   │   └── knowledge.ts  # 知识库 API
│   │
│   ├── store/            # 状态管理（Zustand）
│   │   ├── authStore.ts  # 认证状态
│   │   ├── chatStore.ts  # 聊天状态
│   │   └── uiStore.ts    # UI 状态
│   │
│   ├── types/            # TypeScript 类型定义
│   │   ├── chat.ts
│   │   ├── user.ts
│   │   ├── agent.ts
│   │   └── api.ts
│   │
│   ├── utils/            # 工具函数
│   │   ├── format.ts     # 格式化工具
│   │   ├── storage.ts    # 本地存储
│   │   └── constants.ts  # 常量定义
│   │
│   └── styles/           # 样式文件
│       └── globals.css   # 全局样式
│
├── public/               # 静态资源
│   ├── icons/
│   ├── images/
│   └── favicon.ico
│
├── .env.example          # 环境变量示例
├── .env.local            # 本地环境变量
├── next.config.js        # Next.js 配置
├── tailwind.config.js    # TailwindCSS 配置
├── tsconfig.json         # TypeScript 配置
├── package.json
└── README.md
```

### 组件组织原则

**原子设计模式**：

```
atoms（原子）
  └─ Button, Input, Label

molecules（分子）
  └─ FormField, SearchBar

organisms（有机体）
  └─ ChatBox, SessionList

templates（模板）
  └─ ChatLayout

pages（页面）
  └─ ChatPage
```

## 部署结构 (deploy/)

```
deploy/
├── docker/               # Docker 配置
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   └── docker-compose.yml
│
├── k8s/                  # Kubernetes 配置
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   ├── deployments/      # 部署配置
│   │   ├── gateway.yaml
│   │   ├── user.yaml
│   │   ├── chat.yaml
│   │   └── ...
│   ├── services/         # 服务配置
│   │   ├── gateway.yaml
│   │   └── ...
│   ├── ingress.yaml      # 入口配置
│   └── hpa/              # 自动扩缩容
│       ├── gateway-hpa.yaml
│       └── chat-hpa.yaml
│
├── helm/                 # Helm Charts
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│
├── monitoring/           # 监控配置
│   ├── prometheus.yml
│   └── grafana-dashboard.json
│
└── scripts/              # 部署脚本
    ├── deploy.sh
    ├── rollback.sh
    └── health-check.sh
```

## 配置文件说明

### 后端配置 (config/config.yaml)

```yaml
app:
  name: oblivious
  env: production
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: oblivious

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key
  expire: 7200  # 2小时

ai:
  default_model: gpt-3.5-turbo
  max_tokens: 4096
```

### 前端配置 (.env.local)

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 代码规范

### Go 代码规范

**包命名**：
- 小写字母
- 简短有意义
- 避免下划线

**文件命名**：
- 小写字母 + 下划线
- 例如：`user_service.go`

**函数命名**：
- 驼峰命名
- 导出函数首字母大写
- 私有函数首字母小写

**注释**：
```go
// CreateUser 创建新用户
// 参数：
//   - username: 用户名
//   - email: 邮箱
// 返回：
//   - *User: 创建的用户对象
//   - error: 错误信息
func CreateUser(username, email string) (*User, error) {
    // 实现...
}
```

### TypeScript 代码规范

**命名规范**：
- 组件：PascalCase（如 `ChatBox.tsx`）
- 函数：camelCase（如 `sendMessage`）
- 常量：UPPER_SNAKE_CASE（如 `API_BASE_URL`）
- 类型：PascalCase（如 `UserProfile`）

**导入顺序**：
```typescript
// 1. React 相关
import React, { useState } from 'react'

// 2. 第三方库
import axios from 'axios'

// 3. 本地组件
import { ChatBox } from '@/components/chat/ChatBox'

// 4. 工具函数
import { formatDate } from '@/utils/format'

// 5. 类型
import type { Message } from '@/types/chat'

// 6. 样式
import styles from './page.module.css'
```

## 版本控制

### Git 分支策略

```
main              # 主分支（生产环境）
  ├── develop     # 开发分支
  │   ├── feature/user-auth      # 功能分支
  │   ├── feature/chat-service
  │   └── fix/message-bug        # 修复分支
  └── release/v1.0.0  # 发布分支
```

### Commit 规范

```bash
feat: 添加用户认证功能
fix: 修复消息发送失败的问题
docs: 更新 API 文档
style: 格式化代码
refactor: 重构聊天服务
test: 添加单元测试
chore: 更新依赖
```

## 相关文档

- [架构设计](ARCHITECTURE.md)
- [API 参考](API_REFERENCE.md)
- [开发指南](CONTRIBUTING.md)
