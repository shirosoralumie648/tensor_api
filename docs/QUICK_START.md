# Oblivious 快速启动指南

## 前置要求

- Go 1.24.10 或更高版本
- Docker & Docker Compose (v2)
- PostgreSQL 15 (通过 Docker)
- Redis 7 (通过 Docker)
- Make 工具（可选，用于简化命令）

> **注意**: 确保已安装并配置好 Go 环境，可以运行 `go version` 验证。

## 项目结构

```
oblivious/
├── backend/                    # 后端微服务
│   ├── cmd/                   # 各个微服务的入口
│   │   ├── gateway/           # API 网关 ✅ 已实现
│   │   ├── user/              # 用户服务 ✅ 已实现
│   │   ├── chat/              # 对话服务 ✅ 已实现
│   │   ├── relay/             # 中转服务 ✅ 已实现
│   │   ├── agent/             # 助手服务 🚧 开发中
│   │   ├── billing/           # 计费服务 🚧 开发中
│   │   ├── kb/                # 知识库服务 🚧 开发中
│   │   ├── file/              # 文件服务 📋 规划中
│   │   ├── plugin/            # 插件服务 📋 规划中
│   │   └── worker/            # 异步任务 📋 规划中
│   ├── internal/              # 内部包
│   │   ├── config/            # 配置管理
│   │   ├── database/          # 数据库连接
│   │   ├── middleware/        # HTTP 中间件
│   │   ├── model/             # 数据模型
│   │   ├── repository/        # 数据访问层
│   │   ├── service/           # 业务逻辑层
│   │   └── utils/             # 工具函数
│   ├── pkg/                   # 公共包
│   │   └── logger/            # 日志管理
│   ├── migrations/            # 数据库迁移文件
│   ├── scripts/               # 运维脚本
│   │   ├── start_services.sh  # 启动脚本
│   │   ├── stop_services.sh   # 停止脚本
│   │   └── test_all_services.sh # 测试脚本
│   ├── go.mod                 # Go 依赖管理
│   ├── Makefile              # Make 任务
│   └── env.test              # 测试环境配置
├── deploy/                    # 部署配置
│   └── docker-compose.dev.yml # 开发环境 Docker Compose
├── docs/                      # 文档
│   ├── DEVELOPMENT_PLAN.md    # 开发计划
│   ├── WEEK1_PROGRESS.md      # Week 1 进度
│   ├── WEEK2_PROGRESS.md      # Week 2 进度
│   └── WEEK3_TEST_COMPLETE.md # Week 3 测试报告
└── README.md                  # 项目说明
```

## 启动步骤

### 1. 启动基础设施 (PostgreSQL 和 Redis)

```bash
cd /home/shirosora/windsurf-storage/oblivious
docker compose -f deploy/docker-compose.dev.yml up -d
```

验证容器已启动：
```bash
docker compose -f deploy/docker-compose.dev.yml ps
```

应该看到：
- `oblivious-postgres` - 运行中 (端口 5433)
- `oblivious-redis` - 运行中 (端口 6379)

### 2. 启动所有微服务

```bash
cd /home/shirosora/windsurf-storage/oblivious/backend
./scripts/start_services.sh
```

等待所有服务启动（约 5-10 秒）。

验证服务已启动：
```bash
curl http://localhost:8080/health  # 网关（✅ 应该返回 {"status":"ok"}）
curl http://localhost:8081/health  # 用户服务（✅ 应该返回 {"status":"ok"}）
curl http://localhost:8082/health  # 对话服务（✅ 应该返回 {"status":"ok"}）
curl http://localhost:8083/health  # 中转服务（✅ 应该返回 {"status":"ok"}）
```

所有请求都应该返回：
```json
{"status":"ok"}
```

### 3. 运行测试

```bash
cd /home/shirosora/windsurf-storage/oblivious/backend
./scripts/test_all_services.sh
```

### 4. 查看日志

```bash
# 网关日志
tail -f /home/shirosora/windsurf-storage/oblivious/backend/logs/gateway.log

# 用户服务日志
tail -f /home/shirosora/windsurf-storage/oblivious/backend/logs/user.log

# 对话服务日志
tail -f /home/shirosora/windsurf-storage/oblivious/backend/logs/chat.log
```

### 5. 停止所有服务

```bash
cd /home/shirosora/windsurf-storage/oblivious/backend
./scripts/stop_services.sh
```

## API 使用示例

### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

响应中会包含 `access_token` 和 `refresh_token`。

### 创建对话会话

```bash
curl -X POST http://localhost:8080/api/v1/chat/sessions \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的第一个对话",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个有帮助的助手"
  }'
```

### 发送消息

```bash
curl -X POST http://localhost:8080/api/v1/chat/messages \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "SESSION_UUID",
    "content": "你好，请介绍一下自己"
  }'
```

### 刷新 Token

```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

## 环境变量配置

编辑 `/home/shirosora/windsurf-storage/oblivious/backend/env.test` 文件来修改配置：

```env
# 数据库配置
DATABASE_HOST=localhost
DATABASE_PORT=5433
DATABASE_USER=postgres
DATABASE_PASSWORD=password
DATABASE_NAME=oblivious

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT 配置
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRE_HOURS=1
REFRESH_TOKEN_EXPIRE_DAYS=7

# 应用配置
APP_NAME=Oblivious
APP_ENV=development
APP_PORT=8080

# 服务 URL
USER_SERVICE_URL=http://localhost:8081
CHAT_SERVICE_URL=http://localhost:8082
RELAY_SERVICE_URL=http://localhost:8083
BILLING_SERVICE_URL=http://localhost:8088

# 上游 AI API 密钥（可选，用于测试中转服务）
OPENAI_API_KEY=sk-your-openai-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
```

## 数据库管理

### 运行迁移

```bash
cd /home/shirosora/windsurf-storage/oblivious/backend
export PATH=/usr/local/go/bin:$PATH
make migrate-up
```

### 回滚迁移

```bash
make migrate-down
```

### 检查迁移状态

```bash
make migrate-status
```

## 常见问题

### 问题：端口已被占用

```
Error response from daemon: ports are not available: exposing port TCP 0.0.0.0:5432
```

**解决方案**：修改 `docker-compose.dev.yml` 中的端口映射，或者停止占用该端口的其他容器。

### 问题：数据库连接失败

确保 PostgreSQL 容器正在运行：
```bash
docker compose -f deploy/docker-compose.dev.yml ps
```

### 问题：服务无法启动

检查日志文件：
```bash
tail -100 /home/shirosora/windsurf-storage/oblivious/backend/logs/user.log
```

常见原因：
- 数据库连接失败 → 检查环境变量
- 端口被占用 → 更改配置中的端口
- 数据库迁移失败 → 运行 `make migrate-up`

## 开发流程

### 修改代码后重新编译

```bash
cd /home/shirosora/windsurf-storage/oblivious/backend
export PATH=/usr/local/go/bin:$PATH

# 编译单个服务
go build -o bin/user ./cmd/user
go build -o bin/chat ./cmd/chat
go build -o bin/gateway ./cmd/gateway

# 或者使用 Makefile
make build
```

### 重启服务

```bash
cd /home/shirosora/windsurf-storage/oblivious/backend
./scripts/stop_services.sh
./scripts/start_services.sh
```

## 监控和调试

### 查看请求日志

所有 HTTP 请求都会被记录。检查服务日志：
```bash
tail -f /home/shirosora/windsurf-storage/oblivious/backend/logs/gateway.log
```

### 数据库查询

连接到 PostgreSQL 数据库：
```bash
PGPASSWORD=password psql -h localhost -p 5433 -U postgres -d oblivious
```

常用查询：
```sql
-- 查看所有用户
SELECT * FROM users;

-- 查看所有会话
SELECT * FROM sessions;

-- 查看所有消息
SELECT * FROM messages;

-- 查看用户登录日志
SELECT id, username, last_login_at FROM users;
```

## 下一步

- 👉 查看 [服务开发状态](docs/SERVICE_STATUS.md) 了解各服务实现进度
- 👉 阅读 [API 参考文档](docs/API_REFERENCE.md) 了解接口详情
- 👉 参考 [架构设计](docs/ARCHITECTURE.md) 理解系统架构
- 👉 如需贡献代码，请阅读 [贡献指南](docs/CONTRIBUTING.md)

## 获取帮助

查看相关文档：
- [项目架构文档](docs/ARCHITECTURE.md)
- [API 参考文档](docs/API_REFERENCE.md)
- [快速部署参考](docs/DEPLOYMENT_QUICK_REFERENCE.md)
- [常见问题](docs/FAQ.md)（规划中）

如需帮助：
- 提交 Issue: [GitHub Issues](https://github.com/your-org/oblivious/issues)
- 加入社区: [Discord](https://discord.gg/oblivious)
- 发送邮件: support@oblivious.ai

