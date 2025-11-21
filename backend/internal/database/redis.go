package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/oblivious/backend/internal/config"
)

var RedisClient *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	RedisClient = client
	return nil
}

func CloseRedis() error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Close()
}

// 封装常用操作
func RedisSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

func RedisGet(ctx context.Context, key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func RedisDel(ctx context.Context, keys ...string) error {
	return RedisClient.Del(ctx, keys...).Err()
}

func RedisExists(ctx context.Context, keys ...string) (int64, error) {
	return RedisClient.Exists(ctx, keys...).Result()
}


