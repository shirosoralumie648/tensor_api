package claude

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/sdk-go"
	"github.com/anthropics/sdk-go/option"
	"oblivious/internal/adapter"
)

// ClaudeAdapter Claude (Anthropic) 提供商适配器
type ClaudeAdapter struct {
	client  *sdk.Client
	config  *adapter.ProviderConfig
	timeout time.Duration
}

// NewClaudeAdapter 创建新的 Claude 适配器
func NewClaudeAdapter(apiKey string) *ClaudeAdapter {
	client := sdk.NewClient(option.WithAPIKey(apiKey))

	return &ClaudeAdapter{
		client:  client,
		timeout: 30 * time.Second,
		config: &adapter.ProviderConfig{
			Name:   "claude",
			APIKey: apiKey,
			Models: getDefaultClaudeModels(),
		},
	}
}

// WithTimeout 设置请求超时
func (a *ClaudeAdapter) WithTimeout(duration time.Duration) *ClaudeAdapter {
	a.timeout = duration
	return a
}

// Chat 实现单次请求
func (a *ClaudeAdapter) Chat(ctx context.Context, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("chat request cannot be nil")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// 转换消息格式
	messages := make([]sdk.MessageParam, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = sdk.MessageParam{
			Role: msg.Role,
			Content: msg.Content,
		}
	}

	startTime := time.Now()

	// 调用 Claude API
	message, err := a.client.Messages.New(ctx, sdk.MessageNewParams{
		Model:       sdk.String(req.Model),
		MaxTokens:   sdk.Int64(int64(req.MaxTokens)),
		Temperature: sdk.Float(float64(req.Temperature)),
		Messages:    messages,
	})

	if err != nil {
		return nil, &adapter.AdapterError{
			Provider: "claude",
			Code:     "chat_error",
			Message:  err.Error(),
			Err:      err,
		}
	}

	// 提取响应内容
	content := ""
	if len(message.Content) > 0 {
		// Claude 返回的是 ContentBlock 数组，通常第一个是文本
		if textBlock, ok := message.Content[0].(sdk.TextBlock); ok {
			content = textBlock.Text
		}
	}

	respTime := time.Since(startTime).Milliseconds()

	// 构建响应
	return &adapter.ChatResponse{
		ID:      message.ID,
		Model:   message.Model,
		Content: content,
		Tokens: adapter.Usage{
			PromptTokens:     message.Usage.InputTokens,
			CompletionTokens: message.Usage.OutputTokens,
			TotalTokens:      message.Usage.InputTokens + message.Usage.OutputTokens,
		},
		FinishReason: string(message.StopReason),
		Provider:     "claude",
		ResponseTime: respTime,
	}, nil
}

// ChatStream 实现流式响应
func (a *ClaudeAdapter) ChatStream(ctx context.Context, req *adapter.ChatRequest) (<-chan *adapter.StreamDelta, error) {
	if req == nil {
		return nil, fmt.Errorf("chat request cannot be nil")
	}

	deltaCh := make(chan *adapter.StreamDelta, 100)

	// 异步处理流式响应
	go func() {
		defer close(deltaCh)

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(ctx, a.timeout)
		defer cancel()

		// 转换消息格式
		messages := make([]sdk.MessageParam, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = sdk.MessageParam{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}

		// 创建流式请求
		stream, err := a.client.Messages.NewWithStreaming(ctx, sdk.MessageNewParams{
			Model:       sdk.String(req.Model),
			MaxTokens:   sdk.Int64(int64(req.MaxTokens)),
			Temperature: sdk.Float(float64(req.Temperature)),
			Messages:    messages,
		})

		if err != nil {
			deltaCh <- &adapter.StreamDelta{
				Error: err,
				Done:  true,
			}
			return
		}
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()

			// 处理不同的事件类型
			switch event.(type) {
			case sdk.ContentBlockStartEvent:
				// 内容块开始
				continue

			case sdk.ContentBlockDeltaEvent:
				// 内容块增量
				if delta, ok := event.(sdk.ContentBlockDeltaEvent); ok {
					if textDelta, ok := delta.Delta.(sdk.TextDeltaEvent); ok {
						deltaCh <- &adapter.StreamDelta{
							Content: textDelta.Text,
							Done:    false,
						}
					}
				}

			case sdk.ContentBlockStopEvent:
				// 内容块结束
				continue

			case sdk.MessageStopEvent:
				// 消息结束
				if msgStop, ok := event.(sdk.MessageStopEvent); ok {
					deltaCh <- &adapter.StreamDelta{
						FinishReason: string(msgStop.Message.StopReason),
						Done:         true,
					}
				}
				return
			}
		}

		if err := stream.Err(); err != nil {
			deltaCh <- &adapter.StreamDelta{
				Error: err,
				Done:  true,
			}
		}
	}()

	return deltaCh, nil
}

// ListModels 获取模型列表
func (a *ClaudeAdapter) ListModels(ctx context.Context) ([]adapter.Model, error) {
	// Claude 没有 list models 端点，返回预定义的模型列表
	models := make([]adapter.Model, 0)
	for _, model := range a.config.Models {
		models = append(models, model)
	}
	return models, nil
}

// HealthCheck 检查连接
func (a *ClaudeAdapter) HealthCheck(ctx context.Context) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 发送简单请求来检查连接
	_, err := a.client.Messages.New(ctx, sdk.MessageNewParams{
		Model:     sdk.String("claude-3-sonnet-20240229"),
		MaxTokens: sdk.Int64(100),
		Messages: []sdk.MessageParam{
			{
				Role:    "user",
				Content: "ping",
			},
		},
	})

	if err != nil {
		return &adapter.AdapterError{
			Provider: "claude",
			Code:     "health_check_error",
			Message:  fmt.Sprintf("Claude health check failed: %v", err),
			Err:      err,
		}
	}

	return nil
}

// GetName 获取提供商名称
func (a *ClaudeAdapter) GetName() string {
	return "claude"
}

// getDefaultClaudeModels 返回默认的 Claude 模型列表
func getDefaultClaudeModels() map[string]adapter.Model {
	return map[string]adapter.Model{
		"claude-3-opus-20240229": {
			ID:                "claude-3-opus-20240229",
			Name:              "Claude 3 Opus",
			Provider:          "anthropic",
			Type:              "text",
			ContextSize:       200000,
			MaxOutputTokens:   4096,
			CostPer1KPrompt:   0.015,
			CostPer1KCompletion: 0.075,
			IsActive:          true,
			Description:       "Most powerful Claude model",
			ReleaseDate:       "2024-02-29",
		},
		"claude-3-sonnet-20240229": {
			ID:                "claude-3-sonnet-20240229",
			Name:              "Claude 3 Sonnet",
			Provider:          "anthropic",
			Type:              "text",
			ContextSize:       200000,
			MaxOutputTokens:   4096,
			CostPer1KPrompt:   0.003,
			CostPer1KCompletion: 0.015,
			IsActive:          true,
			Description:       "Balanced Claude model",
			ReleaseDate:       "2024-02-29",
		},
		"claude-3-haiku-20240307": {
			ID:                "claude-3-haiku-20240307",
			Name:              "Claude 3 Haiku",
			Provider:          "anthropic",
			Type:              "text",
			ContextSize:       200000,
			MaxOutputTokens:   4096,
			CostPer1KPrompt:   0.00080,
			CostPer1KCompletion: 0.0024,
			IsActive:          true,
			Description:       "Fast and compact Claude model",
			ReleaseDate:       "2024-03-07",
		},
	}
}

