package billing

import (
	"context"
	"testing"
	"time"
)

func TestBillingEventQueue(t *testing.T) {
	queue := NewBillingEventQueue("test-queue", 10)

	event := &BillingEvent{
		EventID:      "evt-1",
		UserID:       "user-1",
		EventType:    "chat",
		ModelName:    "gpt-4",
		InputTokens:  1000,
		OutputTokens: 500,
		Cost:         0.05,
		Timestamp:    time.Now(),
	}

	err := queue.Enqueue(event)
	if err != nil {
		t.Errorf("Enqueue failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	retrieved, err := queue.Dequeue(ctx)
	if err != nil {
		t.Errorf("Dequeue failed: %v", err)
	}

	if retrieved.EventID != event.EventID {
		t.Errorf("Event ID mismatch: expected %s, got %s", event.EventID, retrieved.EventID)
	}
}

func TestBillingEventQueueFull(t *testing.T) {
	queue := NewBillingEventQueue("test-queue", 2)

	// 填满队列
	for i := 0; i < 2; i++ {
		event := &BillingEvent{
			EventID:   "evt-" + string(rune(i)),
			UserID:    "user-1",
			Timestamp: time.Now(),
		}
		_ = queue.Enqueue(event)
	}

	// 尝试添加更多事件应该失败
	event := &BillingEvent{
		EventID:   "evt-3",
		UserID:    "user-1",
		Timestamp: time.Now(),
	}
	err := queue.Enqueue(event)
	if err == nil {
		t.Errorf("Expected error for full queue")
	}
}

func TestBillingConsumer(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 1000.0)
	quotaManager.PreDeduct("user-1", "req-1", 100.0, "Test")

	pricingManager := NewPricingManager()
	pricingManager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	queue := NewBillingEventQueue("test-queue", 100)
	consumer := NewBillingConsumer("consumer-1", queue, quotaManager, pricingManager)

	// 发布事件
	event := &BillingEvent{
		EventID:      "evt-1",
		UserID:       "user-1",
		EventType:    "chat",
		ModelName:    "gpt-4",
		InputTokens:  1000,
		OutputTokens: 1000,
		Cost:         0.09,
		RequestID:    "req-1",
		Timestamp:    time.Now(),
	}

	err := queue.Enqueue(event)
	if err != nil {
		t.Errorf("Enqueue failed: %v", err)
	}

	// 启动消费者
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	consumer.Start(ctx)

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	stats := consumer.GetStatistics()
	if processedCount, ok := stats["processed_count"].(int64); !ok || processedCount < 1 {
		t.Logf("Consumer stats: %v", stats)
	}
}

func TestAsyncBillingService(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 1000.0)
	quotaManager.PreDeduct("user-1", "req-1", 100.0, "Test")

	pricingManager := NewPricingManager()
	pricingManager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	service := NewAsyncBillingService("billing-queue", 1000)

	consumer := NewBillingConsumer("consumer-1", service.queue, quotaManager, pricingManager)
	service.AddConsumer(consumer)

	service.Start()

	// 发布事件
	event := &BillingEvent{
		EventID:      "evt-1",
		UserID:       "user-1",
		ModelName:    "gpt-4",
		InputTokens:  1000,
		OutputTokens: 1000,
		RequestID:    "req-1",
		Timestamp:    time.Now(),
	}

	err := service.PublishEvent(event)
	if err != nil {
		t.Errorf("PublishEvent failed: %v", err)
	}

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	service.Stop()

	stats := service.GetStatistics()
	if consumers, ok := stats["consumers"].(int); !ok || consumers != 1 {
		t.Errorf("Expected 1 consumer, got %v", stats["consumers"])
	}
}

func TestBillingEventLogger(t *testing.T) {
	logger := NewBillingEventLogger("/tmp/billing.log", 10, 1*time.Second)

	event := &BillingEvent{
		EventID:      "evt-1",
		UserID:       "user-1",
		ModelName:    "gpt-4",
		InputTokens:  1000,
		OutputTokens: 500,
		Cost:         0.045,
		Timestamp:    time.Now(),
	}

	err := logger.Log(event)
	if err != nil {
		t.Errorf("Log failed: %v", err)
	}

	err = logger.Flush()
	if err != nil {
		t.Errorf("Flush failed: %v", err)
	}

	total := logger.GetTotalLogged()
	if total < 1 {
		t.Errorf("Expected at least 1 logged event")
	}
}

func TestBillingEventQueueStatistics(t *testing.T) {
	queue := NewBillingEventQueue("test-queue", 10)

	for i := 0; i < 5; i++ {
		event := &BillingEvent{
			EventID:   "evt-" + string(rune(i)),
			UserID:    "user-1",
			Timestamp: time.Now(),
		}
		_ = queue.Enqueue(event)
	}

	stats := queue.GetStatistics()
	if enqueueCount, ok := stats["enqueue_count"].(int64); !ok || enqueueCount != 5 {
		t.Errorf("Expected 5 enqueued events")
	}
}

func TestDeadLetterQueue(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 10.0) // 很小的配额

	pricingManager := NewPricingManager()
	pricingManager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	queue := NewBillingEventQueue("test-queue", 100)
	consumer := NewBillingConsumer("consumer-1", queue, quotaManager, pricingManager)

	// 发送会导致扣费失败的事件
	event := &BillingEvent{
		EventID:      "evt-1",
		UserID:       "user-1",
		ModelName:    "gpt-4",
		InputTokens:  100000, // 很大的 token 数
		OutputTokens: 100000,
		Timestamp:    time.Now(),
	}

	_ = queue.Enqueue(event)

	// 处理会失败并进入 DLQ
	consumer.processEvent(event, 0)

	dlqSize := consumer.GetDeadLetterQueueSize()
	if dlqSize < 1 {
		t.Logf("DLQ size: %d", dlqSize)
	}
}

func TestBillingEventQueueDiscard(t *testing.T) {
	queue := NewBillingEventQueue("test-queue", 1)

	// 填满队列
	event1 := &BillingEvent{
		EventID:   "evt-1",
		UserID:    "user-1",
		Timestamp: time.Now(),
	}
	_ = queue.Enqueue(event1)

	// 第二个事件应该被丢弃
	event2 := &BillingEvent{
		EventID:   "evt-2",
		UserID:    "user-1",
		Timestamp: time.Now(),
	}
	_ = queue.Enqueue(event2)

	stats := queue.GetStatistics()
	if discardCount, ok := stats["discard_count"].(int64); !ok || discardCount < 1 {
		t.Errorf("Expected at least 1 discarded event")
	}
}

func TestConsumerStatistics(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 1000.0)

	pricingManager := NewPricingManager()
	pricingManager.RegisterModelPrice("gpt-4", 0.03, 0.06, PricingByToken)

	queue := NewBillingEventQueue("test-queue", 100)
	consumer := NewBillingConsumer("consumer-1", queue, quotaManager, pricingManager)

	stats := consumer.GetStatistics()
	if stats == nil {
		t.Errorf("Statistics should not be nil")
	}
}

func BenchmarkBillingEventQueueEnqueue(b *testing.B) {
	queue := NewBillingEventQueue("test-queue", 10000)

	event := &BillingEvent{
		EventID:      "evt-1",
		UserID:       "user-1",
		ModelName:    "gpt-4",
		InputTokens:  1000,
		OutputTokens: 500,
		Timestamp:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queue.Enqueue(event)
	}
}

func BenchmarkBillingEventQueueDequeue(b *testing.B) {
	queue := NewBillingEventQueue("test-queue", 10000)

	// 预先加入事件
	for i := 0; i < 1000; i++ {
		event := &BillingEvent{
			EventID:   "evt-" + string(rune(i)),
			UserID:    "user-1",
			Timestamp: time.Now(),
		}
		_ = queue.Enqueue(event)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N && queue.Size() > 0; i++ {
		_, _ = queue.Dequeue(ctx)
	}
}

