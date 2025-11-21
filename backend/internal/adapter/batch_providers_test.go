package adapter

import (
	"testing"
	"time"
)

func TestNewDeepSeekAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "deepseek",
		BaseURL: "https://api.deepseek.com/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewDeepSeekAdapter(config)

	if adapter.Name() != "deepseek" {
		t.Errorf("Expected adapter name deepseek")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewMoonshotAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "moonshot",
		BaseURL: "https://api.moonshot.cn/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewMoonshotAdapter(config)

	if adapter.Name() != "moonshot" {
		t.Errorf("Expected adapter name moonshot")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewMinimaxAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "minimax",
		BaseURL: "https://api.minimax.chat/v1",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}

	adapter := NewMinimaxAdapter(config)

	if adapter.Name() != "minimax" {
		t.Errorf("Expected adapter name minimax")
	}

	models := adapter.GetSupportedModels()
	if len(models) == 0 {
		t.Errorf("Expected supported models")
	}
}

func TestNewGenericAdapter(t *testing.T) {
	config := &AdapterConfig{
		Type:    "custom",
		BaseURL: "https://api.custom.com/v1",
		Timeout: 30 * time.Second,
	}

	mapping := &AdapterMapping{
		RequestMapping:  make(map[string]string),
		ResponseMapping: make(map[string]string),
		FieldMapping:    make(map[string]string),
	}

	adapter := NewGenericAdapter(config, mapping)

	if adapter.Name() != "custom" {
		t.Errorf("Expected adapter name custom")
	}
}

func TestGenericAdapterConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "custom",
		BaseURL: "https://api.custom.com/v1",
		Timeout: 30 * time.Second,
	}

	customConvert := func(req *OpenAIRequest) (interface{}, error) {
		return map[string]interface{}{
			"model": req.Model,
			"text":  "custom format",
		}, nil
	}

	mapping := &AdapterMapping{
		CustomConvert: customConvert,
	}

	adapter := NewGenericAdapter(config, mapping)

	req := &OpenAIRequest{
		Model: "custom-model",
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}

	customReq := converted.(map[string]interface{})
	if customReq["model"] != "custom-model" {
		t.Errorf("Expected model custom-model")
	}
}

func TestDeepSeekConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "deepseek",
		BaseURL: "https://api.deepseek.com/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewDeepSeekAdapter(config)

	req := &OpenAIRequest{
		Model:       "deepseek-chat",
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

func TestMoonshotConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "moonshot",
		BaseURL: "https://api.moonshot.cn/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewMoonshotAdapter(config)

	req := &OpenAIRequest{
		Model: "moonshot-v1-8k",
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}
}

func TestMinimaxConvertRequest(t *testing.T) {
	config := &AdapterConfig{
		Type:    "minimax",
		BaseURL: "https://api.minimax.chat/v1",
		Timeout: 30 * time.Second,
	}

	adapter := NewMinimaxAdapter(config)

	req := &OpenAIRequest{
		Model: "abab6.5-chat",
	}

	converted, err := adapter.ConvertRequest(req)
	if err != nil {
		t.Errorf("ConvertRequest failed: %v", err)
	}

	if converted == nil {
		t.Errorf("Expected converted request")
	}
}

func BenchmarkAllBatchAdapters(b *testing.B) {
	adapters := []Adapter{
		NewDeepSeekAdapter(&AdapterConfig{
			Type:    "deepseek",
			BaseURL: "https://api.deepseek.com/v1",
			Timeout: 30 * time.Second,
		}),
		NewMoonshotAdapter(&AdapterConfig{
			Type:    "moonshot",
			BaseURL: "https://api.moonshot.cn/v1",
			Timeout: 30 * time.Second,
		}),
		NewMinimaxAdapter(&AdapterConfig{
			Type:    "minimax",
			BaseURL: "https://api.minimax.chat/v1",
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
