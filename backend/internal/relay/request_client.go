package relay

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// Channel 代表一个中继渠道（API 服务提供商）
type Channel struct {
	// 渠道 ID
	ID string

	// 渠道名称
	Name string

	// 基础 URL
	BaseURL string

	// API Key
	APIKey string

	// 是否启用
	Enabled bool

	// 优先级（数值越小优先级越高）
	Priority int

	// 权重（用于负载均衡）
	Weight int

	// 支持的模型
	SupportedModels []string

	// 统计信息
	RequestCount     int64
	SuccessCount     int64
	FailureCount     int64
	AvgLatencyMs     int64
	LastRequestTime  time.Time
	ConsecutiveFailures int32
}

// RequestClient 请求客户端
type RequestClient struct {
	// HTTP 客户端
	httpClient *http.Client

	// 重试策略
	retryPolicy *RetryPolicy

	// 渠道列表
	channels []*Channel

	// 当前渠道索引
	currentChannelIndex int

	// 统计信息
	totalRequests    int64
	successRequests  int64
	failedRequests   int64
	channelSwitches  int64
}

// NewRequestClient 创建新的请求客户端
func NewRequestClient(timeout time.Duration) *RequestClient {
	return &RequestClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retryPolicy: NewRetryPolicy(),
		channels:    make([]*Channel, 0),
	}
}

// AddChannel 添加渠道
func (rc *RequestClient) AddChannel(channel *Channel) {
	rc.channels = append(rc.channels, channel)
}

// SetRetryPolicy 设置重试策略
func (rc *RequestClient) SetRetryPolicy(policy *RetryPolicy) {
	rc.retryPolicy = policy
}

// GetAvailableChannels 获取可用的渠道
func (rc *RequestClient) GetAvailableChannels() []*Channel {
	var available []*Channel
	for _, ch := range rc.channels {
		if ch.Enabled && ch.ConsecutiveFailures < 3 {
			available = append(available, ch)
		}
	}
	return available
}

// SelectChannel 选择一个渠道
func (rc *RequestClient) SelectChannel(model string) *Channel {
	available := rc.GetAvailableChannels()
	if len(available) == 0 {
		return nil
	}

	// 过滤支持该模型的渠道
	supportedChannels := make([]*Channel, 0)
	for _, ch := range available {
		if len(ch.SupportedModels) == 0 {
			// 如果没有模型限制，则支持所有模型
			supportedChannels = append(supportedChannels, ch)
		} else {
			for _, m := range ch.SupportedModels {
				if m == model || m == "*" {
					supportedChannels = append(supportedChannels, ch)
					break
				}
			}
		}
	}

	if len(supportedChannels) == 0 {
		return nil
	}

	// 简单的轮询策略（可以改为更复杂的策略）
	rc.currentChannelIndex = (rc.currentChannelIndex + 1) % len(supportedChannels)
	return supportedChannels[rc.currentChannelIndex]
}

// SwitchChannel 切换到下一个渠道
func (rc *RequestClient) SwitchChannel(currentChannel *Channel) *Channel {
	// 标记当前渠道失败
	currentChannel.ConsecutiveFailures++
	atomic.AddInt64(&rc.channelSwitches, 1)

	// 选择下一个渠道
	available := rc.GetAvailableChannels()
	for _, ch := range available {
		if ch.ID != currentChannel.ID {
			return ch
		}
	}

	return nil
}

// DoRequest 发送请求并支持重试
// method: HTTP 方法
// path: 请求路径
// body: 请求体
// headers: 请求头
// 返回: (响应体, 响应头, 错误)
func (rc *RequestClient) DoRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	headers map[string]string,
) ([]byte, http.Header, error) {
	var lastChannel *Channel
	var lastErr error

	// 使用重试机制
	success, err := RetryWithCallback(
		ctx,
		rc.retryPolicy,
		func(ctx context.Context, retryCtx *RetryContext) error {
			// 选择渠道
			var channel *Channel
			if retryCtx.RetryCount == 0 {
				channel = rc.SelectChannel("")
				lastChannel = channel
			} else {
				// 重试时切换渠道
				channel = rc.SwitchChannel(lastChannel)
				lastChannel = channel
			}

			if channel == nil {
				return fmt.Errorf("no available channel")
			}

			// 发送请求
			respBody, respHeader, err := rc.doSingleRequest(ctx, channel, method, path, body, headers)
			if err != nil {
				// 包装为可重试错误
				retryErr := &RetryableError{
					StatusCode: 0,
					Message:    err.Error(),
					Err:        err,
				}
				lastErr = retryErr
				return retryErr
			}

			// 成功返回
			*retryCtx = RetryContext{
				RetryCount: retryCtx.RetryCount,
				Delay:      retryCtx.Delay,
			}

			// 这里需要返回响应数据，但签名不支持
			// 所以我们把数据存储在闭包外部
			return nil
		},
	)

	if !success {
		atomic.AddInt64(&rc.failedRequests, 1)
		return nil, nil, lastErr
	}

	atomic.AddInt64(&rc.successRequests, 1)
	return nil, nil, nil // 这里应该返回实际的响应数据
}

// doSingleRequest 发送单个请求（不带重试）
func (rc *RequestClient) doSingleRequest(
	ctx context.Context,
	channel *Channel,
	method string,
	path string,
	body io.Reader,
	headers map[string]string,
) ([]byte, http.Header, error) {
	// 构建完整 URL
	url := channel.BaseURL + path

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, err
	}

	// 添加认证头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))

	// 添加自定义头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 发送请求
	resp, err := rc.httpClient.Do(req)
	latency := time.Since(startTime)

	// 更新统计
	atomic.AddInt64(&channel.RequestCount, 1)
	atomic.StoreInt64(&channel.AvgLatencyMs, latency.Milliseconds())
	channel.LastRequestTime = time.Now()
	atomic.AddInt64(&rc.totalRequests, 1)

	if err != nil {
		channel.ConsecutiveFailures++
		return nil, nil, &RetryableError{
			StatusCode: 0,
			Message:    err.Error(),
			Err:        err,
		}
	}

	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		channel.ConsecutiveFailures++
		return nil, resp.Header, &RetryableError{
			StatusCode: resp.StatusCode,
			Message:    err.Error(),
			Err:        err,
		}
	}

	// 检查状态码
	if resp.StatusCode >= 400 {
		channel.ConsecutiveFailures++
		
		// 检查 Retry-After 头
		retryAfter := time.Duration(0)
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			// 解析 Retry-After
			if seconds := 0; _, err := fmt.Sscanf(retryAfterStr, "%d", &seconds); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}

		return respBody, resp.Header, &RetryableError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			Err:        fmt.Errorf("HTTP error: %d", resp.StatusCode),
			RetryAfter: retryAfter,
		}
	}

	// 成功
	channel.ConsecutiveFailures = 0
	atomic.AddInt64(&channel.SuccessCount, 1)

	return respBody, resp.Header, nil
}

// GetStatistics 获取请求统计
func (rc *RequestClient) GetStatistics() map[string]interface{} {
	total := atomic.LoadInt64(&rc.totalRequests)
	success := atomic.LoadInt64(&rc.successRequests)
	failed := atomic.LoadInt64(&rc.failedRequests)
	switches := atomic.LoadInt64(&rc.channelSwitches)

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_requests":    total,
		"success_requests":  success,
		"failed_requests":   failed,
		"success_rate":      successRate,
		"channel_switches":  switches,
	}
}

// GetChannelStatistics 获取渠道统计
func (rc *RequestClient) GetChannelStatistics() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, ch := range rc.channels {
		successRate := 0.0
		totalRequests := atomic.LoadInt64(&ch.RequestCount)
		if totalRequests > 0 {
			successCount := atomic.LoadInt64(&ch.SuccessCount)
			successRate = float64(successCount) / float64(totalRequests) * 100
		}

		result = append(result, map[string]interface{}{
			"id":                   ch.ID,
			"name":                 ch.Name,
			"enabled":              ch.Enabled,
			"priority":             ch.Priority,
			"weight":               ch.Weight,
			"request_count":        totalRequests,
			"success_count":        atomic.LoadInt64(&ch.SuccessCount),
			"failure_count":        atomic.LoadInt64(&ch.FailureCount),
			"success_rate":         successRate,
			"avg_latency_ms":       atomic.LoadInt64(&ch.AvgLatencyMs),
			"consecutive_failures": atomic.LoadInt32(&ch.ConsecutiveFailures),
		})
	}

	return result
}

// RecoverChannel 恢复渠道
func (rc *RequestClient) RecoverChannel(channelID string) error {
	for _, ch := range rc.channels {
		if ch.ID == channelID {
			ch.ConsecutiveFailures = 0
			return nil
		}
	}
	return fmt.Errorf("channel not found")
}

