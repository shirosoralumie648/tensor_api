package chat

import (
	"context"
	"testing"
	"time"
)

func TestFunctionExecutorRegisterFunction(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name:        "greet",
		Description: "Greet someone",
		Parameters: []*FunctionParameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Person's name",
				Required:    true,
			},
		},
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "Hello", nil
	}

	err := fe.RegisterFunction(spec, handler)
	if err != nil {
		t.Errorf("RegisterFunction failed: %v", err)
	}
}

func TestFunctionExecutorGetFunction(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name:        "greet",
		Description: "Greet someone",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "Hello", nil
	}

	fe.RegisterFunction(spec, handler)

	retrieved, err := fe.GetFunction("greet")
	if err != nil {
		t.Errorf("GetFunction failed: %v", err)
	}

	if retrieved.Name != "greet" {
		t.Errorf("Expected greet function")
	}
}

func TestFunctionExecutorExecute(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name:        "add",
		Description: "Add two numbers",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return 5, nil
	}

	fe.RegisterFunction(spec, handler)

	req := &FunctionRequest{
		ID:           "req-1",
		FunctionName: "add",
		Arguments:    map[string]interface{}{"a": 2, "b": 3},
		Context:      context.Background(),
		CreatedAt:    time.Now(),
	}

	response := fe.Execute(req)

	if response.Status != "success" {
		t.Errorf("Expected success status")
	}

	if response.Result != 5 {
		t.Errorf("Expected result 5")
	}
}

func TestFunctionExecutorExecuteTimeout(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name:        "slow",
		Description: "Slow function",
		TimeoutMS:   100,
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "done", nil
	}

	fe.RegisterFunction(spec, handler)

	req := &FunctionRequest{
		ID:           "req-1",
		FunctionName: "slow",
		Arguments:    map[string]interface{}{},
		Context:      context.Background(),
		CreatedAt:    time.Now(),
	}

	response := fe.Execute(req)

	if response.Status != "timeout" {
		t.Errorf("Expected timeout status, got %s", response.Status)
	}
}

func TestFunctionExecutorUnregisterFunction(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name: "test",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return nil, nil
	}

	fe.RegisterFunction(spec, handler)

	err := fe.UnregisterFunction("test")
	if err != nil {
		t.Errorf("UnregisterFunction failed: %v", err)
	}

	_, err = fe.GetFunction("test")
	if err == nil {
		t.Errorf("Expected error after unregistering")
	}
}

func TestFunctionExecutorGetAllFunctions(t *testing.T) {
	fe := NewFunctionExecutor()

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return nil, nil
	}

	for i := 0; i < 3; i++ {
		spec := &FunctionSpec{
			Name: "func-" + string(rune(i)),
		}
		fe.RegisterFunction(spec, handler)
	}

	specs := fe.GetAllFunctions()

	if len(specs) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(specs))
	}
}

func TestFunctionExecutorGetHistory(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name: "test",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	fe.RegisterFunction(spec, handler)

	for i := 0; i < 3; i++ {
		req := &FunctionRequest{
			ID:           "req-" + string(rune(i)),
			FunctionName: "test",
			Arguments:    map[string]interface{}{},
			Context:      context.Background(),
			CreatedAt:    time.Now(),
		}
		fe.Execute(req)
	}

	history := fe.GetHistory(10)

	if len(history) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(history))
	}
}

func TestFunctionExecutorGetStatistics(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name: "test",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	fe.RegisterFunction(spec, handler)

	req := &FunctionRequest{
		ID:           "req-1",
		FunctionName: "test",
		Arguments:    map[string]interface{}{},
		Context:      context.Background(),
		CreatedAt:    time.Now(),
	}

	fe.Execute(req)

	stats := fe.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalCalls, ok := stats["total_calls"].(int64); !ok || totalCalls != 1 {
		t.Errorf("Expected total_calls to be 1")
	}
}

func TestBatchExecutor(t *testing.T) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name: "echo",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "echoed", nil
	}

	fe.RegisterFunction(spec, handler)

	be := NewBatchExecutor(fe, 2)

	requests := make([]*FunctionRequest, 3)
	for i := 0; i < 3; i++ {
		requests[i] = &FunctionRequest{
			ID:           "req-" + string(rune(i)),
			FunctionName: "echo",
			Arguments:    map[string]interface{}{},
			Context:      context.Background(),
			CreatedAt:    time.Now(),
		}
	}

	responses := be.ExecuteBatch(requests)

	if len(responses) != 3 {
		t.Errorf("Expected 3 responses")
	}

	for _, resp := range responses {
		if resp.Status != "success" {
			t.Errorf("Expected success status")
		}
	}
}

func TestFunctionExecutorFunctionNotFound(t *testing.T) {
	fe := NewFunctionExecutor()

	req := &FunctionRequest{
		ID:           "req-1",
		FunctionName: "nonexistent",
		Arguments:    map[string]interface{}{},
		Context:      context.Background(),
		CreatedAt:    time.Now(),
	}

	response := fe.Execute(req)

	if response.Status != "failed" {
		t.Errorf("Expected failed status")
	}

	if response.Error == "" {
		t.Errorf("Expected error message")
	}
}

func TestBatchExecutorGetStatistics(t *testing.T) {
	fe := NewFunctionExecutor()
	be := NewBatchExecutor(fe, 2)

	stats := be.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}
}

func BenchmarkFunctionExecute(b *testing.B) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name: "bench",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	fe.RegisterFunction(spec, handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &FunctionRequest{
			ID:           "req-" + string(rune(i)),
			FunctionName: "bench",
			Arguments:    map[string]interface{}{},
			Context:      context.Background(),
			CreatedAt:    time.Now(),
		}
		_ = fe.Execute(req)
	}
}

func BenchmarkBatchExecute(b *testing.B) {
	fe := NewFunctionExecutor()

	spec := &FunctionSpec{
		Name: "bench",
	}

	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	fe.RegisterFunction(spec, handler)

	be := NewBatchExecutor(fe, 4)

	requests := make([]*FunctionRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = &FunctionRequest{
			ID:           "req-" + string(rune(i)),
			FunctionName: "bench",
			Arguments:    map[string]interface{}{},
			Context:      context.Background(),
			CreatedAt:    time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = be.ExecuteBatch(requests)
	}
}

