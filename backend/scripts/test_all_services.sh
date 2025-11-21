#!/bin/bash

# 综合测试脚本 - 测试用户服务和对话服务

set -e

GATEWAY_URL="http://localhost:8080/api/v1"
USER_SERVICE_URL="http://localhost:8081/api/v1"
CHAT_SERVICE_URL="http://localhost:8082/api/v1"

echo "================================"
echo "Oblivious 服务测试"
echo "================================"
echo ""

# 生成随机用户名
RANDOM_USER="testuser_$(date +%s)"
EMAIL="${RANDOM_USER}@example.com"
PASSWORD="password123"

echo "📋 测试环境信息"
echo "----------------------------"
echo "Gateway URL: ${GATEWAY_URL}"
echo "User Service URL: ${USER_SERVICE_URL}"
echo "Chat Service URL: ${CHAT_SERVICE_URL}"
echo "测试用户: ${RANDOM_USER}"
echo ""

# ============================================
# Part 1: 用户服务测试
# ============================================

echo ""
echo "🔐 Part 1: 用户服务测试"
echo "================================"
echo ""

echo "1.1 测试用户注册"
echo "----------------------------"

REGISTER_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${RANDOM_USER}\",
    \"email\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "响应: ${REGISTER_RESPONSE}"

if echo "${REGISTER_RESPONSE}" | grep -q '"success":true'; then
    echo "✅ 用户注册成功"
else
    echo "❌ 用户注册失败"
    exit 1
fi

echo ""
echo "1.2 测试用户登录"
echo "----------------------------"

LOGIN_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${RANDOM_USER}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "响应: ${LOGIN_RESPONSE}"

if echo "${LOGIN_RESPONSE}" | grep -q '"access_token"'; then
    echo "✅ 用户登录成功"
    
    # 提取令牌
    ACCESS_TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "${LOGIN_RESPONSE}" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
    echo "User ID: ${USER_ID}"
else
    echo "❌ 用户登录失败"
    exit 1
fi

echo ""
echo "1.3 测试获取用户信息"
echo "----------------------------"

USER_INFO_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/user/profile" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

echo "响应: ${USER_INFO_RESPONSE}"

if echo "${USER_INFO_RESPONSE}" | grep -q '"username"'; then
    echo "✅ 获取用户信息成功"
else
    echo "❌ 获取用户信息失败"
    exit 1
fi

echo ""
echo "1.4 测试令牌刷新"
echo "----------------------------"

REFRESH_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"${REFRESH_TOKEN}\"
  }")

echo "响应: ${REFRESH_RESPONSE}"

if echo "${REFRESH_RESPONSE}" | grep -q '"access_token"'; then
    echo "✅ 令牌刷新成功"
    # 更新 access token
    ACCESS_TOKEN=$(echo "${REFRESH_RESPONSE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
else
    echo "❌ 令牌刷新失败"
    exit 1
fi

# ============================================
# Part 2: 对话服务测试
# ============================================

echo ""
echo "💬 Part 2: 对话服务测试"
echo "================================"
echo ""

echo "2.1 测试创建对话会话"
echo "----------------------------"

CREATE_SESSION_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/chat/sessions" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试会话",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个有帮助的助手"
  }')

echo "响应: ${CREATE_SESSION_RESPONSE}"

if echo "${CREATE_SESSION_RESPONSE}" | grep -q '"id"'; then
    echo "✅ 创建会话成功"
    
    # 提取会话ID (UUID格式)
    SESSION_ID=$(echo "${CREATE_SESSION_RESPONSE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "Session ID: ${SESSION_ID}"
else
    echo "❌ 创建会话失败"
    exit 1
fi

echo ""
echo "2.2 测试获取用户会话列表"
echo "----------------------------"

LIST_SESSIONS_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/chat/sessions" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

echo "响应: ${LIST_SESSIONS_RESPONSE}"

if echo "${LIST_SESSIONS_RESPONSE}" | grep -q "${SESSION_ID}"; then
    echo "✅ 获取会话列表成功"
else
    echo "❌ 获取会话列表失败"
    exit 1
fi

echo ""
echo "2.3 测试发送消息"
echo "----------------------------"

SEND_MESSAGE_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/chat/messages" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"${SESSION_ID}\",
    \"content\": \"你好，请介绍一下自己\",
    \"role\": \"user\"
  }")

echo "响应: ${SEND_MESSAGE_RESPONSE}"

if echo "${SEND_MESSAGE_RESPONSE}" | grep -q '"id"'; then
    echo "✅ 发送消息成功"
    
    MESSAGE_ID=$(echo "${SEND_MESSAGE_RESPONSE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "Message ID: ${MESSAGE_ID}"
else
    echo "❌ 发送消息失败"
    exit 1
fi

echo ""
echo "2.4 测试获取会话消息"
echo "----------------------------"

GET_MESSAGES_RESPONSE=$(curl -s -X GET "${GATEWAY_URL}/chat/sessions/${SESSION_ID}/messages" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

echo "响应: ${GET_MESSAGES_RESPONSE}"

if echo "${GET_MESSAGES_RESPONSE}" | grep -q "${MESSAGE_ID}"; then
    echo "✅ 获取消息列表成功"
else
    echo "❌ 获取消息列表失败"
    exit 1
fi

echo ""
echo "2.5 测试更新会话"
echo "----------------------------"

UPDATE_SESSION_RESPONSE=$(curl -s -X PUT "${GATEWAY_URL}/chat/sessions/${SESSION_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新后的测试会话",
    "pinned": true
  }')

echo "响应: ${UPDATE_SESSION_RESPONSE}"

if echo "${UPDATE_SESSION_RESPONSE}" | grep -q '"success":true'; then
    echo "✅ 更新会话成功"
else
    echo "❌ 更新会话失败"
    exit 1
fi

echo ""
echo "2.6 测试删除会话"
echo "----------------------------"

DELETE_SESSION_RESPONSE=$(curl -s -X DELETE "${GATEWAY_URL}/chat/sessions/${SESSION_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

echo "响应: ${DELETE_SESSION_RESPONSE}"

if echo "${DELETE_SESSION_RESPONSE}" | grep -q '"success":true'; then
    echo "✅ 删除会话成功"
else
    echo "❌ 删除会话失败"
    exit 1
fi

# ============================================
# 测试总结
# ============================================

echo ""
echo "================================"
echo "✅ 所有测试通过！"
echo "================================"
echo ""
echo "测试总结:"
echo "- ✅ 用户注册"
echo "- ✅ 用户登录"
echo "- ✅ 获取用户信息"
echo "- ✅ 令牌刷新"
echo "- ✅ 创建对话会话"
echo "- ✅ 获取会话列表"
echo "- ✅ 发送消息"
echo "- ✅ 获取消息列表"
echo "- ✅ 更新会话"
echo "- ✅ 删除会话"
echo ""

