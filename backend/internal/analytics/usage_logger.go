package analytics

import (
	"fmt"
	"sync"
	"time"
)

// UsageRecord 使用记录
type UsageRecord struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	TokenID         string    `json:"token_id"`
	Model           string    `json:"model"`
	Provider        string    `json:"provider"`
	RequestTokens   int64     `json:"request_tokens"`
	ResponseTokens  int64     `json:"response_tokens"`
	TotalTokens     int64     `json:"total_tokens"`
	Cost            float64   `json:"cost"`
	Duration        int64     `json:"duration"` // 毫秒
	Status          string    `json:"status"`   // success, error, timeout
	ErrorMsg        string    `json:"error_msg,omitempty"`
	ClientIP        string    `json:"client_ip"`
	Endpoint        string    `json:"endpoint"`
	Timestamp       time.Time `json:"timestamp"`
	RequestID       string    `json:"request_id"`
	StreamingMode   bool      `json:"streaming_mode"`
	CacheHit        bool      `json:"cache_hit"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UsageLogger 使用日志记录器
type UsageLogger struct {
	mu           sync.RWMutex
	buffer       []*UsageRecord
	bufferSize   int
	flushTicker  *time.Ticker
	store        UsageStore
	channels     map[string]chan *UsageRecord
}

// UsageStore 使用数据存储接口
type UsageStore interface {
	Write(records []*UsageRecord) error
	Query(filter *UsageFilter) ([]*UsageRecord, error)
	GetAggregated(filter *AggregationFilter) (*AggregatedStats, error)
	DeleteOldRecords(before time.Time) error
}

// UsageFilter 使用记录过滤器
type UsageFilter struct {
	UserID    string
	TokenID   string
	Model     string
	Provider  string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Offset    int
}

// AggregationFilter 聚合过滤器
type AggregationFilter struct {
	UserID     string
	Model      string
	Provider   string
	StartTime  time.Time
	EndTime    time.Time
	GroupBy    string // day, hour, model, provider
	Dimensions []string
}

// AggregatedStats 聚合统计
type AggregatedStats struct {
	TotalRequests    int64
	SuccessRequests  int64
	ErrorRequests    int64
	TotalTokens      int64
	TotalCost        float64
	AvgDuration      int64
	AvgTokensPerReq  int64
	AvgCostPerReq    float64
	MaxDuration      int64
	MinDuration      int64
	Timeline         []*TimelinePoint
}

// TimelinePoint 时间线点
type TimelinePoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Requests    int64     `json:"requests"`
	Tokens      int64     `json:"tokens"`
	Cost        float64   `json:"cost"`
	AvgDuration int64     `json:"avg_duration"`
}

// NewUsageLogger 创建使用日志记录器
func NewUsageLogger(store UsageStore, bufferSize int, flushInterval time.Duration) *UsageLogger {
	logger := &UsageLogger{
		buffer:     make([]*UsageRecord, 0, bufferSize),
		bufferSize: bufferSize,
		store:      store,
		channels:   make(map[string]chan *UsageRecord),
	}

	// 启动定期刷新
	logger.flushTicker = time.NewTicker(flushInterval)
	go logger.flushRoutine()

	return logger
}

// RecordUsage 记录使用
func (ul *UsageLogger) RecordUsage(record *UsageRecord) error {
	if record.ID == "" {
		record.ID = fmt.Sprintf("rec_%d_%s", time.Now().UnixNano(), record.UserID)
	}
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	ul.mu.Lock()
	ul.buffer = append(ul.buffer, record)
	shouldFlush := len(ul.buffer) >= ul.bufferSize
	ul.mu.Unlock()

	// 广播到监听通道
	ul.broadcastRecord(record)

	if shouldFlush {
		return ul.Flush()
	}

	return nil
}

// Flush 刷新缓冲区到存储
func (ul *UsageLogger) Flush() error {
	ul.mu.Lock()
	if len(ul.buffer) == 0 {
		ul.mu.Unlock()
		return nil
	}

	records := ul.buffer
	ul.buffer = make([]*UsageRecord, 0, ul.bufferSize)
	ul.mu.Unlock()

	return ul.store.Write(records)
}

// flushRoutine 定期刷新例程
func (ul *UsageLogger) flushRoutine() {
	for range ul.flushTicker.C {
		ul.Flush()
	}
}

// Close 关闭记录器
func (ul *UsageLogger) Close() error {
	ul.flushTicker.Stop()
	return ul.Flush()
}

// Query 查询使用记录
func (ul *UsageLogger) Query(filter *UsageFilter) ([]*UsageRecord, error) {
	return ul.store.Query(filter)
}

// GetAggregatedStats 获取聚合统计
func (ul *UsageLogger) GetAggregatedStats(filter *AggregationFilter) (*AggregatedStats, error) {
	return ul.store.GetAggregated(filter)
}

// Subscribe 订阅使用记录
func (ul *UsageLogger) Subscribe(userID string) <-chan *UsageRecord {
	ul.mu.Lock()
	defer ul.mu.Unlock()

	ch := make(chan *UsageRecord, 100)
	ul.channels[userID] = ch
	return ch
}

// Unsubscribe 取消订阅
func (ul *UsageLogger) Unsubscribe(userID string) {
	ul.mu.Lock()
	defer ul.mu.Unlock()

	if ch, exists := ul.channels[userID]; exists {
		close(ch)
		delete(ul.channels, userID)
	}
}

// broadcastRecord 广播记录到订阅者
func (ul *UsageLogger) broadcastRecord(record *UsageRecord) {
	ul.mu.RLock()
	defer ul.mu.RUnlock()

	if ch, exists := ul.channels[record.UserID]; exists {
		select {
		case ch <- record:
		default:
			// 通道已满，跳过
		}
	}
}

// GetDailyStats 获取日统计
func (ul *UsageLogger) GetDailyStats(userID string, date time.Time) (*AggregatedStats, error) {
	startTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endTime := startTime.AddDate(0, 0, 1)

	return ul.store.GetAggregated(&AggregationFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   "day",
	})
}

// GetMonthlyStats 获取月统计
func (ul *UsageLogger) GetMonthlyStats(userID string, year int, month time.Month) (*AggregatedStats, error) {
	startTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 1, 0)

	return ul.store.GetAggregated(&AggregationFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   "month",
	})
}

// GetModelStats 获取模型统计
func (ul *UsageLogger) GetModelStats(userID string, startTime, endTime time.Time) (map[string]*AggregatedStats, error) {
	stats, err := ul.store.GetAggregated(&AggregationFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   "model",
	})
	if err != nil {
		return nil, err
	}

	// 返回按模型分组的统计
	result := make(map[string]*AggregatedStats)
	result["total"] = stats
	return result, nil
}

// Cleanup 清理旧记录
func (ul *UsageLogger) Cleanup(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	return ul.store.DeleteOldRecords(cutoffTime)
}


