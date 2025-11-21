package adapter

import (
	"testing"
	"time"
)

func TestAdapterRegistry(t *testing.T) {
	registry := NewAdapterRegistry()

	// 测试注册
	err := registry.Register("test-adapter", func(config *AdapterConfig) Adapter {
		return NewOpenAIAdapter(config)
	}, "v1.0.0")
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// 测试重复注册
	err = registry.Register("test-adapter", func(config *AdapterConfig) Adapter {
		return NewOpenAIAdapter(config)
	}, "v1.0.0")
	if err == nil {
		t.Errorf("Expected error for duplicate registration")
	}

	// 测试列表
	adapters := registry.List()
	if len(adapters) == 0 {
		t.Errorf("Expected adapters in registry")
	}

	// 测试版本
	version, err := registry.GetVersion("test-adapter")
	if err != nil {
		t.Errorf("GetVersion failed: %v", err)
	}

	if version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %s", version)
	}

	// 测试卸载
	err = registry.Unregister("test-adapter")
	if err != nil {
		t.Errorf("Unregister failed: %v", err)
	}

	// 测试卸载不存在的适配器
	err = registry.Unregister("test-adapter")
	if err == nil {
		t.Errorf("Expected error for unregistering non-existent adapter")
	}
}

func TestCreateAdapter(t *testing.T) {
	registry := NewAdapterRegistry()

	registry.Register("openai", func(config *AdapterConfig) Adapter {
		return NewOpenAIAdapter(config)
	}, "v1.0.0")

	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter, err := registry.Create("openai", config)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if adapter == nil {
		t.Errorf("Expected adapter instance")
	}

	if adapter.Name() != "openai" {
		t.Errorf("Expected adapter name openai")
	}
}

func TestCreateNonExistentAdapter(t *testing.T) {
	registry := NewAdapterRegistry()

	config := &AdapterConfig{
		Type:    "non-existent",
		BaseURL: "https://api.example.com",
		Timeout: 30 * time.Second,
	}

	adapter, err := registry.Create("non-existent", config)
	if err == nil {
		t.Errorf("Expected error for non-existent adapter")
	}

	if adapter != nil {
		t.Errorf("Expected nil adapter")
	}
}

func TestUpdateAdapter(t *testing.T) {
	registry := NewAdapterRegistry()

	registry.Register("test", func(config *AdapterConfig) Adapter {
		return NewOpenAIAdapter(config)
	}, "v1.0.0")

	// 更新适配器
	err := registry.Update("test", func(config *AdapterConfig) Adapter {
		return NewClaudeAdapter(config)
	}, "v2.0.0")
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	version, _ := registry.GetVersion("test")
	if version != "v2.0.0" {
		t.Errorf("Expected version v2.0.0, got %s", version)
	}
}

func TestUpdateNonExistentAdapter(t *testing.T) {
	registry := NewAdapterRegistry()

	err := registry.Update("non-existent", func(config *AdapterConfig) Adapter {
		return NewOpenAIAdapter(config)
	}, "v1.0.0")

	if err == nil {
		t.Errorf("Expected error for updating non-existent adapter")
	}
}

func TestGlobalRegistry(t *testing.T) {
	registry := GetGlobalRegistry()

	if registry == nil {
		t.Errorf("Expected global registry")
	}

	adapters := ListAdapters()
	if len(adapters) == 0 {
		t.Errorf("Expected adapters in global registry")
	}
}

func TestCoreAdaptersRegistration(t *testing.T) {
	registry := GetGlobalRegistry()

	coreAdapters := []string{"openai", "claude", "gemini", "baidu", "qwen"}

	for _, name := range coreAdapters {
		version, err := registry.GetVersion(name)
		if err != nil {
			t.Errorf("Expected %s adapter to be registered", name)
		}

		if version == "" {
			t.Errorf("Expected version for %s", name)
		}
	}
}

func TestBatchAdaptersRegistration(t *testing.T) {
	registry := GetGlobalRegistry()

	batchAdapters := []string{"deepseek", "moonshot", "minimax"}

	for _, name := range batchAdapters {
		version, err := registry.GetVersion(name)
		if err != nil {
			t.Errorf("Expected %s adapter to be registered", name)
		}

		if version == "" {
			t.Errorf("Expected version for %s", name)
		}
	}
}

func BenchmarkAdapterRegistry(b *testing.B) {
	registry := GetGlobalRegistry()

	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Create("openai", config)
	}
}


