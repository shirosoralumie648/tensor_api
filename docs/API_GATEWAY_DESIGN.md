# API 网关设计文档

## 概述

API 网关是 Oblivious 系统的统一入口，负责请求路由、鉴权、限流、日志记录等横切关注点。

## 核心职责

### 1. 路由转发

将前端请求路由到对应的后端微服务：

```
客户端请求                          后端服务
─────────────────────────────────────────────────
POST /api/v1/auth/login        →   User Service
POST /api/v1/chat/completions  →   Chat Service
GET  /api/v1/models            →   Relay Service
POST /api/v1/files/upload      →   File Service
```

### 2. JWT 鉴权

#### Token 结构

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user_id",
    "username": "john_doe",
    "role": 1,
    "exp": 1735574400,
    "iat": 1735488000,
    "jti": "unique_token_id"
  },
  "signature": "..."
}
```

#### Token 验证流程

1. 从 HTTP Header 提取 Token：`Authorization: Bearer <token>`
2. 验证 Token 签名和过期时间
3. 从 Redis 检查 Token 是否在黑名单（登出后的 Token）
4. 将用户信息注入到请求上下文，传递给下游服务

#### Refresh Token 机制

- Access Token 有效期：2 小时
- Refresh Token 有效期：7 天
- 使用 Refresh Token 换取新的 Access Token

### 3. 限流策略

#### 限流维度

| 维度       | 限流规则                          | 实现方式          |
|------------|-----------------------------------|-------------------|
| IP         | 1000 req/min                      | Redis + 滑动窗口  |
| 用户       | 根据用户等级（普通/VIP/企业）      | Redis + 令牌桶    |
| 接口       | 敏感接口（登录）10 req/min        | Redis             |
| 全局       | 100K req/min                      | Nginx             |

#### 限流算法

**令牌桶算法**（Token Bucket）：

```go
type TokenBucket struct {
    Capacity      int       // 桶容量
    RefillRate    int       // 每秒补充速率
    Tokens        int       // 当前令牌数
    LastRefillTime time.Time
}

func (tb *TokenBucket) AllowRequest() bool {
    now := time.Now()
    elapsed := now.Sub(tb.LastRefillTime).Seconds()
    
    // 补充令牌
    refillTokens := int(elapsed * float64(tb.RefillRate))
    tb.Tokens = min(tb.Capacity, tb.Tokens + refillTokens)
    tb.LastRefillTime = now
    
    // 消费令牌
    if tb.Tokens > 0 {
        tb.Tokens--
        return true
    }
    return false
}
```

### 4. 熔断降级

使用 **hystrix-go** 实现熔断：

```go
hystrix.ConfigureCommand("chat_service", hystrix.CommandConfig{
    Timeout:                1000,  // 超时时间(ms)
    MaxConcurrentRequests:  100,   // 最大并发数
    ErrorPercentThreshold:  50,    // 错误率阈值
    RequestVolumeThreshold: 20,    // 最小请求数
    SleepWindow:            5000,  // 熔断后恢复尝试间隔(ms)
})

err := hystrix.Do("chat_service", func() error {
    // 调用 Chat Service
    return chatService.Process(req)
}, func(err error) error {
    // 降级逻辑：返回缓存或默认响应
    return getFallbackResponse()
})
```

### 5. 日志记录

#### 日志格式（JSON）

```json
{
  "timestamp": "2024-11-20T10:30:15.123Z",
  "request_id": "req_abc123xyz",
  "trace_id": "trace_def456uvw",
  "method": "POST",
  "path": "/api/v1/chat/completions",
  "user_id": 12345,
  "ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "status_code": 200,
  "latency_ms": 342,
  "upstream_service": "chat-service",
  "error": null
}
```

#### 日志输出

- 本地开发：输出到 stdout（彩色格式）
- 生产环境：输出到 stdout，由 Fluentd/Filebeat 采集到 Loki

### 6. CORS 处理

```go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        // 白名单检查（生产环境）
        if isAllowedOrigin(origin) {
            c.Header("Access-Control-Allow-Origin", origin)
            c.Header("Access-Control-Allow-Credentials", "true")
            c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
            c.Header("Access-Control-Max-Age", "86400")
        }
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

---

## 服务间通信

### 1. HTTP REST

**优点**：简单、易调试、跨语言
**缺点**：性能相对较低

**使用场景**：API 网关 → 微服务

```go
// 网关调用用户服务
resp, err := http.Post(
    "http://user-service:8080/internal/user/verify",
    "application/json",
    bytes.NewBuffer(reqBody),
)
```

### 2. gRPC

**优点**：性能高、强类型、支持流式
**缺点**：学习曲线陡、调试困难

**使用场景**：微服务之间的高频调用

```protobuf
// chat.proto
syntax = "proto3";

service ChatService {
  rpc SendMessage(ChatRequest) returns (ChatResponse);
  rpc GetHistory(HistoryRequest) returns (stream Message);
}

message ChatRequest {
  string user_id = 1;
  string session_id = 2;
  string content = 3;
  string model = 4;
}
```

### 3. 消息队列 (RabbitMQ)

**优点**：异步解耦、削峰填谷
**缺点**：增加复杂度、消息可能丢失

**使用场景**：异步任务、事件通知

```go
// 对话服务发布计费事件
event := BillingEvent{
    UserID:   12345,
    Model:    "gpt-4",
    InputTokens:  100,
    OutputTokens: 200,
    Timestamp: time.Now(),
}

ch.Publish(
    "billing_exchange",
    "billing.usage",
    false,
    false,
    amqp.Publishing{
        ContentType: "application/json",
        Body:        json.Marshal(event),
    },
)

// 计费服务消费事件
msgs, _ := ch.Consume(
    "billing_queue",
    "billing_consumer",
    false, // auto-ack
    false,
    false,
    false,
    nil,
)

for msg := range msgs {
    var event BillingEvent
    json.Unmarshal(msg.Body, &event)
    
    // 处理计费逻辑
    processBilling(event)
    
    msg.Ack(false)
}
```

---

## API 设计规范

### RESTful 风格

```
GET    /api/v1/users            # 获取用户列表
GET    /api/v1/users/:id        # 获取单个用户
POST   /api/v1/users            # 创建用户
PUT    /api/v1/users/:id        # 更新用户（全量）
PATCH  /api/v1/users/:id        # 更新用户（部分）
DELETE /api/v1/users/:id        # 删除用户
```

### 统一响应格式

#### 成功响应

```json
{
  "success": true,
  "data": {
    "id": 123,
    "username": "john_doe"
  },
  "message": "操作成功",
  "timestamp": "2024-11-20T10:30:15Z"
}
```

#### 错误响应

```json
{
  "success": false,
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "Token 已过期",
    "details": null
  },
  "timestamp": "2024-11-20T10:30:15Z"
}
```

### 错误码设计

```go
const (
    // 通用错误 (1000-1999)
    ErrInternal        = 1000  // 内部错误
    ErrInvalidRequest  = 1001  // 请求参数错误
    ErrNotFound        = 1004  // 资源不存在
    
    // 鉴权错误 (2000-2999)
    ErrUnauthorized    = 2001  // 未登录
    ErrForbidden       = 2003  // 无权限
    ErrInvalidToken    = 2010  // Token 无效
    ErrTokenExpired    = 2011  // Token 过期
    
    // 业务错误 (3000-3999)
    ErrInsufficientQuota = 3001  // 余额不足
    ErrModelNotAvailable = 3002  // 模型不可用
    ErrRateLimitExceeded = 3003  // 请求频率超限
    
    // 第三方错误 (4000-4999)
    ErrUpstreamTimeout  = 4001  // 上游服务超时
    ErrUpstreamError    = 4002  // 上游服务错误
)
```

### 分页参数

```
GET /api/v1/sessions?page=2&page_size=20&sort=-created_at
```

**响应**：

```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 2,
      "page_size": 20,
      "total": 156,
      "total_pages": 8
    }
  }
}
```

---

## 链路追踪

使用 **OpenTelemetry** 实现分布式追踪：

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func HandleRequest(c *gin.Context) {
    tracer := otel.Tracer("gateway")
    ctx, span := tracer.Start(c.Request.Context(), "HandleChatRequest")
    defer span.End()
    
    // 注入 Trace ID 到 HTTP Header
    carrier := propagation.HeaderCarrier(c.Request.Header)
    otel.GetTextMapPropagator().Inject(ctx, carrier)
    
    // 调用下游服务
    resp, err := callChatService(ctx, req)
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
}
```

**Trace ID 传播**：

```
客户端 → 网关 (trace_id: abc123)
         ↓
      对话服务 (trace_id: abc123, span_id: def456)
         ↓
      中转服务 (trace_id: abc123, span_id: ghi789)
```

---

## 性能优化

### 1. 连接池

```go
// HTTP 客户端连接池
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 10 * time.Second,
}
```

### 2. 响应缓存

对于不常变化的数据（模型列表、插件列表），使用 Redis 缓存：

```go
// 先查缓存
cachedModels, err := redis.Get(ctx, "models:list").Result()
if err == nil {
    return json.Unmarshal(cachedModels, &models)
}

// 缓存未命中，查询数据库
models := getModelsFromDB()

// 写入缓存（TTL 5 分钟）
redis.Set(ctx, "models:list", json.Marshal(models), 5*time.Minute)
```

### 3. 请求合并

对于高频的相同请求，使用 **singleflight** 避免缓存击穿：

```go
import "golang.org/x/sync/singleflight"

var g singleflight.Group

func GetUserInfo(userID int) (*User, error) {
    key := fmt.Sprintf("user:%d", userID)
    
    v, err, _ := g.Do(key, func() (interface{}, error) {
        return db.GetUser(userID)
    })
    
    return v.(*User), err
}
```

---

## 部署配置

### 环境变量

```bash
# 服务配置
PORT=8080
GIN_MODE=release

# 数据库
DATABASE_URL=postgres://user:pass@postgres:5432/oblivious?sslmode=disable

# Redis
REDIS_URL=redis://:password@redis:6379/0

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE_HOURS=2
REFRESH_TOKEN_EXPIRE_DAYS=7

# 限流
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=1000

# 上游服务地址
USER_SERVICE_URL=http://user-service:8080
CHAT_SERVICE_URL=http://chat-service:8080
RELAY_SERVICE_URL=http://relay-service:8080
FILE_SERVICE_URL=http://file-service:8080

# 监控
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
PROMETHEUS_PORT=9090
```

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/gateway

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/gateway .

EXPOSE 8080 9090
CMD ["./gateway"]
```

---

## 监控指标

### Prometheus Metrics

```go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
)
```

### 关键指标

| 指标名称                  | 类型      | 说明                  |
|---------------------------|-----------|----------------------|
| http_requests_total       | Counter   | 总请求数              |
| http_request_duration     | Histogram | 请求延迟分布          |
| active_connections        | Gauge     | 当前活跃连接数        |
| rate_limit_exceeded_total | Counter   | 被限流的请求数        |
| auth_failures_total       | Counter   | 鉴权失败次数          |
| upstream_errors_total     | Counter   | 上游服务错误次数      |

---

## 安全措施

### 1. 防止 SQL 注入

使用 ORM（GORM）参数化查询，禁止拼接 SQL。

### 2. 防止 XSS

前端输出时转义 HTML 特殊字符。

### 3. API Key 加密存储

```go
import "crypto/aes"
import "crypto/cipher"

func EncryptAPIKey(key string) string {
    block, _ := aes.NewCipher([]byte(secretKey))
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    ciphertext := gcm.Seal(nonce, nonce, []byte(key), nil)
    return base64.StdEncoding.EncodeToString(ciphertext)
}
```

### 4. 防止暴力破解

登录接口增加验证码（5 次失败后）和账号锁定（10 次失败后锁定 30 分钟）。

---

## 总结

API 网关是 Oblivious 系统的核心枢纽，承担了流量管控、安全防护、可观测性等关键职责。通过合理的架构设计和工程实践，可以确保系统的高可用性和高性能。

