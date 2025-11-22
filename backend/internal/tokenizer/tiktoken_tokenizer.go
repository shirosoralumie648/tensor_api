package tokenizer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

// TiktokenTokenizer 基于tiktoken的Token计数器
type TiktokenTokenizer struct {
	encoders map[string]*tiktoken.Tiktoken
	mu       sync.RWMutex
}

// NewTiktokenTokenizer 创建tiktoken计数器
func NewTiktokenTokenizer() (*TiktokenTokenizer, error) {
	return &TiktokenTokenizer{
		encoders: make(map[string]*tiktoken.Tiktoken),
	}, nil
}

// CountTokens 计算Token数量
func (t *TiktokenTokenizer) CountTokens(ctx context.Context, req *TokenCountRequest) (*TokenCount, error) {
	// 1. 计算消息Token
	messageTokens, err := t.CountMessages(ctx, req.Messages, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to count message tokens: %w", err)
	}

	// 2. 计算工具Token
	toolTokens := 0
	if len(req.Tools) > 0 {
		toolTokens, err = t.countToolTokens(ctx, req.Tools, req.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to count tool tokens: %w", err)
		}
	}

	// 3. 计算图片Token
	imageTokens := 0
	for _, img := range req.ImageDetails {
		tokens, err := t.EstimateImageTokens(ctx, img)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate image tokens: %w", err)
		}
		imageTokens += tokens
	}

	// 4. 计算音频Token
	audioTokens := 0
	for _, audio := range req.AudioDetails {
		tokens, err := t.EstimateAudioTokens(ctx, audio)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate audio tokens: %w", err)
		}
		audioTokens += tokens
	}

	promptTokens := messageTokens + toolTokens + imageTokens + audioTokens

	return &TokenCount{
		PromptTokens:     promptTokens,
		CompletionTokens: 0, // completion tokens需要在响应后计算
		TotalTokens:      promptTokens,
	}, nil
}

// CountText 计算纯文本Token数量
func (t *TiktokenTokenizer) CountText(ctx context.Context, text string, model string) (int, error) {
	encoder, err := t.getEncoder(model)
	if err != nil {
		return 0, err
	}

	tokens := encoder.Encode(text, nil, nil)
	return len(tokens), nil
}

// CountMessages 计算消息列表Token数量
func (t *TiktokenTokenizer) CountMessages(ctx context.Context, messages []Message, model string) (int, error) {
	encoder, err := t.getEncoder(model)
	if err != nil {
		return 0, err
	}

	tokensPerMessage := 3 // 每条消息的固定开销
	tokensPerName := 1    // 如果有name字段的额外开销

	// 针对不同模型调整
	if strings.HasPrefix(model, "gpt-3.5-turbo") {
		tokensPerMessage = 4
		tokensPerName = -1 // gpt-3.5-turbo中name字段减少1个token
	}

	numTokens := 0

	for _, message := range messages {
		numTokens += tokensPerMessage

		// 计算role
		roleTokens := encoder.Encode(message.Role, nil, nil)
		numTokens += len(roleTokens)

		// 计算content
		contentStr := ""
		switch content := message.Content.(type) {
		case string:
			contentStr = content
		case []interface{}:
			// 多模态内容
			for _, part := range content {
				if partMap, ok := part.(map[string]interface{}); ok {
					if partType, ok := partMap["type"].(string); ok {
						if partType == "text" {
							if text, ok := partMap["text"].(string); ok {
								contentStr += text
							}
						}
						// 图片和音频的token已在外部计算
					}
				}
			}
		}

		if contentStr != "" {
			contentTokens := encoder.Encode(contentStr, nil, nil)
			numTokens += len(contentTokens)
		}

		// 计算name
		if message.Name != "" {
			nameTokens := encoder.Encode(message.Name, nil, nil)
			numTokens += len(nameTokens) + tokensPerName
		}
	}

	// 每次对话的固定开销
	numTokens += 3

	return numTokens, nil
}

// EstimateImageTokens 估算图片Token数量
func (t *TiktokenTokenizer) EstimateImageTokens(ctx context.Context, detail ImageDetail) (int, error) {
	// 根据OpenAI的定价规则计算图片token
	switch detail.Detail {
	case "low":
		// low模式固定85 tokens
		return 85, nil
	case "high", "auto":
		// high模式需要根据图片尺寸计算
		width := detail.Width
		height := detail.Height

		// 1. 缩放到2048以内
		if width > 2048 || height > 2048 {
			scale := 2048.0 / math.Max(float64(width), float64(height))
			width = int(float64(width) * scale)
			height = int(float64(height) * scale)
		}

		// 2. 缩放短边到768
		if width < height {
			scale := 768.0 / float64(width)
			width = int(float64(width) * scale)
			height = int(float64(height) * scale)
		} else {
			scale := 768.0 / float64(height)
			width = int(float64(width) * scale)
			height = int(float64(height) * scale)
		}

		// 3. 计算需要多少个512x512的tile
		tilesWidth := int(math.Ceil(float64(width) / 512.0))
		tilesHeight := int(math.Ceil(float64(height) / 512.0))
		totalTiles := tilesWidth * tilesHeight

		// 每个tile 170 tokens，再加上基础的85 tokens
		return totalTiles*170 + 85, nil
	default:
		// 默认使用low模式
		return 85, nil
	}
}

// EstimateAudioTokens 估算音频Token数量
func (t *TiktokenTokenizer) EstimateAudioTokens(ctx context.Context, detail AudioDetail) (int, error) {
	// 音频token估算（基于Whisper的token使用）
	// 一般来说，1分钟音频约等于150-200 tokens
	tokensPerSecond := 3 // 平均每秒3个token
	return detail.Duration * tokensPerSecond, nil
}

// SupportedModels 返回支持的模型列表
func (t *TiktokenTokenizer) SupportedModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"text-embedding-ada-002",
		"text-embedding-3-small",
		"text-embedding-3-large",
	}
}

// countToolTokens 计算工具定义的Token数量
func (t *TiktokenTokenizer) countToolTokens(ctx context.Context, tools []Tool, model string) (int, error) {
	// 将tools序列化为JSON字符串
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal tools: %w", err)
	}

	// 计算JSON字符串的token
	tokens, err := t.CountText(ctx, string(toolsJSON), model)
	if err != nil {
		return 0, err
	}

	// 添加工具调用的固定开销
	return tokens + len(tools)*10, nil
}

// getEncoder 获取或创建编码器
func (t *TiktokenTokenizer) getEncoder(model string) (*tiktoken.Tiktoken, error) {
	t.mu.RLock()
	encoder, exists := t.encoders[model]
	t.mu.RUnlock()

	if exists {
		return encoder, nil
	}

	// 创建新编码器
	t.mu.Lock()
	defer t.mu.Unlock()

	// 双重检查
	if encoder, exists := t.encoders[model]; exists {
		return encoder, nil
	}

	// 根据模型名称选择编码器
	var enc *tiktoken.Tiktoken
	var err error

	if strings.HasPrefix(model, "gpt-4") || strings.HasPrefix(model, "gpt-3.5") {
		enc, err = tiktoken.EncodingForModel(model)
	} else {
		// 默认使用cl100k_base编码（适用于大多数新模型）
		enc, err = tiktoken.GetEncoding("cl100k_base")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get encoding for model %s: %w", model, err)
	}

	t.encoders[model] = enc
	return enc, nil
}

// Close 关闭并释放资源
func (t *TiktokenTokenizer) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// tiktoken的encoder不需要显式关闭
	t.encoders = make(map[string]*tiktoken.Tiktoken)
	return nil
}
