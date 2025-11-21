package relay

import (
	"sync/atomic"
	"time"
)

// ChannelAbility 渠道能力
type ChannelAbility struct {
	// 支持的模型列表
	SupportedModels []string `json:"supported_models"`

	// 功能特性
	Features map[string]interface{} `json:"features"`

	// 版本信息
	Version string `json:"version"`

	// 最大并发数
	MaxConcurrency int `json:"max_concurrency"`

	// 速率限制（请求/分钟）
	RateLimit int `json:"rate_limit"`

	// 支持流式
	SupportsStreaming bool `json:"supports_streaming"`

	// 支持函数调用
	SupportsFunctionCalling bool `json:"supports_function_calling"`

	// 支持视觉
	SupportsVision bool `json:"supports_vision"`
}

// ChannelStatus 渠道状态
type ChannelStatus int

const (
	// 健康
	ChannelStatusHealthy ChannelStatus = iota
	// 降级
	ChannelStatusDegraded
	// 不可用
	ChannelStatusUnavailable
	// 禁用
	ChannelStatusDisabled
)

// String 返回状态的字符串表示
func (cs ChannelStatus) String() string {
	switch cs {
	case ChannelStatusHealthy:
		return "healthy"
	case ChannelStatusDegraded:
		return "degraded"
	case ChannelStatusUnavailable:
		return "unavailable"
	case ChannelStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// ChannelMetrics 渠道指标
type ChannelMetrics struct {
	// 总请求数
	TotalRequests int64 `json:"total_requests"`

	// 成功请求数
	SuccessfulRequests int64 `json:"successful_requests"`

	// 失败请求数
	FailedRequests int64 `json:"failed_requests"`

	// 平均延迟（毫秒）
	AvgLatency float64 `json:"avg_latency_ms"`

	// 最后一次成功时间
	LastSuccessTime int64 `json:"last_success_time"`

	// 最后一次失败时间
	LastFailureTime int64 `json:"last_failure_time"`

	// 连续失败次数
	ConsecutiveFailures int64 `json:"consecutive_failures"`

	// 当前并发数
	CurrentConcurrency int64 `json:"current_concurrency"`
}

// GetSuccessRate 计算成功率
func (cm *ChannelMetrics) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&cm.TotalRequests)
	if total == 0 {
		return 0
	}
	successful := atomic.LoadInt64(&cm.SuccessfulRequests)
	return float64(successful) / float64(total) * 100
}

// ChannelKey 渠道密钥
type ChannelKey struct {
	// 密钥 ID
	ID string `json:"id"`

	// API 密钥
	APIKey string `json:"api_key"`

	// 密钥类型（Bearer、API Key 等）
	Type string `json:"type"`

	// 是否可用
	Enabled bool `json:"enabled"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 过期时间
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// 使用次数
	UsageCount int64 `json:"usage_count"`

	// 最后使用时间
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`

	// 配额限制
	QuotaLimit *int64 `json:"quota_limit,omitempty"`
}

// Channel 渠道信息
type Channel struct {
	// 渠道 ID
	ID string `json:"id"`

	// 渠道名称
	Name string `json:"name"`

	// API 基础 URL
	BaseURL string `json:"base_url"`

	// 渠道类型（openai、claude、gemini 等）
	Type string `json:"type"`

	// 优先级（数字越小优先级越高）
	Priority int `json:"priority"`

	// 权重（用于负载均衡）
	Weight int `json:"weight"`

	// 当前状态
	Status atomic.Value // ChannelStatus

	// 能力
	Ability *ChannelAbility `json:"ability"`

	// 指标
	Metrics *ChannelMetrics `json:"metrics"`

	// 密钥列表
	Keys []*ChannelKey `json:"keys"`

	// 地理位置
	Region string `json:"region"`

	// 描述
	Description string `json:"description"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 是否启用
	Enabled bool `json:"enabled"`
}

// NewChannel 创建新的渠道
func NewChannel(id, name, baseURL, channelType string) *Channel {
	now := time.Now()
	ch := &Channel{
		ID:        id,
		Name:      name,
		BaseURL:   baseURL,
		Type:      channelType,
		Priority:  100,
		Weight:    1,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
		Metrics: &ChannelMetrics{
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AvgLatency:          0,
			CurrentConcurrency:  0,
			ConsecutiveFailures: 0,
		},
		Ability: &ChannelAbility{
			SupportedModels: make([]string, 0),
			Features:        make(map[string]interface{}),
			Version:         "1.0",
		},
		Keys: make([]*ChannelKey, 0),
	}

	// 设置初始状态为健康
	ch.Status.Store(ChannelStatusHealthy)

	return ch
}

// GetStatus 获取渠道状态
func (ch *Channel) GetStatus() ChannelStatus {
	if !ch.Enabled {
		return ChannelStatusDisabled
	}
	status, ok := ch.Status.Load().(ChannelStatus)
	if !ok {
		return ChannelStatusHealthy
	}
	return status
}

// SetStatus 设置渠道状态
func (ch *Channel) SetStatus(status ChannelStatus) {
	ch.Status.Store(status)
	ch.UpdatedAt = time.Now()
}

// IsAvailable 是否可用
func (ch *Channel) IsAvailable() bool {
	if !ch.Enabled {
		return false
	}
	status := ch.GetStatus()
	return status == ChannelStatusHealthy
}

// RecordSuccess 记录成功请求
func (ch *Channel) RecordSuccess(latency int64) {
	atomic.AddInt64(&ch.Metrics.TotalRequests, 1)
	atomic.AddInt64(&ch.Metrics.SuccessfulRequests, 1)
	atomic.StoreInt64(&ch.Metrics.LastSuccessTime, time.Now().Unix())
	atomic.StoreInt64(&ch.Metrics.ConsecutiveFailures, 0)

	// 更新平均延迟（简单平均）
	_ = atomic.LoadInt64(&ch.Metrics.TotalRequests) // total 暂未使用
	successCount := atomic.LoadInt64(&ch.Metrics.SuccessfulRequests)
	currentAvg := ch.Metrics.AvgLatency
	newAvg := (currentAvg*float64(successCount-1) + float64(latency)) / float64(successCount)
	ch.Metrics.AvgLatency = newAvg
}

// RecordFailure 记录失败请求
func (ch *Channel) RecordFailure() {
	atomic.AddInt64(&ch.Metrics.TotalRequests, 1)
	atomic.AddInt64(&ch.Metrics.FailedRequests, 1)
	atomic.AddInt64(&ch.Metrics.ConsecutiveFailures, 1)
	atomic.StoreInt64(&ch.Metrics.LastFailureTime, time.Now().Unix())

	// 如果连续失败次数过多，设置为降级或不可用
	failureCount := atomic.LoadInt64(&ch.Metrics.ConsecutiveFailures)
	if failureCount >= 10 {
		ch.SetStatus(ChannelStatusUnavailable)
	} else if failureCount >= 5 {
		ch.SetStatus(ChannelStatusDegraded)
	}
}

// RecordConcurrency 记录并发
func (ch *Channel) RecordConcurrency(delta int64) {
	atomic.AddInt64(&ch.Metrics.CurrentConcurrency, delta)
}

// SupportModel 是否支持指定模型
func (ch *Channel) SupportModel(model string) bool {
	if ch.Ability == nil {
		return false
	}
	for _, m := range ch.Ability.SupportedModels {
		if m == model {
			return true
		}
	}
	return false
}

// GetSuccessRate 获取成功率
func (ch *Channel) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&ch.Metrics.TotalRequests)
	if total == 0 {
		return 0
	}
	success := atomic.LoadInt64(&ch.Metrics.SuccessfulRequests)
	return float64(success) / float64(total) * 100
}

// GetMetricsSnapshot 获取指标快照
func (ch *Channel) GetMetricsSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"id":                   ch.ID,
		"name":                 ch.Name,
		"status":               ch.GetStatus().String(),
		"enabled":              ch.Enabled,
		"total_requests":       atomic.LoadInt64(&ch.Metrics.TotalRequests),
		"successful_requests":  atomic.LoadInt64(&ch.Metrics.SuccessfulRequests),
		"failed_requests":      atomic.LoadInt64(&ch.Metrics.FailedRequests),
		"success_rate":         ch.GetSuccessRate(),
		"avg_latency_ms":       ch.Metrics.AvgLatency,
		"current_concurrency":  atomic.LoadInt64(&ch.Metrics.CurrentConcurrency),
		"consecutive_failures": atomic.LoadInt64(&ch.Metrics.ConsecutiveFailures),
		"last_success_time":    atomic.LoadInt64(&ch.Metrics.LastSuccessTime),
		"last_failure_time":    atomic.LoadInt64(&ch.Metrics.LastFailureTime),
	}
}

// ChannelFilter 渠道过滤条件
type ChannelFilter struct {
	// 渠道类型
	Type string

	// 支持的模型
	Model string

	// 地区
	Region string

	// 状态
	Status ChannelStatus

	// 最小可用性
	MinAvailability float64

	// 最小优先级
	MinPriority int

	// 只选择启用的渠道
	OnlyEnabled bool
}

// Matches 检查渠道是否匹配过滤条件
func (ch *Channel) Matches(filter *ChannelFilter) bool {
	if filter == nil {
		return true
	}

	// 检查启用状态
	if filter.OnlyEnabled && !ch.Enabled {
		return false
	}

	// 检查类型
	if filter.Type != "" && ch.Type != filter.Type {
		return false
	}

	// 检查模型支持
	if filter.Model != "" && !ch.SupportModel(filter.Model) {
		return false
	}

	// 检查地区
	if filter.Region != "" && ch.Region != filter.Region {
		return false
	}

	// 检查状态
	if filter.Status != ch.GetStatus() && filter.Status != ChannelStatus(-1) {
		return false
	}

	// 检查可用性
	if filter.MinAvailability > 0 && ch.GetSuccessRate() < filter.MinAvailability {
		return false
	}

	// 检查优先级
	if filter.MinPriority > 0 && ch.Priority > filter.MinPriority {
		return false
	}

	return true
}
