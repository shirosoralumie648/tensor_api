#!/bin/bash

# 测试用户服务的注册和登录功能

BASE_URL="http://localhost:8081/api/v1"

echo "================================"
echo "测试用户服务"
echo "================================"
echo ""

# 生成随机用户名
RANDOM_USER="testuser_$(date +%s)"
EMAIL="${RANDOM_USER}@example.com"
PASSWORD="password123"

echo "1. 测试用户注册"
echo "----------------------------"
echo "注册用户: ${RANDOM_USER}"
echo ""

REGISTER_RESPONSE=$(curl -s -X POST "${BASE_URL}/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${RANDOM_USER}\",
    \"email\": \"${EMAIL}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "响应: ${REGISTER_RESPONSE}"
echo ""

# 检查是否注册成功
if echo "${REGISTER_RESPONSE}" | grep -q '"success":true'; then
    echo "✅ 注册成功"
else
    echo "❌ 注册失败"
    exit 1
fi

echo ""
echo "2. 测试用户登录"
echo "----------------------------"
echo ""

LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${RANDOM_USER}\",
    \"password\": \"${PASSWORD}\"
  }")

echo "响应: ${LOGIN_RESPONSE}"
echo ""

# 检查是否登录成功
if echo "${LOGIN_RESPONSE}" | grep -q '"access_token"'; then
    echo "✅ 登录成功"
    
    # 提取 access_token
    ACCESS_TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    
    # 提取用户ID
    USER_ID=$(echo "${LOGIN_RESPONSE}" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "User ID: ${USER_ID}"
else
    echo "❌ 登录失败"
    exit 1
fi

echo ""
echo "3. 测试获取用户信息"
echo "----------------------------"
echo ""

USER_INFO_RESPONSE=$(curl -s -X GET "${BASE_URL}/user/${USER_ID}" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

echo "响应: ${USER_INFO_RESPONSE}"
echo ""

if echo "${USER_INFO_RESPONSE}" | grep -q '"username"'; then
    echo "✅ 获取用户信息成功"
else
    echo "❌ 获取用户信息失败"
    exit 1
fi

echo ""
echo "================================"
echo "✅ 所有测试通过！"
echo "================================"

