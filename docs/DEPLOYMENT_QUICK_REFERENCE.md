# Oblivious 部署快速参考

## 🐳 Docker 快速部署

### 方式1：使用 Docker Compose（推荐用于本地开发）

```bash
cd deploy

# 构建镜像
docker build -f docker/Dockerfile.backend -t oblivious-backend:latest ..
docker build -f docker/Dockerfile.frontend -t oblivious-frontend:latest ..

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

**访问地址**:
- 前端: http://localhost:3000
- API 网关: http://localhost:8080
- 用户服务: http://localhost:8081
- 对话服务: http://localhost:8082
- 中转服务: http://localhost:8083

### 方式2：使用自动化脚本

```bash
cd deploy
./deploy.sh docker
```

---

## ☸️ Kubernetes 快速部署

### 方式1：使用自动化脚本（推荐）

```bash
cd deploy
./deploy.sh kubernetes
```

### 方式2：手动部署

```bash
cd deploy

# 1. 创建命名空间
kubectl apply -f kubernetes/namespace.yaml

# 2. 部署基础设施
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/redis.yaml

# 等待启动
kubectl wait --for=condition=ready pod -l app=postgres -n oblivious --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n oblivious --timeout=300s

# 3. 部署后端服务
kubectl apply -f kubernetes/backend-services.yaml

# 4. 部署前端
kubectl apply -f kubernetes/frontend.yaml

# 5. 部署 Ingress（可选）
kubectl apply -f kubernetes/ingress.yaml

# 6. 部署监控（可选）
kubectl apply -f kubernetes/monitoring.yaml
```

### 查看部署状态

```bash
# 查看所有资源
kubectl get all -n oblivious

# 查看 Pod 状态
kubectl get pods -n oblivious

# 查看服务
kubectl get svc -n oblivious

# 查看日志
kubectl logs -f deployment/gateway -n oblivious
```

### 端口转发

```bash
# 前端
kubectl port-forward svc/frontend 3000:3000 -n oblivious

# API 网关
kubectl port-forward svc/gateway 8080:8080 -n oblivious

# Prometheus
kubectl port-forward svc/prometheus 9090:9090 -n oblivious

# Grafana
kubectl port-forward svc/grafana 3001:3000 -n oblivious
```

---

## 📊 常见操作

### 查看日志

**Docker:**
```bash
docker-compose logs -f <service-name>
```

**Kubernetes:**
```bash
kubectl logs -f deployment/<deployment-name> -n oblivious
```

### 扩缩容

```bash
# 扩容
kubectl scale deployment gateway -n oblivious --replicas=5

# 缩容
kubectl scale deployment gateway -n oblivious --replicas=2

# 查看 HPA 自动扩缩状态
kubectl get hpa -n oblivious
```

### 重启服务

**Docker:**
```bash
docker-compose restart <service-name>
```

**Kubernetes:**
```bash
kubectl rollout restart deployment/gateway -n oblivious
```

### 查看资源使用

```bash
kubectl top nodes
kubectl top pods -n oblivious
```

### 更新应用

**Docker:**
```bash
# 重新构建镜像
docker build -f docker/Dockerfile.backend -t oblivious-backend:v2 ..

# 更新 docker-compose.yml 中的镜像版本
# 然后重启
docker-compose down
docker-compose up -d
```

**Kubernetes:**
```bash
kubectl set image deployment/gateway gateway=your-registry/oblivious-backend:v2 -n oblivious
kubectl rollout status deployment/gateway -n oblivious
```

---

## 🔧 故障排除

### Pod 无法启动

```bash
# 查看详细信息
kubectl describe pod <pod-name> -n oblivious

# 查看日志
kubectl logs <pod-name> -n oblivious

# 查看事件
kubectl get events -n oblivious --sort-by='.lastTimestamp'
```

### 数据库连接问题

```bash
# 测试 PostgreSQL 连接
kubectl exec -it deployment/postgres -n oblivious -- \
  psql -U postgres -d oblivious -c "SELECT 1"

# 查看 PostgreSQL 日志
kubectl logs deployment/postgres -n oblivious
```

### 服务间通信问题

```bash
# 在容器中测试 DNS
kubectl run -it --rm debug --image=nicolaka/netshoot -n oblivious -- bash

# 测试连接
nslookup gateway
curl http://gateway:8080/health
exit
```

---

## 📈 监控和告警

### 访问 Prometheus

```bash
kubectl port-forward svc/prometheus 9090:9090 -n oblivious
# 访问 http://localhost:9090
```

### 访问 Grafana

```bash
kubectl port-forward svc/grafana 3001:3000 -n oblivious
# 访问 http://localhost:3001
# 用户名: admin
# 密码: admin
```

---

## 🔐 生产环境配置

### 1. 修改密钥

编辑 `kubernetes/backend-services.yaml` 中的 Secret：

```yaml
stringData:
  DATABASE_USER: "postgres"
  DATABASE_PASSWORD: "your-secure-password"  # 修改这里
  JWT_SECRET: "your-secure-jwt-secret"      # 修改这里
```

### 2. 配置资源限制

根据实际需求调整资源：

```yaml
resources:
  requests:
    cpu: 100m        # 修改这些值
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### 3. 配置备份

```bash
# 备份 PostgreSQL
kubectl exec deployment/postgres -n oblivious -- \
  pg_dump -U postgres oblivious > backup-$(date +%Y%m%d).sql

# 恢复数据库
cat backup.sql | kubectl exec -i deployment/postgres -n oblivious -- \
  psql -U postgres oblivious
```

---

## 📚 详细文档

- 📖 [完整部署指南](deploy/DEPLOYMENT_GUIDE.md)
- 📖 [部署总结](docs/DEPLOYMENT_COMPLETE.md)
- 📖 [快速开始](QUICK_START.md)
- 📖 [开发计划](docs/DEVELOPMENT_PLAN.md)

---

## 🆘 获取帮助

**查看常见问题**:
```bash
# Docker Compose 常见问题
less deploy/DEPLOYMENT_GUIDE.md  # 搜索 "故障排除"

# 查看完整服务状态
docker-compose ps          # Docker
kubectl get all -n oblivious  # Kubernetes
```

---

**快速参考完成！祝您部署愉快！** 🚀

