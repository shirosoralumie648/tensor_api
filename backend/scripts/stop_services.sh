#!/bin/bash

# 停止所有微服务

echo "================================"
echo "停止 Oblivious 微服务"
echo "================================"
echo ""

cd /home/shirosora/windsurf-storage/oblivious/backend

if [ -f logs/gateway.pid ]; then
    GATEWAY_PID=$(cat logs/gateway.pid)
    if kill -0 $GATEWAY_PID 2>/dev/null; then
        echo "停止网关服务 (PID: ${GATEWAY_PID})..."
        kill $GATEWAY_PID
    fi
    rm -f logs/gateway.pid
fi

if [ -f logs/user.pid ]; then
    USER_PID=$(cat logs/user.pid)
    if kill -0 $USER_PID 2>/dev/null; then
        echo "停止用户服务 (PID: ${USER_PID})..."
        kill $USER_PID
    fi
    rm -f logs/user.pid
fi

if [ -f logs/chat.pid ]; then
    CHAT_PID=$(cat logs/chat.pid)
    if kill -0 $CHAT_PID 2>/dev/null; then
        echo "停止对话服务 (PID: ${CHAT_PID})..."
        kill $CHAT_PID
    fi
    rm -f logs/chat.pid
fi

echo ""
echo "✅ 所有服务已停止"

