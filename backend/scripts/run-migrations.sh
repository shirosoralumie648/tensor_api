#!/bin/bash

# run-migrations.sh
# 在 Docker 容器中运行数据库迁移

set -e

POSTGRES_HOST=${DATABASE_HOST:-postgres}
POSTGRES_PORT=${DATABASE_PORT:-5432}
POSTGRES_USER=${DATABASE_USER:-postgres}
POSTGRES_PASSWORD=${DATABASE_PASSWORD:-password}
POSTGRES_DB=${DATABASE_NAME:-oblivious}

echo "连接到 PostgreSQL: $POSTGRES_HOST:$POSTGRES_PORT"

# 等待 PostgreSQL 就绪
echo "等待 PostgreSQL 启动..."
for i in {1..30}; do
  if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d postgres -c "SELECT 1" 2>/dev/null; then
    echo "✅ PostgreSQL 已就绪"
    break
  fi
  echo "尝试连接... ($i/30)"
  sleep 1
done

# 创建数据库（如果不存在）
echo "创建数据库 $POSTGRES_DB..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d postgres -c "CREATE DATABASE $POSTGRES_DB;" 2>/dev/null || echo "数据库已存在"

# 启用 pgvector 扩展
echo "启用 pgvector 扩展..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c "CREATE EXTENSION IF NOT EXISTS vector;" 2>/dev/null || true

# 运行所有迁移文件
echo "运行数据库迁移..."
MIGRATION_DIR="/app/migrations"

# 确保 schema_migrations 表存在
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB << EOF
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty BOOLEAN NOT NULL DEFAULT FALSE
);
EOF

# 遍历所有 .up.sql 文件并执行
for file in $(ls $MIGRATION_DIR/*.up.sql 2>/dev/null | sort); do
  filename=$(basename "$file")
  version=$(echo "$filename" | sed 's/_create.*//' | sed 's/^0*//')
  
  # 检查是否已执行
  if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1 FROM schema_migrations WHERE version = $version;" 2>/dev/null | grep -q 1; then
    echo "执行迁移: $filename"
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -f "$file" 2>&1 || true
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c "INSERT INTO schema_migrations (version, dirty) VALUES ($version, FALSE) ON CONFLICT DO NOTHING;" 2>/dev/null || true
  fi
done

echo "✅ 迁移完成！"

# 显示表列表
echo ""
echo "📊 已创建的表:"
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c "\dt" || true

