package tokenizer

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// GenericTokenizer 通用Token计数器（用于非OpenAI模型）
type GenericTokenizer struct {
	model string
}

// NewGenericTokenizer 创建通用Token计数器
func NewGenericTokenizer(model string) *GenericTokenizer {
	return &GenericTokenizer{
		model: model,
	}
}

// CountTokens 计算Token数量（估算）
func (g *GenericTokenizer) CountTokens(ctx context.Context, req *TokenCountRequest) (*TokenCount, error) {
	// 计算消息Token
	messageTokens, err := g.CountMessages(ctx, req.Messages, req.Model)
	if err != nil {
		return nil, err
	}

	// 计算工具Token
	toolTokens := 0
	if len(req.Tools) > 0 {
		toolTokens = g.estimateToolTokens(req.Tools)
	}

	// 计算图片Token
	imageTokens := 0
	for _, img := range req.ImageDetails {
		tokens, _ := g.EstimateImageTokens(ctx, img)
		imageTokens += tokens
	}

	// 计算音频Token
	audioTokens := 0
	for _, audio := range req.AudioDetails {
		tokens, _ := g.EstimateAudioTokens(ctx, audio)
		audioTokens += tokens
	}

	promptTokens := messageTokens + toolTokens + imageTokens + audioTokens

	return &TokenCount{
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		TotalTokens:      promptTokens,
	}, nil
}

// CountText 计算纯文本Token数量（基于启发式规则）
func (g *GenericTokenizer) CountText(ctx context.Context, text string, model string) (int, error) {
	// 根据模型选择不同的估算策略
	switch {
	case strings.Contains(model, "claude"):
		return g.estimateClaudeTokens(text), nil
	case strings.Contains(model, "gemini"):
		return g.estimateGeminiTokens(text), nil
	case strings.Contains(model, "qwen"), strings.Contains(model, "通义"):
		return g.estimateChineseTokens(text), nil
	case strings.Contains(model, "glm"), strings.Contains(model, "智谱"):
		return g.estimateChineseTokens(text), nil
	default:
		// 默认估算
		return g.estimateDefaultTokens(text), nil
	}
}

// CountMessages 计算消息列表Token数量
func (g *GenericTokenizer) CountMessages(ctx context.Context, messages []Message, model string) (int, error) {
	total := 0

	for _, msg := range messages {
		// Role token
		roleTokens, _ := g.CountText(ctx, msg.Role, model)
		total += roleTokens

		// Content tokens
		switch content := msg.Content.(type) {
		case string:
			contentTokens, _ := g.CountText(ctx, content, model)
			total += contentTokens
		case []interface{}:
			for _, part := range content {
				if partMap, ok := part.(map[string]interface{}); ok {
					if partType, ok := partMap["type"].(string); ok && partType == "text" {
						if text, ok := partMap["text"].(string); ok {
							textTokens, _ := g.CountText(ctx, text, model)
							total += textTokens
						}
					}
				}
			}
		}

		// Name token
		if msg.Name != "" {
			nameTokens, _ := g.CountText(ctx, msg.Name, model)
			total += nameTokens
		}

		// 每条消息的固定开销
		total += 4
	}

	return total, nil
}

// EstimateImageTokens 估算图片Token数量
func (g *GenericTokenizer) EstimateImageTokens(ctx context.Context, detail ImageDetail) (int, error) {
	// 统一使用简化的图片token估算
	switch detail.Detail {
	case "low":
		return 85, nil
	case "high", "auto":
		// 简化计算：基于像素数
		pixels := detail.Width * detail.Height
		// 大约每10000像素1个token
		return pixels/10000 + 85, nil
	default:
		return 85, nil
	}
}

// EstimateAudioTokens 估算音频Token数量
func (g *GenericTokenizer) EstimateAudioTokens(ctx context.Context, detail AudioDetail) (int, error) {
	// 每秒约3个token
	return detail.Duration * 3, nil
}

// SupportedModels 返回支持的模型列表
func (g *GenericTokenizer) SupportedModels() []string {
	return []string{
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-3.5-sonnet",
		"gemini-pro",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-2.0-flash",
		"qwen-max",
		"qwen-plus",
		"qwen-turbo",
		"glm-4",
		"glm-3-turbo",
		"deepseek-chat",
		"moonshot-v1",
	}
}

// estimateClaudeTokens Claude模型的Token估算
func (g *GenericTokenizer) estimateClaudeTokens(text string) int {
	// Claude使用类似GPT的tokenization
	// 平均1个token约4个字符（英文）
	charCount := len(text)

	// 检测中文字符
	chineseCount := 0
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
		}
	}

	// 中文字符通常每个字符1-2个token
	englishCount := charCount - chineseCount
	return (englishCount / 4) + (chineseCount * 2)
}

// estimateGeminiTokens Gemini模型的Token估算
func (g *GenericTokenizer) estimateGeminiTokens(text string) int {
	// Gemini使用SentencePiece
	// 英文：约4字符/token
	// 中文：约1.5-2字符/token
	charCount := utf8.RuneCountInString(text)

	// 简化估算
	if containsChinese(text) {
		return charCount * 6 / 10 // 0.6 tokens per character
	}
	return charCount / 4
}

// estimateChineseTokens 中文模型的Token估算
func (g *GenericTokenizer) estimateChineseTokens(text string) int {
	// 中文模型通常对中文更友好
	// 英文：约4字符/token
	// 中文：约1个字符/token
	runeCount := utf8.RuneCountInString(text)
	chineseCount := 0

	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
		}
	}

	englishCount := runeCount - chineseCount
	return chineseCount + (englishCount / 4)
}

// estimateDefaultTokens 默认Token估算
func (g *GenericTokenizer) estimateDefaultTokens(text string) int {
	// 最保守的估算：英文4字符/token，中文1.5字符/token
	charCount := len(text)
	runeCount := utf8.RuneCountInString(text)

	if containsChinese(text) {
		return runeCount * 2 / 3
	}
	return charCount / 4
}

// estimateToolTokens 估算工具定义的Token数量
func (g *GenericTokenizer) estimateToolTokens(tools []Tool) int {
	toolsJSON, _ := json.Marshal(tools)
	return len(toolsJSON)/4 + len(tools)*10
}

// containsChinese 检查是否包含中文字符
func containsChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}
