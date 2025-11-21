package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	URL              string
	ClusterMode      bool
	ClusterAddresses []string
	PrefetchCount    int
	PrefetchSize     int
	MaxRetries       int
	ConnectTimeout   time.Duration
	Heartbeat        time.Duration
}

// QueueConfig 队列配置
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]interface{}
}

// ExchangeConfig 交换机配置
type ExchangeConfig struct {
	Name       string
	Kind       string // direct, fanout, topic, headers
	Durable    bool
	AutoDelete bool
	NoWait     bool
	Args       map[string]interface{}
}

// Message 消息
type Message struct {
	ID          string    `json:"id"`
	Body        []byte    `json:"body"`
	ContentType string    `json:"content_type"`
	Timestamp   time.Time `json:"timestamp"`
	Retry       int       `json:"retry"`
	MaxRetry    int       `json:"max_retry"`
	Headers     map[string]string `json:"headers"`
}

// RabbitMQClient RabbitMQ 客户端
type RabbitMQClient struct {
	mu              sync.RWMutex
	config          *RabbitMQConfig
	isConnected     bool
	queues          map[string]*QueueConfig
	exchanges       map[string]*ExchangeConfig
	bindings        map[string][]string
	consumers       map[string]chan *Message
	errorChan       chan error
	stats           *QueueStats
}

// QueueStats 队列统计
type QueueStats struct {
	mu          sync.RWMutex
	Published   int64
	Consumed    int64
	Failed      int64
	Acknowledged int64
	Requeued    int64
}

// NewRabbitMQClient 创建 RabbitMQ 客户端
func NewRabbitMQClient(config *RabbitMQConfig) (*RabbitMQClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	client := &RabbitMQClient{
		config:      config,
		isConnected: false,
		queues:      make(map[string]*QueueConfig),
		exchanges:   make(map[string]*ExchangeConfig),
		bindings:    make(map[string][]string),
		consumers:   make(map[string]chan *Message),
		errorChan:   make(chan error, 100),
		stats:       &QueueStats{},
	}

	return client, nil
}

// Connect 连接到 RabbitMQ
func (rc *RabbitMQClient) Connect(ctx context.Context) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.isConnected {
		return fmt.Errorf("already connected")
	}

	// 模拟连接
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		rc.isConnected = true
		return nil
	}
}

// Disconnect 断开连接
func (rc *RabbitMQClient) Disconnect() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.isConnected = false

	// 关闭所有消费者
	for _, ch := range rc.consumers {
		close(ch)
	}
	rc.consumers = make(map[string]chan *Message)

	return nil
}

// IsConnected 检查是否连接
func (rc *RabbitMQClient) IsConnected() bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	return rc.isConnected
}

// DeclareQueue 声明队列
func (rc *RabbitMQClient) DeclareQueue(ctx context.Context, queueConfig *QueueConfig) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if !rc.isConnected {
		return fmt.Errorf("not connected")
	}

	if queueConfig == nil || queueConfig.Name == "" {
		return fmt.Errorf("invalid queue config")
	}

	rc.queues[queueConfig.Name] = queueConfig

	return nil
}

// DeclareExchange 声明交换机
func (rc *RabbitMQClient) DeclareExchange(ctx context.Context, exchangeConfig *ExchangeConfig) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if !rc.isConnected {
		return fmt.Errorf("not connected")
	}

	if exchangeConfig == nil || exchangeConfig.Name == "" {
		return fmt.Errorf("invalid exchange config")
	}

	rc.exchanges[exchangeConfig.Name] = exchangeConfig

	return nil
}

// BindQueue 绑定队列到交换机
func (rc *RabbitMQClient) BindQueue(ctx context.Context, queueName, exchangeName, routingKey string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, qExists := rc.queues[queueName]; !qExists {
		return fmt.Errorf("queue %s not found", queueName)
	}

	if _, eExists := rc.exchanges[exchangeName]; !eExists {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}

	bindingKey := exchangeName + ":" + routingKey
	rc.bindings[queueName] = append(rc.bindings[queueName], bindingKey)

	return nil
}

// Publish 发布消息
func (rc *RabbitMQClient) Publish(ctx context.Context, exchangeName, routingKey string, message *Message) error {
	rc.mu.RLock()
	isConnected := rc.isConnected
	rc.mu.RUnlock()

	if !isConnected {
		return fmt.Errorf("not connected")
	}

	if message == nil {
		return fmt.Errorf("message is required")
	}

	message.Timestamp = time.Now()
	if message.ID == "" {
		message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	rc.stats.mu.Lock()
	rc.stats.Published++
	rc.stats.mu.Unlock()

	return nil
}

// Consume 消费消息
func (rc *RabbitMQClient) Consume(ctx context.Context, queueName string, autoAck bool) (chan *Message, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if !rc.isConnected {
		return nil, fmt.Errorf("not connected")
	}

	if _, exists := rc.queues[queueName]; !exists {
		return nil, fmt.Errorf("queue %s not found", queueName)
	}

	msgChan := make(chan *Message, 100)
	rc.consumers[queueName] = msgChan

	return msgChan, nil
}

// Acknowledge 确认消息
func (rc *RabbitMQClient) Acknowledge(ctx context.Context, message *Message) error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	rc.stats.mu.Lock()
	rc.stats.Acknowledged++
	rc.stats.mu.Unlock()

	return nil
}

// Nack 拒绝消息
func (rc *RabbitMQClient) Nack(ctx context.Context, message *Message, requeue bool) error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	rc.stats.mu.Lock()
	if requeue {
		rc.stats.Requeued++
	} else {
		rc.stats.Failed++
	}
	rc.stats.mu.Unlock()

	return nil
}

// GetQueueStats 获取队列统计
func (rc *RabbitMQClient) GetQueueStats(queueName string) map[string]interface{} {
	rc.mu.RLock()
	queue, exists := rc.queues[queueName]
	rc.mu.RUnlock()

	if !exists {
		return nil
	}

	rc.stats.mu.RLock()
	defer rc.stats.mu.RUnlock()

	return map[string]interface{}{
		"queue_name":    queue.Name,
		"durable":       queue.Durable,
		"published":     rc.stats.Published,
		"consumed":      rc.stats.Consumed,
		"failed":        rc.stats.Failed,
		"acknowledged":  rc.stats.Acknowledged,
		"requeued":      rc.stats.Requeued,
	}
}

// PurgeQueue 清空队列
func (rc *RabbitMQClient) PurgeQueue(ctx context.Context, queueName string) (int, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.queues[queueName]; !exists {
		return 0, fmt.Errorf("queue %s not found", queueName)
	}

	// 模拟清空队列
	return 0, nil
}

// DeleteQueue 删除队列
func (rc *RabbitMQClient) DeleteQueue(ctx context.Context, queueName string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.queues[queueName]; !exists {
		return fmt.Errorf("queue %s not found", queueName)
	}

	delete(rc.queues, queueName)
	delete(rc.consumers, queueName)

	return nil
}

// DeleteExchange 删除交换机
func (rc *RabbitMQClient) DeleteExchange(ctx context.Context, exchangeName string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if _, exists := rc.exchanges[exchangeName]; !exists {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}

	delete(rc.exchanges, exchangeName)

	return nil
}

// GetHealth 获取健康状态
func (rc *RabbitMQClient) GetHealth(ctx context.Context) map[string]interface{} {
	rc.mu.RLock()
	isConnected := rc.isConnected
	queueCount := len(rc.queues)
	exchangeCount := len(rc.exchanges)
	consumerCount := len(rc.consumers)
	rc.mu.RUnlock()

	rc.stats.mu.RLock()
	defer rc.stats.mu.RUnlock()

	return map[string]interface{}{
		"connected":      isConnected,
		"queue_count":    queueCount,
		"exchange_count": exchangeCount,
		"consumer_count": consumerCount,
		"published":      rc.stats.Published,
		"consumed":       rc.stats.Consumed,
		"failed":         rc.stats.Failed,
	}
}

