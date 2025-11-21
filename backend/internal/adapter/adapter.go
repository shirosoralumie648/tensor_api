package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// OpenAIRequest OpenAI 标准请求格式
type OpenAIRequest struct {
	Model            string                 `json:"model"`
	Messages         []Message              `json:"messages"`
	Temperature      float32                `json:"temperature,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	TopP             float32                `json:"top_p,omitempty"`
	FrequencyPenalty float32                `json:"frequency_penalty,omitempty"`
	PresencePenalty  float32                `json:"presence_penalty,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	User             string                 `json:"user,omitempty"`
	Extra            map[string]interface{} `json:"extra,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
	Name    string      `json:"name,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// OpenAIResponse OpenAI 标准响应格式
type OpenAIResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Created int64      `json:"created"`
	Model   string     `json:"model"`
	Choices []Choice   `json:"choices"`
	Usage   Usage      `json:"usage"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

// Choice 完成选择
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
	Delta        *Message `json:"delta,omitempty"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	ID      string   `json:"id,omitempty"`
	Object  string   `json:"object,omitempty"`
	Created int64    `json:"created,omitempty"`
	Model   string   `json:"model,omitempty"`
	Choices []Choice `json:"choices,omitempty"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// AdapterConfig 适配器配置
type AdapterConfig struct {
	// API 类型
	Type string

	// API 基础 URL
	BaseURL string

	// API 密钥
	APIKey string

	// API 版本（可选）
	Version string

	// 超时时间
	Timeout time.Duration

	// 额外配置
	Extra map[string]interface{}
}

// Adapter 适配器接口
type Adapter interface {
	// 获取名称
	Name() string

	// 获取支持的模型列表
	GetSupportedModels() []string

	// 转换请求格式
	ConvertRequest(req *OpenAIRequest) (interface{}, error)

	// 发送请求
	DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error)

	// 解析响应
	ParseResponse(resp *http.Response) (*OpenAIResponse, error)

	// 解析流式响应
	ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error)

	// 提取使用量
	ExtractUsage(resp interface{}) (*Usage, error)

	// 获取错误信息
	GetError(resp *http.Response) error

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// BaseAdapter 基础适配器
type BaseAdapter struct {
	config          *AdapterConfig
	httpClient      *http.Client
	supportedModels []string
	mu              sync.RWMutex
}

// NewBaseAdapter 创建基础适配器
func NewBaseAdapter(config *AdapterConfig) *BaseAdapter {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &BaseAdapter{
		config:     config,
		httpClient: client,
	}
}

// Name 获取名称
func (ba *BaseAdapter) Name() string {
	return ba.config.Type
}

// GetSupportedModels 获取支持的模型
func (ba *BaseAdapter) GetSupportedModels() []string {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	models := make([]string, len(ba.supportedModels))
	copy(models, ba.supportedModels)

	return models
}

// SetSupportedModels 设置支持的模型
func (ba *BaseAdapter) SetSupportedModels(models []string) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	ba.supportedModels = models
}

// NewRequest 创建 HTTP 请求
func (ba *BaseAdapter) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	url := ba.config.BaseURL + path

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %v", err)
		}
	}

	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加认证
	ba.addAuthHeader(req)

	return req, nil
}

// addAuthHeader 添加认证头
func (ba *BaseAdapter) addAuthHeader(req *http.Request) {
	if ba.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ba.config.APIKey))
	}
}

// DoHTTPRequest 执行 HTTP 请求
func (ba *BaseAdapter) DoHTTPRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	req, err := ba.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	resp, err := ba.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}

	return resp, nil
}

// AdapterMetrics 适配器指标
type AdapterMetrics struct {
	// 总请求数
	TotalRequests int64

	// 成功请求数
	SuccessRequests int64

	// 失败请求数
	FailedRequests int64

	// 平均响应时间（毫秒）
	AvgResponseTime int64

	// 最后请求时间
	LastRequestTime time.Time
}

// AdapterInfo 适配器信息
type AdapterInfo struct {
	// 适配器名称
	Name string

	// 适配器类型
	Type string

	// 支持的模型数
	SupportedModelCount int

	// 版本
	Version string

	// 状态
	Status string

	// 指标
	Metrics AdapterMetrics

	// 创建时间
	CreatedAt time.Time
}

// AdapterError 适配器错误
type AdapterError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error 实现 error 接口
func (ae *AdapterError) Error() string {
	return fmt.Sprintf("adapter error [%s]: %s", ae.Code, ae.Message)
}

// NewAdapterError 创建适配器错误
func NewAdapterError(code, message string) *AdapterError {
	return &AdapterError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// DefaultLogFunc 默认日志函数
func DefaultLogFunc(level, msg string, args ...interface{}) {
	// 默认实现：忽略日志
}
