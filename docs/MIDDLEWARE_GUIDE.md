# Oblivious 中间件系统完整指南

> 📅 生成时间: 2025-11-22  
> 🎯 目标: 详细说明所有中间件的实现和使用方法

---

## 📋 目录

- [1. 中间件概览](#1-中间件概览)
- [2. 认证中间件](#2-认证中间件)
- [3. 权限控制中间件](#3-权限控制中间件)
- [4. 限流中间件](#4-限流中间件)
- [5. 日志中间件](#5-日志中间件)
- [6. CORS中间件](#6-cors中间件)
- [7. 安全中间件](#7-安全中间件)
- [8. 其他中间件](#8-其他中间件)
- [9. 使用示例](#9-使用示例)

---

## 1. 中间件概览

### 文件列表

```
backend/internal/middleware/
├── auth.go              - JWT认证中间件
├── auth_cached.go       - 带缓存的认证中间件
├── auth_factory.go      - 认证工厂
├── auth_handler.go      - 认证处理器
├── rbac.go              - 基于角色的访问控制(RBAC)
├── rate_limit.go        - 限流中间件
├── cors.go              - 跨域资源共享
├── logger.go            - 请求日志
├── request_id.go        - 请求ID追踪
├── security.go          - 安全防护
└── README.md            - 说明文档
```

### 中间件执行顺序

```
请求 → Recovery → RequestID → Logger → CORS → Auth → LoadPermissions → RateLimit → 业务Handler
```

---

## 2. 认证中间件

### 2.1 文件: `auth.go`

#### 核心类型

```go
// Claims JWT声明结构
type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}
```

#### 核心函数

##### `AuthMiddleware(signingKey []byte) gin.HandlerFunc`
**功能**: 标准JWT认证中间件

**流程**:
1. 从Authorization头提取Bearer token
2. 解析并验证JWT签名
3. 提取用户信息到上下文
4. 失败返回401

**使用**:
```go
router.Use(middleware.AuthMiddleware([]byte("your-secret-key")))
```

**上下文变量**:
- `user_id`: 用户ID
- `token`: JWT token字符串

---

##### `APIKeyAuthMiddleware() gin.HandlerFunc`
**功能**: API密钥认证中间件

**流程**:
1. 从X-API-Key头或api_key查询参数提取密钥
2. 验证API密钥（需要数据库查询）
3. 失败返回401

**使用**:
```go
router.Use(middleware.APIKeyAuthMiddleware())
```

**适用场景**:
- 机器对机器通信
- 第三方API集成
- 无用户上下文的服务调用

---

##### `ParseToken(tokenString string, signingKey []byte) (*Claims, error)`
**功能**: 解析JWT token

**参数**:
- `tokenString`: JWT字符串
- `signingKey`: 签名密钥

**返回**:
- `*Claims`: 解析后的声明
- `error`: 解析错误

**验证项**:
- HMAC签名算法
- token有效性
- 过期时间

---

##### `ExtractUserID(c *gin.Context) (string, error)`
**功能**: 从上下文提取用户ID

**返回**:
- 用户ID字符串
- 错误（如果不存在）

**使用**:
```go
userID, err := middleware.ExtractUserID(c)
if err != nil {
    c.JSON(401, gin.H{"error": "unauthorized"})
    return
}
```

---

##### `ExtractToken(c *gin.Context) (string, error)`
**功能**: 从上下文提取token

**使用**:
```go
token, err := middleware.ExtractToken(c)
```

---

### 2.2 文件: `auth_cached.go`

#### 核心函数

##### `CachedAuthMiddleware(cache cache.Cache, signingKey []byte) gin.HandlerFunc`
**功能**: 带缓存的认证中间件

**优化点**:
- 缓存已验证的token
- 减少重复的JWT解析
- 降低CPU消耗

**缓存策略**:
- Key: `auth:token:{token_hash}`
- TTL: 5分钟
- 存储: Claims结构

**流程**:
```
1. 提取token
2. 检查缓存中是否存在
   └─> 存在: 直接使用缓存的Claims
   └─> 不存在: 解析token并缓存
3. 设置上下文变量
```

**使用**:
```go
router.Use(middleware.CachedAuthMiddleware(redisCache, []byte("secret")))
```

---

### 2.3 文件: `auth_factory.go`

#### 核心函数

##### `AuthFactory` 结构体
**功能**: 创建不同类型的认证中间件

**方法**:

```go
type AuthFactory struct {
    jwtSecret  []byte
    cache      cache.Cache
    enableCache bool
}

// NewAuthFactory 创建认证工厂
func NewAuthFactory(jwtSecret []byte, cache cache.Cache) *AuthFactory

// JWT 创建JWT认证中间件
func (af *AuthFactory) JWT() gin.HandlerFunc

// JWTWithCache 创建带缓存的JWT认证
func (af *AuthFactory) JWTWithCache() gin.HandlerFunc

// APIKey 创建API密钥认证
func (af *AuthFactory) APIKey() gin.HandlerFunc

// Optional 创建可选认证中间件（认证失败不阻止请求）
func (af *AuthFactory) Optional() gin.HandlerFunc
```

**使用示例**:
```go
authFactory := middleware.NewAuthFactory([]byte("secret"), redisCache)

// 必须认证
protectedRoutes.Use(authFactory.JWT())

// 可选认证
publicRoutes.Use(authFactory.Optional())
```

---

### 2.4 文件: `auth_handler.go`

#### 核心函数

##### `GetAuthInfo(c *gin.Context) (userID int, username string, role int, ok bool)`
**功能**: 从上下文获取完整的认证信息

**返回值**:
- `userID`: 用户ID
- `username`: 用户名
- `role`: 角色ID
- `ok`: 是否成功获取

**使用**:
```go
userID, username, role, ok := middleware.GetAuthInfo(c)
if !ok {
    c.JSON(401, gin.H{"error": "unauthorized"})
    return
}
```

---

##### `RequireAuth(c *gin.Context) bool`
**功能**: 检查是否已认证

**使用**:
```go
if !middleware.RequireAuth(c) {
    return
}
```

---

## 3. 权限控制中间件

### 3.1 文件: `rbac.go`

#### RBAC Manager

```go
type RBACManager struct {
    permissionCache cache.Cache
    ttl             time.Duration
}
```

#### 核心函数

##### `NewRBACManager(ttl time.Duration) *RBACManager`
**功能**: 创建RBAC管理器

**参数**:
- `ttl`: 权限缓存有效期

---

##### `GetUserPermissions(c *gin.Context, userID int) (*model.UserPermissions, error)`
**功能**: 获取用户的所有权限

**缓存策略**:
- Key: `user_permissions:{userID}`
- TTL: 配置的ttl值
- 自动刷新

**返回结构**:
```go
type UserPermissions struct {
    UserID      int
    Roles       []string          // 角色列表
    Permissions []PermissionDTO   // 权限列表
    CachedAt    time.Time
    ExpireAt    time.Time
}
```

---

##### `LoadUserPermissions(rbacManager *RBACManager) gin.HandlerFunc`
**功能**: 加载用户权限到上下文

**使用位置**: 在认证中间件之后

**使用**:
```go
router.Use(middleware.AuthMiddleware(secret))
router.Use(middleware.LoadUserPermissions(rbacManager))
```

---

#### 权限检查中间件

##### `RequirePermission(permission string) gin.HandlerFunc`
**功能**: 要求特定权限

**使用**:
```go
// 需要user.create权限
router.POST("/api/users", middleware.RequirePermission("user.create"), handler)
```

**示例权限名**:
- `user.create` - 创建用户
- `user.update` - 更新用户
- `user.delete` - 删除用户
- `channel.manage` - 管理渠道
- `pricing.edit` - 编辑定价

---

##### `RequirePermissions(permissions ...string) gin.HandlerFunc`
**功能**: 要求任意一个权限（OR关系）

**使用**:
```go
// 需要data.create 或 data.admin 任一权限
router.POST("/api/data", 
    middleware.RequirePermissions("data.create", "data.admin"), 
    handler)
```

---

##### `RequireAllPermissions(permissions ...string) gin.HandlerFunc`
**功能**: 要求所有权限（AND关系）

**使用**:
```go
// 需要同时具备data.delete和data.verify权限
router.DELETE("/api/data/:id", 
    middleware.RequireAllPermissions("data.delete", "data.verify"), 
    handler)
```

---

#### 角色检查中间件

##### `RequireRole(role string) gin.HandlerFunc`
**功能**: 要求特定角色

**使用**:
```go
// 需要admin角色
router.GET("/api/admin/settings", middleware.RequireRole("admin"), handler)
```

**常见角色**:
- `admin` - 管理员
- `user` - 普通用户
- `vip` - VIP用户
- `auditor` - 审计员

---

##### `RequireRoles(roles ...string) gin.HandlerFunc`
**功能**: 要求任意一个角色

**使用**:
```go
// 需要admin或auditor角色
router.POST("/api/audit", 
    middleware.RequireRoles("admin", "auditor"), 
    handler)
```

---

#### 资源访问控制

##### `CheckResourceAccess(userPerms *UserPermissions, resource string, action string) (bool, string)`
**功能**: 检查用户是否可以访问特定资源

**参数**:
- `userPerms`: 用户权限集合
- `resource`: 资源名称
- `action`: 动作类型

**返回**:
- `bool`: 是否有权限
- `string`: 失败原因

**使用**:
```go
userPerms, _ := c.Get("user_permissions")
hasAccess, reason := middleware.CheckResourceAccess(
    userPerms.(*model.UserPermissions),
    "channel",
    "delete",
)

if !hasAccess {
    c.JSON(403, gin.H{"error": reason})
    return
}
```

---

#### 辅助函数

##### `GetUserRoleNames(c *gin.Context) []string`
**功能**: 获取用户的所有角色名称

**使用**:
```go
roles := middleware.GetUserRoleNames(c)
if containsRole(roles, "admin") {
    // 管理员逻辑
}
```

---

##### `GetUserPermissionNames(c *gin.Context) []string`
**功能**: 获取用户的所有权限名称

**使用**:
```go
permissions := middleware.GetUserPermissionNames(c)
```

---

## 4. 限流中间件

### 4.1 文件: `rate_limit.go`

#### 配置结构

```go
type RateLimitConfig struct {
    Rate  int           // 每秒请求数
    Burst int           // 突发容量
    TTL   time.Duration // 过期时间
}
```

#### 核心函数

##### `RateLimitMiddleware(cfg *RateLimitConfig) gin.HandlerFunc`
**功能**: 令牌桶算法限流

**算法**: Token Bucket (令牌桶)

**限流维度**:
1. 已认证用户: 按user_id限流
2. 未认证请求: 按IP地址限流

**实现方式**: Redis + Lua脚本（原子操作）

**使用**:
```go
// 每分钟10个请求，突发容量20
router.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
    Rate:  10,
    Burst: 20,
    TTL:   time.Minute,
}))
```

**响应**:
- 成功: 继续处理
- 超限: 返回429 (Too Many Requests)

---

#### Lua脚本逻辑

```lua
-- 令牌桶算法
1. 获取上次时间和剩余令牌数
2. 计算时间间隔
3. 补充令牌 = min(burst, tokens + elapsed * rate)
4. 尝试消费1个令牌
5. 成功: 更新状态，返回1
6. 失败: 返回0
```

**优势**:
- 原子操作，并发安全
- 分布式限流
- 平滑流量

---

##### `checkRateLimit(ctx context.Context, key string, cfg *RateLimitConfig) (bool, error)`
**功能**: 执行限流检查

**参数**:
- `key`: 限流键（user:123 或 ip:192.168.1.1）
- `cfg`: 限流配置

**返回**:
- `bool`: 是否允许通过
- `error`: 检查错误

---

## 5. 日志中间件

### 5.1 文件: `logger.go`

#### 核心函数

##### `LoggerMiddleware() gin.HandlerFunc`
**功能**: 记录HTTP请求日志

**记录内容**:
```go
{
    "timestamp": "2025-11-22T10:00:00Z",
    "request_id": "abc123",
    "method": "POST",
    "path": "/api/chat/completions",
    "status": 200,
    "latency": "125ms",
    "user_id": "123",
    "ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "error": ""  // 如果有错误
}
```

**日志级别**:
- 2xx: Info
- 4xx: Warn
- 5xx: Error

**使用**:
```go
router.Use(middleware.LoggerMiddleware())
```

---

##### `LoggerWithConfig(config LoggerConfig) gin.HandlerFunc`
**功能**: 带配置的日志中间件

**配置选项**:
```go
type LoggerConfig struct {
    SkipPaths    []string  // 跳过的路径
    TimeFormat   string    // 时间格式
    UTC          bool      // 使用UTC时间
    SkipBodyLog  bool      // 跳过请求体日志
}
```

**使用**:
```go
router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
    UTC: true,
}))
```

---

## 6. CORS中间件

### 6.1 文件: `cors.go`

#### 核心函数

##### `CORSMiddleware() gin.HandlerFunc`
**功能**: 处理跨域资源共享

**配置**:
```go
config := cors.DefaultConfig()
config.AllowAllOrigins = true
config.AllowCredentials = true
config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
config.AllowHeaders = []string{"*"}
config.ExposeHeaders = []string{"Content-Length", "Content-Type", "Authorization"}
config.MaxAge = 86400  // 24小时
```

**响应头**:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Credentials: true`
- `Access-Control-Allow-Methods: GET, POST, ...`
- `Access-Control-Allow-Headers: *`
- `Access-Control-Max-Age: 86400`

**使用**:
```go
router.Use(middleware.CORSMiddleware())
```

---

##### `CORSWithConfig(config CORSConfig) gin.HandlerFunc`
**功能**: 自定义CORS配置

**使用**:
```go
router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST"},
}))
```

---

## 7. 安全中间件

### 7.1 文件: `security.go`

#### 核心函数

##### `SecureHeaders() gin.HandlerFunc`
**功能**: 设置安全响应头

**响应头**:
```go
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
```

**使用**:
```go
router.Use(middleware.SecureHeaders())
```

---

##### `SQLInjectionProtection() gin.HandlerFunc`
**功能**: SQL注入防护

**检测内容**:
- 查询参数
- 请求体
- 路径参数

**危险模式**:
```go
var sqlInjectionPatterns = []string{
    `(\s|^)(union|select|insert|update|delete|drop|create|alter)(\s|$)`,
    `--`,
    `/\*`,
    `\*/`,
    `;`,
    `'`,
    `"`,
}
```

**使用**:
```go
router.Use(middleware.SQLInjectionProtection())
```

---

##### `XSSProtection() gin.HandlerFunc`
**功能**: XSS攻击防护

**清理内容**:
- HTML标签
- JavaScript代码
- 事件处理器

**使用**:
```go
router.Use(middleware.XSSProtection())
```

---

## 8. 其他中间件

### 8.1 文件: `request_id.go`

#### 核心函数

##### `RequestIDMiddleware() gin.HandlerFunc`
**功能**: 为每个请求分配唯一ID

**生成方式**: UUID v4

**设置位置**:
- 上下文: `request_id`
- 响应头: `X-Request-ID`

**使用**:
```go
router.Use(middleware.RequestIDMiddleware())

// 在handler中获取
requestID := c.GetString("request_id")
```

---

##### `GetRequestID(c *gin.Context) string`
**功能**: 获取请求ID

**使用**:
```go
requestID := middleware.GetRequestID(c)
logger.Info("processing request", zap.String("request_id", requestID))
```

---

## 9. 使用示例

### 9.1 完整的中间件栈

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/oblivious/backend/internal/middleware"
    "github.com/oblivious/backend/internal/cache"
)

func setupRouter() *gin.Engine {
    r := gin.New()
    
    // 1. Recovery - 必须第一个
    r.Use(gin.Recovery())
    
    // 2. 请求ID - 用于追踪
    r.Use(middleware.RequestIDMiddleware())
    
    // 3. 日志 - 记录所有请求
    r.Use(middleware.LoggerMiddleware())
    
    // 4. CORS - 跨域处理
    r.Use(middleware.CORSMiddleware())
    
    // 5. 安全头
    r.Use(middleware.SecureHeaders())
    
    // 公开路由
    public := r.Group("/api/v1")
    {
        // 限流: 10 req/min
        public.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
            Rate:  10,
            Burst: 20,
            TTL:   time.Minute,
        }))
        
        public.POST("/register", registerHandler)
        public.POST("/login", loginHandler)
    }
    
    // 需要认证的路由
    protected := r.Group("/api/v1")
    {
        // 认证
        authFactory := middleware.NewAuthFactory([]byte("secret"), redisCache)
        protected.Use(authFactory.JWTWithCache())
        
        // 加载权限
        rbacManager := middleware.NewRBACManager(5 * time.Minute)
        protected.Use(middleware.LoadUserPermissions(rbacManager))
        
        // 限流: 100 req/min
        protected.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
            Rate:  100,
            Burst: 200,
            TTL:   time.Minute,
        }))
        
        // 用户路由
        protected.GET("/profile", getUserProfile)
        protected.PUT("/profile", updateUserProfile)
        
        // 聊天路由
        protected.POST("/chat/messages", sendMessage)
    }
    
    // 管理员路由
    admin := r.Group("/api/v1/admin")
    {
        admin.Use(authFactory.JWT())
        admin.Use(middleware.LoadUserPermissions(rbacManager))
        admin.Use(middleware.RequireRole("admin"))
        
        // 渠道管理
        admin.POST("/channels", 
            middleware.RequirePermission("channel.create"),
            createChannel)
        
        admin.DELETE("/channels/:id", 
            middleware.RequireAllPermissions("channel.delete", "channel.manage"),
            deleteChannel)
        
        // 定价管理
        admin.POST("/pricing",
            middleware.RequirePermission("pricing.edit"),
            createPricing)
    }
    
    return r
}
```

### 9.2 自定义认证逻辑

```go
// 可选认证 - 认证成功则加载用户信息，失败也继续
router.GET("/api/public/data", 
    authFactory.Optional(),
    func(c *gin.Context) {
        userID, _, _, ok := middleware.GetAuthInfo(c)
        if ok {
            // 用户已登录，返回个性化数据
            return personalizedData(c, userID)
        }
        // 用户未登录，返回公开数据
        return publicData(c)
    })
```

### 9.3 动态权限检查

```go
func deleteResourceHandler(c *gin.Context) {
    resourceID := c.Param("id")
    
    // 获取用户权限
    userPerms, _ := c.Get("user_permissions")
    perms := userPerms.(*model.UserPermissions)
    
    // 动态检查权限
    if hasAccess, reason := middleware.CheckResourceAccess(perms, "resource", "delete"); !hasAccess {
        c.JSON(403, gin.H{"error": reason})
        return
    }
    
    // 执行删除逻辑
    ...
}
```

### 9.4 多层级限流

```go
// 全局限流
router.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
    Rate:  1000,  // 每秒1000请求
    Burst: 2000,
    TTL:   time.Second,
}))

// API路由组限流
apiGroup.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
    Rate:  100,  // 每分钟100请求
    Burst: 200,
    TTL:   time.Minute,
}))

// 特定端点限流
apiGroup.POST("/expensive-operation",
    middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
        Rate:  1,   // 每小时1请求
        Burst: 2,
        TTL:   time.Hour,
    }),
    expensiveHandler)
```

---

## 10. 性能优化

### 10.1 认证缓存

使用`CachedAuthMiddleware`减少JWT解析开销：

```go
// 普通认证: 每次都解析JWT
router.Use(middleware.AuthMiddleware(secret))  // ~500µs/request

// 缓存认证: 第一次解析，后续从缓存读取
router.Use(middleware.CachedAuthMiddleware(cache, secret))  // ~50µs/request

// 性能提升: 10倍
```

### 10.2 权限缓存

RBAC权限自动缓存5分钟：

```go
// 首次加载权限: 需要数据库查询 (~10ms)
// 后续请求: 从Redis缓存读取 (~1ms)

rbacManager := middleware.NewRBACManager(5 * time.Minute)
```

### 10.3 限流性能

令牌桶算法使用Redis Lua脚本：

```go
// 原子操作，单次Redis调用
// 性能: ~2ms per request
// 并发安全，支持分布式
```

---

## 11. 故障排查

### 11.1 认证失败

**问题**: 返回401 Unauthorized

**排查步骤**:
1. 检查token格式: `Bearer <token>`
2. 验证token有效期
3. 确认签名密钥正确
4. 检查token是否被修改

```bash
# 解码JWT查看内容
echo "token" | base64 -d
```

### 11.2 权限不足

**问题**: 返回403 Forbidden

**排查步骤**:
1. 检查用户角色
2. 确认权限配置
3. 查看权限缓存是否过期
4. 验证RBAC规则

```go
// 调试权限
roles := middleware.GetUserRoleNames(c)
permissions := middleware.GetUserPermissionNames(c)
log.Printf("User roles: %v, permissions: %v", roles, permissions)
```

### 11.3 限流问题

**问题**: 频繁返回429

**排查步骤**:
1. 检查限流配置是否合理
2. 确认Redis连接正常
3. 查看是否有恶意请求
4. 考虑调整Rate和Burst值

```bash
# 查看限流键
redis-cli KEYS "rate_limit:*"

# 查看特定用户的限流状态
redis-cli HGETALL "rate_limit:user:123"
```

---

## 12. 最佳实践

### 12.1 中间件顺序

遵循以下顺序：
1. Recovery（必须第一个）
2. RequestID
3. Logger
4. CORS
5. Security
6. Auth
7. LoadPermissions
8. RateLimit
9. 业务中间件

### 12.2 性能考虑

- ✅ 使用缓存认证降低开销
- ✅ 权限缓存5-10分钟
- ✅ 限流使用Redis集群
- ❌ 避免在中间件中进行复杂计算
- ❌ 避免同步IO操作

### 12.3 安全建议

- ✅ 所有敏感端点使用认证
- ✅ 重要操作添加权限检查
- ✅ 启用CORS和安全头
- ✅ 实施多层级限流
- ❌ 不要在日志中记录敏感信息

---

## 结语

本文档详细说明了Oblivious平台的所有中间件实现。通过合理使用这些中间件，可以构建安全、高效、可维护的API服务。

**下一步**: 查看 [SERVICE_GUIDE.md](./SERVICE_GUIDE.md) 了解业务服务层的详细实现。

---

📅 最后更新: 2025-11-22  
📝 文档版本: v1.0.0  
👥 维护者: Oblivious Team

