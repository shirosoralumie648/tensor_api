package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/cache"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	signingKey    []byte
	cacheManager  *cache.CacheManager
	factory       *AuthExtractorFactory
	allowedMethods map[AuthMethod]bool
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(
	signingKey []byte,
	cacheManager *cache.CacheManager,
) *AuthHandler {
	return &AuthHandler{
		signingKey:    signingKey,
		cacheManager:  cacheManager,
		factory:       GetDefaultFactory(),
		allowedMethods: map[AuthMethod]bool{
			AuthMethodBearer:     true,
			AuthMethodClaude:     true,
			AuthMethodGemini:     true,
			AuthMethodWebSocket:  true,
		},
	}
}

// EnableAuthMethod 启用认证方法
func (ah *AuthHandler) EnableAuthMethod(method AuthMethod, enabled bool) {
	ah.allowedMethods[method] = enabled
}

// RegisterExtractor 注册自定义提取器
func (ah *AuthHandler) RegisterExtractor(method AuthMethod, extractor TokenExtractor) {
	ah.factory.RegisterExtractor(method, extractor)
	ah.allowedMethods[method] = true
}

// Authenticate 执行认证
// 返回: (userID, tokenHash, authMethod, error)
func (ah *AuthHandler) Authenticate(c *gin.Context) (int, string, AuthMethod, error) {
	// 1. 提取 Token
	tokenHash, authMethod, err := ah.factory.ExtractToken(c)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to extract token: %w", err)
	}

	// 2. 检查认证方法是否启用
	if !ah.allowedMethods[authMethod] {
		return 0, "", authMethod, fmt.Errorf("auth method %s is not enabled", authMethod)
	}

	// 3. 验证 Token 签名（对于 Bearer Token）
	if authMethod == AuthMethodBearer {
		// Bearer Token 需要通过 JWT 验证
		// 这里假设 tokenHash 实际上是 JWT token
		claims, err := ParseToken(tokenHash, ah.signingKey)
		if err != nil {
			return 0, "", authMethod, fmt.Errorf("invalid token: %w", err)
		}

		// 从 claims 中提取用户 ID
		userID := 0
		fmt.Sscanf(claims.UserID, "%d", &userID)
		if userID == 0 {
			return 0, "", authMethod, fmt.Errorf("invalid user id in token")
		}

		return userID, tokenHash, authMethod, nil
	}

	// 4. 对于其他方法，假设 tokenHash 是实际的 API token hash
	// 这里需要从数据库查询来验证 token（这部分由 repository 层实现）
	return 0, tokenHash, authMethod, nil
}

// ValidateAndCacheUser 验证并缓存用户信息
func (ah *AuthHandler) ValidateAndCacheUser(
	c *gin.Context,
	userID int,
	tokenHash string,
	authMethod AuthMethod,
) (*cache.UserCache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 从缓存获取用户信息
	userCache, err := ah.cacheManager.GetUserCache(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache: %w", err)
	}

	// 检查用户状态
	if userCache.Status != 1 { // 1 = 激活
		return nil, fmt.Errorf("user is not active, status: %d", userCache.Status)
	}

	// 检查 Token 过期
	if time.Now().After(userCache.ExpireAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// 检查配额
	if userCache.Quota <= 0 {
		return nil, fmt.Errorf("user quota exhausted")
	}

	return userCache, nil
}

// HandleAuth 处理认证的中间件函数
func (ah *AuthHandler) HandleAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 执行认证
		userID, tokenHash, authMethod, err := ah.Authenticate(c)
		if err != nil {
			c.JSON(401, gin.H{
				"error":  "Unauthorized",
				"detail": err.Error(),
			})
			c.Abort()
			return
		}

		// 2. 验证并获取用户缓存
		userCache, err := ah.ValidateAndCacheUser(c, userID, tokenHash, authMethod)
		if err != nil {
			c.JSON(401, gin.H{
				"error":  "Unauthorized",
				"detail": err.Error(),
			})
			c.Abort()
			return
		}

		// 3. 将认证信息存储在上下文中
		c.Set("user_id", userID)
		c.Set("user_cache", userCache)
		c.Set("token_hash", tokenHash)
		c.Set("auth_method", authMethod.String())
		c.Set("auth_time", time.Now().UnixMilli())

		// 4. 继续处理请求
		c.Next()
	}
}

// HandleAuthOptional 可选的认证中间件（不会拒绝请求，但会设置用户信息）
func (ah *AuthHandler) HandleAuthOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试认证，但如果失败不会阻止请求
		userID, tokenHash, authMethod, err := ah.Authenticate(c)
		if err != nil {
			// 认证失败，但继续处理（作为匿名用户）
			c.Set("user_id", 0)
			c.Set("auth_method", "anonymous")
			c.Next()
			return
		}

		// 认证成功，验证用户
		userCache, err := ah.ValidateAndCacheUser(c, userID, tokenHash, authMethod)
		if err != nil {
			// 用户验证失败，但继续处理（作为匿名用户）
			c.Set("user_id", 0)
			c.Set("auth_method", "anonymous")
			c.Next()
			return
		}

		// 设置用户信息
		c.Set("user_id", userID)
		c.Set("user_cache", userCache)
		c.Set("token_hash", tokenHash)
		c.Set("auth_method", authMethod.String())
		c.Set("auth_time", time.Now().UnixMilli())

		c.Next()
	}
}

// GetAuthInfo 从上下文获取认证信息
func GetAuthInfo(c *gin.Context) (userID int, userCache *cache.UserCache, authMethod AuthMethod, ok bool) {
	userIDVal, ok := c.Get("user_id")
	if !ok {
		return 0, nil, "", false
	}

	userID, ok = userIDVal.(int)
	if !ok {
		return 0, nil, "", false
	}

	userCacheVal, ok := c.Get("user_cache")
	if !ok {
		return userID, nil, "", false
	}

	userCache, ok = userCacheVal.(*cache.UserCache)
	if !ok {
		return userID, nil, "", false
	}

	authMethodVal, ok := c.Get("auth_method")
	if !ok {
		return userID, userCache, "", true
	}

	authMethodStr, ok := authMethodVal.(string)
	if !ok {
		return userID, userCache, "", true
	}

	authMethod = AuthMethod(authMethodStr)
	return userID, userCache, authMethod, true
}

// RequireAuth 验证必须已认证（包装的中间件）
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _, _, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAuthMethod 验证特定的认证方法
func RequireAuthMethod(allowedMethods ...AuthMethod) gin.HandlerFunc {
	allowedSet := make(map[AuthMethod]bool)
	for _, method := range allowedMethods {
		allowedSet[method] = true
	}

	return func(c *gin.Context) {
		userID, _, authMethod, ok := GetAuthInfo(c)
		if !ok || userID == 0 {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		if !allowedSet[authMethod] {
			c.JSON(403, gin.H{
				"error": "Forbidden: auth method not allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// String 返回字符串表示
func (am AuthMethod) String() string {
	return string(am)
}

