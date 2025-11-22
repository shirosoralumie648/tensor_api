package selector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/gorm"
)

// DefaultChannelSelector 默认渠道选择器实现
type DefaultChannelSelector struct {
	db         *gorm.DB
	cache      *ChannelCache
	stats      *StatsManager
	strategies map[SelectStrategy]StrategyFunc
	mu         sync.RWMutex
}

// NewDefaultChannelSelector 创建默认渠道选择器
func NewDefaultChannelSelector(db *gorm.DB, cache *ChannelCache, stats *StatsManager) *DefaultChannelSelector {
	selector := &DefaultChannelSelector{
		db:         db,
		cache:      cache,
		stats:      stats,
		strategies: make(map[SelectStrategy]StrategyFunc),
	}

	// 注册所有策略
	selector.registerStrategies()

	return selector
}

// registerStrategies 注册所有选择策略
func (s *DefaultChannelSelector) registerStrategies() {
	s.strategies[StrategyWeight] = s.selectByWeight
	s.strategies[StrategyPriority] = s.selectByPriority
	s.strategies[StrategyRoundRobin] = s.selectByRoundRobin
	s.strategies[StrategyLowestLatency] = s.selectByLowestLatency
	s.strategies[StrategyRandom] = s.selectByRandom
}

// Select 选择渠道
func (s *DefaultChannelSelector) Select(ctx context.Context, req *SelectRequest) (*SelectResult, error) {
	// 从缓存获取可用渠道
	channels, err := s.cache.GetAvailableChannels(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get available channels: %w", err)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no available channels for model: %s", req.Model)
	}

	// 过滤掉排除的渠道
	if len(req.ExcludeIDs) > 0 {
		channels = s.filterExcludedChannels(channels, req.ExcludeIDs)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("all channels are excluded")
	}

	// 根据策略选择渠道
	strategy := req.Strategy
	if strategy == "" {
		strategy = StrategyWeight // 默认使用权重策略
	}

	strategyFunc, exists := s.strategies[strategy]
	if !exists {
		return nil, fmt.Errorf("unsupported strategy: %s", strategy)
	}

	channel, err := strategyFunc(ctx, channels, req)
	if err != nil {
		return nil, fmt.Errorf("strategy failed: %w", err)
	}

	return &SelectResult{
		Channel:       channel,
		TotalAttempts: 1,
	}, nil
}

// SelectWithRetry 带重试的选择
func (s *DefaultChannelSelector) SelectWithRetry(ctx context.Context, req *SelectRequest, maxRetries int) (*SelectResult, error) {
	if maxRetries <= 0 {
		maxRetries = 3 // 默认最多重试3次
	}

	excludeIDs := make([]int, 0)
	if req.ExcludeIDs != nil {
		excludeIDs = append(excludeIDs, req.ExcludeIDs...)
	}

	var lastErr error
	failedCount := 0

	for attempt := 0; attempt < maxRetries; attempt++ {
		reqCopy := *req
		reqCopy.ExcludeIDs = excludeIDs

		result, err := s.Select(ctx, &reqCopy)
		if err != nil {
			lastErr = err
			failedCount++
			// 将失败的渠道加入排除列表，避免重复尝试同一渠道
			if result != nil && result.Channel != nil {
				excludeIDs = append(excludeIDs, result.Channel.ID)
			}
			continue
		}

		result.FailedCount = failedCount
		result.TotalAttempts = attempt + 1
		return result, nil
	}

	return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
}

// UpdateStats 更新渠道统计信息
func (s *DefaultChannelSelector) UpdateStats(ctx context.Context, channelID int, success bool, responseTime time.Duration) error {
	return s.stats.UpdateStats(ctx, channelID, success, responseTime)
}

// GetStats 获取渠道统计信息
func (s *DefaultChannelSelector) GetStats(ctx context.Context, channelID int) (*ChannelStats, error) {
	return s.stats.GetStats(ctx, channelID)
}

// MarkChannelFailed 标记渠道失败
func (s *DefaultChannelSelector) MarkChannelFailed(ctx context.Context, channelID int, reason string) error {
	// 更新统计
	if err := s.stats.RecordFailure(ctx, channelID); err != nil {
		return err
	}

	// 检查是否需要禁用渠道
	stats, err := s.stats.GetStats(ctx, channelID)
	if err != nil {
		return err
	}

	// 如果连续失败超过阈值，自动禁用渠道
	if stats.FailureCount >= 10 && stats.TotalRequests > 0 {
		failureRate := float64(stats.FailureCount) / float64(stats.TotalRequests)
		if failureRate > 0.5 { // 失败率超过50%
			return s.disableChannel(ctx, channelID, reason)
		}
	}

	return nil
}

// RefreshCache 刷新渠道缓存
func (s *DefaultChannelSelector) RefreshCache(ctx context.Context) error {
	return s.cache.Refresh(ctx)
}

// filterExcludedChannels 过滤掉排除的渠道
func (s *DefaultChannelSelector) filterExcludedChannels(channels []*model.Channel, excludeIDs []int) []*model.Channel {
	excludeMap := make(map[int]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	filtered := make([]*model.Channel, 0)
	for _, ch := range channels {
		if !excludeMap[ch.ID] {
			filtered = append(filtered, ch)
		}
	}

	return filtered
}

// disableChannel 禁用渠道
func (s *DefaultChannelSelector) disableChannel(ctx context.Context, channelID int, reason string) error {
	return s.db.WithContext(ctx).
		Model(&model.Channel{}).
		Where("id = ?", channelID).
		Updates(map[string]interface{}{
			"status":     1, // 1 表示禁用
			"updated_at": time.Now(),
		}).Error
}
