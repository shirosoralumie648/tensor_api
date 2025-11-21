package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
)

// SecurityHeaders 添加安全响应头
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HSTS (HTTP Strict-Transport-Security)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// XSS 保护
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// CSP (Content Security Policy)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:")

		// 其他安全头
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// 禁用缓存（敏感数据）
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(requestsPerSecond float64) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(requestsPerSecond, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Hour,
	})

	// 为 API 密钥和用户 ID 设置自定义限制
	lmt.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})

	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": httpError.RetryAfter,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证 Content-Type
		contentType := c.ContentType()
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if contentType != "application/json" && contentType != "application/x-www-form-urlencoded" {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "unsupported content type",
				})
				c.Abort()
				return
			}
		}

		// 验证请求大小（防止大请求 DoS）
		maxRequestSize := int64(10 * 1024 * 1024) // 10MB
		if c.Request.ContentLength > maxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request too large",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestSignatureVerification 请求签名验证中间件
func RequestSignatureVerification(secret string, enabledPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查路径是否需要验证
		needsVerification := false
		for _, path := range enabledPaths {
			if c.Request.URL.Path == path {
				needsVerification = true
				break
			}
		}

		if !needsVerification {
			c.Next()
			return
		}

		// 获取签名
		signature := c.GetHeader("X-Signature")
		timestamp := c.GetHeader("X-Timestamp")

		if signature == "" || timestamp == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing signature or timestamp",
			})
			c.Abort()
			return
		}

		// 验证时间戳（防止重放攻击）
		ts := parseTimestamp(timestamp)
		if time.Since(ts) > 5*time.Minute {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "request expired",
			})
			c.Abort()
			return
		}

		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "failed to read request body",
			})
			c.Abort()
			return
		}

		// 恢复请求体
		c.Request.Body = io.NopCloser(io.MultiReader(
			io.Reader(nil),
			io.Reader(nil),
		))
		c.Request.Body = io.NopCloser(io.MultiReader(
			io.NopCloser(io.MultiReader()),
		))

		// 验证签名
		expectedSignature := generateSignature(string(body), timestamp, secret)
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid signature",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SQLInjectionProtection SQL 注入保护
func SQLInjectionProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查查询参数和路径参数中的 SQL 关键字
		dangerousKeywords := []string{"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE", "UNION", "SELECT", ";", "--"}

		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				for _, keyword := range dangerousKeywords {
					if contains(value, keyword) {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": fmt.Sprintf("invalid parameter value in %s", key),
						})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// XSSProtection XSS 保护
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		dangerousPatterns := []string{
			"<script",
			"javascript:",
			"onerror=",
			"onload=",
			"onclick=",
			"eval(",
			"iframe",
		}

		// 检查请求 body 中的危险模式
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			body, _ := io.ReadAll(c.Request.Body)
			bodyStr := string(body)

			for _, pattern := range dangerousPatterns {
				if contains(bodyStr, pattern) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "potentially dangerous content detected",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// TLSEnforcementMiddleware 强制 TLS
func TLSEnforcementMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "HTTPS required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware 添加请求 ID 用于追踪
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware 安全的日志记录
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// 不记录敏感字段
		sensitiveFields := map[string]bool{
			"password":    true,
			"api_key":     true,
			"token":       true,
			"secret":      true,
			"credit_card": true,
		}

		query := c.Request.URL.Query()
		sanitizedQuery := make(map[string][]string)
		for key, values := range query {
			if sensitiveFields[key] {
				sanitizedQuery[key] = []string{"***"}
			} else {
				sanitizedQuery[key] = values
			}
		}

		// 记录请求信息
		duration := time.Since(startTime)
		_ = fmt.Sprintf(
			"[%s] %s %s %d (%dms)",
			c.GetString("request_id"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.Milliseconds(),
		)
	}
}

// CORS 配置安全版本
func SecureCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 白名单检查
		allowedOrigins := map[string]bool{
			"https://oblivious.com": true,
			"https://app.oblivious.com": true,
		}

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Signature, X-Timestamp")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ============== 辅助函数 ==============

// generateSignature 生成签名
func generateSignature(body, timestamp, secret string) string {
	message := fmt.Sprintf("%s.%s", body, timestamp)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// parseTimestamp 解析时间戳
func parseTimestamp(ts string) time.Time {
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t
	}
	return time.Now()
}

// contains 不区分大小写的包含检查
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// generateRequestID 生成唯一的请求 ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ValidateTLSConfig 验证 TLS 配置
func ValidateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	return tlsConfig, nil
}

