#!/bin/bash

# 启动后端服务的脚本

set -e

echo "════════════════════════════════════════════════════════════"
echo "🚀 启动 Oblivious 后端服务"
echo "════════════════════════════════════════════════════════════"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 创建 .env 文件（如果不存在）
if [ ! -f ".env" ]; then
    echo "📝 创建 .env 文件..."
    cat > .env << 'EOF'
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=oblivious
DB_PORT=5433
REDIS_PORT=6379
APP_ENV=development
JWT_SECRET=your-super-secret-jwt-key
EOF
    echo "✅ .env 文件已创建"
    echo ""
fi

# 启动服务
echo "🚀 使用 Docker Compose 启动服务..."
docker compose -f docker-compose-backend-only.yml up -d

# 等待服务启动
echo ""
echo "⏳ 等待服务启动（约 10 秒）..."
sleep 10

# 显示服务状态
echo ""
echo "════════════════════════════════════════════════════════════"
echo "✅ 服务启动状态"
echo "════════════════════════════════════════════════════════════"
echo ""

docker compose -f docker-compose-backend-only.yml ps

echo ""
echo "📝 API 访问地址:"
echo "   🌐 API 网关:    http://localhost:8080"
echo "   👤 用户服务:    http://localhost:8081"
echo "   💬 对话服务:    http://localhost:8082"
echo "   🔄 中转服务:    http://localhost:8083"
echo "   🗄️  PostgreSQL:  localhost:5433"
echo "   💾 Redis:       localhost:6379"
echo ""

# 测试 API
echo "🧪 测试 API..."
sleep 2

if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo "✅ API 网关正常运行"
else
    echo "⚠️ API 网关可能还在启动中，请稍后重试"
fi

echo ""
echo "💡 常用命令:"
echo "   查看日志:        docker compose -f docker-compose-backend-only.yml logs -f"
echo "   查看特定服务:    docker compose -f docker-compose-backend-only.yml logs -f gateway"
echo "   停止服务:        docker compose -f docker-compose-backend-only.yml down"
echo "   重启服务:        docker compose -f docker-compose-backend-only.yml restart"
echo ""
echo "📚 下一步:"
echo "   1. 测试用户注册: curl -X POST http://localhost:8080/api/v1/register -H 'Content-Type: application/json' -d '{\"username\":\"test\",\"email\":\"test@test.com\",\"password\":\"Pass123!\"'"
echo "   2. 查看完整部署指南: cat DEPLOYMENT_GUIDE.md"
echo "   3. 前端部署: 需要修复前端 TypeScript 错误后再构建"
echo ""
echo "🎉 启动完成！"

