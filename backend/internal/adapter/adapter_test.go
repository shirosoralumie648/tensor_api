package adapter

import (
	"context"
	"testing"
	"time"
)

func TestNewBaseAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewBaseAdapter(config)

	if adapter == nil {
		t.Errorf("Expected adapter to be created")
	}

	if adapter.Name() != "openai" {
		t.Errorf("Expected adapter name openai")
	}
}

func TestOpenAIRequest(t *testing.T) {
	req := &OpenAIRequest{
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2000,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	if req.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4")
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message")
	}
}

func TestOpenAIResponse(t *testing.T) {
	resp := &OpenAIResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("Expected ID chatcmpl-123")
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice")
	}

	if resp.Usage.TotalTokens != 15 {
		t.Errorf("Expected 15 total tokens")
	}
}

func TestStreamChunk(t *testing.T) {
	chunk := &StreamChunk{
		ID:      "chatcmpl-123",
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   "gpt-4",
		Choices: []Choice{
			{
				Index: 0,
				Delta: &Message{
					Role:    "assistant",
					Content: "Hello",
				},
				FinishReason: "",
			},
		},
	}

	if chunk.ID != "chatcmpl-123" {
		t.Errorf("Expected ID chatcmpl-123")
	}

	if len(chunk.Choices) != 1 {
		t.Errorf("Expected 1 choice")
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello",
		Name:    "John",
	}

	if msg.Role != "user" {
		t.Errorf("Expected role user")
	}

	if msg.Content != "Hello" {
		t.Errorf("Expected content Hello")
	}
}

func TestTool(t *testing.T) {
	tool := Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "search",
			Description: "Search the web",
			Parameters: map[string]interface{}{
				"type": "object",
			},
		},
	}

	if tool.Type != "function" {
		t.Errorf("Expected type function")
	}

	if tool.Function.Name != "search" {
		t.Errorf("Expected function name search")
	}
}

func TestAdapterConfig(t *testing.T) {
	config := &AdapterConfig{
		Type:    "claude",
		BaseURL: "https://api.anthropic.com/v1",
		APIKey:  "sk-ant-",
		Timeout: 60 * time.Second,
	}

	if config.Type != "claude" {
		t.Errorf("Expected type claude")
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s")
	}
}

func TestBaseAdapterSetSupportedModels(t *testing.T) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewBaseAdapter(config)

	models := []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"}
	adapter.SetSupportedModels(models)

	retrieved := adapter.GetSupportedModels()
	if len(retrieved) != 3 {
		t.Errorf("Expected 3 models")
	}

	if retrieved[0] != "gpt-4" {
		t.Errorf("Expected first model gpt-4")
	}
}

func TestAdapterError(t *testing.T) {
	err := NewAdapterError("INVALID_REQUEST", "Request validation failed")

	if err.Code != "INVALID_REQUEST" {
		t.Errorf("Expected code INVALID_REQUEST")
	}

	if err.Message != "Request validation failed" {
		t.Errorf("Expected correct message")
	}

	errStr := err.Error()
	if len(errStr) == 0 {
		t.Errorf("Expected non-empty error string")
	}
}

func TestAdapterMetrics(t *testing.T) {
	metrics := AdapterMetrics{
		TotalRequests:   100,
		SuccessRequests: 95,
		FailedRequests:  5,
		AvgResponseTime: 150,
		LastRequestTime: time.Now(),
	}

	if metrics.TotalRequests != 100 {
		t.Errorf("Expected 100 total requests")
	}

	if metrics.SuccessRequests != 95 {
		t.Errorf("Expected 95 success requests")
	}
}

func TestAdapterInfo(t *testing.T) {
	info := AdapterInfo{
		Name:                 "OpenAI",
		Type:                 "openai",
		SupportedModelCount:  5,
		Version:              "1.0.0",
		Status:               "active",
		CreatedAt:            time.Now(),
	}

	if info.Name != "OpenAI" {
		t.Errorf("Expected name OpenAI")
	}

	if info.Status != "active" {
		t.Errorf("Expected status active")
	}
}

func TestNewRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewBaseAdapter(config)

	ctx := context.Background()
	req, err := adapter.NewRequest(ctx, "POST", "/chat/completions", nil)

	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	if req != nil && req.Method != "POST" { //nolint:SA5011
		t.Errorf("Expected method POST, got %s", req.Method)
	}
}

func TestToolFunction(t *testing.T) {
	fn := ToolFunction{
		Name:        "get_weather",
		Description: "Get weather information",
		Parameters: map[string]interface{}{
			"location": "string",
		},
	}

	if fn.Name != "get_weather" {
		t.Errorf("Expected name get_weather")
	}

	if fn.Description != "Get weather information" {
		t.Errorf("Expected correct description")
	}
}

func TestUsage(t *testing.T) {
	usage := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	if usage.TotalTokens != 150 {
		t.Errorf("Expected 150 total tokens")
	}

	if usage.PromptTokens != 100 {
		t.Errorf("Expected 100 prompt tokens")
	}
}

func TestErrorInfo(t *testing.T) {
	errInfo := ErrorInfo{
		Message: "Invalid API key",
		Type:    "invalid_request_error",
		Code:    "invalid_api_key",
	}

	if errInfo.Message != "Invalid API key" {
		t.Errorf("Expected correct message")
	}

	if errInfo.Type != "invalid_request_error" {
		t.Errorf("Expected correct type")
	}
}

func BenchmarkNewBaseAdapter(b *testing.B) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBaseAdapter(config)
	}
}

func BenchmarkGetSupportedModels(b *testing.B) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewBaseAdapter(config)
	adapter.SetSupportedModels([]string{"gpt-4", "gpt-3.5-turbo"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.GetSupportedModels()
	}
}

