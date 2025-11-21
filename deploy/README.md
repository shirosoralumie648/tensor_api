# Oblivious 部署

本目录包含 Docker 和 Kubernetes 部署配置。

## 快速开始

### Docker（本地开发）

```bash
# 构建镜像
docker build -f docker/Dockerfile.backend -t oblivious-backend:latest ..
docker build -f docker/Dockerfile.frontend -t oblivious-frontend:latest ..

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止所有服务
docker-compose down
```

### Kubernetes（生产环境）

```bash
# 使用部署脚本
chmod +x deploy.sh
./deploy.sh kubernetes

# 或者手动部署
# 1. 创建命名空间
kubectl apply -f kubernetes/namespace.yaml

# 2. 部署基础设施
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/redis.yaml

# 3. 部署服务
kubectl apply -f kubernetes/backend-services.yaml
kubectl apply -f kubernetes/frontend.yaml

# 4. （可选）部署监控
kubectl apply -f kubernetes/monitoring.yaml
```

## 文件结构

```
deploy/
├── docker/                    # Docker 配置
│   ├── Dockerfile.backend     # 后端多阶段构建
│   └── Dockerfile.frontend    # 前端多阶段构建
├── kubernetes/                # Kubernetes 配置
│   ├── namespace.yaml         # 命名空间
│   ├── postgres.yaml          # PostgreSQL 部署
│   ├── redis.yaml             # Redis 部署
│   ├── backend-services.yaml  # 后端微服务
│   ├── frontend.yaml          # 前端部署
│   ├── ingress.yaml           # Ingress 配置
│   └── monitoring.yaml        # Prometheus & Grafana
├── docker-compose.yml         # Docker Compose 完整配置
├── deploy.sh                  # 快速部署脚本
└── DEPLOYMENT_GUIDE.md        # 详细部署指南
```

## 详细指南

请查看 [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) 获取完整的部署说明。

## Docker Compose 环境变量

编辑 `.env` 文件来自定义配置：

```bash
# 数据库
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=oblivious
DB_PORT=5433

# Redis
REDIS_PORT=6379

# 应用
APP_ENV=development
JWT_SECRET=your-super-secret-key
```

## Kubernetes 扩缩容

### 手动扩容

```bash
kubectl scale deployment gateway -n oblivious --replicas=5
```

### 自动扩缩容

```bash
kubectl autoscale deployment gateway -n oblivious --min=2 --max=10 --cpu-percent=70
```

## 监控

### Prometheus

```bash
kubectl port-forward svc/prometheus 9090:9090 -n oblivious
# 访问 http://localhost:9090
```

### Grafana

```bash
kubectl port-forward svc/grafana 3001:3000 -n oblivious
# 访问 http://localhost:3001
# 用户名: admin
# 密码: admin
```

## 常见问题

### 如何查看服务日志？

**Docker:**
```bash
docker-compose logs -f <service-name>
```

**Kubernetes:**
```bash
kubectl logs -f deployment/<deployment-name> -n oblivious
```

### 如何更新应用？

**Docker:**
```bash
# 停止服务
docker-compose down

# 重建镜像
docker build -f docker/Dockerfile.backend -t oblivious-backend:latest ..

# 重新启动
docker-compose up -d
```

**Kubernetes:**
```bash
# 更新镜像
kubectl set image deployment/gateway gateway=your-registry/oblivious-backend:2.0.0 -n oblivious

# 查看更新状态
kubectl rollout status deployment/gateway -n oblivious
```

### 如何备份数据库？

**Docker:**
```bash
docker-compose exec postgres pg_dump -U postgres oblivious > backup.sql
```

**Kubernetes:**
```bash
kubectl exec -it deployment/postgres -n oblivious -- \
  pg_dump -U postgres oblivious > backup.sql
```

## 相关链接

- [Docker 官方文档](https://docs.docker.com/)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [项目主目录](../)

