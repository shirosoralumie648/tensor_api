#!/bin/bash

# 完整的端到端测试脚本

set -e

GATEWAY_URL="http://localhost:8080/api/v1"
RELAY_URL="http://localhost:8083/v1"

echo "════════════════════════════════════════════════════════════"
echo "Oblivious 项目完整流程测试"
echo "════════════════════════════════════════════════════════════"
echo ""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数
TESTS_PASSED=0
TESTS_FAILED=0

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=$4
    local token=$5
    
    local headers="-H 'Content-Type: application/json'"
    if [ -n "$token" ]; then
        headers="$headers -H 'Authorization: Bearer $token'"
    fi
    
    local response
    if [ "$method" = "GET" ]; then
        response=$(eval "curl -s -w '\n%{http_code}' -X $method '$GATEWAY_URL$endpoint' $headers")
    else
        response=$(eval "curl -s -w '\n%{http_code}' -X $method '$GATEWAY_URL$endpoint' $headers -d '$data'")
    fi
    
    local body=$(echo "$response" | head -n -1)
    local http_code=$(echo "$response" | tail -n 1)
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ PASS${NC} [$method] $endpoint (HTTP $http_code)"
        ((TESTS_PASSED++))
        echo "$body"
    else
        echo -e "${RED}❌ FAIL${NC} [$method] $endpoint (Expected: $expected_code, Got: $http_code)"
        echo "Response: $body"
        ((TESTS_FAILED++))
    fi
}

echo ""
echo "📋 Part 1: 用户认证测试"
echo "════════════════════════════════════════════"
echo ""

# 生成唯一用户名
RANDOM_USER="testuser_$(date +%s)"
EMAIL="${RANDOM_USER}@test.com"
PASSWORD="TestPass123!"

echo "1️⃣  测试用户注册"
REGISTER_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$RANDOM_USER\",
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}✅ 用户注册成功${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 用户注册失败${NC}"
    echo "Response: $REGISTER_RESPONSE"
    ((TESTS_FAILED++))
fi

echo ""
echo "2️⃣  测试用户登录"
LOGIN_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$RANDOM_USER\",
    \"password\": \"$PASSWORD\"
  }")

if echo "$LOGIN_RESPONSE" | grep -q '"access_token"'; then
    echo -e "${GREEN}✅ 用户登录成功${NC}"
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    echo "Token: ${ACCESS_TOKEN:0:30}..."
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 用户登录失败${NC}"
    echo "Response: $LOGIN_RESPONSE"
    ((TESTS_FAILED++))
    exit 1
fi

echo ""
echo "3️⃣  测试获取用户信息"
USER_PROFILE=$(curl -s -X GET "$GATEWAY_URL/user/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$USER_PROFILE" | grep -q "$RANDOM_USER"; then
    echo -e "${GREEN}✅ 获取用户信息成功${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 获取用户信息失败${NC}"
    ((TESTS_FAILED++))
fi

echo ""
echo "📋 Part 2: 对话功能测试"
echo "════════════════════════════════════════════"
echo ""

echo "1️⃣  测试创建对话会话"
CREATE_SESSION=$(curl -s -X POST "$GATEWAY_URL/chat/sessions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试会话",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个有帮助的助手",
    "context_length": 4
  }')

if echo "$CREATE_SESSION" | grep -q '"id"'; then
    echo -e "${GREEN}✅ 创建会话成功${NC}"
    SESSION_ID=$(echo "$CREATE_SESSION" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | head -1)
    echo "Session ID: ${SESSION_ID:0:30}..."
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 创建会话失败${NC}"
    echo "Response: $CREATE_SESSION"
    ((TESTS_FAILED++))
    exit 1
fi

echo ""
echo "2️⃣  测试获取会话列表"
SESSIONS_LIST=$(curl -s -X GET "$GATEWAY_URL/chat/sessions" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$SESSIONS_LIST" | grep -q '"id"'; then
    echo -e "${GREEN}✅ 获取会话列表成功${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 获取会话列表失败${NC}"
    ((TESTS_FAILED++))
fi

echo ""
echo "3️⃣  测试发送消息"
SEND_MESSAGE=$(curl -s -X POST "$GATEWAY_URL/chat/messages" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"content\": \"你好，请介绍一下你自己\"
  }")

if echo "$SEND_MESSAGE" | grep -q '"content"'; then
    echo -e "${GREEN}✅ 发送消息成功${NC}"
    MESSAGE_ID=$(echo "$SEND_MESSAGE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | head -1)
    echo "Message ID: ${MESSAGE_ID:0:30}..."
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 发送消息失败${NC}"
    echo "Response: $SEND_MESSAGE"
    ((TESTS_FAILED++))
fi

echo ""
echo "4️⃣  测试获取消息列表"
MESSAGES=$(curl -s -X GET "$GATEWAY_URL/chat/sessions/$SESSION_ID/messages" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$MESSAGES" | grep -q '"role"'; then
    echo -e "${GREEN}✅ 获取消息列表成功${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 获取消息列表失败${NC}"
    ((TESTS_FAILED++))
fi

echo ""
echo "📋 Part 3: 中转服务测试"
echo "════════════════════════════════════════════"
echo ""

echo "1️⃣  测试获取支持的模型"
MODELS=$(curl -s -X GET "$RELAY_URL/models")
if echo "$MODELS" | grep -q '"data"'; then
    echo -e "${GREEN}✅ 获取模型列表成功${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 获取模型列表失败${NC}"
    ((TESTS_FAILED++))
fi

echo ""
echo "2️⃣  测试获取渠道列表"
CHANNELS=$(curl -s -X GET "$RELAY_URL/channels")
if echo "$CHANNELS" | grep -q '"data"'; then
    echo -e "${GREEN}✅ 获取渠道列表成功${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}❌ 获取渠道列表失败${NC}"
    ((TESTS_FAILED++))
fi

echo ""
echo "════════════════════════════════════════════════════════════"
echo "📊 测试结果总结"
echo "════════════════════════════════════════════════════════════"
echo ""
echo -e "通过测试: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败测试: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 有 $TESTS_FAILED 个测试失败${NC}"
    exit 1
fi

