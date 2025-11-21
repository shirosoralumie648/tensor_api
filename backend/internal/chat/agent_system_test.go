package chat

import (
	"testing"
	"time"
)

func TestSystemPromptManagerAddPrompt(t *testing.T) {
	spm := NewSystemPromptManager()

	prompt := &SystemPrompt{
		ID:          "prompt-1",
		Content:     "You are a helpful assistant",
		Role:        "assistant",
		Model:       "gpt-4",
		Temperature: 0.7,
	}

	err := spm.AddPrompt(prompt)
	if err != nil {
		t.Errorf("AddPrompt failed: %v", err)
	}

	if prompt.Version != 1 {
		t.Errorf("Expected version 1, got %d", prompt.Version)
	}
}

func TestSystemPromptManagerUpdatePrompt(t *testing.T) {
	spm := NewSystemPromptManager()

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "Original content",
		Role:    "assistant",
	}

	spm.AddPrompt(prompt)

	updated, err := spm.UpdatePrompt("prompt-1", "Updated content")
	if err != nil {
		t.Errorf("UpdatePrompt failed: %v", err)
	}

	if updated.Content != "Updated content" {
		t.Errorf("Expected 'Updated content'")
	}

	if updated.Version != 2 {
		t.Errorf("Expected version 2")
	}
}

func TestSystemPromptManagerGetPrompt(t *testing.T) {
	spm := NewSystemPromptManager()

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "Test content",
	}

	spm.AddPrompt(prompt)

	retrieved, err := spm.GetPrompt("prompt-1")
	if err != nil {
		t.Errorf("GetPrompt failed: %v", err)
	}

	if retrieved.Content != "Test content" {
		t.Errorf("Expected 'Test content'")
	}
}

func TestSystemPromptManagerDeletePrompt(t *testing.T) {
	spm := NewSystemPromptManager()

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "Test content",
	}

	spm.AddPrompt(prompt)

	err := spm.DeletePrompt("prompt-1")
	if err != nil {
		t.Errorf("DeletePrompt failed: %v", err)
	}

	_, err = spm.GetPrompt("prompt-1")
	if err == nil {
		t.Errorf("Expected error when getting deleted prompt")
	}
}

func TestToolRegistryRegisterTool(t *testing.T) {
	tr := NewToolRegistry()

	toolDef := &ToolDefinition{
		Name:        "search",
		Description: "Search for information",
		Parameters:  map[string]interface{}{"query": "string"},
	}

	handler := func(args map[string]interface{}) (interface{}, error) {
		return "search result", nil
	}

	err := tr.RegisterTool(toolDef, handler)
	if err != nil {
		t.Errorf("RegisterTool failed: %v", err)
	}
}

func TestToolRegistryCallTool(t *testing.T) {
	tr := NewToolRegistry()

	toolDef := &ToolDefinition{
		Name:        "add",
		Description: "Add two numbers",
	}

	handler := func(args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	tr.RegisterTool(toolDef, handler)

	result, err := tr.CallTool("add", map[string]interface{}{})
	if err != nil {
		t.Errorf("CallTool failed: %v", err)
	}

	if result != "result" {
		t.Errorf("Expected 'result'")
	}
}

func TestToolRegistryUnregisterTool(t *testing.T) {
	tr := NewToolRegistry()

	toolDef := &ToolDefinition{
		Name: "search",
	}

	handler := func(args map[string]interface{}) (interface{}, error) {
		return nil, nil
	}

	tr.RegisterTool(toolDef, handler)

	err := tr.UnregisterTool("search")
	if err != nil {
		t.Errorf("UnregisterTool failed: %v", err)
	}

	_, err = tr.GetTool("search")
	if err == nil {
		t.Errorf("Expected error when getting unregistered tool")
	}
}

func TestAgentCreation(t *testing.T) {
	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	agent := NewAgent("agent-1", "Test Agent", prompt)

	if agent.ID != "agent-1" {
		t.Errorf("Expected agent-1")
	}

	if agent.Name != "Test Agent" {
		t.Errorf("Expected Test Agent")
	}
}

func TestAgentAddTool(t *testing.T) {
	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	agent := NewAgent("agent-1", "Test Agent", prompt)

	tool := &ToolDefinition{
		Name:        "search",
		Description: "Search for info",
	}

	agent.AddTool(tool)

	if len(agent.Tools) != 1 {
		t.Errorf("Expected 1 tool")
	}
}

func TestAgentManagerCreateAgent(t *testing.T) {
	spm := NewSystemPromptManager()
	tr := NewToolRegistry()
	am := NewAgentManager(spm, tr)

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	spm.AddPrompt(prompt)

	agent, err := am.CreateAgent("agent-1", "Test Agent", "prompt-1")
	if err != nil {
		t.Errorf("CreateAgent failed: %v", err)
	}

	if agent.ID != "agent-1" {
		t.Errorf("Expected agent-1")
	}
}

func TestAgentManagerGetAgent(t *testing.T) {
	spm := NewSystemPromptManager()
	tr := NewToolRegistry()
	am := NewAgentManager(spm, tr)

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	spm.AddPrompt(prompt)
	am.CreateAgent("agent-1", "Test Agent", "prompt-1")

	agent, err := am.GetAgent("agent-1")
	if err != nil {
		t.Errorf("GetAgent failed: %v", err)
	}

	if agent.Name != "Test Agent" {
		t.Errorf("Expected Test Agent")
	}
}

func TestAgentManagerBindTool(t *testing.T) {
	spm := NewSystemPromptManager()
	tr := NewToolRegistry()
	am := NewAgentManager(spm, tr)

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	spm.AddPrompt(prompt)
	am.CreateAgent("agent-1", "Test Agent", "prompt-1")

	toolDef := &ToolDefinition{
		Name: "search",
	}

	handler := func(args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	tr.RegisterTool(toolDef, handler)

	err := am.BindTool("agent-1", "search")
	if err != nil {
		t.Errorf("BindTool failed: %v", err)
	}

	agent, _ := am.GetAgent("agent-1")
	if len(agent.Tools) != 1 {
		t.Errorf("Expected 1 tool after binding")
	}
}

func TestAgentManagerExecuteToolCall(t *testing.T) {
	spm := NewSystemPromptManager()
	tr := NewToolRegistry()
	am := NewAgentManager(spm, tr)

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	spm.AddPrompt(prompt)
	am.CreateAgent("agent-1", "Test Agent", "prompt-1")

	toolDef := &ToolDefinition{
		Name: "add",
	}

	handler := func(args map[string]interface{}) (interface{}, error) {
		return 42, nil
	}

	tr.RegisterTool(toolDef, handler)
	am.BindTool("agent-1", "add")

	call, err := am.ExecuteToolCall("agent-1", "add", map[string]interface{}{})
	if err != nil {
		t.Errorf("ExecuteToolCall failed: %v", err)
	}

	if call.ToolName != "add" {
		t.Errorf("Expected add tool")
	}

	if call.Result != 42 {
		t.Errorf("Expected result 42")
	}
}

func TestAgentManagerHotUpdatePrompt(t *testing.T) {
	spm := NewSystemPromptManager()
	tr := NewToolRegistry()
	am := NewAgentManager(spm, tr)

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "Old content",
	}

	spm.AddPrompt(prompt)
	am.CreateAgent("agent-1", "Test Agent", "prompt-1")

	err := am.HotUpdatePrompt("agent-1", "New content")
	if err != nil {
		t.Errorf("HotUpdatePrompt failed: %v", err)
	}

	agent, _ := am.GetAgent("agent-1")
	if agent.SystemPrompt.Content != "New content" {
		t.Errorf("Expected 'New content'")
	}
}

func TestAgentRecordToolCall(t *testing.T) {
	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	agent := NewAgent("agent-1", "Test Agent", prompt)

	call := &ToolCall{
		ID:       "call-1",
		ToolName: "search",
		Status:   "completed",
	}

	agent.RecordToolCall(call)

	calls := agent.GetToolCalls()
	if len(calls) != 1 {
		t.Errorf("Expected 1 tool call")
	}
}

func TestAgentManagerGetAgentStatistics(t *testing.T) {
	spm := NewSystemPromptManager()
	tr := NewToolRegistry()
	am := NewAgentManager(spm, tr)

	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	spm.AddPrompt(prompt)
	am.CreateAgent("agent-1", "Test Agent", "prompt-1")

	stats, err := am.GetAgentStatistics("agent-1")
	if err != nil {
		t.Errorf("GetAgentStatistics failed: %v", err)
	}

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if agentID, ok := stats["agent_id"].(string); !ok || agentID != "agent-1" {
		t.Errorf("Expected agent-1 in statistics")
	}
}

func BenchmarkSystemPromptManagerAddPrompt(b *testing.B) {
	spm := NewSystemPromptManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prompt := &SystemPrompt{
			ID:      "prompt-" + string(rune(i)),
			Content: "Test content",
		}
		_ = spm.AddPrompt(prompt)
	}
}

func BenchmarkToolCallExecution(b *testing.B) {
	tr := NewToolRegistry()

	toolDef := &ToolDefinition{
		Name: "bench-tool",
	}

	handler := func(args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	tr.RegisterTool(toolDef, handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tr.CallTool("bench-tool", map[string]interface{}{})
	}
}

func BenchmarkAgentCreation(b *testing.B) {
	prompt := &SystemPrompt{
		ID:      "prompt-1",
		Content: "You are helpful",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAgent("agent-"+string(rune(i)), "Test", prompt)
	}
}

