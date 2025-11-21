package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 统一响应结构
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type ErrorInfo struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// 错误码定义
const (
	ErrInternal          = 1000
	ErrInvalidRequest    = 1001
	ErrNotFound          = 1004
	ErrUnauthorized      = 2001
	ErrForbidden         = 2003
	ErrInvalidToken      = 2010
	ErrTokenExpired      = 2011
	ErrInsufficientQuota = 3001
	ErrModelNotAvailable = 3002
	ErrRateLimitExceeded = 3003
)

var errorMessages = map[int]string{
	ErrInternal:          "内部服务器错误",
	ErrInvalidRequest:    "请求参数错误",
	ErrNotFound:          "资源不存在",
	ErrUnauthorized:      "未登录",
	ErrForbidden:         "无权限访问",
	ErrInvalidToken:      "Token 无效",
	ErrTokenExpired:      "Token 已过期",
	ErrInsufficientQuota: "余额不足",
	ErrModelNotAvailable: "模型不可用",
	ErrRateLimitExceeded: "请求频率超限",
}

// Success 成功响应
func Success(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "操作成功"
	}
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, errCode int, customMessage string, details interface{}) {
	message := errorMessages[errCode]
	if customMessage != "" {
		message = customMessage
	}

	c.JSON(httpStatus, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// 快捷方法
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrInvalidRequest, message, nil)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, ErrUnauthorized, message, nil)
}

func Forbidden(c *gin.Context) {
	Error(c, http.StatusForbidden, ErrForbidden, "", nil)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, ErrNotFound, message, nil)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, ErrInternal, message, nil)
}


