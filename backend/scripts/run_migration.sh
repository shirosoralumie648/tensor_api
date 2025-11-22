#!/bin/bash

# 数据库迁移脚本
# Usage: ./scripts/run_migration.sh [up|down|status|sync]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MIGRATE_DIR="$PROJECT_ROOT/cmd/migrate"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查环境变量
check_env() {
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        warn ".env file not found, using default configuration"
        return
    fi
    
    source "$PROJECT_ROOT/.env"
    
    if [ -z "$DATABASE_HOST" ]; then
        error "DATABASE_HOST not set in .env"
        exit 1
    fi
    
    info "Database: ${DATABASE_USER}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}"
}

# 备份数据库
backup_db() {
    info "Creating database backup..."
    
    BACKUP_DIR="$PROJECT_ROOT/backups"
    mkdir -p "$BACKUP_DIR"
    
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    BACKUP_FILE="$BACKUP_DIR/backup_${TIMESTAMP}.sql"
    
    PGPASSWORD="$DATABASE_PASSWORD" pg_dump \
        -h "$DATABASE_HOST" \
        -p "$DATABASE_PORT" \
        -U "$DATABASE_USER" \
        -d "$DATABASE_NAME" \
        > "$BACKUP_FILE"
    
    if [ $? -eq 0 ]; then
        info "Backup created: $BACKUP_FILE"
        gzip "$BACKUP_FILE"
        info "Backup compressed: ${BACKUP_FILE}.gz"
    else
        error "Backup failed"
        exit 1
    fi
}

# 执行迁移
run_migration() {
    local command="${1:-up}"
    
    info "Running migration: $command"
    
    cd "$MIGRATE_DIR"
    go run main.go "$command"
    
    if [ $? -eq 0 ]; then
        info "Migration $command completed successfully!"
    else
        error "Migration $command failed"
        exit 1
    fi
}

# 主函数
main() {
    local command="${1:-up}"
    
    info "=== Oblivious Database Migration ===" 
    
    # 检查环境变量
    check_env
    
    # 对于 up 操作，先备份
    if [ "$command" == "up" ]; then
        warn "This will modify the database schema"
        read -p "Do you want to create a backup first? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            backup_db
        fi
    fi
    
    # 执行迁移
    run_migration "$command"
    
    info "=== Migration Complete ==="
}

# 运行
main "$@"
