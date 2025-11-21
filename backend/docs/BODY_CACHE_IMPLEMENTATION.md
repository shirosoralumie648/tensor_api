# 请求体缓存与恢复实现文档

## 概述

请求体缓存与恢复系统是 Oblivious AI 平台的核心组件，用于处理任意大小的 HTTP 请求体，支持内存、磁盘混合存储，以及智能恢复机制。

**核心特性：**
- ✅ 支持任意大小请求体（仅受磁盘限制）
- ✅ 内存 + 磁盘混合存储（自适应）
- ✅ 多种恢复策略（重试、备用渠道、分块）
- ✅ 自动清理机制（LRU + 过期删除）
- ✅ 流式读写支持
- ✅ 完整的统计和监控

## 架构设计

### 核心组件

#### 1. BodyCache（请求体缓存）

负责请求体的存储和检索。

```
┌─────────────────────────────────────────┐
│         BodyCache Manager               │
├─────────────────────────────────────────┤
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  Memory Cache (sync.Map)         │  │
│  │  - 小文件（<1MB）                 │  │
│  │  - 快速访问                      │  │
│  └──────────────────────────────────┘  │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  Disk Cache (Temp Files)         │  │
│  │  - 大文件（>1MB）                 │  │
│  │  - 持久存储                      │  │
│  └──────────────────────────────────┘  │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │  Cleanup Routine                 │  │
│  │  - 过期删除（24小时）             │  │
│  │  - LRU 清理（>10GB）              │  │
│  └──────────────────────────────────┘  │
│                                         │
└─────────────────────────────────────────┘
```

#### 2. BodyRecoveryManager（请求体恢复）

负责失败请求的恢复。

```
┌──────────────────────────────────────────┐
│    BodyRecoveryManager                   │
├──────────────────────────────────────────┤
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Retry Strategy                    │ │
│  │  - 指数退避 (3次)                  │ │
│  │  - 自动渠道切换                    │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Alternate Channel Strategy        │ │
│  │  - 使用备用 API 渠道                │ │
│  │  - 自动故障转移                    │ │
│  └────────────────────────────────────┘ │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │  Chunked Strategy                  │ │
│  │  - 分块上传 (512KB/块)              │ │
│  │  - 支持断点续传                    │ │
│  └────────────────────────────────────┘ │
│                                          │
└──────────────────────────────────────────┘
```

## 使用指南

### 基础用法

#### 1. 初始化缓存管理器

```go
// 创建缓存管理器
cache := relay.NewBodyCache("/tmp/request-cache")

// 配置参数
cache.SetMode(relay.BodyCacheModeHybrid)    // 混合模式
cache.SetMemoryThreshold(1024 * 1024)       // 1MB 阈值

// 启动清理线程
cache.Start()
defer cache.Stop()
```

#### 2. 缓存请求体

```go
// 从 HTTP 请求读取并缓存
cacheID, err := cache.CacheRequestBody(request.Body)
if err != nil {
    log.Fatalf("Cache failed: %v", err)
}

log.Printf("Cached as: %s", cacheID)
```

#### 3. 检索请求体

```go
// 获取缓存的请求体
body, err := cache.GetCachedBody(cacheID)
if err != nil {
    log.Fatalf("Get failed: %v", err)
}

// 使用缓存数据
log.Printf("Retrieved %d bytes", len(body))
```

#### 4. 流式读取

```go
// 获取读取器用于流式处理
reader, err := cache.GetCachedBodyReader(cacheID)
if err != nil {
    log.Fatalf("Get reader failed: %v", err)
}

// 流式读取
data, err := io.ReadAll(reader)
if err != nil {
    log.Fatalf("Read failed: %v", err)
}
```

### 高级用法

#### 1. 请求体恢复

```go
// 创建恢复管理器
recovery := relay.NewBodyRecoveryManager(cache)

// 配置恢复策略
recovery.SetStrategy(relay.RecoveryStrategyChunked)
recovery.SetChunkSize(512 * 1024)           // 512KB 分块
recovery.SetMaxRecoveries(3)                // 最多 3 次恢复

// 启动恢复过程
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := recovery.InitiateRecovery(ctx, "request-123", cacheID)
if err != nil {
    log.Fatalf("Recovery failed: %v", err)
}
```

#### 2. 可恢复流读取器

```go
// 创建可恢复的流读取器
streamReader := relay.NewBodyRecoveryStreamReader(
    originalReader,
    cache,
    recovery,
    "request-123",
)

// 读取数据（自动缓存）
buf := make([]byte, 4096)
n, err := streamReader.Read(buf)
if err != nil && err != io.EOF {
    log.Fatalf("Read failed: %v", err)
}

// 获取缓存 ID
cacheID := streamReader.GetCacheID()
log.Printf("Stream cached as: %s", cacheID)
```

#### 3. 统计和监控

```go
// 获取缓存统计
stats := cache.GetStatistics()
log.Printf("Cache stats: %+v", stats)

// 输出:
// Cache stats: map[string]interface{}{
//     "memory_count": 10,
//     "disk_count": 5,
//     "total_size": 52428800,        // 50MB
//     "total_hits": 1000,
//     "total_misses": 50,
//     "hit_rate": 95.23,
//     "evictions": 2,
// }

// 获取恢复统计
recStats := recovery.GetStatistics()
log.Printf("Recovery stats: %+v", recStats)

// 输出:
// Recovery stats: map[string]interface{}{
//     "total_recoveries": 100,
//     "successful_recoveries": 98,
//     "failed_recoveries": 2,
//     "success_rate": 98.0,
//     "total_bytes_recovered": 1048576,
//     "active_recoveries": 5,
// }
```

## 缓存模式详解

### 1. 内存缓存模式 (BodyCacheModeMemory)

**适用场景：** 小型、频繁访问的请求

```
请求体 → 读取全部 → 存入内存 → 快速检索
```

**优点：**
- 速度快（纳秒级）
- 无 I/O 开销

**缺点：**
- 内存占用大
- 重启丢失

**配置：**
```go
cache.SetMode(relay.BodyCacheModeMemory)
```

### 2. 磁盘缓存模式 (BodyCacheModeDisk)

**适用场景：** 大型、非频繁访问的请求

```
请求体 → 读取全部 → 写入磁盘 → 需要时读取
```

**优点：**
- 支持任意大小
- 持久存储
- 重启不丢失

**缺点：**
- I/O 延迟较高
- 磁盘空间消耗

**配置：**
```go
cache.SetMode(relay.BodyCacheModeDisk)
```

### 3. 混合模式 (BodyCacheModeHybrid) - 推荐

**适用场景：** 生产环境（平衡速度和存储）

```
         ↓ 请求体
     是否 > 1MB?
    ↙            ↘
  NO             YES
  ↓              ↓
内存缓存        磁盘缓存
  ↓              ↓
快速访问        持久存储
```

**配置：**
```go
cache.SetMode(relay.BodyCacheModeHybrid)
cache.SetMemoryThreshold(1024 * 1024)  // 1MB
```

## 恢复策略

### 1. 重试策略 (RecoveryStrategyRetry)

使用指数退避重新发送请求。

```
发送请求 → 失败
    ↓
等待 100ms → 重试 1
    ↓
等待 200ms → 重试 2
    ↓
等待 400ms → 重试 3
    ↓
放弃/超时
```

**何时使用：** 网络瞬间故障、服务器临时不可用

**配置：**
```go
recovery.SetStrategy(relay.RecoveryStrategyRetry)
recovery.SetMaxRecoveries(3)
```

### 2. 备用渠道策略 (RecoveryStrategyAlternateChannel)

切换到备用 API 渠道重新发送。

```
主渠道 → 失败
    ↓
备用渠道 1 → 成功
    或失败 ↓
备用渠道 2 → ...
```

**何时使用：** 主渠道服务降级、限流

**配置：**
```go
recovery.SetStrategy(relay.RecoveryStrategyAlternateChannel)
```

### 3. 分块策略 (RecoveryStrategyChunked)

将大请求体分块上传，支持断点续传。

```
大请求体 (100MB)
    ↓
分块 (512KB × 200)
    ↓
逐块上传
    ↓
失败的块重新上传
    ↓
合并完成
```

**何时使用：** 处理大文件、不稳定网络

**配置：**
```go
recovery.SetStrategy(relay.RecoveryStrategyChunked)
recovery.SetChunkSize(512 * 1024)  // 512KB
```

## 性能特性

### 缓存性能

| 操作 | 内存模式 | 磁盘模式 | 混合模式 |
|------|---------|---------|---------|
| 小文件缓存 | <1ms | ~5-10ms | <1ms |
| 小文件检索 | <0.1ms | ~2-5ms | <0.1ms |
| 大文件缓存 | N/A | ~100ms | ~100ms |
| 大文件检索 | N/A | ~50-100ms | ~50-100ms |

### 恢复性能

| 策略 | 首次延迟 | 渠道切换时间 | 成功率 |
|------|---------|-----------|--------|
| 重试 | 100ms | N/A | >95% |
| 备用渠道 | <5ms | <10ms | >98% |
| 分块 | <5ms | N/A | 99%+ |

### 清理性能

- **清理间隔：** 5 分钟
- **过期检查：** O(n) 遍历
- **LRU 清理：** 触发时 O(n log n)
- **清理吞吐量：** >1GB/分钟

## 清理机制

### 自动清理

系统每 5 分钟执行一次清理：

```
触发清理
    ↓
1. 检查过期缓存 (>24小时)
    ↓
2. 删除过期缓存
    ↓
3. 如果总大小 > 10GB
    ↓
4. 按创建时间排序（LRU）
    ↓
5. 删除最旧的缓存直到 <= 8GB
```

### 手动清理

```go
// 失效单个缓存
cache.InvalidateCache(cacheID)

// 清空所有缓存
cache.PurgeAll()
```

## 配置建议

### 开发环境

```go
cache := relay.NewBodyCache("/tmp/request-cache")
cache.SetMode(relay.BodyCacheModeMemory)
cache.SetMemoryThreshold(10 * 1024 * 1024)  // 10MB
```

### 生产环境 - 小规模

```go
cache := relay.NewBodyCache("/var/cache/oblivious/request-cache")
cache.SetMode(relay.BodyCacheModeHybrid)
cache.SetMemoryThreshold(1 * 1024 * 1024)  // 1MB
// 默认: 最大 10GB, 过期 24 小时
```

### 生产环境 - 大规模

```go
cache := relay.NewBodyCache("/data/cache/request-cache")
cache.SetMode(relay.BodyCacheModeHybrid)
cache.SetMemoryThreshold(512 * 1024)       // 512KB
cache.maxCacheSize = 100 * 1024 * 1024 * 1024  // 100GB
cache.maxCacheDuration = 48 * time.Hour
```

## 监控和诊断

### 关键指标

```go
stats := cache.GetStatistics()

// 缓存命中率 (目标: >90%)
hitRate := stats["hit_rate"].(float64)

// 总缓存大小 (监控磁盘占用)
totalSize := stats["total_size"].(int64)

// 清理次数 (异常高表示配置不当)
evictions := stats["evictions"].(int64)
```

### 告警规则

- **缓存命中率 <85%：** 阈值配置可能不当
- **缓存大小 >90% maxCacheSize：** 需要清理或扩容
- **清理频繁 (>5次/小时)：** 缓存容量不足
- **恢复失败率 >5%：** 网络或服务故障

### 日志示例

```
[info] recovery with chunked strategy requestID=req-123 size=104857600 chunks=200
[info] sending chunk requestID=req-123 chunk=1 of=200 size=524288
[info] sending chunk requestID=req-123 chunk=2 of=200 size=524288
...
[info] recovery succeeded requestID=req-123 size=104857600
```

## 与其他组件的集成

### 与 RequestClient 集成

```go
// 创建请求客户端
client := relay.NewRequestClient(30 * time.Second)

// 集成缓存和恢复
cache := relay.NewBodyCache("/tmp/cache")
recovery := relay.NewBodyRecoveryManager(cache)

client.cache = cache
client.recovery = recovery

// 现在请求会自动缓存和恢复
```

## 最佳实践

1. **合理设置阈值**
   - 根据内存容量选择内存阈值
   - 不要设置过小（频繁磁盘 I/O）
   - 不要设置过大（内存溢出）

2. **定期监控**
   - 每小时检查缓存统计
   - 监控命中率趋势
   - 监控磁盘空间

3. **及时清理**
   - 定期手动清理过期数据
   - 设置合理的过期时间
   - 监控清理频率

4. **恢复策略选择**
   - 小文件用重试策略
   - 大文件用分块策略
   - 关键业务用备用渠道

5. **错误处理**
   - 缓存失败不应导致请求失败
   - 恢复失败需要重新发送
   - 记录所有失败日志

## 常见问题

### Q: 如何处理超大文件（>1GB）？

A: 使用分块策略，将大文件分成多个小块分别上传和缓存。

### Q: 缓存会占满磁盘怎么办？

A: 系统会自动触发 LRU 清理，保持在最大容量的 80%。

### Q: 如何验证缓存的完整性？

A: 系统自动计算 MD5 哈希值，可以在恢复时验证。

### Q: 内存缓存和磁盘缓存能并存吗？

A: 可以，混合模式会智能选择存储方式。

### Q: 如何处理缓存热点问题？

A: 监控缓存统计，识别频繁访问的缓存，考虑优化应用逻辑。

## 参考资源

- [RequestClient 文档](./RETRY_MECHANISM_IMPLEMENTATION.md)
- [SSE 流式文档](./SSE_STREAM_IMPLEMENTATION.md)
- [项目状态](./PHASE1_PROGRESS.md)

