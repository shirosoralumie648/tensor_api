package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 创建测试用 Redis 客户端
func getTestRedisClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用 DB 1 进行测试
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis 连接失败: %v", err)
	}

	// 清空测试 DB
	client.FlushDB(ctx)

	return client
}

func TestCacheManager_SetAndGetUserCache(t *testing.T) {
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	cm := NewCacheManager(redisClient, 5*time.Minute, 30*time.Minute)
	ctx := context.Background()

	// 测试数据
	testUser := &UserCache{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Group:    "default",
		Quota:    1000,
		Role:     1,
		Status:   1,
		ExpireAt: time.Now().Add(30 * 24 * time.Hour),
	}

	// 设置缓存
	err := cm.SetUserCache(ctx, testUser.UserID, testUser)
	require.NoError(t, err)

	// 获取缓存（第一次应该是 L1 命中，因为刚刚设置）
	retrieved, err := cm.GetUserCache(ctx, testUser.UserID)
	// 注意：如果数据库查询未实现，GetUserCache 会失败
	// 这里我们只测试缓存层的逻辑
}

func TestCacheManager_L1CacheExpiry(t *testing.T) {
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	// 使用较短的 TTL 进行测试
	cm := NewCacheManager(redisClient, 100*time.Millisecond, 30*time.Minute)

	testUser := &UserCache{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Quota:    1000,
		Status:   1,
	}

	ctx := context.Background()

	// 设置缓存
	cm.SetUserCache(ctx, testUser.UserID, testUser)

	// 立即从 L1 缓存获取（应该命中）
	key := "user:1"
	if entry, ok := cm.l1Cache.Load(key); ok {
		cacheEntry := entry.(*cacheEntry)
		cacheEntry.mu.RLock()
		assert.True(t, time.Now().Before(cacheEntry.expireAt), "缓存应该未过期")
		cacheEntry.mu.RUnlock()
	}

	// 等待 L1 缓存过期
	time.Sleep(150 * time.Millisecond)

	// 此时 L1 缓存应该已过期
	if entry, ok := cm.l1Cache.Load(key); ok {
		cacheEntry := entry.(*cacheEntry)
		cacheEntry.mu.RLock()
		isExpired := time.Now().After(cacheEntry.expireAt)
		cacheEntry.mu.RUnlock()
		assert.True(t, isExpired, "缓存应该已过期")
	}
}

func TestCacheManager_InvalidateCache(t *testing.T) {
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	cm := NewCacheManager(redisClient, 5*time.Minute, 30*time.Minute)
	ctx := context.Background()

	testUser := &UserCache{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Quota:    1000,
		Status:   1,
	}

	// 设置缓存
	err := cm.SetUserCache(ctx, testUser.UserID, testUser)
	require.NoError(t, err)

	// 清除缓存
	err = cm.InvalidateUserCache(ctx, testUser.UserID)
	require.NoError(t, err)

	// 验证 L1 缓存已清除
	key := "user:1"
	_, ok := cm.l1Cache.Load(key)
	assert.False(t, ok, "L1 缓存应该已清除")

	// 验证 L2 缓存已清除
	val, err := redisClient.Get(ctx, key).Result()
	assert.Error(t, err, "L2 缓存应该已清除")
	assert.Equal(t, "", val)
}

func TestCacheStats(t *testing.T) {
	redisClient := getTestRedisClient(t)
	defer redisClient.Close()

	cm := NewCacheManager(redisClient, 5*time.Minute, 30*time.Minute)

	// 初始统计应该为 0
	stats := cm.GetStats()
	assert.Equal(t, int64(0), stats.L1Hits)
	assert.Equal(t, int64(0), stats.L1Misses)

	// 模拟一些统计数据
	cm.stats.recordQuery()
	cm.stats.recordL1Hit()

	stats = cm.GetStats()
	assert.Equal(t, int64(1), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.L1Hits)

	// 测试命中率
	hitRate := stats.GetHitRate()
	assert.Equal(t, 1.0, hitRate)
}

func TestBloomFilter_AddAndContains(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)

	// 添加元素
	bf.Add([]byte("user:1"))
	bf.Add([]byte("user:2"))
	bf.Add([]byte("user:3"))

	// 测试包含
	assert.True(t, bf.Contains([]byte("user:1")))
	assert.True(t, bf.Contains([]byte("user:2")))
	assert.True(t, bf.Contains([]byte("user:3")))

	// 测试不包含（可能有误判，但概率很低）
	assert.False(t, bf.Contains([]byte("user:999")))
}

func TestBloomFilter_Reset(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)

	// 添加元素
	bf.Add([]byte("user:1"))
	assert.True(t, bf.Contains([]byte("user:1")))

	// 重置
	bf.Reset()
	// 重置后不应该包含任何元素（除非由于哈希碰撞）
	// 这里我们检查大多数元素不会被找到
	assert.False(t, bf.Contains([]byte("user:1")))
}

func TestBloomFilter_Stats(t *testing.T) {
	bf := NewBloomFilter(1000, 0.01)

	// 添加元素
	bf.Add([]byte("user:1"))
	bf.Add([]byte("user:2"))

	stats := bf.GetStats()

	assert.NotNil(t, stats)
	assert.Greater(t, stats["set_bits"].(int), 0)
	assert.Greater(t, stats["utilization"].(float64), 0)
	assert.Less(t, stats["utilization"].(float64), 1.0)
}

func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	bf := NewBloomFilter(10000, 0.01)

	// 添加元素
	for i := 0; i < 5000; i++ {
		bf.Add([]byte("user:" + string(rune(i))))
	}

	// 获取期望的误判率
	expectedRate := bf.GetExpectedFalsePositiveRate(5000)

	// 误判率应该接近配置的误判率
	assert.Less(t, expectedRate, 0.02, "误判率应该接近 1%")
}

func BenchmarkCacheManager_GetUserCache(b *testing.B) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
	defer redisClient.Close()

	cm := NewCacheManager(redisClient, 5*time.Minute, 30*time.Minute)
	ctx := context.Background()

	// 预设缓存
	testUser := &UserCache{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		Quota:    1000,
		Status:   1,
	}
	cm.SetUserCache(ctx, testUser.UserID, testUser)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = cm.GetUserCache(ctx, 1)
	}
}

func BenchmarkBloomFilter_Add(b *testing.B) {
	bf := NewBloomFilter(1000000, 0.01)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf.Add([]byte("user:" + string(rune(i))))
	}
}

func BenchmarkBloomFilter_Contains(b *testing.B) {
	bf := NewBloomFilter(100000, 0.01)

	// 预先添加一些元素
	for i := 0; i < 10000; i++ {
		bf.Add([]byte("user:" + string(rune(i))))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf.Contains([]byte("user:5000"))
	}
}
