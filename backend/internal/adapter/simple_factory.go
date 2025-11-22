package adapter

import (
	"fmt"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// GetAdapterByChannel 根据渠道获取适配器
func GetAdapterByChannel(channel *model.Channel) (Adapter, error) {
	providerType := ParseProviderType(channel.Type)

	// 构建基本配置
	config := &AdapterConfig{
		Type:    channel.Type,
		BaseURL: channel.BaseURL,
		APIKey:  channel.APIKey,
		Timeout: 30 * 1000000000, // 30s
	}

	return CreateAdapterFactory(providerType, config)
}

// CreateAdapterFactory 创建适配器实例
func CreateAdapterFactory(providerType ProviderType, config *AdapterConfig) (Adapter, error) {
	switch providerType {
	case ProviderOpenAI:
		return NewOpenAIAdapter(config), nil
	case ProviderAnthropic:
		return NewClaudeAdapter(config), nil
	case ProviderGoogle:
		return NewGeminiAdapter(config), nil
	case ProviderBaidu:
		return NewBaiduAdapter(config), nil
	case ProviderQwen:
		return NewQwenAdapter(config), nil
	// Azure 暂时复用 OpenAI
	case ProviderAzure:
		return NewOpenAIAdapter(config), nil

	// 对于尚未实现的适配器，暂时返回错误
	default:
		return nil, fmt.Errorf("adapter not implemented for provider type: %s", providerType)
	}
}
