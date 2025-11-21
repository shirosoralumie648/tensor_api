# API 网关设计

## 概述

API 网关是 Oblivious 平台的统一入口，负责请求路由、认证鉴权、限流熔断、日志记录等横切关注点。所有客户端请求都通过网关转发到后端微服务。

## 设计目标

- **统一入口**：所有 API 请求的单一入口点
- **认证鉴权**：集中处理 JWT 验证和权限控制
- **流量控制**：限流、熔断、降级保护
- **服务发现**：动态路由到后端微服务
- **监控审计**：请求日志、指标收集、链路追踪
- **高性能**：低延迟、高吞吐量

## 架构设计

```
┌──────────────────────────────────────────────────────────────┐
│                        客户端请求                              │
└────────────────────────┬─────────────────────────────────────┘
                         │ HTTPS
                         ▼
┌──────────────────────────────────────────────────────────────┐
│                      API 网关 (Gateway)                        │
│                                                                │
│  ┌────────────────────────────────────────────────────────┐  │
│  │              中间件链 (Middleware Chain)                │  │
│  │                                                          │  │
│  │  1. CORS 跨域处理                                        │  │
│  │  2. 日志记录 (Logger)                                    │  │
│  │  3. 链路追踪 (Tracing)                                   │  │
│  │  4. 限流控制 (Rate Limiter)                              │  │
│  │  5. JWT 认证 (Auth Middleware)                           │  │
│  │  6. 权限验证 (Permission Check)                          │  │
│  │  7. 熔断降级 (Circuit Breaker)                           │  │
│  │                                                          │  │
│  └────────────────────────┬───────────────────────────────┘  │
│                           │                                   │
│  ┌────────────────────────▼───────────────────────────────┐  │
│  │                  路由器 (Router)                        │  │
│  │                                                          │  │
│  │  - 路径匹配                                              │  │
│  │  - 方法匹配                                              │  │
│  │  - 参数提取                                              │  │
│  │                                                          │  │
│  └────────────────────────┬───────────────────────────────┘  │
│                           │                                   │
│  ┌────────────────────────▼───────────────────────────────┐  │
│  │              代理处理器 (Proxy Handler)                 │  │
│  │                                                          │  │
│  │  - 请求转发                                              │  │
│  │  - 响应聚合                                              │  │
│  │  - 错误处理                                              │  │
│  │                                                          │  │
│  └────────────────────────┬───────────────────────────────┘  │
└───────────────────────────┼───────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
            ▼               ▼               ▼
    ┌────────────┐  ┌────────────┐  ┌────────────┐
    │ 用户服务   │  │ 对话服务   │  │ 计费服务   │
    └────────────┘  └────────────┘  └────────────┘
```

## 核心功能模块

### 1. 路由管理

#### 路由配置

```go
// backend/cmd/gateway/main.go
func setupRoutes(r *gin.Engine) {
    // API v1 版本组
    api := r.Group("/api/v1")
    
    // 公开接口（无需认证）
    public := api.Group("")
    {
        public.POST("/auth/register", handler.Register)
        public.POST("/auth/login", handler.Login)
        public.GET("/health", handler.Health)
    }
    
    // 认证接口（需要 JWT）
    auth := api.Group("")
    auth.Use(middleware.AuthMiddleware())
    {
        // 用户相关
        auth.GET("/user/profile", proxyToUser)
        auth.PUT("/user/profile", proxyToUser)
        
        // 对话相关
        auth.POST("/chat/completions", proxyToChat)
        auth.GET("/chat/sessions", proxyToChat)
        
        // 计费相关
        auth.GET("/billing/usage", proxyToBilling)
        auth.POST("/billing/recharge", proxyToBilling)
    }
}
```

#### 路由表

| 方法 | 路径 | 目标服务 | 认证 | 说明 |
|-----|------|---------|------|------|
| POST | /api/v1/auth/register | user | ❌ | 用户注册 |
| POST | /api/v1/auth/login | user | ❌ | 用户登录 |
| GET | /api/v1/user/profile | user | ✅ | 获取用户信息 |
| PUT | /api/v1/user/profile | user | ✅ | 更新用户信息 |
| POST | /api/v1/chat/completions | chat | ✅ | AI 对话 |
| GET | /api/v1/chat/sessions | chat | ✅ | 获取会话列表 |
| GET | /api/v1/billing/usage | billing | ✅ | 查询用户额度 |
| POST | /api/v1/billing/recharge | billing | ✅ | 充值 |
| GET | /api/v1/models | relay | ✅ | 获取模型列表 |

### 2. 认证中间件

#### JWT 认证流程

```
┌─────────────────────────────────────────────────────────────┐
│                      认证中间件流程                           │
└─────────────────────────────────────────────────────────────┘

1. 提取 Token
   ├─> 从 Header 获取: Authorization: Bearer <token>
   └─> 从 Query 获取: ?token=<token> (备用)

2. 验证 Token
   ├─> 解析 JWT
   ├─> 验证签名 (HMAC-SHA256)
   ├─> 检查过期时间
   └─> 验证 Issuer 和 Audience

3. 提取用户信息
   ├─> 从 Claims 获取 user_id
   ├─> 从 Claims 获取 role
   └─> 从 Claims 获取 permissions

4. 设置上下文
   ├─> ctx.Set("user_id", userId)
   ├─> ctx.Set("role", role)
   └─> ctx.Next() 继续执行

5. 错误处理
   ├─> Token 缺失: 401 Unauthorized
   ├─> Token 无效: 401 Unauthorized
   ├─> Token 过期: 401 Token Expired
   └─> 权限不足: 403 Forbidden
```

#### 实现代码

```go
// backend/internal/middleware/auth.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取 Token
        token := c.GetHeader("Authorization")
        if token == "" {
            token = c.Query("token")
        }
        
        if token == "" {
            c.JSON(401, gin.H{"error": "Authorization token required"})
            c.Abort()
            return
        }
        
        // 移除 "Bearer " 前缀
        token = strings.TrimPrefix(token, "Bearer ")
        
        // 2. 验证 Token
        claims, err := utils.ValidateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // 3. 提取用户信息
        userId := claims.UserID
        role := claims.Role
        
        // 4. 设置上下文
        c.Set("user_id", userId)
        c.Set("role", role)
        
        c.Next()
    }
}
```

### 3. 限流控制

#### 限流策略

**多层限流**：

1. **IP 限流**：防止单个 IP 恶意攻击
   - 默认：100 请求/分钟
   - 算法：令牌桶

2. **用户限流**：防止用户滥用
   - 免费用户：20 请求/分钟
   - 付费用户：100 请求/分钟
   - VIP 用户：500 请求/分钟

3. **API 限流**：保护特定接口
   - AI 对话：10 请求/分钟
   - 文件上传：5 请求/分钟

#### 实现代码

```go
// backend/internal/middleware/ratelimit.go
type RateLimitConfig struct {
    Rate  int           // 请求速率（每分钟）
    Burst int           // 突发容量
    Key   func(*gin.Context) string  // 限流键生成函数
}

func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
    limiters := sync.Map{}
    
    return func(c *gin.Context) {
        key := config.Key(c)
        
        // 获取或创建限流器
        limiterInterface, _ := limiters.LoadOrStore(key, 
            rate.NewLimiter(rate.Limit(config.Rate), config.Burst))
        limiter := limiterInterface.(*rate.Limiter)
        
        // 检查是否允许请求
        if !limiter.Allow() {
            c.JSON(429, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": limiter.Reserve().Delay().Seconds(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// IP 限流
func IPRateLimitKey(c *gin.Context) string {
    return "ip:" + c.ClientIP()
}

// 用户限流
func UserRateLimitKey(c *gin.Context) string {
    userId, _ := c.Get("user_id")
    return "user:" + fmt.Sprint(userId)
}
```

### 4. 熔断降级

#### 熔断器状态机

```
         成功次数达到阈值
    ┌──────────────────────┐
    │                      │
    ▼                      │
┌────────┐  失败率过高  ┌──────────┐  超时时间  ┌────────────┐
│ CLOSED │─────────────>│  OPEN    │───────────>│ HALF_OPEN  │
└────────┘              └──────────┘            └──────┬─────┘
    ▲                        │                         │
    │      请求全部失败       │                         │
    └────────────────────────┴─────────────────────────┘
                       请求成功
```

**状态说明**：

- **CLOSED（关闭）**：正常状态，请求正常转发
- **OPEN（打开）**：故障状态，直接返回错误，不转发请求
- **HALF_OPEN（半开）**：恢复测试，允许少量请求通过

#### 熔断配置

```go
type CircuitBreakerConfig struct {
    Timeout            time.Duration  // 请求超时时间
    MaxConcurrent      int           // 最大并发请求数
    ErrorThreshold     int           // 错误阈值（触发熔断）
    SuccessThreshold   int           // 成功阈值（恢复正常）
    SleepWindow        time.Duration // 熔断时间窗口
}
```

### 5. 请求代理

#### 代理逻辑

```go
// backend/internal/gateway/handler/proxy.go
func ProxyToService(serviceName string, serviceURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 构建目标 URL
        targetURL := serviceURL + c.Request.URL.Path
        if c.Request.URL.RawQuery != "" {
            targetURL += "?" + c.Request.URL.RawQuery
        }
        
        // 2. 创建代理请求
        proxyReq, err := http.NewRequest(
            c.Request.Method,
            targetURL,
            c.Request.Body,
        )
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to create proxy request"})
            return
        }
        
        // 3. 复制请求头
        for key, values := range c.Request.Header {
            for _, value := range values {
                proxyReq.Header.Add(key, value)
            }
        }
        
        // 4. 添加追踪头
        proxyReq.Header.Set("X-Request-ID", generateRequestID())
        proxyReq.Header.Set("X-Forwarded-For", c.ClientIP())
        
        // 5. 发送请求
        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(proxyReq)
        if err != nil {
            c.JSON(502, gin.H{"error": "Service unavailable"})
            return
        }
        defer resp.Body.Close()
        
        // 6. 复制响应头
        for key, values := range resp.Header {
            for _, value := range values {
                c.Header(key, value)
            }
        }
        
        // 7. 返回响应
        c.Status(resp.StatusCode)
        io.Copy(c.Writer, resp.Body)
    }
}
```

#### 流式响应代理

```go
func ProxyStreamResponse(c *gin.Context, resp *http.Response) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    
    flusher, ok := c.Writer.(http.Flusher)
    if !ok {
        c.JSON(500, gin.H{"error": "Streaming not supported"})
        return
    }
    
    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            break
        }
        
        c.Writer.Write(line)
        flusher.Flush()
    }
}
```

### 6. 日志和监控

#### 请求日志

```go
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        // 处理请求
        c.Next()
        
        // 记录日志
        latency := time.Since(start)
        statusCode := c.Writer.Status()
        
        log.Info().
            Str("method", c.Request.Method).
            Str("path", path).
            Int("status", statusCode).
            Dur("latency", latency).
            Str("ip", c.ClientIP()).
            Msg("Request completed")
    }
}
```

#### 指标收集

```go
var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "gateway_requests_total",
            Help: "Total number of requests",
        },
        []string{"method", "path", "status"},
    )
    
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "gateway_request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"method", "path"},
    )
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := fmt.Sprint(c.Writer.Status())
        
        requestCounter.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
        
        requestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}
```

## 服务发现

### 静态配置

```yaml
# config/gateway.yaml
services:
  user:
    url: http://user-service:8081
    timeout: 5s
    
  chat:
    url: http://chat-service:8082
    timeout: 30s
    
  relay:
    url: http://relay-service:8083
    timeout: 60s
    
  billing:
    url: http://billing-service:8084
    timeout: 5s
```

### Kubernetes 服务发现

```go
func getServiceURL(serviceName string) string {
    // Kubernetes 内部 DNS
    return fmt.Sprintf("http://%s-service.default.svc.cluster.local", serviceName)
}
```

## 错误处理

### 统一错误响应

```json
{
    "error": {
        "code": "INVALID_TOKEN",
        "message": "The provided token is invalid or expired",
        "details": {
            "reason": "token_expired",
            "expired_at": "2024-01-01T00:00:00Z"
        }
    },
    "request_id": "req_abc123"
}
```

### 错误码定义

| 错误码 | HTTP 状态 | 说明 |
|-------|----------|------|
| INVALID_TOKEN | 401 | Token 无效或过期 |
| INSUFFICIENT_QUOTA | 403 | 额度不足 |
| RATE_LIMIT_EXCEEDED | 429 | 请求频率超限 |
| SERVICE_UNAVAILABLE | 502 | 后端服务不可用 |
| GATEWAY_TIMEOUT | 504 | 网关超时 |

## 性能优化

### 连接池

```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

### 响应缓存

对于幂等的 GET 请求，可以使用 Redis 缓存响应：

```go
func CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != "GET" {
            c.Next()
            return
        }
        
        key := "cache:" + c.Request.URL.Path
        
        // 尝试从缓存获取
        cached, err := redis.Get(key).Result()
        if err == nil {
            c.Data(200, "application/json", []byte(cached))
            return
        }
        
        // 缓存未命中，继续处理
        c.Next()
        
        // 缓存响应
        if c.Writer.Status() == 200 {
            // 保存响应到缓存
        }
    }
}
```

## 相关文档

- [架构设计](ARCHITECTURE.md)
- [API 参考](API_REFERENCE.md)
- [部署指南](PRODUCTION_DEPLOYMENT_GUIDE.md)
