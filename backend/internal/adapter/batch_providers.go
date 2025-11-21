package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ==================== DeepSeek 适配器 ====================

// DeepSeekAdapter DeepSeek 适配器
type DeepSeekAdapter struct {
	*BaseAdapter
}

// NewDeepSeekAdapter 创建 DeepSeek 适配器
func NewDeepSeekAdapter(config *AdapterConfig) *DeepSeekAdapter {
	adapter := &DeepSeekAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"deepseek-coder", "deepseek-chat",
	})

	return adapter
}

// ConvertRequest 转换请求
func (da *DeepSeekAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	deepseekReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	return deepseekReq, nil
}

// DoRequest 发送请求
func (da *DeepSeekAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return da.DoHTTPRequest(ctx, "POST", "/chat/completions", convertedReq)
}

// ParseResponse 解析响应
func (da *DeepSeekAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var result OpenAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// ParseStreamResponse 解析流式响应
func (da *DeepSeekAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
		defer resp.Body.Close()
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (da *DeepSeekAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	response, ok := resp.(*OpenAIResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	return &response.Usage, nil
}

// GetError 获取错误
func (da *DeepSeekAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (da *DeepSeekAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== Moonshot 适配器 ====================

// MoonshotAdapter Moonshot 适配器
type MoonshotAdapter struct {
	*BaseAdapter
}

// NewMoonshotAdapter 创建 Moonshot 适配器
func NewMoonshotAdapter(config *AdapterConfig) *MoonshotAdapter {
	adapter := &MoonshotAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k",
	})

	return adapter
}

// ConvertRequest 转换请求
func (ma *MoonshotAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	moonshotReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	return moonshotReq, nil
}

// DoRequest 发送请求
func (ma *MoonshotAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return ma.DoHTTPRequest(ctx, "POST", "/v1/chat/completions", convertedReq)
}

// ParseResponse 解析响应
func (ma *MoonshotAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var result OpenAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// ParseStreamResponse 解析流式响应
func (ma *MoonshotAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (ma *MoonshotAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	return &Usage{}, nil
}

// GetError 获取错误
func (ma *MoonshotAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (ma *MoonshotAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== MiniMax 适配器 ====================

// MinimaxAdapter MiniMax 适配器
type MinimaxAdapter struct {
	*BaseAdapter
}

// NewMinimaxAdapter 创建 MiniMax 适配器
func NewMinimaxAdapter(config *AdapterConfig) *MinimaxAdapter {
	adapter := &MinimaxAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"abab6.5-chat", "abab6.5s-chat", "abab5.5s-chat",
	})

	return adapter
}

// ConvertRequest 转换请求
func (mma *MinimaxAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	minimaxReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	return minimaxReq, nil
}

// DoRequest 发送请求
func (mma *MinimaxAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return mma.DoHTTPRequest(ctx, "POST", "/chat/completions", convertedReq)
}

// ParseResponse 解析响应
func (mma *MinimaxAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var result OpenAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// ParseStreamResponse 解析流式响应
func (mma *MinimaxAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (mma *MinimaxAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	return &Usage{}, nil
}

// GetError 获取错误
func (mma *MinimaxAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (mma *MinimaxAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== 通用适配器 ====================

// GenericAdapter 通用适配器
type GenericAdapter struct {
	*BaseAdapter
	mapping *AdapterMapping
}

// AdapterMapping 适配器映射配置
type AdapterMapping struct {
	// 请求映射
	RequestMapping map[string]string

	// 响应映射
	ResponseMapping map[string]string

	// 字段映射
	FieldMapping map[string]string

	// 自定义转换函数
	CustomConvert func(*OpenAIRequest) (interface{}, error)
}

// NewGenericAdapter 创建通用适配器
func NewGenericAdapter(config *AdapterConfig, mapping *AdapterMapping) *GenericAdapter {
	adapter := &GenericAdapter{
		BaseAdapter: NewBaseAdapter(config),
		mapping:     mapping,
	}

	return adapter
}

// ConvertRequest 转换请求
func (ga *GenericAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	if ga.mapping.CustomConvert != nil {
		return ga.mapping.CustomConvert(req)
	}

	// 默认转换逻辑
	return req, nil
}

// DoRequest 发送请求
func (ga *GenericAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	path := "/chat/completions"
	if pathMapping, ok := ga.mapping.RequestMapping["path"]; ok {
		path = pathMapping
	}

	return ga.DoHTTPRequest(ctx, "POST", path, convertedReq)
}

// ParseResponse 解析响应
func (ga *GenericAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var result OpenAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// ParseStreamResponse 解析流式响应
func (ga *GenericAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (ga *GenericAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	return &Usage{}, nil
}

// GetError 获取错误
func (ga *GenericAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (ga *GenericAdapter) HealthCheck(ctx context.Context) error {
	return nil
}


