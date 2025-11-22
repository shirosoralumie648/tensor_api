package tokenizer

import (
	"context"
)

// TokenCount Token计数结果
type TokenCount struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TokenCountRequest Token计数请求
type TokenCountRequest struct {
	Model        string        `json:"model"`
	Messages     []Message     `json:"messages"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
	Tools        []Tool        `json:"tools,omitempty"`
	ImageDetails []ImageDetail `json:"image_details,omitempty"`
	AudioDetails []AudioDetail `json:"audio_details,omitempty"`
}

// Message 消息
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string 或 []ContentPart
	Name    string      `json:"name,omitempty"`
}

// ContentPart 内容部分（多模态）
type ContentPart struct {
	Type     string    `json:"type"` // text, image_url, audio_url
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
	AudioURL *AudioURL `json:"audio_url,omitempty"`
}

// ImageURL 图片URL
type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // auto, low, high
}

// AudioURL 音频URL
type AudioURL struct {
	URL    string `json:"url"`
	Format string `json:"format,omitempty"` // mp3, wav, opus
}

// ImageDetail 图片详情
type ImageDetail struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Detail string `json:"detail"` // auto, low, high
}

// AudioDetail 音频详情
type AudioDetail struct {
	Duration int    `json:"duration"` // 秒
	Format   string `json:"format"`   // mp3, wav, opus
}

// Tool 工具定义
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 工具函数
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// Tokenizer Token计数器接口
type Tokenizer interface {
	// CountTokens 计算Token数量
	CountTokens(ctx context.Context, req *TokenCountRequest) (*TokenCount, error)

	// CountText 计算纯文本Token数量
	CountText(ctx context.Context, text string, model string) (int, error)

	// CountMessages 计算消息列表Token数量
	CountMessages(ctx context.Context, messages []Message, model string) (int, error)

	// EstimateImageTokens 估算图片Token数量
	EstimateImageTokens(ctx context.Context, detail ImageDetail) (int, error)

	// EstimateAudioTokens 估算音频Token数量
	EstimateAudioTokens(ctx context.Context, detail AudioDetail) (int, error)

	// SupportedModels 返回支持的模型列表
	SupportedModels() []string
}

// StreamTokenCounter 流式Token计数器
type StreamTokenCounter interface {
	// AddChunk 添加流式数据块
	AddChunk(chunk string) error

	// GetCurrentCount 获取当前Token计数
	GetCurrentCount() int

	// Reset 重置计数器
	Reset()

	// Finalize 完成计数（返回最终结果）
	Finalize() int
}
