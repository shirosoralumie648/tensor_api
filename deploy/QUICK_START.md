# 🚀 Oblivious AI Platform - 快速开始

## ⚡ 一键部署（3分钟）

### 前提条件
- ✅ Docker Desktop 正在运行

### 执行命令

```bash
cd /home/shirosora/windsurf-storage/oblivious/deploy
./deploy-and-test.sh
```

就这么简单！脚本会自动完成所有部署和测试。

---

## 📱 部署后访问

| 服务 | 地址 | 说明 |
|------|------|------|
| 🌐 **前端** | http://localhost:3000 | 用户界面 |
| 🔌 **API** | http://localhost:8080 | API 网关 |
| 💡 **健康检查** | http://localhost:8080/health | 服务状态 |

---

## 🧪 手动测试

如果服务已经运行，快速验证：

```bash
./manual-test.sh
```

---

## 📊 查看结果

部署完成后会生成：

1. **测试日志**: `deployment-test-YYYYMMDD-HHMMSS.log`
2. **测试报告**: `test-report-YYYYMMDD-HHMMSS.md`

```bash
# 查看最新报告
ls -lt test-report-*.md | head -1 | awk '{print $NF}' | xargs cat
```

---

## 🛠️ 常用命令

```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f gateway

# 重启服务
docker compose restart

# 停止服务
docker compose down
```

---

## 🐛 问题排查

### Docker 未运行
```bash
# 启动 Docker Desktop 应用
```

### 端口被占用
```bash
# 修改 .env 文件中的端口配置
```

### 服务启动失败
```bash
# 查看日志
docker compose logs [service-name]

# 重新部署
docker compose down
docker compose up -d
```

---

## 📖 详细文档

- **完整指南**: `DEPLOYMENT_INSTRUCTIONS.md`
- **资源总结**: `DEPLOYMENT_SUMMARY.md`

---

**准备好了吗？开始部署吧！** 🎉

```bash
./deploy-and-test.sh
```
