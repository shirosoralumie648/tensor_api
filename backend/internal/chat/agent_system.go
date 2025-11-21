package chat

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SystemPrompt 系统提示词
type SystemPrompt struct {
	// 提示词 ID
	ID string `json:"id"`

	// 提示词内容
	Content string `json:"content"`

	// 提示词版本
	Version int64 `json:"version"`

	// 角色
	Role string `json:"role"`

	// 模型
	Model string `json:"model"`

	// 温度
	Temperature float64 `json:"temperature"`

	// 最大 token 数
	MaxTokens int `json:"max_tokens"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 启用状态
	Enabled bool `json:"enabled"`

	// 标签
	Tags []string `json:"tags"`
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	// 工具名称
	Name string `json:"name"`

	// 工具描述
	Description string `json:"description"`

	// 参数定义
	Parameters map[string]interface{} `json:"parameters"`

	// 返回类型
	ReturnType string `json:"return_type"`

	// 是否必需
	Required bool `json:"required"`

	// 超时时间（毫秒）
	TimeoutMS int `json:"timeout_ms"`

	// 重试次数
	RetryCount int `json:"retry_count"`
}

// ToolCall 工具调用
type ToolCall struct {
	// 调用 ID
	ID string `json:"id"`

	// 工具名称
	ToolName string `json:"tool_name"`

	// 参数
	Arguments map[string]interface{} `json:"arguments"`

	// 结果
	Result interface{} `json:"result"`

	// 错误信息
	Error string `json:"error"`

	// 执行时间（毫秒）
	ExecutionTime int64 `json:"execution_time"`

	// 状态
	Status string `json:"status"` // pending, running, completed, failed

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 完成时间
	CompletedAt *time.Time `json:"completed_at"`
}

// SystemPromptManager 系统提示词管理器
type SystemPromptManager struct {
	// 提示词存储
	prompts map[string]*SystemPrompt
	promptsMu sync.RWMutex

	// 当前版本
	currentVersion int64

	// 统计信息
	totalPrompts int64
	totalUpdates int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewSystemPromptManager 创建系统提示词管理器
func NewSystemPromptManager() *SystemPromptManager {
	return &SystemPromptManager{
		prompts: make(map[string]*SystemPrompt),
		logFunc: defaultLogFunc,
	}
}

// AddPrompt 添加提示词
func (spm *SystemPromptManager) AddPrompt(prompt *SystemPrompt) error {
	spm.promptsMu.Lock()
	defer spm.promptsMu.Unlock()

	if _, exists := spm.prompts[prompt.ID]; exists {
		return fmt.Errorf("prompt %s already exists", prompt.ID)
	}

	prompt.Version = atomic.AddInt64(&spm.currentVersion, 1)
	prompt.CreatedAt = time.Now()
	prompt.UpdatedAt = time.Now()

	spm.prompts[prompt.ID] = prompt
	atomic.AddInt64(&spm.totalPrompts, 1)

	spm.logFunc("info", fmt.Sprintf("Added prompt %s (version %d)", prompt.ID, prompt.Version))

	return nil
}

// UpdatePrompt 更新提示词
func (spm *SystemPromptManager) UpdatePrompt(promptID string, content string) (*SystemPrompt, error) {
	spm.promptsMu.Lock()
	defer spm.promptsMu.Unlock()

	prompt, exists := spm.prompts[promptID]
	if !exists {
		return nil, fmt.Errorf("prompt %s not found", promptID)
	}

	prompt.Content = content
	prompt.Version = atomic.AddInt64(&spm.currentVersion, 1)
	prompt.UpdatedAt = time.Now()

	atomic.AddInt64(&spm.totalUpdates, 1)

	spm.logFunc("info", fmt.Sprintf("Updated prompt %s (version %d)", promptID, prompt.Version))

	return prompt, nil
}

// GetPrompt 获取提示词
func (spm *SystemPromptManager) GetPrompt(promptID string) (*SystemPrompt, error) {
	spm.promptsMu.RLock()
	defer spm.promptsMu.RUnlock()

	prompt, exists := spm.prompts[promptID]
	if !exists {
		return nil, fmt.Errorf("prompt %s not found", promptID)
	}

	return prompt, nil
}

// GetAllPrompts 获取所有提示词
func (spm *SystemPromptManager) GetAllPrompts() []*SystemPrompt {
	spm.promptsMu.RLock()
	defer spm.promptsMu.RUnlock()

	prompts := make([]*SystemPrompt, 0, len(spm.prompts))
	for _, prompt := range spm.prompts {
		prompts = append(prompts, prompt)
	}

	return prompts
}

// DeletePrompt 删除提示词
func (spm *SystemPromptManager) DeletePrompt(promptID string) error {
	spm.promptsMu.Lock()
	defer spm.promptsMu.Unlock()

	if _, exists := spm.prompts[promptID]; !exists {
		return fmt.Errorf("prompt %s not found", promptID)
	}

	delete(spm.prompts, promptID)

	return nil
}

// GetStatistics 获取统计信息
func (spm *SystemPromptManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_prompts": atomic.LoadInt64(&spm.totalPrompts),
		"total_updates": atomic.LoadInt64(&spm.totalUpdates),
		"current_version": atomic.LoadInt64(&spm.currentVersion),
	}
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	// 工具定义
	tools map[string]*ToolDefinition
	toolsMu sync.RWMutex

	// 工具执行函数
	handlers map[string]func(map[string]interface{}) (interface{}, error)
	handlersMu sync.RWMutex

	// 统计信息
	totalTools int64
	totalCalls int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]*ToolDefinition),
		handlers: make(map[string]func(map[string]interface{}) (interface{}, error)),
		logFunc:  defaultLogFunc,
	}
}

// RegisterTool 注册工具
func (tr *ToolRegistry) RegisterTool(def *ToolDefinition, handler func(map[string]interface{}) (interface{}, error)) error {
	tr.toolsMu.Lock()
	defer tr.toolsMu.Unlock()

	if _, exists := tr.tools[def.Name]; exists {
		return fmt.Errorf("tool %s already registered", def.Name)
	}

	if def.TimeoutMS == 0 {
		def.TimeoutMS = 5000 // 默认 5 秒
	}

	if def.RetryCount == 0 {
		def.RetryCount = 1
	}

	tr.tools[def.Name] = def

	tr.handlersMu.Lock()
	defer tr.handlersMu.Unlock()

	tr.handlers[def.Name] = handler

	atomic.AddInt64(&tr.totalTools, 1)

	tr.logFunc("info", fmt.Sprintf("Registered tool: %s", def.Name))

	return nil
}

// UnregisterTool 注销工具
func (tr *ToolRegistry) UnregisterTool(toolName string) error {
	tr.toolsMu.Lock()
	defer tr.toolsMu.Unlock()

	if _, exists := tr.tools[toolName]; !exists {
		return fmt.Errorf("tool %s not found", toolName)
	}

	delete(tr.tools, toolName)

	tr.handlersMu.Lock()
	defer tr.handlersMu.Unlock()

	delete(tr.handlers, toolName)

	return nil
}

// GetTool 获取工具定义
func (tr *ToolRegistry) GetTool(toolName string) (*ToolDefinition, error) {
	tr.toolsMu.RLock()
	defer tr.toolsMu.RUnlock()

	tool, exists := tr.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	return tool, nil
}

// GetAllTools 获取所有工具
func (tr *ToolRegistry) GetAllTools() []*ToolDefinition {
	tr.toolsMu.RLock()
	defer tr.toolsMu.RUnlock()

	tools := make([]*ToolDefinition, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}

	return tools
}

// CallTool 调用工具
func (tr *ToolRegistry) CallTool(toolName string, arguments map[string]interface{}) (interface{}, error) {
	tr.handlersMu.RLock()
	handler, exists := tr.handlers[toolName]
	tr.handlersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	atomic.AddInt64(&tr.totalCalls, 1)

	return handler(arguments)
}

// GetStatistics 获取统计信息
func (tr *ToolRegistry) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_tools": atomic.LoadInt64(&tr.totalTools),
		"total_calls": atomic.LoadInt64(&tr.totalCalls),
	}
}

// Agent AI Agent
type Agent struct {
	// Agent ID
	ID string `json:"id"`

	// Agent 名称
	Name string `json:"name"`

	// 系统提示词
	SystemPrompt *SystemPrompt `json:"system_prompt"`

	// 可用工具
	Tools []*ToolDefinition `json:"tools"`

	// 模型配置
	ModelConfig map[string]interface{} `json:"model_config"`

	// 工具调用历史
	ToolCalls []*ToolCall `json:"tool_calls"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 互斥锁
	mu sync.RWMutex
}

// NewAgent 创建 Agent
func NewAgent(id string, name string, systemPrompt *SystemPrompt) *Agent {
	return &Agent{
		ID:           id,
		Name:         name,
		SystemPrompt: systemPrompt,
		Tools:        make([]*ToolDefinition, 0),
		ModelConfig:  make(map[string]interface{}),
		ToolCalls:    make([]*ToolCall, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// AddTool 添加工具
func (a *Agent) AddTool(tool *ToolDefinition) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.Tools = append(a.Tools, tool)
	a.UpdatedAt = time.Now()
}

// RecordToolCall 记录工具调用
func (a *Agent) RecordToolCall(call *ToolCall) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ToolCalls = append(a.ToolCalls, call)
	a.UpdatedAt = time.Now()
}

// GetToolCalls 获取工具调用历史
func (a *Agent) GetToolCalls() []*ToolCall {
	a.mu.RLock()
	defer a.mu.RUnlock()

	calls := make([]*ToolCall, len(a.ToolCalls))
	copy(calls, a.ToolCalls)

	return calls
}

// ClearToolCalls 清除工具调用历史
func (a *Agent) ClearToolCalls() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ToolCalls = make([]*ToolCall, 0)
}

// UpdateSystemPrompt 更新系统提示词
func (a *Agent) UpdateSystemPrompt(prompt *SystemPrompt) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.SystemPrompt = prompt
	a.UpdatedAt = time.Now()
}

// AgentManager Agent 管理器
type AgentManager struct {
	// Agent 存储
	agents map[string]*Agent
	agentsMu sync.RWMutex

	// 系统提示词管理器
	promptManager *SystemPromptManager

	// 工具注册表
	toolRegistry *ToolRegistry

	// 统计信息
	totalAgents int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewAgentManager 创建 Agent 管理器
func NewAgentManager(promptManager *SystemPromptManager, toolRegistry *ToolRegistry) *AgentManager {
	return &AgentManager{
		agents:        make(map[string]*Agent),
		promptManager: promptManager,
		toolRegistry:  toolRegistry,
		logFunc:       defaultLogFunc,
	}
}

// CreateAgent 创建 Agent
func (am *AgentManager) CreateAgent(agentID string, name string, promptID string) (*Agent, error) {
	prompt, err := am.promptManager.GetPrompt(promptID)
	if err != nil {
		return nil, err
	}

	am.agentsMu.Lock()
	defer am.agentsMu.Unlock()

	if _, exists := am.agents[agentID]; exists {
		return nil, fmt.Errorf("agent %s already exists", agentID)
	}

	agent := NewAgent(agentID, name, prompt)

	am.agents[agentID] = agent
	atomic.AddInt64(&am.totalAgents, 1)

	am.logFunc("info", fmt.Sprintf("Created agent: %s", agentID))

	return agent, nil
}

// GetAgent 获取 Agent
func (am *AgentManager) GetAgent(agentID string) (*Agent, error) {
	am.agentsMu.RLock()
	defer am.agentsMu.RUnlock()

	agent, exists := am.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// DeleteAgent 删除 Agent
func (am *AgentManager) DeleteAgent(agentID string) error {
	am.agentsMu.Lock()
	defer am.agentsMu.Unlock()

	if _, exists := am.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(am.agents, agentID)

	return nil
}

// BindTool 绑定工具到 Agent
func (am *AgentManager) BindTool(agentID string, toolName string) error {
	agent, err := am.GetAgent(agentID)
	if err != nil {
		return err
	}

	tool, err := am.toolRegistry.GetTool(toolName)
	if err != nil {
		return err
	}

	agent.AddTool(tool)

	return nil
}

// ExecuteToolCall 执行工具调用
func (am *AgentManager) ExecuteToolCall(agentID string, toolName string, arguments map[string]interface{}) (*ToolCall, error) {
	agent, err := am.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	result, err := am.toolRegistry.CallTool(toolName, arguments)

	executionTime := time.Since(start).Milliseconds()

	call := &ToolCall{
		ID:            fmt.Sprintf("%s_%d", agentID, time.Now().UnixNano()),
		ToolName:      toolName,
		Arguments:     arguments,
		ExecutionTime: executionTime,
		CreatedAt:     time.Now(),
		Status:        "completed",
	}

	if err != nil {
		call.Error = err.Error()
		call.Status = "failed"
	} else {
		call.Result = result
	}

	now := time.Now()
	call.CompletedAt = &now

	agent.RecordToolCall(call)

	return call, nil
}

// GetAgentStatistics 获取 Agent 统计信息
func (am *AgentManager) GetAgentStatistics(agentID string) (map[string]interface{}, error) {
	agent, err := am.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	calls := agent.GetToolCalls()
	successCount := 0
	failedCount := 0

	for _, call := range calls {
		if call.Status == "completed" && call.Error == "" {
			successCount++
		} else {
			failedCount++
		}
	}

	return map[string]interface{}{
		"agent_id":       agentID,
		"name":           agent.Name,
		"tool_count":     len(agent.Tools),
		"call_count":     len(calls),
		"success_count":  successCount,
		"failed_count":   failedCount,
		"created_at":     agent.CreatedAt,
		"updated_at":     agent.UpdatedAt,
	}, nil
}

// GetStatistics 获取全局统计信息
func (am *AgentManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_agents":  atomic.LoadInt64(&am.totalAgents),
		"prompt_stats":  am.promptManager.GetStatistics(),
		"tool_stats":    am.toolRegistry.GetStatistics(),
	}
}

// HotUpdatePrompt 热更新提示词
func (am *AgentManager) HotUpdatePrompt(agentID string, newContent string) error {
	agent, err := am.GetAgent(agentID)
	if err != nil {
		return err
	}

	// 更新系统提示词管理器中的提示词
	prompt, err := am.promptManager.UpdatePrompt(agent.SystemPrompt.ID, newContent)
	if err != nil {
		return err
	}

	// 更新 Agent 中的提示词
	agent.UpdateSystemPrompt(prompt)

	am.logFunc("info", fmt.Sprintf("Hot updated prompt for agent %s", agentID))

	return nil
}

