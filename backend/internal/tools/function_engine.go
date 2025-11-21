package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// JSONSchema JSON 模式定义
type JSONSchema struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
}

// ToolParameter 工具参数
type ToolParameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Schema      *JSONSchema
}

// Tool 工具定义
type Tool struct {
	// 工具名称
	Name string `json:"name"`

	// 工具描述
	Description string `json:"description"`

	// 参数定义
	Parameters *JSONSchema `json:"parameters,omitempty"`

	// 执行处理函数
	Handler func(ctx context.Context, args map[string]interface{}) (interface{}, error) `json:"-"`

	// 执行超时
	Timeout time.Duration

	// 最大调用频率
	MaxCallsPerMinute int

	// 创建时间
	CreatedAt time.Time

	// 最后调用时间
	LastCalledAt *time.Time

	// 调用计数
	CallCount int64
}

// ToolCall 工具调用请求
type ToolCall struct {
	// 调用 ID
	ID string `json:"id"`

	// 工具名称
	ToolName string `json:"tool_name"`

	// 参数
	Arguments map[string]interface{} `json:"arguments"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	// 调用 ID
	CallID string `json:"call_id"`

	// 工具名称
	ToolName string `json:"tool_name"`

	// 执行状态
	Status string `json:"status"` // success, error, timeout

	// 结果数据
	Result interface{} `json:"result,omitempty"`

	// 错误信息
	Error string `json:"error,omitempty"`

	// 执行耗时（毫秒）
	Duration int64 `json:"duration"`

	// 完成时间
	CompletedAt time.Time `json:"completed_at"`
}

// FunctionEngine Function Calling 引擎
type FunctionEngine struct {
	// 工具集合
	tools   map[string]*Tool
	toolsMu sync.RWMutex

	// 调用历史
	callHistory []*ToolCallResult
	historyMu   sync.RWMutex

	// 统计信息
	totalCalls   int64
	successCalls int64
	failedCalls  int64
	timeoutCalls int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewFunctionEngine 创建 Function Calling 引擎
func NewFunctionEngine() *FunctionEngine {
	return &FunctionEngine{
		tools:       make(map[string]*Tool),
		callHistory: make([]*ToolCallResult, 0),
		logFunc:     defaultLogFuncTools,
	}
}

// defaultLogFuncTools 默认日志函数
func defaultLogFuncTools(level, msg string, args ...interface{}) {
	// 默认实现：忽略日志
}

// RegisterTool 注册工具
func (fe *FunctionEngine) RegisterTool(tool *Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Handler == nil {
		return fmt.Errorf("tool handler cannot be nil")
	}

	if tool.Timeout == 0 {
		tool.Timeout = 30 * time.Second
	}

	tool.CreatedAt = time.Now()

	fe.toolsMu.Lock()
	fe.tools[tool.Name] = tool
	fe.toolsMu.Unlock()

	fe.logFunc("info", fmt.Sprintf("Registered tool: %s", tool.Name))

	return nil
}

// UnregisterTool 注销工具
func (fe *FunctionEngine) UnregisterTool(toolName string) error {
	fe.toolsMu.Lock()
	defer fe.toolsMu.Unlock()

	if _, exists := fe.tools[toolName]; !exists {
		return fmt.Errorf("tool not found: %s", toolName)
	}

	delete(fe.tools, toolName)

	fe.logFunc("info", fmt.Sprintf("Unregistered tool: %s", toolName))

	return nil
}

// GetTool 获取工具
func (fe *FunctionEngine) GetTool(toolName string) (*Tool, error) {
	fe.toolsMu.RLock()
	defer fe.toolsMu.RUnlock()

	tool, exists := fe.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	return tool, nil
}

// ListTools 列出所有工具
func (fe *FunctionEngine) ListTools() []*Tool {
	fe.toolsMu.RLock()
	defer fe.toolsMu.RUnlock()

	tools := make([]*Tool, 0)
	for _, tool := range fe.tools {
		tools = append(tools, tool)
	}

	return tools
}

// ValidateArguments 验证参数
func (fe *FunctionEngine) ValidateArguments(toolName string, args map[string]interface{}) error {
	tool, err := fe.GetTool(toolName)
	if err != nil {
		return err
	}

	if tool.Parameters == nil {
		return nil
	}

	// 检查必需参数
	for _, required := range tool.Parameters.Required {
		if _, exists := args[required]; !exists {
			return fmt.Errorf("missing required parameter: %s", required)
		}
	}

	// TODO: 可以添加更详细的类型检查

	return nil
}

// ExecuteTool 执行工具
func (fe *FunctionEngine) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (*ToolCallResult, error) {
	// 验证参数
	if err := fe.ValidateArguments(toolName, args); err != nil {
		return nil, err
	}

	tool, err := fe.GetTool(toolName)
	if err != nil {
		return nil, err
	}

	// 生成调用 ID
	callID := fmt.Sprintf("call-%d", time.Now().UnixNano())

	// 记录调用信息
	_ = &ToolCall{
		ID:        callID,
		ToolName:  toolName,
		Arguments: args,
		CreatedAt: time.Now(),
	}

	startTime := time.Now()

	// 创建超时 context
	execCtx, cancel := context.WithTimeout(ctx, tool.Timeout)
	defer cancel()

	// 执行工具
	var result interface{}
	var execErr error
	var status string

	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		res, err := tool.Handler(execCtx, args)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- res
		}
	}()

	// 等待结果或超时
	select {
	case result = <-resultChan:
		status = "success"
		atomic.AddInt64(&fe.successCalls, 1)

	case execErr = <-errChan:
		status = "error"
		atomic.AddInt64(&fe.failedCalls, 1)

	case <-execCtx.Done():
		status = "timeout"
		atomic.AddInt64(&fe.timeoutCalls, 1)
		execErr = fmt.Errorf("tool execution timeout")
	}

	duration := time.Since(startTime)
	atomic.AddInt64(&fe.totalCalls, 1)

	// 更新工具信息
	now := time.Now()
	tool.LastCalledAt = &now
	atomic.AddInt64(&tool.CallCount, 1)

	// 创建结果
	callResult := &ToolCallResult{
		CallID:      callID,
		ToolName:    toolName,
		Status:      status,
		Result:      result,
		Duration:    duration.Milliseconds(),
		CompletedAt: now,
	}

	if execErr != nil {
		callResult.Error = execErr.Error()
	}

	// 保存历史
	fe.historyMu.Lock()
	fe.callHistory = append(fe.callHistory, callResult)
	// 保持最近 1000 条记录
	if len(fe.callHistory) > 1000 {
		fe.callHistory = fe.callHistory[1:]
	}
	fe.historyMu.Unlock()

	fe.logFunc("debug", fmt.Sprintf("Executed tool %s: %s (%.2fms)", toolName, status, float64(duration.Milliseconds())))

	if execErr != nil {
		return callResult, execErr
	}

	return callResult, nil
}

// BatchExecute 批量执行工具调用
func (fe *FunctionEngine) BatchExecute(ctx context.Context, calls []*ToolCall) ([]*ToolCallResult, error) {
	results := make([]*ToolCallResult, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)

		go func(index int, toolCall *ToolCall) {
			defer wg.Done()

			result, _ := fe.ExecuteTool(ctx, toolCall.ToolName, toolCall.Arguments)
			if result != nil {
				results[index] = result
			}
		}(i, call)
	}

	wg.Wait()

	return results, nil
}

// GetStatistics 获取统计信息
func (fe *FunctionEngine) GetStatistics() map[string]interface{} {
	fe.toolsMu.RLock()
	toolCount := len(fe.tools)
	fe.toolsMu.RUnlock()

	total := atomic.LoadInt64(&fe.totalCalls)
	success := atomic.LoadInt64(&fe.successCalls)

	successRate := float32(0)
	if total > 0 {
		successRate = float32(success) / float32(total)
	}

	return map[string]interface{}{
		"total_tools":   toolCount,
		"total_calls":   total,
		"success_calls": success,
		"failed_calls":  atomic.LoadInt64(&fe.failedCalls),
		"timeout_calls": atomic.LoadInt64(&fe.timeoutCalls),
		"success_rate":  successRate,
	}
}

// GetCallHistory 获取调用历史
func (fe *FunctionEngine) GetCallHistory(limit int) []*ToolCallResult {
	fe.historyMu.RLock()
	defer fe.historyMu.RUnlock()

	start := len(fe.callHistory) - limit
	if start < 0 {
		start = 0
	}

	history := make([]*ToolCallResult, len(fe.callHistory)-start)
	copy(history, fe.callHistory[start:])

	return history
}

// ConvertToOpenAIFormat 转换为 OpenAI 格式
func (fe *FunctionEngine) ConvertToOpenAIFormat() []map[string]interface{} {
	tools := fe.ListTools()

	var functions []map[string]interface{}

	for _, tool := range tools {
		function := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
		}

		if tool.Parameters != nil {
			function["parameters"] = tool.Parameters
		} else {
			function["parameters"] = map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			}
		}

		functions = append(functions, map[string]interface{}{
			"type":     "function",
			"function": function,
		})
	}

	return functions
}

// ParseOpenAIToolCall 解析 OpenAI 工具调用格式
func (fe *FunctionEngine) ParseOpenAIToolCall(toolCall map[string]interface{}) (*ToolCall, error) {
	// 提取工具名称
	toolName, ok := toolCall["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid tool name")
	}

	// 解析参数
	var args map[string]interface{}

	if argumentsStr, ok := toolCall["arguments"].(string); ok {
		// JSON 字符串格式
		if err := json.Unmarshal([]byte(argumentsStr), &args); err != nil {
			return nil, fmt.Errorf("invalid arguments JSON: %v", err)
		}
	} else if arguments, ok := toolCall["arguments"].(map[string]interface{}); ok {
		// 直接对象格式
		args = arguments
	} else {
		args = make(map[string]interface{})
	}

	return &ToolCall{
		ID:        fmt.Sprintf("call-%d", time.Now().UnixNano()),
		ToolName:  toolName,
		Arguments: args,
		CreatedAt: time.Now(),
	}, nil
}

// ToolBuilder 工具构建器
type ToolBuilder struct {
	name        string
	description string
	parameters  *JSONSchema
	handler     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
	timeout     time.Duration
}

// NewToolBuilder 创建工具构建器
func NewToolBuilder(name string) *ToolBuilder {
	return &ToolBuilder{
		name:    name,
		timeout: 30 * time.Second,
	}
}

// Description 设置描述
func (tb *ToolBuilder) Description(desc string) *ToolBuilder {
	tb.description = desc
	return tb
}

// Parameters 设置参数
func (tb *ToolBuilder) Parameters(params *JSONSchema) *ToolBuilder {
	tb.parameters = params
	return tb
}

// Handler 设置处理函数
func (tb *ToolBuilder) Handler(handler func(ctx context.Context, args map[string]interface{}) (interface{}, error)) *ToolBuilder {
	tb.handler = handler
	return tb
}

// Timeout 设置超时
func (tb *ToolBuilder) Timeout(timeout time.Duration) *ToolBuilder {
	tb.timeout = timeout
	return tb
}

// Build 构建工具
func (tb *ToolBuilder) Build() *Tool {
	return &Tool{
		Name:        tb.name,
		Description: tb.description,
		Parameters:  tb.parameters,
		Handler:     tb.handler,
		Timeout:     tb.timeout,
	}
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	// 执行 ID
	ID string

	// 工具引擎
	Engine *FunctionEngine

	// 调用堆栈（防止递归）
	CallStack []string

	// 最大调用深度
	MaxDepth int

	// 创建时间
	CreatedAt time.Time

	// 互斥锁
	mu sync.RWMutex
}

// NewExecutionContext 创建执行上下文
func NewExecutionContext(engine *FunctionEngine) *ExecutionContext {
	return &ExecutionContext{
		ID:        fmt.Sprintf("ctx-%d", time.Now().UnixNano()),
		Engine:    engine,
		CallStack: make([]string, 0),
		MaxDepth:  10,
		CreatedAt: time.Now(),
	}
}

// PushCall 推入调用栈
func (ec *ExecutionContext) PushCall(toolName string) error {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if len(ec.CallStack) >= ec.MaxDepth {
		return fmt.Errorf("maximum call depth exceeded")
	}

	// 检查递归
	for _, call := range ec.CallStack {
		if call == toolName {
			return fmt.Errorf("recursive call detected: %s", toolName)
		}
	}

	ec.CallStack = append(ec.CallStack, toolName)

	return nil
}

// PopCall 弹出调用栈
func (ec *ExecutionContext) PopCall() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if len(ec.CallStack) > 0 {
		ec.CallStack = ec.CallStack[:len(ec.CallStack)-1]
	}
}

// GetDepth 获取调用深度
func (ec *ExecutionContext) GetDepth() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return len(ec.CallStack)
}
