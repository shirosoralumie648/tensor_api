package adapter

import (
	"testing"
	"time"
)

func TestNewOpenAIAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewOpenAIAdapter(config)

	if adapter.Name() != "openai" {
		t.Errorf("Expected adapter name openai")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewClaudeAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "claude",
		BaseURL: "https://api.anthropic.com/v1",
		APIKey:  "sk-ant-",
		Timeout: 30 * time.Second,
	}

	adapter := NewClaudeAdapter(config)

	if adapter.Name() != "claude" {
		t.Errorf("Expected adapter name claude")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewGeminiAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "gemini",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewGeminiAdapter(config)

	if adapter.Name() != "gemini" {
		t.Errorf("Expected adapter name gemini")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewBaiduAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "baidu",
		BaseURL: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewBaiduAdapter(config)

	if adapter.Name() != "baidu" {
		t.Errorf("Expected adapter name baidu")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewQwenAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "qwen",
		BaseURL: "https://dashscope.aliyuncs.com/api/v1",
		APIKey:  "sk-",
		Timeout: 30 * time.Second,
	}

	adapter := NewQwenAdapter(config)

	if adapter.Name() != "qwen" {
		t.Errorf("Expected adapter name qwen")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestOpenAIConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewOpenAIAdapter(config)

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

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}
}

func TestClaudeConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "claude",
		BaseURL: "https://api.anthropic.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewClaudeAdapter(config)

	req := &OpenAIRequest{
		Model:       "claude-3-opus",
		Temperature: 0.7,
		MaxTokens:   2000,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}

	claudeReq, ok := converted.(map[string]interface{})
	if !ok {
		t.Errorf("Expected map[string]interface{}")
	}

	if claudeReq["model"] != "claude-3-opus" {
		t.Errorf("Expected model claude-3-opus")
	}
}

func TestGeminiConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "gemini",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		Timeout: 30 * time.Second,
	}

	adapter := NewGeminiAdapter(config)

	req := &OpenAIRequest{
		Model:       "gemini-pro",
		Temperature: 0.7,
		MaxTokens:   2000,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}
}

func TestBaiduConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "baidu",
		BaseURL: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop",
		Timeout: 30 * time.Second,
	}

	adapter := NewBaiduAdapter(config)

	req := &OpenAIRequest{
		Model:       "eb-4",
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}
}

func TestQwenConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "qwen",
		BaseURL: "https://dashscope.aliyuncs.com/api/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewQwenAdapter(config)

	req := &OpenAIRequest{
		Model:       "qwen-turbo",
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}
}

func TestExtractUsage(t *testing.T) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewOpenAIAdapter(config)

	resp := &OpenAIResponse{
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	usage, err := adapter.ExtractUsage(resp)
	if err != nil {
		t.Errorf("ExtractUsage failed: %v", err)
	}

	if usage.TotalTokens != 15 {
		t.Errorf("Expected 15 total tokens")
	}
}

func BenchmarkOpenAIConvertRequest(b *testing.B) {
	config := &AdapterConfig{
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewOpenAIAdapter(config)

	req := &OpenAIRequest{
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.ConvertRequest(req)
	}
}

func BenchmarkAllAdapters(b *testing.B) {
	adapters := []Adapter{
		NewOpenAIAdapter(&AdapterConfig{
			Type:    "openai",
			BaseURL: "https://api.openai.com/v1",
			Timeout: 30 * time.Second,
		}),
		NewClaudeAdapter(&AdapterConfig{
			Type:    "claude",
			BaseURL: "https://api.anthropic.com/v1",
			Timeout: 30 * time.Second,
		}),
		NewGeminiAdapter(&AdapterConfig{
			Type:    "gemini",
			BaseURL: "https://generativelanguage.googleapis.com/v1beta/models",
			Timeout: 30 * time.Second,
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, adapter := range adapters {
			adapter.GetSupportedModels()
		}
	}
}

