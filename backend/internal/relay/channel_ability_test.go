package relay

import (
	"testing"
	"time"
)

func TestChannelAbilityManagerRegister(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4", "gpt-3.5"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureStreaming: {
				Supported:   true,
				Description: "Streaming support",
			},
		},
		Description: "Initial version",
	}

	if err := manager.RegisterAbility("ch-1", ability); err != nil {
		t.Errorf("RegisterAbility failed: %v", err)
	}

	// 验证已注册
	retrieved, err := manager.GetAbility("ch-1", "v1")
	if err != nil {
		t.Errorf("GetAbility failed: %v", err)
	}

	if retrieved.Version != "v1" {
		t.Errorf("Expected version v1, got %s", retrieved.Version)
	}
}

func TestChannelAbilityManagerGetLatest(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability1 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	ability2 := &ChannelAbilityVersion{
		Version:         "v2",
		ReleasedAt:      time.Now().Add(1 * time.Hour),
		SupportedModels: []string{"gpt-4", "gpt-3.5"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	manager.RegisterAbility("ch-1", ability1)
	manager.RegisterAbility("ch-1", ability2)

	latest, err := manager.GetLatestAbility("ch-1")
	if err != nil {
		t.Errorf("GetLatestAbility failed: %v", err)
	}

	if latest.Version != "v2" {
		t.Errorf("Expected v2, got %s", latest.Version)
	}
}

func TestChannelAbilitySupportsModel(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4", "gpt-3.5"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	manager.RegisterAbility("ch-1", ability)

	// 支持的模型
	if supported, _ := manager.SupportsModel("ch-1", "gpt-4", "v1"); !supported {
		t.Errorf("Expected gpt-4 to be supported")
	}

	// 不支持的模型
	if supported, _ := manager.SupportsModel("ch-1", "claude", "v1"); supported {
		t.Errorf("Expected claude to not be supported")
	}
}

func TestChannelAbilitySupportsFeature(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureStreaming: {
				Supported:   true,
				Description: "Streaming support",
			},
			FeatureFunctionCalling: {
				Supported:   true,
				Description: "Function calling support",
			},
		},
	}

	manager.RegisterAbility("ch-1", ability)

	// 支持的功能
	if supported, _ := manager.SupportsFeature("ch-1", "v1", FeatureStreaming); !supported {
		t.Errorf("Expected Streaming to be supported")
	}

	// 不支持的功能
	if supported, _ := manager.SupportsFeature("ch-1", "v1", FeatureVision); supported {
		t.Errorf("Expected Vision to not be supported")
	}
}

func TestChannelAbilityFilterByModel(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability1 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	ability2 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"claude-3"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	manager.RegisterAbility("ch-1", ability1)
	manager.RegisterAbility("ch-2", ability2)

	// 过滤支持 gpt-4 的渠道
	channels, err := manager.FilterChannelsByModel("gpt-4")
	if err != nil {
		t.Errorf("FilterChannelsByModel failed: %v", err)
	}

	if len(channels) != 1 || channels[0] != "ch-1" {
		t.Errorf("Expected [ch-1], got %v", channels)
	}
}

func TestChannelAbilityFilterByFeature(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability1 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureStreaming: {Supported: true},
		},
	}

	ability2 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"claude-3"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureFunctionCalling: {Supported: true},
		},
	}

	manager.RegisterAbility("ch-1", ability1)
	manager.RegisterAbility("ch-2", ability2)

	// 过滤支持 Streaming 的渠道
	channels, err := manager.FilterChannelsByFeature(FeatureStreaming)
	if err != nil {
		t.Errorf("FilterChannelsByFeature failed: %v", err)
	}

	if len(channels) != 1 || channels[0] != "ch-1" {
		t.Errorf("Expected [ch-1], got %v", channels)
	}
}

func TestChannelAbilityDeprecate(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	manager.RegisterAbility("ch-1", ability)

	// 弃用版本
	if err := manager.DeprecateVersion("ch-1", "v1", "Use v2 instead"); err != nil {
		t.Errorf("DeprecateVersion failed: %v", err)
	}

	// 验证弃用标记
	retrieved, _ := manager.GetAbility("ch-1", "v1")
	if !retrieved.Deprecated {
		t.Errorf("Expected version to be deprecated")
	}

	if retrieved.DeprecationMessage != "Use v2 instead" {
		t.Errorf("Expected deprecation message")
	}
}

func TestChannelAbilityComparison(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability1 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureStreaming: {Supported: true},
		},
	}

	ability2 := &ChannelAbilityVersion{
		Version:         "v2",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4", "gpt-3.5"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureStreaming:       {Supported: true},
			FeatureFunctionCalling: {Supported: true},
		},
	}

	manager.RegisterAbility("ch-1", ability1)
	manager.RegisterAbility("ch-1", ability2)

	comparison, err := manager.GetAbilityComparison("ch-1", "v1", "v2")
	if err != nil {
		t.Errorf("GetAbilityComparison failed: %v", err)
	}

	// 验证新增的模型
	newModels := comparison["new_models"].([]string)
	if len(newModels) != 1 || newModels[0] != "gpt-3.5" {
		t.Errorf("Expected new model gpt-3.5")
	}

	// 验证新增的功能
	newFeatures := comparison["new_features"].([]string)
	if len(newFeatures) != 1 {
		t.Errorf("Expected new feature")
	}
}

func TestChannelAbilityValidator(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features: map[ChannelAbilityFeature]FeatureConfig{
			FeatureStreaming: {Supported: true},
		},
	}

	manager.RegisterAbility("ch-1", ability)

	validator := NewChannelAbilityValidator(manager)

	// 有效的请求
	err := validator.ValidateRequest("ch-1", "gpt-4", map[ChannelAbilityFeature]bool{
		FeatureStreaming: true,
	}, "v1")

	if err != nil {
		t.Errorf("ValidateRequest failed for valid request: %v", err)
	}

	// 无效的模型
	err = validator.ValidateRequest("ch-1", "claude", map[ChannelAbilityFeature]bool{}, "v1")
	if err == nil {
		t.Errorf("Expected error for invalid model")
	}

	// 缺失的功能
	err = validator.ValidateRequest("ch-1", "gpt-4", map[ChannelAbilityFeature]bool{
		FeatureFunctionCalling: true,
	}, "v1")
	if err == nil {
		t.Errorf("Expected error for missing feature")
	}
}

func TestChannelAbilityListVersions(t *testing.T) {
	manager := NewChannelAbilityManager()

	ability1 := &ChannelAbilityVersion{
		Version:         "v1",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	ability2 := &ChannelAbilityVersion{
		Version:         "v2",
		ReleasedAt:      time.Now(),
		SupportedModels: []string{"gpt-4"},
		Features:        make(map[ChannelAbilityFeature]FeatureConfig),
	}

	manager.RegisterAbility("ch-1", ability1)
	manager.RegisterAbility("ch-1", ability2)

	versions, err := manager.ListVersions("ch-1")
	if err != nil {
		t.Errorf("ListVersions failed: %v", err)
	}

	if len(versions) != 2 || versions[0] != "v1" || versions[1] != "v2" {
		t.Errorf("Expected [v1, v2], got %v", versions)
	}
}

func TestChannelAbilityConfig(t *testing.T) {
	config := &ChannelAbilityConfig{
		ChannelID: "ch-1",
		Version:   "v1",
		Models:    []string{"gpt-4", "gpt-3.5"},
		Features: []ChannelAbilityFeature{
			FeatureStreaming,
			FeatureFunctionCalling,
		},
	}

	ability := config.BuildAbilityVersion("Test version")

	if ability.Version != "v1" {
		t.Errorf("Expected version v1")
	}

	if len(ability.SupportedModels) != 2 {
		t.Errorf("Expected 2 models")
	}

	if len(ability.Features) != 2 {
		t.Errorf("Expected 2 features")
	}
}

