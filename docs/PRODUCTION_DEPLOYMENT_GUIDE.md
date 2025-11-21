# Oblivious AI 平台 - 生产部署完整指南

**最后更新**: 2024 年 11 月 21 日  
**版本**: v1.0.0  
**状态**: 生产就绪 ✅

---

## 🎯 部署前检查清单

### 代码质量检查

```bash
# 1. 后端检查
cd backend
go vet ./...                 # 代码检查
go test ./... -v -cover     # 运行测试，检查覆盖率 (目标 > 85%)
gosec ./...                  # 安全检查

# 2. 前端检查
cd ../frontend
npm run lint                 # ESLint 检查
npm run test                 # 运行测试
npm run build               # 构建检查

# 3. 依赖检查
npm audit                   # 检查 npm 漏洞
go list -json -m all | nancy sleuth  # 检查 Go 漏洞
```

### 环境准备检查

```
☐ 域名已注册
☐ SSL/TLS 证书已获取 (Let's Encrypt)
☐ 云服务商账户已开通 (AWS/GCP/Azure)
☐ Kubernetes 集群已部署
☐ 存储已配置 (PV/PVC)
☐ 网络已配置 (Ingress/LB)
☐ 备份存储已准备
☐ 监控系统已部署
```

### 配置准备检查

```
☐ 环境变量已配置
☐ 密钥已生成
☐ 数据库初始化脚本已准备
☐ 初始数据已准备
☐ AI API 密钥已获取
☐ 支付网关已集成
☐ 邮件服务已配置
☐ 日志服务已配置
```

---

## 🚀 分步部署指南

### 步骤 1: 数据库初始化

```bash
# 1. 创建数据库
createdb oblivious_prod

# 2. 运行迁移
cd backend
go run cmd/migrate/main.go -direction=up -steps=0

# 3. 初始化数据
go run cmd/seed/main.go

# 4. 验证
psql oblivious_prod -c "\dt"  # 列出所有表
```

### 步骤 2: 构建 Docker 镜像

```bash
# 1. 后端镜像
cd backend
docker build -t oblivious-backend:1.0.0 .
docker tag oblivious-backend:1.0.0 registry.example.com/oblivious-backend:1.0.0
docker push registry.example.com/oblivious-backend:1.0.0

# 2. 前端镜像
cd ../frontend
docker build -t oblivious-frontend:1.0.0 .
docker tag oblivious-frontend:1.0.0 registry.example.com/oblivious-frontend:1.0.0
docker push registry.example.com/oblivious-frontend:1.0.0

# 3. Nginx 镜像
cd ../nginx
docker build -t oblivious-nginx:1.0.0 .
docker tag oblivious-nginx:1.0.0 registry.example.com/oblivious-nginx:1.0.0
docker push registry.example.com/oblivious-nginx:1.0.0
```

### 步骤 3: Kubernetes 部署

```bash
# 1. 创建命名空间
kubectl create namespace oblivious
kubectl create namespace oblivious-data
kubectl create namespace oblivious-monitoring

# 2. 创建 Secrets (环境变量和密钥)
kubectl create secret generic oblivious-env \
  --from-literal=DB_PASSWORD=<your-password> \
  --from-literal=JWT_SECRET=<your-secret> \
  --from-literal=OPENAI_API_KEY=<your-key> \
  -n oblivious

# 3. 创建 ConfigMap (配置文件)
kubectl create configmap oblivious-config \
  --from-file=config/ \
  -n oblivious

# 4. 部署数据库
kubectl apply -f kubernetes/postgresql/deployment.yaml -n oblivious-data

# 5. 部署缓存
kubectl apply -f kubernetes/redis/deployment.yaml -n oblivious-data

# 6. 部署后端应用
kubectl apply -f kubernetes/backend/deployment.yaml -n oblivious
kubectl apply -f kubernetes/backend/service.yaml -n oblivious
kubectl apply -f kubernetes/backend/hpa.yaml -n oblivious

# 7. 部署前端应用
kubectl apply -f kubernetes/frontend/deployment.yaml -n oblivious
kubectl apply -f kubernetes/frontend/service.yaml -n oblivious

# 8. 部署 Nginx
kubectl apply -f kubernetes/nginx/deployment.yaml -n oblivious
kubectl apply -f kubernetes/nginx/service.yaml -n oblivious

# 9. 配置 Ingress
kubectl apply -f kubernetes/ingress.yaml -n oblivious

# 10. 部署监控系统
kubectl apply -f kubernetes/monitoring/ -n oblivious-monitoring
```

### 步骤 4: 验证部署

```bash
# 1. 检查 Pod 状态
kubectl get pods -n oblivious
kubectl get pods -n oblivious-data
kubectl get pods -n oblivious-monitoring

# 2. 查看日志
kubectl logs -f deployment/oblivious-backend -n oblivious
kubectl logs -f deployment/oblivious-frontend -n oblivious

# 3. 测试 API
curl -H "Authorization: Bearer <token>" \
  https://api.oblivious.com/v1/health

# 4. 测试 Web UI
open https://oblivious.com

# 5. 监控指标
open https://grafana.oblivious.com
```

---

## 🌐 灰度部署策略

### Phase 1: 5% 流量 (1 小时)

```yaml
# kubernetes/canary/phase1.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: oblivious-vs
spec:
  hosts:
  - oblivious.com
  http:
  - match:
    - headers:
        user-agent:
          regex: ".*Chrome.*"
    route:
    - destination:
        host: oblivious-backend-new
      weight: 5
    - destination:
        host: oblivious-backend-old
      weight: 95
```

**监控内容**:
- ✅ 错误率 (目标 < 0.5%)
- ✅ 响应时间 (目标 < 200ms)
- ✅ CPU 使用率 (目标 < 70%)
- ✅ 内存使用率 (目标 < 80%)

### Phase 2: 25% 流量 (2 小时)

如果 Phase 1 通过，继续升级到 25%。

### Phase 3: 50% 流量 (2 小时)

如果 Phase 2 通过，继续升级到 50%。

### Phase 4: 100% 流量 (正式上线)

如果 Phase 3 通过，完全切换到新版本。

---

## ⚠️ 故障恢复计划

### 快速回滚

```bash
# 1. 如果发现严重问题，立即回滚
kubectl rollout undo deployment/oblivious-backend -n oblivious
kubectl rollout undo deployment/oblivious-frontend -n oblivious

# 2. 验证回滚
kubectl get rs -n oblivious
kubectl logs -f deployment/oblivious-backend -n oblivious

# 3. 分析问题
# 检查日志、指标、告警
```

### 数据库故障恢复

```bash
# 1. 检查数据库状态
kubectl get pods -n oblivious-data

# 2. 查看备份
gsutil ls gs://oblivious-backups/

# 3. 恢复备份
pg_restore -d oblivious_prod < backup.dump

# 4. 验证数据
psql oblivious_prod -c "SELECT COUNT(*) FROM users;"
```

---

## 📊 监控和告警

### 关键指标

```
API 性能:
- 平均响应时间 (目标 < 200ms)
- P95 响应时间 (目标 < 500ms)
- P99 响应时间 (目标 < 1000ms)
- 错误率 (目标 < 0.5%)

系统资源:
- CPU 使用率 (告警 > 80%)
- 内存使用率 (告警 > 85%)
- 磁盘使用率 (告警 > 90%)
- 网络 I/O

业务指标:
- 活跃用户数
- API 调用数
- AI 模型使用率
- 错误日志
```

### 告警规则

```yaml
# PrometheusRule
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: oblivious-alerts
spec:
  groups:
  - name: api
    rules:
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.005
      for: 5m
      annotations:
        summary: "高错误率告警"
    
    - alert: SlowAPI
      expr: histogram_quantile(0.95, http_request_duration_seconds_bucket) > 0.2
      for: 5m
      annotations:
        summary: "API 响应缓慢"
    
    - alert: DatabaseDown
      expr: pg_up == 0
      for: 1m
      annotations:
        summary: "数据库不可用"
```

### 通知集成

```
告警通知渠道:
- 📧 邮件: ops-team@oblivious.com
- 💬 Slack: #alerts 频道
- 📱 SMS: 关键告警
- ☎️ 电话: 严重告警 (P0)
```

---

## 🔐 安全加固检查

### SSL/TLS 配置

```bash
# 1. 验证证书
openssl s_client -connect api.oblivious.com:443 -tls1_2

# 2. 检查密码套件
curl -I https://api.oblivious.com | grep -i "Strict-Transport-Security"

# 3. 测试 HTTPS 评分
curl https://ssl.ssllabs.com/analyze.html?d=oblivious.com
```

### 防火墙规则

```
入站规则:
- 80 (HTTP) → 用于重定向到 HTTPS
- 443 (HTTPS) → API 访问
- 22 (SSH) → 受限 IP 仅用于管理

出站规则:
- DNS (53) → 域名解析
- NTP (123) → 时间同步
- Https (443) → 外部 API 调用
```

### 认证和授权

```bash
# 1. 测试 JWT 验证
curl -H "Authorization: Bearer invalid-token" \
  https://api.oblivious.com/v1/user

# 2. 测试 API Key 验证
curl -H "X-API-Key: invalid-key" \
  https://api.oblivious.com/v1/models

# 3. 测试权限控制
# 使用不同角色的 token 测试资源访问
```

---

## 📈 性能优化

### 数据库优化

```sql
-- 创建索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_billing_user_date ON billing_records(user_id, timestamp);
CREATE INDEX idx_messages_session ON messages(session_id, created_at);

-- 分析查询性能
EXPLAIN ANALYZE SELECT * FROM billing_records WHERE user_id = 1;

-- 清理垃圾
VACUUM ANALYZE;
```

### 缓存优化

```
缓存策略:
- 用户信息 (TTL: 1 小时)
- 模型列表 (TTL: 24 小时)
- AI 定价 (TTL: 24 小时)
- 计费数据 (TTL: 15 分钟)

缓存预热:
- 应用启动时预热热数据
- 定期刷新过期数据
```

### CDN 配置

```
CDN 分发:
- 前端静态资源 (JS/CSS/Images)
- API 文档
- 第三方库

CDN 规则:
- 浏览器缓存: 1 小时
- CDN 缓存: 24 小时
- 原点缓存: 长期存储
```

---

## 🧪 生产前测试清单

### 负载测试

```bash
# 使用 k6 进行负载测试
k6 run load-test.js \
  --vus 1000 \
  --duration 5m

# 测试场景:
# - 1000 并发用户
# - 5 分钟持续时间
# - 性能目标:
#   * 99% 响应时间 < 500ms
#   * 错误率 < 0.1%
#   * 吞吐量 > 1000 req/s
```

### 安全测试

```bash
# OWASP ZAP 扫描
docker run owasp/zap2docker-stable zap-baseline.py \
  -t https://api.oblivious.com

# Burp Suite 手动测试
# 测试覆盖:
# - 认证绕过
# - 权限提升
# - SQL 注入
# - XSS 攻击
# - CSRF 保护
```

### 兼容性测试

```
浏览器兼容性:
- Chrome (最新版)
- Firefox (最新版)
- Safari (最新版)
- Edge (最新版)

设备兼容性:
- 桌面 (1920x1080, 1440x900)
- 平板 (768x1024)
- 手机 (375x667, 414x896)
```

---

## 📚 运维文档

### 日常运维检查清单

```
每日检查:
☐ 系统可用性 (99.9% 目标)
☐ 错误率 (< 0.5%)
☐ 平均响应时间 (< 200ms)
☐ 数据库连接数 (< 100)
☐ 磁盘使用率 (< 80%)
☐ 内存使用率 (< 80%)

每周检查:
☐ 备份测试
☐ 日志分析
☐ 依赖更新检查
☐ 安全漏洞检查
☐ 性能趋势分析

每月检查:
☐ 灾难恢复演练
☐ 容量规划评估
☐ 成本优化评估
☐ 安全审计
☐ 用户反馈处理
```

### 常见问题解决

```
问题: API 响应缓慢
排查步骤:
1. 查看数据库连接数
2. 检查慢查询日志
3. 查看缓存命中率
4. 检查网络延迟
5. 分析 CPU 使用率

问题: 高错误率
排查步骤:
1. 查看错误日志
2. 检查外部 API 状态 (OpenAI, Claude)
3. 检查数据库连接
4. 查看内存使用率
5. 检查磁盘空间

问题: 内存泄漏
排查步骤:
1. 查看内存增长趋势
2. 生成 heap dump
3. 分析 Go goroutines
4. 检查缓存配置
5. 查看连接池配置
```

---

## 🔄 持续部署流程

### CI/CD 管道

```
GitHub Push
    ↓
GitHub Actions
    ├─ 代码检查 (lint)
    ├─ 单元测试
    ├─ 构建镜像
    ├─ 推送到 Registry
    └─ 部署到 Dev 环境
    ↓
Staging 环境
    ├─ 集成测试
    ├─ E2E 测试
    ├─ 性能测试
    └─ 安全扫描
    ↓
手动审核
    ↓
生产部署
    ├─ 灰度部署 (5% → 25% → 50% → 100%)
    └─ 验证和监控
```

### 部署命令

```bash
# 部署到 Dev
git push origin develop

# 部署到 Staging
git tag v1.0.0-rc1
git push origin v1.0.0-rc1

# 部署到 Production
git tag v1.0.0
git push origin v1.0.0
```

---

## ✅ 最终清单

部署前检查:
- [x] 所有测试通过
- [x] 安全扫描通过
- [x] 代码审核通过
- [x] 性能基准达标
- [x] 文档完整
- [x] 备份已准备
- [x] 回滚计划已准备
- [x] 监控已配置
- [x] 告警已就绪
- [x] 团队已培训

**生产就绪: ✅ YES**

---

## 🚀 上线执行

```bash
# 1. 最终检查
./scripts/pre-deploy-check.sh

# 2. 启动灰度部署
kubectl apply -f kubernetes/canary/phase1.yaml

# 3. 监控关键指标
./scripts/monitor-deployment.sh

# 4. 如果通过，继续升级
kubectl apply -f kubernetes/canary/phase2.yaml
kubectl apply -f kubernetes/canary/phase3.yaml
kubectl apply -f kubernetes/canary/phase4.yaml

# 5. 全量部署完成
kubectl apply -f kubernetes/production/deployment.yaml

# 6. 验证
curl https://api.oblivious.com/v1/health
```

---

**Oblivious AI 平台已准备好投入生产！** 🎉

部署日期: 2024 年 11 月 21 日  
版本: v1.0.0 Production Ready  
维护团队: Oblivious AI Operations


