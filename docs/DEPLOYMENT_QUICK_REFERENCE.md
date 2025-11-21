# 部署快速参考

## Docker Compose 部署

### 启动所有服务

```bash
cd deploy
docker-compose up -d
```

### 查看服务状态

```bash
docker-compose ps
docker-compose logs -f gateway
```

### 停止服务

```bash
docker-compose down
docker-compose down -v  # 同时删除数据卷
```

## Kubernetes 部署

### 创建命名空间

```bash
kubectl create namespace oblivious
kubectl config set-context --current --namespace=oblivious
```

### 应用配置

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secrets.yaml
```

### 部署基础设施

```bash
kubectl apply -f deploy/k8s/postgres.yaml
kubectl apply -f deploy/k8s/redis.yaml
kubectl apply -f deploy/k8s/minio.yaml
kubectl apply -f deploy/k8s/rabbitmq.yaml
```

### 部署微服务

```bash
kubectl apply -f deploy/k8s/deployments/
kubectl apply -f deploy/k8s/services/
kubectl apply -f deploy/k8s/ingress.yaml
```

### 配置自动扩缩容

```bash
kubectl apply -f deploy/k8s/hpa/
```

### 查看状态

```bash
kubectl get pods
kubectl get svc
kubectl logs -f deployment/gateway
```

## 数据库迁移

```bash
migrate -path ./backend/migrations \
  -database "postgresql://user:pass@host:5432/oblivious?sslmode=disable" \
  up
```

## 健康检查

```bash
# 网关
curl http://localhost:8080/health

# 各个服务
curl http://localhost:8081/health  # user
curl http://localhost:8082/health  # chat
curl http://localhost:8083/health  # relay
```

## 快速回滚

### Docker Compose

```bash
docker-compose down
git checkout <previous-commit>
docker-compose up -d
```

### Kubernetes

```bash
kubectl rollout undo deployment/gateway
kubectl rollout undo deployment/chat
```

## 常用命令

### Docker

```bash
# 查看日志
docker logs -f container_name

# 进入容器
docker exec -it container_name /bin/bash

# 重启容器
docker restart container_name

# 清理无用资源
docker system prune -a
```

### Kubernetes

```bash
# 扩容
kubectl scale deployment gateway --replicas=3

# 查看事件
kubectl get events --sort-by='.lastTimestamp'

# 删除 Pod（自动重建）
kubectl delete pod pod_name

# 查看资源使用
kubectl top nodes
kubectl top pods
```

## 监控面板

- **Grafana**：http://your-domain/grafana
- **Prometheus**：http://your-domain/prometheus
- **RabbitMQ**：http://your-domain/rabbitmq
- **MinIO**：http://your-domain/minio

## 相关文档

- [生产部署指南](PRODUCTION_DEPLOYMENT_GUIDE.md)
- [快速开始](QUICK_START.md)
