#!/bin/bash

# 启动所有微服务

set -e

echo "================================"
echo "启动 Oblivious 微服务"
echo "================================"
echo ""

# 设置工作目录
cd /home/shirosora/windsurf-storage/oblivious/backend

# 加载环境变量
set -a
source env.test
set +a

# 检查数据库是否已初始化
echo "📊 检查数据库状态..."
export PATH=/usr/local/go/bin:$PATH
if ! make migrate-status 2>/dev/null | grep -q "up"; then
    echo "⚠️  数据库未初始化，正在运行迁移..."
    make migrate-up
fi

echo ""
echo "🔨 编译服务..."
go build -o bin/user ./cmd/user
go build -o bin/chat ./cmd/chat
go build -o bin/gateway ./cmd/gateway

echo ""
echo "🚀 启动用户服务 (端口 8081)..."
env DATABASE_HOST=$DATABASE_HOST DATABASE_PORT=$DATABASE_PORT DATABASE_USER=$DATABASE_USER DATABASE_PASSWORD=$DATABASE_PASSWORD DATABASE_NAME=$DATABASE_NAME REDIS_HOST=$REDIS_HOST REDIS_PORT=$REDIS_PORT REDIS_PASSWORD=$REDIS_PASSWORD JWT_SECRET=$JWT_SECRET JWT_EXPIRE_HOURS=$JWT_EXPIRE_HOURS REFRESH_TOKEN_EXPIRE_DAYS=$REFRESH_TOKEN_EXPIRE_DAYS APP_ENV=$APP_ENV ./bin/user > logs/user.log 2>&1 &
USER_PID=$!
echo "用户服务 PID: ${USER_PID}"

sleep 2

echo ""
echo "🚀 启动对话服务 (端口 8082)..."
env DATABASE_HOST=$DATABASE_HOST DATABASE_PORT=$DATABASE_PORT DATABASE_USER=$DATABASE_USER DATABASE_PASSWORD=$DATABASE_PASSWORD DATABASE_NAME=$DATABASE_NAME REDIS_HOST=$REDIS_HOST REDIS_PORT=$REDIS_PORT REDIS_PASSWORD=$REDIS_PASSWORD JWT_SECRET=$JWT_SECRET JWT_EXPIRE_HOURS=$JWT_EXPIRE_HOURS REFRESH_TOKEN_EXPIRE_DAYS=$REFRESH_TOKEN_EXPIRE_DAYS APP_ENV=$APP_ENV ./bin/chat > logs/chat.log 2>&1 &
CHAT_PID=$!
echo "对话服务 PID: ${CHAT_PID}"

sleep 2

echo ""
echo "🚀 启动中转服务 (端口 8083)..."
env DATABASE_HOST=$DATABASE_HOST DATABASE_PORT=$DATABASE_PORT DATABASE_USER=$DATABASE_USER DATABASE_PASSWORD=$DATABASE_PASSWORD DATABASE_NAME=$DATABASE_NAME REDIS_HOST=$REDIS_HOST REDIS_PORT=$REDIS_PORT REDIS_PASSWORD=$REDIS_PASSWORD JWT_SECRET=$JWT_SECRET JWT_EXPIRE_HOURS=$JWT_EXPIRE_HOURS REFRESH_TOKEN_EXPIRE_DAYS=$REFRESH_TOKEN_EXPIRE_DAYS APP_ENV=$APP_ENV ./bin/relay > logs/relay.log 2>&1 &
RELAY_PID=$!
echo "中转服务 PID: ${RELAY_PID}"

sleep 2

echo ""
echo "🚀 启动网关服务 (端口 8080)..."
env DATABASE_HOST=$DATABASE_HOST DATABASE_PORT=$DATABASE_PORT DATABASE_USER=$DATABASE_USER DATABASE_PASSWORD=$DATABASE_PASSWORD DATABASE_NAME=$DATABASE_NAME REDIS_HOST=$REDIS_HOST REDIS_PORT=$REDIS_PORT REDIS_PASSWORD=$REDIS_PASSWORD JWT_SECRET=$JWT_SECRET JWT_EXPIRE_HOURS=$JWT_EXPIRE_HOURS REFRESH_TOKEN_EXPIRE_DAYS=$REFRESH_TOKEN_EXPIRE_DAYS APP_ENV=$APP_ENV ./bin/gateway > logs/gateway.log 2>&1 &
GATEWAY_PID=$!
echo "网关服务 PID: ${GATEWAY_PID}"

sleep 2

echo ""
echo "================================"
echo "✅ 所有服务已启动"
echo "================================"
echo ""
echo "服务信息:"
echo "- 网关服务: http://localhost:8080 (PID: ${GATEWAY_PID})"
echo "- 用户服务: http://localhost:8081 (PID: ${USER_PID})"
echo "- 对话服务: http://localhost:8082 (PID: ${CHAT_PID})"
echo "- 中转服务: http://localhost:8083 (PID: ${RELAY_PID})"
echo ""
echo "日志文件:"
echo "- 网关服务: logs/gateway.log"
echo "- 用户服务: logs/user.log"
echo "- 对话服务: logs/chat.log"
echo "- 中转服务: logs/relay.log"
echo ""
echo "停止服务: kill ${GATEWAY_PID} ${USER_PID} ${CHAT_PID} ${RELAY_PID}"
echo ""
echo "保存 PID 到文件..."
echo "${GATEWAY_PID}" > logs/gateway.pid
echo "${USER_PID}" > logs/user.pid
echo "${CHAT_PID}" > logs/chat.pid
echo "${RELAY_PID}" > logs/relay.pid

