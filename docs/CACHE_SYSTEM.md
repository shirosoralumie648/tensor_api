# 缓存系统 (Cache System)

## 文件位置
- `backend/internal/cache/cache.go` - 缓存接口
- `backend/internal/cache/redis.go` - Redis实现
- `backend/internal/cache/user_cache.go` - 用户缓存

---

## 1. Cache 接口

```go
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

---

## 2. RedisClient 结构体

```go
type RedisClient struct {
    mu        sync.RWMutex
    data      map[string]*CacheEntry  // 缓存数据
    ttls      map[string]time.Time    // 过期时间
    cluster   bool                    // 集群模式
    password  string                  // 密码
    addresses []string                // 地址列表
    db        int                     // 数据库编号
    stats     *CacheStats            // 统计信息
}
```

---

## 3. RedisClient 方法

### `NewRedisClient(cfg *CacheConfig) (*RedisClient, error)`
**功能**: 创建Redis客户端  
**输入**: 
- `cfg *CacheConfig` - 配置
  - `Addrs []string` - 地址列表（必需）
  - `Password string` - 密码
  - `DB int` - 数据库编号
  - `PoolSize int` - 连接池大小
  - `MaxRetries int` - 最大重试次数
  - `TTL time.Duration` - 默认TTL
  - `ClusterMode bool` - 集群模式  
**输出**: 
- `*RedisClient` - Redis客户端
- `error` - 创建错误

**自动启动**: 过期清理goroutine

---

### `Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error`
**功能**: 设置缓存值  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键
- `value interface{}` - 值
- `ttl time.Duration` - 过期时间（0表示不过期）  
**输出**: 
- `error` - 设置错误

**统计**: 更新`Sets`计数

---

### `Get(ctx context.Context, key string) (interface{}, error)`
**功能**: 获取缓存值  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键  
**输出**: 
- `interface{}` - 值
- `error` - 获取错误（键不存在或已过期）

**统计**: 
- 命中: 更新`Hits`计数
- 未命中: 更新`Misses`计数

---

### `Delete(ctx context.Context, key string) error`
**功能**: 删除缓存值  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键  
**输出**: 
- `error` - 删除错误

**统计**: 更新`Deletes`计数

---

### `Exists(ctx context.Context, key string) bool`
**功能**: 检查键是否存在  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键  
**输出**: 
- `bool` - 是否存在

---

### `Expire(ctx context.Context, key string, ttl time.Duration) error`
**功能**: 设置过期时间  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键
- `ttl time.Duration` - 过期时间  
**输出**: 
- `error` - 设置错误（键不存在）

---

### `TTL(ctx context.Context, key string) (time.Duration, error)`
**功能**: 获取剩余TTL  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键  
**输出**: 
- `time.Duration` - 剩余时间
- `error` - 获取错误（键不存在）

---

### `Incr(ctx context.Context, key string, delta int64) (int64, error)`
**功能**: 原子增量  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键
- `delta int64` - 增量  
**输出**: 
- `int64` - 新值
- `error` - 操作错误

---

### `Decr(ctx context.Context, key string, delta int64) (int64, error)`
**功能**: 原子减量  
**输入**: 
- `ctx context.Context` - 上下文
- `key string` - 键
- `delta int64` - 减量  
**输出**: 
- `int64` - 新值
- `error` - 操作错误

**实现**: 调用`Incr(ctx, key, -delta)`

---

### `MGet(ctx context.Context, keys ...string) map[string]interface{}`
**功能**: 批量获取  
**输入**: 
- `ctx context.Context` - 上下文
- `keys ...string` - 键列表  
**输出**: 
- `map[string]interface{}` - 键值映射（只包含存在的键）

---

### `MSet(ctx context.Context, kvs map[string]interface{}) error`
**功能**: 批量设置  
**输入**: 
- `ctx context.Context` - 上下文
- `kvs map[string]interface{}` - 键值映射  
**输出**: 
- `error` - 设置错误

**统计**: 更新`Sets`计数

---

### `Del(ctx context.Context, keys ...string) int64`
**功能**: 删除多个键  
**输入**: 
- `ctx context.Context` - 上下文
- `keys ...string` - 键列表  
**输出**: 
- `int64` - 删除的数量

---

### `Clear(ctx context.Context) error`
**功能**: 清空所有缓存  
**输入**: 
- `ctx context.Context` - 上下文  
**输出**: 
- `error` - 清空错误

---

### `Keys(ctx context.Context, pattern string) []string`
**功能**: 获取所有键（支持模式匹配）  
**输入**: 
- `ctx context.Context` - 上下文
- `pattern string` - 匹配模式  
**输出**: 
- `[]string` - 键列表

---

### `GetStats() map[string]interface{}`
**功能**: 获取统计信息  
**输入**: 无  
**输出**: 
- `map[string]interface{}` - 统计信息
  - `hits` - 命中次数
  - `misses` - 未命中次数
  - `hit_rate` - 命中率（%）
  - `sets` - 设置次数
  - `deletes` - 删除次数
  - `expirations` - 过期次数
  - `cache_size` - 缓存大小

---

### `Close() error`
**功能**: 关闭连接  
**输入**: 无  
**输出**: 
- `error` - 关闭错误

---

## 4. CacheEntry 结构体

```go
type CacheEntry struct {
    Key       string        // 键
    Value     interface{}   // 值
    TTL       time.Duration // 过期时间
    CreatedAt time.Time    // 创建时间
    UpdatedAt time.Time    // 更新时间
    HitCount  int64        // 命中次数
}
```

---

## 5. CacheStats 结构体

```go
type CacheStats struct {
    Hits         int64  // 命中次数
    Misses       int64  // 未命中次数
    Sets         int64  // 设置次数
    Deletes      int64  // 删除次数
    Expirations  int64  // 过期次数
    EvictionSize int64  // 驱逐大小
}
```

---

## 6. Pipeline 管道操作

### `NewPipeline() *Pipeline`
**功能**: 创建管道  
**输入**: 无  
**输出**: `*Pipeline`

---

### `Set(key string, value interface{}, ttl time.Duration) *Pipeline`
**功能**: 添加Set命令到管道  
**输入**: 
- `key string` - 键
- `value interface{}` - 值
- `ttl time.Duration` - 过期时间  
**输出**: 
- `*Pipeline` - 管道对象（链式调用）

---

### `Get(key string) *Pipeline`
**功能**: 添加Get命令到管道  
**输入**: 
- `key string` - 键  
**输出**: 
- `*Pipeline` - 管道对象

---

### `Execute(ctx context.Context) error`
**功能**: 执行管道中的所有命令  
**输入**: 
- `ctx context.Context` - 上下文  
**输出**: 
- `error` - 执行错误

---

## 7. 过期清理

### `cleanupExpired()`
**功能**: 自动清理过期项（后台goroutine）  
**输入**: 无  
**输出**: 无

**逻辑**:
- 每分钟执行一次
- 检查所有键的过期时间
- 删除已过期的键
- 更新过期统计

---

## 8. 用户缓存

### UserCache
**文件**: `backend/internal/cache/user_cache.go`

**功能**: 用户相关的缓存操作

**方法**:
- `GetUser(userID int)` - 获取用户缓存
- `SetUser(user *model.User, ttl time.Duration)` - 设置用户缓存
- `InvalidateUser(userID int)` - 失效用户缓存

**缓存键格式**: `user:{userID}`

---

## 使用示例

```go
// 创建Redis客户端
cfg := &cache.CacheConfig{
    Addrs: []string{"localhost:6379"},
    Password: "",
    DB: 0,
    ClusterMode: false,
}
client, err := cache.NewRedisClient(cfg)

// 设置缓存
err = client.Set(ctx, "key1", "value1", 5*time.Minute)

// 获取缓存
value, err := client.Get(ctx, "key1")

// 批量操作
values := client.MGet(ctx, "key1", "key2", "key3")
err = client.MSet(ctx, map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
})

// 原子操作
newValue, err := client.Incr(ctx, "counter", 1)

// 管道操作
pipeline := client.NewPipeline()
pipeline.Set("key1", "value1", time.Minute)
pipeline.Get("key2")
err = pipeline.Execute(ctx)

// 获取统计
stats := client.GetStats()
fmt.Printf("命中率: %.2f%%\n", stats["hit_rate"])
```

---

📅 最后更新: 2025-11-22

