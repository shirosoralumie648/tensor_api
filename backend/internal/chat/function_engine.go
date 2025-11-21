package chat

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// FunctionParameter 函数参数
type FunctionParameter struct {
	// 参数名
	Name string `json:"name"`

	// 参数类型
	Type string `json:"type"` // string, number, boolean, array, object

	// 参数描述
	Description string `json:"description"`

	// 是否必需
	Required bool `json:"required"`

	// 默认值
	Default interface{} `json:"default"`

	// 枚举值
	Enum []interface{} `json:"enum"`
}

// FunctionSpec 函数规范
type FunctionSpec struct {
	// 函数名
	Name string `json:"name"`

	// 函数描述
	Description string `json:"description"`

	// 参数列表
	Parameters []*FunctionParameter `json:"parameters"`

	// 返回类型
	ReturnType string `json:"return_type"`

	// 是否异步
	Async bool `json:"async"`

	// 超时时间（毫秒）
	TimeoutMS int `json:"timeout_ms"`
}

// FunctionRequest 函数请求
type FunctionRequest struct {
	// 请求 ID
	ID string `json:"id"`

	// 函数名
	FunctionName string `json:"function_name"`

	// 函数参数
	Arguments map[string]interface{} `json:"arguments"`

	// 上下文
	Context context.Context `json:"-"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// FunctionResponse 函数响应
type FunctionResponse struct {
	// 请求 ID
	RequestID string `json:"request_id"`

	// 函数名
	FunctionName string `json:"function_name"`

	// 结果
	Result interface{} `json:"result"`

	// 错误信息
	Error string `json:"error"`

	// 执行时间（毫秒）
	ExecutionTime int64 `json:"execution_time"`

	// 状态
	Status string `json:"status"` // success, failed, timeout

	// 创建时间
	CreatedAt time.Time `json:"created_at"`
}

// FunctionExecutor 函数执行器
type FunctionExecutor struct {
	// 函数定义
	specs map[string]*FunctionSpec
	specsMu sync.RWMutex

	// 函数实现
	handlers map[string]func(context.Context, map[string]interface{}) (interface{}, error)
	handlersMu sync.RWMutex

	// 执行历史
	history []*FunctionResponse
	historyMu sync.RWMutex

	// 统计信息
	totalCalls int64
	successCalls int64
	failedCalls int64
	totalTime int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewFunctionExecutor 创建函数执行器
func NewFunctionExecutor() *FunctionExecutor {
	return &FunctionExecutor{
		specs:    make(map[string]*FunctionSpec),
		handlers: make(map[string]func(context.Context, map[string]interface{}) (interface{}, error)),
		history:  make([]*FunctionResponse, 0),
		logFunc:  defaultLogFunc,
	}
}

// RegisterFunction 注册函数
func (fe *FunctionExecutor) RegisterFunction(spec *FunctionSpec, handler func(context.Context, map[string]interface{}) (interface{}, error)) error {
	fe.specsMu.Lock()
	defer fe.specsMu.Unlock()

	if _, exists := fe.specs[spec.Name]; exists {
		return fmt.Errorf("function %s already registered", spec.Name)
	}

	if spec.TimeoutMS == 0 {
		spec.TimeoutMS = 5000 // 默认 5 秒
	}

	fe.specs[spec.Name] = spec

	fe.handlersMu.Lock()
	defer fe.handlersMu.Unlock()

	fe.handlers[spec.Name] = handler

	fe.logFunc("info", fmt.Sprintf("Registered function: %s", spec.Name))

	return nil
}

// UnregisterFunction 注销函数
func (fe *FunctionExecutor) UnregisterFunction(functionName string) error {
	fe.specsMu.Lock()
	defer fe.specsMu.Unlock()

	if _, exists := fe.specs[functionName]; !exists {
		return fmt.Errorf("function %s not found", functionName)
	}

	delete(fe.specs, functionName)

	fe.handlersMu.Lock()
	defer fe.handlersMu.Unlock()

	delete(fe.handlers, functionName)

	return nil
}

// GetFunction 获取函数规范
func (fe *FunctionExecutor) GetFunction(functionName string) (*FunctionSpec, error) {
	fe.specsMu.RLock()
	defer fe.specsMu.RUnlock()

	spec, exists := fe.specs[functionName]
	if !exists {
		return nil, fmt.Errorf("function %s not found", functionName)
	}

	return spec, nil
}

// GetAllFunctions 获取所有函数规范
func (fe *FunctionExecutor) GetAllFunctions() []*FunctionSpec {
	fe.specsMu.RLock()
	defer fe.specsMu.RUnlock()

	specs := make([]*FunctionSpec, 0, len(fe.specs))
	for _, spec := range fe.specs {
		specs = append(specs, spec)
	}

	return specs
}

// Execute 执行函数
func (fe *FunctionExecutor) Execute(req *FunctionRequest) *FunctionResponse {
	start := time.Now()

	fe.handlersMu.RLock()
	handler, exists := fe.handlers[req.FunctionName]
	fe.handlersMu.RUnlock()

	if !exists {
		response := &FunctionResponse{
			RequestID:     req.ID,
			FunctionName:  req.FunctionName,
			Error:         fmt.Sprintf("function %s not found", req.FunctionName),
			Status:        "failed",
			ExecutionTime: time.Since(start).Milliseconds(),
			CreatedAt:     time.Now(),
		}

		fe.recordResponse(response)
		atomic.AddInt64(&fe.failedCalls, 1)

		return response
	}

	// 获取超时时间
	spec, _ := fe.GetFunction(req.FunctionName)
	timeoutDuration := time.Duration(spec.TimeoutMS) * time.Millisecond

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(req.Context, timeoutDuration)
	defer cancel()

	// 执行函数
	resultCh := make(chan interface{}, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := handler(ctx, req.Arguments)
		if err != nil {
			errCh <- err
		} else {
			resultCh <- result
		}
	}()

	var result interface{}
	var execErr error
	var status string = "success"

	select {
	case <-ctx.Done():
		execErr = fmt.Errorf("function execution timeout")
		status = "timeout"
		atomic.AddInt64(&fe.failedCalls, 1)

	case err := <-errCh:
		execErr = err
		status = "failed"
		atomic.AddInt64(&fe.failedCalls, 1)

	case res := <-resultCh:
		result = res
		atomic.AddInt64(&fe.successCalls, 1)
	}

	executionTime := time.Since(start).Milliseconds()

	response := &FunctionResponse{
		RequestID:     req.ID,
		FunctionName:  req.FunctionName,
		Result:        result,
		ExecutionTime: executionTime,
		Status:        status,
		CreatedAt:     time.Now(),
	}

	if execErr != nil {
		response.Error = execErr.Error()
	}

	fe.recordResponse(response)
	atomic.AddInt64(&fe.totalCalls, 1)
	atomic.AddInt64(&fe.totalTime, executionTime)

	fe.logFunc("debug", fmt.Sprintf("Function %s executed in %d ms (status: %s)", req.FunctionName, executionTime, status))

	return response
}

// recordResponse 记录响应
func (fe *FunctionExecutor) recordResponse(response *FunctionResponse) {
	fe.historyMu.Lock()
	defer fe.historyMu.Unlock()

	fe.history = append(fe.history, response)

	// 限制历史记录大小（最多 1000 条）
	if len(fe.history) > 1000 {
		fe.history = fe.history[len(fe.history)-1000:]
	}
}

// GetHistory 获取执行历史
func (fe *FunctionExecutor) GetHistory(limit int) []*FunctionResponse {
	fe.historyMu.RLock()
	defer fe.historyMu.RUnlock()

	if limit <= 0 || limit > len(fe.history) {
		limit = len(fe.history)
	}

	history := make([]*FunctionResponse, limit)
	copy(history, fe.history[len(fe.history)-limit:])

	return history
}

// GetStatistics 获取统计信息
func (fe *FunctionExecutor) GetStatistics() map[string]interface{} {
	totalCalls := atomic.LoadInt64(&fe.totalCalls)
	successCalls := atomic.LoadInt64(&fe.successCalls)
	failedCalls := atomic.LoadInt64(&fe.failedCalls)
	totalTime := atomic.LoadInt64(&fe.totalTime)

	var avgTime int64 = 0
	if totalCalls > 0 {
		avgTime = totalTime / totalCalls
	}

	return map[string]interface{}{
		"total_calls":    totalCalls,
		"success_calls":  successCalls,
		"failed_calls":   failedCalls,
		"success_rate":   fmt.Sprintf("%.2f%%", float64(successCalls)*100/float64(totalCalls)),
		"average_time":   avgTime,
		"total_time":     totalTime,
	}
}

// BatchExecute 批量执行函数
type BatchExecutor struct {
	// 函数执行器
	executor *FunctionExecutor

	// 并发控制
	semaphore chan struct{}

	// 统计信息
	totalBatches int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewBatchExecutor 创建批量执行器
func NewBatchExecutor(executor *FunctionExecutor, maxConcurrency int) *BatchExecutor {
	return &BatchExecutor{
		executor:   executor,
		semaphore:  make(chan struct{}, maxConcurrency),
		logFunc:    defaultLogFunc,
	}
}

// ExecuteBatch 批量执行函数
func (be *BatchExecutor) ExecuteBatch(requests []*FunctionRequest) []*FunctionResponse {
	responses := make([]*FunctionResponse, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)

		go func(idx int, r *FunctionRequest) {
			defer wg.Done()

			// 获取信号量
			be.semaphore <- struct{}{}
			defer func() { <-be.semaphore }()

			responses[idx] = be.executor.Execute(r)
		}(i, req)
	}

	wg.Wait()

	atomic.AddInt64(&be.totalBatches, 1)

	be.logFunc("info", fmt.Sprintf("Executed batch of %d functions", len(requests)))

	return responses
}

// GetStatistics 获取统计信息
func (be *BatchExecutor) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_batches":       atomic.LoadInt64(&be.totalBatches),
		"executor_stats":      be.executor.GetStatistics(),
	}
}

