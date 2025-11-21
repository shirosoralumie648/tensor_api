package adapter

import "context"

// AIProvider 定义 AI 提供商统一接口
type AIProvider interface {
	// Chat 发送单次请求并获取完整响应
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// ChatStream 流式响应，返回通道接收增量消息
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamDelta, error)

	// ListModels 获取该提供商支持的模型列表
	ListModels(ctx context.Context) ([]Model, error)

	// HealthCheck 检查连接状态
	HealthCheck(ctx context.Context) error

	// GetName 获取提供商名称
	GetName() string
}

// ChatRequest 统一的聊天请求格式
type ChatRequest struct {
	Model       string     `json:"model"`              // 模型名称，如 "gpt-4", "claude-3-opus"
	Messages    []Message  `json:"messages"`           // 对话消息历史
	Temperature float32    `json:"temperature"`        // 生成温度 (0-2)，控制随机性
	MaxTokens   int        `json:"max_tokens"`         // 最大输出 token 数
	TopP        float32    `json:"top_p"`              // TopP 采样参数
	TopK        int        `json:"top_k,omitempty"`    // TopK 采样参数（某些模型支持）
	Stream      bool       `json:"stream"`             // 是否使用流式响应
	User        string     `json:"user,omitempty"`     // 用户标识
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // 额外元数据
}

// Message 对话消息
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // 消息内容
	Name    string `json:"name,omitempty"` // 可选的消息发送者名称
}

// ChatResponse 统一的完整响应格式
type ChatResponse struct {
	ID            string `json:"id"`             // 响应 ID
	Model         string `json:"model"`          // 使用的模型
	Content       string `json:"content"`        // 响应内容
	Tokens        Usage  `json:"tokens"`         // token 使用情况
	FinishReason  string `json:"finish_reason"`  // 完成原因：stop, max_tokens, error 等
	Provider      string `json:"provider"`       // 提供商名称
	ResponseTime  int64  `json:"response_time"`  // 响应时间（毫秒）
}

// StreamDelta 流式响应中的增量数据
type StreamDelta struct {
	Content     string      `json:"content,omitempty"`    // 增量内容
	Index       int         `json:"index,omitempty"`      // 消息索引
	FinishReason string      `json:"finish_reason,omitempty"` // 完成原因
	Tokens      Usage       `json:"tokens,omitempty"`     // token 使用情况
	Error       error       `json:"-"`                     // 错误信息
	Done        bool        `json:"done"`                  // 是否完成
}

// Usage token 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 输入 token 数
	CompletionTokens int `json:"completion_tokens"` // 输出 token 数
	TotalTokens      int `json:"total_tokens"`      // 总 token 数
	CostUSD          float32 `json:"cost_usd,omitempty"` // 成本（USD）
}

// Model AI 模型信息
type Model struct {
	ID                string  `json:"id"`                   // 模型标识符
	Name              string  `json:"name"`                 // 模型显示名称
	Provider          string  `json:"provider"`             // 提供商名称
	Type              string  `json:"type"`                 // 模型类型：text, vision, embedding
	ContextSize       int     `json:"context_size"`         // 上下文窗口大小
	MaxOutputTokens   int     `json:"max_output_tokens"`    // 最大输出 token 数
	CostPer1KPrompt   float32 `json:"cost_per_1k_prompt"`   // 每 1000 个输入 token 的成本（USD）
	CostPer1KCompletion float32 `json:"cost_per_1k_completion"` // 每 1000 个输出 token 的成本（USD）
	IsActive          bool    `json:"is_active"`            // 是否可用
	Description       string  `json:"description,omitempty"` // 模型描述
	ReleaseDate       string  `json:"release_date,omitempty"` // 发布日期
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name        string            // 提供商名称
	APIKey      string            // API 密钥
	APIBaseURL  string            // API 基础 URL
	Timeout     int               // 请求超时（秒）
	MaxRetries  int               // 最大重试次数
	Models      map[string]Model  // 支持的模型映射
	Metadata    map[string]interface{} // 额外配置
}

// AdapterError 适配器特定错误
type AdapterError struct {
	Provider string
	Code     string
	Message  string
	Err      error
}

func (e *AdapterError) Error() string {
	return "adapter error: " + e.Message
}

