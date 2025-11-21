package queue

import (
	"context"
	"testing"
	"time"
)

func TestRabbitMQClient(t *testing.T) {
	config := &RabbitMQConfig{
		URL:            "amqp://guest:guest@localhost:5672/",
		ClusterMode:    false,
		PrefetchCount:  10,
		MaxRetries:     3,
		ConnectTimeout: 5 * time.Second,
		Heartbeat:      60 * time.Second,
	}

	client, err := NewRabbitMQClient(config)
	if err != nil {
		t.Fatalf("NewRabbitMQClient failed: %v", err)
	}

	ctx := context.Background()

	// 测试连接
	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// 测试声明队列
	queueConfig := &QueueConfig{
		Name:    "test_queue",
		Durable: true,
	}

	err = client.DeclareQueue(ctx, queueConfig)
	if err != nil {
		t.Fatalf("DeclareQueue failed: %v", err)
	}

	// 测试声明交换机
	exchangeConfig := &ExchangeConfig{
		Name:    "test_exchange",
		Kind:    "direct",
		Durable: true,
	}

	err = client.DeclareExchange(ctx, exchangeConfig)
	if err != nil {
		t.Fatalf("DeclareExchange failed: %v", err)
	}

	// 测试绑定队列
	err = client.BindQueue(ctx, "test_queue", "test_exchange", "test_routing_key")
	if err != nil {
		t.Fatalf("BindQueue failed: %v", err)
	}

	// 测试发布消息
	message := &Message{
		Body:        []byte("test message"),
		ContentType: "text/plain",
		Headers:     make(map[string]string),
	}

	err = client.Publish(ctx, "test_exchange", "test_routing_key", message)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// 测试消费消息
	msgChan, err := client.Consume(ctx, "test_queue", false)
	if err != nil {
		t.Fatalf("Consume failed: %v", err)
	}

	if msgChan == nil {
		t.Error("Message channel should not be nil")
	}

	// 测试断开连接
	err = client.Disconnect()
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("Client should not be connected after disconnect")
	}
}

func TestRabbitMQHealth(t *testing.T) {
	config := &RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	}

	client, _ := NewRabbitMQClient(config)
	ctx := context.Background()

	client.Connect(ctx)
	defer client.Disconnect()

	queueConfig := &QueueConfig{
		Name:    "health_test_queue",
		Durable: true,
	}

	client.DeclareQueue(ctx, queueConfig)

	health := client.GetHealth(ctx)

	if health["connected"] != true {
		t.Error("Client should be connected")
	}

	if health["queue_count"].(int) < 1 {
		t.Error("Should have at least 1 queue")
	}
}

func TestAsyncQueue(t *testing.T) {
	config := &RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	}

	client, _ := NewRabbitMQClient(config)
	client.Connect(context.Background())
	defer client.Disconnect()

	asyncQueue := NewAsyncQueue(client, 5)

	// 测试提交任务
	task := &Task{
		Type:   TaskBilling,
		Status: StatusPending,
		Payload: map[string]interface{}{
			"user_id": "user123",
			"amount":  100.0,
		},
	}

	ctx := context.Background()
	err := asyncQueue.Submit(ctx, task)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if task.ID == "" {
		t.Error("Task ID should not be empty")
	}

	// 等待任务完成
	time.Sleep(100 * time.Millisecond)

	// 测试获取任务
	retrievedTask, err := asyncQueue.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if retrievedTask == nil {
		t.Error("Retrieved task should not be nil")
	}

	// 测试统计
	stats := asyncQueue.GetStats()
	if stats["created"].(int64) != 1 {
		t.Errorf("Expected 1 created task, got %v", stats["created"])
	}
}

func TestDeadLetterQueue(t *testing.T) {
	dlq := NewDeadLetterQueue()

	// 添加任务
	task1 := &Task{
		ID:   "task1",
		Type: TaskBilling,
		Status: StatusFailed,
		Error: "test error",
	}

	dlq.Add(task1)

	if dlq.Size() != 1 {
		t.Errorf("Expected DLQ size 1, got %d", dlq.Size())
	}

	// 获取任务
	retrieved, err := dlq.Get("task1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != "task1" {
		t.Errorf("Expected task1, got %s", retrieved.ID)
	}

	// 重试任务
	_, err = dlq.Retry("task1")
	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}

	if dlq.Size() != 0 {
		t.Errorf("Expected DLQ size 0 after retry, got %d", dlq.Size())
	}
}

func TestTaskConsumer(t *testing.T) {
	config := &RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	}

	client, _ := NewRabbitMQClient(config)
	asyncQueue := NewAsyncQueue(client, 5)
	consumer := NewTaskConsumer(asyncQueue)

	if consumer.IsRunning() {
		t.Error("Consumer should not be running initially")
	}

	ctx := context.Background()

	// 启动消费者
	err := consumer.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !consumer.IsRunning() {
		t.Error("Consumer should be running")
	}

	// 停止消费者
	err = consumer.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if consumer.IsRunning() {
		t.Error("Consumer should not be running after stop")
	}
}

func TestMultipleTasks(t *testing.T) {
	config := &RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	}

	client, _ := NewRabbitMQClient(config)
	client.Connect(context.Background())
	defer client.Disconnect()

	asyncQueue := NewAsyncQueue(client, 10)

	ctx := context.Background()

	// 提交多个任务
	for i := 0; i < 5; i++ {
		task := &Task{
			Type:   TaskBilling,
			Status: StatusPending,
			Payload: map[string]interface{}{
				"task_index": i,
			},
		}

		err := asyncQueue.Submit(ctx, task)
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}
	}

	// 等待任务处理
	time.Sleep(200 * time.Millisecond)

	stats := asyncQueue.GetStats()
	if stats["created"].(int64) != 5 {
		t.Errorf("Expected 5 created tasks, got %v", stats["created"])
	}
}

func BenchmarkPublish(b *testing.B) {
	config := &RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	}

	client, _ := NewRabbitMQClient(config)
	client.Connect(context.Background())
	defer client.Disconnect()

	ctx := context.Background()

	queueConfig := &QueueConfig{
		Name:    "bench_queue",
		Durable: true,
	}

	client.DeclareQueue(ctx, queueConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message := &Message{
			Body: []byte("benchmark message"),
		}

		client.Publish(ctx, "", "", message)
	}
}

func BenchmarkAsyncQueueSubmit(b *testing.B) {
	config := &RabbitMQConfig{
		URL: "amqp://guest:guest@localhost:5672/",
	}

	client, _ := NewRabbitMQClient(config)
	client.Connect(context.Background())
	defer client.Disconnect()

	asyncQueue := NewAsyncQueue(client, 10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			Type:   TaskBilling,
			Status: StatusPending,
			Payload: map[string]interface{}{
				"user_id": "user123",
			},
		}

		asyncQueue.Submit(ctx, task)
	}
}

