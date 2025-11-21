#!/bin/bash

# 快速启动脚本

set -e

echo "════════════════════════════════════════════════════════════"
echo "🚀 Oblivious 快速启动"
echo "════════════════════════════════════════════════════════════"
echo ""

# 获取脚本目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 检查 .env 文件
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

# 检查镜像是否存在
if ! docker image inspect oblivious-backend:latest >/dev/null 2>&1; then
    echo "⚠️  后端镜像不存在，开始构建..."
    bash build-images.sh
    echo ""
fi

if ! docker image inspect oblivious-frontend:latest >/dev/null 2>&1; then
    echo "⚠️  前端镜像不存在，开始构建..."
    bash build-images.sh
    echo ""
fi

# 启动容器
echo "🚀 启动 Docker Compose..."
docker compose up -d

# 等待服务启动
echo ""
echo "⏳ 等待服务启动..."
sleep 5

# 显示服务状态
echo ""
echo "════════════════════════════════════════════════════════════"
echo "✅ 所有服务已启动"
echo "════════════════════════════════════════════════════════════"
echo ""

docker compose ps

echo ""
echo "📝 访问地址:"
echo "   前端:        http://localhost:3000"
echo "   API 网关:    http://localhost:8080"
echo "   用户服务:    http://localhost:8081"
echo "   对话服务:    http://localhost:8082"
echo "   中转服务:    http://localhost:8083"
echo ""
echo "💡 常用命令:"
echo "   查看日志:    docker compose logs -f"
echo "   停止服务:    docker compose down"
echo "   重启服务:    docker compose restart"
echo ""
echo "🎉 快速启动完成！"

