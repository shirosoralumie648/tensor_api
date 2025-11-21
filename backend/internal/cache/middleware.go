package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// CacheMiddleware 缓存中间件
type CacheMiddleware struct {
	client      *RedisClient
	cacheTTL    time.Duration
	patterns    map[string]time.Duration
	mu          sync.RWMutex
	enabled     bool
	maxSize     int64
	currentSize int64
}

// NewCacheMiddleware 创建缓存中间件
func NewCacheMiddleware(client *RedisClient, ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		client:      client,
		cacheTTL:    ttl,
		patterns:    make(map[string]time.Duration),
		enabled:     true,
		maxSize:     1 * 1024 * 1024 * 1024, // 1GB
		currentSize: 0,
	}
}

// SetPattern 设置特定路径的缓存 TTL
func (cm *CacheMiddleware) SetPattern(path string, ttl time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.patterns[path] = ttl
}

// getTTL 获取 TTL
func (cm *CacheMiddleware) getTTL(path string) time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if ttl, exists := cm.patterns[path]; exists {
		return ttl
	}

	return cm.cacheTTL
}

// Middleware HTTP 中间件
func (cm *CacheMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cm.enabled || r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		cacheKey := cm.generateCacheKey(r)
		
		// 尝试从缓存获取
		if cached, err := cm.client.Get(r.Context(), cacheKey); err == nil {
			if cachedResp, ok := cached.(*CachedResponse); ok {
				writeResponseFromCache(w, cachedResp)
				return
			}
		}

		// 捕获响应
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           make([]byte, 0),
		}

		next.ServeHTTP(recorder, r)

		// 缓存成功的响应
		if recorder.statusCode == http.StatusOK {
			cachedResp := &CachedResponse{
				Status:  recorder.statusCode,
				Headers: recorder.Header().Clone(),
				Body:    recorder.body,
			}

			cm.client.Set(r.Context(), cacheKey, cachedResp, cm.getTTL(r.URL.Path))
		}
	})
}

// CachedResponse 缓存的响应
type CachedResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	rr.body = append(rr.body, b...)
	return rr.ResponseWriter.Write(b)
}

// generateCacheKey 生成缓存键
func (cm *CacheMiddleware) generateCacheKey(r *http.Request) string {
	h := md5.New()
	io.WriteString(h, r.RequestURI)
	io.WriteString(h, r.Header.Get("Accept"))
	io.WriteString(h, r.Header.Get("Accept-Encoding"))
	
	return fmt.Sprintf("http:%x", h.Sum(nil))
}

// writeResponseFromCache 从缓存写入响应
func writeResponseFromCache(w http.ResponseWriter, resp *CachedResponse) {
	for k, v := range resp.Headers {
		w.Header()[k] = v
	}

	w.Header().Set("X-Cache", "HIT")
	w.WriteHeader(resp.Status)
	w.Write(resp.Body)
}

// InvalidatePattern 删除匹配模式的缓存
func (cm *CacheMiddleware) InvalidatePattern(ctx context.Context, pattern string) error {
	keys := cm.client.Keys(ctx, pattern)
	
	for _, key := range keys {
		cm.client.Delete(ctx, key)
	}

	return nil
}

// SetEnabled 启用/禁用缓存
func (cm *CacheMiddleware) SetEnabled(enabled bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.enabled = enabled
}

// IsEnabled 检查是否启用
func (cm *CacheMiddleware) IsEnabled() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.enabled
}

// CacheInvalidator 缓存失效管理器
type CacheInvalidator struct {
	client       *RedisClient
	mu           sync.RWMutex
	subscriptions map[string][]func()
}

// NewCacheInvalidator 创建缓存失效管理器
func NewCacheInvalidator(client *RedisClient) *CacheInvalidator {
	return &CacheInvalidator{
		client:        client,
		subscriptions: make(map[string][]func()),
	}
}

// Subscribe 订阅缓存失效事件
func (ci *CacheInvalidator) Subscribe(pattern string, callback func()) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ci.subscriptions[pattern] = append(ci.subscriptions[pattern], callback)
}

// Invalidate 失效缓存
func (ci *CacheInvalidator) Invalidate(ctx context.Context, pattern string) error {
	ci.mu.RLock()
	callbacks, exists := ci.subscriptions[pattern]
	ci.mu.RUnlock()

	if exists {
		for _, callback := range callbacks {
			go callback()
		}
	}

	keys := ci.client.Keys(ctx, pattern)
	for _, key := range keys {
		ci.client.Delete(ctx, key)
	}

	return nil
}

// HotDataStrategy 热数据策略
type HotDataStrategy struct {
	client        *RedisClient
	mu            sync.RWMutex
	accessCounts  map[string]int64
	threshold     int64
	preloadFunc   map[string]func(context.Context) (interface{}, error)
}

// NewHotDataStrategy 创建热数据策略
func NewHotDataStrategy(client *RedisClient, threshold int64) *HotDataStrategy {
	return &HotDataStrategy{
		client:       client,
		accessCounts: make(map[string]int64),
		threshold:    threshold,
		preloadFunc:  make(map[string]func(context.Context) (interface{}, error)),
	}
}

// RecordAccess 记录访问
func (hds *HotDataStrategy) RecordAccess(key string) {
	hds.mu.Lock()
	defer hds.mu.Unlock()

	hds.accessCounts[key]++
}

// IsHotData 检查是否为热数据
func (hds *HotDataStrategy) IsHotData(key string) bool {
	hds.mu.RLock()
	defer hds.mu.RUnlock()

	return hds.accessCounts[key] >= hds.threshold
}

// PreloadHotData 预加载热数据
func (hds *HotDataStrategy) PreloadHotData(ctx context.Context) error {
	hds.mu.Lock()
	defer hds.mu.Unlock()

	for key, preload := range hds.preloadFunc {
		if hds.accessCounts[key] >= hds.threshold {
			if data, err := preload(ctx); err == nil {
				hds.client.Set(ctx, key, data, 1*time.Hour)
			}
		}
	}

	return nil
}

// RegisterPreloadFunc 注册预加载函数
func (hds *HotDataStrategy) RegisterPreloadFunc(key string, fn func(context.Context) (interface{}, error)) {
	hds.mu.Lock()
	defer hds.mu.Unlock()

	hds.preloadFunc[key] = fn
}

// GetHotDataStats 获取热数据统计
func (hds *HotDataStrategy) GetHotDataStats() map[string]int64 {
	hds.mu.RLock()
	defer hds.mu.RUnlock()

	stats := make(map[string]int64)
	for key, count := range hds.accessCounts {
		if count >= hds.threshold {
			stats[key] = count
		}
	}

	return stats
}

// CachePrewarmer 缓存预热器
type CachePrewarmer struct {
	client    *RedisClient
	mu        sync.Mutex
	dataFuncs map[string]func(context.Context) (interface{}, error)
	ttls      map[string]time.Duration
}

// NewCachePrewarmer 创建缓存预热器
func NewCachePrewarmer(client *RedisClient) *CachePrewarmer {
	return &CachePrewarmer{
		client:    client,
		dataFuncs: make(map[string]func(context.Context) (interface{}, error)),
		ttls:      make(map[string]time.Duration),
	}
}

// Register 注册预热数据源
func (cp *CachePrewarmer) Register(key string, fn func(context.Context) (interface{}, error), ttl time.Duration) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.dataFuncs[key] = fn
	cp.ttls[key] = ttl
}

// Preheat 执行预热
func (cp *CachePrewarmer) Preheat(ctx context.Context) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for key, fn := range cp.dataFuncs {
		if data, err := fn(ctx); err == nil {
			ttl := cp.ttls[key]
			if ttl == 0 {
				ttl = 1 * time.Hour
			}

			cp.client.Set(ctx, key, data, ttl)
		}
	}

	return nil
}

// PreheatAsync 异步执行预热
func (cp *CachePrewarmer) PreheatAsync(ctx context.Context) {
	go cp.Preheat(ctx)
}

