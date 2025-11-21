package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "X-Request-ID"

// RequestIDMiddleware 为每个请求生成唯一 ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 Header 获取 Request ID
		requestID := c.GetHeader(RequestIDKey)
		
		// 如果没有，生成新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到 Context 和响应 Header
		c.Set("request_id", requestID)
		c.Header(RequestIDKey, requestID)

		c.Next()
	}
}

