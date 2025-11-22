package relay

import (
	"context"
	"sync"
	"time"
)

// StreamMonitor 流式请求监控器
type StreamMonitor struct {
	mu            sync.RWMutex
	activeStreams map[string]*StreamMetrics
}

// StreamMetrics 流式请求指标
type StreamMetrics struct {
	RequestID    string
	UserID       int
	Model        string
	StartTime    time.Time
	LastActivity time.Time
	ChunkCount   int
	BytesSent    int64
	TokenCount   int
	Status       string // active, completed, failed, timeout
}

// NewStreamMonitor 创建流式监控器
func NewStreamMonitor() *StreamMonitor {
	return &StreamMonitor{
		activeStreams: make(map[string]*StreamMetrics),
	}
}

// StartStream 开始监控流式请求
func (m *StreamMonitor) StartStream(requestID string, userID int, model string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeStreams[requestID] = &StreamMetrics{
		RequestID:    requestID,
		UserID:       userID,
		Model:        model,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Status:       "active",
	}
}

// UpdateActivity 更新活动时间
func (m *StreamMonitor) UpdateActivity(requestID string, chunkSize int, tokens int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.activeStreams[requestID]; exists {
		metrics.LastActivity = time.Now()
		metrics.ChunkCount++
		metrics.BytesSent += int64(chunkSize)
		metrics.TokenCount += tokens
	}
}

// CompleteStream 标记流完成
func (m *StreamMonitor) CompleteStream(requestID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.activeStreams[requestID]; exists {
		metrics.Status = "completed"
	}
}

// FailStream 标记流失败
func (m *StreamMonitor) FailStream(requestID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.activeStreams[requestID]; exists {
		metrics.Status = "failed"
	}
}

// RemoveStream 移除流记录
func (m *StreamMonitor) RemoveStream(requestID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.activeStreams, requestID)
}

// GetMetrics 获取流指标
func (m *StreamMonitor) GetMetrics(requestID string) *StreamMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics, exists := m.activeStreams[requestID]; exists {
		// 返回副本
		copy := *metrics
		return &copy
	}

	return nil
}

// GetActiveStreams 获取所有活跃流
func (m *StreamMonitor) GetActiveStreams() []*StreamMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*StreamMetrics, 0, len(m.activeStreams))
	for _, metrics := range m.activeStreams {
		if metrics.Status == "active" {
			copy := *metrics
			result = append(result, &copy)
		}
	}

	return result
}

// CleanupStaleStreams 清理过期的流记录
func (m *StreamMonitor) CleanupStaleStreams(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	now := time.Now()

	for requestID, metrics := range m.activeStreams {
		if metrics.Status != "active" && now.Sub(metrics.LastActivity) > maxAge {
			delete(m.activeStreams, requestID)
			cleaned++
		}
	}

	return cleaned
}

// StartCleanupWorker 启动定期清理工作
func (m *StreamMonitor) StartCleanupWorker(ctx context.Context, interval time.Duration, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cleaned := m.CleanupStaleStreams(maxAge)
			if cleaned > 0 {
				// 可以记录日志
				_ = cleaned
			}

		case <-ctx.Done():
			return
		}
	}
}

// GetStats 获取总体统计
func (m *StreamMonitor) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_streams":  len(m.activeStreams),
		"active_streams": 0,
		"total_tokens":   0,
		"total_bytes":    int64(0),
	}

	for _, metrics := range m.activeStreams {
		if metrics.Status == "active" {
			stats["active_streams"] = stats["active_streams"].(int) + 1
		}
		stats["total_tokens"] = stats["total_tokens"].(int) + metrics.TokenCount
		stats["total_bytes"] = stats["total_bytes"].(int64) + metrics.BytesSent
	}

	return stats
}
