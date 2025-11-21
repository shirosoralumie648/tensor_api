package relay

import (
	"testing"
	"time"
)

func TestChannelSelectorRandom(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建测试渠道
	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 多次选择，应该有随机分布
	distribution := make(map[string]int)
	for i := 0; i < 100; i++ {
		ch, err := selector.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}
		distribution[ch.ID]++
	}

	// 两个渠道都应该被选择
	if len(distribution) != 2 {
		t.Errorf("Expected 2 channels to be selected, got %d", len(distribution))
	}

	// 两个渠道的选择次数应该都>0
	for _, count := range distribution {
		if count == 0 {
			t.Errorf("Expected all channels to be selected at least once")
		}
	}
}

func TestChannelSelectorRoundRobin(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建测试渠道
	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyRoundRobin)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 轮询选择应该交替
	lastID := ""
	for i := 0; i < 6; i++ {
		ch, err := selector.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		if i > 0 && i%2 == 1 {
			// 奇数位置应该与前一个不同
			if ch.ID == lastID {
				t.Errorf("Expected alternating channels in round-robin")
			}
		}
		lastID = ch.ID
	}
}

func TestChannelSelectorWeightedRoundRobin(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建权重不同的渠道
	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Weight = 3
	ch1.Ability.SupportedModels = []string{"gpt-4"}

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Weight = 1
	ch2.Ability.SupportedModels = []string{"gpt-4"}

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyWeightedRoundRobin)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 多次选择，ch1 应该被选择 3 倍
	distribution := make(map[string]int)
	for i := 0; i < 400; i++ {
		ch, err := selector.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}
		distribution[ch.ID]++
	}

	// 大约 75% 应该是 ch1（3/(3+1) = 75%）
	ch1Count := distribution["ch-1"]
	expectedCh1 := 300 // 大约 75% 的 400

	// 允许 ±50 的误差范围
	if ch1Count < expectedCh1-50 || ch1Count > expectedCh1+50 {
		t.Errorf("Expected ~%d selections of ch-1, got %d", expectedCh1, ch1Count)
	}
}

func TestChannelSelectorLeastConnection(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建测试渠道
	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.RecordConcurrency(5)

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}
	ch2.RecordConcurrency(2)

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyLeastConnection)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 应该总是选择 ch2（连接较少）
	for i := 0; i < 10; i++ {
		ch, err := selector.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		if ch.ID != "ch-2" {
			t.Errorf("Expected ch-2 (least connection), got %s", ch.ID)
		}
	}
}

func TestChannelSelectorLowestLatency(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建测试渠道
	ch1 := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.Metrics.AvgLatency = 500.0

	ch2 := NewChannel("ch-2", "Channel 2", "https://api2.test.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}
	ch2.Metrics.AvgLatency = 100.0

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyLowestLatency)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 应该总是选择 ch2（延迟较低）
	for i := 0; i < 10; i++ {
		ch, err := selector.SelectChannel(options)
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		if ch.ID != "ch-2" {
			t.Errorf("Expected ch-2 (lowest latency), got %s", ch.ID)
		}
	}
}

func TestWildcardRuleMatching(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)
	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	tests := []struct {
		pattern string
		text    string
		matches bool
	}{
		{"gpt-*", "gpt-4", true},
		{"gpt-*", "gpt-3.5", true},
		{"gpt-*", "claude-3", false},
		{"*-vision", "gpt-4-vision", true},
		{"*-vision", "gpt-4", false},
		{"claude-*", "claude-3-opus", true},
		{"*", "anything", true},
		{"exact-match", "exact-match", true},
		{"exact-match", "no-match", false},
	}

	for _, tt := range tests {
		result := selector.matchPattern(tt.text, tt.pattern)
		if result != tt.matches {
			t.Errorf("Pattern %s against %s: expected %v, got %v",
				tt.pattern, tt.text, tt.matches, result)
		}
	}
}

func TestWildcardRuleApplication(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建渠道
	ch1 := NewChannel("ch-1", "OpenAI 1", "https://api1.openai.com", "openai")
	ch1.Ability.SupportedModels = []string{"gpt-4", "gpt-3.5"}
	ch2 := NewChannel("ch-2", "OpenAI 2", "https://api2.openai.com", "openai")
	ch2.Ability.SupportedModels = []string{"gpt-4"}

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	// 创建规则
	rule := &WildcardRule{
		ID:               "gpt-rule",
		Pattern:          "gpt-*",
		ChannelType:      "openai",
		PriorityChannels: []string{"ch-1"},
		Weight:           10,
		Enabled:          true,
	}

	if err := selector.AddWildcardRule(rule); err != nil {
		t.Errorf("Failed to add rule: %v", err)
	}

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 应该优先选择 ch-1
	ch, err := selector.SelectChannel(options)
	if err != nil {
		t.Errorf("Selection failed: %v", err)
	}

	if ch.ID != "ch-1" {
		t.Errorf("Expected ch-1 (priority channel), got %s", ch.ID)
	}
}

func TestChannelSelectorManager(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch.Ability.SupportedModels = []string{"gpt-4"}
	cache.AddChannel(ch)

	manager := NewChannelSelectorManager(cache)

	// 设置 OpenAI 的选择策略
	manager.SetStrategy("openai", SelectorStrategyWeightedRoundRobin)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	selected, err := manager.SelectChannel(options)
	if err != nil {
		t.Errorf("Selection failed: %v", err)
	}

	if selected.ID != "ch-1" {
		t.Errorf("Expected ch-1, got %s", selected.ID)
	}

	// 验证规则管理器
	ruleManager := manager.GetRuleManager()
	if ruleManager == nil {
		t.Errorf("Rule manager is nil")
	}
}

func TestWildcardRuleManager(t *testing.T) {
	manager := NewWildcardRuleManager()

	rule := &WildcardRule{
		ID:               "rule-1",
		Pattern:          "gpt-*",
		ChannelType:      "openai",
		PriorityChannels: []string{"ch-1"},
		Enabled:          true,
	}

	// 添加规则
	if err := manager.AddRule(rule); err != nil {
		t.Errorf("Failed to add rule: %v", err)
	}

	// 获取规则
	retrieved, err := manager.GetRule("rule-1")
	if err != nil {
		t.Errorf("Failed to get rule: %v", err)
	}

	if retrieved.ID != rule.ID {
		t.Errorf("Rule ID mismatch")
	}

	// 获取所有规则
	all := manager.GetAllRules()
	if len(all) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(all))
	}

	// 移除规则
	if err := manager.RemoveRule("rule-1"); err != nil {
		t.Errorf("Failed to remove rule: %v", err)
	}

	// 验证已移除
	_, err = manager.GetRule("rule-1")
	if err == nil {
		t.Errorf("Expected error after removal")
	}
}

func TestChannelSelectorStatistics(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api1.test.com", "openai")
	ch.Ability.SupportedModels = []string{"gpt-4"}
	cache.AddChannel(ch)

	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	// 进行多次选择
	for i := 0; i < 10; i++ {
		_, _ = selector.SelectChannel(options)
	}

	stats := selector.GetStatistics()
	if totalSelections, ok := stats["total_selections"].(int64); !ok || totalSelections != 10 {
		t.Errorf("Expected 10 total selections")
	}
}

func TestChannelSelectorNoAvailable(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)
	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	_, err := selector.SelectChannel(options)
	if err == nil {
		t.Errorf("Expected error when no channels available")
	}
}

func BenchmarkChannelSelectorRandom(b *testing.B) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建 100 个渠道
	for i := 0; i < 100; i++ {
		ch := NewChannel("ch-"+string(rune(i)), "Channel "+string(rune(i)), "https://api"+string(rune(i))+".test.com", "openai")
		ch.Ability.SupportedModels = []string{"gpt-4"}
		cache.AddChannel(ch)
	}

	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = selector.SelectChannel(options)
	}
}

func BenchmarkChannelSelectorWeightedRoundRobin(b *testing.B) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建 100 个渠道，权重不同
	for i := 0; i < 100; i++ {
		ch := NewChannel("ch-"+string(rune(i)), "Channel "+string(rune(i)), "https://api"+string(rune(i))+".test.com", "openai")
		ch.Weight = (i % 10) + 1
		ch.Ability.SupportedModels = []string{"gpt-4"}
		cache.AddChannel(ch)
	}

	selector := NewChannelSelector(cache, SelectorStrategyWeightedRoundRobin)

	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = selector.SelectChannel(options)
	}
}

func TestChannelSelectorFiltering(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	// 创建不同区域的渠道
	ch1 := NewChannel("ch-1", "US Channel", "https://api1.test.com", "openai")
	ch1.Region = "us"
	ch1.Ability.SupportedModels = []string{"gpt-4"}
	ch1.SetStatus(ChannelStatusHealthy)

	ch2 := NewChannel("ch-2", "EU Channel", "https://api2.test.com", "openai")
	ch2.Region = "eu"
	ch2.Ability.SupportedModels = []string{"gpt-4"}
	ch2.SetStatus(ChannelStatusHealthy)

	cache.AddChannel(ch1)
	cache.AddChannel(ch2)

	selector := NewChannelSelector(cache, SelectorStrategyRandom)

	// 选择 US 区域
	options := &ChannelSelectOptions{
		ChannelType: "openai",
		Model:       "gpt-4",
		Region:      "us",
	}

	ch, err := selector.SelectChannel(options)
	if err != nil {
		t.Errorf("Selection failed: %v", err)
	}

	if ch.Region != "us" {
		t.Errorf("Expected US channel, got %s region", ch.Region)
	}
}

