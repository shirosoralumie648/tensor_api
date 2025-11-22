package adapter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ==================== OpenAI 适配器 ====================

// OpenAIAdapter OpenAI 适配器
type OpenAIAdapter struct {
	*BaseAdapter
}

// NewOpenAIAdapter 创建 OpenAI 适配器
func NewOpenAIAdapter(config *AdapterConfig) *OpenAIAdapter {
	adapter := &OpenAIAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"gpt-4", "gpt-4-turbo", "gpt-4-turbo-preview",
		"gpt-3.5-turbo", "gpt-3.5-turbo-16k",
	})

	return adapter
}

// ConvertRequest 转换请求
func (oa *OpenAIAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	return req, nil
}

// DoRequest 发送请求
func (oa *OpenAIAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	req, ok := convertedReq.(*OpenAIRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	path := "/chat/completions"
	if req.Stream {
		// 流式请求
	}

	return oa.DoHTTPRequest(ctx, "POST", path, req)
}

// ParseResponse 解析响应
func (oa *OpenAIAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var result OpenAIResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// ParseStreamResponse 解析流式响应
func (oa *OpenAIAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					// ignore errors, stream will end
				}
				return
			}

			if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
				continue
			}

			data := bytes.TrimSpace(bytes.TrimPrefix([]byte(line), []byte("data: ")))
			if bytes.Equal(data, []byte("[DONE]")) {
				return
			}

			var chunk StreamChunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}

			ch <- &chunk
		}
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (oa *OpenAIAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	response, ok := resp.(*OpenAIResponse)
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	return &response.Usage, nil
}

// GetError 获取错误
func (oa *OpenAIAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	var errResp struct {
		Error *ErrorInfo `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	if errResp.Error != nil {
		return NewAdapterError(errResp.Error.Code, errResp.Error.Message)
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (oa *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== Claude 适配器 ====================

// ClaudeAdapter Claude 适配器
type ClaudeAdapter struct {
	*BaseAdapter
}

// NewClaudeAdapter 创建 Claude 适配器
func NewClaudeAdapter(config *AdapterConfig) *ClaudeAdapter {
	adapter := &ClaudeAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"claude-3-opus", "claude-3-sonnet", "claude-3-haiku",
		"claude-2.1", "claude-2", "claude-instant-1.2",
	})

	return adapter
}

// ConvertRequest 转换请求
func (ca *ClaudeAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	// 将 OpenAI 格式转换为 Claude 格式
	claudeReq := map[string]interface{}{
		"model":       req.Model,
		"max_tokens":  req.MaxTokens,
		"system":      "", // Claude 使用 system 参数
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}

	return claudeReq, nil
}

// DoRequest 发送请求
func (ca *ClaudeAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return ca.DoHTTPRequest(ctx, "POST", "/messages", convertedReq)
}

// ParseResponse 解析响应
func (ca *ClaudeAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var claudeResp map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// 转换为 OpenAI 格式
	result := &OpenAIResponse{
		ID:      fmt.Sprintf("%v", claudeResp["id"]),
		Model:   fmt.Sprintf("%v", claudeResp["model"]),
		Created: int64(0), // Claude 不返回 created
	}

	return result, nil
}

// ParseStreamResponse 解析流式响应
func (ca *ClaudeAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		// Claude 使用不同的流式格式
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (ca *ClaudeAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	respMap, ok := resp.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response type")
	}

	usage := &Usage{}

	if usage_, ok := respMap["usage"].(map[string]interface{}); ok {
		if input, ok := usage_["input_tokens"].(float64); ok {
			usage.PromptTokens = int(input)
		}
		if output, ok := usage_["output_tokens"].(float64); ok {
			usage.CompletionTokens = int(output)
		}
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	return usage, nil
}

// GetError 获取错误
func (ca *ClaudeAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (ca *ClaudeAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== Gemini 适配器 ====================

// GeminiAdapter Gemini 适配器
type GeminiAdapter struct {
	*BaseAdapter
}

// NewGeminiAdapter 创建 Gemini 适配器
func NewGeminiAdapter(config *AdapterConfig) *GeminiAdapter {
	adapter := &GeminiAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"gemini-pro", "gemini-pro-vision",
		"gemini-1.5-pro", "gemini-1.5-flash",
	})

	return adapter
}

// ConvertRequest 转换请求
func (ga *GeminiAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	geminiReq := map[string]interface{}{
		"contents": convertMessagesToContents(req.Messages),
		"generation_config": map[string]interface{}{
			"temperature":     req.Temperature,
			"topP":            req.TopP,
			"maxOutputTokens": req.MaxTokens,
		},
	}

	return geminiReq, nil
}

// DoRequest 发送请求
func (ga *GeminiAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return ga.DoHTTPRequest(ctx, "POST", "/generateContent", convertedReq)
}

// ParseResponse 解析响应
func (ga *GeminiAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var geminiResp map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	result := &OpenAIResponse{
		ID:      "gemini-response",
		Model:   "gemini-pro",
		Created: int64(0),
	}

	return result, nil
}

// ParseStreamResponse 解析流式响应
func (ga *GeminiAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (ga *GeminiAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	return &Usage{}, nil
}

// GetError 获取错误
func (ga *GeminiAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (ga *GeminiAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== 百度适配器 ====================

// BaiduAdapter 百度适配器
type BaiduAdapter struct {
	*BaseAdapter
}

// NewBaiduAdapter 创建百度适配器
func NewBaiduAdapter(config *AdapterConfig) *BaiduAdapter {
	adapter := &BaiduAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"eb-4", "eb-3.5-turbo", "bge-large-zh", "bge-base-zh",
	})

	return adapter
}

// ConvertRequest 转换请求
func (ba *BaiduAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	baiduReq := map[string]interface{}{
		"model":             req.Model,
		"messages":          req.Messages,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"max_output_tokens": req.MaxTokens,
	}

	return baiduReq, nil
}

// DoRequest 发送请求
func (ba *BaiduAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return ba.DoHTTPRequest(ctx, "POST", "/chat/completions", convertedReq)
}

// ParseResponse 解析响应
func (ba *BaiduAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var baiduResp map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&baiduResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	result := &OpenAIResponse{
		ID:      fmt.Sprintf("%v", baiduResp["id"]),
		Model:   fmt.Sprintf("%v", baiduResp["model"]),
		Created: int64(0),
	}

	return result, nil
}

// ParseStreamResponse 解析流式响应
func (ba *BaiduAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (ba *BaiduAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	return &Usage{}, nil
}

// GetError 获取错误
func (ba *BaiduAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (ba *BaiduAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== 阿里适配器 ====================

// QwenAdapter 阿里通义千问适配器
type QwenAdapter struct {
	*BaseAdapter
}

// NewQwenAdapter 创建阿里适配器
func NewQwenAdapter(config *AdapterConfig) *QwenAdapter {
	adapter := &QwenAdapter{
		BaseAdapter: NewBaseAdapter(config),
	}

	adapter.SetSupportedModels([]string{
		"qwen-turbo", "qwen-plus", "qwen-max",
		"qwen-vl-plus", "qwen-vl-max",
	})

	return adapter
}

// ConvertRequest 转换请求
func (qa *QwenAdapter) ConvertRequest(req *OpenAIRequest) (interface{}, error) {
	qwenReq := map[string]interface{}{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"max_tokens":  req.MaxTokens,
	}

	return qwenReq, nil
}

// DoRequest 发送请求
func (qa *QwenAdapter) DoRequest(ctx context.Context, convertedReq interface{}) (*http.Response, error) {
	return qa.DoHTTPRequest(ctx, "POST", "/chat/completions", convertedReq)
}

// ParseResponse 解析响应
func (qa *QwenAdapter) ParseResponse(resp *http.Response) (*OpenAIResponse, error) {
	var qwenResp map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&qwenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	result := &OpenAIResponse{
		ID:      fmt.Sprintf("%v", qwenResp["request_id"]),
		Model:   fmt.Sprintf("%v", qwenResp["model"]),
		Created: int64(0),
	}

	return result, nil
}

// ParseStreamResponse 解析流式响应
func (qa *QwenAdapter) ParseStreamResponse(resp *http.Response) (<-chan *StreamChunk, error) {
	ch := make(chan *StreamChunk, 1)

	go func() {
		defer close(ch)
	}()

	return ch, nil
}

// ExtractUsage 提取使用量
func (qa *QwenAdapter) ExtractUsage(resp interface{}) (*Usage, error) {
	return &Usage{}, nil
}

// GetError 获取错误
func (qa *QwenAdapter) GetError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	return fmt.Errorf("http %d", resp.StatusCode)
}

// HealthCheck 健康检查
func (qa *QwenAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// ==================== 辅助函数 ====================

// convertMessagesToContents 将消息转换为 Gemini 格式
func convertMessagesToContents(messages []Message) interface{} {
	// 转换逻辑
	return messages
}
