package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PluginManagerAPI 插件管理 API
type PluginManagerAPI struct {
	mcpManager   *MCPPluginManager
	funcEngine   *FunctionEngine
	builtinTools map[string]*Tool
	plugins      map[string]*ManagedPlugin
	mu           sync.RWMutex
	logFunc      func(level, msg string, args ...interface{})
}

// ManagedPlugin 受管理的插件
type ManagedPlugin struct {
	// 插件 ID
	ID string `json:"id"`

	// 插件名称
	Name string `json:"name"`

	// 插件描述
	Description string `json:"description"`

	// 插件类型
	Type string `json:"type"` // "builtin", "mcp"

	// 是否启用
	Enabled bool `json:"enabled"`

	// 配置
	Config map[string]interface{} `json:"config"`

	// 状态
	Status string `json:"status"` // "running", "stopped", "error"

	// 工具列表
	Tools []string `json:"tools"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 版本
	Version string `json:"version"`

	// 统计信息
	Stats PluginStats `json:"stats"`
}

// PluginStats 插件统计
type PluginStats struct {
	// 总调用数
	TotalCalls int64 `json:"total_calls"`

	// 成功调用数
	SuccessCalls int64 `json:"success_calls"`

	// 失败调用数
	FailedCalls int64 `json:"failed_calls"`

	// 平均执行时间（毫秒）
	AvgDuration int64 `json:"avg_duration"`

	// 最后调用时间
	LastCallTime time.Time `json:"last_call_time"`
}

// PluginInstallRequest 插件安装请求
type PluginInstallRequest struct {
	// 插件名称
	Name string `json:"name"`

	// 插件描述
	Description string `json:"description"`

	// 插件模式
	Mode string `json:"mode"` // "stdio" 或 "http"

	// 命令（stdio 模式）
	Command string `json:"command,omitempty"`

	// 命令参数（stdio 模式）
	Args []string `json:"args,omitempty"`

	// 服务地址（HTTP 模式）
	URL string `json:"url,omitempty"`

	// 配置
	Config map[string]interface{} `json:"config,omitempty"`

	// 自动重启
	AutoRestart bool `json:"auto_restart"`
}

// PluginConfigRequest 插件配置请求
type PluginConfigRequest struct {
	// 配置数据
	Config map[string]interface{} `json:"config"`
}

// PluginEnableRequest 插件启用请求
type PluginEnableRequest struct {
	// 是否启用
	Enabled bool `json:"enabled"`
}

// NewPluginManagerAPI 创建插件管理 API
func NewPluginManagerAPI(mcpManager *MCPPluginManager, funcEngine *FunctionEngine) *PluginManagerAPI {
	return &PluginManagerAPI{
		mcpManager:   mcpManager,
		funcEngine:   funcEngine,
		builtinTools: make(map[string]*Tool),
		plugins:      make(map[string]*ManagedPlugin),
		logFunc:      defaultLogFuncMCP,
	}
}

// RegisterBuiltinTool 注册内置工具
func (pma *PluginManagerAPI) RegisterBuiltinTool(name string, tool *Tool) error {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	pma.builtinTools[name] = tool

	// 添加到 Function Engine
	if err := pma.funcEngine.RegisterTool(tool); err != nil {
		return err
	}

	pma.logFunc("info", fmt.Sprintf("Registered builtin tool: %s", name))

	return nil
}

// InstallPlugin 安装插件
func (pma *PluginManagerAPI) InstallPlugin(ctx context.Context, req *PluginInstallRequest) (*ManagedPlugin, error) {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	// 创建 MCP 配置
	config := &MCPPluginConfig{
		Name:        req.Name,
		Description: req.Description,
		Mode:        req.Mode,
		Command:     req.Command,
		Args:        req.Args,
		URL:         req.URL,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		AutoRestart: req.AutoRestart,
	}

	// 注册到 MCP 管理器
	plugin, err := pma.mcpManager.RegisterPlugin(req.Name, config)
	if err != nil {
		return nil, fmt.Errorf("failed to register plugin: %v", err)
	}

	// 启动插件
	if err := plugin.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start plugin: %v", err)
	}

	// 创建受管插件
	managed := &ManagedPlugin{
		ID:          generatePluginID(),
		Name:        req.Name,
		Description: req.Description,
		Type:        "mcp",
		Enabled:     true,
		Config:      req.Config,
		Status:      "running",
		Tools:       getToolNames(plugin.GetTools()),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     "1.0.0",
	}

	pma.plugins[req.Name] = managed

	pma.logFunc("info", fmt.Sprintf("Installed plugin: %s", req.Name))

	return managed, nil
}

// UninstallPlugin 卸载插件
func (pma *PluginManagerAPI) UninstallPlugin(name string) error {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	_, exists := pma.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// 从 MCP 管理器卸载
	if err := pma.mcpManager.UnregisterPlugin(name); err != nil {
		return err
	}

	delete(pma.plugins, name)

	pma.logFunc("info", fmt.Sprintf("Uninstalled plugin: %s", name))

	return nil
}

// EnablePlugin 启用插件
func (pma *PluginManagerAPI) EnablePlugin(ctx context.Context, name string) error {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	plugin, exists := pma.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if plugin.Enabled {
		return fmt.Errorf("plugin already enabled: %s", name)
	}

	// 启动插件
	mcpPlugin, err := pma.mcpManager.GetPlugin(name)
	if err != nil {
		return err
	}

	if err := mcpPlugin.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin: %v", err)
	}

	plugin.Enabled = true
	plugin.Status = "running"
	plugin.UpdatedAt = time.Now()

	pma.logFunc("info", fmt.Sprintf("Enabled plugin: %s", name))

	return nil
}

// DisablePlugin 禁用插件
func (pma *PluginManagerAPI) DisablePlugin(name string) error {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	plugin, exists := pma.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if !plugin.Enabled {
		return fmt.Errorf("plugin already disabled: %s", name)
	}

	// 停止插件
	if err := pma.mcpManager.StopPlugin(name); err != nil {
		return err
	}

	plugin.Enabled = false
	plugin.Status = "stopped"
	plugin.UpdatedAt = time.Now()

	pma.logFunc("info", fmt.Sprintf("Disabled plugin: %s", name))

	return nil
}

// UpdatePluginConfig 更新插件配置
func (pma *PluginManagerAPI) UpdatePluginConfig(name string, req *PluginConfigRequest) error {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	plugin, exists := pma.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	plugin.Config = req.Config
	plugin.UpdatedAt = time.Now()

	pma.logFunc("info", fmt.Sprintf("Updated config for plugin: %s", name))

	return nil
}

// GetPlugin 获取插件信息
func (pma *PluginManagerAPI) GetPlugin(name string) (*ManagedPlugin, error) {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	plugin, exists := pma.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (pma *PluginManagerAPI) ListPlugins() []*ManagedPlugin {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	plugins := make([]*ManagedPlugin, 0)
	for _, plugin := range pma.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetPluginStatus 获取插件状态
func (pma *PluginManagerAPI) GetPluginStatus(name string) (*ManagedPlugin, error) {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	managedPlugin, exists := pma.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	// 获取 MCP 插件状态
	mcpPlugin, err := pma.mcpManager.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	status := mcpPlugin.GetStatus()
	managedPlugin.Status = status.State
	managedPlugin.Stats.TotalCalls = status.TotalCalls
	managedPlugin.Stats.SuccessCalls = status.SuccessCalls
	managedPlugin.Stats.FailedCalls = status.FailedCalls

	return managedPlugin, nil
}

// GetPluginTools 获取插件工具列表
func (pma *PluginManagerAPI) GetPluginTools(name string) ([]interface{}, error) {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	_, exists := pma.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	// 获取 MCP 插件工具
	mcpPlugin, err := pma.mcpManager.GetPlugin(name)
	if err != nil {
		return nil, err
	}

	tools := mcpPlugin.GetTools()
	result := make([]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		}
	}

	return result, nil
}

// CallPluginTool 调用插件工具
func (pma *PluginManagerAPI) CallPluginTool(ctx context.Context, pluginName, toolName string, args map[string]interface{}) (interface{}, error) {
	pma.mu.RLock()
	plugin, exists := pma.plugins[pluginName]
	pma.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	if !plugin.Enabled {
		return nil, fmt.Errorf("plugin is disabled: %s", pluginName)
	}

	// 调用工具
	result, err := pma.mcpManager.CallTool(ctx, pluginName, toolName, args)
	if err != nil {
		return nil, err
	}

	// 更新统计
	pma.mu.Lock()
	plugin.Stats.TotalCalls++
	plugin.Stats.SuccessCalls++
	plugin.Stats.LastCallTime = time.Now()
	pma.mu.Unlock()

	return result, nil
}

// GetAllPluginsStatus 获取所有插件状态
func (pma *PluginManagerAPI) GetAllPluginsStatus() map[string]interface{} {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	status := map[string]interface{}{
		"total_plugins":   len(pma.plugins),
		"enabled_plugins": 0,
		"plugins":         make([]*ManagedPlugin, 0),
	}

	enabledCount := 0
	for _, plugin := range pma.plugins {
		if plugin.Enabled {
			enabledCount++
		}
		status["plugins"] = append(status["plugins"].([]*ManagedPlugin), plugin)
	}

	status["enabled_plugins"] = enabledCount

	return status
}

// ReloadPlugin 重新加载插件
func (pma *PluginManagerAPI) ReloadPlugin(ctx context.Context, name string) error {
	pma.mu.Lock()
	defer pma.mu.Unlock()

	managedPlugin, exists := pma.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// 停止插件
	if err := pma.mcpManager.StopPlugin(name); err != nil {
		return err
	}

	// 启动插件
	if err := pma.mcpManager.StartPlugin(ctx, name); err != nil {
		return err
	}

	managedPlugin.Status = "running"
	managedPlugin.UpdatedAt = time.Now()

	pma.logFunc("info", fmt.Sprintf("Reloaded plugin: %s", name))

	return nil
}

// ExportPluginConfig 导出插件配置
func (pma *PluginManagerAPI) ExportPluginConfig(name string) (string, error) {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	plugin, exists := pma.plugins[name]
	if !exists {
		return "", fmt.Errorf("plugin not found: %s", name)
	}

	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %v", err)
	}

	return string(data), nil
}

// ImportPluginConfig 导入插件配置
func (pma *PluginManagerAPI) ImportPluginConfig(data string) (*ManagedPlugin, error) {
	var plugin ManagedPlugin

	if err := json.Unmarshal([]byte(data), &plugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	pma.mu.Lock()
	pma.plugins[plugin.Name] = &plugin
	pma.mu.Unlock()

	pma.logFunc("info", fmt.Sprintf("Imported plugin config: %s", plugin.Name))

	return &plugin, nil
}

// GetPluginMetrics 获取插件指标
func (pma *PluginManagerAPI) GetPluginMetrics(name string) (map[string]interface{}, error) {
	pma.mu.RLock()
	defer pma.mu.RUnlock()

	plugin, exists := pma.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	metrics := map[string]interface{}{
		"name":           plugin.Name,
		"enabled":        plugin.Enabled,
		"status":         plugin.Status,
		"total_calls":    plugin.Stats.TotalCalls,
		"success_calls":  plugin.Stats.SuccessCalls,
		"failed_calls":   plugin.Stats.FailedCalls,
		"success_rate":   float64(0),
		"avg_duration":   plugin.Stats.AvgDuration,
		"last_call_time": plugin.Stats.LastCallTime,
	}

	if plugin.Stats.TotalCalls > 0 {
		metrics["success_rate"] = float64(plugin.Stats.SuccessCalls) / float64(plugin.Stats.TotalCalls)
	}

	return metrics, nil
}

// generatePluginID 生成插件 ID
func generatePluginID() string {
	return fmt.Sprintf("plugin-%d", time.Now().UnixNano())
}

// getToolNames 获取工具名称列表
func getToolNames(tools []*MCPTool) []string {
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}
