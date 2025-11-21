# API 参考文档

## 基本信息

- **Base URL**: `https://api.oblivious.ai/api/v1`
- **认证方式**: JWT Bearer Token
- **内容类型**: `application/json`
- **字符编码**: `UTF-8`

## 通用响应格式

### 成功响应

```json
{
  "data": {...},
  "message": "Success"
}
```

### 错误响应

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "错误描述",
    "details": {}
  },
  "request_id": "req_abc123"
}
```

### HTTP 状态码

| 状态码 | 说明 |
|-------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（Token 无效或过期）|
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 429 | 请求频率超限 |
| 500 | 服务器内部错误 |
| 502 | 网关错误 |
| 503 | 服务不可用 |

## 认证接口

### 用户注册

```http
POST /auth/register
```

**请求体**：

```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecureP@ss123",
  "invite_code": "ABC123"
}
```

**响应**：

```json
{
  "data": {
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "display_name": "John Doe",
      "avatar_url": null,
      "role": 1,
      "quota": 100000,
      "created_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 用户登录

```http
POST /auth/login
```

**请求体**：

```json
{
  "email": "john@example.com",
  "password": "SecureP@ss123"
}
```

**响应**：

```json
{
  "data": {
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "quota": 95000
    },
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2024-01-02T00:00:00Z"
  }
}
```

### 刷新 Token

```http
POST /auth/refresh
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2024-01-03T00:00:00Z"
  }
}
```

## 用户接口

### 获取用户信息

```http
GET /user/profile
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "display_name": "John Doe",
    "avatar_url": "https://cdn.example.com/avatar.jpg",
    "role": 1,
    "quota": 95000,
    "total_quota": 100000,
    "used_quota": 5000,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 更新用户信息

```http
PUT /user/profile
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "display_name": "John Smith",
  "avatar_url": "https://cdn.example.com/new-avatar.jpg"
}
```

### 修改密码

```http
PUT /user/password
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "old_password": "OldPass123",
  "new_password": "NewSecureP@ss456"
}
```

## 对话接口

### 创建对话

```http
POST /chat/completions
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "session_id": "uuid-optional",
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "你是一个有帮助的AI助手"
    },
    {
      "role": "user",
      "content": "你好，请介绍一下你自己"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 2000,
  "stream": true
}
```

**流式响应**（SSE）：

```
data: {"id":"msg_123","choices":[{"delta":{"content":"你"}}]}

data: {"id":"msg_123","choices":[{"delta":{"content":"好"}}]}

data: {"id":"msg_123","choices":[{"delta":{"content":"！"}}]}

data: {"id":"msg_123","choices":[{"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

**非流式响应**：

```json
{
  "data": {
    "id": "msg_123",
    "model": "gpt-3.5-turbo",
    "choices": [
      {
        "message": {
          "role": "assistant",
          "content": "你好！我是一个AI助手..."
        },
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 25,
      "completion_tokens": 150,
      "total_tokens": 175
    },
    "cost": 35
  }
}
```

### 获取会话列表

```http
GET /chat/sessions?page=1&limit=20&archived=false
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "sessions": [
      {
        "id": "uuid-123",
        "title": "关于 AI 的讨论",
        "model": "gpt-3.5-turbo",
        "message_count": 12,
        "pinned": false,
        "archived": false,
        "created_at": "2024-01-01T10:00:00Z",
        "updated_at": "2024-01-01T12:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45
    }
  }
}
```

### 获取会话详情

```http
GET /chat/sessions/{session_id}
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "id": "uuid-123",
    "title": "关于 AI 的讨论",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个专业的AI助手",
    "messages": [
      {
        "id": "msg_1",
        "role": "user",
        "content": "什么是人工智能？",
        "created_at": "2024-01-01T10:00:00Z"
      },
      {
        "id": "msg_2",
        "role": "assistant",
        "content": "人工智能是...",
        "model": "gpt-3.5-turbo",
        "input_tokens": 10,
        "output_tokens": 50,
        "created_at": "2024-01-01T10:00:05Z"
      }
    ],
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T12:30:00Z"
  }
}
```

### 更新会话

```http
PUT /chat/sessions/{session_id}
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "title": "新的会话标题",
  "pinned": true,
  "archived": false
}
```

### 删除会话

```http
DELETE /chat/sessions/{session_id}
Authorization: Bearer {token}
```

## 模型接口

### 获取可用模型列表

```http
GET /models
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "models": [
      {
        "id": "gpt-3.5-turbo",
        "name": "GPT-3.5 Turbo",
        "provider": "OpenAI",
        "context_length": 4096,
        "input_price": 0.0015,
        "output_price": 0.002,
        "available": true
      },
      {
        "id": "gpt-4",
        "name": "GPT-4",
        "provider": "OpenAI",
        "context_length": 8192,
        "input_price": 0.03,
        "output_price": 0.06,
        "available": true
      }
    ]
  }
}
```

## 计费接口

### 查询额度

```http
GET /billing/quota
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "quota": 95000,
    "total_quota": 100000,
    "used_quota": 5000,
    "quota_usd": 9.50
  }
}
```

### 查询消费记录

```http
GET /billing/logs?start_date=2024-01-01&end_date=2024-01-31&page=1&limit=20
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "logs": [
      {
        "id": 1,
        "model": "gpt-3.5-turbo",
        "input_tokens": 25,
        "output_tokens": 150,
        "total_tokens": 175,
        "cost": 35,
        "cost_usd": 0.035,
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 120
    },
    "summary": {
      "total_cost": 5000,
      "total_cost_usd": 5.0,
      "total_tokens": 25000
    }
  }
}
```

### 充值

```http
POST /billing/recharge
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "amount": 100,
  "payment_method": "alipay",
  "return_url": "https://example.com/return"
}
```

**响应**：

```json
{
  "data": {
    "order_id": "order_123",
    "amount": 100,
    "payment_url": "https://payment.example.com/pay/order_123",
    "expires_at": "2024-01-01T01:00:00Z"
  }
}
```

## 知识库接口

### 创建知识库

```http
POST /knowledge/bases
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "name": "技术文档库",
  "description": "存储技术文档和API说明",
  "embedding_model": "text-embedding-3-small",
  "chunk_size": 512,
  "chunk_overlap": 50
}
```

**响应**：

```json
{
  "data": {
    "id": 1,
    "name": "技术文档库",
    "description": "存储技术文档和API说明",
    "embedding_model": "text-embedding-3-small",
    "document_count": 0,
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

### 上传文档

```http
POST /knowledge/bases/{kb_id}/documents
Authorization: Bearer {token}
Content-Type: multipart/form-data
```

**表单字段**：

```
file: [PDF/Word/TXT文件]
title: "API 设计指南"
```

**响应**：

```json
{
  "data": {
    "id": "uuid-doc-123",
    "title": "API 设计指南",
    "file_type": "pdf",
    "file_size": 1024000,
    "status": 1,
    "created_at": "2024-01-01T10:05:00Z"
  }
}
```

### 检索知识库

```http
POST /knowledge/bases/{kb_id}/search
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "query": "如何设计 RESTful API",
  "top_k": 5
}
```

**响应**：

```json
{
  "data": {
    "results": [
      {
        "chunk_id": "uuid-chunk-1",
        "content": "RESTful API 设计原则包括...",
        "similarity": 0.92,
        "metadata": {
          "document_title": "API 设计指南",
          "page": 5
        }
      }
    ]
  }
}
```

## 助手接口

### 获取助手市场列表

```http
GET /agents?category=coding&page=1&limit=20
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "agents": [
      {
        "id": 1,
        "name": "代码助手",
        "description": "帮助你编写和调试代码",
        "avatar_url": "https://cdn.example.com/agent1.jpg",
        "category": "coding",
        "tags": ["编程", "调试", "代码审查"],
        "usage_count": 1250,
        "like_count": 89,
        "is_official": true
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45
    }
  }
}
```

### 获取助手详情

```http
GET /agents/{agent_id}
Authorization: Bearer {token}
```

**响应**：

```json
{
  "data": {
    "id": 1,
    "name": "代码助手",
    "description": "帮助你编写和调试代码",
    "avatar_url": "https://cdn.example.com/agent1.jpg",
    "system_prompt": "你是一个专业的编程助手...",
    "welcome_message": "你好！我可以帮你解决编程问题。",
    "suggested_messages": [
      "如何实现二分查找？",
      "解释一下闭包的概念"
    ],
    "default_model": "gpt-4",
    "temperature": 0.5,
    "is_official": true,
    "usage_count": 1250,
    "like_count": 89
  }
}
```

### 创建自定义助手

```http
POST /agents
Authorization: Bearer {token}
```

**请求体**：

```json
{
  "name": "我的翻译助手",
  "description": "中英文互译助手",
  "system_prompt": "你是一个专业的翻译助手，擅长中英文互译...",
  "welcome_message": "你好！请输入需要翻译的内容。",
  "default_model": "gpt-3.5-turbo",
  "temperature": 0.3,
  "is_public": false
}
```

## 文件接口

### 上传文件

```http
POST /files/upload
Authorization: Bearer {token}
Content-Type: multipart/form-data
```

**表单字段**：

```
file: [文件]
purpose: "chat" | "knowledge" | "avatar"
```

**响应**：

```json
{
  "data": {
    "file_id": "file_123",
    "file_name": "document.pdf",
    "file_size": 1024000,
    "file_type": "pdf",
    "url": "https://cdn.example.com/files/file_123.pdf",
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

## 速率限制

### 默认限制

| 用户类型 | 请求限制 |
|---------|---------|
| 免费用户 | 20 请求/分钟 |
| 付费用户 | 100 请求/分钟 |
| VIP 用户 | 500 请求/分钟 |

### 响应头

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### 超限响应

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求频率超限，请稍后重试",
    "details": {
      "retry_after": 30
    }
  }
}
```

## WebSocket 接口

### 连接

```
wss://api.oblivious.ai/ws?token={jwt_token}
```

### 消息格式

**客户端发送**：

```json
{
  "type": "chat",
  "data": {
    "session_id": "uuid-123",
    "message": "你好"
  }
}
```

**服务端推送**：

```json
{
  "type": "message",
  "data": {
    "role": "assistant",
    "content": "你好！",
    "delta": true
  }
}
```

## SDK 示例

### JavaScript/TypeScript

```typescript
import { ObliviousClient } from '@oblivious/sdk'

const client = new ObliviousClient({
  apiKey: 'your-api-key',
  baseURL: 'https://api.oblivious.ai/api/v1'
})

// 创建对话
const response = await client.chat.create({
  model: 'gpt-3.5-turbo',
  messages: [
    { role: 'user', content: '你好' }
  ]
})
```

### Python

```python
from oblivious import Client

client = Client(api_key='your-api-key')

# 创建对话
response = client.chat.create(
    model='gpt-3.5-turbo',
    messages=[
        {'role': 'user', 'content': '你好'}
    ]
)
```

## 错误码参考

| 错误码 | 说明 | HTTP 状态 |
|-------|------|----------|
| INVALID_TOKEN | Token 无效或过期 | 401 |
| INSUFFICIENT_QUOTA | 额度不足 | 403 |
| RATE_LIMIT_EXCEEDED | 请求频率超限 | 429 |
| MODEL_NOT_FOUND | 模型不存在 | 404 |
| INVALID_PARAMETERS | 参数错误 | 400 |
| SERVICE_UNAVAILABLE | 服务不可用 | 503 |
| INTERNAL_ERROR | 内部服务器错误 | 500 |

## 相关文档

- [架构设计](ARCHITECTURE.md)
- [快速开始](QUICK_START.md)
- [API 网关设计](API_GATEWAY_DESIGN.md)
