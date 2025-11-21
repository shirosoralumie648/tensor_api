#!/bin/bash

# 运行数据库迁移脚本

set -e

echo "════════════════════════════════════════════════════════════"
echo "🔄 运行数据库迁移"
echo "════════════════════════════════════════════════════════════"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# 检查 migrate 命令
if ! command -v migrate &> /dev/null; then
    echo "⚠️ migrate 命令未找到，正在安装..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# 获取数据库配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5433}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}
DB_NAME=${DB_NAME:-oblivious}

echo "📝 数据库配置:"
echo "   Host: $DB_HOST:$DB_PORT"
echo "   User: $DB_USER"
echo "   Database: $DB_NAME"
echo ""

# 构建数据库连接字符串
DATABASE_URL="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

echo "🔄 运行迁移..."
migrate -path "$PROJECT_ROOT/backend/migrations" -database "$DATABASE_URL" up

echo ""
echo "✅ 迁移完成！"
echo ""

# 验证表是否存在
echo "📊 验证数据库表..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt" 2>/dev/null | head -20

echo ""
echo "✅ 数据库已准备就绪！"

