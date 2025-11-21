package tools

import (
	"context"
	"os"
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

func TestFileOperationToolRead(t *testing.T) {
	tool := NewFileOperationTool("/tmp")
	ctx := context.Background()

	// Create test file
	testFile := "/tmp/test_read.txt"
	content := "test content"
	os.WriteFile(testFile, []byte(content), 0644)
	defer os.Remove(testFile)

	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "read",
		"path":      "test_read.txt",
	})

	if err != nil {
		t.Errorf("FileOperationTool read failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected read result")
	}
}

func TestFileOperationToolWrite(t *testing.T) {
	tool := NewFileOperationTool("/tmp")
	ctx := context.Background()

	testFile := "/tmp/test_write.txt"
	defer os.Remove(testFile)

	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "write",
		"path":      "test_write.txt",
		"content":   "test write content",
	})

	if err != nil {
		t.Errorf("FileOperationTool write failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected write result")
	}

	// Verify file was written
	data, err := os.ReadFile(testFile)
	if err != nil || string(data) != "test write content" {
		t.Errorf("File content mismatch")
	}
}

func TestFileOperationToolDelete(t *testing.T) {
	tool := NewFileOperationTool("/tmp")
	ctx := context.Background()

	testFile := "/tmp/test_delete.txt"
	os.WriteFile(testFile, []byte("content"), 0644)

	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "delete",
		"path":      "test_delete.txt",
	})

	if err != nil {
		t.Errorf("FileOperationTool delete failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected delete result")
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); err == nil {
		t.Errorf("File should be deleted")
	}
}

func TestFileOperationToolList(t *testing.T) {
	tool := NewFileOperationTool("/tmp")
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "list",
		"path":      "",
	})

	if err != nil {
		t.Errorf("FileOperationTool list failed: %v", err)
	}

	if result == nil {
		t.Errorf("Expected list result")
	}
}

func TestFileOperationToolPathSecurity(t *testing.T) {
	tool := NewFileOperationTool("/tmp")
	ctx := context.Background()

	// Try to access path outside allowed directory
	_, err := tool.Execute(ctx, map[string]interface{}{
		"operation": "read",
		"path":      "../../etc/passwd",
	})

	if err == nil {
		t.Errorf("Expected security error for path traversal")
	}
}

func TestFileOperationToolGetTool(t *testing.T) {
	tool := NewFileOperationTool("/tmp")
	toolDef := tool.GetTool()

	if toolDef.Name != "file_operations" {
		t.Errorf("Expected tool name file_operations")
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
	if len(tools) != 4 {
		t.Errorf("Expected 4 tools, got %d", len(tools))
	}

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{"web_search", "code_executor", "file_operations", "http_request"}
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

func BenchmarkFileOperationToolRead(b *testing.B) {
	tool := NewFileOperationTool("/tmp")
	ctx := context.Background()

	// Create test file
	testFile := "/tmp/bench_test.txt"
	os.WriteFile(testFile, []byte("test content"), 0644)
	defer os.Remove(testFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, map[string]interface{}{
			"operation": "read",
			"path":      "bench_test.txt",
		})
	}
}

