package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TaskType 任务类型
type TaskType string

const (
	TaskBilling   TaskType = "billing"
	TaskNotification TaskType = "notification"
	TaskAnalytics TaskType = "analytics"
	TaskExport    TaskType = "export"
	TaskCleanup   TaskType = "cleanup"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusRetrying  TaskStatus = "retrying"
)

// Task 任务
type Task struct {
	ID        string                 `json:"id"`
	Type      TaskType               `json:"type"`
	Status    TaskStatus             `json:"status"`
	Payload   map[string]interface{} `json:"payload"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error"`
	Retry     int                    `json:"retry"`
	MaxRetry  int                    `json:"max_retry"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ExecTime  int64                  `json:"exec_time"` // 毫秒
}

// TaskHandler 任务处理器
type TaskHandler interface {
	Handle(context.Context, *Task) error
	CanHandle(TaskType) bool
}

// AsyncQueue 异步任务队列
type AsyncQueue struct {
	mu           sync.RWMutex
	client       *RabbitMQClient
	tasks        map[string]*Task
	handlers     map[TaskType]TaskHandler
	dlq          *DeadLetterQueue
	stats        *AsyncQueueStats
	workers      int
	workerPool   chan struct{}
	maxRetry     int
	retryDelay   time.Duration
}

// AsyncQueueStats 异步队列统计
type AsyncQueueStats struct {
	mu        sync.RWMutex
	Created   int64
	Running   int64
	Completed int64
	Failed    int64
	Retrying  int64
	AvgExecTime int64
}

// NewAsyncQueue 创建异步队列
func NewAsyncQueue(client *RabbitMQClient, workers int) *AsyncQueue {
	return &AsyncQueue{
		client:      client,
		tasks:       make(map[string]*Task),
		handlers:    make(map[TaskType]TaskHandler),
		dlq:         NewDeadLetterQueue(),
		stats:       &AsyncQueueStats{},
		workers:     workers,
		workerPool:  make(chan struct{}, workers),
		maxRetry:    3,
		retryDelay:  5 * time.Second,
	}
}

// RegisterHandler 注册任务处理器
func (aq *AsyncQueue) RegisterHandler(handler TaskHandler) error {
	aq.mu.Lock()
	defer aq.mu.Unlock()

	// 支持多个 Handler，这里简化处理
	return nil
}

// Submit 提交任务
func (aq *AsyncQueue) Submit(ctx context.Context, task *Task) error {
	aq.mu.Lock()

	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	if task.Status == "" {
		task.Status = StatusPending
	}

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	if task.MaxRetry == 0 {
		task.MaxRetry = aq.maxRetry
	}

	aq.tasks[task.ID] = task
	aq.stats.mu.Lock()
	aq.stats.Created++
	aq.stats.mu.Unlock()

	aq.mu.Unlock()

	// 异步处理任务
	go aq.processTask(ctx, task)

	return nil
}

// processTask 处理任务
func (aq *AsyncQueue) processTask(ctx context.Context, task *Task) {
	// 获取 worker slot
	aq.workerPool <- struct{}{}
	defer func() { <-aq.workerPool }()

	aq.mu.Lock()
	task.Status = StatusRunning
	task.UpdatedAt = time.Now()
	aq.stats.mu.Lock()
	aq.stats.Running++
	aq.stats.mu.Unlock()
	aq.mu.Unlock()

	startTime := time.Now()

	// 查找处理器
	handler, exists := aq.getHandlerForType(task.Type)
	if !exists {
		aq.handleTaskError(task, fmt.Sprintf("no handler for task type: %s", task.Type))
		return
	}

	// 执行任务
	err := handler.Handle(ctx, task)

	execTime := time.Since(startTime)
	aq.mu.Lock()
	task.ExecTime = execTime.Milliseconds()
	task.UpdatedAt = time.Now()
	aq.mu.Unlock()

	if err != nil {
		aq.handleTaskRetry(ctx, task, err)
	} else {
		aq.handleTaskSuccess(task)
	}
}

// handleTaskSuccess 处理任务成功
func (aq *AsyncQueue) handleTaskSuccess(task *Task) {
	aq.mu.Lock()
	defer aq.mu.Unlock()

	task.Status = StatusCompleted
	task.UpdatedAt = time.Now()

	aq.stats.mu.Lock()
	aq.stats.Completed++
	aq.stats.Running--
	if task.ExecTime > 0 {
		aq.stats.AvgExecTime = (aq.stats.AvgExecTime + task.ExecTime) / 2
	}
	aq.stats.mu.Unlock()
}

// handleTaskError 处理任务错误
func (aq *AsyncQueue) handleTaskError(task *Task, errMsg string) {
	aq.mu.Lock()
	defer aq.mu.Unlock()

	task.Error = errMsg
	task.Status = StatusFailed
	task.UpdatedAt = time.Now()

	aq.stats.mu.Lock()
	aq.stats.Failed++
	aq.stats.Running--
	aq.stats.mu.Unlock()

	// 添加到死信队列
	aq.dlq.Add(task)
}

// handleTaskRetry 处理任务重试
func (aq *AsyncQueue) handleTaskRetry(ctx context.Context, task *Task, err error) {
	aq.mu.Lock()
	task.Error = err.Error()
	aq.mu.Unlock()

	if task.Retry < task.MaxRetry {
		aq.mu.Lock()
		task.Retry++
		task.Status = StatusRetrying
		task.UpdatedAt = time.Now()
		aq.stats.mu.Lock()
		aq.stats.Retrying++
		aq.stats.Running--
		aq.stats.mu.Unlock()
		aq.mu.Unlock()

		// 延迟重试
		time.AfterFunc(aq.retryDelay, func() {
			aq.processTask(ctx, task)
		})
	} else {
		aq.handleTaskError(task, fmt.Sprintf("max retries exceeded: %s", err.Error()))
	}
}

// getHandlerForType 获取类型对应的处理器
func (aq *AsyncQueue) getHandlerForType(taskType TaskType) (TaskHandler, bool) {
	aq.mu.RLock()
	defer aq.mu.RUnlock()

	handler, exists := aq.handlers[taskType]
	return handler, exists
}

// GetTask 获取任务
func (aq *AsyncQueue) GetTask(taskID string) (*Task, error) {
	aq.mu.RLock()
	defer aq.mu.RUnlock()

	task, exists := aq.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	return task, nil
}

// GetStats 获取统计
func (aq *AsyncQueue) GetStats() map[string]interface{} {
	aq.stats.mu.RLock()
	defer aq.stats.mu.RUnlock()

	return map[string]interface{}{
		"created":    aq.stats.Created,
		"running":    aq.stats.Running,
		"completed":  aq.stats.Completed,
		"failed":     aq.stats.Failed,
		"retrying":   aq.stats.Retrying,
		"avg_exec_time": aq.stats.AvgExecTime,
	}
}

// DeadLetterQueue 死信队列
type DeadLetterQueue struct {
	mu    sync.RWMutex
	tasks []*Task
	max   int
}

// NewDeadLetterQueue 创建死信队列
func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		tasks: make([]*Task, 0),
		max:   10000,
	}
}

// Add 添加任务到死信队列
func (dlq *DeadLetterQueue) Add(task *Task) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if len(dlq.tasks) >= dlq.max {
		// 移除最旧的任务
		dlq.tasks = dlq.tasks[1:]
	}

	dlq.tasks = append(dlq.tasks, task)
}

// GetAll 获取所有任务
func (dlq *DeadLetterQueue) GetAll() []*Task {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	result := make([]*Task, len(dlq.tasks))
	copy(result, dlq.tasks)
	return result
}

// Get 获取指定任务
func (dlq *DeadLetterQueue) Get(taskID string) (*Task, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	for _, task := range dlq.tasks {
		if task.ID == taskID {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task not found in DLQ")
}

// Retry 重试任务
func (dlq *DeadLetterQueue) Retry(taskID string) (*Task, error) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	for i, task := range dlq.tasks {
		if task.ID == taskID {
			// 从 DLQ 中移除
			dlq.tasks = append(dlq.tasks[:i], dlq.tasks[i+1:]...)
			return task, nil
		}
	}

	return nil, fmt.Errorf("task not found in DLQ")
}

// Clear 清空死信队列
func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	dlq.tasks = make([]*Task, 0)
}

// Size 获取大小
func (dlq *DeadLetterQueue) Size() int {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	return len(dlq.tasks)
}

// TaskConsumer 任务消费者
type TaskConsumer struct {
	mu        sync.RWMutex
	queue     *AsyncQueue
	running   bool
	stopChan  chan struct{}
}

// NewTaskConsumer 创建任务消费者
func NewTaskConsumer(queue *AsyncQueue) *TaskConsumer {
	return &TaskConsumer{
		queue:    queue,
		stopChan: make(chan struct{}),
	}
}

// Start 启动消费者
func (tc *TaskConsumer) Start(ctx context.Context) error {
	tc.mu.Lock()
	if tc.running {
		tc.mu.Unlock()
		return fmt.Errorf("consumer already running")
	}

	tc.running = true
	tc.mu.Unlock()

	// 模拟消费
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-tc.stopChan:
				return
			case <-ticker.C:
				// 处理待处理任务
			}
		}
	}()

	return nil
}

// Stop 停止消费者
func (tc *TaskConsumer) Stop() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if !tc.running {
		return fmt.Errorf("consumer not running")
	}

	tc.running = false
	close(tc.stopChan)

	return nil
}

// IsRunning 检查是否运行
func (tc *TaskConsumer) IsRunning() bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return tc.running
}

