# 生产部署指南

## 概述

本文档提供 Oblivious 平台在生产环境的完整部署指南，涵盖 Kubernetes 集群部署、监控配置、安全加固等内容。

## 前置条件

### 基础设施要求

- **Kubernetes 集群**：v1.24+
- **Helm**：v3.0+
- **kubectl**：与集群版本匹配
- **域名**：已备案的域名及 SSL 证书
- **存储**：支持 PVC 的存储类（推荐 Ceph/NFS）

### 资源配置建议

#### 最小配置（适合测试/小规模）

| 组件 | CPU | 内存 | 存储 | 副本数 |
|-----|-----|------|------|-------|
| Gateway | 0.5核 | 512MB | - | 2 |
| User Service | 0.25核 | 256MB | - | 2 |
| Chat Service | 0.5核 | 512MB | - | 2 |
| Relay Service | 1核 | 1GB | - | 2 |
| PostgreSQL | 2核 | 4GB | 50GB | 1 |
| Redis | 0.5核 | 1GB | 10GB | 3 |

#### 推荐配置（生产环境）

| 组件 | CPU | 内存 | 存储 | 副本数 |
|-----|-----|------|------|-------|
| Gateway | 2核 | 2GB | - | 3 |
| User Service | 1核 | 1GB | - | 3 |
| Chat Service | 2核 | 2GB | - | 3 |
| Relay Service | 4核 | 4GB | - | 3 |
| PostgreSQL | 4核 | 8GB | 200GB | 3 (主从) |
| Redis | 2核 | 4GB | 20GB | 6 (集群) |
| MinIO | 2核 | 4GB | 500GB | 4 |

## 部署步骤

### 1. 准备 Kubernetes 集群

#### 创建命名空间

```bash
kubectl create namespace oblivious-prod
kubectl label namespace oblivious-prod env=production
```

#### 配置 RBAC

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: oblivious-sa
  namespace: oblivious-prod
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: oblivious-role
  namespace: oblivious-prod
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: oblivious-rolebinding
  namespace: oblivious-prod
subjects:
- kind: ServiceAccount
  name: oblivious-sa
  namespace: oblivious-prod
roleRef:
  kind: Role
  name: oblivious-role
  apiGroup: rbac.authorization.k8s.io
```

```bash
kubectl apply -f rbac.yaml
```

### 2. 配置 Secrets

#### 创建数据库密钥

```bash
kubectl create secret generic db-secret \
  --from-literal=username=oblivious \
  --from-literal=password='YOUR_STRONG_PASSWORD' \
  --from-literal=database=oblivious \
  -n oblivious-prod
```

#### 创建 JWT 密钥

```bash
kubectl create secret generic jwt-secret \
  --from-literal=secret-key='YOUR_JWT_SECRET_KEY' \
  -n oblivious-prod
```

#### 创建 AI API 密钥

```bash
kubectl create secret generic ai-api-keys \
  --from-literal=openai-key='sk-...' \
  --from-literal=claude-key='sk-ant-...' \
  -n oblivious-prod
```

#### 创建 TLS 证书

```bash
kubectl create secret tls tls-cert \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  -n oblivious-prod
```

### 3. 配置 ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oblivious-config
  namespace: oblivious-prod
data:
  app.env: "production"
  log.level: "info"
  redis.addr: "redis-cluster:6379"
  postgres.host: "postgres-primary"
  postgres.port: "5432"
  minio.endpoint: "minio:9000"
  rabbitmq.host: "rabbitmq"
```

```bash
kubectl apply -f configmap.yaml
```

### 4. 部署数据库（PostgreSQL）

使用 StatefulSet 部署高可用 PostgreSQL：

```yaml
# postgres-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: oblivious-prod
spec:
  serviceName: postgres
  replicas: 3
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        - name: POSTGRES_DB
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: database
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            cpu: 2000m
            memory: 4Gi
          limits:
            cpu: 4000m
            memory: 8Gi
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 200Gi
```

```bash
kubectl apply -f postgres-statefulset.yaml
```

### 5. 部署 Redis 集群

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install redis bitnami/redis-cluster \
  --namespace oblivious-prod \
  --set cluster.nodes=6 \
  --set persistence.size=20Gi
```

### 6. 部署 MinIO

```bash
helm repo add minio https://charts.min.io/
helm install minio minio/minio \
  --namespace oblivious-prod \
  --set mode=distributed \
  --set replicas=4 \
  --set persistence.size=500Gi
```

### 7. 部署微服务

#### Gateway 部署

```yaml
# gateway-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
  namespace: oblivious-prod
spec:
  replicas: 3
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
        image: oblivious/gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret-key
        envFrom:
        - configMapRef:
            name: oblivious-config
        resources:
          requests:
            cpu: 1000m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 2Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: gateway
  namespace: oblivious-prod
spec:
  selector:
    app: gateway
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

```bash
kubectl apply -f gateway-deployment.yaml
```

#### 其他服务部署

类似地部署 User、Chat、Relay、Billing 等服务。

### 8. 配置 Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: oblivious-ingress
  namespace: oblivious-prod
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.oblivious.ai
    - app.oblivious.ai
    secretName: tls-cert
  rules:
  - host: api.oblivious.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gateway
            port:
              number: 80
  - host: app.oblivious.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend
            port:
              number: 80
```

```bash
kubectl apply -f ingress.yaml
```

### 9. 配置自动扩缩容（HPA）

```yaml
# gateway-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway-hpa
  namespace: oblivious-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gateway
  minReplicas: 3
  maxReplicas: 10
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

```bash
kubectl apply -f gateway-hpa.yaml
```

### 10. 数据库迁移

```bash
# 创建迁移 Job
kubectl create job migrate-db \
  --image=oblivious/migrate:latest \
  -n oblivious-prod \
  -- migrate -path /migrations -database "postgresql://..." up
```

## 监控和日志

### 1. 部署 Prometheus

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

### 2. 部署 Loki（日志聚合）

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set grafana.enabled=true
```

### 3. 配置 ServiceMonitor

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: oblivious-services
  namespace: oblivious-prod
spec:
  selector:
    matchLabels:
      monitoring: "true"
  endpoints:
  - port: metrics
    interval: 30s
```

## 安全加固

### 1. 网络策略

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-ingress
  namespace: oblivious-prod
spec:
  podSelector: {}
  policyTypes:
  - Ingress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-gateway
  namespace: oblivious-prod
spec:
  podSelector:
    matchLabels:
      app: gateway
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
```

### 2. Pod Security Policy

```yaml
# pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
  - ALL
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
  - configMap
  - secret
  - persistentVolumeClaim
```

### 3. 敏感信息加密

使用 Sealed Secrets 或 External Secrets Operator。

## 备份策略

### 数据库备份

```bash
# 创建 CronJob 定时备份
kubectl create -f - <<EOF
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
  namespace: oblivious-prod
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15-alpine
            command:
            - /bin/sh
            - -c
            - pg_dump -h postgres -U oblivious oblivious | gzip > /backup/backup-\$(date +%Y%m%d-%H%M%S).sql.gz
            volumeMounts:
            - name: backup-volume
              mountPath: /backup
          volumes:
          - name: backup-volume
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
EOF
```

### MinIO 数据备份

配置 MinIO 镜像同步到另一个存储桶或 S3。

## 性能优化

### 1. 数据库优化

```sql
-- 调整 PostgreSQL 参数
ALTER SYSTEM SET shared_buffers = '2GB';
ALTER SYSTEM SET effective_cache_size = '6GB';
ALTER SYSTEM SET maintenance_work_mem = '512MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
```

### 2. Redis 优化

```bash
# 配置 Redis 参数
kubectl exec -it redis-0 -n oblivious-prod -- redis-cli
CONFIG SET maxmemory 4gb
CONFIG SET maxmemory-policy allkeys-lru
```

### 3. 启用 CDN

为静态资源配置 CDN 加速。

## 滚动更新

### 零停机更新

```bash
# 更新镜像
kubectl set image deployment/gateway \
  gateway=oblivious/gateway:v1.1.0 \
  -n oblivious-prod

# 查看更新状态
kubectl rollout status deployment/gateway -n oblivious-prod

# 回滚
kubectl rollout undo deployment/gateway -n oblivious-prod
```

### 金丝雀发布

```yaml
# canary-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway-canary
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gateway
      version: canary
  template:
    metadata:
      labels:
        app: gateway
        version: canary
    spec:
      containers:
      - name: gateway
        image: oblivious/gateway:v1.1.0-canary
```

## 故障排查

### 查看日志

```bash
# Pod 日志
kubectl logs -f deployment/gateway -n oblivious-prod

# 多个 Pod 日志
kubectl logs -l app=gateway -n oblivious-prod --tail=100

# 前一个容器日志
kubectl logs pod-name -n oblivious-prod --previous
```

### 进入容器调试

```bash
kubectl exec -it pod-name -n oblivious-prod -- /bin/sh
```

### 查看事件

```bash
kubectl get events -n oblivious-prod --sort-by='.lastTimestamp'
```

## 灾难恢复

### 数据恢复流程

1. 停止所有服务
2. 恢复数据库备份
3. 恢复 MinIO 数据
4. 验证数据完整性
5. 重启服务

```bash
# 恢复数据库
gunzip < backup-20240101.sql.gz | kubectl exec -i postgres-0 -n oblivious-prod -- psql -U oblivious
```

## 相关文档

- [快速部署参考](DEPLOYMENT_QUICK_REFERENCE.md)
- [架构设计](ARCHITECTURE.md)
- [监控配置](MONITORING.md)
