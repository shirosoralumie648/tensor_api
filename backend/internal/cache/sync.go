package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SyncStrategy 缓存同步策略
type SyncStrategy interface {
	Sync(ctx context.Context, key string, value interface{}) error
	Invalidate(ctx context.Context, key string) error
}

// WriteThroughCache 写穿策略
type WriteThroughCache struct {
	cache         *RedisClient
	persistence   func(context.Context, string, interface{}) error
	mu            sync.RWMutex
}

// NewWriteThroughCache 创建写穿缓存
func NewWriteThroughCache(cache *RedisClient, persist func(context.Context, string, interface{}) error) *WriteThroughCache {
	return &WriteThroughCache{
		cache:       cache,
		persistence: persist,
	}
}

// Set 写穿设置
func (wtc *WriteThroughCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	wtc.mu.Lock()
	defer wtc.mu.Unlock()

	// 先写入持久化层
	if err := wtc.persistence(ctx, key, value); err != nil {
		return fmt.Errorf("persistence failed: %w", err)
	}

	// 再更新缓存
	return wtc.cache.Set(ctx, key, value, ttl)
}

// Get 写穿获取
func (wtc *WriteThroughCache) Get(ctx context.Context, key string) (interface{}, error) {
	wtc.mu.RLock()
	defer wtc.mu.RUnlock()

	// 优先从缓存获取
	if val, err := wtc.cache.Get(ctx, key); err == nil {
		return val, nil
	}

	// 缓存未命中，触发回源（实际应该实现）
	return nil, fmt.Errorf("cache miss")
}

// WriteBackCache 写回策略
type WriteBackCache struct {
	cache         *RedisClient
	persistence   func(context.Context, string, interface{}) error
	flushInterval time.Duration
	flushBatch    int
	mu            sync.RWMutex
	dirty         map[string]interface{}
	ticker        *time.Ticker
}

// NewWriteBackCache 创建写回缓存
func NewWriteBackCache(cache *RedisClient, persist func(context.Context, string, interface{}) error) *WriteBackCache {
	wbc := &WriteBackCache{
		cache:         cache,
		persistence:   persist,
		flushInterval: 5 * time.Second,
		flushBatch:    100,
		dirty:         make(map[string]interface{}),
	}

	// 启动后台刷新
	go wbc.startFlusher()

	return wbc
}

// Set 写回设置
func (wbc *WriteBackCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	wbc.mu.Lock()
	defer wbc.mu.Unlock()

	// 先更新缓存
	if err := wbc.cache.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// 标记为脏数据
	wbc.dirty[key] = value

	return nil
}

// Get 写回获取
func (wbc *WriteBackCache) Get(ctx context.Context, key string) (interface{}, error) {
	wbc.mu.RLock()
	defer wbc.mu.RUnlock()

	return wbc.cache.Get(ctx, key)
}

// Flush 刷新脏数据
func (wbc *WriteBackCache) Flush(ctx context.Context) error {
	wbc.mu.Lock()
	defer wbc.mu.Unlock()

	count := 0
	for key, value := range wbc.dirty {
		if err := wbc.persistence(ctx, key, value); err != nil {
			continue
		}

		delete(wbc.dirty, key)
		count++

		if count >= wbc.flushBatch {
			break
		}
	}

	return nil
}

// startFlusher 启动后台刷新
func (wbc *WriteBackCache) startFlusher() {
	wbc.ticker = time.NewTicker(wbc.flushInterval)
	defer wbc.ticker.Stop()

	ctx := context.Background()

	for range wbc.ticker.C {
		wbc.Flush(ctx)
	}
}

// Close 关闭写回缓存
func (wbc *WriteBackCache) Close() {
	if wbc.ticker != nil {
		wbc.ticker.Stop()
	}

	ctx := context.Background()
	wbc.Flush(ctx)
}

// CacheConsistency 缓存一致性管理
type CacheConsistency struct {
	cache       *RedisClient
	invalidator *CacheInvalidator
	versioning  map[string]int64
	mu          sync.RWMutex
}

// NewCacheConsistency 创建缓存一致性管理
func NewCacheConsistency(cache *RedisClient) *CacheConsistency {
	return &CacheConsistency{
		cache:       cache,
		invalidator: NewCacheInvalidator(cache),
		versioning:  make(map[string]int64),
	}
}

// SetWithVersion 设置带版本的缓存
func (cc *CacheConsistency) SetWithVersion(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// 递增版本号
	cc.versioning[key]++
	version := cc.versioning[key]

	// 使用版本号作为缓存键后缀
	versionedKey := fmt.Sprintf("%s:v%d", key, version)

	return cc.cache.Set(ctx, versionedKey, value, ttl)
}

// GetWithVersion 获取带版本的缓存
func (cc *CacheConsistency) GetWithVersion(ctx context.Context, key string) (interface{}, error) {
	cc.mu.RLock()
	version := cc.versioning[key]
	cc.mu.RUnlock()

	if version == 0 {
		return nil, fmt.Errorf("key not found")
	}

	versionedKey := fmt.Sprintf("%s:v%d", key, version)
	return cc.cache.Get(ctx, versionedKey)
}

// Invalidate 失效所有版本
func (cc *CacheConsistency) Invalidate(ctx context.Context, key string) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	for i := int64(1); i <= cc.versioning[key]; i++ {
		versionedKey := fmt.Sprintf("%s:v%d", key, i)
		cc.cache.Delete(ctx, versionedKey)
	}

	cc.versioning[key] = 0

	return nil
}

// CacheWarmup 缓存预热
type CacheWarmup struct {
	cache      *RedisClient
	sources    map[string]DataSource
	mu         sync.RWMutex
	stats      *WarmupStats
}

// DataSource 数据源
type DataSource interface {
	Load(ctx context.Context) (map[string]interface{}, error)
	Key() string
}

// WarmupStats 预热统计
type WarmupStats struct {
	mu        sync.RWMutex
	Loaded    int64
	Failed    int64
	Duration  time.Duration
}

// NewCacheWarmup 创建缓存预热
func NewCacheWarmup(cache *RedisClient) *CacheWarmup {
	return &CacheWarmup{
		cache:   cache,
		sources: make(map[string]DataSource),
		stats:   &WarmupStats{},
	}
}

// Register 注册数据源
func (cw *CacheWarmup) Register(source DataSource) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	cw.sources[source.Key()] = source
}

// Warmup 执行预热
func (cw *CacheWarmup) Warmup(ctx context.Context, ttl time.Duration) error {
	startTime := time.Now()

	cw.mu.RLock()
	sources := cw.sources
	cw.mu.RUnlock()

	for _, source := range sources {
		data, err := source.Load(ctx)
		if err != nil {
			cw.stats.mu.Lock()
			cw.stats.Failed++
			cw.stats.mu.Unlock()
			continue
		}

		for key, value := range data {
			cw.cache.Set(ctx, key, value, ttl)
			cw.stats.mu.Lock()
			cw.stats.Loaded++
			cw.stats.mu.Unlock()
		}
	}

	cw.stats.mu.Lock()
	cw.stats.Duration = time.Since(startTime)
	cw.stats.mu.Unlock()

	return nil
}

// GetStats 获取预热统计
func (cw *CacheWarmup) GetStats() map[string]interface{} {
	cw.stats.mu.RLock()
	defer cw.stats.mu.RUnlock()

	return map[string]interface{}{
		"loaded":   cw.stats.Loaded,
		"failed":   cw.stats.Failed,
		"duration": cw.stats.Duration,
	}
}

// CacheEviction 缓存驱逐策略
type CacheEviction struct {
	cache       *RedisClient
	maxMemory   int64
	evictPolicy string // LRU, LFU, FIFO
	mu          sync.RWMutex
}

// NewCacheEviction 创建缓存驱逐
func NewCacheEviction(cache *RedisClient, maxMemory int64, policy string) *CacheEviction {
	return &CacheEviction{
		cache:       cache,
		maxMemory:   maxMemory,
		evictPolicy: policy,
	}
}

// Evict 执行驱逐
func (ce *CacheEviction) Evict(ctx context.Context) error {
	stats := ce.cache.GetStats()
	cacheSize := stats["cache_size"].(int)

	if int64(cacheSize) > ce.maxMemory {
		// 简单的 FIFO 驱逐（实际应该实现更复杂的算法）
		keys := ce.cache.Keys(ctx, "*")

		for i := 0; i < len(keys)/10; i++ {
			ce.cache.Delete(ctx, keys[i])
		}
	}

	return nil
}

// CacheRefresh 缓存刷新策略
type CacheRefresh struct {
	cache         *RedisClient
	refreshFunc   map[string]func(context.Context) (interface{}, error)
	refreshTTL    map[string]time.Duration
	refreshTime   map[string]time.Time
	mu            sync.RWMutex
}

// NewCacheRefresh 创建缓存刷新
func NewCacheRefresh(cache *RedisClient) *CacheRefresh {
	return &CacheRefresh{
		cache:       cache,
		refreshFunc: make(map[string]func(context.Context) (interface{}, error)),
		refreshTTL:  make(map[string]time.Duration),
		refreshTime: make(map[string]time.Time),
	}
}

// Register 注册刷新函数
func (cr *CacheRefresh) Register(key string, fn func(context.Context) (interface{}, error), interval time.Duration) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.refreshFunc[key] = fn
	cr.refreshTTL[key] = interval
	cr.refreshTime[key] = time.Now()
}

// RefreshIfNeeded 需要时刷新缓存
func (cr *CacheRefresh) RefreshIfNeeded(ctx context.Context, key string) error {
	cr.mu.Lock()
	fn, exists := cr.refreshFunc[key]
	interval, _ := cr.refreshTTL[key]
	lastRefresh, _ := cr.refreshTime[key]
	cr.mu.Unlock()

	if !exists {
		return fmt.Errorf("key not registered")
	}

	if time.Since(lastRefresh) < interval {
		return nil
	}

	// 执行刷新
	data, err := fn(ctx)
	if err != nil {
		return err
	}

	cr.cache.Set(ctx, key, data, interval*2)

	cr.mu.Lock()
	cr.refreshTime[key] = time.Now()
	cr.mu.Unlock()

	return nil
}

