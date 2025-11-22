package tools

import (
	"context"
	"testing"
	"time"
)

func TestWebSearchTool(t *testing.T) {
	tool := NewWebSearchTool("google", "test-api-key")
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"query": "golang",
	})

	if err != nil {
		t.Errorf("WebSearchTool failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected search results")
	}

	results, ok := result.([]SearchResult)
	if !ok || len(results) == 0 {
		t.Errorf("Expected search results array")
	}
}

func TestWebSearchToolGetTool(t *testing.T) {
	tool := NewWebSearchTool("google", "")
	toolDef := tool.GetTool()

	if toolDef.Name != "web_search" {
		t.Errorf("Expected tool name web_search")
	}

	if toolDef.Handler == nil {
		t.Errorf("Expected handler to be set")
	}
}

func TestCodeExecutorTool(t *testing.T) {
	tool := NewCodeExecutorTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"language": "python",
		"code":     "print('hello')",
	})

	if err != nil {
		t.Errorf("CodeExecutorTool failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected execution result")
	}
}

func TestCodeExecutorToolUnsupported(t *testing.T) {
	tool := NewCodeExecutorTool()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"language": "rust",
		"code":     "fn main() {}",
	})

	if err == nil {
		t.Errorf("Expected error for unsupported language")
	}
}

func TestHTTPRequestTool(t *testing.T) {
	tool := NewHTTPRequestTool()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url":    "https://www.google.com",
		"method": "GET",
	})

	// Note: This test might fail if internet is not available
	if err == nil && result == nil {
		t.Errorf("Expected HTTP result")
	}
}

func TestHTTPRequestToolAllowedDomains(t *testing.T) {
	tool := NewHTTPRequestTool()

	if !tool.isDomainAllowed("example.com") {
		t.Errorf("Expected domain to be allowed")
	}

	tool.allowedDomains = []string{"example.com"}
	if tool.isDomainAllowed("other.com") {
		t.Errorf("Expected domain to be blocked")
	}
}

func TestHTTPRequestToolGetTool(t *testing.T) {
	tool := NewHTTPRequestTool()
	toolDef := tool.GetTool()

	if toolDef.Name != "http_request" {
		t.Errorf("Expected tool name http_request")
	}
}

func TestBuiltinToolsFactory(t *testing.T) {
	engine := NewFunctionEngine()
	factory := NewBuiltinToolsFactory(engine)

	err := factory.RegisterAllTools()
	if err != nil {
		t.Errorf("RegisterAllTools failed: %v", err)
	}

	tools := engine.ListTools()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{"web_search", "code_executor", "http_request"}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestCodeExecutorToolGetTool(t *testing.T) {
	tool := NewCodeExecutorTool()
	toolDef := tool.GetTool()

	if toolDef.Name != "code_executor" {
		t.Errorf("Expected tool name code_executor")
	}

	if toolDef.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s")
	}
}

func BenchmarkWebSearchTool(b *testing.B) {
	tool := NewWebSearchTool("google", "")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]interface{}{
			"query": "golang",
		})
	}
}
