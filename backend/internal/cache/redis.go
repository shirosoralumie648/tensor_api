package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RedisClient Redis 客户端包装
type RedisClient struct {
	mu        sync.RWMutex
	data      map[string]*CacheEntry
	ttls      map[string]time.Time
	cluster   bool
	password  string
	addresses []string
	db        int
	stats     *CacheStats
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key      string
	Value    interface{}
	TTL      time.Duration
	CreatedAt time.Time
	UpdatedAt time.Time
	HitCount  int64
}

// CacheStats 缓存统计
type CacheStats struct {
	mu           sync.RWMutex
	Hits         int64
	Misses       int64
	Sets         int64
	Deletes      int64
	Expirations  int64
	EvictionSize int64
}

// CacheConfig Redis 配置
type CacheConfig struct {
	Addrs      []string
	Password   string
	DB         int
	PoolSize   int
	MaxRetries int
	TTL        time.Duration
	ClusterMode bool
}

// NewRedisClient 创建 Redis 客户端
func NewRedisClient(cfg *CacheConfig) (*RedisClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if len(cfg.Addrs) == 0 {
		return nil, fmt.Errorf("addresses are required")
	}

	client := &RedisClient{
		data:      make(map[string]*CacheEntry),
		ttls:      make(map[string]time.Time),
		cluster:   cfg.ClusterMode,
		password:  cfg.Password,
		addresses: cfg.Addrs,
		db:        cfg.DB,
		stats:     &CacheStats{},
	}

	// 启动过期清理 goroutine
	go client.cleanupExpired()

	return client, nil
}

// Set 设置缓存
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	rc.data[key] = &CacheEntry{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if ttl > 0 {
		rc.ttls[key] = now.Add(ttl)
	}

	rc.stats.mu.Lock()
	rc.stats.Sets++
	rc.stats.mu.Unlock()

	return nil
}

// Get 获取缓存
func (rc *RedisClient) Get(ctx context.Context, key string) (interface{}, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	entry, exists := rc.data[key]
	if !exists {
		rc.stats.mu.Lock()
		rc.stats.Misses++
		rc.stats.mu.Unlock()
		return nil, fmt.Errorf("key not found")
	}

	// 检查过期
	if expTime, hasExpiry := rc.ttls[key]; hasExpiry && time.Now().After(expTime) {
		rc.stats.mu.Lock()
		rc.stats.Expirations++
		rc.stats.mu.Unlock()
		return nil, fmt.Errorf("key expired")
	}

	entry.HitCount++
	rc.stats.mu.Lock()
	rc.stats.Hits++
	rc.stats.mu.Unlock()

	return entry.Value, nil
}

// Delete 删除缓存
func (rc *RedisClient) Delete(ctx context.Context, key string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.data[key]; exists {
		delete(rc.data, key)
		delete(rc.ttls, key)
		
		rc.stats.mu.Lock()
		rc.stats.Deletes++
		rc.stats.mu.Unlock()
	}

	return nil
}

// Exists 检查键是否存在
func (rc *RedisClient) Exists(ctx context.Context, key string) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	_, exists := rc.data[key]
	return exists
}

// Expire 设置过期时间
func (rc *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.data[key]; !exists {
		return fmt.Errorf("key not found")
	}

	if ttl > 0 {
		rc.ttls[key] = time.Now().Add(ttl)
	}

	return nil
}

// TTL 获取剩余 TTL
func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	if _, exists := rc.data[key]; !exists {
		return 0, fmt.Errorf("key not found")
	}

	if expTime, hasExpiry := rc.ttls[key]; hasExpiry {
		remaining := time.Until(expTime)
		if remaining > 0 {
			return remaining, nil
		}
	}

	return 0, nil
}

// Incr 原子增量
func (rc *RedisClient) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	var value int64
	if entry, exists := rc.data[key]; exists {
		if v, ok := entry.Value.(int64); ok {
			value = v
		}
	}

	newValue := value + delta
	rc.data[key] = &CacheEntry{
		Key:       key,
		Value:     newValue,
		UpdatedAt: time.Now(),
	}

	return newValue, nil
}

// Decr 原子减量
func (rc *RedisClient) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return rc.Incr(ctx, key, -delta)
}

// MGet 批量获取
func (rc *RedisClient) MGet(ctx context.Context, keys ...string) map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	result := make(map[string]interface{})
	for _, key := range keys {
		if entry, exists := rc.data[key]; exists {
			if expTime, hasExpiry := rc.ttls[key]; !hasExpiry || time.Now().Before(expTime) {
				result[key] = entry.Value
			}
		}
	}

	return result
}

// MSet 批量设置
func (rc *RedisClient) MSet(ctx context.Context, kvs map[string]interface{}) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	for key, value := range kvs {
		rc.data[key] = &CacheEntry{
			Key:       key,
			Value:     value,
			UpdatedAt: now,
		}
	}

	rc.stats.mu.Lock()
	rc.stats.Sets += int64(len(kvs))
	rc.stats.mu.Unlock()

	return nil
}

// Del 删除多个键
func (rc *RedisClient) Del(ctx context.Context, keys ...string) int64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	count := int64(0)
	for _, key := range keys {
		if _, exists := rc.data[key]; exists {
			delete(rc.data, key)
			delete(rc.ttls, key)
			count++
		}
	}

	rc.stats.mu.Lock()
	rc.stats.Deletes += count
	rc.stats.mu.Unlock()

	return count
}

// Clear 清空所有缓存
func (rc *RedisClient) Clear(ctx context.Context) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.data = make(map[string]*CacheEntry)
	rc.ttls = make(map[string]time.Time)

	return nil
}

// Keys 获取所有键
func (rc *RedisClient) Keys(ctx context.Context, pattern string) []string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	var keys []string
	for k := range rc.data {
		keys = append(keys, k)
	}

	return keys
}

// GetStats 获取统计信息
func (rc *RedisClient) GetStats() map[string]interface{} {
	rc.stats.mu.RLock()
	defer rc.stats.mu.RUnlock()

	total := rc.stats.Hits + rc.stats.Misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(rc.stats.Hits) / float64(total) * 100
	}

	rc.mu.RLock()
	cacheSize := len(rc.data)
	rc.mu.RUnlock()

	return map[string]interface{}{
		"hits":        rc.stats.Hits,
		"misses":      rc.stats.Misses,
		"hit_rate":    hitRate,
		"sets":        rc.stats.Sets,
		"deletes":     rc.stats.Deletes,
		"expirations": rc.stats.Expirations,
		"cache_size":  cacheSize,
	}
}

// Close 关闭连接
func (rc *RedisClient) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.data = make(map[string]*CacheEntry)
	rc.ttls = make(map[string]time.Time)

	return nil
}

// cleanupExpired 清理过期项
func (rc *RedisClient) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rc.mu.Lock()

		now := time.Now()
		expiredKeys := make([]string, 0)

		for key, expTime := range rc.ttls {
			if now.After(expTime) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(rc.data, key)
			delete(rc.ttls, key)
			
			rc.stats.mu.Lock()
			rc.stats.Expirations++
			rc.stats.mu.Unlock()
		}

		rc.mu.Unlock()
	}
}

// Pipeline 管道操作
type Pipeline struct {
	client    *RedisClient
	commands  []*pipelineCommand
	mu        sync.Mutex
}

type pipelineCommand struct {
	name string
	args []interface{}
}

// NewPipeline 创建管道
func (rc *RedisClient) NewPipeline() *Pipeline {
	return &Pipeline{
		client:   rc,
		commands: make([]*pipelineCommand, 0),
	}
}

// Set 添加 Set 命令到管道
func (p *Pipeline) Set(key string, value interface{}, ttl time.Duration) *Pipeline {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.commands = append(p.commands, &pipelineCommand{
		name: "SET",
		args: []interface{}{key, value, ttl},
	})

	return p
}

// Get 添加 Get 命令到管道
func (p *Pipeline) Get(key string) *Pipeline {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.commands = append(p.commands, &pipelineCommand{
		name: "GET",
		args: []interface{}{key},
	})

	return p
}

// Execute 执行管道
func (p *Pipeline) Execute(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, cmd := range p.commands {
		switch cmd.name {
		case "SET":
			if len(cmd.args) >= 3 {
				p.client.Set(ctx, cmd.args[0].(string), cmd.args[1], cmd.args[2].(time.Duration))
			}
		case "GET":
			if len(cmd.args) >= 1 {
				p.client.Get(ctx, cmd.args[0].(string))
			}
		}
	}

	p.commands = make([]*pipelineCommand, 0)
	return nil
}

