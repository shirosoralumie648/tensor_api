package service

import (
	"context"
	"fmt"

	"github.com/shirosoralumie648/Oblivious/backend/internal/adapter"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/relay"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
)

// RelayService 中转服务
type RelayService struct {
	selector       *relay.ChannelSelector
	channelRepo    *repository.ChannelRepository
	modelPriceRepo *repository.ModelPriceRepository
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

	// 2. 获取适配器
	adaptor, err := adapter.GetAdapterByChannel(channel)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// 3. 转换请求
	// 注意：adapter 包使用的是 adapter.OpenAIRequest，我们需要做类型转换
	adapterReq := s.convertToAdapterRequest(req)
	convertedReq, err := adaptor.ConvertRequest(adapterReq)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// 4. 发送请求
	httpResp, err := adaptor.DoRequest(ctx, convertedReq)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}

	// 5. 解析响应
	adapterResp, err := adaptor.ParseResponse(httpResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 6. 转换响应回 Relay 格式
	return s.convertFromAdapterResponse(adapterResp), nil
}

// RelayChatCompletionStream 中转流式 Chat Completion 请求
func (s *RelayService) RelayChatCompletionStream(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error {
	// 1. 选择渠道
	channel, err := s.selector.SelectChannel(ctx, req.Model)
	if err != nil {
		return fmt.Errorf("failed to select channel: %w", err)
	}

	// 2. 获取适配器
	adaptor, err := adapter.GetAdapterByChannel(channel)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}

	// 3. 转换请求
	req.Stream = true
	adapterReq := s.convertToAdapterRequest(req)
	convertedReq, err := adaptor.ConvertRequest(adapterReq)
	if err != nil {
		return fmt.Errorf("failed to convert request: %w", err)
	}

	// 4. 发送请求
	httpResp, err := adaptor.DoRequest(ctx, convertedReq)
	if err != nil {
		return fmt.Errorf("upstream request failed: %w", err)
	}

	// 5. 解析流式响应
	streamChan, err := adaptor.ParseStreamResponse(httpResp)
	if err != nil {
		return fmt.Errorf("failed to parse stream response: %w", err)
	}

	// 6. 处理流式数据
	for chunk := range streamChan {
		relayChunk := s.convertFromAdapterStreamChunk(chunk)
		if err := handler(relayChunk); err != nil {
			return err
		}
	}

	return nil
}

// 辅助函数：类型转换
func (s *RelayService) convertToAdapterRequest(req *relay.ChatCompletionRequest) *adapter.OpenAIRequest {
	messages := make([]adapter.Message, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = adapter.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	return &adapter.OpenAIRequest{
		Model:            req.Model,
		Messages:         messages,
		Temperature:      float32(req.Temperature),
		MaxTokens:        req.MaxTokens,
		TopP:             float32(req.TopP),
		FrequencyPenalty: float32(req.FrequencyPenalty),
		PresencePenalty:  float32(req.PresencePenalty),
		Stream:           req.Stream,
	}
}

func (s *RelayService) convertFromAdapterResponse(resp *adapter.OpenAIResponse) *relay.ChatCompletionResponse {
	choices := make([]struct {
		Index        int                `json:"index"`
		Message      relay.ChatMessage  `json:"message"`
		Delta        *relay.ChatMessage `json:"delta,omitempty"`
		FinishReason string             `json:"finish_reason"`
	}, len(resp.Choices))

	for i, c := range resp.Choices {
		choices[i] = struct {
			Index        int                `json:"index"`
			Message      relay.ChatMessage  `json:"message"`
			Delta        *relay.ChatMessage `json:"delta,omitempty"`
			FinishReason string             `json:"finish_reason"`
		}{
			Index: c.Index,
			Message: relay.ChatMessage{
				Role:    c.Message.Role,
				Content: fmt.Sprintf("%v", c.Message.Content), // 简单处理 content
			},
			FinishReason: c.FinishReason,
		}
	}

	return &relay.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

func (s *RelayService) convertFromAdapterStreamChunk(chunk *adapter.StreamChunk) *relay.ChatCompletionResponse {
	choices := make([]struct {
		Index        int                `json:"index"`
		Message      relay.ChatMessage  `json:"message"`
		Delta        *relay.ChatMessage `json:"delta,omitempty"`
		FinishReason string             `json:"finish_reason"`
	}, len(chunk.Choices))

	for i, c := range chunk.Choices {
		var delta *relay.ChatMessage
		if c.Delta != nil {
			delta = &relay.ChatMessage{
				Role:    c.Delta.Role,
				Content: fmt.Sprintf("%v", c.Delta.Content),
			}
		}

		choices[i] = struct {
			Index        int                `json:"index"`
			Message      relay.ChatMessage  `json:"message"`
			Delta        *relay.ChatMessage `json:"delta,omitempty"`
			FinishReason string             `json:"finish_reason"`
		}{
			Index:        c.Index,
			Delta:        delta,
			FinishReason: c.FinishReason,
		}
	}

	return &relay.ChatCompletionResponse{
		ID:      chunk.ID,
		Object:  chunk.Object,
		Created: chunk.Created,
		Model:   chunk.Model,
		Choices: choices,
	}
}

// GetAvailableChannels 获取所有可用渠道
func (s *RelayService) GetAvailableChannels(ctx context.Context) ([]*model.Channel, error) {
	return s.selector.GetAllChannels(ctx)
}

// StreamChatCompletion 流式 Chat Completion 的别名
func (s *RelayService) StreamChatCompletion(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error {
	return s.RelayChatCompletionStream(ctx, req, handler)
}

// ChatCompletionStream 流式 Chat Completion（另一个别名）
func (s *RelayService) ChatCompletionStream(ctx context.Context, req *relay.ChatCompletionRequest, handler func(chunk *relay.ChatCompletionResponse) error) error {
	return s.RelayChatCompletionStream(ctx, req, handler)
}

// GetModelPrice 获取指定渠道的模型价格
func (s *RelayService) GetModelPrice(ctx context.Context, channelID int, modelName string) (*model.ModelPrice, error) {
	return s.modelPriceRepo.FindByChannelAndModel(ctx, channelID, modelName)
}

// GetModelLowestPrice 获取模型的最低价格
func (s *RelayService) GetModelLowestPrice(ctx context.Context, modelName string) (*model.ModelPrice, error) {
	return s.modelPriceRepo.FindByModel(ctx, modelName)
}
