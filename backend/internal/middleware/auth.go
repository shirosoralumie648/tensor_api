package middleware

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// UserIDKey 用户 ID 在上下文中的键
	UserIDKey = "user_id"
	// TokenKey 令牌在上下文中的键
	TokenKey = "token"
)

// Claims JWT 声明
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ExtractUserID 从上下文中提取用户 ID
func ExtractUserID(c *gin.Context) (string, error) {
	userID, ok := c.Get(UserIDKey)
	if !ok {
		return "", errors.New("user id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user id type")
	}

	return userIDStr, nil
}

// ExtractToken 从上下文中提取令牌
func ExtractToken(c *gin.Context) (string, error) {
	token, ok := c.Get(TokenKey)
	if !ok {
		return "", errors.New("token not found in context")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return "", errors.New("invalid token type")
	}

	return tokenStr, nil
}

// ParseToken 解析 JWT 令牌
func ParseToken(tokenString string, signingKey []byte) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(signingKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取令牌
		tokenString, err := extractTokenFromHeader(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// 解析令牌
		claims, err := ParseToken(tokenString, signingKey)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 将用户 ID 存储在上下文中
		c.Set(UserIDKey, claims.UserID)
		c.Set(TokenKey, tokenString)

		c.Next()
	}
}

// extractTokenFromHeader 从请求头中提取令牌
func extractTokenFromHeader(c *gin.Context) (string, error) {
	const bearerSchema = "Bearer "

	// 从 Authorization 请求头获取令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	// 检查是否以 "Bearer " 开头
	if len(authHeader) <= len(bearerSchema) || authHeader[:len(bearerSchema)] != bearerSchema {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerSchema):], nil
}

// APIKeyAuthMiddleware API 密钥认证中间件
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 API 密钥
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// 尝试从查询参数中获取
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(401, gin.H{"error": "api key required"})
			c.Abort()
			return
		}

		// TODO: 验证 API 密钥（应该从数据库中查询）
		// 这里仅作示例

		c.Set("api_key", apiKey)
		c.Next()
	}
}

// RateLimitMiddleware 和 CORSMiddleware 已在其他文件中定义
