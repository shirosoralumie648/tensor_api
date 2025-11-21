package billing

import (
	"testing"
	"time"
)

func TestPricingManagerRegister(t *testing.T) {
	manager := NewPricingManager()

	err := manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)
	if err != nil {
		t.Errorf("RegisterModelPrice failed: %v", err)
	}

	price, err := manager.GetModelPrice("gpt-4")
	if err != nil {
		t.Errorf("GetModelPrice failed: %v", err)
	}

	if price.InputPrice != 0.03 || price.OutputPrice != 0.06 {
		t.Errorf("Price mismatch: input=%.2f, output=%.2f", price.InputPrice, price.OutputPrice)
	}
}

func TestPricingManagerUpdate(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	// 更新价格
	err := manager.UpdateModelPrice("gpt-4", 0.04, 0.08, "Price adjustment")
	if err != nil {
		t.Errorf("UpdateModelPrice failed: %v", err)
	}

	price, _ := manager.GetModelPrice("gpt-4")
	if price.InputPrice != 0.04 || price.OutputPrice != 0.08 {
		t.Errorf("Price update failed: input=%.2f, output=%.2f", price.InputPrice, price.OutputPrice)
	}

	// 检查历史记录
	history := manager.GetPriceHistory("gpt-4")
	if len(history) != 1 {
		t.Errorf("Expected 1 history record, got %d", len(history))
	}
}

func TestPricingManagerCalculatePrice(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	// 计算价格：1000 input tokens, 1000 output tokens
	price, err := manager.CalculatePrice("gpt-4", 1000, 1000)
	if err != nil {
		t.Errorf("CalculatePrice failed: %v", err)
	}

	// 预期：(1000/1000)*0.03 + (1000/1000)*0.06 = 0.09
	expected := 0.09
	if price < expected-0.0001 || price > expected+0.0001 {
		t.Errorf("Price calculation error: expected %.4f, got %.4f", expected, price)
	}
}

func TestPricingManagerCalculatePriceByRequest(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-3.5", 0.01, 0.0, PricingByRequest)

	// 计算价格（按次计费）
	price, err := manager.CalculatePrice("gpt-3.5", 0, 0)
	if err != nil {
		t.Errorf("CalculatePrice failed: %v", err)
	}

	if price != 0.01 {
		t.Errorf("Expected 0.01, got %.4f", price)
	}
}

func TestPriceGroup(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)
	manager.RegisterModelPrice("gpt-3.5", 0.001, 0.002, PricingByToken)

	// 创建价格组
	err := manager.CreatePriceGroup("group-1", "Premium Models", []string{"gpt-4", "gpt-3.5"}, 1.5)
	if err != nil {
		t.Errorf("CreatePriceGroup failed: %v", err)
	}

	// 计算带分组倍率的价格
	price, err := manager.CalculatePriceWithGroup("gpt-4", "group-1", 1000, 1000)
	if err != nil {
		t.Errorf("CalculatePriceWithGroup failed: %v", err)
	}

	// 预期：0.09 * 1.5 = 0.135
	expected := 0.09 * 1.5
	if price < expected-0.0001 || price > expected+0.0001 {
		t.Errorf("Price with group error: expected %.4f, got %.4f", expected, price)
	}
}

func TestPricingStrategy(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	// 创建定价策略
	err := manager.CreatePricingStrategy("strategy-1", "Standard Pricing", []string{"gpt-4"})
	if err != nil {
		t.Errorf("CreatePricingStrategy failed: %v", err)
	}

	// 激活策略
	err = manager.ActivatePricingStrategy("strategy-1")
	if err != nil {
		t.Errorf("ActivatePricingStrategy failed: %v", err)
	}

	// 获取活跃策略
	strategy := manager.GetActiveStrategy()
	if strategy == nil {
		t.Errorf("Active strategy is nil")
	}

	if strategy.StrategyID != "strategy-1" {
		t.Errorf("Strategy ID mismatch: %s", strategy.StrategyID)
	}
}

func TestPricingCache(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	cache := NewPricingCache(manager, 1*time.Second)

	// 第一次查询（缓存未命中）
	price1, err := cache.GetPrice("gpt-4")
	if err != nil {
		t.Errorf("GetPrice failed: %v", err)
	}

	// 第二次查询（缓存命中）
	price2, err := cache.GetPrice("gpt-4")
	if err != nil {
		t.Errorf("GetPrice failed: %v", err)
	}

	// 检查缓存命中率
	hitRate := cache.GetCacheHitRate()
	if hitRate < 40 || hitRate > 60 {
		t.Logf("Cache hit rate: %.2f%%", hitRate)
	}

	if price1.InputPrice != price2.InputPrice {
		t.Errorf("Price mismatch between queries")
	}
}

func TestPricingCacheCalculate(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	cache := NewPricingCache(manager, 1*time.Second)

	// 计算价格（使用缓存）
	price, err := cache.CalculatePrice("gpt-4", 1000, 1000)
	if err != nil {
		t.Errorf("CalculatePrice failed: %v", err)
	}

	expected := 0.09
	if price < expected-0.0001 || price > expected+0.0001 {
		t.Errorf("Price calculation error: expected %.4f, got %.4f", expected, price)
	}
}

func TestPricingStatistics(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)
	manager.RegisterModelPrice("gpt-3.5", 0.001, 0.002, PricingByToken)
	manager.CreatePriceGroup("group-1", "Premium", []string{"gpt-4"}, 1.5)

	stats := manager.GetStatistics()

	if modelCount, ok := stats["model_count"].(int); !ok || modelCount != 2 {
		t.Errorf("Expected 2 models, got %v", stats["model_count"])
	}

	if groupCount, ok := stats["group_count"].(int); !ok || groupCount != 1 {
		t.Errorf("Expected 1 group, got %v", stats["group_count"])
	}
}

func TestPriceHistoryTracking(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	// 执行多次更新
	manager.UpdateModelPrice("gpt-4", 0.04, 0.08, "Update 1")
	manager.UpdateModelPrice("gpt-4", 0.05, 0.10, "Update 2")

	history := manager.GetPriceHistory("gpt-4")
	if len(history) != 2 {
		t.Errorf("Expected 2 history records, got %d", len(history))
	}

	// 验证版本号递增
	if history[0].Version != 2 || history[1].Version != 3 {
		t.Errorf("Version numbers incorrect: %d, %d", history[0].Version, history[1].Version)
	}
}

func TestGetAllPrices(t *testing.T) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)
	manager.RegisterModelPrice("gpt-3.5", 0.001, 0.002, PricingByToken)
	manager.RegisterModelPrice("claude", 0.015, 0.03, PricingByToken)

	prices := manager.GetAllModelPrices()
	if len(prices) != 3 {
		t.Errorf("Expected 3 prices, got %d", len(prices))
	}
}

func BenchmarkCalculatePrice(b *testing.B) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.CalculatePrice("gpt-4", 1000, 1000)
	}
}

func BenchmarkCachedCalculatePrice(b *testing.B) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)
	cache := NewPricingCache(manager, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.CalculatePrice("gpt-4", 1000, 1000)
	}
}

func BenchmarkPricingCacheGetPrice(b *testing.B) {
	manager := NewPricingManager()
	manager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)
	cache := NewPricingCache(manager, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.GetPrice("gpt-4")
	}
}

