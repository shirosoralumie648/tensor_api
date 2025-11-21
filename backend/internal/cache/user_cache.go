package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

// UserCache 用户信息缓存结构
type UserCache struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Group    string    `json:"group"`
	Quota    int64     `json:"quota"`
	Role     int       `json:"role"`
	Status   int       `json:"status"`
	ExpireAt time.Time `json:"expire_at"`
	CachedAt time.Time `json:"cached_at"`
}

// CacheManager 多级缓存管理器
type CacheManager struct {
	// L1 缓存：本地内存（同步Map）
	l1Cache sync.Map // key: string, value: *cacheEntry

	// L2 缓存：Redis
	l2Cache *redis.Client

	// 单次飞行控制（防止缓存击穿）
	singleflight singleflight.Group

	// 缓存配置
	l1TTL time.Duration // L1 缓存过期时间（默认5分钟）
	l2TTL time.Duration // L2 缓存过期时间（默认30分钟）

	// 布隆过滤器用于防止缓存穿透
	bloomFilter BloomFilter

	// 统计数据
	stats *CacheStats
}

// cacheEntry L1 缓存条目
type cacheEntry struct {
	data     *UserCache
	expireAt time.Time
	mu       sync.RWMutex
}

// CacheStats 缓存统计数据
type CacheStats struct {
	L1Hits       int64
	L1Misses     int64
	L2Hits       int64
	L2Misses     int64
	DBHits       int64
	TotalQueries int64
	mu           sync.RWMutex
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(redisClient *redis.Client, l1TTL, l2TTL time.Duration) *CacheManager {
	return &CacheManager{
		l2Cache:     redisClient,
		l1TTL:       l1TTL,
		l2TTL:       l2TTL,
		bloomFilter: NewBloomFilter(100000, 0.01), // 10万容量，1%误判率
		stats:       &CacheStats{},
	}
}

// GetUserCache 获取用户缓存（三级缓存策略）
func (cm *CacheManager) GetUserCache(ctx context.Context, userID int) (*UserCache, error) {
	cm.stats.recordQuery()

	// 检查布隆过滤器（快速判断是否可能存在）
	userIDStr := fmt.Sprintf("user:%d", userID)
	if !cm.bloomFilter.Contains([]byte(userIDStr)) {
		// 用户很可能不存在
		return nil, fmt.Errorf("user not found (bloom filter)")
	}

	// L1 缓存检查
	if entry, ok := cm.l1Cache.Load(userIDStr); ok {
		cacheEntry := entry.(*cacheEntry)
		cacheEntry.mu.RLock()
		defer cacheEntry.mu.RUnlock()

		if time.Now().Before(cacheEntry.expireAt) {
			cm.stats.recordL1Hit()
			return cacheEntry.data, nil
		}
		// 过期则删除
		cm.l1Cache.Delete(userIDStr)
	}
	cm.stats.recordL1Miss()

	// 使用 singleflight 防止缓存击穿
	val, err, shared := cm.singleflight.Do(userIDStr, func() (interface{}, error) {
		// L2 缓存检查
		cacheData, err := cm.getFromL2Cache(ctx, userIDStr)
		if err == nil && cacheData != nil {
			cm.stats.recordL2Hit()
			// 更新 L1 缓存
			cm.setL1Cache(userIDStr, cacheData)
			return cacheData, nil
		}
		cm.stats.recordL2Miss()

		// 从数据库获取
		cacheData, err = cm.getFromDatabase(ctx, userID)
		if err != nil {
			return nil, err
		}

		if cacheData != nil {
			cm.stats.recordDBHit()
			// 同时更新 L1 和 L2 缓存
			cm.setL1Cache(userIDStr, cacheData)
			_ = cm.setL2Cache(ctx, userIDStr, cacheData)
			// 更新布隆过滤器
			cm.bloomFilter.Add([]byte(userIDStr))
		}

		return cacheData, nil
	})

	if err != nil {
		return nil, err
	}

	// 如果是并发请求，直接返回共享的结果
	if shared {
		if val == nil {
			return nil, fmt.Errorf("user not found")
		}
	}

	return val.(*UserCache), nil
}

// SetUserCache 设置用户缓存
func (cm *CacheManager) SetUserCache(ctx context.Context, userID int, cache *UserCache) error {
	userIDStr := fmt.Sprintf("user:%d", userID)

	// 更新 L1 缓存
	cm.setL1Cache(userIDStr, cache)

	// 更新 L2 缓存
	if err := cm.setL2Cache(ctx, userIDStr, cache); err != nil {
		return fmt.Errorf("failed to set L2 cache: %w", err)
	}

	// 更新布隆过滤器
	cm.bloomFilter.Add([]byte(userIDStr))

	return nil
}

// InvalidateUserCache 清除用户缓存
func (cm *CacheManager) InvalidateUserCache(ctx context.Context, userID int) error {
	userIDStr := fmt.Sprintf("user:%d", userID)

	// 删除 L1 缓存
	cm.l1Cache.Delete(userIDStr)

	// 删除 L2 缓存
	if err := cm.l2Cache.Del(ctx, userIDStr).Err(); err != nil {
		return fmt.Errorf("failed to invalidate L2 cache: %w", err)
	}

	return nil
}

// 私有方法

// setL1Cache 设置 L1 缓存
func (cm *CacheManager) setL1Cache(key string, data *UserCache) {
	entry := &cacheEntry{
		data:     data,
		expireAt: time.Now().Add(cm.l1TTL),
	}
	cm.l1Cache.Store(key, entry)
}

// getFromL2Cache 从 L2 缓存获取
func (cm *CacheManager) getFromL2Cache(ctx context.Context, key string) (*UserCache, error) {
	val, err := cm.l2Cache.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在
		}
		return nil, err
	}

	var cache UserCache
	if err := json.Unmarshal([]byte(val), &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return &cache, nil
}

// setL2Cache 设置 L2 缓存
func (cm *CacheManager) setL2Cache(ctx context.Context, key string, data *UserCache) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return cm.l2Cache.Set(ctx, key, jsonData, cm.l2TTL).Err()
}

// getFromDatabase 从数据库获取（这里是占位符，实际使用时需要调用真实的数据库操作）
func (cm *CacheManager) getFromDatabase(ctx context.Context, userID int) (*UserCache, error) {
	// TODO: 实现真实的数据库查询
	// 这里应该调用 UserRepository 的方法
	return nil, fmt.Errorf("database query not implemented")
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() *CacheStats {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()

	return &CacheStats{
		L1Hits:       cm.stats.L1Hits,
		L1Misses:     cm.stats.L1Misses,
		L2Hits:       cm.stats.L2Hits,
		L2Misses:     cm.stats.L2Misses,
		DBHits:       cm.stats.DBHits,
		TotalQueries: cm.stats.TotalQueries,
	}
}

// GetHitRate 获取缓存命中率
func (s *CacheStats) GetHitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.TotalQueries == 0 {
		return 0
	}

	hits := s.L1Hits + s.L2Hits
	return float64(hits) / float64(s.TotalQueries)
}

// 统计方法

func (s *CacheStats) recordQuery() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalQueries++
}

func (s *CacheStats) recordL1Hit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.L1Hits++
}

func (s *CacheStats) recordL1Miss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.L1Misses++
}

func (s *CacheStats) recordL2Hit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.L2Hits++
}

func (s *CacheStats) recordL2Miss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.L2Misses++
}

func (s *CacheStats) recordDBHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DBHits++
}
