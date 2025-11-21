package relay

import (
	"testing"
	"time"
)

func TestLoadBalancerWeightedRoundRobin(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Weight = 3
	ch1.Ability.SupportedModels = []string{"gpt-4"}

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Weight = 1
	ch2.Ability.SupportedModels = []string{"gpt-4"}

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	config := DefaultLoadBalancerConfig()
	config.Strategy = LBStrategyWeightedRoundRobin

	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	distribution := make(map[string]int)
	for i := 0; i < 400; i++ {
		ch, err := lb.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}
		distribution[ch.ID]++
	}

	// ch1 应该被选择约 75% (3/(3+1))
	ch1Count := distribution["ch-1"]
	if ch1Count < 300 || ch1Count > 350 {
		t.Errorf("Expected ch-1 ~300 selections, got %d", ch1Count)
	}
}

func TestLoadBalancerRandom(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	config := DefaultLoadBalancerConfig()
	config.Strategy = LBStrategyRandom

	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	distribution := make(map[string]int)
	for i := 0; i < 100; i++ {
		ch, err := lb.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}
		distribution[ch.ID]++
	}

	if len(distribution) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(distribution))
	}
}

func TestLoadBalancerLeastConnection(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.RecordConcurrency(5)

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}
	ch2.RecordConcurrency(2)

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	config := DefaultLoadBalancerConfig()
	config.Strategy = LBStrategyLeastConnection

	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	for i := 0; i < 10; i++ {
		ch, err := lb.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		if ch.ID != "ch-2" {
			t.Errorf("Expected ch-2, got %s", ch.ID)
		}
	}
}

func TestLoadBalancerLowestLatency(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.Metrics.AvgLatency = 500.0

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}
	ch2.Metrics.AvgLatency = 100.0

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	config := DefaultLoadBalancerConfig()
	config.Strategy = LBStrategyLowestLatency

	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	for i := 0; i < 10; i++ {
		ch, err := lb.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		if ch.ID != "ch-2" {
			t.Errorf("Expected ch-2, got %s", ch.ID)
		}
	}
}

func TestLoadBalancerWeightedByLatency(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.Metrics.AvgLatency = 200.0

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}
	ch2.Metrics.AvgLatency = 100.0

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	config := DefaultLoadBalancerConfig()
	config.Strategy = LBStrategyWeightedByLatency

	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	distribution := make(map[string]int)
	for i := 0; i < 100; i++ {
		ch, err := lb.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}
		distribution[ch.ID]++
	}

	// ch2 (low latency) 应该被选择更多
	if distribution["ch-2"] <= distribution["ch-1"] {
		t.Errorf("Expected ch-2 to be selected more than ch-1")
	}
}

func TestLoadBalancerRecordRequest(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "openai")
	cache.AddChannel(ch)

	config := DefaultLoadBalancerConfig()
	config.EnableCircuitBreaker = true

	lb := NewLoadBalancer(cache, config)

	// 记录成功
	if err := lb.RecordRequest("ch-1", true, 100); err != nil {
		t.Errorf("RecordRequest failed: %v", err)
	}

	// 记录失败
	if err := lb.RecordRequest("ch-1", false, 0); err != nil {
		t.Errorf("RecordRequest failed: %v", err)
	}
}

func TestLoadBalancerCircuitBreaker(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "openai")
	ch.Ability.SupportedModels = []string{"gpt-4"}
	cache.AddChannel(ch)

	config := DefaultLoadBalancerConfig()
	config.EnableCircuitBreaker = true
	config.CircuitBreakerFailureThreshold = 3

	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 记录 3 次失败
	for i := 0; i < 3; i++ {
		lb.RecordRequest("ch-1", false, 0)
	}

	// 第 4 次选择应该失败（断路器打开）
	_, err := lb.SelectChannel(options)
	if err == nil {
		t.Errorf("Expected selection to fail after circuit breaker opens")
	}
}

func TestLoadBalancerStatistics(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "openai")
	ch.Ability.SupportedModels = []string{"gpt-4"}
	cache.AddChannel(ch)

	config := DefaultLoadBalancerConfig()
	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	for i := 0; i < 10; i++ {
		_, _ = lb.SelectChannel(options)
	}

	stats := lb.GetStatistics()

	if total, ok := stats["total_requests"].(int64); !ok || total != 10 {
		t.Errorf("Expected 10 total requests")
	}
}

func TestLoadBalancerManager(t *testing.T) {
	manager := NewLoadBalancerManager()

	cache := NewChannelCache(ChannelCacheLevelMemory)
	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "openai")
	ch.Ability.SupportedModels = []string{"gpt-4"}
	cache.AddChannel(ch)

	config := DefaultLoadBalancerConfig()
	lb := NewLoadBalancer(cache, config)
	manager.RegisterBalancer("openai", lb)

	retrieved := manager.GetBalancer("openai")
	if retrieved == nil {
		t.Errorf("Expected balancer to be registered")
	}

	stats := manager.GetAllStatistics()
	if len(stats) != 1 {
		t.Errorf("Expected 1 balancer in statistics")
	}
}

func TestLoadBalancerConfig(t *testing.T) {
	config := DefaultLoadBalancerConfig()

	if config.Strategy != LBStrategyWeightedRoundRobin {
		t.Errorf("Expected WeightedRoundRobin strategy")
	}

	if !config.EnableHealthCheck {
		t.Errorf("Expected health check to be enabled")
	}

	if !config.EnableCircuitBreaker {
		t.Errorf("Expected circuit breaker to be enabled")
	}
}

func BenchmarkLoadBalancerSelection(b *testing.B) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建 100 个渠道
	for i := 0; i < 100; i++ {
		ch := NewChannel("ch-"+string(rune(i)), "Channel "+string(rune(i)), "https://api"+string(rune(i))+".test.com", "openai")
		ch.Weight = (i % 10) + 1
		ch.Ability.SupportedModels = []string{"gpt-4"}
		cache.AddChannel(ch)
	}

	config := DefaultLoadBalancerConfig()
	lb := NewLoadBalancer(cache, config)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lb.SelectChannel(options)
	}
}

