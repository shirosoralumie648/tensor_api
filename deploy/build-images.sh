#!/bin/bash

# 构建 Docker 镜像脚本

set -e

echo "════════════════════════════════════════════════════════════"
echo "开始构建 Docker 镜像"
echo "════════════════════════════════════════════════════════════"
echo ""

# 进入项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

echo "📦 项目根目录: $(pwd)"
echo ""

# 构建后端镜像
echo "🔨 构建后端镜像 (oblivious-backend:latest)..."
docker build \
  -f "$PROJECT_ROOT/deploy/docker/Dockerfile.backend" \
  -t oblivious-backend:latest \
  "$PROJECT_ROOT"
echo "✅ 后端镜像构建完成"
echo ""

# 构建前端镜像
echo "🔨 构建前端镜像 (oblivious-frontend:latest)..."
docker build \
  -f "$PROJECT_ROOT/deploy/docker/Dockerfile.frontend" \
  -t oblivious-frontend:latest \
  "$PROJECT_ROOT"
echo "✅ 前端镜像构建完成"
echo ""

# 显示镜像信息
echo "📊 已构建的镜像:"
docker images | grep -E "oblivious-backend|oblivious-frontend"
echo ""

echo "════════════════════════════════════════════════════════════"
echo "✅ 镜像构建完成！"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "接下来可以运行:"
echo "  cd deploy"
echo "  docker-compose up -d"

