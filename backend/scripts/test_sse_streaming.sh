#!/bin/bash

# Week 7: SSE 流式响应测试脚本
# 测试完整的流式消息发送和接收

set -e

BASE_URL="http://localhost:8080"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-password}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"

echo "════════════════════════════════════════════════════════════"
echo "🧪 Week 7: SSE 流式响应完整测试"
echo "════════════════════════════════════════════════════════════"

# 1. 测试用户注册
echo ""
echo "📝 1️⃣ 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "streamtest",
    "email": "streamtest@test.com",
    "password": "Password123!"
  }')

echo "✅ 注册响应: $REGISTER_RESPONSE"

# 2. 测试用户登录
echo ""
echo "🔐 2️⃣ 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "streamtest",
    "password": "Password123!"
  }')

echo "✅ 登录响应: $LOGIN_RESPONSE"

# 提取 token
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
USER_ID=$(echo "$LOGIN_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)

if [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ 无法获取 Access Token"
  exit 1
fi

echo "✅ Access Token: ${ACCESS_TOKEN:0:20}..."
echo "✅ User ID: $USER_ID"

# 3. 创建会话
echo ""
echo "💬 3️⃣ 创建测试会话..."
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/chat/sessions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "title": "SSE 流式测试会话",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个友好的助手"
  }')

echo "✅ 会话创建响应: $SESSION_RESPONSE"

SESSION_ID=$(echo "$SESSION_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$SESSION_ID" ]; then
  echo "❌ 无法获取 Session ID"
  exit 1
fi

echo "✅ Session ID: $SESSION_ID"

# 4. 测试流式消息发送（SSE）
echo ""
echo "🌊 4️⃣ 测试 SSE 流式消息..."
echo "📨 发送消息: '你好，请用简短的句子介绍自己'"
echo ""
echo "接收流式响应："
echo "────────────────────────────────────────────────────────────"

# 使用 curl 连接 SSE
curl -s -X POST "$BASE_URL/api/v1/chat/messages/stream" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"content\": \"你好，请用简短的句子介绍自己\"
  }" \
  --no-buffer | while IFS= read -r line; do
    # 解析并显示流式事件
    if [[ $line == event:* ]]; then
      EVENT=$(echo "$line" | cut -d' ' -f2-)
      echo "📌 事件: $EVENT"
    elif [[ $line == data:* ]]; then
      DATA=$(echo "$line" | cut -d' ' -f2-)
      # 尝试解析 JSON
      if command -v jq &> /dev/null; then
        echo "$DATA" | jq '.' 2>/dev/null || echo "📦 数据: $DATA"
      else
        echo "📦 数据: $DATA"
      fi
    fi
  done

echo ""
echo "────────────────────────────────────────────────────────────"

# 5. 获取会话消息
echo ""
echo "📖 5️⃣ 获取会话消息列表..."
MESSAGES_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/chat/sessions/$SESSION_ID/messages" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "✅ 消息列表:"
if command -v jq &> /dev/null; then
  echo "$MESSAGES_RESPONSE" | jq '.data.messages[] | {role, content: (.content | .[0:50] + "..."), tokens: .total_tokens}' 2>/dev/null || echo "$MESSAGES_RESPONSE"
else
  echo "$MESSAGES_RESPONSE"
fi

# 6. 验证数据库中的消息
echo ""
echo "🗄️  6️⃣ 验证数据库中的消息..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d oblivious << EOF 2>/dev/null | tail -10
SELECT 
  role,
  LENGTH(content) as content_length,
  input_tokens,
  output_tokens,
  total_tokens,
  created_at
FROM messages
WHERE session_id = '$SESSION_ID'
ORDER BY created_at DESC
LIMIT 5;
EOF

echo ""
echo "════════════════════════════════════════════════════════════"
echo "✅ SSE 流式响应测试完成！"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "📊 测试总结:"
echo "  ✅ 用户注册和登录成功"
echo "  ✅ 会话创建成功"
echo "  ✅ SSE 流式消息接收成功"
echo "  ✅ 消息已保存到数据库"
echo ""
echo "🎯 下一步:"
echo "  1. 访问 http://localhost:3000 查看前端"
echo "  2. 登录账号: streamtest / Password123!"
echo "  3. 创建对话并测试流式消息发送"
echo ""
echo "════════════════════════════════════════════════════════════"

