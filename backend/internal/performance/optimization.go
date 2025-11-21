package performance

import (
	"runtime"
	"sync"
	"time"
)

// PerformanceOptimizer 性能优化器
type PerformanceOptimizer struct {
	mu sync.RWMutex

	// 数据库连接池
	dbPoolSize     int
	dbMaxOverflow  int

	// Goroutine 池
	goroutineLimit int

	// 缓存配置
	cacheMaxSize int
	cacheTTL     time.Duration

	// 内存指标
	lastMemStats runtime.MemStats
	updateTime   time.Time
}

// NewPerformanceOptimizer 创建新的性能优化器
func NewPerformanceOptimizer() *PerformanceOptimizer {
	return &PerformanceOptimizer{
		dbPoolSize:     20,
		dbMaxOverflow:  40,
		goroutineLimit: 1000,
		cacheMaxSize:   100000,
		cacheTTL:       5 * time.Minute,
		updateTime:     time.Now(),
	}
}

// GetMemoryStats 获取内存统计
func (po *PerformanceOptimizer) GetMemoryStats() runtime.MemStats {
	po.mu.Lock()
	defer po.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	po.lastMemStats = m
	po.updateTime = time.Now()
	return m
}

// GetGoroutineCount 获取 Goroutine 数量
func (po *PerformanceOptimizer) GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// GetCPUStats 获取 CPU 统计
func (po *PerformanceOptimizer) GetCPUStats() map[string]interface{} {
	return map[string]interface{}{
		"num_cpu":           runtime.NumCPU(),
		"goroutine_count":   runtime.NumGoroutine(),
		"num_goroutine_max": po.goroutineLimit,
	}
}

// OptimizeMemory 优化内存使用
func (po *PerformanceOptimizer) OptimizeMemory() {
	// 强制垃圾回收
	runtime.GC()

	// 释放内存
	debug := runtime.DebugInfo()
	if len(debug) > 0 {
		// 强制进行一次垃圾回收
		runtime.GC()
	}
}

// CheckMemoryUsage 检查内存使用
func (po *PerformanceOptimizer) CheckMemoryUsage() (used, limit uint64, percentage float64) {
	stats := po.GetMemoryStats()

	// RSS 内存 (实际使用)
	used = stats.Alloc

	// 总分配内存
	limit = stats.TotalAlloc

	// 内存使用百分比
	if limit > 0 {
		percentage = float64(used) / float64(limit) * 100
	}

	return used, limit, percentage
}

// DatabaseOptimization 数据库优化配置
type DatabaseOptimization struct {
	PoolSize      int           // 连接池大小
	MaxOverflow   int           // 最大溢出连接数
	MaxIdleTime   time.Duration // 最大空闲时间
	MaxLifetime   time.Duration // 最大生命周期
	ReadTimeout   time.Duration // 读超时
	WriteTimeout  time.Duration // 写超时
	IdleConns     int           // 空闲连接数
	OpenConns     int           // 打开连接数
}

// NewDatabaseOptimization 创建数据库优化配置
func NewDatabaseOptimization() *DatabaseOptimization {
	return &DatabaseOptimization{
		PoolSize:     20,
		MaxOverflow:  40,
		MaxIdleTime:  5 * time.Minute,
		MaxLifetime:  30 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleConns:    5,
	}
}

// CacheOptimization 缓存优化配置
type CacheOptimization struct {
	MaxSize           int           // 最大缓存条数
	TTL               time.Duration // 过期时间
	EvictionPolicy    string        // 驱逐策略 (LRU/LFU)
	CompressionRatio  float64       // 压缩率
	PrewarmPercentage float64       // 预热百分比
}

// NewCacheOptimization 创建缓存优化配置
func NewCacheOptimization() *CacheOptimization {
	return &CacheOptimization{
		MaxSize:           100000,
		TTL:               5 * time.Minute,
		EvictionPolicy:    "LRU",
		CompressionRatio:  0.8,
		PrewarmPercentage: 0.2,
	}
}

// ConcurrencyOptimization 并发优化配置
type ConcurrencyOptimization struct {
	MaxGoroutines      int           // 最大 Goroutine 数
	WorkerPoolSize     int           // 工作池大小
	QueueSize          int           // 队列大小
	RequestTimeout     time.Duration // 请求超时
	ContextDeadline    time.Duration // Context 超时
	ShutdownTimeout    time.Duration // 关闭超时
	RateLimitPerSecond int           // 每秒速率限制
}

// NewConcurrencyOptimization 创建并发优化配置
func NewConcurrencyOptimization() *ConcurrencyOptimization {
	return &ConcurrencyOptimization{
		MaxGoroutines:      10000,
		WorkerPoolSize:     100,
		QueueSize:          1000,
		RequestTimeout:     30 * time.Second,
		ContextDeadline:    60 * time.Second,
		ShutdownTimeout:    10 * time.Second,
		RateLimitPerSecond: 10000,
	}
}

// NetworkOptimization 网络优化配置
type NetworkOptimization struct {
	EnableHTTP2         bool          // 启用 HTTP/2
	EnableKeepAlive     bool          // 启用 Keep-Alive
	MaxIdleConns        int           // 最大空闲连接数
	MaxIdleConnsPerHost int           // 每个主机最大空闲连接数
	IdleConnTimeout     time.Duration // 空闲连接超时
	DialTimeout         time.Duration // 拨号超时
	DialKeepAlive       time.Duration // 拨号 Keep-Alive
	CompressResponse    bool          // 压缩响应
	CompressionLevel    int           // 压缩级别
}

// NewNetworkOptimization 创建网络优化配置
func NewNetworkOptimization() *NetworkOptimization {
	return &NetworkOptimization{
		EnableHTTP2:         true,
		EnableKeepAlive:     true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DialTimeout:         5 * time.Second,
		DialKeepAlive:       30 * time.Second,
		CompressResponse:    true,
		CompressionLevel:    6,
	}
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	mu              sync.RWMutex
	requestCount    int64
	requestDuration time.Duration
	errorCount      int64
	cacheHits       int64
	cacheMisses     int64
}

// RecordRequest 记录请求
func (pm *PerformanceMonitor) RecordRequest(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.requestCount++
	pm.requestDuration += duration
}

// RecordError 记录错误
func (pm *PerformanceMonitor) RecordError() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.errorCount++
}

// RecordCacheHit 记录缓存命中
func (pm *PerformanceMonitor) RecordCacheHit() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.cacheHits++
}

// RecordCacheMiss 记录缓存未命中
func (pm *PerformanceMonitor) RecordCacheMiss() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.cacheMisses++
}

// GetStats 获取统计信息
func (pm *PerformanceMonitor) GetStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	avgDuration := time.Duration(0)
	if pm.requestCount > 0 {
		avgDuration = pm.requestDuration / time.Duration(pm.requestCount)
	}

	cacheHitRate := float64(0)
	total := pm.cacheHits + pm.cacheMisses
	if total > 0 {
		cacheHitRate = float64(pm.cacheHits) / float64(total) * 100
	}

	errorRate := float64(0)
	if pm.requestCount > 0 {
		errorRate = float64(pm.errorCount) / float64(pm.requestCount) * 100
	}

	return map[string]interface{}{
		"request_count":      pm.requestCount,
		"avg_duration_ms":    avgDuration.Milliseconds(),
		"error_count":        pm.errorCount,
		"error_rate_percent": errorRate,
		"cache_hits":         pm.cacheHits,
		"cache_misses":       pm.cacheMisses,
		"cache_hit_rate_percent": cacheHitRate,
	}
}

