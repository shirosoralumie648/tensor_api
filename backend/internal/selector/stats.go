package selector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StatsManager 统计管理器
type StatsManager struct {
	stats map[int]*ChannelStats // key: channelID
	mu    sync.RWMutex
}

// NewStatsManager 创建统计管理器
func NewStatsManager() *StatsManager {
	return &StatsManager{
		stats: make(map[int]*ChannelStats),
	}
}

// UpdateStats 更新渠道统计信息
func (sm *StatsManager) UpdateStats(ctx context.Context, channelID int, success bool, responseTime time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stats, exists := sm.stats[channelID]
	if !exists {
		stats = &ChannelStats{
			ChannelID: channelID,
		}
		sm.stats[channelID] = stats
	}

	stats.TotalRequests++
	stats.LastUsedAt = time.Now()

	if success {
		stats.SuccessCount++

		// 更新平均响应时间（使用指数移动平均）
		if stats.AvgResponseTime == 0 {
			stats.AvgResponseTime = responseTime
		} else {
			// EMA: α = 0.3
			alpha := 0.3
			stats.AvgResponseTime = time.Duration(
				float64(stats.AvgResponseTime)*(1-alpha) + float64(responseTime)*alpha,
			)
		}
	} else {
		stats.FailureCount++
		stats.LastFailedAt = time.Now()
	}

	return nil
}

// GetStats 获取渠道统计信息
func (sm *StatsManager) GetStats(ctx context.Context, channelID int) (*ChannelStats, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats, exists := sm.stats[channelID]
	if !exists {
		return &ChannelStats{
			ChannelID: channelID,
		}, nil
	}

	// 返回副本
	statsCopy := *stats
	return &statsCopy, nil
}

// RecordFailure 记录失败
func (sm *StatsManager) RecordFailure(ctx context.Context, channelID int) error {
	return sm.UpdateStats(ctx, channelID, false, 0)
}

// RecordSuccess 记录成功
func (sm *StatsManager) RecordSuccess(ctx context.Context, channelID int, responseTime time.Duration) error {
	return sm.UpdateStats(ctx, channelID, true, responseTime)
}

// GetAllStats 获取所有统计信息
func (sm *StatsManager) GetAllStats() map[int]*ChannelStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[int]*ChannelStats)
	for id, stats := range sm.stats {
		statsCopy := *stats
		result[id] = &statsCopy
	}

	return result
}

// ResetStats 重置统计信息
func (sm *StatsManager) ResetStats(channelID int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if channelID == 0 {
		// 重置所有
		sm.stats = make(map[int]*ChannelStats)
	} else {
		// 重置指定渠道
		if _, exists := sm.stats[channelID]; !exists {
			return fmt.Errorf("channel %d stats not found", channelID)
		}
		delete(sm.stats, channelID)
	}

	return nil
}

// GetSuccessRate 获取成功率
func (sm *StatsManager) GetSuccessRate(channelID int) float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats, exists := sm.stats[channelID]
	if !exists || stats.TotalRequests == 0 {
		return 0.0
	}

	return float64(stats.SuccessCount) / float64(stats.TotalRequests)
}
