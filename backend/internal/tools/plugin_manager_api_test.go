package tools

import (
	"context"
	"testing"
	"time"
)

func TestNewPluginManagerAPI(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	if api == nil {
		t.Errorf("Expected API to be created")
	}
}

func TestPluginManagerAPIRegisterBuiltinTool(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	tool := &Tool{
		Name:        "test_tool",
		Description: "Test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	err := api.RegisterBuiltinTool("test_tool", tool)
	if err != nil {
		t.Errorf("RegisterBuiltinTool failed: %v", err)
	}

	builtins := api.builtinTools
	if len(builtins) == 0 {
		t.Errorf("Expected builtin tool to be registered")
	}
}

func TestManagedPlugin(t *testing.T) {
	plugin := &ManagedPlugin{
		ID:          "plugin-123",
		Name:        "test",
		Description: "Test plugin",
		Type:        "mcp",
		Enabled:     true,
		Status:      "running",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if plugin.Name != "test" {
		t.Errorf("Expected plugin name test")
	}

	if !plugin.Enabled {
		t.Errorf("Expected plugin to be enabled")
	}
}

func TestPluginManagerAPIListPlugins(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	plugins := api.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins initially")
	}
}

func TestPluginManagerAPIGetPlugin(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	_, err := api.GetPlugin("nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent plugin")
	}
}

func TestPluginInstallRequest(t *testing.T) {
	req := &PluginInstallRequest{
		Name:        "test",
		Description: "Test plugin",
		Mode:        "http",
		URL:         "http://localhost:8080",
		Config: map[string]interface{}{
			"key": "value",
		},
	}

	if req.Name != "test" {
		t.Errorf("Expected name test")
	}

	if req.Mode != "http" {
		t.Errorf("Expected mode http")
	}
}

func TestPluginConfigRequest(t *testing.T) {
	req := &PluginConfigRequest{
		Config: map[string]interface{}{
			"setting1": "value1",
			"setting2": 42,
		},
	}

	if len(req.Config) != 2 {
		t.Errorf("Expected 2 config items")
	}
}

func TestPluginEnableRequest(t *testing.T) {
	req := &PluginEnableRequest{
		Enabled: true,
	}

	if !req.Enabled {
		t.Errorf("Expected enabled to be true")
	}
}

func TestPluginStats(t *testing.T) {
	stats := PluginStats{
		TotalCalls:   100,
		SuccessCalls: 95,
		FailedCalls:  5,
		AvgDuration:  50,
	}

	if stats.TotalCalls != 100 {
		t.Errorf("Expected 100 total calls")
	}

	if stats.SuccessCalls != 95 {
		t.Errorf("Expected 95 success calls")
	}
}

func TestPluginManagerAPIGetAllPluginsStatus(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	status := api.GetAllPluginsStatus()

	if status["total_plugins"] != 0 {
		t.Errorf("Expected 0 total plugins")
	}

	if status["enabled_plugins"] != 0 {
		t.Errorf("Expected 0 enabled plugins")
	}
}

func TestPluginManagerAPIExportPluginConfig(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	_, err := api.ExportPluginConfig("nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent plugin")
	}
}

func TestPluginManagerAPIImportPluginConfig(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	data := `{
		"id": "plugin-123",
		"name": "test",
		"description": "Test plugin",
		"type": "mcp",
		"enabled": true,
		"status": "running"
	}`

	plugin, err := api.ImportPluginConfig(data)
	if err != nil {
		t.Errorf("ImportPluginConfig failed: %v", err)
	}

	if plugin.Name != "test" {
		t.Errorf("Expected plugin name test")
	}
}

func TestPluginManagerAPIGetPluginMetrics(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	_, err := api.GetPluginMetrics("nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent plugin")
	}
}

func TestPluginManagerAPIUpdatePluginConfig(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	// Add a plugin manually
	plugin := &ManagedPlugin{
		ID:      "plugin-123",
		Name:    "test",
		Enabled: true,
	}
	api.plugins["test"] = plugin

	req := &PluginConfigRequest{
		Config: map[string]interface{}{
			"new_key": "new_value",
		},
	}

	err := api.UpdatePluginConfig("test", req)
	if err != nil {
		t.Errorf("UpdatePluginConfig failed: %v", err)
	}

	if len(plugin.Config) == 0 {
		t.Errorf("Expected config to be updated")
	}
}

func TestGeneratePluginID(t *testing.T) {
	id1 := generatePluginID()
	id2 := generatePluginID()

	if id1 == id2 {
		t.Errorf("Expected different plugin IDs")
	}

	if len(id1) == 0 {
		t.Errorf("Expected non-empty plugin ID")
	}
}

func TestGetToolNames(t *testing.T) {
	tools := []*MCPTool{
		{Name: "tool1"},
		{Name: "tool2"},
		{Name: "tool3"},
	}

	names := getToolNames(tools)

	if len(names) != 3 {
		t.Errorf("Expected 3 tool names")
	}

	if names[0] != "tool1" {
		t.Errorf("Expected tool1")
	}
}

func TestManagedPluginFields(t *testing.T) {
	plugin := &ManagedPlugin{
		ID:          "id-123",
		Name:        "plugin",
		Description: "desc",
		Type:        "mcp",
		Enabled:     true,
		Config:      map[string]interface{}{},
		Status:      "running",
		Tools:       []string{"tool1", "tool2"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     "1.0.0",
	}

	if plugin.ID != "id-123" {
		t.Errorf("Expected ID id-123")
	}

	if len(plugin.Tools) != 2 {
		t.Errorf("Expected 2 tools")
	}
}

func TestPluginManagerAPIDisablePlugin(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	err := api.DisablePlugin("nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent plugin")
	}
}

func TestPluginManagerAPIConcurrency(t *testing.T) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	// Test concurrent access
	for i := 0; i < 10; i++ {
		go func() {
			api.ListPlugins()
		}()
	}

	time.Sleep(100 * time.Millisecond)
}

func BenchmarkPluginManagerAPIListPlugins(b *testing.B) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	// Add some plugins
	for i := 0; i < 10; i++ {
		plugin := &ManagedPlugin{
			ID:      generatePluginID(),
			Name:    string(rune(i)),
			Enabled: true,
		}
		api.plugins[string(rune(i))] = plugin
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api.ListPlugins()
	}
}

func BenchmarkPluginManagerAPIGetPlugin(b *testing.B) {
	mcpManager := NewMCPPluginManager()
	funcEngine := NewFunctionEngine()
	api := NewPluginManagerAPI(mcpManager, funcEngine)

	// Add a plugin
	plugin := &ManagedPlugin{
		ID:      "plugin-123",
		Name:    "test",
		Enabled: true,
	}
	api.plugins["test"] = plugin

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api.GetPlugin("test")
	}
}

