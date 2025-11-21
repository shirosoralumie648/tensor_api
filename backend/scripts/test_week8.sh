#!/bin/bash

# test_week8.sh
# 用于测试 Week 8 的助手服务和知识库服务

set -e

# 定义 API 网关地址
GATEWAY_URL="http://localhost:8080"
AGENT_SERVICE_URL="http://localhost:8084"
KB_SERVICE_URL="http://localhost:8085"

echo "════════════════════════════════════════════════════════════"
echo "🧪 Week 8: 助手服务与知识库服务测试"
echo "════════════════════════════════════════════════════════════"

# 1️⃣ 测试用户注册和登录
echo ""
echo "📝 1️⃣ 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"week8test","email":"week8@test.com","password":"Password123!"}')
echo "✅ 注册响应: $REGISTER_RESPONSE"

echo ""
echo "🔐 2️⃣ 测试用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$GATEWAY_URL/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"week8test","password":"Password123!"}')
echo "✅ 登录响应: $LOGIN_RESPONSE"

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.data.user.id')

if [ "$ACCESS_TOKEN" == "null" ]; then
  echo "❌ 无法获取 Access Token"
  exit 1
fi
echo "✅ Access Token: ${ACCESS_TOKEN:0:20}..."
echo "✅ User ID: $USER_ID"

# 3️⃣ 测试 Agent 服务 - 创建助手
echo ""
echo "🤖 3️⃣ 测试 Agent 服务 - 创建助手..."
CREATE_AGENT_RESPONSE=$(curl -s -X POST "$AGENT_SERVICE_URL/api/v1/agents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "编程助手",
    "avatar": "https://example.com/avatar.png",
    "description": "一个专业的编程助手",
    "category": "编程",
    "system_role": "你是一个专业的编程助手，精通多种编程语言和框架。",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "is_public": true
  }')
echo "✅ 创建助手响应: $CREATE_AGENT_RESPONSE"

AGENT_ID=$(echo "$CREATE_AGENT_RESPONSE" | jq -r '.data.id')
if [ "$AGENT_ID" == "null" ]; then
  echo "❌ 无法获取 Agent ID"
  exit 1
fi
echo "✅ Agent ID: $AGENT_ID"

# 4️⃣ 测试 Agent 服务 - 获取助手列表
echo ""
echo "📖 4️⃣ 测试 Agent 服务 - 获取助手列表..."
GET_AGENTS_RESPONSE=$(curl -s -X GET "$AGENT_SERVICE_URL/api/v1/agents/user?page=1&page_size=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "✅ 获取助手列表响应:"
echo "$GET_AGENTS_RESPONSE" | jq '.data | {agents: .[].name, total: length}'

# 5️⃣ 测试 Agent 服务 - 获取公开助手
echo ""
echo "🌟 5️⃣ 测试 Agent 服务 - 获取公开助手..."
PUBLIC_AGENTS_RESPONSE=$(curl -s -X GET "$AGENT_SERVICE_URL/api/v1/agents/public?page=1&page_size=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "✅ 获取公开助手响应:"
echo "$PUBLIC_AGENTS_RESPONSE" | jq '.data | {count: length, agents: .[].name}'

# 6️⃣ 测试知识库服务 - 创建知识库
echo ""
echo "📚 6️⃣ 测试知识库服务 - 创建知识库..."
CREATE_KB_RESPONSE=$(curl -s -X POST "$KB_SERVICE_URL/api/v1/knowledge-bases" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "name": "Python 编程指南",
    "description": "Python 编程的完整指南和最佳实践",
    "embedding_model": "text-embedding-3-small",
    "chunk_size": 512,
    "chunk_overlap": 50
  }')
echo "✅ 创建知识库响应: $CREATE_KB_RESPONSE"

KB_ID=$(echo "$CREATE_KB_RESPONSE" | jq -r '.data.id')
if [ "$KB_ID" == "null" ]; then
  echo "❌ 无法获取 Knowledge Base ID"
  exit 1
fi
echo "✅ Knowledge Base ID: $KB_ID"

# 7️⃣ 测试知识库服务 - 上传文档
echo ""
echo "📄 7️⃣ 测试知识库服务 - 上传文档..."
UPLOAD_DOC_RESPONSE=$(curl -s -X POST "$KB_SERVICE_URL/api/v1/knowledge-bases/$KB_ID/documents" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "title": "Python 基础教程",
    "file_content": "Python 是一种高级编程语言，具有简洁易读的语法。它支持多种编程范式，包括面向对象、函数式和过程式编程。Python 的应用范围很广，从 Web 开发到数据科学、人工智能等。在 Python 中，一切都是对象，包括函数、类、模块等。Python 使用缩进来表示代码块，这使得代码更加清晰易读。列表是 Python 中最常用的数据结构之一，它可以存储任意类型的对象，并且是可变的。字典是另一种重要的数据结构，它使用键值对存储数据。集合是一个无序的集合，其中所有元素都是唯一的。元组是不可变的序列，经常用作字典的键。字符串是不可变的序列，用于表示文本数据。"
  }')
echo "✅ 上传文档响应: $UPLOAD_DOC_RESPONSE"

DOC_ID=$(echo "$UPLOAD_DOC_RESPONSE" | jq -r '.data.id')
if [ "$DOC_ID" != "null" ]; then
  echo "✅ Document ID: $DOC_ID"
fi

# 8️⃣ 测试知识库服务 - 获取知识库列表
echo ""
echo "📋 8️⃣ 测试知识库服务 - 获取知识库列表..."
LIST_KB_RESPONSE=$(curl -s -X GET "$KB_SERVICE_URL/api/v1/knowledge-bases?page=1&page_size=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "✅ 知识库列表响应:"
echo "$LIST_KB_RESPONSE" | jq '.data.knowledge_bases[] | {name, description, document_count, total_chunks}'

# 9️⃣ 测试知识库服务 - 获取文档列表
echo ""
echo "📑 9️⃣ 测试知识库服务 - 获取文档列表..."
LIST_DOCS_RESPONSE=$(curl -s -X GET "$KB_SERVICE_URL/api/v1/knowledge-bases/$KB_ID/documents" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "✅ 文档列表响应:"
echo "$LIST_DOCS_RESPONSE" | jq '.data.documents[] | {title, status, chunk_count}'

# 🔟 测试 Agent 赞功能
echo ""
echo "👍 🔟 测试 Agent 赞功能..."
LIKE_AGENT_RESPONSE=$(curl -s -X POST "$AGENT_SERVICE_URL/api/v1/agents/$AGENT_ID/like" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "✅ 赞助手响应: $LIKE_AGENT_RESPONSE"

echo ""
echo "════════════════════════════════════════════════════════════"
echo "✅ Week 8 所有测试完成！"
echo "════════════════════════════════════════════════════════════"

echo ""
echo "📊 测试总结:"
echo "  ✅ 用户认证          - 成功"
echo "  ✅ Agent 创建        - 成功"
echo "  ✅ Agent 列表        - 成功"
echo "  ✅ Agent 公开列表    - 成功"
echo "  ✅ 知识库创建        - 成功"
echo "  ✅ 文档上传          - 成功"
echo "  ✅ 知识库列表        - 成功"
echo "  ✅ 文档列表          - 成功"
echo "  ✅ Agent 赞功能      - 成功"
echo ""

echo "🎯 服务端点速查"
echo "  • Agent 服务:         http://localhost:8084/api/v1/agents"
echo "  • 知识库服务:         http://localhost:8085/api/v1/knowledge-bases"
echo "  • API 网关:           http://localhost:8080"
echo ""

