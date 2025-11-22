package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/gorm"
)

// ConfigManager 配置管理器（从数据库加载适配器配置）
type ConfigManager struct {
	db          *gorm.DB
	registry    *AdapterRegistry
	configs     map[string]*model.AdapterConfig
	mu          sync.RWMutex
	initialized bool
}

// NewConfigManager 创建配置管理器
func NewConfigManager(db *gorm.DB, registry *AdapterRegistry) *ConfigManager {
	return &ConfigManager{
		db:       db,
		registry: registry,
		configs:  make(map[string]*model.AdapterConfig),
	}
}

// Initialize 初始化配置管理器（从数据库加载配置）
func (cm *ConfigManager) Initialize(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 从数据库加载所有启用的适配器配置
	var configs []model.AdapterConfig
	if err := cm.db.WithContext(ctx).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&configs).Error; err != nil {
		return fmt.Errorf("failed to load adapter configs: %w", err)
	}

	// 清空现有配置
	cm.configs = make(map[string]*model.AdapterConfig)

	// 加载所有配置
	for i := range configs {
		cfg := &configs[i]
		cm.configs[cfg.Name] = cfg
	}

	cm.initialized = true
	return nil
}

// GetAdapter 获取适配器（动态创建）
func (cm *ConfigManager) GetAdapter(name string, channelConfig *AdapterConfig) (Adapter, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.initialized {
		return nil, fmt.Errorf("config manager not initialized")
	}

	// 使用registry创建适配器
	return cm.registry.Create(name, channelConfig)
}

// ListAdapters 列出所有可用适配器
func (cm *ConfigManager) ListAdapters() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.configs))
	for name := range cm.configs {
		names = append(names, name)
	}

	return names
}

// ReloadConfig 重新加载配置（热更新）
func (cm *ConfigManager) ReloadConfig(ctx context.Context, name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 从数据库加载最新配置
	var cfg model.AdapterConfig
	if err := cm.db.WithContext(ctx).
		Where("name = ? AND enabled = ? AND deleted_at IS NULL", name, true).
		First(&cfg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 配置被删除或禁用，移除配置
			delete(cm.configs, name)
			return nil
		}
		return fmt.Errorf("failed to load adapter config: %w", err)
	}

	// 更新配置
	cm.configs[name] = &cfg
	return nil
}

// ReloadAllConfigs 重新加载所有配置
func (cm *ConfigManager) ReloadAllConfigs(ctx context.Context) error {
	return cm.Initialize(ctx)
}

// GetConfig 获取适配器配置
func (cm *ConfigManager) GetConfig(name string) (*model.AdapterConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cfg, exists := cm.configs[name]
	if !exists {
		return nil, fmt.Errorf("adapter config %s not found", name)
	}

	return cfg, nil
}

// IsInitialized 检查是否已初始化
func (cm *ConfigManager) IsInitialized() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.initialized
}
