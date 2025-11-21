package relay

import (
	"context"
	"io"
)

// RequestType 请求类型
type RequestType int

const (
	// Chat 请求
	RequestTypeChat RequestType = iota
	// Embedding 请求
	RequestTypeEmbedding
	// Image 请求
	RequestTypeImage
	// Audio 请求
	RequestTypeAudio
)

// String 返回请求类型的字符串表示
func (rt RequestType) String() string {
	switch rt {
	case RequestTypeChat:
		return "chat"
	case RequestTypeEmbedding:
		return "embedding"
	case RequestTypeImage:
		return "image"
	case RequestTypeAudio:
		return "audio"
	default:
		return "unknown"
	}
}

// HandlerRequest 处理器请求
type HandlerRequest struct {
	// 请求类型
	Type RequestType

	// 请求 ID
	ID string

	// 用户 ID
	UserID int

	// Token ID
	TokenID int

	// 模型名称
	Model string

	// API 端点
	Endpoint string

	// 请求头
	Headers map[string]string

	// 请求体（原始字节）
	Body []byte

	// 请求体读取器（用于流式）
	BodyReader io.Reader

	// 缓存 ID（如果启用了缓存）
	CacheID string

	// 上下文（用于取消、超时等）
	Context context.Context

	// 元数据（扩展信息）
	Metadata map[string]interface{}
}

// HandlerResponse 处理器响应
type HandlerResponse struct {
	// 状态码
	StatusCode int

	// 响应头
	Headers map[string]string

	// 响应体（原始字节）
	Body []byte

	// 响应体读取器（用于流式）
	BodyReader io.Reader

	// 错误信息
	Error string

	// 元数据
	Metadata map[string]interface{}
}

// RelayHandler 中继处理器接口
type RelayHandler interface {
	// 获取处理器类型
	GetType() RequestType

	// 获取处理器名称
	GetName() string

	// 是否支持流式处理
	SupportsStreaming() bool

	// 处理请求（同步）
	Handle(ctx context.Context, req *HandlerRequest) (*HandlerResponse, error)

	// 处理流式请求（异步）
	HandleStream(ctx context.Context, req *HandlerRequest, responseCh chan *HandlerResponse) error

	// 验证请求（在处理前）
	ValidateRequest(req *HandlerRequest) error

	// 验证响应（在返回前）
	ValidateResponse(resp *HandlerResponse) error

	// 获取处理统计
	GetStatistics() map[string]interface{}

	// 重置统计
	ResetStatistics()
}

// BaseRelayHandler 基础处理器（实现通用逻辑）
type BaseRelayHandler struct {
	// 处理器类型
	handlerType RequestType

	// 处理器名称
	name string

	// 是否支持流式
	supportsStreaming bool

	// 统计信息
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	totalBytes         int64
}

// NewBaseRelayHandler 创建新的基础处理器
func NewBaseRelayHandler(handlerType RequestType, name string, supportsStreaming bool) *BaseRelayHandler {
	return &BaseRelayHandler{
		handlerType:       handlerType,
		name:              name,
		supportsStreaming: supportsStreaming,
	}
}

// GetType 获取处理器类型
func (brh *BaseRelayHandler) GetType() RequestType {
	return brh.handlerType
}

// GetName 获取处理器名称
func (brh *BaseRelayHandler) GetName() string {
	return brh.name
}

// SupportsStreaming 是否支持流式处理
func (brh *BaseRelayHandler) SupportsStreaming() bool {
	return brh.supportsStreaming
}

// ValidateRequest 验证请求（可被重写）
func (brh *BaseRelayHandler) ValidateRequest(req *HandlerRequest) error {
	if req == nil {
		return ErrInvalidRequest{Message: "request is nil"}
	}
	if req.Type != brh.handlerType {
		return ErrInvalidRequest{Message: "request type mismatch"}
	}
	if req.Model == "" {
		return ErrInvalidRequest{Message: "model is required"}
	}
	if req.Endpoint == "" {
		return ErrInvalidRequest{Message: "endpoint is required"}
	}
	return nil
}

// ValidateResponse 验证响应（可被重写）
func (brh *BaseRelayHandler) ValidateResponse(resp *HandlerResponse) error {
	if resp == nil {
		return ErrInvalidResponse{Message: "response is nil"}
	}
	if resp.StatusCode < 0 || resp.StatusCode >= 600 {
		return ErrInvalidResponse{Message: "invalid status code"}
	}
	return nil
}

// GetStatistics 获取统计信息
func (brh *BaseRelayHandler) GetStatistics() map[string]interface{} {
	successRate := 0.0
	if brh.totalRequests > 0 {
		successRate = float64(brh.successfulRequests) / float64(brh.totalRequests) * 100
	}

	return map[string]interface{}{
		"handler_type":        brh.handlerType.String(),
		"handler_name":        brh.name,
		"total_requests":      brh.totalRequests,
		"successful_requests": brh.successfulRequests,
		"failed_requests":     brh.failedRequests,
		"success_rate":        successRate,
		"total_bytes":         brh.totalBytes,
		"supports_streaming":  brh.supportsStreaming,
	}
}

// ResetStatistics 重置统计
func (brh *BaseRelayHandler) ResetStatistics() {
	brh.totalRequests = 0
	brh.successfulRequests = 0
	brh.failedRequests = 0
	brh.totalBytes = 0
}

// RecordSuccess 记录成功请求
func (brh *BaseRelayHandler) RecordSuccess(bodySize int64) {
	brh.totalRequests++
	brh.successfulRequests++
	brh.totalBytes += bodySize
}

// RecordFailure 记录失败请求
func (brh *BaseRelayHandler) RecordFailure() {
	brh.totalRequests++
	brh.failedRequests++
}

// 错误类型
type ErrInvalidRequest struct {
	Message string
}

func (e ErrInvalidRequest) Error() string {
	return "invalid request: " + e.Message
}

type ErrInvalidResponse struct {
	Message string
}

func (e ErrInvalidResponse) Error() string {
	return "invalid response: " + e.Message
}

type ErrUnsupportedType struct {
	Type RequestType
}

func (e ErrUnsupportedType) Error() string {
	return "unsupported request type: " + e.Type.String()
}
