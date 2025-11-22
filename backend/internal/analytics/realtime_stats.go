package analytics

import (
	"sync"
	"time"
)

// RealtimeStats 实时统计
type RealtimeStats struct {
	Timestamp       time.Time
	Requests        int64
	SuccessRequests int64
	ErrorRequests   int64
	TotalTokens     int64
	TotalCost       float64
	AvgDuration     int64
	QPS             float64
	ErrorRate       float64
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	windowSize time.Duration
	buckets    map[int64]*WindowBucket
	mu         sync.RWMutex
}

// WindowBucket 窗口桶
type WindowBucket struct {
	Timestamp       int64
	Requests        int64
	SuccessRequests int64
	ErrorRequests   int64
	TotalTokens     int64
	TotalCost       float64
	TotalDuration   int64
}

// RealtimeStatsEngine 实时统计引擎
type RealtimeStatsEngine struct {
	mu              sync.RWMutex
	userStats       map[string]*SlidingWindow
	modelStats      map[string]*SlidingWindow
	providerStats   map[string]*SlidingWindow
	windowSize      time.Duration
	bucketDuration  time.Duration
	cache           map[string]*RealtimeStats
	cacheTTL        time.Duration
	lastCacheUpdate map[string]time.Time
}

// NewRealtimeStatsEngine 创建实时统计引擎
func NewRealtimeStatsEngine(windowSize time.Duration) *RealtimeStatsEngine {
	return &RealtimeStatsEngine{
		userStats:       make(map[string]*SlidingWindow),
		modelStats:      make(map[string]*SlidingWindow),
		providerStats:   make(map[string]*SlidingWindow),
		windowSize:      windowSize,
		bucketDuration:  time.Second, // 1 秒桶
		cache:           make(map[string]*RealtimeStats),
		cacheTTL:        5 * time.Second,
		lastCacheUpdate: make(map[string]time.Time),
	}
}

// RecordMetric 记录指标
func (rse *RealtimeStatsEngine) RecordMetric(userID, model, provider string, record *UsageRecord) {
	now := time.Now()
	bucket := now.Unix() / int64(rse.bucketDuration.Seconds())

	// 用户统计
	rse.recordToWindow(rse.userStats, userID, bucket, record)
	// 模型统计
	rse.recordToWindow(rse.modelStats, model, bucket, record)
	// 提供商统计
	rse.recordToWindow(rse.providerStats, provider, bucket, record)

	// 清除缓存
	rse.invalidateCache(userID)
}

// recordToWindow 记录到窗口
func (rse *RealtimeStatsEngine) recordToWindow(windows map[string]*SlidingWindow, key string, bucket int64, record *UsageRecord) {
	rse.mu.Lock()
	window, exists := windows[key]
	if !exists {
		window = &SlidingWindow{
			windowSize: rse.windowSize,
			buckets:    make(map[int64]*WindowBucket),
		}
		windows[key] = window
	}
	rse.mu.Unlock()

	window.AddRecord(bucket, record)
}

// AddRecord 向滑动窗口添加记录
func (sw *SlidingWindow) AddRecord(bucket int64, record *UsageRecord) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if _, exists := sw.buckets[bucket]; !exists {
		sw.buckets[bucket] = &WindowBucket{
			Timestamp: bucket,
		}
	}

	b := sw.buckets[bucket]
	b.Requests++
	b.TotalTokens += record.TotalTokens
	b.TotalCost += record.Cost
	b.TotalDuration += record.Duration

	if record.Status == "success" {
		b.SuccessRequests++
	} else {
		b.ErrorRequests++
	}

	// 清理过期桶
	sw.cleanOldBuckets()
}

// cleanOldBuckets 清理过期桶
func (sw *SlidingWindow) cleanOldBuckets() {
	now := time.Now().Unix() / int64(time.Second)
	windowBuckets := int64(sw.windowSize.Seconds())
	cutoff := now - windowBuckets

	for ts := range sw.buckets {
		if ts < cutoff {
			delete(sw.buckets, ts)
		}
	}
}

// GetStats 获取统计
func (sw *SlidingWindow) GetStats() *RealtimeStats {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	stats := &RealtimeStats{
		Timestamp: time.Now(),
	}

	for _, bucket := range sw.buckets {
		stats.Requests += bucket.Requests
		stats.SuccessRequests += bucket.SuccessRequests
		stats.ErrorRequests += bucket.ErrorRequests
		stats.TotalTokens += bucket.TotalTokens
		stats.TotalCost += bucket.TotalCost
		stats.AvgDuration += bucket.TotalDuration
	}

	if stats.Requests > 0 {
		stats.AvgDuration /= stats.Requests
		stats.ErrorRate = float64(stats.ErrorRequests) / float64(stats.Requests)
		stats.QPS = float64(stats.Requests) / sw.windowSize.Seconds()
	}

	return stats
}

// GetUserStats 获取用户统计
func (rse *RealtimeStatsEngine) GetUserStats(userID string) *RealtimeStats {
	// 检查缓存
	cacheKey := "user:" + userID
	if stats, cached := rse.getFromCache(cacheKey); cached {
		return stats
	}

	rse.mu.RLock()
	window, exists := rse.userStats[userID]
	rse.mu.RUnlock()

	if !exists {
		return &RealtimeStats{Timestamp: time.Now()}
	}

	stats := window.GetStats()
	rse.setCache(cacheKey, stats)
	return stats
}

// GetModelStats 获取模型统计
func (rse *RealtimeStatsEngine) GetModelStats(model string) *RealtimeStats {
	cacheKey := "model:" + model
	if stats, cached := rse.getFromCache(cacheKey); cached {
		return stats
	}

	rse.mu.RLock()
	window, exists := rse.modelStats[model]
	rse.mu.RUnlock()

	if !exists {
		return &RealtimeStats{Timestamp: time.Now()}
	}

	stats := window.GetStats()
	rse.setCache(cacheKey, stats)
	return stats
}

// GetProviderStats 获取提供商统计
func (rse *RealtimeStatsEngine) GetProviderStats(provider string) *RealtimeStats {
	cacheKey := "provider:" + provider
	if stats, cached := rse.getFromCache(cacheKey); cached {
		return stats
	}

	rse.mu.RLock()
	window, exists := rse.providerStats[provider]
	rse.mu.RUnlock()

	if !exists {
		return &RealtimeStats{Timestamp: time.Now()}
	}

	stats := window.GetStats()
	rse.setCache(cacheKey, stats)
	return stats
}

// getFromCache 从缓存获取
func (rse *RealtimeStatsEngine) getFromCache(key string) (*RealtimeStats, bool) {
	rse.mu.RLock()
	defer rse.mu.RUnlock()

	stats, exists := rse.cache[key]
	if !exists {
		return nil, false
	}

	lastUpdate := rse.lastCacheUpdate[key]
	if time.Since(lastUpdate) > rse.cacheTTL {
		return nil, false
	}

	return stats, true
}

// setCache 设置缓存
func (rse *RealtimeStatsEngine) setCache(key string, stats *RealtimeStats) {
	rse.mu.Lock()
	defer rse.mu.Unlock()

	rse.cache[key] = stats
	rse.lastCacheUpdate[key] = time.Now()
}

// invalidateCache 清除缓存
func (rse *RealtimeStatsEngine) invalidateCache(userID string) {
	rse.mu.Lock()
	defer rse.mu.Unlock()

	delete(rse.cache, "user:"+userID)
	delete(rse.lastCacheUpdate, "user:"+userID)
}

// ModelStatItem 模型统计项
type ModelStatItem struct {
	Model    string
	Requests int64
	Cost     float64
}

// GetTopModels 获取热门模型
func (rse *RealtimeStatsEngine) GetTopModels(limit int) []*ModelStatItem {
	rse.mu.RLock()
	defer rse.mu.RUnlock()

	var items []*ModelStatItem
	for model, window := range rse.modelStats {
		stats := window.GetStats()
		items = append(items, &ModelStatItem{
			Model:    model,
			Requests: stats.Requests,
			Cost:     stats.TotalCost,
		})
	}

	// 简单排序（按请求数）
	if len(items) > limit {
		items = items[:limit]
	}

	return items
}
