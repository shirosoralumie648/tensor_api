package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MCPMessage MCP 协议消息
type MCPMessage struct {
	// 消息类型
	Type string `json:"type"`

	// 消息 ID
	ID string `json:"id,omitempty"`

	// 工具定义
	Tool *MCPTool `json:"tool,omitempty"`

	// 工具调用
	ToolCall *MCPToolCall `json:"tool_call,omitempty"`

	// 结果
	Result interface{} `json:"result,omitempty"`

	// 错误信息
	Error string `json:"error,omitempty"`

	// 健康检查
	Status string `json:"status,omitempty"`
}

// MCPTool MCP 工具定义
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// MCPToolCall MCP 工具调用
type MCPToolCall struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPPluginConfig MCP 插件配置
type MCPPluginConfig struct {
	// 插件名称
	Name string

	// 插件描述
	Description string

	// 运行模式
	Mode string // "stdio" 或 "http"

	// 启动命令（stdio 模式）
	Command string
	Args    []string

	// 服务地址（http 模式）
	URL string

	// 超时时间
	Timeout time.Duration

	// 最大重试次数
	MaxRetries int

	// 自动恢复
	AutoRestart bool
}

// MCPPluginStatus 插件状态
type MCPPluginStatus struct {
	// 插件名称
	Name string

	// 状态
	State string // "running", "stopped", "error", "crashed"

	// 错误信息
	Error string

	// 启动时间
	StartedAt time.Time

	// 最后心跳
	LastHeartbeat time.Time

	// 工具数
	ToolCount int

	// 调用统计
	TotalCalls int64
	SuccessCalls int64
	FailedCalls int64
}

// StdioMCPPlugin stdio 模式 MCP 插件
type StdioMCPPlugin struct {
	config     *MCPPluginConfig
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	status     MCPPluginStatus
	tools      map[string]*MCPTool
	toolsMu    sync.RWMutex
	running    int32
	exitChan   chan error
	logFunc    func(level, msg string, args ...interface{})
}

// NewStdioMCPPlugin 创建 stdio 模式插件
func NewStdioMCPPlugin(config *MCPPluginConfig) *StdioMCPPlugin {
	return &StdioMCPPlugin{
		config:   config,
		tools:    make(map[string]*MCPTool),
		exitChan: make(chan error, 1),
		status: MCPPluginStatus{
			Name:      config.Name,
			State:     "stopped",
			StartedAt: time.Now(),
		},
		logFunc: defaultLogFuncMCP,
	}
}

// defaultLogFuncMCP 默认日志函数
func defaultLogFuncMCP(level, msg string, args ...interface{}) {
	// 默认实现：忽略日志
}

// Start 启动插件
func (smp *StdioMCPPlugin) Start(ctx context.Context) error {
	if atomic.LoadInt32(&smp.running) == 1 {
		return fmt.Errorf("plugin already running")
	}

	// 创建命令
	smp.cmd = exec.CommandContext(ctx, smp.config.Command, smp.config.Args...)

	// 获取 stdin/stdout
	stdin, err := smp.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	stdout, err := smp.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	smp.stdin = stdin
	smp.stdout = stdout

	// 启动进程
	if err := smp.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start plugin: %v", err)
	}

	atomic.StoreInt32(&smp.running, 1)
	smp.status.State = "running"
	smp.status.StartedAt = time.Now()
	smp.status.LastHeartbeat = time.Now()

	smp.logFunc("info", fmt.Sprintf("Started plugin: %s", smp.config.Name))

	// 后台监听输出
	go smp.readOutput()

	// 后台监听进程退出
	go smp.waitForExit()

	// 启动心跳检查
	go smp.healthCheck(ctx)

	return nil
}

// Stop 停止插件
func (smp *StdioMCPPlugin) Stop() error {
	if atomic.LoadInt32(&smp.running) == 0 {
		return fmt.Errorf("plugin not running")
	}

	atomic.StoreInt32(&smp.running, 0)

	if smp.stdin != nil {
		smp.stdin.Close()
	}

	if smp.cmd != nil && smp.cmd.Process != nil {
		smp.cmd.Process.Kill()
	}

	smp.status.State = "stopped"

	smp.logFunc("info", fmt.Sprintf("Stopped plugin: %s", smp.config.Name))

	return nil
}

// Call 调用工具
func (smp *StdioMCPPlugin) Call(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	if atomic.LoadInt32(&smp.running) == 0 {
		return nil, fmt.Errorf("plugin not running")
	}

	// 构建请求
	msg := MCPMessage{
		Type: "tool_call",
		ID:   fmt.Sprintf("call-%d", time.Now().UnixNano()),
		ToolCall: &MCPToolCall{
			ToolName:  toolName,
			Arguments: args,
		},
	}

	// 发送请求
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	if _, err := smp.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// 等待响应
	responseChan := make(chan *MCPMessage, 1)
	go smp.readResponse(msg.ID, responseChan)

	select {
	case response := <-responseChan:
		if response.Error != "" {
			atomic.AddInt64(&smp.status.FailedCalls, 1)
			return nil, fmt.Errorf(response.Error)
		}
		atomic.AddInt64(&smp.status.SuccessCalls, 1)
		return response.Result, nil

	case <-ctx.Done():
		atomic.AddInt64(&smp.status.FailedCalls, 1)
		return nil, ctx.Err()
	}
}

// GetTools 获取工具列表
func (smp *StdioMCPPlugin) GetTools() []*MCPTool {
	smp.toolsMu.RLock()
	defer smp.toolsMu.RUnlock()

	tools := make([]*MCPTool, 0)
	for _, tool := range smp.tools {
		tools = append(tools, tool)
	}

	return tools
}

// GetStatus 获取插件状态
func (smp *StdioMCPPlugin) GetStatus() MCPPluginStatus {
	smp.toolsMu.RLock()
	defer smp.toolsMu.RUnlock()

	smp.status.ToolCount = len(smp.tools)
	smp.status.TotalCalls = smp.status.SuccessCalls + smp.status.FailedCalls

	return smp.status
}

// readOutput 读取输出
func (smp *StdioMCPPlugin) readOutput() {
	scanner := bufio.NewScanner(smp.stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var msg MCPMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			smp.logFunc("error", fmt.Sprintf("Failed to unmarshal message: %v", err))
			continue
		}

		// 处理不同类型的消息
		switch msg.Type {
		case "tool_definition":
			smp.toolsMu.Lock()
			smp.tools[msg.Tool.Name] = msg.Tool
			smp.toolsMu.Unlock()

		case "heartbeat":
			smp.status.LastHeartbeat = time.Now()
		}
	}

	if err := scanner.Err(); err != nil {
		smp.logFunc("error", fmt.Sprintf("Scanner error: %v", err))
	}
}

// readResponse 读取响应
func (smp *StdioMCPPlugin) readResponse(callID string, responseChan chan<- *MCPMessage) {
	scanner := bufio.NewScanner(smp.stdout)

	for scanner.Scan() {
		var msg MCPMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		if msg.ID == callID {
			responseChan <- &msg
			return
		}
	}
}

// waitForExit 等待进程退出
func (smp *StdioMCPPlugin) waitForExit() {
	if err := smp.cmd.Wait(); err != nil {
		smp.status.State = "crashed"
		smp.status.Error = err.Error()
		smp.logFunc("error", fmt.Sprintf("Plugin crashed: %v", err))

		// 自动恢复
		if smp.config.AutoRestart {
			time.Sleep(time.Second)
			if err := smp.Start(context.Background()); err != nil {
				smp.logFunc("error", fmt.Sprintf("Failed to restart plugin: %v", err))
			}
		}
	} else {
		smp.status.State = "stopped"
	}

	atomic.StoreInt32(&smp.running, 0)
}

// healthCheck 健康检查
func (smp *StdioMCPPlugin) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送心跳
			msg := MCPMessage{
				Type:   "heartbeat",
				ID:     fmt.Sprintf("hb-%d", time.Now().UnixNano()),
				Status: "ping",
			}

			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			if _, err := smp.stdin.Write(append(data, '\n')); err != nil {
				smp.logFunc("warn", "Heartbeat failed, plugin may be unresponsive")
			}

			// 检查最后心跳时间
			if time.Since(smp.status.LastHeartbeat) > 2*time.Minute {
				smp.logFunc("error", "Plugin heartbeat timeout")
				smp.status.State = "error"
			}

		case <-ctx.Done():
			return
		}
	}
}

// HTTPMCPPlugin HTTP 模式 MCP 插件
type HTTPMCPPlugin struct {
	config     *MCPPluginConfig
	client     *http.Client
	status     MCPPluginStatus
	tools      map[string]*MCPTool
	toolsMu    sync.RWMutex
	logFunc    func(level, msg string, args ...interface{})
}

// NewHTTPMCPPlugin 创建 HTTP 模式插件
func NewHTTPMCPPlugin(config *MCPPluginConfig) *HTTPMCPPlugin {
	return &HTTPMCPPlugin{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		tools: make(map[string]*MCPTool),
		status: MCPPluginStatus{
			Name:      config.Name,
			State:     "stopped",
			StartedAt: time.Now(),
		},
		logFunc: defaultLogFuncMCP,
	}
}

// Start 启动插件（HTTP 模式只进行连接检查）
func (hmp *HTTPMCPPlugin) Start(ctx context.Context) error {
	// 检查服务是否可达
	resp, err := hmp.client.Get(hmp.config.URL + "/health")
	if err != nil {
		hmp.status.State = "error"
		hmp.status.Error = err.Error()
		return fmt.Errorf("failed to connect to plugin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hmp.status.State = "error"
		return fmt.Errorf("plugin returned status %d", resp.StatusCode)
	}

	hmp.status.State = "running"
	hmp.status.StartedAt = time.Now()
	hmp.status.LastHeartbeat = time.Now()

	hmp.logFunc("info", fmt.Sprintf("Connected to HTTP plugin: %s", hmp.config.Name))

	// 启动健康检查
	go hmp.healthCheck(ctx)

	return nil
}

// Stop 停止插件
func (hmp *HTTPMCPPlugin) Stop() error {
	hmp.status.State = "stopped"
	hmp.logFunc("info", fmt.Sprintf("Disconnected from HTTP plugin: %s", hmp.config.Name))
	return nil
}

// Call 调用工具
func (hmp *HTTPMCPPlugin) Call(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	msg := MCPMessage{
		Type: "tool_call",
		ID:   fmt.Sprintf("call-%d", time.Now().UnixNano()),
		ToolCall: &MCPToolCall{
			ToolName:  toolName,
			Arguments: args,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", hmp.config.URL+"/call", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := hmp.client.Do(req)
	if err != nil {
		atomic.AddInt64(&hmp.status.FailedCalls, 1)
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var response MCPMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		atomic.AddInt64(&hmp.status.FailedCalls, 1)
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != "" {
		atomic.AddInt64(&hmp.status.FailedCalls, 1)
		return nil, fmt.Errorf(response.Error)
	}

	atomic.AddInt64(&hmp.status.SuccessCalls, 1)
	return response.Result, nil
}

// GetTools 获取工具列表
func (hmp *HTTPMCPPlugin) GetTools() []*MCPTool {
	hmp.toolsMu.RLock()
	defer hmp.toolsMu.RUnlock()

	tools := make([]*MCPTool, 0)
	for _, tool := range hmp.tools {
		tools = append(tools, tool)
	}

	return tools
}

// GetStatus 获取插件状态
func (hmp *HTTPMCPPlugin) GetStatus() MCPPluginStatus {
	hmp.toolsMu.RLock()
	defer hmp.toolsMu.RUnlock()

	hmp.status.ToolCount = len(hmp.tools)
	hmp.status.TotalCalls = hmp.status.SuccessCalls + hmp.status.FailedCalls

	return hmp.status
}

// healthCheck 健康检查
func (hmp *HTTPMCPPlugin) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := hmp.client.Get(hmp.config.URL + "/health")
			if err != nil {
				hmp.status.State = "error"
				hmp.status.Error = err.Error()
				hmp.logFunc("warn", fmt.Sprintf("Health check failed: %v", err))
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				hmp.status.State = "running"
				hmp.status.LastHeartbeat = time.Now()
			} else {
				hmp.status.State = "error"
				hmp.status.Error = fmt.Sprintf("health check returned %d", resp.StatusCode)
			}

		case <-ctx.Done():
			return
		}
	}
}

// MCPPlugin MCPPlugin 接口
type MCPPlugin interface {
	Start(ctx context.Context) error
	Stop() error
	Call(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error)
	GetTools() []*MCPTool
	GetStatus() MCPPluginStatus
}

// MCPPluginManager MCP 插件管理器
type MCPPluginManager struct {
	plugins map[string]MCPPlugin
	mu      sync.RWMutex
	logFunc func(level, msg string, args ...interface{})
}

// NewMCPPluginManager 创建 MCP 插件管理器
func NewMCPPluginManager() *MCPPluginManager {
	return &MCPPluginManager{
		plugins: make(map[string]MCPPlugin),
		logFunc: defaultLogFuncMCP,
	}
}

// RegisterPlugin 注册插件
func (mpm *MCPPluginManager) RegisterPlugin(name string, config *MCPPluginConfig) (MCPPlugin, error) {
	mpm.mu.Lock()
	defer mpm.mu.Unlock()

	var plugin MCPPlugin

	switch config.Mode {
	case "stdio":
		plugin = NewStdioMCPPlugin(config)
	case "http":
		plugin = NewHTTPMCPPlugin(config)
	default:
		return nil, fmt.Errorf("unsupported mode: %s", config.Mode)
	}

	mpm.plugins[name] = plugin
	mpm.logFunc("info", fmt.Sprintf("Registered plugin: %s (%s)", name, config.Mode))

	return plugin, nil
}

// UnregisterPlugin 注销插件
func (mpm *MCPPluginManager) UnregisterPlugin(name string) error {
	mpm.mu.Lock()
	defer mpm.mu.Unlock()

	plugin, exists := mpm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}

	if err := plugin.Stop(); err != nil {
		return err
	}

	delete(mpm.plugins, name)
	mpm.logFunc("info", fmt.Sprintf("Unregistered plugin: %s", name))

	return nil
}

// GetPlugin 获取插件
func (mpm *MCPPluginManager) GetPlugin(name string) (MCPPlugin, error) {
	mpm.mu.RLock()
	defer mpm.mu.RUnlock()

	plugin, exists := mpm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}

	return plugin, nil
}

// ListPlugins 列出所有插件
func (mpm *MCPPluginManager) ListPlugins() []string {
	mpm.mu.RLock()
	defer mpm.mu.RUnlock()

	names := make([]string, 0)
	for name := range mpm.plugins {
		names = append(names, name)
	}

	return names
}

// StartPlugin 启动插件
func (mpm *MCPPluginManager) StartPlugin(ctx context.Context, name string) error {
	plugin, err := mpm.GetPlugin(name)
	if err != nil {
		return err
	}

	return plugin.Start(ctx)
}

// StopPlugin 停止插件
func (mpm *MCPPluginManager) StopPlugin(name string) error {
	plugin, err := mpm.GetPlugin(name)
	if err != nil {
		return err
	}

	return plugin.Stop()
}

// CallTool 调用工具
func (mpm *MCPPluginManager) CallTool(ctx context.Context, pluginName, toolName string, args map[string]interface{}) (interface{}, error) {
	plugin, err := mpm.GetPlugin(pluginName)
	if err != nil {
		return nil, err
	}

	return plugin.Call(ctx, toolName, args)
}

// GetAllTools 获取所有工具
func (mpm *MCPPluginManager) GetAllTools() map[string][]*MCPTool {
	mpm.mu.RLock()
	defer mpm.mu.RUnlock()

	allTools := make(map[string][]*MCPTool)

	for name, plugin := range mpm.plugins {
		allTools[name] = plugin.GetTools()
	}

	return allTools
}

// GetAllStatus 获取所有插件状态
func (mpm *MCPPluginManager) GetAllStatus() []MCPPluginStatus {
	mpm.mu.RLock()
	defer mpm.mu.RUnlock()

	statuses := make([]MCPPluginStatus, 0)

	for _, plugin := range mpm.plugins {
		statuses = append(statuses, plugin.GetStatus())
	}

	return statuses
}

