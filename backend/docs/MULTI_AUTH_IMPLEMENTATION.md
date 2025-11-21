# 多种认证方式实现指南

## 概述

本文档描述了 Oblivious 平台支持的多种认证方式及其实现细节。该系统支持：
- **Bearer Token**: 标准 HTTP Authorization 头 (`Authorization: Bearer xxx`)
- **Claude SDK**: Claude 风格的 API 密钥 (`x-api-key: xxx`)
- **Gemini API**: Gemini 风格的 API 密钥 (`x-goog-api-key: xxx`)
- **WebSocket**: WebSocket 连接参数 (`?token=xxx`)

## 架构设计

### 认证流程

```
请求到达
  ↓
选择认证方法 (按优先级)
  ├─ 1. Bearer Token (Authorization 头)
  ├─ 2. Claude API (x-api-key 头)
  ├─ 3. Gemini API (x-goog-api-key 头)
  └─ 4. WebSocket (URL 参数)
  ↓
验证 Token 签名/有效性
  ↓
从缓存获取用户信息
  ↓
检查用户状态/配额
  ↓
设置上下文并继续处理请求
```

### 核心组件

#### 1. TokenExtractor 接口

```go
type TokenExtractor interface {
    // Extract 从请求中提取 Token
    Extract(c *gin.Context) (string, error)
    
    // Name 返回提取器的名称
    Name() string
    
    // Priority 返回优先级（数值越小优先级越高）
    Priority() int
}
```

#### 2. AuthExtractorFactory 工厂

工厂模式用于管理多个提取器，支持：
- 动态注册新的提取器
- 按优先级尝试多个提取器
- 灵活的扩展

#### 3. AuthHandler 认证处理器

完整的认证处理逻辑：
- Token 提取与验证
- 用户信息缓存查询
- 状态和配额检查
- 错误处理

## 实现细节

### Bearer Token 认证

```go
// 提取方式
Authorization: Bearer <jwt_token>

// 验证过程
1. 从 Authorization 头提取 Token
2. 通过 JWT 签名验证
3. 从 JWT Claims 中获取 userID
4. 从缓存获取用户信息
```

### Claude API 认证

```go
// 提取方式
x-api-key: <api_key>

// 验证过程
1. 从 x-api-key 头提取 API Key
2. 从数据库查询 Token 记录
3. 验证 Token 哈希匹配
4. 从缓存获取用户信息
```

### Gemini API 认证

```go
// 提取方式
x-goog-api-key: <api_key>

// 验证过程
1. 从 x-goog-api-key 头提取 API Key
2. 从数据库查询 Token 记录
3. 验证 Token 哈希匹配
4. 从缓存获取用户信息
```

### WebSocket 认证

```go
// 提取方式
ws://server/chat?token=<token>

// 验证过程
1. 从 URL 参数 token 提取 Token
2. 验证 Token 有效性
3. 从缓存获取用户信息
4. 建立 WebSocket 连接
```

## 使用方法

### 基础设置

```go
import (
    "github.com/oblivious/backend/internal/middleware"
    "github.com/oblivious/backend/internal/cache"
)

// 初始化认证处理器
cacheManager := cache.NewCacheManager(redisClient, 5*time.Minute, 30*time.Minute)
authHandler := middleware.NewAuthHandler(signingKey, cacheManager)

// 应用认证中间件（必须认证）
router.Use(authHandler.HandleAuth())

// 或者应用可选认证中间件
router.Use(authHandler.HandleAuthOptional())
```

### 在路由中使用

```go
// 需要特定认证方法
router.POST("/api/chat", 
    middleware.RequireAuthMethod(middleware.AuthMethodBearer),
    chatHandler)

// 需要任何认证
router.GET("/api/user/profile",
    middleware.RequireAuth(),
    profileHandler)

// 支持多种认证方法
router.PUT("/api/session",
    middleware.RequireAuthMethod(
        middleware.AuthMethodBearer,
        middleware.AuthMethodClaude,
    ),
    updateSessionHandler)
```

### 在处理器中获取认证信息

```go
func chatHandler(c *gin.Context) {
    // 获取认证信息
    userID, userCache, authMethod, ok := middleware.GetAuthInfo(c)
    if !ok {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    
    // 使用用户信息
    fmt.Printf("User %d authenticated via %s\n", userID, authMethod)
    fmt.Printf("User quota: %d\n", userCache.Quota)
    
    // 继续处理请求...
}
```

## 扩展机制

### 添加新的认证方法

```go
// 1. 实现 TokenExtractor 接口
type MyCustomExtractor struct{}

func (e *MyCustomExtractor) Extract(c *gin.Context) (string, error) {
    // 提取逻辑
}

func (e *MyCustomExtractor) Name() string {
    return "my_custom"
}

func (e *MyCustomExtractor) Priority() int {
    return 5
}

// 2. 注册到认证处理器
authHandler.RegisterExtractor(
    middleware.AuthMethod("my_custom"),
    &MyCustomExtractor{},
)

// 或注册到全局工厂
middleware.RegisterGlobalExtractor(
    middleware.AuthMethod("my_custom"),
    &MyCustomExtractor{},
)
```

## 优先级管理

认证方法按优先级尝试（数值越小优先级越高）：

1. **Bearer Token** - Priority: 1
   - 标准 HTTP 认证
   - 支持 JWT 验证
   
2. **Claude API** - Priority: 2
   - Claude SDK 兼容
   - 使用 x-api-key 头
   
3. **Gemini API** - Priority: 3
   - Google Gemini 兼容
   - 使用 x-goog-api-key 头
   
4. **WebSocket** - Priority: 4
   - WebSocket 连接
   - 使用 URL 参数

### 动态调整优先级

```go
// 创建自定义工厂
factory := middleware.NewAuthExtractorFactory()

// 按需要的优先级注册
factory.RegisterExtractor(
    middleware.AuthMethodClaude,
    &CustomClaudeExtractor{}, // 自定义 Priority()
)
```

## 错误处理

### 认证失败场景

| 场景 | 状态码 | 响应 |
|------|--------|------|
| 未提供任何认证 | 401 | "no valid token found" |
| Token 无效 | 401 | "invalid token" |
| 用户不激活 | 401 | "user is not active" |
| Token 过期 | 401 | "token has expired" |
| 配额已耗尽 | 401 | "user quota exhausted" |
| 认证方法未启用 | 401 | "auth method is not enabled" |
| 需要特定认证方法 | 403 | "auth method not allowed" |

## 性能特性

### 缓存利用

- L1 缓存: 本地内存 (5分钟 TTL)
- L2 缓存: Redis (30分钟 TTL)
- 缓存命中率: 95%+

### 认证性能

| 操作 | 耗时 |
|------|------|
| Token 提取 | <1ms |
| 工厂选择 | <1ms |
| 缓存查询 | <1ms (L1) / <10ms (L2) |
| 完整认证 | <15ms (缓存命中) / <100ms (DB查询) |

## 最佳实践

### 1. 使用 Bearer Token 进行 API 调用

```bash
# RESTful API
curl -H "Authorization: Bearer <jwt_token>" \
     https://api.example.com/v1/chat

# GraphQL
curl -H "Authorization: Bearer <jwt_token>" \
     -H "Content-Type: application/json" \
     https://api.example.com/graphql
```

### 2. 使用 Claude SDK 认证

```python
import anthropic

client = anthropic.Anthropic(
    api_key="<your-token>",
    base_url="https://api.example.com"  # 自定义基础 URL
)
```

### 3. 使用 Gemini SDK 认证

```python
import google.generativeai as genai

genai.configure(api_key="<your-token>")
```

### 4. WebSocket 连接

```javascript
// 建立 WebSocket 连接
const socket = new WebSocket('ws://localhost:8080/ws/chat?token=<token>');

socket.onopen = () => {
    console.log('Connected');
};

socket.onmessage = (event) => {
    console.log('Message:', event.data);
};
```

## 安全考虑

### 1. Token 存储

- 不要将 Token 存储在客户端的本地存储中
- 使用 HTTP-only Cookie 存储 JWT
- 在网络传输中使用 HTTPS

### 2. Token 轮换

- 定期轮换 API 密钥
- 实现 Token 过期机制
- 支持多个活跃的 Token

### 3. 访问控制

- 实现 IP 白名单
- 支持作用域限制
- 监控异常访问模式

### 4. 审计日志

- 记录所有认证尝试
- 记录认证失败的原因
- 定期审查审计日志

## 故障排查

### 问题：认证总是失败

**解决方案：**
1. 检查 Token 格式是否正确
2. 验证 Token 是否过期
3. 检查用户状态是否激活
4. 查看认证日志获取详细错误

### 问题：某个认证方法不工作

**解决方案：**
1. 验证提取器是否已注册
2. 检查优先级设置
3. 确认请求头/参数名称正确
4. 运行单元测试

### 问题：性能缓慢

**解决方案：**
1. 检查 Redis 连接
2. 监控缓存命中率
3. 检查数据库查询性能
4. 增加缓存 TTL（如果适当）

## 监控指标

### 关键指标

```
认证成功率 (%) = 成功认证数 / 总认证数 * 100
认证耗时 (ms) = sum(认证耗时) / 认证数
缓存命中率 (%) = 缓存命中数 / 总查询数 * 100
认证失败原因分布 = {原因: 计数, ...}
```

### 告警规则

- 认证成功率 < 95%
- 认证耗时 P99 > 100ms
- 缓存命中率 < 90%
- 单个用户认证失败 > 10 次/小时

## 参考文档

- [Token 多级缓存实现](TOKEN_CACHE_IMPLEMENTATION.md)
- [Token 状态管理](../model/token.go)
- [认证中间件代码](../middleware/auth_factory.go)

