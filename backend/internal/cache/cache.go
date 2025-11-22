package cache

import (
	"context"
	"sync"
	"time"
)

// CacheStats 缓存统计信息（统一定义）
type CacheStats struct {
	mu          sync.RWMutex
	Hits        int64   // 命中次数
	Misses      int64   // 未命中次数
	Sets        int64   // 设置次数
	Deletes     int64   // 删除次数
	Expirations int64   // 过期次数
	HitRate     float64 // 命中率

	L1Hits       int64 // L1缓存命中（多级缓存使用）
	L1Misses     int64 // L1缓存未命中
	L2Hits       int64 // L2缓存命中
	L2Misses     int64 // L2缓存未命中
	DBHits       int64 // 数据库命中
	TotalQueries int64 // 总查询次数
}

// CalculateHitRate 计算命中率
func (s *CacheStats) CalculateHitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	s.HitRate = float64(s.Hits) / float64(total) * 100
	return s.HitRate
}

// Cache 通用缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (interface{}, error)

	// Set 设置缓存值
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete 删除缓存值
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// GetStats 获取统计信息
	GetStats() *CacheStats
}
