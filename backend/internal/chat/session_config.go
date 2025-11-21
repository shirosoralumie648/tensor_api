package chat

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SessionConfig 会话配置
type SessionConfig struct {
	// 配置 ID
	ID string `json:"id"`

	// 会话 ID
	SessionID string `json:"session_id"`

	// 配置内容（JSONB）
	Config map[string]interface{} `json:"config"`

	// 版本号
	Version int64 `json:"version"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionConfigManager 会话配置管理器
type SessionConfigManager struct {
	// 配置存储
	configs map[string]*SessionConfig
	configsMu sync.RWMutex

	// 版本历史
	versions map[string][]*SessionConfig
	versionsMu sync.RWMutex

	// 统计信息
	totalConfigs int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewSessionConfigManager 创建会话配置管理器
func NewSessionConfigManager() *SessionConfigManager {
	return &SessionConfigManager{
		configs:  make(map[string]*SessionConfig),
		versions: make(map[string][]*SessionConfig),
		logFunc:  defaultLogFunc,
	}
}

// CreateConfig 创建配置
func (scm *SessionConfigManager) CreateConfig(sessionID string, config map[string]interface{}) (*SessionConfig, error) {
	scm.configsMu.Lock()
	defer scm.configsMu.Unlock()

	configID := fmt.Sprintf("config-%s", sessionID)

	cfg := &SessionConfig{
		ID:        configID,
		SessionID: sessionID,
		Config:    config,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	scm.configs[sessionID] = cfg

	// 记录版本历史
	scm.versionsMu.Lock()
	scm.versions[sessionID] = append(scm.versions[sessionID], cfg)
	scm.versionsMu.Unlock()

	scm.logFunc("info", fmt.Sprintf("Created config for session %s", sessionID))

	return cfg, nil
}

// GetConfig 获取配置
func (scm *SessionConfigManager) GetConfig(sessionID string) (*SessionConfig, error) {
	scm.configsMu.RLock()
	defer scm.configsMu.RUnlock()

	cfg, exists := scm.configs[sessionID]
	if !exists {
		return nil, fmt.Errorf("config for session %s not found", sessionID)
	}

	return cfg, nil
}

// UpdateConfig 更新配置
func (scm *SessionConfigManager) UpdateConfig(sessionID string, config map[string]interface{}) (*SessionConfig, error) {
	scm.configsMu.Lock()
	defer scm.configsMu.Unlock()

	cfg, exists := scm.configs[sessionID]
	if !exists {
		return nil, fmt.Errorf("config for session %s not found", sessionID)
	}

	cfg.Config = config
	cfg.Version++
	cfg.UpdatedAt = time.Now()

	// 记录版本历史
	scm.versionsMu.Lock()
	scm.versions[sessionID] = append(scm.versions[sessionID], cfg)
	scm.versionsMu.Unlock()

	scm.logFunc("info", fmt.Sprintf("Updated config for session %s (version %d)", sessionID, cfg.Version))

	return cfg, nil
}

// MergeConfig 合并配置
func (scm *SessionConfigManager) MergeConfig(sessionID string, updates map[string]interface{}) (*SessionConfig, error) {
	scm.configsMu.Lock()
	defer scm.configsMu.Unlock()

	cfg, exists := scm.configs[sessionID]
	if !exists {
		return nil, fmt.Errorf("config for session %s not found", sessionID)
	}

	// 合并配置
	for key, value := range updates {
		cfg.Config[key] = value
	}

	cfg.Version++
	cfg.UpdatedAt = time.Now()

	// 记录版本历史
	scm.versionsMu.Lock()
	scm.versions[sessionID] = append(scm.versions[sessionID], cfg)
	scm.versionsMu.Unlock()

	return cfg, nil
}

// DeleteConfig 删除配置
func (scm *SessionConfigManager) DeleteConfig(sessionID string) error {
	scm.configsMu.Lock()
	defer scm.configsMu.Unlock()

	if _, exists := scm.configs[sessionID]; !exists {
		return fmt.Errorf("config for session %s not found", sessionID)
	}

	delete(scm.configs, sessionID)

	return nil
}

// GetConfigVersion 获取配置版本
func (scm *SessionConfigManager) GetConfigVersion(sessionID string, version int64) (*SessionConfig, error) {
	scm.versionsMu.RLock()
	defer scm.versionsMu.RUnlock()

	versions, exists := scm.versions[sessionID]
	if !exists {
		return nil, fmt.Errorf("no versions for session %s", sessionID)
	}

	for _, cfg := range versions {
		if cfg.Version == version {
			return cfg, nil
		}
	}

	return nil, fmt.Errorf("version %d not found", version)
}

// GetConfigVersions 获取所有版本
func (scm *SessionConfigManager) GetConfigVersions(sessionID string) []*SessionConfig {
	scm.versionsMu.RLock()
	defer scm.versionsMu.RUnlock()

	versions, exists := scm.versions[sessionID]
	if !exists {
		return make([]*SessionConfig, 0)
	}

	result := make([]*SessionConfig, len(versions))
	copy(result, versions)

	return result
}

// ExportConfig 导出配置为 JSON
func (scm *SessionConfigManager) ExportConfig(sessionID string) (string, error) {
	cfg, err := scm.GetConfig(sessionID)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(cfg.Config, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportConfig 导入配置从 JSON
func (scm *SessionConfigManager) ImportConfig(sessionID string, jsonStr string) (*SessionConfig, error) {
	var config map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	return scm.UpdateConfig(sessionID, config)
}

// GetConfigValue 获取配置值
func (scm *SessionConfigManager) GetConfigValue(sessionID string, key string) (interface{}, bool) {
	cfg, err := scm.GetConfig(sessionID)
	if err != nil {
		return nil, false
	}

	value, exists := cfg.Config[key]
	return value, exists
}

// SetConfigValue 设置配置值
func (scm *SessionConfigManager) SetConfigValue(sessionID string, key string, value interface{}) (*SessionConfig, error) {
	cfg, err := scm.GetConfig(sessionID)
	if err != nil {
		return nil, err
	}

	cfg.Config[key] = value
	cfg.Version++
	cfg.UpdatedAt = time.Now()

	// 记录版本历史
	scm.versionsMu.Lock()
	scm.versions[sessionID] = append(scm.versions[sessionID], cfg)
	scm.versionsMu.Unlock()

	return cfg, nil
}

// GetStatistics 获取统计信息
func (scm *SessionConfigManager) GetStatistics() map[string]interface{} {
	scm.configsMu.RLock()
	defer scm.configsMu.RUnlock()

	scm.versionsMu.RLock()
	totalVersions := 0
	for _, versions := range scm.versions {
		totalVersions += len(versions)
	}
	scm.versionsMu.RUnlock()

	return map[string]interface{}{
		"total_configs": len(scm.configs),
		"total_versions": totalVersions,
	}
}

// ConfigInheritance 配置继承
type ConfigInheritance struct {
	// 父配置 ID
	ParentSessionID string `json:"parent_session_id"`

	// 子配置 ID
	ChildSessionID string `json:"child_session_id"`

	// 覆盖值
	Overrides map[string]interface{} `json:"overrides"`
}

// InheritanceManager 配置继承管理器
type InheritanceManager struct {
	// 继承关系
	inheritance map[string]*ConfigInheritance
	inheritanceMu sync.RWMutex

	// 配置管理器
	configManager *SessionConfigManager

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewInheritanceManager 创建继承管理器
func NewInheritanceManager(configManager *SessionConfigManager) *InheritanceManager {
	return &InheritanceManager{
		inheritance:   make(map[string]*ConfigInheritance),
		configManager: configManager,
		logFunc:       defaultLogFunc,
	}
}

// CreateInheritance 创建继承关系
func (im *InheritanceManager) CreateInheritance(parentSessionID, childSessionID string) error {
	im.inheritanceMu.Lock()
	defer im.inheritanceMu.Unlock()

	inheritance := &ConfigInheritance{
		ParentSessionID: parentSessionID,
		ChildSessionID:  childSessionID,
		Overrides:       make(map[string]interface{}),
	}

	im.inheritance[childSessionID] = inheritance

	im.logFunc("info", fmt.Sprintf("Created inheritance from %s to %s", parentSessionID, childSessionID))

	return nil
}

// GetEffectiveConfig 获取有效配置（包括继承）
func (im *InheritanceManager) GetEffectiveConfig(sessionID string) (map[string]interface{}, error) {
	im.inheritanceMu.RLock()
	inheritance, exists := im.inheritance[sessionID]
	im.inheritanceMu.RUnlock()

	// 如果没有继承关系，直接返回配置
	if !exists {
		cfg, err := im.configManager.GetConfig(sessionID)
		if err != nil {
			return nil, err
		}

		result := make(map[string]interface{})
		for key, value := range cfg.Config {
			result[key] = value
		}

		return result, nil
	}

	// 获取父配置
	parentCfg, err := im.configManager.GetConfig(inheritance.ParentSessionID)
	if err != nil {
		return nil, err
	}

	// 合并配置：先是父配置，然后是覆盖值
	result := make(map[string]interface{})
	for key, value := range parentCfg.Config {
		result[key] = value
	}

	for key, value := range inheritance.Overrides {
		result[key] = value
	}

	return result, nil
}

// SetOverride 设置覆盖值
func (im *InheritanceManager) SetOverride(childSessionID string, key string, value interface{}) error {
	im.inheritanceMu.Lock()
	defer im.inheritanceMu.Unlock()

	inheritance, exists := im.inheritance[childSessionID]
	if !exists {
		return fmt.Errorf("inheritance not found for session %s", childSessionID)
	}

	inheritance.Overrides[key] = value

	return nil
}

