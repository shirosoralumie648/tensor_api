#!/bin/bash

# 开发模式启动脚本

set -e

echo "🔧 Starting Oblivious in development mode..."

# 设置开发环境变量
export GIN_MODE="debug"
export DATABASE_URL="${DATABASE_URL:-host=localhost user=postgres password=postgres dbname=oblivious port=5432 sslmode=disable}"

# 使用air进行热重载（如果安装了）
if command -v air &> /dev/null; then
    echo "🔄 Using air for hot reload..."
    air
else
    echo "💡 Tip: Install air for hot reload: go install github.com/cosmtrek/air@latest"
    echo "🚀 Starting with go run..."
    go run cmd/server/main_example.go
fi
