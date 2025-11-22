package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
)

type RateLimitConfig struct {
	Rate  int           // 每秒请求数
	Burst int           // 突发容量
	TTL   time.Duration // 过期时间
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(cfg *RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取限流 Key（优先使用 user_id，否则使用 IP）
		var key string
		if userID, exists := c.Get("user_id"); exists {
			key = fmt.Sprintf("rate_limit:user:%d", userID.(int))
		} else {
			key = fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		}

		// 检查是否允许请求
		allowed, err := checkRateLimit(c.Request.Context(), key, cfg)
		if err != nil {
			utils.InternalError(c, "限流检查失败")
			c.Abort()
			return
		}

		if !allowed {
			utils.Error(c, 429, utils.ErrRateLimitExceeded, "请求过于频繁，请稍后再试", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit 使用令牌桶算法检查限流
func checkRateLimit(ctx context.Context, key string, cfg *RateLimitConfig) (bool, error) {
	now := time.Now().Unix()

	// Lua 脚本实现令牌桶算法
	script := `
        local key = KEYS[1]
        local rate = tonumber(ARGV[1])
        local burst = tonumber(ARGV[2])
        local now = tonumber(ARGV[3])
        local ttl = tonumber(ARGV[4])
        
        local last = redis.call('HGET', key, 'last')
        local tokens = redis.call('HGET', key, 'tokens')
        
        if last == false then
            last = now
            tokens = burst
        else
            last = tonumber(last)
            tokens = tonumber(tokens)
            
            -- 计算新增令牌
            local elapsed = now - last
            local newTokens = math.min(burst, tokens + elapsed * rate)
            tokens = newTokens
        end
        
        -- 消费一个令牌
        if tokens >= 1 then
            tokens = tokens - 1
            redis.call('HMSET', key, 'last', now, 'tokens', tokens)
            redis.call('EXPIRE', key, ttl)
            return 1
        else
            return 0
        end
    `

	result, err := database.RedisClient.Eval(
		ctx,
		script,
		[]string{key},
		cfg.Rate,
		cfg.Burst,
		now,
		int(cfg.TTL.Seconds()),
	).Result()

	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}


