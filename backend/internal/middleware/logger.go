package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/shirosoralumie648/Oblivious/backend/internal/logging"
	"go.uber.org/zap"
)

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录日志
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// 如果有用户信息，添加到日志
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.Int("user_id", userID.(int)))
		}

		// 如果有错误，记录错误
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request", fields...)
		}
	}
}

