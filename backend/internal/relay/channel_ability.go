package relay

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ChannelAbilityFeature 渠道功能特性
type ChannelAbilityFeature string

const (
	// 流式输出
	FeatureStreaming ChannelAbilityFeature = "streaming"
	// 函数调用
	FeatureFunctionCalling ChannelAbilityFeature = "function_calling"
	// 视觉/图像识别
	FeatureVision ChannelAbilityFeature = "vision"
	// 文件上传
	FeatureFileUpload ChannelAbilityFeature = "file_upload"
	// JSON 模式
	FeatureJSONMode ChannelAbilityFeature = "json_mode"
	// 系统提示词
	FeatureSystemPrompt ChannelAbilityFeature = "system_prompt"
	// 温度参数
	FeatureTemperature ChannelAbilityFeature = "temperature"
	// 最大 token
	FeatureMaxTokens ChannelAbilityFeature = "max_tokens"
	// 上下文窗口
	FeatureContextWindow ChannelAbilityFeature = "context_window"
	// 并行函数调用
	FeatureParallelFunctions ChannelAbilityFeature = "parallel_functions"
)

// ChannelAbilityVersion 渠道能力版本
type ChannelAbilityVersion struct {
	// 版本号
	Version string

	// 发布时间
	ReleasedAt time.Time

	// 支持的模型
	SupportedModels []string

	// 支持的功能
	Features map[ChannelAbilityFeature]FeatureConfig

	// 描述
	Description string

	// 是否已弃用
	Deprecated bool

	// 弃用信息
	DeprecationMessage string
}

// FeatureConfig 功能配置
type FeatureConfig struct {
	// 是否支持
	Supported bool

	// 功能描述
	Description string

	// 限制条件
	Limits map[string]interface{}

	// 附加信息
	Extra map[string]interface{}
}

// ChannelAbilityManager 渠道能力管理器
type ChannelAbilityManager struct {
	// 能力版本映射 (channelID -> version -> ChannelAbilityVersion)
	abilities map[string]map[string]*ChannelAbilityVersion
	abilitiesMu sync.RWMutex

	// 默认版本映射 (channelID -> defaultVersion)
	defaultVersions map[string]string
	versionsMu      sync.RWMutex

	// 版本历史 (channelID -> []version)
	versionHistory map[string][]string
	historyMu      sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewChannelAbilityManager 创建能力管理器
func NewChannelAbilityManager() *ChannelAbilityManager {
	return &ChannelAbilityManager{
		abilities:       make(map[string]map[string]*ChannelAbilityVersion),
		defaultVersions: make(map[string]string),
		versionHistory:  make(map[string][]string),
		logFunc:         defaultLogFunc,
	}
}

// RegisterAbility 注册渠道能力
func (cam *ChannelAbilityManager) RegisterAbility(channelID string, ability *ChannelAbilityVersion) error {
	if channelID == "" || ability == nil {
		return fmt.Errorf("channel ID and ability cannot be empty")
	}

	cam.abilitiesMu.Lock()
	defer cam.abilitiesMu.Unlock()

	if _, ok := cam.abilities[channelID]; !ok {
		cam.abilities[channelID] = make(map[string]*ChannelAbilityVersion)
	}

	cam.abilities[channelID][ability.Version] = ability

	// 记录版本历史
	cam.historyMu.Lock()
	defer cam.historyMu.Unlock()

	if cam.versionHistory[channelID] == nil {
		cam.versionHistory[channelID] = make([]string, 0)
	}
	cam.versionHistory[channelID] = append(cam.versionHistory[channelID], ability.Version)

	// 设置为默认版本（如果是第一个版本）
	cam.versionsMu.Lock()
	defer cam.versionsMu.Unlock()

	if _, ok := cam.defaultVersions[channelID]; !ok {
		cam.defaultVersions[channelID] = ability.Version
	}

	cam.logFunc("info", fmt.Sprintf("Registered ability for %s (version %s)", channelID, ability.Version))

	return nil
}

// GetAbility 获取渠道能力
func (cam *ChannelAbilityManager) GetAbility(channelID, version string) (*ChannelAbilityVersion, error) {
	cam.abilitiesMu.RLock()
	defer cam.abilitiesMu.RUnlock()

	versions, ok := cam.abilities[channelID]
	if !ok {
		return nil, fmt.Errorf("channel %s not found", channelID)
	}

	// 如果版本为空，使用默认版本
	if version == "" {
		cam.versionsMu.RLock()
		version = cam.defaultVersions[channelID]
		cam.versionsMu.RUnlock()

		if version == "" {
			return nil, fmt.Errorf("no default version for channel %s", channelID)
		}
	}

	ability, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("version %s not found for channel %s", version, channelID)
	}

	return ability, nil
}

// GetLatestAbility 获取最新版本的能力
func (cam *ChannelAbilityManager) GetLatestAbility(channelID string) (*ChannelAbilityVersion, error) {
	cam.historyMu.RLock()
	history, ok := cam.versionHistory[channelID]
	cam.historyMu.RUnlock()

	if !ok || len(history) == 0 {
		return nil, fmt.Errorf("no versions found for channel %s", channelID)
	}

	// 最后一个版本是最新的
	latestVersion := history[len(history)-1]

	return cam.GetAbility(channelID, latestVersion)
}

// SetDefaultVersion 设置默认版本
func (cam *ChannelAbilityManager) SetDefaultVersion(channelID, version string) error {
	// 验证版本存在
	_, err := cam.GetAbility(channelID, version)
	if err != nil {
		return err
	}

	cam.versionsMu.Lock()
	defer cam.versionsMu.Unlock()

	cam.defaultVersions[channelID] = version
	cam.logFunc("info", fmt.Sprintf("Set default version for %s to %s", channelID, version))

	return nil
}

// ListVersions 列出所有版本
func (cam *ChannelAbilityManager) ListVersions(channelID string) ([]string, error) {
	cam.historyMu.RLock()
	defer cam.historyMu.RUnlock()

	history, ok := cam.versionHistory[channelID]
	if !ok {
		return nil, fmt.Errorf("channel %s not found", channelID)
	}

	// 返回副本
	result := make([]string, len(history))
	copy(result, history)

	return result, nil
}

// SupportsModel 检查是否支持某个模型
func (cam *ChannelAbilityManager) SupportsModel(channelID, model, version string) (bool, error) {
	ability, err := cam.GetAbility(channelID, version)
	if err != nil {
		return false, err
	}

	for _, m := range ability.SupportedModels {
		if m == model {
			return true, nil
		}
	}

	return false, nil
}

// SupportsFeature 检查是否支持某个功能
func (cam *ChannelAbilityManager) SupportsFeature(channelID, version string, feature ChannelAbilityFeature) (bool, error) {
	ability, err := cam.GetAbility(channelID, version)
	if err != nil {
		return false, err
	}

	config, ok := ability.Features[feature]
	if !ok {
		return false, nil
	}

	return config.Supported, nil
}

// GetFeatureConfig 获取功能配置
func (cam *ChannelAbilityManager) GetFeatureConfig(channelID, version string, feature ChannelAbilityFeature) (*FeatureConfig, error) {
	ability, err := cam.GetAbility(channelID, version)
	if err != nil {
		return nil, err
	}

	config, ok := ability.Features[feature]
	if !ok {
		return nil, fmt.Errorf("feature %s not found", feature)
	}

	return &config, nil
}

// FilterChannelsByModel 按模型过滤渠道
func (cam *ChannelAbilityManager) FilterChannelsByModel(model string) ([]string, error) {
	cam.abilitiesMu.RLock()
	defer cam.abilitiesMu.RUnlock()

	result := make([]string, 0)

	for channelID, versions := range cam.abilities {
		// 检查最新版本是否支持该模型
		cam.historyMu.RLock()
		history := cam.versionHistory[channelID]
		cam.historyMu.RUnlock()

		if len(history) == 0 {
			continue
		}

		latestVersion := history[len(history)-1]
		ability, ok := versions[latestVersion]
		if !ok {
			continue
		}

		for _, m := range ability.SupportedModels {
			if m == model {
				result = append(result, channelID)
				break
			}
		}
	}

	return result, nil
}

// FilterChannelsByFeature 按功能过滤渠道
func (cam *ChannelAbilityManager) FilterChannelsByFeature(feature ChannelAbilityFeature) ([]string, error) {
	cam.abilitiesMu.RLock()
	defer cam.abilitiesMu.RUnlock()

	result := make([]string, 0)

	for channelID, versions := range cam.abilities {
		// 检查最新版本是否支持该功能
		cam.historyMu.RLock()
		history := cam.versionHistory[channelID]
		cam.historyMu.RUnlock()

		if len(history) == 0 {
			continue
		}

		latestVersion := history[len(history)-1]
		ability, ok := versions[latestVersion]
		if !ok {
			continue
		}

		if config, ok := ability.Features[feature]; ok && config.Supported {
			result = append(result, channelID)
		}
	}

	return result, nil
}

// DeprecateVersion 弃用某个版本
func (cam *ChannelAbilityManager) DeprecateVersion(channelID, version, message string) error {
	ability, err := cam.GetAbility(channelID, version)
	if err != nil {
		return err
	}

	ability.Deprecated = true
	ability.DeprecationMessage = message

	cam.logFunc("warn", fmt.Sprintf("Deprecated version %s for channel %s: %s", version, channelID, message))

	return nil
}

// GetAbilityComparison 获取两个版本的能力对比
func (cam *ChannelAbilityManager) GetAbilityComparison(channelID, version1, version2 string) (map[string]interface{}, error) {
	ability1, err := cam.GetAbility(channelID, version1)
	if err != nil {
		return nil, err
	}

	ability2, err := cam.GetAbility(channelID, version2)
	if err != nil {
		return nil, err
	}

	// 对比支持的模型
	newModels := make([]string, 0)
	removedModels := make([]string, 0)

	modelMap1 := make(map[string]bool)
	for _, m := range ability1.SupportedModels {
		modelMap1[m] = true
	}

	modelMap2 := make(map[string]bool)
	for _, m := range ability2.SupportedModels {
		modelMap2[m] = true
	}

	for m := range modelMap2 {
		if !modelMap1[m] {
			newModels = append(newModels, m)
		}
	}

	for m := range modelMap1 {
		if !modelMap2[m] {
			removedModels = append(removedModels, m)
		}
	}

	// 对比功能
	newFeatures := make([]string, 0)
	removedFeatures := make([]string, 0)

	for feature, config2 := range ability2.Features {
		config1, ok := ability1.Features[feature]
		if !ok {
			if config2.Supported {
				newFeatures = append(newFeatures, string(feature))
			}
		} else if !config1.Supported && config2.Supported {
			newFeatures = append(newFeatures, string(feature))
		}
	}

	for feature, config1 := range ability1.Features {
		config2, ok := ability2.Features[feature]
		if !ok {
			if config1.Supported {
				removedFeatures = append(removedFeatures, string(feature))
			}
		} else if config1.Supported && !config2.Supported {
			removedFeatures = append(removedFeatures, string(feature))
		}
	}

	return map[string]interface{}{
		"new_models":      newModels,
		"removed_models":  removedModels,
		"new_features":    newFeatures,
		"removed_features": removedFeatures,
	}, nil
}

// ExportAbilityJSON 导出能力为 JSON
func (cam *ChannelAbilityManager) ExportAbilityJSON(channelID, version string) (string, error) {
	ability, err := cam.GetAbility(channelID, version)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(ability, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportAbilityJSON 从 JSON 导入能力
func (cam *ChannelAbilityManager) ImportAbilityJSON(channelID, jsonData string) error {
	var ability ChannelAbilityVersion
	if err := json.Unmarshal([]byte(jsonData), &ability); err != nil {
		return err
	}

	return cam.RegisterAbility(channelID, &ability)
}

// ChannelAbilityValidator 能力验证器
type ChannelAbilityValidator struct {
	manager *ChannelAbilityManager
}

// NewChannelAbilityValidator 创建验证器
func NewChannelAbilityValidator(manager *ChannelAbilityManager) *ChannelAbilityValidator {
	return &ChannelAbilityValidator{manager: manager}
}

// ValidateRequest 验证请求是否与渠道能力兼容
func (cav *ChannelAbilityValidator) ValidateRequest(channelID, model string, features map[ChannelAbilityFeature]bool, version string) error {
	// 检查模型支持
	supported, err := cav.manager.SupportsModel(channelID, model, version)
	if err != nil {
		return err
	}

	if !supported {
		return fmt.Errorf("model %s is not supported by channel %s", model, channelID)
	}

	// 检查功能支持
	for feature, required := range features {
		if !required {
			continue
		}

		supported, err := cav.manager.SupportsFeature(channelID, version, feature)
		if err != nil {
			return err
		}

		if !supported {
			return fmt.Errorf("feature %s is not supported by channel %s", feature, channelID)
		}
	}

	return nil
}

// GetCompatibleChannels 获取兼容的渠道列表
func (cav *ChannelAbilityValidator) GetCompatibleChannels(model string, features map[ChannelAbilityFeature]bool) ([]string, error) {
	// 先按模型过滤
	channels, err := cav.manager.FilterChannelsByModel(model)
	if err != nil {
		return nil, err
	}

	// 再按功能过滤
	compatible := make([]string, 0)
	for _, channelID := range channels {
		valid := true

		for feature, required := range features {
			if !required {
				continue
			}

			supported, err := cav.manager.SupportsFeature(channelID, "", feature)
			if err != nil || !supported {
				valid = false
				break
			}
		}

		if valid {
			compatible = append(compatible, channelID)
		}
	}

	return compatible, nil
}

// ChannelAbilityConfig 渠道能力配置助手
type ChannelAbilityConfig struct {
	// 渠道 ID
	ChannelID string

	// 版本
	Version string

	// 支持的模型
	Models []string

	// 支持的功能
	Features []ChannelAbilityFeature
}

// BuildAbilityVersion 构建能力版本
func (cac *ChannelAbilityConfig) BuildAbilityVersion(description string) *ChannelAbilityVersion {
	version := &ChannelAbilityVersion{
		Version:         cac.Version,
		ReleasedAt:      time.Now(),
		SupportedModels: cac.Models,
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
		Description:     description,
	}

	// 构建功能配置
	for _, feature := range cac.Features {
		version.Features[feature] = FeatureConfig{
			Supported:   true,
			Description: fmt.Sprintf("%s support", feature),
			Limits:      make(map[string]interface{}),
			Extra:       make(map[string]interface{}),
		}
	}

	return version
}

