package relay

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// Channel 已在 channel_model.go 中定义

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
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	channelSwitches int64
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
		if ch.Enabled {
			// 检查连续失败次数
			if ch.Metrics == nil || atomic.LoadInt64(&ch.Metrics.ConsecutiveFailures) < 3 {
				available = append(available, ch)
			}
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
		// 如果没有 Ability 配置，默认支持所有模型
		if ch.Ability == nil || len(ch.Ability.SupportedModels) == 0 {
			supportedChannels = append(supportedChannels, ch)
		} else {
			for _, m := range ch.Ability.SupportedModels {
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
	if currentChannel.Metrics != nil {
		atomic.AddInt64(&currentChannel.Metrics.ConsecutiveFailures, 1)
	}
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
	var respBody []byte
	var respHeader http.Header

	// 缓存请求体以便重试
	bodyBytes, errRead := io.ReadAll(body)
	if errRead != nil {
		return nil, nil, errRead
	}

	_, err := RetryWithCallback(
		ctx,
		rc.retryPolicy,
		func(ctx context.Context, retryCtx *RetryContext) error {
			var channel *Channel
			if retryCtx.RetryCount == 0 {
				channel = rc.SelectChannel("")
				lastChannel = channel
			} else {
				channel = rc.SwitchChannel(lastChannel)
				lastChannel = channel
			}

			if channel == nil {
				return fmt.Errorf("no available channel")
			}

			respBody, respHeader, err = rc.doSingleRequest(ctx, channel, method, path, bytes.NewReader(bodyBytes), headers)
			if err != nil {
				lastErr = err
				return &RetryableError{
					StatusCode: 0,
					Message:    err.Error(),
					Err:        err,
				}
			}

			return nil
		},
	)

	if err != nil {
		atomic.AddInt64(&rc.failedRequests, 1)
		return nil, nil, lastErr
	}

	atomic.AddInt64(&rc.successRequests, 1)
	return respBody, respHeader, nil
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
	if len(channel.Keys) > 0 && channel.Keys[0].APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+channel.Keys[0].APIKey)
	}

	// 添加自定义头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 发送请求
	resp, err := rc.httpClient.Do(req)
	_ = time.Since(startTime) // latency 未使用

	// 更新统计
	if channel.Metrics != nil {
		atomic.AddInt64(&channel.Metrics.TotalRequests, 1)
		latencyMs := time.Since(startTime).Milliseconds()
		channel.RecordSuccess(latencyMs)
	}
	// LastRequestTime 已在 Metrics 中跟踪
	atomic.AddInt64(&rc.totalRequests, 1)

	if err != nil {
		if channel.Metrics != nil {
			atomic.AddInt64(&channel.Metrics.ConsecutiveFailures, 1)
		}
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
		if channel.Metrics != nil {
			atomic.AddInt64(&channel.Metrics.ConsecutiveFailures, 1)
		}
		return nil, resp.Header, &RetryableError{
			StatusCode: resp.StatusCode,
			Message:    err.Error(),
			Err:        err,
		}
	}

	// 检查状态码
	if resp.StatusCode >= 400 {
		if channel.Metrics != nil {
			atomic.AddInt64(&channel.Metrics.ConsecutiveFailures, 1)
		}

		// 检查 Retry-After 头
		retryAfter := time.Duration(0)
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			// 解析 Retry-After
			var seconds int
			_, err := fmt.Sscanf(retryAfterStr, "%d", &seconds)
			if err == nil && seconds > 0 {
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
	if channel.Metrics != nil {
		atomic.StoreInt64(&channel.Metrics.ConsecutiveFailures, 0)
		atomic.AddInt64(&channel.Metrics.SuccessfulRequests, 1)
	}

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
		"total_requests":   total,
		"success_requests": success,
		"failed_requests":  failed,
		"success_rate":     successRate,
		"channel_switches": switches,
	}
}

// GetChannelStatistics 获取渠道统计
func (rc *RequestClient) GetChannelStatistics() []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, ch := range rc.channels {
		successRate := 0.0
		var totalRequests int64
		var successCount int64
		var failureCount int64
		var avgLatency float64
		var consecutiveFailures int64

		if ch.Metrics != nil {
			totalRequests = atomic.LoadInt64(&ch.Metrics.TotalRequests)
			successCount = atomic.LoadInt64(&ch.Metrics.SuccessfulRequests)
			failureCount = atomic.LoadInt64(&ch.Metrics.FailedRequests)
			avgLatency = ch.Metrics.AvgLatency
			consecutiveFailures = atomic.LoadInt64(&ch.Metrics.ConsecutiveFailures)

			if totalRequests > 0 {
				successRate = float64(successCount) / float64(totalRequests) * 100
			}
		}

		result = append(result, map[string]interface{}{
			"id":                   ch.ID,
			"name":                 ch.Name,
			"enabled":              ch.Enabled,
			"priority":             ch.Priority,
			"weight":               ch.Weight,
			"request_count":        totalRequests,
			"success_count":        successCount,
			"failure_count":        failureCount,
			"success_rate":         successRate,
			"avg_latency_ms":       avgLatency,
			"consecutive_failures": consecutiveFailures,
		})
	}

	return result
}

// RecoverChannel 恢复渠道
func (rc *RequestClient) RecoverChannel(channelID string) error {
	for _, ch := range rc.channels {
		if ch.ID == channelID {
			if ch.Metrics != nil {
				atomic.StoreInt64(&ch.Metrics.ConsecutiveFailures, 0)
			}
			return nil
		}
	}
	return fmt.Errorf("channel not found")
}
