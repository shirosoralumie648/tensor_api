package selector

import (
	"context"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// SelectStrategy 选择策略类型
type SelectStrategy string

const (
	// StrategyWeight 权重选择（根据权重随机选择）
	StrategyWeight SelectStrategy = "weight"

	// StrategyPriority 优先级选择（按优先级从高到低）
	StrategyPriority SelectStrategy = "priority"

	// StrategyRoundRobin 轮询选择
	StrategyRoundRobin SelectStrategy = "round_robin"

	// StrategyLowestLatency 最低延迟选择
	StrategyLowestLatency SelectStrategy = "lowest_latency"

	// StrategyRandom 随机选择
	StrategyRandom SelectStrategy = "random"
)

// SelectRequest 选择请求
type SelectRequest struct {
	Model      string         // 请求的模型
	Strategy   SelectStrategy // 选择策略
	UserID     int            // 用户ID
	UserGroup  string         // 用户分组
	ExcludeIDs []int          // 排除的渠道ID（用于重试）
}

// SelectResult 选择结果
type SelectResult struct {
	Channel       *model.Channel // 选中的渠道
	FailedCount   int            // 失败次数
	TotalAttempts int            // 总尝试次数
}

// ChannelStats 渠道统计信息
type ChannelStats struct {
	ChannelID       int
	SuccessCount    int64         // 成功次数
	FailureCount    int64         // 失败次数
	TotalRequests   int64         // 总请求数
	AvgResponseTime time.Duration // 平均响应时间
	LastUsedAt      time.Time     // 最后使用时间
	LastFailedAt    time.Time     // 最后失败时间
}

// ChannelSelector 渠道选择器接口
type ChannelSelector interface {
	// Select 选择渠道
	Select(ctx context.Context, req *SelectRequest) (*SelectResult, error)

	// SelectWithRetry 带重试的选择
	SelectWithRetry(ctx context.Context, req *SelectRequest, maxRetries int) (*SelectResult, error)

	// UpdateStats 更新渠道统计信息
	UpdateStats(ctx context.Context, channelID int, success bool, responseTime time.Duration) error

	// GetStats 获取渠道统计信息
	GetStats(ctx context.Context, channelID int) (*ChannelStats, error)

	// MarkChannelFailed 标记渠道失败
	MarkChannelFailed(ctx context.Context, channelID int, reason string) error

	// RefreshCache 刷新渠道缓存
	RefreshCache(ctx context.Context) error
}

// StrategyFunc 策略函数类型
type StrategyFunc func(ctx context.Context, channels []*model.Channel, req *SelectRequest) (*model.Channel, error)
