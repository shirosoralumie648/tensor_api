package relay

import (
	"testing"
	"time"
)

func TestChannelCreation(t *testing.T) {
	ch := NewChannel("ch-1", "OpenAI Channel", "https://api.openai.com", "openai")

	if ch.ID != "ch-1" {
		t.Errorf("Expected ID ch-1, got %s", ch.ID)
	}

	if ch.Name != "OpenAI Channel" {
		t.Errorf("Expected name OpenAI Channel, got %s", ch.Name)
	}

	if ch.Type != "openai" {
		t.Errorf("Expected type openai, got %s", ch.Type)
	}

	if ch.GetStatus() != ChannelStatusHealthy {
		t.Errorf("Expected status healthy, got %v", ch.GetStatus())
	}

	if !ch.IsAvailable() {
		t.Errorf("Expected channel to be available")
	}
}

func TestChannelStatus(t *testing.T) {
	ch := NewChannel("ch-1", "Test Channel", "https://api.test.com", "test")

	// 测试初始状态
	if ch.GetStatus() != ChannelStatusHealthy {
		t.Errorf("Expected initial status healthy")
	}

	// 测试状态变更
	ch.SetStatus(ChannelStatusDegraded)
	if ch.GetStatus() != ChannelStatusDegraded {
		t.Errorf("Expected status degraded")
	}

	// 测试禁用
	ch.Enabled = false
	if ch.GetStatus() != ChannelStatusDisabled {
		t.Errorf("Expected status disabled when channel is disabled")
	}
}

func TestChannelMetrics(t *testing.T) {
	ch := NewChannel("ch-1", "Test Channel", "https://api.test.com", "test")

	// 初始指标
	if ch.Metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 initial requests")
	}

	// 记录成功
	ch.RecordSuccess(100)
	if ch.Metrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request after success")
	}

	if ch.GetSuccessRate() != 100 {
		t.Errorf("Expected 100% success rate")
	}

	// 记录失败
	ch.RecordFailure()
	if ch.Metrics.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests")
	}

	expectedRate := 50.0
	if ch.GetSuccessRate() != expectedRate {
		t.Errorf("Expected %.1f%% success rate, got %.1f%%", expectedRate, ch.GetSuccessRate())
	}
}

func TestChannelSupportModel(t *testing.T) {
	ch := NewChannel("ch-1", "Test Channel", "https://api.test.com", "test")
	ch.Ability.SupportedModels = []string{"gpt-4", "gpt-3.5-turbo"}

	if !ch.SupportModel("gpt-4") {
		t.Errorf("Expected channel to support gpt-4")
	}

	if !ch.SupportModel("gpt-3.5-turbo") {
		t.Errorf("Expected channel to support gpt-3.5-turbo")
	}

	if ch.SupportModel("davinci") {
		t.Errorf("Expected channel not to support davinci")
	}
}

func TestChannelFilter(t *testing.T) {
	ch := NewChannel("ch-1", "OpenAI", "https://api.openai.com", "openai")
	ch.Ability.SupportedModels = []string{"gpt-4"}
	ch.Region = "us-east-1"

	tests := []struct {
		name      string
		filter    *ChannelFilter
		expected  bool
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: true,
		},
		{
			name: "type filter match",
			filter: &ChannelFilter{
				Type: "openai",
			},
			expected: true,
		},
		{
			name: "type filter no match",
			filter: &ChannelFilter{
				Type: "claude",
			},
			expected: false,
		},
		{
			name: "model filter match",
			filter: &ChannelFilter{
				Model: "gpt-4",
			},
			expected: true,
		},
		{
			name: "model filter no match",
			filter: &ChannelFilter{
				Model: "claude-3",
			},
			expected: false,
		},
		{
			name: "region filter match",
			filter: &ChannelFilter{
				Region: "us-east-1",
			},
			expected: true,
		},
		{
			name: "multiple filters match",
			filter: &ChannelFilter{
				Type:   "openai",
				Model:  "gpt-4",
				Region: "us-east-1",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ch.Matches(tt.filter)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestChannelCacheAddRemove(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "type1")
	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "type2")

	// 测试添加
	if err := cc.AddChannel(ch1); err != nil {
		t.Errorf("Failed to add channel: %v", err)
	}

	if err := cc.AddChannel(ch2); err != nil {
		t.Errorf("Failed to add channel: %v", err)
	}

	// 测试获取
	retrieved, err := cc.GetChannel("ch-1")
	if err != nil {
		t.Errorf("Failed to get channel: %v", err)
	}

	if retrieved.Name != "Channel 1" {
		t.Errorf("Expected Channel 1, got %s", retrieved.Name)
	}

	// 测试移除
	if err := cc.RemoveChannel("ch-1"); err != nil {
		t.Errorf("Failed to remove channel: %v", err)
	}

	_, err = cc.GetChannel("ch-1")
	if err == nil {
		t.Errorf("Expected error when getting removed channel")
	}
}

func TestChannelCacheIndexing(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "OpenAI", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4", "gpt-3.5"}
	ch1.Region = "us"

	ch2 := NewChannel("ch-2", "Claude", "https://api2.test.com", "anthropic")
	ch2.Ability.SupportedModels = []string{"claude-3"}
	ch2.Region = "eu"

	cc.AddChannel(ch1)
	cc.AddChannel(ch2)

	// 测试按类型索引
	openaiChannels := cc.GetChannelsByType("openai")
	if len(openaiChannels) != 1 || openaiChannels[0].ID != "ch-1" {
		t.Errorf("Expected 1 openai channel")
	}

	// 测试按模型索引
	gpt4Channels := cc.GetChannelsByModel("gpt-4")
	if len(gpt4Channels) != 1 || gpt4Channels[0].ID != "ch-1" {
		t.Errorf("Expected 1 channel supporting gpt-4")
	}

	// 测试按地区索引
	usChannels := cc.GetChannelsByRegion("us")
	if len(usChannels) != 1 || usChannels[0].ID != "ch-1" {
		t.Errorf("Expected 1 us region channel")
	}
}

func TestChannelCacheFilter(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "OpenAI", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.Region = "us"
	ch1.Priority = 10
	ch1.Enabled = true

	ch2 := NewChannel("ch-2", "Claude", "https://api2.test.com", "anthropic")
	ch2.Ability.SupportedModels = []string{"claude-3"}
	ch2.Region = "eu"
	ch2.Priority = 20
	ch2.Enabled = true

	cc.AddChannel(ch1)
	cc.AddChannel(ch2)

	// 过滤 openai 类型
	filtered := cc.FilterChannels(&ChannelFilter{
		Type:        "openai",
		OnlyEnabled: true,
	})
	if len(filtered) != 1 || filtered[0].ID != "ch-1" {
		t.Errorf("Expected 1 openai channel")
	}

	// 过滤 gpt-4 支持
	filtered = cc.FilterChannels(&ChannelFilter{
		Model: "gpt-4",
	})
	if len(filtered) != 1 || filtered[0].ID != "ch-1" {
		t.Errorf("Expected 1 channel supporting gpt-4")
	}
}

func TestChannelCacheStatistics(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "type1")
	cc.AddChannel(ch1)

	stats := cc.GetStatistics()

	if channelCount, ok := stats["channel_count"].(int); !ok || channelCount != 1 {
		t.Errorf("Expected 1 channel in stats")
	}
}

func TestChannelCacheRefresh(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	channels := make([]*Channel, 2)
	channels[0] = NewChannel("ch-1", "Channel 1", "https://api1.test.com", "type1")
	channels[1] = NewChannel("ch-2", "Channel 2", "https://api2.test.com", "type2")

	if err := cc.RefreshCache(channels); err != nil {
		t.Errorf("Failed to refresh cache: %v", err)
	}

	allChannels := cc.GetAllChannels()
	if len(allChannels) != 2 {
		t.Errorf("Expected 2 channels after refresh, got %d", len(allChannels))
	}
}

func TestChannelCacheClear(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "type1")
	cc.AddChannel(ch1)

	allChannels := cc.GetAllChannels()
	if len(allChannels) != 1 {
		t.Errorf("Expected 1 channel before clear")
	}

	cc.ClearCache()

	allChannels = cc.GetAllChannels()
	if len(allChannels) != 0 {
		t.Errorf("Expected 0 channels after clear")
	}
}

func TestChannelCacheManager(t *testing.T) {
	// 模拟数据源
	dataSource := func() ([]*Channel, error) {
		channels := make([]*Channel, 2)
		channels[0] = NewChannel("ch-1", "Channel 1", "https://api1.test.com", "type1")
		channels[1] = NewChannel("ch-2", "Channel 2", "https://api2.test.com", "type2")
		return channels, nil
	}

	manager := NewChannelCacheManager(dataSource)
	manager.SetRefreshInterval(100 * time.Millisecond)

	if err := manager.Start(); err != nil {
		t.Errorf("Failed to start manager: %v", err)
	}

	defer manager.Stop()

	// 等待初始化
	time.Sleep(50 * time.Millisecond)

	cache := manager.GetCache()
	allChannels := cache.GetAllChannels()

	if len(allChannels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(allChannels))
	}
}

func BenchmarkChannelCacheGetChannel(b *testing.B) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	for i := 0; i < 1000; i++ {
		ch := NewChannel(
			"ch-"+string(rune(i)),
			"Channel "+string(rune(i)),
			"https://api"+string(rune(i))+".test.com",
			"type"+string(rune(i%5)),
		)
		cc.AddChannel(ch)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cc.GetChannel("ch-500")
	}
}

func BenchmarkChannelCacheFilter(b *testing.B) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	for i := 0; i < 100; i++ {
		ch := NewChannel(
			"ch-"+string(rune(i)),
			"Channel "+string(rune(i)),
			"https://api"+string(rune(i))+".test.com",
			"openai",
		)
		ch.Ability.SupportedModels = []string{"gpt-4"}
		ch.Region = "us"
		cc.AddChannel(ch)
	}

	filter := &ChannelFilter{
		Type:   "openai",
		Model:  "gpt-4",
		Region: "us",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cc.FilterChannels(filter)
	}
}

func TestChannelStatusString(t *testing.T) {
	tests := []struct {
		status   ChannelStatus
		expected string
	}{
		{ChannelStatusHealthy, "healthy"},
		{ChannelStatusDegraded, "degraded"},
		{ChannelStatusUnavailable, "unavailable"},
		{ChannelStatusDisabled, "disabled"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.status.String())
		}
	}
}

func TestChannelConcurrentOps(t *testing.T) {
	cc := NewChannelCache(ChannelCacheLevelMemory)

	done := make(chan error, 10)

	// 并发添加和获取
	for i := 0; i < 10; i++ {
		go func(id int) {
			ch := NewChannel(
				"ch-"+string(rune(id)),
				"Channel "+string(rune(id)),
				"https://api"+string(rune(id))+".test.com",
				"type",
			)
			if err := cc.AddChannel(ch); err != nil {
				done <- err
				return
			}

			_, err := cc.GetChannel("ch-" + string(rune(id)))
			done <- err
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}

