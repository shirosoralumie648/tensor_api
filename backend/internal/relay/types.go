package relay

// ChatMessage 代表对话中的一条消息
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatCompletionRequest 标准的 OpenAI 格式请求
type ChatCompletionRequest struct {
	Model            string                 `json:"model" binding:"required"`
	Messages         []ChatMessage          `json:"messages" binding:"required"`
	Temperature      float64                `json:"temperature"`
	TopP             float64                `json:"top_p"`
	MaxTokens        int                    `json:"max_tokens"`
	Stream           bool                   `json:"stream"`
	FrequencyPenalty float64                `json:"frequency_penalty"`
	PresencePenalty  float64                `json:"presence_penalty"`
	Functions        []map[string]interface{} `json:"functions"`
	FunctionCall     interface{}            `json:"function_call"`
	Tools            []map[string]interface{} `json:"tools"`
	ToolChoice       interface{}            `json:"tool_choice"`
}

// ChatCompletionResponse 标准的 OpenAI 格式响应
type ChatCompletionResponse struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	Choices           []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		Delta        *ChatMessage `json:"delta,omitempty"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *ErrorResponse `json:"error,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

