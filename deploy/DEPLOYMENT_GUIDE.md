# Oblivious 部署指南

本指南涵盖了使用 Docker 和 Kubernetes 部署 Oblivious 项目的完整步骤。

---

## 📋 目录

1. [Docker 部署（本地开发）](#docker-部署本地开发)
2. [Kubernetes 部署（生产环境）](#kubernetes-部署生产环境)
3. [扩缩容配置](#扩缩容配置)
4. [监控和日志](#监控和日志)
5. [故障排除](#故障排除)

---

## Docker 部署（本地开发）

### 前置条件

- Docker 20.10+
- Docker Compose 2.0+
- 至少 4GB RAM
- 至少 10GB 磁盘空间

### 快速开始

#### 1. 克隆项目

```bash
cd /home/shirosora/windsurf-storage/oblivious
```

#### 2. 创建 .env 文件

```bash
cat > deploy/.env << 'EOF'
# 数据库配置
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=oblivious
DB_PORT=5433

# Redis 配置
REDIS_PORT=6379

# 应用配置
APP_ENV=development
JWT_SECRET=your-super-secret-jwt-key

# 镜像配置
COMPOSE_PROJECT_NAME=oblivious
EOF
```

#### 3. 构建镜像

```bash
# 构建后端镜像
docker build -f deploy/docker/Dockerfile.backend -t oblivious-backend:latest .

# 构建前端镜像
docker build -f deploy/docker/Dockerfile.frontend -t oblivious-frontend:latest .
```

#### 4. 启动所有服务

```bash
cd deploy
docker-compose up -d
```

#### 5. 检查服务状态

```bash
docker-compose ps
```

应该看到所有服务都是 `Up` 状态：

```
NAME                    STATUS
oblivious-postgres      Up (healthy)
oblivious-redis         Up (healthy)
oblivious-gateway       Up
oblivious-user          Up
oblivious-chat          Up
oblivious-relay         Up
oblivious-frontend      Up
```

#### 6. 运行数据库迁移

```bash
# 进入网关容器
docker-compose exec gateway sh

# 运行迁移
migrate -path /app/migrations -database "postgresql://$DATABASE_USER:$DATABASE_PASSWORD@postgres:5432/$DATABASE_NAME?sslmode=disable" up

# 退出容器
exit
```

#### 7. 访问应用

- **前端**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **User Service**: http://localhost:8081
- **Chat Service**: http://localhost:8082
- **Relay Service**: http://localhost:8083

#### 8. 测试功能

```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Password123!"
  }'

# 登录
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Password123!"
  }'
```

#### 9. 查看日志

```bash
# 查看所有日志
docker-compose logs -f

# 查看特定服务的日志
docker-compose logs -f gateway
docker-compose logs -f frontend
```

#### 10. 停止服务

```bash
docker-compose down

# 包括删除卷
docker-compose down -v
```

---

## Kubernetes 部署（生产环境）

### 前置条件

- Kubernetes 1.20+
- kubectl 配置正确
- 可用的 Docker 仓库（Docker Hub、ECR、GCR 等）
- 至少 4 个 CPU 和 8GB RAM 的节点

### 部署步骤

#### 1. 准备镜像

```bash
# 构建后端镜像
docker build -f deploy/docker/Dockerfile.backend -t your-registry/oblivious-backend:1.0.0 .
docker push your-registry/oblivious-backend:1.0.0

# 构建前端镜像
docker build -f deploy/docker/Dockerfile.frontend -t your-registry/oblivious-frontend:1.0.0 .
docker push your-registry/oblivious-frontend:1.0.0
```

#### 2. 创建命名空间

```bash
kubectl apply -f deploy/kubernetes/namespace.yaml
```

#### 3. 创建 Secrets

```bash
# 编辑后端 secret 中的敏感信息
kubectl create secret generic backend-secret \
  -n oblivious \
  --from-literal=DATABASE_USER=postgres \
  --from-literal=DATABASE_PASSWORD=your-secure-password \
  --from-literal=JWT_SECRET=your-secure-jwt-secret \
  --dry-run=client -o yaml | kubectl apply -f -
```

#### 4. 部署数据库和缓存

```bash
# 部署 PostgreSQL
kubectl apply -f deploy/kubernetes/postgres.yaml

# 部署 Redis
kubectl apply -f deploy/kubernetes/redis.yaml

# 等待 Pod 启动
kubectl wait --for=condition=ready pod -l app=postgres -n oblivious --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n oblivious --timeout=300s
```

#### 5. 运行数据库迁移

```bash
# 创建一个 Job 来运行迁移
kubectl run migrate -n oblivious \
  --image=migrate/migrate \
  --rm -it \
  --restart=Never \
  -- -path /app/migrations \
  -database "postgresql://postgres:password@postgres:5432/oblivious?sslmode=disable" \
  up
```

#### 6. 部署后端服务

```bash
# 编辑后端服务配置中的镜像地址
sed -i 's/oblivious-backend:latest/your-registry\/oblivious-backend:1.0.0/g' deploy/kubernetes/backend-services.yaml

# 部署
kubectl apply -f deploy/kubernetes/backend-services.yaml

# 检查部署状态
kubectl get deployments -n oblivious
kubectl get pods -n oblivious
```

#### 7. 部署前端

```bash
# 编辑前端配置中的镜像地址
sed -i 's/oblivious-frontend:latest/your-registry\/oblivious-frontend:1.0.0/g' deploy/kubernetes/frontend.yaml

# 部署
kubectl apply -f deploy/kubernetes/frontend.yaml
```

#### 8. 配置 Ingress（可选）

```bash
# 安装 NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.5.1/deploy/static/provider/cloud/deploy.yaml

# 等待 Ingress Controller 启动
kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout=300s

# 部署 Ingress
kubectl apply -f deploy/kubernetes/ingress.yaml

# 获取 Ingress IP
kubectl get ingress -n oblivious
```

#### 9. 部署监控（可选）

```bash
kubectl apply -f deploy/kubernetes/monitoring.yaml
```

#### 10. 检查服务

```bash
# 检查所有 Pod
kubectl get pods -n oblivious

# 检查所有 Service
kubectl get svc -n oblivious

# 检查 Ingress
kubectl get ingress -n oblivious
```

---

## 扩缩容配置

### 手动扩容

```bash
# 扩容网关
kubectl scale deployment gateway -n oblivious --replicas=5

# 扩容对话服务
kubectl scale deployment chat-service -n oblivious --replicas=5

# 扩容前端
kubectl scale deployment frontend -n oblivious --replicas=5
```

### 自动扩缩容 (HPA)

```bash
# 为网关创建 HPA
kubectl autoscale deployment gateway -n oblivious --min=2 --max=10 --cpu-percent=70

# 查看 HPA 状态
kubectl get hpa -n oblivious
```

---

## 监控和日志

### 查看日志

```bash
# 查看 Pod 日志
kubectl logs -f pod/gateway-xxxxx -n oblivious

# 查看所有 Pod 日志
kubectl logs -f deployment/gateway -n oblivious

# 实时查看
kubectl logs -f deployment/gateway -n oblivious --timestamps=true
```

### 监控资源使用

```bash
# 查看 Pod 资源使用
kubectl top pod -n oblivious

# 查看节点资源使用
kubectl top nodes
```

### 访问 Prometheus

```bash
kubectl port-forward svc/prometheus 9090:9090 -n oblivious
# 访问 http://localhost:9090
```

### 访问 Grafana

```bash
kubectl port-forward svc/grafana 3001:3000 -n oblivious
# 访问 http://localhost:3001
# 默认用户名: admin
# 默认密码: admin
```

---

## 故障排除

### Pod 无法启动

```bash
# 查看 Pod 详细信息
kubectl describe pod <pod-name> -n oblivious

# 查看 Pod 日志
kubectl logs <pod-name> -n oblivious

# 检查事件
kubectl get events -n oblivious --sort-by='.lastTimestamp'
```

### 数据库连接失败

```bash
# 检查 PostgreSQL Pod
kubectl get pods -n oblivious | grep postgres

# 检查 PostgreSQL 日志
kubectl logs -f deployment/postgres -n oblivious

# 测试连接
kubectl run psql -n oblivious --image=postgres:16-alpine \
  --rm -it --restart=Never -- \
  psql -h postgres -U postgres -d oblivious
```

### 服务无法通信

```bash
# 检查服务 DNS
kubectl run -it --rm debug --image=nicolaka/netshoot -n oblivious -- bash

# 在容器内测试
nslookup gateway
curl -v http://gateway:8080/health
exit
```

### 高 CPU/内存 使用

```bash
# 查看资源使用最多的 Pod
kubectl top pods -n oblivious --sort-by=memory

# 增加资源限制
kubectl set resources deployment gateway -n oblivious --limits=cpu=1000m,memory=1Gi
```

---

## 生产环境最佳实践

### 1. 安全性

```bash
# 使用 RBAC
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ServiceAccount
metadata:
  name: oblivious
  namespace: oblivious
EOF

# 使用 NetworkPolicy 限制流量
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: oblivious-default-deny
  namespace: oblivious
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: oblivious
EOF
```

### 2. 备份

```bash
# 备份 PostgreSQL
kubectl exec -it deployment/postgres -n oblivious -- \
  pg_dump -U postgres oblivious > backup.sql

# 恢复 PostgreSQL
cat backup.sql | kubectl exec -i deployment/postgres -n oblivious -- \
  psql -U postgres oblivious
```

### 3. 更新应用

```bash
# 使用 RollingUpdate 更新镜像
kubectl set image deployment/gateway gateway=your-registry/oblivious-backend:2.0.0 \
  -n oblivious --record

# 查看更新状态
kubectl rollout status deployment/gateway -n oblivious

# 回滚更新
kubectl rollout undo deployment/gateway -n oblivious
```

---

## 常用命令速查表

```bash
# 创建资源
kubectl apply -f <file>

# 删除资源
kubectl delete -f <file>

# 查看资源
kubectl get <resource> -n oblivious
kubectl describe <resource> <name> -n oblivious

# 查看日志
kubectl logs <pod-name> -n oblivious
kubectl logs -f <pod-name> -n oblivious

# 执行命令
kubectl exec -it <pod-name> -n oblivious -- bash

# 端口转发
kubectl port-forward svc/<service-name> <local-port>:<pod-port> -n oblivious

# 查看事件
kubectl get events -n oblivious

# 查看资源使用
kubectl top nodes
kubectl top pods -n oblivious
```

---

## 相关链接

- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [kubectl 命令参考](https://kubernetes.io/docs/reference/kubectl/)

---

**最后更新**: 2025年11月20日

