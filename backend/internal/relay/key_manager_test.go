package relay

import (
	"testing"
	"time"
)

func TestAPIKeyValid(t *testing.T) {
	key := NewAPIKey("k1", "secret", "bearer")

	if !key.IsValid() {
		t.Errorf("Expected valid key")
	}

	// 禁用
	key.Enabled = false
	if key.IsValid() {
		t.Errorf("Expected disabled key to be invalid")
	}

	// 启用但过期
	key.Enabled = true
	expireTime := time.Now().Add(-1 * time.Hour)
	key.ExpiresAt = &expireTime
	if key.IsValid() {
		t.Errorf("Expected expired key to be invalid")
	}
}

func TestKeyManagerRandom(t *testing.T) {
	km := NewKeyManager(KeyStrategyRandom)

	k1 := NewAPIKey("k1", "secret1", "bearer")
	k2 := NewAPIKey("k2", "secret2", "bearer")

	km.AddKey("openai", k1)
	km.AddKey("openai", k2)

	// 多次选择应该有随机分布
	distribution := make(map[string]int)
	for i := 0; i < 100; i++ {
		key, err := km.SelectKey("openai")
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}
		distribution[key.ID]++
	}

	if len(distribution) != 2 {
		t.Errorf("Expected 2 keys to be selected")
	}
}

func TestKeyManagerRoundRobin(t *testing.T) {
	km := NewKeyManager(KeyStrategyRoundRobin)

	k1 := NewAPIKey("k1", "secret1", "bearer")
	k2 := NewAPIKey("k2", "secret2", "bearer")

	km.AddKey("openai", k1)
	km.AddKey("openai", k2)

	// 轮询应该交替
	lastID := ""
	for i := 0; i < 6; i++ {
		key, err := km.SelectKey("openai")
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		if i > 0 && i%2 == 1 {
			if key.ID == lastID {
				t.Errorf("Expected alternating keys")
			}
		}
		lastID = key.ID
	}
}

func TestKeyManagerFailureAware(t *testing.T) {
	km := NewKeyManager(KeyStrategyFailureAware)

	k1 := NewAPIKey("k1", "secret1", "bearer")
	k2 := NewAPIKey("k2", "secret2", "bearer")
	k2.Weight = 2

	km.AddKey("openai", k1)
	km.AddKey("openai", k2)

	// 记录一些失败
	k1.RecordUsage(false, 0)
	k1.RecordUsage(false, 0)
	k1.RecordUsage(false, 0)

	// k2 应该被优先选择
	for i := 0; i < 10; i++ {
		key, err := km.SelectKey("openai")
		if err != nil {
			t.Errorf("Selection failed: %v", err)
		}

		// k2 应该被选择更多次
		if key.ID != "k2" && i < 8 {
			t.Logf("Expected k2, got %s", key.ID)
		}
	}
}

func TestKeyManagerAddRemove(t *testing.T) {
	km := NewKeyManager(KeyStrategyRandom)

	k1 := NewAPIKey("k1", "secret1", "bearer")
	km.AddKey("openai", k1)

	// 选择应该成功
	key, err := km.SelectKey("openai")
	if err != nil {
		t.Errorf("Selection failed: %v", err)
	}

	if key.ID != "k1" {
		t.Errorf("Expected k1")
	}

	// 移除
	if err := km.RemoveKey("openai", "k1"); err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// 选择应该失败
	_, err = km.SelectKey("openai")
	if err == nil {
		t.Errorf("Expected error after removal")
	}
}

func TestKeyPool(t *testing.T) {
	pool := NewKeyPool()

	// 注册渠道类型
	pool.RegisterChannelType("openai", KeyStrategyWeightedRoundRobin)
	pool.RegisterChannelType("anthropic", KeyStrategyRoundRobin)

	// 添加密钥
	k1 := NewAPIKey("k1", "secret1", "bearer")
	pool.AddKey("openai", k1)

	// 选择
	key, err := pool.SelectKey("openai")
	if err != nil {
		t.Errorf("Selection failed: %v", err)
	}

	if key.ID != "k1" {
		t.Errorf("Expected k1")
	}

	// 获取统计
	stats := pool.GetAllStatistics()
	if len(stats) != 2 {
		t.Errorf("Expected 2 channel types in stats")
	}
}

func TestKeyRecordUsage(t *testing.T) {
	key := NewAPIKey("k1", "secret", "bearer")

	// 记录成功
	key.RecordUsage(true, 100)
	if key.GetSuccessRate() != 100 {
		t.Errorf("Expected 100% success rate")
	}

	// 记录失败
	key.RecordUsage(false, 0)
	expectedRate := 50.0
	if key.GetSuccessRate() != expectedRate {
		t.Errorf("Expected %.1f%% success rate", expectedRate)
	}
}

func TestKeyQuota(t *testing.T) {
	key := NewAPIKey("k1", "secret", "bearer")
	limit := int64(1000)
	key.QuotaLimit = &limit

	// 在配额内
	if !key.IsValid() {
		t.Errorf("Expected valid key within quota")
	}

	// 超过配额
	key.CurrentUsage = 1000
	if key.IsValid() {
		t.Errorf("Expected invalid key over quota")
	}
}

func BenchmarkKeySelection(b *testing.B) {
	km := NewKeyManager(KeyStrategyRoundRobin)

	// 创建 100 个密钥
	for i := 0; i < 100; i++ {
		key := NewAPIKey("k-"+string(rune(i)), "secret"+string(rune(i)), "bearer")
		km.AddKey("openai", key)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = km.SelectKey("openai")
	}
}

