package analytics

import (
	"testing"
	"time"
)

// MockUsageStore 模拟使用存储
type MockUsageStore struct {
	records []*UsageRecord
}

func NewMockUsageStore() *MockUsageStore {
	return &MockUsageStore{
		records: make([]*UsageRecord, 0),
	}
}

func (m *MockUsageStore) Write(records []*UsageRecord) error {
	m.records = append(m.records, records...)
	return nil
}

func (m *MockUsageStore) Query(filter *UsageFilter) ([]*UsageRecord, error) {
	var result []*UsageRecord

	for _, record := range m.records {
		if filter.UserID != "" && record.UserID != filter.UserID {
			continue
		}
		if filter.Model != "" && record.Model != filter.Model {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if !filter.StartTime.IsZero() && record.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && record.Timestamp.After(filter.EndTime) {
			continue
		}

		result = append(result, record)
	}

	return result, nil
}

func (m *MockUsageStore) GetAggregated(filter *AggregationFilter) (*AggregatedStats, error) {
	stats := &AggregatedStats{
		Timeline: []*TimelinePoint{},
	}

	for _, record := range m.records {
		if filter.UserID != "" && record.UserID != filter.UserID {
			continue
		}

		stats.TotalRequests++
		stats.TotalTokens += record.TotalTokens
		stats.TotalCost += record.Cost

		if record.Status == "success" {
			stats.SuccessRequests++
		} else {
			stats.ErrorRequests++
		}

		stats.AvgDuration += record.Duration
	}

	if stats.TotalRequests > 0 {
		stats.AvgDuration /= stats.TotalRequests
		stats.AvgTokensPerReq = stats.TotalTokens / stats.TotalRequests
		stats.AvgCostPerReq = stats.TotalCost / float64(stats.TotalRequests)
	}

	return stats, nil
}

func (m *MockUsageStore) DeleteOldRecords(before time.Time) error {
	var filtered []*UsageRecord
	for _, record := range m.records {
		if record.Timestamp.After(before) {
			filtered = append(filtered, record)
		}
	}
	m.records = filtered
	return nil
}

func TestUsageLogger(t *testing.T) {
	store := NewMockUsageStore()
	logger := NewUsageLogger(store, 100, 1*time.Second)

	record := &UsageRecord{
		UserID:          "user123",
		TokenID:         "token123",
		Model:           "gpt-4",
		Provider:        "openai",
		RequestTokens:   100,
		ResponseTokens:  50,
		TotalTokens:     150,
		Cost:            0.01,
		Duration:        100,
		Status:          "success",
		ClientIP:        "127.0.0.1",
		Endpoint:        "/chat/completions",
		Timestamp:       time.Now(),
		StreamingMode:   false,
		CacheHit:        false,
	}

	err := logger.RecordUsage(record)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	logger.Close()
}

func TestRealtimeStatsEngine(t *testing.T) {
	engine := NewRealtimeStatsEngine(1 * time.Minute)

	record := &UsageRecord{
		UserID:       "user123",
		Model:        "gpt-4",
		Provider:     "openai",
		TotalTokens:  150,
		Cost:         0.01,
		Duration:     100,
		Status:       "success",
		Timestamp:    time.Now(),
	}

	engine.RecordMetric("user123", "gpt-4", "openai", record)

	stats := engine.GetUserStats("user123")
	if stats.Requests != 1 {
		t.Errorf("Expected 1 request, got %d", stats.Requests)
	}

	if stats.SuccessRequests != 1 {
		t.Errorf("Expected 1 success, got %d", stats.SuccessRequests)
	}

	if stats.TotalTokens != 150 {
		t.Errorf("Expected 150 tokens, got %d", stats.TotalTokens)
	}
}

func TestAnalyticsAPI(t *testing.T) {
	store := NewMockUsageStore()
	logger := NewUsageLogger(store, 100, 1*time.Second)
	realtime := NewRealtimeStatsEngine(1 * time.Minute)
	api := NewAnalyticsAPI(logger, realtime)

	record := &UsageRecord{
		UserID:         "user123",
		TokenID:        "token123",
		Model:          "gpt-4",
		Provider:       "openai",
		RequestTokens:  100,
		ResponseTokens: 50,
		TotalTokens:    150,
		Cost:           0.01,
		Duration:       100,
		Status:         "success",
		ClientIP:       "127.0.0.1",
		Endpoint:       "/chat/completions",
		Timestamp:      time.Now(),
	}

	logger.RecordUsage(record)
	realtime.RecordMetric("user123", "gpt-4", "openai", record)

	// 测试查询
	req := &QueryRequest{
		UserID: "user123",
		Limit:  100,
	}

	resp, err := api.Query(req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Total == 0 {
		t.Error("Expected non-zero results")
	}
}

func TestGetRealtimeMetrics(t *testing.T) {
	store := NewMockUsageStore()
	logger := NewUsageLogger(store, 100, 1*time.Second)
	realtime := NewRealtimeStatsEngine(1 * time.Minute)
	api := NewAnalyticsAPI(logger, realtime)

	record := &UsageRecord{
		UserID:        "user123",
		Model:         "gpt-4",
		Provider:      "openai",
		TotalTokens:   150,
		Cost:          0.01,
		Duration:      100,
		Status:        "success",
		Timestamp:     time.Now(),
	}

	realtime.RecordMetric("user123", "gpt-4", "openai", record)

	metrics := api.GetRealtimeMetrics("user123")
	if metrics == nil {
		t.Error("Expected metrics")
	}

	if metrics["requests"] != int64(1) {
		t.Errorf("Expected 1 request, got %v", metrics["requests"])
	}
}

func BenchmarkRecordUsage(b *testing.B) {
	store := NewMockUsageStore()
	logger := NewUsageLogger(store, 10000, 5*time.Second)

	record := &UsageRecord{
		UserID:        "user123",
		Model:         "gpt-4",
		Provider:      "openai",
		TotalTokens:   150,
		Cost:          0.01,
		Duration:      100,
		Status:        "success",
		Timestamp:     time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.RecordUsage(record)
	}

	logger.Close()
}

func BenchmarkRealtimeStats(b *testing.B) {
	engine := NewRealtimeStatsEngine(1 * time.Minute)

	record := &UsageRecord{
		UserID:       "user123",
		Model:        "gpt-4",
		Provider:     "openai",
		TotalTokens:  150,
		Cost:         0.01,
		Duration:     100,
		Status:       "success",
		Timestamp:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.RecordMetric("user123", "gpt-4", "openai", record)
	}
}


