# 管理API文档

## 渠道管理API

### 1. 查询渠道列表
```
GET /api/admin/channels?page=1&page_size=20&type=openai&enabled=true
```

响应:
```json
{
  "total": 10,
  "page": 1,
  "page_size": 20,
  "data": [
    {
      "id": 1,
      "name": "OpenAI主渠道",
      "type": "openai",
      "group": "default",
      "priority": 100,
      "weight": 10,
      "enabled": true,
      "status": 0
    }
  ]
}
```

### 2. 创建渠道
```
POST /api/admin/channels
```

请求体:
```json
{
  "name": "新OpenAI渠道",
  "type": "openai",
  "group": "default",
  "base_url": "https://api.openai.com/v1",
  "api_keys": "sk-xxx,sk-yyy",
  "support_models": "gpt-4o,gpt-4o-mini,gpt-3.5-turbo",
  "priority": 100,
  "weight": 10,
  "enabled": true
}
```

### 3. 更新渠道
```
PUT /api/admin/channels/1
```

请求体:
```json
{
  "priority": 150,
  "weight": 20,
  "enabled": false
}
```

### 4. 删除渠道
```
DELETE /api/admin/channels/1
```

### 5. 测试渠道
```
POST /api/admin/channels/1/test
```

响应:
```json
{
  "channel_id": 1,
  "status": "healthy",
  "latency_ms": 150,
  "message": "连接正常"
}
```

### 6. 批量操作
```
POST /api/admin/channels/batch
```

请求体:
```json
{
  "ids": [1, 2, 3],
  "operation": "enable"  // enable | disable | delete
}
```

---

## 定价管理API

### 1. 查询定价列表
```
GET /api/admin/pricing?enabled=true
```

### 2. 获取模型定价
```
GET /api/admin/pricing/gpt-4o?group=default
```

### 3. 创建定价
```
POST /api/admin/pricing
```

请求体:
```json
{
  "model": "gpt-4o-pro",
  "group": "default",
  "quota_type": 0,
  "model_ratio": 20.0,
  "completion_ratio": 2.0,
  "group_ratio": 1.0,
  "vendor_id": "openai",
  "enabled": true,
  "description": "GPT-4o Pro定价"
}
```

### 4. 计算配额
```
POST /api/admin/pricing/calculate
```

请求体:
```json
{
  "model": "gpt-4o",
  "group": "default",
  "prompt_tokens": 100,
  "completion_tokens": 500
}
```

响应:
```json
{
  "model": "gpt-4o",
  "group": "default",
  "prompt_tokens": 100,
  "completion_tokens": 500,
  "quota": 16500,
  "group_ratio": 1.0
}
```

---

## 统计监控API

### 1. 总览统计
```
GET /api/admin/stats/overview
```

响应:
```json
{
  "total_channels": 10,
  "active_channels": 8,
  "total_requests": 50000,
  "total_tokens": 10000000,
  "total_quota": 500000,
  "avg_response_time": 250.5,
  "success_rate": 0.98,
  "today": {
    "requests": 1500,
    "tokens": 300000,
    "quota": 15000
  }
}
```

### 2. 渠道统计
```
GET /api/admin/stats/channels?days=7
```

响应:
```json
[
  {
    "channel_id": 1,
    "channel_name": "OpenAI主渠道",
    "type": "openai",
    "status": 0,
    "enabled": true,
    "requests": 5000,
    "tokens": 1000000,
    "quota": 50000,
    "avg_latency": 200.5,
    "success_rate": 0.99
  }
]
```

### 3. 模型统计
```
GET /api/admin/stats/models?days=7
```

响应:
```json
[
  {
    "model": "gpt-4o",
    "requests": 3000,
    "tokens": 800000,
    "quota": 40000,
    "avg_latency": 250.0
  }
]
```

### 4. 时间序列数据
```
GET /api/admin/stats/timeseries?days=30
```

响应:
```json
[
  {
    "date": "2025-11-01",
    "requests": 1500,
    "tokens": 300000,
    "quota": 15000
  }
]
```

---

## 使用示例

### 创建渠道完整流程
```bash
# 1. 创建渠道
curl -X POST http://localhost:8080/api/admin/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpenAI测试渠道",
    "type": "openai",
    "api_keys": "sk-test",
    "support_models": "gpt-4o,gpt-3.5-turbo",
    "priority": 100,
    "weight": 10,
    "enabled": true
  }'

# 2. 测试连接
curl -X POST http://localhost:8080/api/admin/channels/1/test

# 3. 查看统计
curl http://localhost:8080/api/admin/stats/channels
```

### 管理定价
```bash
# 1. 创建定价
curl -X POST http://localhost:8080/api/admin/pricing \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "quota_type": 0,
    "model_ratio": 15.0,
    "completion_ratio": 2.0
  }'

# 2. 计算配额
curl -X POST http://localhost:8080/api/admin/pricing/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "prompt_tokens": 100,
    "completion_tokens": 500
  }'
```
