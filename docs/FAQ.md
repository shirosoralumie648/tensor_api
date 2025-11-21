# 常见问题解答 (FAQ)

本文档列出使用 Oblivious 平台时的常见问题和解决方案。

**最后更新**: 2024 年 11 月 21 日

---

## 目录

- [安装和环境配置](#安装和环境配置)
- [开发相关](#开发相关)
- [部署相关](#部署相关)
- [服务使用](#服务使用)
- [错误排查](#错误排查)
- [性能优化](#性能优化)

---

## 安装和环境配置

### Q: 需要什么版本的 Go？

**A**: 项目要求 Go 1.24.10 或更高版本。运行以下命令检查版本：

```bash
go version
```

如果版本过低，请访问 [Go 官网](https://go.dev/dl/) 下载最新版本。

---

### Q: PostgreSQL 端口被占用怎么办？

**A**: 检查端口占用情况：

```bash
lsof -i :5432  # macOS/Linux
netstat -ano | findstr :5432  # Windows
```

解决方案：
1. 停止占用端口的进程
2. 或修改 `docker-compose.dev.yml` 中的端口映射：

```yaml
ports:
  - "5433:5432"  # 将主机端口改为 5433
```

然后更新环境变量中的 `DATABASE_PORT`。

---

### Q: Redis 连接失败怎么办？

**A**: 确保 Redis 容器正在运行：

```bash
docker ps | grep redis
```

如果未运行，启动 Redis：

```bash
docker compose -f deploy/docker-compose.dev.yml up -d redis
```

测试连接：

```bash
redis-cli -h localhost -p 6379 ping
# 应返回 PONG
```

---

### Q: 如何重置数据库？

**A**: 

```bash
cd backend

# 回滚所有迁移
make migrate-down

# 重新运行迁移
make migrate-up

# 或者完全重建数据库
docker compose -f ../deploy/docker-compose.dev.yml down -v
docker compose -f ../deploy/docker-compose.dev.yml up -d postgres
make migrate-up
```

---

## 开发相关

### Q: 如何添加新的 API 接口？

**A**: 

1. 在对应服务的 `main.go` 中添加路由
2. 在 `internal/service/` 中实现业务逻辑
3. 在 `internal/repository/` 中实现数据访问（如需要）
4. 更新 API 文档 (`docs/API_REFERENCE.md`)
5. 添加单元测试

示例：

```go
// backend/cmd/chat/main.go
api.POST("/chat/sessions/:id/archive", func(c *gin.Context) {
    sessionID := c.Param("id")
    userID := c.GetInt("user_id")
    
    err := chatService.ArchiveSession(c.Request.Context(), userID, sessionID)
    if err != nil {
        utils.InternalError(c, err.Error())
        return
    }
    
    utils.Success(c, nil, "会话已归档")
})
```

---

### Q: 如何运行单个服务进行调试？

**A**:

```bash
cd backend/cmd/gateway  # 或其他服务目录
go run main.go
```

或使用 VSCode 调试配置（`.vscode/launch.json`）:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Gateway",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/gateway",
      "env": {
        "APP_ENV": "development"
      }
    }
  ]
}
```

---

### Q: 如何添加新的数据库表？

**A**:

1. 创建迁移文件：

```bash
cd backend
migrate create -ext sql -dir migrations -seq add_your_table_name
```

2. 编辑生成的 `.up.sql` 和 `.down.sql` 文件

3. 运行迁移：

```bash
make migrate-up
```

4. 在 `internal/model/` 中创建对应的 Go 结构体

5. 更新数据库设计文档 (`docs/DATABASE_DESIGN.md`)

---

### Q: 如何运行测试？

**A**:

```bash
# 运行所有测试
cd backend
go test ./... -v

# 运行特定包的测试
go test ./internal/service -v

# 查看测试覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

### Q: 代码格式化工具怎么用？

**A**:

```bash
# Go 代码格式化
cd backend
gofmt -w .
go vet ./...

# 或使用 golangci-lint（推荐）
golangci-lint run

# 前端代码格式化
cd frontend
npm run lint
npm run lint:fix
npm run format
```

---

## 部署相关

### Q: Docker 镜像构建失败怎么办？

**A**: 

常见原因：
1. 网络问题：检查 Docker 镜像源配置
2. 权限问题：确保有足够的磁盘空间和权限
3. 缓存问题：清理 Docker 缓存

```bash
# 清理 Docker 缓存
docker system prune -a

# 重新构建（不使用缓存）
docker build --no-cache -f deploy/docker/Dockerfile.backend -t oblivious-backend:latest .
```

---

### Q: Kubernetes Pod 启动失败？

**A**:

1. 查看 Pod 状态：

```bash
kubectl get pods -n oblivious
kubectl describe pod <pod-name> -n oblivious
```

2. 查看日志：

```bash
kubectl logs <pod-name> -n oblivious
kubectl logs <pod-name> -n oblivious --previous  # 查看上一次运行的日志
```

3. 常见问题：
   - ImagePullBackOff: 镜像拉取失败，检查镜像名称和 Registry 配置
   - CrashLoopBackOff: 容器启动后崩溃，查看日志排查错误
   - Pending: 资源不足或 PVC 未就绪

---

### Q: 如何查看服务日志？

**A**:

**Docker Compose**:

```bash
docker compose -f deploy/docker-compose.dev.yml logs -f gateway
```

**Kubernetes**:

```bash
kubectl logs -f deployment/gateway -n oblivious
```

**本地运行**:

```bash
tail -f backend/logs/gateway.log
```

---

### Q: 如何更新已部署的服务？

**A**:

**Docker Compose**:

```bash
docker compose -f deploy/docker-compose.dev.yml down
docker compose -f deploy/docker-compose.dev.yml up -d --build
```

**Kubernetes**:

```bash
# 更新镜像
kubectl set image deployment/gateway gateway=oblivious-gateway:v1.1.0 -n oblivious

# 或重新应用配置
kubectl apply -f deploy/k8s/deployments/gateway.yaml

# 查看滚动更新状态
kubectl rollout status deployment/gateway -n oblivious
```

---

## 服务使用

### Q: 如何获取 API Token？

**A**:

1. 注册账号：

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

2. 从响应中获取 `access_token`

3. 在后续请求中使用：

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8080/api/v1/user/profile
```

---

### Q: Token 过期了怎么办？

**A**:

使用 Refresh Token 获取新的 Access Token：

```bash
curl -X POST http://localhost:8080/api/v1/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

---

### Q: 如何测试流式对话？

**A**:

使用 `curl` 测试 SSE 流式响应：

```bash
curl -X POST http://localhost:8080/api/v1/chat/messages/stream \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "your-session-id",
    "content": "你好"
  }' \
  --no-buffer
```

或使用 EventSource API（JavaScript）：

```javascript
const eventSource = new EventSource(
  'http://localhost:8080/api/v1/chat/messages/stream?token=YOUR_TOKEN'
);

eventSource.onmessage = (event) => {
  console.log('Received:', event.data);
};
```

---

### Q: 如何查看我的配额使用情况？

**A**:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/user/profile
```

响应中的 `quota` 字段显示剩余配额（单位：分，1元=100分）。

---

## 错误排查

### Q: 出现 "database connection failed" 错误？

**A**:

1. 检查数据库是否运行：

```bash
docker ps | grep postgres
```

2. 检查环境变量配置：

```bash
echo $DATABASE_HOST
echo $DATABASE_PORT
```

3. 测试数据库连接：

```bash
PGPASSWORD=password psql -h localhost -p 5433 -U postgres -d oblivious -c "SELECT 1"
```

4. 检查防火墙规则

---

### Q: "JWT token invalid" 错误？

**A**:

可能原因：
1. Token 已过期 → 使用 Refresh Token 刷新
2. Token 格式错误 → 检查 Authorization header 格式：`Bearer <token>`
3. JWT Secret 不匹配 → 检查环境变量 `JWT_SECRET`

---

### Q: "rate limit exceeded" 错误？

**A**:

您的请求频率超过了限制。默认限流规则：
- 公开接口：10 请求/分钟
- 受保护接口：100 请求/分钟

解决方案：
1. 降低请求频率
2. 升级账户获得更高限额
3. 开发环境可调整 `middleware/rate_limit.go` 中的限流参数

---

### Q: OpenAI API 调用失败？

**A**:

1. 检查 API Key 是否正确：

```bash
echo $OPENAI_API_KEY
```

2. 检查渠道配置：

```sql
SELECT * FROM channels WHERE type = 1 AND status = 1;
```

3. 测试 OpenAI API：

```bash
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer YOUR_OPENAI_KEY"
```

4. 查看 Relay Service 日志：

```bash
tail -f backend/logs/relay.log
```

---

## 性能优化

### Q: 如何提升 API 响应速度？

**A**:

1. **启用 Redis 缓存**：缓存热点数据
2. **数据库查询优化**：
   - 添加索引
   - 优化 N+1 查询问题
   - 使用查询分析：`EXPLAIN ANALYZE`
3. **启用 HTTP/2**
4. **使用 CDN** 加速静态资源
5. **增加服务副本数**（Kubernetes HPA）

---

### Q: 数据库查询慢怎么办？

**A**:

1. 分析慢查询：

```sql
-- 查看慢查询
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

2. 添加索引：

```sql
-- 为常用查询字段添加索引
CREATE INDEX idx_messages_session_created 
ON messages(session_id, created_at DESC);
```

3. 使用连接池：检查 `database.go` 中的连接池配置

4. 分库分表（数据量 > 500GB 时考虑）

---

### Q: 内存使用过高怎么办？

**A**:

1. 分析内存使用：

```bash
# Go 程序内存分析
go tool pprof http://localhost:8080/debug/pprof/heap
```

2. 常见原因：
   - 连接泄漏：确保关闭数据库连接和 HTTP 连接
   - Goroutine 泄漏：检查 goroutine 数量
   - 缓存过大：调整 Redis 和内存缓存配置

3. 优化建议：
   - 使用对象池减少 GC 压力
   - 限制并发数
   - 调整 `GOMAXPROCS`

---

## 其他问题

### Q: 找不到某个功能的文档？

**A**: 

- 查看 [服务开发状态](SERVICE_STATUS.md) 确认功能是否已实现
- 搜索现有文档：`docs/` 目录
- 查看源代码注释
- 提交 Issue 请求补充文档

---

### Q: 如何贡献代码？

**A**: 

请阅读 [贡献指南](CONTRIBUTING.md)，步骤包括：
1. Fork 仓库
2. 创建特性分支
3. 提交代码
4. 发起 Pull Request

---

### Q: 如何报告 Bug？

**A**:

访问 [GitHub Issues](https://github.com/your-org/oblivious/issues/new) 提交 Bug 报告，请包含：
- Bug 描述
- 复现步骤
- 期望行为 vs 实际行为
- 环境信息（OS、版本等）
- 相关日志和截图

---

### Q: 如何联系开发团队？

**A**:

- GitHub Issues: 技术问题和 Bug 报告
- Discord: 社区讨论
- Email: support@oblivious.ai

---

## 相关文档

- [快速启动指南](QUICK_START.md)
- [架构设计](ARCHITECTURE.md)
- [API 参考](API_REFERENCE.md)
- [贡献指南](CONTRIBUTING.md)
- [服务开发状态](SERVICE_STATUS.md)

---

**文档版本**: v1.0.0  
**最后更新**: 2024 年 11 月 21 日  
**维护团队**: Oblivious 开发团队

**持续更新中，如有其他问题欢迎提交 Issue！**
