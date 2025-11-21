package service

import (
	"context"
	"fmt"

	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/internal/relay"
	"github.com/oblivious/backend/internal/repository"
)

// RelayService 中转服务
type RelayService struct {
	selector              *relay.ChannelSelector
	channelRepo           *repository.ChannelRepository
	modelPriceRepo        *repository.ModelPriceRepository
}

// NewRelayService 创建中转服务
func NewRelayService() *RelayService {
	return &RelayService{
		selector:       relay.NewChannelSelector(),
		channelRepo:    repository.NewChannelRepository(),
		modelPriceRepo: repository.NewModelPriceRepository(),
	}
}

// RelayChatCompletion 中转 Chat Completion 请求
func (s *RelayService) RelayChatCompletion(ctx context.Context, req *relay.ChatCompletionRequest) (*relay.ChatCompletionResponse, error) {
	// 1. 选择渠道
	channel, err := s.selector.SelectChannel(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to select channel: %w", err)
	}

	// 2. 创建适配器（根据渠道类型选择对应的适配器）
	adapter, err := s.createAdapter(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// 3. 调用上游 API
	resp, err := adapter.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call upstream API: %w", err)
	}

	return resp, nil
}

// RelayChatCompletionStream 中转流式 Chat Completion 请求
func (s *RelayService) RelayChatCompletionStream(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error {
	// 1. 选择渠道
	channel, err := s.selector.SelectChannel(ctx, req.Model)
	if err != nil {
		return fmt.Errorf("failed to select channel: %w", err)
	}

	// 2. 创建适配器
	adapter, err := s.createAdapter(channel)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	// 3. 调用上游 API（流式）
	// 这将在 Week 7 完整实现
	return adapter.ChatStream(ctx, req, handler)
}

// createAdapter 根据渠道类型创建对应的适配器
func (s *RelayService) createAdapter(channel *model.Channel) (interface {
	Chat(ctx context.Context, req *relay.ChatCompletionRequest) (*relay.ChatCompletionResponse, error)
	ChatStream(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error
}, error) {
	switch channel.Type {
	case "openai":
		return relay.NewOpenAIAdapter(channel), nil
	case "azure":
		// TODO: 实现 Azure OpenAI 适配器
		return nil, fmt.Errorf("azure adapter not yet implemented")
	case "claude":
		// TODO: 实现 Claude 适配器
		return nil, fmt.Errorf("claude adapter not yet implemented")
	case "gemini":
		// TODO: 实现 Gemini 适配器
		return nil, fmt.Errorf("gemini adapter not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
}

// GetAvailableChannels 获取所有可用渠道
func (s *RelayService) GetAvailableChannels(ctx context.Context) ([]*model.Channel, error) {
	return s.selector.GetAllChannels(ctx)
}

// StreamChatCompletion 流式 Chat Completion 的别名（Week 7 SSE 使用）
func (s *RelayService) StreamChatCompletion(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error {
	return s.RelayChatCompletionStream(ctx, req, handler)
}

// ChatCompletionStream 流式 Chat Completion（另一个别名，为了兼容性）
func (s *RelayService) ChatCompletionStream(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error {
	return s.RelayChatCompletionStream(ctx, req, handler)
}

// GetModelPrice 获取指定渠道的模型价格
func (s *RelayService) GetModelPrice(ctx context.Context, channelID int, modelName string) (*model.ModelPrice, error) {
	return s.modelPriceRepo.FindByChannelAndModel(ctx, channelID, modelName)
}

// GetModelLowestPrice 获取模型的最低价格（在所有渠道中）
func (s *RelayService) GetModelLowestPrice(ctx context.Context, modelName string) (*model.ModelPrice, error) {
	return s.modelPriceRepo.FindByModel(ctx, modelName)
}

