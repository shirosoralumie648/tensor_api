package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/cache"
)

// CachedAuthMiddleware 带缓存的认证中间件
// 使用多级缓存提升认证性能
func CachedAuthMiddleware(signingKey []byte, cacheManager *cache.CacheManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中提取令牌
		tokenString, err := extractTokenFromHeader(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// 解析令牌获取用户ID
		claims, err := ParseToken(tokenString, signingKey)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 使用缓存管理器获取用户信息
		// 这里假设将 userID 转换为整数
		userID := 0 // 从 claims.UserID 转换
		fmt.Sscanf(claims.UserID, "%d", &userID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userCache, err := cacheManager.GetUserCache(ctx, userID)
		if err != nil {
			// 如果缓存不可用，可以选择允许请求继续
			// 或者返回 401
			c.JSON(401, gin.H{"error": "failed to verify user"})
			c.Abort()
			return
		}

		// 检查用户状态
		if userCache.Status != 1 { // 假设 1 表示激活状态
			c.JSON(401, gin.H{"error": "user inactive"})
			c.Abort()
			return
		}

		// 检查配额
		if userCache.Quota <= 0 {
			c.JSON(429, gin.H{"error": "quota exceeded"})
			c.Abort()
			return
		}

		// 检查 Token 过期
		if time.Now().After(userCache.ExpireAt) {
			c.JSON(401, gin.H{"error": "token expired"})
			c.Abort()
			return
		}

		// 将用户信息和令牌存储在上下文中
		c.Set(UserIDKey, claims.UserID)
		c.Set(TokenKey, tokenString)
		c.Set("user_cache", userCache)

		// 记录认证时间用于性能分析
		c.Set("auth_time_ms", time.Now().UnixMilli())

		c.Next()
	}
}

// 认证方式提取函数

// ExtractTokenFromBearer 从 Bearer 令牌提取
func ExtractTokenFromBearer(c *gin.Context) (string, error) {
	return extractTokenFromHeader(c)
}

// ExtractTokenFromWebSocket 从 WebSocket 连接参数提取
func ExtractTokenFromWebSocket(c *gin.Context) (string, error) {
	token := c.Query("token")
	if token == "" {
		return "", fmt.Errorf("token parameter not found")
	}
	return token, nil
}

// ExtractTokenFromClaudeHeader 从 Claude 风格的请求头提取
func ExtractTokenFromClaudeHeader(c *gin.Context) (string, error) {
	token := c.GetHeader("x-api-key")
	if token == "" {
		return "", fmt.Errorf("x-api-key header not found")
	}
	return token, nil
}

// ExtractTokenFromGeminiHeader 从 Gemini 风格的请求头提取
func ExtractTokenFromGeminiHeader(c *gin.Context) (string, error) {
	token := c.GetHeader("x-goog-api-key")
	if token == "" {
		return "", fmt.Errorf("x-goog-api-key header not found")
	}
	return token, nil
}

// MultiAuthMiddleware 支持多种认证方式的中间件
// 按优先级尝试多种认证方式
func MultiAuthMiddleware(signingKey []byte, cacheManager *cache.CacheManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		var err error

		// 按优先级尝试多种认证方式
		authMethods := []func(*gin.Context) (string, error){
			ExtractTokenFromBearer,
			ExtractTokenFromClaudeHeader,
			ExtractTokenFromGeminiHeader,
			ExtractTokenFromWebSocket,
		}

		for _, method := range authMethods {
			tokenString, err = method(c)
			if err == nil && tokenString != "" {
				break
			}
		}

		if tokenString == "" {
			c.JSON(401, gin.H{"error": "no valid authentication method found"})
			c.Abort()
			return
		}

		// 解析并验证令牌
		claims, err := ParseToken(tokenString, signingKey)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 使用缓存验证用户
		userID := 0
		fmt.Sscanf(claims.UserID, "%d", &userID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userCache, err := cacheManager.GetUserCache(ctx, userID)
		if err != nil {
			c.JSON(401, gin.H{"error": "failed to verify user"})
			c.Abort()
			return
		}

		// 设置上下文
		c.Set(UserIDKey, claims.UserID)
		c.Set(TokenKey, tokenString)
		c.Set("user_cache", userCache)
		c.Set("auth_method", getAuthMethodName(method))

		c.Next()
	}
}

// 工具函数

// getAuthMethodName 获取认证方法的名称
func getAuthMethodName(method func(*gin.Context) (string, error)) string {
	switch fmt.Sprintf("%v", method) {
	case "ExtractTokenFromBearer":
		return "bearer"
	case "ExtractTokenFromClaudeHeader":
		return "claude"
	case "ExtractTokenFromGeminiHeader":
		return "gemini"
	case "ExtractTokenFromWebSocket":
		return "websocket"
	default:
		return "unknown"
	}
}

// GetCachedUserInfo 从上下文获取缓存的用户信息
func GetCachedUserInfo(c *gin.Context) (*cache.UserCache, error) {
	userCacheVal, ok := c.Get("user_cache")
	if !ok {
		return nil, fmt.Errorf("user cache not found in context")
	}

	userCache, ok := userCacheVal.(*cache.UserCache)
	if !ok {
		return nil, fmt.Errorf("invalid user cache type")
	}

	return userCache, nil
}

