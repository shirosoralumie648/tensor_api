# 快速开始指南

## 前置条件

在开始之前，请确保你的系统已安装以下软件：

- **Docker** 20.10+ 和 **Docker Compose** v2
- **Go** 1.24.10 或更高版本
- **Node.js** 20+ 和 **npm** 或 **pnpm**
- **Make** 工具（可选，用于简化命令）
- **Git** 版本控制

### 检查环境

```bash
# 检查 Docker
docker --version
docker-compose --version

# 检查 Go
go version

# 检查 Node.js
node --version
npm --version

# 检查 Make
make --version
```

## 快速启动（推荐）

### 1. 克隆仓库

```bash
git clone https://github.com/your-org/oblivious.git
cd oblivious
```

### 2. 启动基础设施

使用 Docker Compose 一键启动所有依赖服务：

```bash
cd deploy
docker-compose up -d
```

这将启动以下服务：
- PostgreSQL（端口 5432）
- Redis（端口 6379）
- MinIO（端口 9000）
- RabbitMQ（端口 5672，管理界面 15672）

### 3. 配置环境变量

```bash
# 后端配置
cd ../backend
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml，填入数据库连接信息

# 前端配置
cd ../frontend
cp .env.example .env.local
# 编辑 .env.local，配置 API 地址
```

### 4. 数据库迁移

```bash
cd ../backend
make migrate-up
# 或者直接使用 migrate 命令
# migrate -path ./migrations -database "postgresql://user:pass@localhost:5432/oblivious?sslmode=disable" up
```

### 5. 启动后端服务

**方式一：使用 Make（推荐）**

```bash
cd backend
make run-gateway  # 在新终端运行
make run-user     # 在新终端运行
make run-chat     # 在新终端运行
make run-relay    # 在新终端运行
```

**方式二：直接运行**

```bash
# 终端1：API 网关
cd backend/cmd/gateway
go run main.go

# 终端2：用户服务
cd backend/cmd/user
go run main.go

# 终端3：对话服务
cd backend/cmd/chat
go run main.go

# 终端4：中转服务
cd backend/cmd/relay
go run main.go
```

### 6. 启动前端

```bash
cd frontend
npm install
# 或 pnpm install

npm run dev
# 或 pnpm dev
```

### 7. 访问应用

打开浏览器访问：
- **前端应用**：http://localhost:3000
- **API 网关**：http://localhost:8080
- **RabbitMQ 管理界面**：http://localhost:15672（默认账号：guest/guest）
- **MinIO 控制台**：http://localhost:9001（默认账号：minioadmin/minioadmin）

## 详细步骤说明

### PostgreSQL 初始化

如果需要手动创建数据库：

```sql
-- 连接到 PostgreSQL
psql -U postgres -h localhost

-- 创建数据库
CREATE DATABASE oblivious;

-- 创建用户
CREATE USER oblivious_user WITH PASSWORD 'your_password';

-- 授权
GRANT ALL PRIVILEGES ON DATABASE oblivious TO oblivious_user;

-- 启用扩展
\c oblivious
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;
```

### Redis 配置

默认配置无需修改，如需持久化：

```bash
# 修改 docker-compose.yml
redis:
  command: redis-server --appendonly yes
  volumes:
    - redis-data:/data
```

### MinIO 初始化

创建必要的 Bucket：

```bash
# 使用 mc 客户端
docker exec -it minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec -it minio mc mb local/oblivious
docker exec -it minio mc mb local/oblivious-kb
```

## 开发工具

### Makefile 命令

```bash
# 后端
make build          # 编译所有服务
make test           # 运行测试
make lint           # 代码检查
make migrate-up     # 应用数据库迁移
make migrate-down   # 回滚数据库迁移

# 前端
cd frontend
npm run build       # 构建生产版本
npm run lint        # 代码检查
npm run test        # 运行测试
```

### 调试配置

**VSCode launch.json**：

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Gateway",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/backend/cmd/gateway",
      "env": {
        "CONFIG_PATH": "${workspaceFolder}/backend/config/config.yaml"
      }
    }
  ]
}
```

## 常见问题

### 1. 端口冲突

如果端口被占用，修改 `docker-compose.yml` 或配置文件中的端口号。

```bash
# 查看端口占用
lsof -i :5432
lsof -i :6379
```

### 2. 数据库连接失败

检查配置文件中的数据库连接信息：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: oblivious
```

### 3. Go 模块下载慢

配置 Go 代理：

```bash
export GOPROXY=https://goproxy.cn,direct
# 或
export GOPROXY=https://goproxy.io,direct
```

### 4. npm 安装慢

配置 npm 镜像：

```bash
npm config set registry https://registry.npmmirror.com
# 或使用 pnpm
pnpm config set registry https://registry.npmmirror.com
```

### 5. 前端无法连接后端

检查 `.env.local` 中的 API 地址配置：

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 测试账号

首次启动后可以注册账号，或使用以下测试数据：

```bash
# 使用 curl 创建测试用户
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_user",
    "email": "test@example.com",
    "password": "Test123456"
  }'
```

## 下一步

- 📖 阅读 [架构设计文档](ARCHITECTURE.md)
- 🔌 配置 [AI 适配器](AI_ADAPTER_SETUP.md)
- 🚀 查看 [生产部署指南](PRODUCTION_DEPLOYMENT_GUIDE.md)
- 🤝 参与 [贡献开发](CONTRIBUTING.md)

## 获取帮助

- GitHub Issues：https://github.com/your-org/oblivious/issues
- 开发文档：https://docs.oblivious.ai
- 社区讨论：https://discord.gg/oblivious
