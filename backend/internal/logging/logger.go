package logging

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Service    string                 `json:"service,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   float64                `json:"duration_ms,omitempty"`
	Error      string                 `json:"error,omitempty"`
	ErrorType  string                 `json:"error_type,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
}

// Logger 结构化日志记录器
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewLogger 创建新的 Logger
func NewLogger(development bool) (*Logger, error) {
	var config zap.Config

	if development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	}

	config.EncoderConfig.CallerKey = ""
	config.EncoderConfig.StacktraceKey = ""

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		logger: logger,
		sugar:  logger.Sugar(),
	}, nil
}

// WithContext 从 context 提取追踪信息
func (l *Logger) WithContext(ctx context.Context) *Logger {
	requestID := ctx.Value("request_id")
	if requestID != nil {
		l.logger = l.logger.With(zap.String("request_id", requestID.(string)))
	}

	userID := ctx.Value("user_id")
	if userID != nil {
		l.logger = l.logger.With(zap.String("user_id", userID.(string)))
	}

	traceID := ctx.Value("trace_id")
	if traceID != nil {
		l.logger = l.logger.With(zap.String("trace_id", traceID.(string)))
	}

	return l
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	zapFields := mapToZapFields(fields)
	l.logger.Info(msg, zapFields...)
}

// Warning 记录警告级别日志
func (l *Logger) Warning(msg string, fields map[string]interface{}) {
	zapFields := mapToZapFields(fields)
	l.logger.Warn(msg, zapFields...)
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	zapFields := mapToZapFields(fields)
	if err != nil {
		zapFields = append(zapFields, zap.Error(err))
	}
	l.logger.Error(msg, zapFields...)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	zapFields := mapToZapFields(fields)
	l.logger.Debug(msg, zapFields...)
}

// AuditLog 记录审计日志
func (l *Logger) AuditLog(userID, action string, resource string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"user_id":  userID,
		"action":   action,
		"resource": resource,
		"details":  details,
		"tags":     []string{"audit"},
	}
	l.Info("audit_action", fields)
}

// RequestLog 记录 HTTP 请求日志
func (l *Logger) RequestLog(method, path string, statusCode int, duration time.Duration, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	}

	for k, v := range fields {
		allFields[k] = v
	}

	if statusCode >= 400 {
		l.logger.Warn("http_request", mapToZapFields(allFields)...)
	} else {
		l.logger.Info("http_request", mapToZapFields(allFields)...)
	}
}

// ErrorLog 记录详细错误日志
func (l *Logger) ErrorLog(msg string, err error, errorType string, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"error_type": errorType,
	}

	for k, v := range fields {
		allFields[k] = v
	}

	l.Error(msg, err, allFields)
}

// PerformanceLog 记录性能日志
func (l *Logger) PerformanceLog(operation string, duration time.Duration, fields map[string]interface{}) {
	allFields := map[string]interface{}{
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
	}

	for k, v := range fields {
		allFields[k] = v
	}

	if duration > 100*time.Millisecond {
		l.logger.Warn("performance", mapToZapFields(allFields)...)
	}
}

// SecurityLog 记录安全相关日志
func (l *Logger) SecurityLog(event string, userID string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"event":   event,
		"user_id": userID,
		"details": details,
		"tags":    []string{"security"},
	}
	l.logger.Info("security_event", mapToZapFields(fields)...)
}

// Close 关闭 logger
func (l *Logger) Close() error {
	return l.logger.Sync()
}

// mapToZapFields 将 map 转换为 zap fields
func mapToZapFields(fields map[string]interface{}) []zapcore.Field {
	var zapFields []zapcore.Field

	for k, v := range fields {
		switch val := v.(type) {
		case string:
			zapFields = append(zapFields, zap.String(k, val))
		case int:
			zapFields = append(zapFields, zap.Int(k, val))
		case int64:
			zapFields = append(zapFields, zap.Int64(k, val))
		case float64:
			zapFields = append(zapFields, zap.Float64(k, val))
		case bool:
			zapFields = append(zapFields, zap.Bool(k, val))
		case error:
			zapFields = append(zapFields, zap.Error(val))
		default:
			// 序列化为 JSON
			if data, err := json.Marshal(val); err == nil {
				zapFields = append(zapFields, zap.ByteString(k, data))
			}
		}
	}

	return zapFields
}

// ==================== 全局兼容层 ====================

var DefaultLogger *Logger

// Init 初始化全局 Logger (兼容旧 API)
func Init(env string) error {
	var err error
	development := env == "development"
	DefaultLogger, err = NewLogger(development)
	return err
}

// Sync 同步日志 (兼容旧 API)
func Sync() {
	if DefaultLogger != nil {
		_ = DefaultLogger.Close()
	}
}

// Info 记录信息日志 (全局)
func Info(msg string, fields ...zapcore.Field) {
	if DefaultLogger != nil {
		DefaultLogger.logger.Info(msg, fields...)
	}
}

// Warn 记录警告日志 (全局)
func Warn(msg string, fields ...zapcore.Field) {
	if DefaultLogger != nil {
		DefaultLogger.logger.Warn(msg, fields...)
	}
}

// Error 记录错误日志 (全局)
func Error(msg string, fields ...zapcore.Field) {
	if DefaultLogger != nil {
		DefaultLogger.logger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误日志 (全局)
func Fatal(msg string, fields ...zapcore.Field) {
	if DefaultLogger != nil {
		DefaultLogger.logger.Fatal(msg, fields...)
	}
}

// Debug 记录调试日志 (全局)
func Debug(msg string, fields ...zapcore.Field) {
	if DefaultLogger != nil {
		DefaultLogger.logger.Debug(msg, fields...)
	}
}
