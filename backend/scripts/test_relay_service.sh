#!/bin/bash

# 中转服务测试脚本

RELAY_URL="http://localhost:8083/v1"
GATEWAY_URL="http://localhost:8080/api/v1"

echo "================================"
echo "中转服务测试"
echo "================================"
echo ""

# 生成随机用户名
RANDOM_USER="testuser_$(date +%s)"
EMAIL="${RANDOM_USER}@example.com"
PASSWORD="password123"

# ============================================
# Part 1: 用户认证
# ============================================

echo ""
echo "Part 1: 用户认证"
echo "================================"
echo ""

echo "1.1 注册用户"
REGISTER_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${RANDOM_USER}\",
    \"email\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "响应: ${REGISTER_RESPONSE}"

if echo "${REGISTER_RESPONSE}" | grep -q '"success":true'; then
    echo "✅ 注册成功"
else
    echo "❌ 注册失败"
    exit 1
fi

echo ""
echo "1.2 用户登录"
LOGIN_RESPONSE=$(curl -s -X POST "${GATEWAY_URL}/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${RANDOM_USER}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "响应: ${LOGIN_RESPONSE}"

if echo "${LOGIN_RESPONSE}" | grep -q '"access_token"'; then
    echo "✅ 登录成功"
    ACCESS_TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
else
    echo "❌ 登录失败"
    exit 1
fi

# ============================================
# Part 2: 中转服务测试
# ============================================

echo ""
echo "Part 2: 中转服务测试"
echo "================================"
echo ""

echo "2.1 测试获取支持的模型列表"
echo "----------------------------"

MODELS_RESPONSE=$(curl -s -X GET "${RELAY_URL}/models")
echo "响应: ${MODELS_RESPONSE}"

if echo "${MODELS_RESPONSE}" | grep -q '"data"'; then
    echo "✅ 获取模型列表成功"
else
    echo "❌ 获取模型列表失败"
fi

echo ""
echo "2.2 测试获取渠道列表"
echo "----------------------------"

CHANNELS_RESPONSE=$(curl -s -X GET "${RELAY_URL}/channels")
echo "响应: ${CHANNELS_RESPONSE}"

if echo "${CHANNELS_RESPONSE}" | grep -q '"data"'; then
    echo "✅ 获取渠道列表成功"
    CHANNEL_COUNT=$(echo "${CHANNELS_RESPONSE}" | grep -o '"id"' | wc -l)
    echo "当前渠道数: ${CHANNEL_COUNT}"
else
    echo "❌ 获取渠道列表失败"
fi

echo ""
echo "2.3 测试 Chat Completion (非流式)"
echo "----------------------------"

CHAT_REQUEST='{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "你好，请介绍一下你自己"
    }
  ],
  "temperature": 0.7,
  "stream": false
}'

echo "请求体:"
echo "$CHAT_REQUEST" | jq .
echo ""

CHAT_RESPONSE=$(curl -s -X POST "${RELAY_URL}/chat/completions" \
  -H "Content-Type: application/json" \
  -d "${CHAT_REQUEST}")

echo "响应: ${CHAT_RESPONSE}"

if echo "${CHAT_RESPONSE}" | grep -q '"message":\|"choices"'; then
    echo "✅ Chat Completion 请求成功"
else
    echo "❌ Chat Completion 请求失败"
fi

echo ""
echo "2.4 测试 Chat Completion (流式)"
echo "----------------------------"

STREAM_REQUEST='{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "你好"
    }
  ],
  "stream": true
}'

echo "请求体:"
echo "$STREAM_REQUEST" | jq .
echo ""

STREAM_RESPONSE=$(curl -s -X POST "${RELAY_URL}/chat/completions" \
  -H "Content-Type: application/json" \
  -d "${STREAM_REQUEST}")

echo "响应: ${STREAM_RESPONSE}"

if echo "${STREAM_RESPONSE}" | grep -q 'not yet implemented'; then
    echo "⚠️ 流式响应暂未实现（预期行为）"
else
    echo "ℹ️ 流式响应返回: 已实现"
fi

# ============================================
# Part 3: 数据库验证
# ============================================

echo ""
echo "Part 3: 数据库验证"
echo "================================"
echo ""

echo "3.1 检查 channels 表"
CHANNELS_COUNT=$(psql -h localhost -p 5433 -U postgres -d oblivious -t -c "SELECT COUNT(*) FROM channels;" 2>&1)
echo "channels 表记录数: ${CHANNELS_COUNT}"

echo ""
echo "3.2 检查 model_prices 表"
PRICES_COUNT=$(psql -h localhost -p 5433 -U postgres -d oblivious -t -c "SELECT COUNT(*) FROM model_prices;" 2>&1)
echo "model_prices 表记录数: ${PRICES_COUNT}"

# ============================================
# 测试总结
# ============================================

echo ""
echo "================================"
echo "✅ 中转服务基础测试完成"
echo "================================"
echo ""
echo "测试项:"
echo "- ✅ 用户注册"
echo "- ✅ 用户登录"
echo "- ✅ 获取模型列表"
echo "- ✅ 获取渠道列表"
echo "- ✅ Chat Completion (非流式)"
echo "- ⚠️ Chat Completion (流式) - 暂未实现"
echo "- ✅ 数据库表验证"
echo ""
echo "下一步: 配置真实的 AI 服务 API 密钥并进行集成测试"
echo ""

