package openai

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sashabaranov/go-openai"
	"oblivious/internal/adapter"
)

// OpenAIAdapter OpenAI 提供商适配器
type OpenAIAdapter struct {
	client  *openai.Client
	config  *adapter.ProviderConfig
	timeout time.Duration
}

// NewOpenAIAdapter 创建新的 OpenAI 适配器
func NewOpenAIAdapter(apiKey string) *OpenAIAdapter {
	client := openai.NewClient(apiKey)
	
	return &OpenAIAdapter{
		client:  client,
		timeout: 30 * time.Second,
		config: &adapter.ProviderConfig{
			Name:   "openai",
			APIKey: apiKey,
			Models: getDefaultModels(),
		},
	}
}

// WithTimeout 设置请求超时
func (a *OpenAIAdapter) WithTimeout(duration time.Duration) *OpenAIAdapter {
	a.timeout = duration
	return a
}

// Chat 实现单次请求
func (a *OpenAIAdapter) Chat(ctx context.Context, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("chat request cannot be nil")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// 转换消息格式
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	startTime := time.Now()

	// 调用 OpenAI API
	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		User:        req.User,
	})

	if err != nil {
		return nil, &adapter.AdapterError{
			Provider: "openai",
			Code:     "chat_error",
			Message:  err.Error(),
			Err:      err,
		}
	}

	// 构建响应
	respTime := time.Since(startTime).Milliseconds()
	return &adapter.ChatResponse{
		ID:      resp.ID,
		Model:   resp.Model,
		Content: resp.Choices[0].Message.Content,
		Tokens: adapter.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: string(resp.Choices[0].FinishReason),
		Provider:     "openai",
		ResponseTime: respTime,
	}, nil
}

// ChatStream 实现流式响应
func (a *OpenAIAdapter) ChatStream(ctx context.Context, req *adapter.ChatRequest) (<-chan *adapter.StreamDelta, error) {
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
		messages := make([]openai.ChatCompletionMessage, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
				Name:    msg.Name,
			}
		}

		// 创建流式请求
		stream, err := a.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:       req.Model,
			Messages:    messages,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			TopP:        req.TopP,
		})

		if err != nil {
			deltaCh <- &adapter.StreamDelta{
				Error: err,
				Done:  true,
			}
			return
		}
		defer stream.Close()

		totalTokens := 0

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				deltaCh <- &adapter.StreamDelta{
					Done: true,
				}
				return
			}

			if err != nil {
				deltaCh <- &adapter.StreamDelta{
					Error: err,
					Done:  true,
				}
				return
			}

			// 提取增量内容
			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				finishReason := string(response.Choices[0].FinishReason)

				if delta.Content != "" {
					totalTokens += len(delta.Content) / 4 // 粗略估计
				}

				deltaCh <- &adapter.StreamDelta{
					Content:      delta.Content,
					Index:        response.Choices[0].Index,
					FinishReason: finishReason,
					Done:         finishReason != "",
				}
			}
		}
	}()

	return deltaCh, nil
}

// ListModels 获取模型列表
func (a *OpenAIAdapter) ListModels(ctx context.Context) ([]adapter.Model, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	resp, err := a.client.ListModels(ctx)
	if err != nil {
		return nil, &adapter.AdapterError{
			Provider: "openai",
			Code:     "list_models_error",
			Message:  err.Error(),
			Err:      err,
		}
	}

	models := make([]adapter.Model, 0)
	for _, model := range resp.Models {
		if isOpenAIModel(model.ID) {
			models = append(models, adapter.Model{
				ID:              model.ID,
				Name:            model.ID,
				Provider:        "openai",
				Type:            "text",
				ContextSize:     8192,
				MaxOutputTokens: 4096,
				IsActive:        true,
			})
		}
	}

	return models, nil
}

// HealthCheck 检查连接
func (a *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := a.client.ListModels(ctx)
	if err != nil {
		return &adapter.AdapterError{
			Provider: "openai",
			Code:     "health_check_error",
			Message:  fmt.Sprintf("OpenAI health check failed: %v", err),
			Err:      err,
		}
	}

	return nil
}

// GetName 获取提供商名称
func (a *OpenAIAdapter) GetName() string {
	return "openai"
}

// getDefaultModels 返回默认的 OpenAI 模型列表
func getDefaultModels() map[string]adapter.Model {
	return map[string]adapter.Model{
		"gpt-4": {
			ID:                "gpt-4",
			Name:              "GPT-4",
			Provider:          "openai",
			Type:              "text",
			ContextSize:       8192,
			MaxOutputTokens:   4096,
			CostPer1KPrompt:   0.03,
			CostPer1KCompletion: 0.06,
			IsActive:          true,
			Description:       "Most capable model",
			ReleaseDate:       "2023-03-14",
		},
		"gpt-4-turbo": {
			ID:                "gpt-4-turbo-preview",
			Name:              "GPT-4 Turbo",
			Provider:          "openai",
			Type:              "text",
			ContextSize:       128000,
			MaxOutputTokens:   4096,
			CostPer1KPrompt:   0.01,
			CostPer1KCompletion: 0.03,
			IsActive:          true,
			Description:       "Most capable model with 128K context",
			ReleaseDate:       "2023-11-06",
		},
		"gpt-3.5-turbo": {
			ID:                "gpt-3.5-turbo",
			Name:              "GPT-3.5 Turbo",
			Provider:          "openai",
			Type:              "text",
			ContextSize:       4096,
			MaxOutputTokens:   4096,
			CostPer1KPrompt:   0.0015,
			CostPer1KCompletion: 0.002,
			IsActive:          true,
			Description:       "Fast and efficient model",
			ReleaseDate:       "2023-03-01",
		},
	}
}

// isOpenAIModel 检查是否是有效的 OpenAI 模型
func isOpenAIModel(modelID string) bool {
	validModels := map[string]bool{
		"gpt-4":                  true,
		"gpt-4-turbo-preview":    true,
		"gpt-3.5-turbo":          true,
		"gpt-4-vision-preview":   true,
	}
	return validModels[modelID]
}

