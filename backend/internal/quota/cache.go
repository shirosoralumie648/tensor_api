package quota

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisQuotaCache Redis配额缓存实现
type RedisQuotaCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisQuotaCache 创建Redis配额缓存
func NewRedisQuotaCache(client *redis.Client) *RedisQuotaCache {
	return &RedisQuotaCache{
		client: client,
		ttl:    15 * time.Minute, // 预扣费记录15分钟过期
	}
}

// SetPreConsumed 设置预扣费记录
func (c *RedisQuotaCache) SetPreConsumed(record *PreConsumedRecord) error {
	ctx := context.Background()
	key := fmt.Sprintf("pre_consumed:%s", record.RequestID)

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// GetPreConsumed 获取预扣费记录
func (c *RedisQuotaCache) GetPreConsumed(requestID string) (*PreConsumedRecord, error) {
	ctx := context.Background()
	key := fmt.Sprintf("pre_consumed:%s", requestID)

	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("pre-consumed record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	var record PreConsumedRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return &record, nil
}

// DeletePreConsumed 删除预扣费记录
func (c *RedisQuotaCache) DeletePreConsumed(requestID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("pre_consumed:%s", requestID)
	return c.client.Del(ctx, key).Err()
}

// GetUserBalance 获取用户余额缓存
func (c *RedisQuotaCache) GetUserBalance(userID int) (float64, bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user_balance:%d", userID)

	val, err := c.client.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get user balance: %w", err)
	}

	return val, true, nil
}

// SetUserBalance 设置用户余额缓存
func (c *RedisQuotaCache) SetUserBalance(userID int, balance float64) error {
	ctx := context.Background()
	key := fmt.Sprintf("user_balance:%d", userID)
	return c.client.Set(ctx, key, balance, 5*time.Minute).Err()
}

// InvalidateUserBalance 失效用户余额缓存
func (c *RedisQuotaCache) InvalidateUserBalance(userID int) error {
	ctx := context.Background()
	key := fmt.Sprintf("user_balance:%d", userID)
	return c.client.Del(ctx, key).Err()
}
