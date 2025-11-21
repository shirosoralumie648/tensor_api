package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// BillingEvent 计费事件
type BillingEvent struct {
	// 事件 ID
	EventID string `json:"event_id"`

	// 用户 ID
	UserID string `json:"user_id"`

	// 事件类型
	EventType string `json:"event_type"`

	// 模型名称
	ModelName string `json:"model_name"`

	// 输入 token 数
	InputTokens int64 `json:"input_tokens"`

	// 输出 token 数
	OutputTokens int64 `json:"output_tokens"`

	// 费用
	Cost float64 `json:"cost"`

	// 请求 ID
	RequestID string `json:"request_id"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BillingEventQueue 计费事件队列
type BillingEventQueue struct {
	// 队列名称
	QueueName string

	// 事件缓冲区
	events chan *BillingEvent

	// 缓冲区大小
	bufferSize int

	// 是否运行
	running bool
	runMu   sync.Mutex

	// 停止信号
	stopCh chan struct{}

	// 统计信息
	enqueueCount  int64
	dequeueCount  int64
	discardCount  int64
	errorCount    int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewBillingEventQueue 创建计费事件队列
func NewBillingEventQueue(queueName string, bufferSize int) *BillingEventQueue {
	return &BillingEventQueue{
		QueueName:  queueName,
		events:     make(chan *BillingEvent, bufferSize),
		bufferSize: bufferSize,
		stopCh:     make(chan struct{}),
		logFunc:    defaultLogFunc,
	}
}

// Enqueue 将事件加入队列
func (beq *BillingEventQueue) Enqueue(event *BillingEvent) error {
	beq.runMu.Lock()
	if !beq.running {
		beq.runMu.Unlock()
		return fmt.Errorf("queue is not running")
	}
	beq.runMu.Unlock()

	select {
	case beq.events <- event:
		atomic.AddInt64(&beq.enqueueCount, 1)
		return nil
	default:
		atomic.AddInt64(&beq.discardCount, 1)
		beq.logFunc("warn", fmt.Sprintf("Queue %s is full, event discarded", beq.QueueName))
		return fmt.Errorf("queue is full")
	}
}

// Dequeue 从队列取出事件（非阻塞）
func (beq *BillingEventQueue) Dequeue(ctx context.Context) (*BillingEvent, error) {
	select {
	case event := <-beq.events:
		atomic.AddInt64(&beq.dequeueCount, 1)
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-beq.stopCh:
		return nil, fmt.Errorf("queue stopped")
	}
}

// Size 获取队列当前大小
func (beq *BillingEventQueue) Size() int {
	return len(beq.events)
}

// GetStatistics 获取统计信息
func (beq *BillingEventQueue) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"enqueue_count": atomic.LoadInt64(&beq.enqueueCount),
		"dequeue_count": atomic.LoadInt64(&beq.dequeueCount),
		"discard_count": atomic.LoadInt64(&beq.discardCount),
		"error_count":   atomic.LoadInt64(&beq.errorCount),
		"queue_size":    len(beq.events),
	}
}

// BillingConsumer 计费消费者
type BillingConsumer struct {
	// 消费者 ID
	ConsumerID string

	// 事件队列
	queue *BillingEventQueue

	// 配额管理器
	quotaManager *QuotaManager

	// 定价管理器
	pricingManager *PricingManager

	// 最大重试次数
	maxRetries int

	// 重试间隔
	retryInterval time.Duration

	// 批处理大小
	batchSize int

	// 死信队列
	deadLetterQueue *BillingEventQueue

	// 是否运行
	running bool
	runMu   sync.Mutex

	// 停止信号
	stopCh chan struct{}

	// 统计信息
	processedCount  int64
	successCount    int64
	failureCount    int64
	retryCount      int64
	dlqCount        int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewBillingConsumer 创建计费消费者
func NewBillingConsumer(
	consumerID string,
	queue *BillingEventQueue,
	quotaManager *QuotaManager,
	pricingManager *PricingManager,
) *BillingConsumer {
	dlq := NewBillingEventQueue(fmt.Sprintf("%s-dlq", queue.QueueName), 10000)
	
	return &BillingConsumer{
		ConsumerID:     consumerID,
		queue:          queue,
		quotaManager:   quotaManager,
		pricingManager: pricingManager,
		maxRetries:     3,
		retryInterval:  1 * time.Second,
		batchSize:      100,
		deadLetterQueue: dlq,
		stopCh:         make(chan struct{}),
		logFunc:        defaultLogFunc,
	}
}

// Start 启动消费者
func (bc *BillingConsumer) Start(ctx context.Context) {
	bc.runMu.Lock()
	if bc.running {
		bc.runMu.Unlock()
		return
	}
	bc.running = true
	bc.runMu.Unlock()

	go bc.run(ctx)
	bc.logFunc("info", fmt.Sprintf("Billing consumer %s started", bc.ConsumerID))
}

// run 消费者运行循环
func (bc *BillingConsumer) run(ctx context.Context) {
	batch := make([]*BillingEvent, 0, bc.batchSize)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			bc.processBatch(batch)
			bc.runMu.Lock()
			bc.running = false
			bc.runMu.Unlock()
			bc.logFunc("info", fmt.Sprintf("Billing consumer %s stopped", bc.ConsumerID))
			return

		case <-bc.stopCh:
			bc.processBatch(batch)
			bc.runMu.Lock()
			bc.running = false
			bc.runMu.Unlock()
			bc.logFunc("info", fmt.Sprintf("Billing consumer %s stopped", bc.ConsumerID))
			return

		case <-ticker.C:
			// 定期处理批次
			if len(batch) > 0 {
				bc.processBatch(batch)
				batch = make([]*BillingEvent, 0, bc.batchSize)
			}

		default:
			// 尝试从队列获取事件
			event, err := bc.queue.Dequeue(ctx)
			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					continue
				}
				continue
			}

			batch = append(batch, event)
			if len(batch) >= bc.batchSize {
				bc.processBatch(batch)
				batch = make([]*BillingEvent, 0, bc.batchSize)
			}
		}
	}
}

// processBatch 处理一批事件
func (bc *BillingConsumer) processBatch(events []*BillingEvent) {
	for _, event := range events {
		if err := bc.processEvent(event, 0); err != nil {
			atomic.AddInt64(&bc.failureCount, 1)
			bc.logFunc("error", fmt.Sprintf("Failed to process event %s: %v", event.EventID, err))
		} else {
			atomic.AddInt64(&bc.successCount, 1)
		}
		atomic.AddInt64(&bc.processedCount, 1)
	}
}

// processEvent 处理单个事件
func (bc *BillingConsumer) processEvent(event *BillingEvent, retryCount int) error {
	// 计算费用
	cost, err := bc.pricingManager.CalculatePrice(event.ModelName, event.InputTokens, event.OutputTokens)
	if err != nil {
		if retryCount < bc.maxRetries {
			atomic.AddInt64(&bc.retryCount, 1)
			time.Sleep(bc.retryInterval)
			return bc.processEvent(event, retryCount+1)
		}
		// 重试失败，发送到死信队列
		_ = bc.deadLetterQueue.Enqueue(event)
		atomic.AddInt64(&bc.dlqCount, 1)
		return err
	}

	// 确认扣费
	recordID := fmt.Sprintf("record-%s", event.EventID)
	err = bc.quotaManager.ConfirmDeduction(event.UserID, recordID, cost)
	if err != nil {
		if retryCount < bc.maxRetries {
			atomic.AddInt64(&bc.retryCount, 1)
			time.Sleep(bc.retryInterval)
			return bc.processEvent(event, retryCount+1)
		}
		// 重试失败，发送到死信队列
		_ = bc.deadLetterQueue.Enqueue(event)
		atomic.AddInt64(&bc.dlqCount, 1)
		return err
	}

	bc.logFunc("debug", fmt.Sprintf("Event %s processed: user=%s, cost=$%.4f", event.EventID, event.UserID, cost))

	return nil
}

// Stop 停止消费者
func (bc *BillingConsumer) Stop() {
	bc.runMu.Lock()
	if !bc.running {
		bc.runMu.Unlock()
		return
	}
	bc.runMu.Unlock()

	close(bc.stopCh)
}

// GetStatistics 获取统计信息
func (bc *BillingConsumer) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"processed_count": atomic.LoadInt64(&bc.processedCount),
		"success_count":   atomic.LoadInt64(&bc.successCount),
		"failure_count":   atomic.LoadInt64(&bc.failureCount),
		"retry_count":     atomic.LoadInt64(&bc.retryCount),
		"dlq_count":       atomic.LoadInt64(&bc.dlqCount),
	}
}

// GetDeadLetterQueueSize 获取死信队列大小
func (bc *BillingConsumer) GetDeadLetterQueueSize() int {
	return bc.deadLetterQueue.Size()
}

// ProcessDeadLetterQueue 处理死信队列
func (bc *BillingConsumer) ProcessDeadLetterQueue(ctx context.Context, callback func(*BillingEvent) error) error {
	for {
		event, err := bc.deadLetterQueue.Dequeue(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				break
			}
			return err
		}

		if event == nil {
			break
		}

		if callback != nil {
			if err := callback(event); err != nil {
				bc.logFunc("error", fmt.Sprintf("DLQ callback failed for event %s: %v", event.EventID, err))
				// 重新入队
				_ = bc.deadLetterQueue.Enqueue(event)
			}
		}
	}

	return nil
}

// BillingEventLogger 计费事件日志记录器
type BillingEventLogger struct {
	// 日志文件路径
	logPath string

	// 日志缓冲区
	buffer []*BillingEvent
	bufMu  sync.Mutex

	// 缓冲区大小
	bufferSize int

	// 刷新间隔
	flushInterval time.Duration

	// 统计信息
	totalLogged int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewBillingEventLogger 创建计费事件日志记录器
func NewBillingEventLogger(logPath string, bufferSize int, flushInterval time.Duration) *BillingEventLogger {
	return &BillingEventLogger{
		logPath:       logPath,
		buffer:        make([]*BillingEvent, 0, bufferSize),
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		logFunc:       defaultLogFunc,
	}
}

// Log 记录计费事件
func (bel *BillingEventLogger) Log(event *BillingEvent) error {
	bel.bufMu.Lock()
	defer bel.bufMu.Unlock()

	bel.buffer = append(bel.buffer, event)

	if len(bel.buffer) >= bel.bufferSize {
		return bel.flush()
	}

	return nil
}

// flush 刷新缓冲区到磁盘
func (bel *BillingEventLogger) flush() error {
	if len(bel.buffer) == 0 {
		return nil
	}

	data := make([]map[string]interface{}, len(bel.buffer))
	for i, event := range bel.buffer {
		data[i] = map[string]interface{}{
			"event_id":      event.EventID,
			"user_id":       event.UserID,
			"event_type":    event.EventType,
			"model_name":    event.ModelName,
			"input_tokens":  event.InputTokens,
			"output_tokens": event.OutputTokens,
			"cost":          event.Cost,
			"request_id":    event.RequestID,
			"timestamp":     event.Timestamp,
			"metadata":      event.Metadata,
		}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		bel.logFunc("error", fmt.Sprintf("Failed to marshal billing events: %v", err))
		return err
	}

	// 这里应该写入文件或数据库，简化实现仅记录
	bel.logFunc("info", fmt.Sprintf("Flushed %d billing events to log", len(bel.buffer)))
	bel.logFunc("debug", fmt.Sprintf("Billing events JSON: %s", string(jsonData)))

	atomic.AddInt64(&bel.totalLogged, int64(len(bel.buffer)))
	bel.buffer = make([]*BillingEvent, 0, bel.bufferSize)

	return nil
}

// Flush 手动刷新缓冲区
func (bel *BillingEventLogger) Flush() error {
	bel.bufMu.Lock()
	defer bel.bufMu.Unlock()

	return bel.flush()
}

// GetTotalLogged 获取已记录的事件总数
func (bel *BillingEventLogger) GetTotalLogged() int64 {
	return atomic.LoadInt64(&bel.totalLogged)
}

// AsyncBillingService 异步计费服务
type AsyncBillingService struct {
	// 事件队列
	queue *BillingEventQueue

	// 消费者列表
	consumers []*BillingConsumer
	consumersMu sync.RWMutex

	// 事件日志记录器
	logger *BillingEventLogger

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewAsyncBillingService 创建异步计费服务
func NewAsyncBillingService(queueName string, queueSize int) *AsyncBillingService {
	ctx, cancel := context.WithCancel(context.Background())

	queue := NewBillingEventQueue(queueName, queueSize)

	return &AsyncBillingService{
		queue:     queue,
		consumers: make([]*BillingConsumer, 0),
		ctx:       ctx,
		cancel:    cancel,
		logFunc:   defaultLogFunc,
	}
}

// AddConsumer 添加消费者
func (abs *AsyncBillingService) AddConsumer(consumer *BillingConsumer) {
	abs.consumersMu.Lock()
	defer abs.consumersMu.Unlock()

	abs.consumers = append(abs.consumers, consumer)
}

// PublishEvent 发布计费事件
func (abs *AsyncBillingService) PublishEvent(event *BillingEvent) error {
	return abs.queue.Enqueue(event)
}

// Start 启动服务
func (abs *AsyncBillingService) Start() {
	abs.consumersMu.RLock()
	defer abs.consumersMu.RLock()

	for _, consumer := range abs.consumers {
		consumer.Start(abs.ctx)
	}

	abs.logFunc("info", "Async billing service started")
}

// Stop 停止服务
func (abs *AsyncBillingService) Stop() {
	abs.consumersMu.Lock()
	defer abs.consumersMu.Unlock()

	for _, consumer := range abs.consumers {
		consumer.Stop()
	}

	abs.cancel()
	abs.logFunc("info", "Async billing service stopped")
}

// GetStatistics 获取统计信息
func (abs *AsyncBillingService) GetStatistics() map[string]interface{} {
	abs.consumersMu.RLock()
	defer abs.consumersMu.RUnlock()

	stats := map[string]interface{}{
		"queue_stats": abs.queue.GetStatistics(),
		"consumers":   len(abs.consumers),
	}

	for i, consumer := range abs.consumers {
		stats[fmt.Sprintf("consumer_%d", i)] = consumer.GetStatistics()
	}

	return stats
}

