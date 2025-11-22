#!/bin/bash

set -e

echo "🚀 Initializing Oblivious Backend..."

# 检查环境变量
if [ -z "$DATABASE_URL" ]; then
    echo "⚠️  DATABASE_URL not set, using default"
    export DATABASE_URL="host=localhost user=postgres password=postgres dbname=oblivious port=5432 sslmode=disable"
fi

# 1. 检查PostgreSQL连接
echo "📡 Checking database connection..."
if ! psql "$DATABASE_URL" -c "SELECT 1" > /dev/null 2>&1; then
    echo "❌ Cannot connect to database"
    echo "   Please ensure PostgreSQL is running and DATABASE_URL is correct"
    exit 1
fi
echo "✅ Database connection OK"

# 2. 运行数据库迁移
echo "📦 Running database migrations..."
for migration in migrations/*.sql; do
    if [ -f "$migration" ]; then
        echo "   Executing: $migration"
        psql "$DATABASE_URL" -f "$migration" || {
            echo "⚠️  Migration failed: $migration (may already exist)"
        }
    fi
done
echo "✅ Migrations completed"

# 3. 安装Go依赖
echo "📦 Installing Go dependencies..."
go mod download
echo "✅ Dependencies installed"

# 4. 构建应用
echo "🔨 Building application..."
go build -o bin/oblivious cmd/server/main_example.go
echo "✅ Build completed"

# 5. 启动服务
echo "🚀 Starting server..."
./bin/oblivious
