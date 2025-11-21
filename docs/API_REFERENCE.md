# Oblivious API 参考文档

## 概述

Oblivious 提供 RESTful API 接口，所有接口均通过 API 网关统一访问。

**基础 URL**: `https://api.oblivious.ai/api/v1` (生产环境)  
**基础 URL**: `http://localhost:8080/api/v1` (本地开发)

---

## 认证

### JWT Token 认证

除公开接口外，所有接口都需要在请求头中携带 JWT Token：

```http
Authorization: Bearer {your_access_token}
```

### Token 刷新

Access Token 有效期为 2 小时，可使用 Refresh Token 获取新的 Access Token。

---

## 响应格式

### 成功响应

```json
{
  "success": true,
  "data": {
    // 响应数据
  },
  "message": "操作成功",
  "timestamp": "2024-11-21T10:00:00Z"
}
```

### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "Token 已过期",
    "details": null
  },
  "timestamp": "2024-11-21T10:00:00Z"
}
```

### 错误码列表

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `INVALID_REQUEST` | 400 | 请求参数错误 |
| `UNAUTHORIZED` | 401 | 未登录或 Token 无效 |
| `FORBIDDEN` | 403 | 无权限访问 |
| `NOT_FOUND` | 404 | 资源不存在 |
| `INTERNAL_ERROR` | 500 | 服务器内部错误 |
| `SERVICE_UNAVAILABLE` | 503 | 服务暂时不可用 |

---

## 用户认证相关

### 用户注册

**端点**: `POST /register`  
**认证**: 不需要

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "display_name": null,
      "avatar_url": null,
      "role": 1,
      "quota": 0,
      "created_at": "2024-11-21T10:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200
  },
  "message": "注册成功"
}
```

### 用户登录

**端点**: `POST /login`  
**认证**: 不需要

**请求体**:
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应**: 同注册接口

### 刷新 Token

**端点**: `POST /refresh`  
**认证**: 不需要

**请求体**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200
  }
}
```

---

## 用户管理相关

### 获取用户信息

**端点**: `GET /user/profile`  
**认证**: 需要

**响应**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "display_name": "测试用户",
    "avatar_url": "https://cdn.example.com/avatar.jpg",
    "role": 1,
    "quota": 5000,
    "created_at": "2024-11-21T10:00:00Z"
  }
}
```

### 更新用户信息

**端点**: `PUT /user/profile`  
**认证**: 需要

**请求体**:
```json
{
  "display_name": "新昵称",
  "avatar_url": "https://cdn.example.com/new-avatar.jpg"
}
```

**响应**: 同获取用户信息

---

## 对话会话相关

### 创建会话

**端点**: `POST /chat/sessions`  
**认证**: 需要

**请求体**:
```json
{
  "title": "我的第一个对话",
  "model": "gpt-3.5-turbo",
  "temperature": 0.7,
  "system_role": "你是一个有帮助的助手"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": 1,
    "title": "我的第一个对话",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个有帮助的助手",
    "created_at": "2024-11-21T10:00:00Z",
    "updated_at": "2024-11-21T10:00:00Z"
  },
  "message": "会话创建成功"
}
```

### 获取会话列表

**端点**: `GET /chat/sessions`  
**认证**: 需要

**查询参数**:
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 20

**响应**:
```json
{
  "success": true,
  "data": {
    "sessions": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "我的第一个对话",
        "model": "gpt-3.5-turbo",
        "created_at": "2024-11-21T10:00:00Z",
        "updated_at": "2024-11-21T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "pageSize": 20
  }
}
```

### 获取会话详情

**端点**: `GET /chat/sessions/:id`  
**认证**: 需要

**响应**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": 1,
    "title": "我的第一个对话",
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "system_role": "你是一个有帮助的助手",
    "created_at": "2024-11-21T10:00:00Z",
    "updated_at": "2024-11-21T10:00:00Z"
  }
}
```

### 更新会话

**端点**: `PUT /chat/sessions/:id`  
**认证**: 需要

**请求体**:
```json
{
  "title": "新标题"
}
```

**响应**: 同获取会话详情

### 删除会话

**端点**: `DELETE /chat/sessions/:id`  
**认证**: 需要

**响应**:
```json
{
  "success": true,
  "message": "会话删除成功"
}
```

---

## 消息相关

### 获取会话消息历史

**端点**: `GET /chat/sessions/:id/messages`  
**认证**: 需要

**查询参数**:
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 50

**响应**:
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "session_id": "550e8400-e29b-41d4-a716-446655440000",
        "role": "user",
        "content": "你好",
        "created_at": "2024-11-21T10:00:00Z"
      },
      {
        "id": "660e8400-e29b-41d4-a716-446655440002",
        "session_id": "550e8400-e29b-41d4-a716-446655440000",
        "role": "assistant",
        "content": "你好！有什么我可以帮助你的吗？",
        "model": "gpt-3.5-turbo",
        "input_tokens": 10,
        "output_tokens": 20,
        "total_tokens": 30,
        "cost": 15,
        "created_at": "2024-11-21T10:00:05Z"
      }
    ],
    "total": 2,
    "page": 1,
    "pageSize": 50
  }
}
```

### 发送消息（非流式）

**端点**: `POST /chat/messages`  
**认证**: 需要

**请求体**:
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "你好，请介绍一下自己"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "user_message": {
      "id": "660e8400-e29b-41d4-a716-446655440003",
      "session_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "user",
      "content": "你好，请介绍一下自己",
      "created_at": "2024-11-21T10:10:00Z"
    },
    "assistant_message": {
      "id": "660e8400-e29b-41d4-a716-446655440004",
      "session_id": "550e8400-e29b-41d4-a716-446655440000",
      "role": "assistant",
      "content": "你好！我是一个 AI 助手...",
      "model": "gpt-3.5-turbo",
      "input_tokens": 25,
      "output_tokens": 50,
      "total_tokens": 75,
      "cost": 38,
      "created_at": "2024-11-21T10:10:05Z"
    }
  },
  "message": "消息发送成功"
}
```

### 发送消息（流式）

**端点**: `POST /chat/messages/stream`  
**认证**: 需要

**请求体**: 同非流式接口

**响应**: 使用 Server-Sent Events (SSE) 格式

```
data: {"delta":"你"}
data: {"delta":"好"}
data: {"delta":"！"}
data: {"delta":"我"}
data: {"delta":"是"}
...
data: [DONE]
```

---

## 限流规则

| 接口类型 | 限流规则 |
|---------|---------|
| 公开接口（注册、登录） | 10 请求/分钟 |
| 受保护接口（需认证） | 100 请求/分钟 |
| 消息发送接口 | 20 请求/分钟 |

当触发限流时，API 会返回 `429 Too Many Requests` 状态码。

---

## 分页

所有列表接口都支持分页，使用以下查询参数：

- `page`: 页码（从 1 开始）
- `page_size`: 每页数量（默认 20，最大 100）

分页响应格式：

```json
{
  "success": true,
  "data": {
    "items": [...],
    "total": 156,
    "page": 2,
    "pageSize": 20
  }
}
```

---

## WebSocket 接口（规划中）

未来版本将支持 WebSocket 实时通信，用于：
- 实时消息推送
- 流式对话
- 在线状态同步

---

## SDK 和示例代码

### JavaScript/TypeScript

```typescript
import axios from 'axios';

const client = axios.create({
  baseURL: 'https://api.oblivious.ai/api/v1',
  headers: {
    'Content-Type': 'application/json'
  }
});

// 登录
const login = async (username: string, password: string) => {
  const response = await client.post('/login', { username, password });
  return response.data;
};

// 发送消息
const sendMessage = async (sessionId: string, content: string, token: string) => {
  const response = await client.post(
    '/chat/messages',
    { session_id: sessionId, content },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  return response.data;
};
```

### Python

```python
import requests

class ObliviousClient:
    def __init__(self, base_url='https://api.oblivious.ai/api/v1'):
        self.base_url = base_url
        self.token = None
    
    def login(self, username, password):
        response = requests.post(f'{self.base_url}/login', json={
            'username': username,
            'password': password
        })
        data = response.json()
        self.token = data['data']['access_token']
        return data
    
    def send_message(self, session_id, content):
        response = requests.post(
            f'{self.base_url}/chat/messages',
            json={'session_id': session_id, 'content': content},
            headers={'Authorization': f'Bearer {self.token}'}
        )
        return response.json()
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type Client struct {
    BaseURL string
    Token   string
}

func (c *Client) Login(username, password string) error {
    body, _ := json.Marshal(map[string]string{
        "username": username,
        "password": password,
    })
    
    resp, err := http.Post(c.BaseURL+"/login", "application/json", bytes.NewBuffer(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    c.Token = result["data"].(map[string]interface{})["access_token"].(string)
    return nil
}
```

---

## 变更历史

### v1.0.0 (2024-11-21)
- 初始版本发布
- 实现用户认证相关接口
- 实现对话会话相关接口
- 实现消息发送接口

---

**文档版本**: v1.0.0  
**最后更新**: 2024 年 11 月 21 日  
**维护团队**: Oblivious 开发团队
