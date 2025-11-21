package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// RateLimitStrategy 限流策略
type RateLimitStrategy string

const (
	StrategyTokenBucket  RateLimitStrategy = "token_bucket"
	StrategySlidingWindow RateLimitStrategy = "sliding_window"
	StrategyFixedWindow  RateLimitStrategy = "fixed_window"
)

// Limiter 限流器接口
type Limiter interface {
	Allow(key string) bool
	AllowN(key string, n int64) bool
	GetRemaining(key string) int64
	Reset(key string)
}

// ==================== 令牌桶限流器 ====================

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	mu           sync.RWMutex
	buckets      map[string]*TokenBucket
	capacity     int64         // 桶容量
	refillRate   int64         // 每秒补充速率
	refillTicker *time.Ticker
}

// TokenBucket 令牌桶
type TokenBucket struct {
	Tokens    int64
	LastRefill time.Time
	Capacity  int64
	RefillRate int64
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(capacity int64, refillRate int64) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}

	// 启动定期补充
	limiter.refillTicker = time.NewTicker(time.Second)
	go limiter.refillRoutine()

	return limiter
}

// Allow 允许一个请求
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	return tbl.AllowN(key, 1)
}

// AllowN 允许 N 个请求
func (tbl *TokenBucketLimiter) AllowN(key string, n int64) bool {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	bucket, exists := tbl.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			Tokens:     tbl.capacity,
			LastRefill: time.Now(),
			Capacity:   tbl.capacity,
			RefillRate: tbl.refillRate,
		}
		tbl.buckets[key] = bucket
	}

	if bucket.Tokens >= n {
		bucket.Tokens -= n
		return true
	}

	return false
}

// GetRemaining 获取剩余令牌
func (tbl *TokenBucketLimiter) GetRemaining(key string) int64 {
	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	if bucket, exists := tbl.buckets[key]; exists {
		return bucket.Tokens
	}

	return tbl.capacity
}

// Reset 重置限流器
func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	if bucket, exists := tbl.buckets[key]; exists {
		bucket.Tokens = bucket.Capacity
		bucket.LastRefill = time.Now()
	}
}

// refillRoutine 补充例程
func (tbl *TokenBucketLimiter) refillRoutine() {
	for range tbl.refillTicker.C {
		tbl.mu.Lock()
		for _, bucket := range tbl.buckets {
			if bucket.Tokens < bucket.Capacity {
				newTokens := bucket.Tokens + bucket.RefillRate
				if newTokens > bucket.Capacity {
					bucket.Tokens = bucket.Capacity
				} else {
					bucket.Tokens = newTokens
				}
			}
		}
		tbl.mu.Unlock()
	}
}

// Close 关闭限流器
func (tbl *TokenBucketLimiter) Close() {
	tbl.refillTicker.Stop()
}

// ==================== 滑动窗口限流器 ====================

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	mu         sync.RWMutex
	windows    map[string]*SlidingWindowCounter
	windowSize time.Duration
	maxRequests int64
}

// SlidingWindowCounter 滑动窗口计数器
type SlidingWindowCounter struct {
	Requests  []time.Time
	MaxReqs   int64
	WindowSize time.Duration
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(windowSize time.Duration, maxRequests int64) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows:     make(map[string]*SlidingWindowCounter),
		windowSize:  windowSize,
		maxRequests: maxRequests,
	}
}

// Allow 允许一个请求
func (swl *SlidingWindowLimiter) Allow(key string) bool {
	return swl.AllowN(key, 1)
}

// AllowN 允许 N 个请求
func (swl *SlidingWindowLimiter) AllowN(key string, n int64) bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()
	counter, exists := swl.windows[key]
	if !exists {
		counter = &SlidingWindowCounter{
			Requests:  make([]time.Time, 0),
			MaxReqs:   swl.maxRequests,
			WindowSize: swl.windowSize,
		}
		swl.windows[key] = counter
	}

	// 移除窗口外的请求
	cutoff := now.Add(-swl.windowSize)
	var validRequests []time.Time
	for _, req := range counter.Requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	counter.Requests = validRequests

	// 检查是否可以添加新请求
	if int64(len(counter.Requests))+n <= swl.maxRequests {
		for i := int64(0); i < n; i++ {
			counter.Requests = append(counter.Requests, now)
		}
		return true
	}

	return false
}

// GetRemaining 获取剩余请求数
func (swl *SlidingWindowLimiter) GetRemaining(key string) int64 {
	swl.mu.RLock()
	defer swl.mu.RUnlock()

	if counter, exists := swl.windows[key]; exists {
		return swl.maxRequests - int64(len(counter.Requests))
	}

	return swl.maxRequests
}

// Reset 重置限流器
func (swl *SlidingWindowLimiter) Reset(key string) {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	if counter, exists := swl.windows[key]; exists {
		counter.Requests = make([]time.Time, 0)
	}
}

// ==================== 固定窗口限流器 ====================

// FixedWindowLimiter 固定窗口限流器
type FixedWindowLimiter struct {
	mu          sync.RWMutex
	windows     map[string]*FixedWindowCounter
	windowSize  time.Duration
	maxRequests int64
}

// FixedWindowCounter 固定窗口计数器
type FixedWindowCounter struct {
	Count     int64
	WindowStart time.Time
	MaxReqs   int64
	WindowSize time.Duration
}

// NewFixedWindowLimiter 创建固定窗口限流器
func NewFixedWindowLimiter(windowSize time.Duration, maxRequests int64) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windows:     make(map[string]*FixedWindowCounter),
		windowSize:  windowSize,
		maxRequests: maxRequests,
	}
}

// Allow 允许一个请求
func (fwl *FixedWindowLimiter) Allow(key string) bool {
	return fwl.AllowN(key, 1)
}

// AllowN 允许 N 个请求
func (fwl *FixedWindowLimiter) AllowN(key string, n int64) bool {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	now := time.Now()
	counter, exists := fwl.windows[key]
	if !exists {
		counter = &FixedWindowCounter{
			Count:       0,
			WindowStart: now,
			MaxReqs:     fwl.maxRequests,
			WindowSize:  fwl.windowSize,
		}
		fwl.windows[key] = counter
	}

	// 检查窗口是否过期
	if now.After(counter.WindowStart.Add(fwl.windowSize)) {
		counter.Count = 0
		counter.WindowStart = now
	}

	// 检查是否可以添加新请求
	if counter.Count+n <= fwl.maxRequests {
		counter.Count += n
		return true
	}

	return false
}

// GetRemaining 获取剩余请求数
func (fwl *FixedWindowLimiter) GetRemaining(key string) int64 {
	fwl.mu.RLock()
	defer fwl.mu.RUnlock()

	if counter, exists := fwl.windows[key]; exists {
		return fwl.maxRequests - counter.Count
	}

	return fwl.maxRequests
}

// Reset 重置限流器
func (fwl *FixedWindowLimiter) Reset(key string) {
	fwl.mu.Lock()
	defer fwl.mu.Unlock()

	if counter, exists := fwl.windows[key]; exists {
		counter.Count = 0
	}
}

// ==================== 多级限流器 ====================

// MultiLevelLimiter 多级限流器
type MultiLevelLimiter struct {
	limiters map[string]Limiter
	mu       sync.RWMutex
}

// NewMultiLevelLimiter 创建多级限流器
func NewMultiLevelLimiter() *MultiLevelLimiter {
	return &MultiLevelLimiter{
		limiters: make(map[string]Limiter),
	}
}

// AddLimiter 添加限流器
func (mll *MultiLevelLimiter) AddLimiter(name string, limiter Limiter) {
	mll.mu.Lock()
	defer mll.mu.Unlock()

	mll.limiters[name] = limiter
}

// Allow 允许请求（所有限流器都必须允许）
func (mll *MultiLevelLimiter) Allow(key string) bool {
	mll.mu.RLock()
	defer mll.mu.RUnlock()

	for _, limiter := range mll.limiters {
		if !limiter.Allow(key) {
			return false
		}
	}

	return true
}

// GetStatus 获取状态
func (mll *MultiLevelLimiter) GetStatus(key string) map[string]interface{} {
	mll.mu.RLock()
	defer mll.mu.RUnlock()

	status := make(map[string]interface{})
	for name, limiter := range mll.limiters {
		status[name] = map[string]interface{}{
			"remaining": limiter.GetRemaining(key),
		}
	}

	return status
}


