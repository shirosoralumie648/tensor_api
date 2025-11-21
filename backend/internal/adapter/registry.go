package adapter

import (
	"fmt"
	"sync"
)

// AdapterRegistry 适配器注册表
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]AdapterFactory
	versions map[string]string // 版本管理
}

// AdapterFactory 适配器工厂函数类型
type AdapterFactory func(*AdapterConfig) Adapter

// NewAdapterRegistry 创建适配器注册表
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]AdapterFactory),
		versions: make(map[string]string),
	}
}

// Register 注册适配器
func (ar *AdapterRegistry) Register(name string, factory AdapterFactory, version string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.adapters[name]; exists {
		return fmt.Errorf("adapter %s already registered", name)
	}

	ar.adapters[name] = factory
	ar.versions[name] = version

	return nil
}

// Unregister 卸载适配器
func (ar *AdapterRegistry) Unregister(name string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.adapters[name]; !exists {
		return fmt.Errorf("adapter %s not found", name)
	}

	delete(ar.adapters, name)
	delete(ar.versions, name)

	return nil
}

// Update 更新适配器（热更新）
func (ar *AdapterRegistry) Update(name string, factory AdapterFactory, version string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.adapters[name]; !exists {
		return fmt.Errorf("adapter %s not found", name)
	}

	ar.adapters[name] = factory
	ar.versions[name] = version

	return nil
}

// Create 创建适配器实例
func (ar *AdapterRegistry) Create(name string, config *AdapterConfig) (Adapter, error) {
	ar.mu.RLock()
	factory, exists := ar.adapters[name]
	ar.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("adapter %s not registered", name)
	}

	return factory(config), nil
}

// GetVersion 获取适配器版本
func (ar *AdapterRegistry) GetVersion(name string) (string, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	version, exists := ar.versions[name]
	if !exists {
		return "", fmt.Errorf("adapter %s not found", name)
	}

	return version, nil
}

// List 列出所有注册的适配器
func (ar *AdapterRegistry) List() map[string]string {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	result := make(map[string]string)
	for name, version := range ar.versions {
		result[name] = version
	}

	return result
}

// ==================== 全局注册表 ====================

var globalRegistry *AdapterRegistry

func init() {
	globalRegistry = NewAdapterRegistry()

	// 注册核心提供商
	registerCoreAdapters()
	// 注册批量提供商
	registerBatchAdapters()
}

// registerCoreAdapters 注册核心适配器
func registerCoreAdapters() {
	globalRegistry.Register("openai", func(config *AdapterConfig) Adapter {
		return NewOpenAIAdapter(config)
	}, "v1.0.0")

	globalRegistry.Register("claude", func(config *AdapterConfig) Adapter {
		return NewClaudeAdapter(config)
	}, "v1.0.0")

	globalRegistry.Register("gemini", func(config *AdapterConfig) Adapter {
		return NewGeminiAdapter(config)
	}, "v1.0.0")

	globalRegistry.Register("baidu", func(config *AdapterConfig) Adapter {
		return NewBaiduAdapter(config)
	}, "v1.0.0")

	globalRegistry.Register("qwen", func(config *AdapterConfig) Adapter {
		return NewQwenAdapter(config)
	}, "v1.0.0")
}

// registerBatchAdapters 注册批量适配器
func registerBatchAdapters() {
	globalRegistry.Register("deepseek", func(config *AdapterConfig) Adapter {
		return NewDeepSeekAdapter(config)
	}, "v1.0.0")

	globalRegistry.Register("moonshot", func(config *AdapterConfig) Adapter {
		return NewMoonshotAdapter(config)
	}, "v1.0.0")

	globalRegistry.Register("minimax", func(config *AdapterConfig) Adapter {
		return NewMinimaxAdapter(config)
	}, "v1.0.0")
}

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *AdapterRegistry {
	return globalRegistry
}

// CreateAdapter 使用全局注册表创建适配器
func CreateAdapter(name string, config *AdapterConfig) (Adapter, error) {
	return globalRegistry.Create(name, config)
}

// RegisterAdapter 向全局注册表注册适配器
func RegisterAdapter(name string, factory AdapterFactory, version string) error {
	return globalRegistry.Register(name, factory, version)
}

// UpdateAdapter 更新全局注册表中的适配器
func UpdateAdapter(name string, factory AdapterFactory, version string) error {
	return globalRegistry.Update(name, factory, version)
}

// ListAdapters 列出全局注册表中的所有适配器
func ListAdapters() map[string]string {
	return globalRegistry.List()
}


