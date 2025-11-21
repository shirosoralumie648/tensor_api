package adapter

import (
	"fmt"
	"sync"
)

// AdapterFactory 适配器工厂，用于创建和管理各个 AI 提供商的适配器
type AdapterFactory struct {
	providers map[string]AIProvider
	configs   map[string]*ProviderConfig
	mu        sync.RWMutex
}

// NewAdapterFactory 创建新的适配器工厂
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{
		providers: make(map[string]AIProvider),
		configs:   make(map[string]*ProviderConfig),
	}
}

// Register 注册一个新的 AI 提供商适配器
func (f *AdapterFactory) Register(name string, provider AIProvider, config *ProviderConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	f.providers[name] = provider
	f.configs[name] = config

	return nil
}

// Get 获取指定的提供商适配器
func (f *AdapterFactory) Get(name string) AIProvider {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.providers[name]
}

// GetConfig 获取指定提供商的配置
func (f *AdapterFactory) GetConfig(name string) *ProviderConfig {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.configs[name]
}

// List 获取所有已注册的提供商名称
func (f *AdapterFactory) List() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}

// Unregister 注销一个提供商适配器
func (f *AdapterFactory) Unregister(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.providers, name)
	delete(f.configs, name)
}

// Exists 检查提供商是否已注册
func (f *AdapterFactory) Exists(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.providers[name]
	return ok
}

// GetAllModels 获取所有提供商的所有模型
func (f *AdapterFactory) GetAllModels() map[string][]Model {
	f.mu.RLock()
	defer f.mu.RUnlock()

	allModels := make(map[string][]Model)
	for name, config := range f.configs {
		models := make([]Model, 0)
		for _, model := range config.Models {
			models = append(models, model)
		}
		allModels[name] = models
	}
	return allModels
}

// GetModelByID 根据模型 ID 获取模型信息
func (f *AdapterFactory) GetModelByID(modelID string) *Model {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, config := range f.configs {
		if model, ok := config.Models[modelID]; ok {
			return &model
		}
	}
	return nil
}

// FindProviderByModel 根据模型 ID 查找提供商
func (f *AdapterFactory) FindProviderByModel(modelID string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for name, config := range f.configs {
		if _, ok := config.Models[modelID]; ok {
			return name
		}
	}
	return ""
}

// UpdateConfig 更新提供商配置
func (f *AdapterFactory) UpdateConfig(name string, config *ProviderConfig) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.providers[name]; !ok {
		return fmt.Errorf("provider %s not registered", name)
	}

	f.configs[name] = config
	return nil
}

