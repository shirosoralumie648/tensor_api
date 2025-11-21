package billing

import (
	"testing"
)

func TestTokenCounterRegisterModel(t *testing.T) {
	counter := NewTokenCounter()

	config := &ModelTokenConfig{
		ModelName:        "test-model",
		InputTokenRatio:  1.0,
		OutputTokenRatio: 1.5,
		MaxContextLength: 4096,
		ReservedTokens:   50,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	}

	if err := counter.RegisterModel(config); err != nil {
		t.Errorf("RegisterModel failed: %v", err)
	}

	retrieved, err := counter.GetModelConfig("test-model")
	if err != nil {
		t.Errorf("GetModelConfig failed: %v", err)
	}

	if retrieved.ModelName != "test-model" {
		t.Errorf("Expected test-model, got %s", retrieved.ModelName)
	}
}

func TestTokenCounterCountTokens(t *testing.T) {
	counter := NewTokenCounter()

	input := "Hello world"
	output := "The world is beautiful"

	result, err := counter.CountTokens("gpt-3.5-turbo", input, output)
	if err != nil {
		t.Errorf("CountTokens failed: %v", err)
	}

	if result == nil {
		t.Errorf("Result is nil")
	}

	if result.TotalTokens <= 0 {
		t.Errorf("Expected positive token count")
	}

	if result.InputTokens <= 0 {
		t.Errorf("Expected positive input token count")
	}

	if result.OutputTokens <= 0 {
		t.Errorf("Expected positive output token count")
	}
}

func TestTokenCounterCache(t *testing.T) {
	counter := NewTokenCounter()

	input := "Hello world"
	output := "Hello"

	// 第一次调用（缓存未命中）
	result1, _ := counter.CountTokens("gpt-4", input, output)

	// 第二次调用（应该命中缓存）
	result2, _ := counter.CountTokens("gpt-4", input, output)

	if result1.TotalTokens != result2.TotalTokens {
		t.Errorf("Cache results should be identical")
	}

	stats := counter.GetStatistics()
	if hits, ok := stats["cache_hits"].(int64); !ok || hits < 1 {
		t.Errorf("Expected cache hits")
	}
}

func TestTokenCounterWithCost(t *testing.T) {
	counter := NewTokenCounter()

	input := "Hello"
	output := "World"

	result, err := counter.CountTokensWithCost("gpt-3.5-turbo", input, output, 0.0015, 0.002)
	if err != nil {
		t.Errorf("CountTokensWithCost failed: %v", err)
	}

	if result.TotalCost <= 0 {
		t.Errorf("Expected positive cost")
	}

	if result.InputCost <= 0 {
		t.Errorf("Expected positive input cost")
	}

	if result.OutputCost <= 0 {
		t.Errorf("Expected positive output cost")
	}
}

func TestTokenCounterBatch(t *testing.T) {
	counter := NewTokenCounter()
	batch := NewTokenCountingBatch(counter)

	// 添加请求
	batch.AddRequest(&TokenCountRequest{
		ModelName:   "gpt-4",
		InputText:   "Hello",
		OutputText:  "World",
		InputPrice:  0.003,
		OutputPrice: 0.006,
	})

	batch.AddRequest(&TokenCountRequest{
		ModelName:  "gpt-3.5-turbo",
		InputText:  "Hi",
		OutputText: "Hey",
	})

	if err := batch.Process(); err != nil {
		t.Errorf("Process failed: %v", err)
	}

	results := batch.GetResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	total := batch.GetTotalResult()
	if total.TotalTokens <= 0 {
		t.Errorf("Expected positive total tokens")
	}
}

func TestTokenCounterClearCache(t *testing.T) {
	counter := NewTokenCounter()

	// 计数一些文本
	counter.CountTokens("gpt-4", "Hello", "World")

	stats1 := counter.GetStatistics()
	if entries, ok := stats1["cache_entries"].(int); !ok || entries < 1 {
		t.Errorf("Expected cache entries")
	}

	// 清空缓存
	counter.ClearCache()

	stats2 := counter.GetStatistics()
	if entries, ok := stats2["cache_entries"].(int); !ok || entries != 0 {
		t.Errorf("Expected 0 cache entries after clear")
	}
}

func TestTokenCounterStatistics(t *testing.T) {
	counter := NewTokenCounter()

	// 计数一些文本
	counter.CountTokens("gpt-4", "Hello", "World")
	counter.CountTokens("gpt-3.5-turbo", "Hi", "Hey")

	stats := counter.GetStatistics()

	if total, ok := stats["total_count"].(int64); !ok || total < 2 {
		t.Errorf("Expected at least 2 counts")
	}

	if tokens, ok := stats["total_tokens"].(int64); !ok || tokens <= 0 {
		t.Errorf("Expected positive total tokens")
	}
}

func TestTokenCounterGetConfig(t *testing.T) {
	counter := NewTokenCounter()

	// 获取默认配置
	config, err := counter.GetModelConfig("gpt-4")
	if err != nil {
		t.Errorf("GetModelConfig failed: %v", err)
	}

	if config.InputTokenRatio != 1.0 {
		t.Errorf("Expected input ratio 1.0")
	}

	// 获取不存在的模型
	_, err = counter.GetModelConfig("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent model")
	}
}

func TestTokenCounterCacheHitRate(t *testing.T) {
	counter := NewTokenCounter()

	// 计数相同文本两次（第二次应命中缓存）
	counter.CountTokens("gpt-4", "Hello", "World")
	counter.CountTokens("gpt-4", "Hello", "World")

	rate := counter.GetCacheHitRate()
	if rate < 50 {
		t.Errorf("Expected high cache hit rate, got %.1f%%", rate)
	}
}

func BenchmarkTokenCounting(b *testing.B) {
	counter := NewTokenCounter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = counter.CountTokens("gpt-4", "Hello world", "This is a response")
	}
}

func BenchmarkBatchCounting(b *testing.B) {
	counter := NewTokenCounter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := NewTokenCountingBatch(counter)
		batch.AddRequest(&TokenCountRequest{
			ModelName:  "gpt-4",
			InputText:  "Hello",
			OutputText: "World",
		})
		batch.Process()
	}
}

