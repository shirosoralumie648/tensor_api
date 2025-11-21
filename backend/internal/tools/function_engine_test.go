package tools

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFunctionEngineRegisterTool(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name:        "test_tool",
		Description: "Test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "success", nil
		},
	}

	err := engine.RegisterTool(tool)
	if err != nil {
		t.Errorf("RegisterTool failed: %v", err)
	}

	retrieved, err := engine.GetTool("test_tool")
	if err != nil {
		t.Errorf("GetTool failed: %v", err)
	}

	if retrieved.Name != "test_tool" {
		t.Errorf("Expected tool name test_tool")
	}
}

func TestFunctionEngineRegisterToolError(t *testing.T) {
	engine := NewFunctionEngine()

	// Test empty name
	tool := &Tool{
		Name: "",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	err := engine.RegisterTool(tool)
	if err == nil {
		t.Errorf("Expected error for empty tool name")
	}

	// Test nil handler
	tool.Name = "test"
	tool.Handler = nil

	err = engine.RegisterTool(tool)
	if err == nil {
		t.Errorf("Expected error for nil handler")
	}
}

func TestFunctionEngineExecuteTool(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name:        "add",
		Description: "Add two numbers",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return a + b, nil
		},
		Timeout: 5 * time.Second,
	}

	engine.RegisterTool(tool)

	ctx := context.Background()
	result, err := engine.ExecuteTool(ctx, "add", map[string]interface{}{
		"a": 5.0,
		"b": 3.0,
	})

	if err != nil {
		t.Errorf("ExecuteTool failed: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("Expected status success, got %s", result.Status)
	}

	if result.Result != 8.0 {
		t.Errorf("Expected result 8.0, got %v", result.Result)
	}
}

func TestFunctionEngineExecuteToolError(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "error_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, errors.New("test error")
		},
		Timeout: 5 * time.Second,
	}

	engine.RegisterTool(tool)

	ctx := context.Background()
	result, err := engine.ExecuteTool(ctx, "error_tool", map[string]interface{}{})

	if err == nil {
		t.Errorf("Expected error")
	}

	if result.Status != "error" {
		t.Errorf("Expected status error")
	}
}

func TestFunctionEngineTimeout(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "slow_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			time.Sleep(2 * time.Second)
			return "done", nil
		},
		Timeout: 500 * time.Millisecond,
	}

	engine.RegisterTool(tool)

	ctx := context.Background()
	result, err := engine.ExecuteTool(ctx, "slow_tool", map[string]interface{}{})

	if err == nil {
		t.Errorf("Expected timeout error")
	}

	if result.Status != "timeout" {
		t.Errorf("Expected status timeout, got %s", result.Status)
	}
}

func TestFunctionEngineListTools(t *testing.T) {
	engine := NewFunctionEngine()

	tool1 := &Tool{
		Name: "tool1",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	tool2 := &Tool{
		Name: "tool2",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	engine.RegisterTool(tool1)
	engine.RegisterTool(tool2)

	tools := engine.ListTools()

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestFunctionEngineUnregisterTool(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "test",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	engine.RegisterTool(tool)
	err := engine.UnregisterTool("test")

	if err != nil {
		t.Errorf("UnregisterTool failed: %v", err)
	}

	_, err = engine.GetTool("test")
	if err == nil {
		t.Errorf("Expected error after unregistering")
	}
}

func TestFunctionEngineStatistics(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "test",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
		Timeout: 5 * time.Second,
	}

	engine.RegisterTool(tool)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		engine.ExecuteTool(ctx, "test", map[string]interface{}{})
	}

	stats := engine.GetStatistics()

	if totalCalls, ok := stats["total_calls"].(int64); !ok || totalCalls != 5 {
		t.Errorf("Expected 5 total calls")
	}

	if successRate, ok := stats["success_rate"].(float32); !ok || successRate != 1.0 {
		t.Error("Expected 100% success rate")
	}
}

func TestToolBuilder(t *testing.T) {
	tool := NewToolBuilder("multiply").
		Description("Multiply two numbers").
		Parameters(&JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"a": {Type: "number"},
				"b": {Type: "number"},
			},
			Required: []string{"a", "b"},
		}).
		Handler(func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return a * b, nil
		}).
		Timeout(10 * time.Second).
		Build()

	if tool.Name != "multiply" {
		t.Errorf("Expected tool name multiply")
	}

	if tool.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s")
	}
}

func TestConvertToOpenAIFormat(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name:        "test",
		Description: "Test tool",
		Parameters: &JSONSchema{
			Type: "object",
			Properties: map[string]*JSONSchema{
				"param": {Type: "string"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	engine.RegisterTool(tool)

	openaiFormat := engine.ConvertToOpenAIFormat()

	if len(openaiFormat) != 1 {
		t.Errorf("Expected 1 function in OpenAI format")
	}

	function := openaiFormat[0]["function"].(map[string]interface{})
	if function["name"] != "test" {
		t.Errorf("Expected function name test")
	}
}

func TestParseOpenAIToolCall(t *testing.T) {
	engine := NewFunctionEngine()

	toolCall := map[string]interface{}{
		"name":      "test_tool",
		"arguments": `{"param": "value"}`,
	}

	call, err := engine.ParseOpenAIToolCall(toolCall)
	if err != nil {
		t.Errorf("ParseOpenAIToolCall failed: %v", err)
	}

	if call.ToolName != "test_tool" {
		t.Errorf("Expected tool name test_tool")
	}

	if call.Arguments["param"] != "value" {
		t.Errorf("Expected param value")
	}
}

func TestExecutionContext(t *testing.T) {
	engine := NewFunctionEngine()
	ctx := NewExecutionContext(engine)

	err := ctx.PushCall("tool1")
	if err != nil {
		t.Errorf("PushCall failed: %v", err)
	}

	if ctx.GetDepth() != 1 {
		t.Errorf("Expected depth 1")
	}

	// Test recursive call detection
	err = ctx.PushCall("tool1")
	if err == nil {
		t.Errorf("Expected recursion error")
	}

	ctx.PopCall()
	if ctx.GetDepth() != 0 {
		t.Errorf("Expected depth 0 after pop")
	}
}

func TestBatchExecute(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "add",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return a + b, nil
		},
		Timeout: 5 * time.Second,
	}

	engine.RegisterTool(tool)

	calls := []*ToolCall{
		{
			ID:       "call1",
			ToolName: "add",
			Arguments: map[string]interface{}{
				"a": 1.0,
				"b": 2.0,
			},
		},
		{
			ID:       "call2",
			ToolName: "add",
			Arguments: map[string]interface{}{
				"a": 3.0,
				"b": 4.0,
			},
		},
	}

	ctx := context.Background()
	results, err := engine.BatchExecute(ctx, calls)

	if err != nil {
		t.Errorf("BatchExecute failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results")
	}

	if results[0].Status != "success" || results[1].Status != "success" {
		t.Errorf("Expected both calls to succeed")
	}
}

func TestCallHistory(t *testing.T) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "test",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
		Timeout: 5 * time.Second,
	}

	engine.RegisterTool(tool)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		engine.ExecuteTool(ctx, "test", map[string]interface{}{})
	}

	history := engine.GetCallHistory(10)

	if len(history) != 5 {
		t.Errorf("Expected 5 calls in history")
	}
}

func BenchmarkExecuteTool(b *testing.B) {
	engine := NewFunctionEngine()

	tool := &Tool{
		Name: "test",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
		Timeout: 5 * time.Second,
	}

	engine.RegisterTool(tool)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ExecuteTool(ctx, "test", map[string]interface{}{})
	}
}

func BenchmarkConvertToOpenAIFormat(b *testing.B) {
	engine := NewFunctionEngine()

	for i := 0; i < 10; i++ {
		tool := &Tool{
			Name: string(rune(i)),
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		}
		engine.RegisterTool(tool)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ConvertToOpenAIFormat()
	}
}

