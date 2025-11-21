package tools

import (
	"context"
	"testing"
	"time"
)

func TestMCPPluginConfig(t *testing.T) {
	config := &MCPPluginConfig{
		Name:        "test_plugin",
		Description: "Test plugin",
		Mode:        "stdio",
		Command:     "echo",
		Args:        []string{"test"},
		Timeout:     10 * time.Second,
		MaxRetries:  3,
		AutoRestart: true,
	}

	if config.Name != "test_plugin" {
		t.Errorf("Expected plugin name test_plugin")
	}

	if config.Mode != "stdio" {
		t.Errorf("Expected mode stdio")
	}
}

func TestHTTPMCPPlugin(t *testing.T) {
	config := &MCPPluginConfig{
		Name:        "http_test",
		Description: "HTTP test plugin",
		Mode:        "http",
		URL:         "http://localhost:8080",
		Timeout:     5 * time.Second,
	}

	plugin := NewHTTPMCPPlugin(config)

	if plugin.config.Name != "http_test" {
		t.Errorf("Expected plugin name http_test")
	}

	status := plugin.GetStatus()
	if status.State != "stopped" {
		t.Errorf("Expected state stopped")
	}
}

func TestHTTPMCPPluginStatus(t *testing.T) {
	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	plugin := NewHTTPMCPPlugin(config)
	status := plugin.GetStatus()

	if status.Name != "test" {
		t.Errorf("Expected name test")
	}

	if status.State != "stopped" {
		t.Errorf("Expected state stopped")
	}
}

func TestMCPPluginManager(t *testing.T) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	plugin, err := manager.RegisterPlugin("test", config)
	if err != nil {
		t.Errorf("RegisterPlugin failed: %v", err)
	}

	if plugin == nil {
		t.Errorf("Expected plugin to be returned")
	}

	plugins := manager.ListPlugins()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin")
	}

	if plugins[0] != "test" {
		t.Errorf("Expected plugin name test")
	}
}

func TestMCPPluginManagerGetPlugin(t *testing.T) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	manager.RegisterPlugin("test", config)

	plugin, err := manager.GetPlugin("test")
	if err != nil {
		t.Errorf("GetPlugin failed: %v", err)
	}

	if plugin == nil {
		t.Errorf("Expected plugin to be returned")
	}
}

func TestMCPPluginManagerGetPluginNotFound(t *testing.T) {
	manager := NewMCPPluginManager()

	_, err := manager.GetPlugin("nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent plugin")
	}
}

func TestMCPPluginManagerUnregister(t *testing.T) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	manager.RegisterPlugin("test", config)

	err := manager.UnregisterPlugin("test")
	if err != nil {
		t.Errorf("UnregisterPlugin failed: %v", err)
	}

	plugins := manager.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins after unregister")
	}
}

func TestMCPMessage(t *testing.T) {
	msg := MCPMessage{
		Type: "tool_call",
		ID:   "test-123",
		ToolCall: &MCPToolCall{
			ToolName:  "test_tool",
			Arguments: map[string]interface{}{"key": "value"},
		},
	}

	if msg.Type != "tool_call" {
		t.Errorf("Expected type tool_call")
	}

	if msg.ToolCall.ToolName != "test_tool" {
		t.Errorf("Expected tool name test_tool")
	}
}

func TestMCPTool(t *testing.T) {
	tool := &MCPTool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"param1": "string",
		},
	}

	if tool.Name != "test_tool" {
		t.Errorf("Expected tool name test_tool")
	}

	if tool.Description != "A test tool" {
		t.Errorf("Expected description")
	}
}

func TestStdioMCPPluginConfig(t *testing.T) {
	config := &MCPPluginConfig{
		Name:        "stdio_plugin",
		Mode:        "stdio",
		Command:     "python",
		Args:        []string{"plugin.py"},
		Timeout:     10 * time.Second,
		MaxRetries:  3,
		AutoRestart: true,
	}

	plugin := NewStdioMCPPlugin(config)

	if plugin.config.Name != "stdio_plugin" {
		t.Errorf("Expected plugin name stdio_plugin")
	}

	status := plugin.GetStatus()
	if status.State != "stopped" {
		t.Errorf("Expected state stopped")
	}
}

func TestMCPPluginManagerMultiplePlugins(t *testing.T) {
	manager := NewMCPPluginManager()

	for i := 0; i < 3; i++ {
		config := &MCPPluginConfig{
			Name:    string(rune(i)) + "_plugin",
			Mode:    "http",
			URL:     "http://localhost:8080",
			Timeout: 5 * time.Second,
		}

		_, err := manager.RegisterPlugin(string(rune(i))+"_plugin", config)
		if err != nil {
			t.Errorf("RegisterPlugin failed: %v", err)
		}
	}

	plugins := manager.ListPlugins()
	if len(plugins) != 3 {
		t.Errorf("Expected 3 plugins")
	}
}

func TestMCPPluginManagerGetAllStatus(t *testing.T) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	manager.RegisterPlugin("test", config)

	statuses := manager.GetAllStatus()
	if len(statuses) != 1 {
		t.Errorf("Expected 1 status")
	}

	if statuses[0].Name != "test" {
		t.Errorf("Expected name test")
	}
}

func TestHTTPMCPPluginTools(t *testing.T) {
	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	plugin := NewHTTPMCPPlugin(config)

	tools := plugin.GetTools()
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools initially")
	}
}

func TestMCPPluginStatusFields(t *testing.T) {
	status := MCPPluginStatus{
		Name:      "test",
		State:     "running",
		StartedAt: time.Now(),
	}

	if status.Name != "test" {
		t.Errorf("Expected name test")
	}

	if status.State != "running" {
		t.Errorf("Expected state running")
	}
}

func BenchmarkMCPPluginManagerRegister(b *testing.B) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.RegisterPlugin("test"+string(rune(i)), config)
	}
}

func BenchmarkMCPPluginManagerGetPlugin(b *testing.B) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	manager.RegisterPlugin("test", config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetPlugin("test")
	}
}

func TestMCPPluginManagerStopPlugin(t *testing.T) {
	manager := NewMCPPluginManager()

	config := &MCPPluginConfig{
		Name:    "test",
		Mode:    "http",
		URL:     "http://localhost:8080",
		Timeout: 5 * time.Second,
	}

	manager.RegisterPlugin("test", config)

	err := manager.StopPlugin("test")
	if err != nil {
		t.Errorf("StopPlugin failed: %v", err)
	}

	status := manager.plugins["test"].GetStatus()
	if status.State != "stopped" {
		t.Errorf("Expected state stopped after stop")
	}
}

func TestMCPPluginManagerCallToolNotFound(t *testing.T) {
	manager := NewMCPPluginManager()

	ctx := context.Background()
	_, err := manager.CallTool(ctx, "nonexistent", "tool", map[string]interface{}{})

	if err == nil {
		t.Errorf("Expected error for nonexistent plugin")
	}
}

