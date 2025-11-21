package chat

import (
	"testing"
	"time"
)

func TestConversationContextCreation(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)

	if ctx.SessionID != "session-1" {
		t.Errorf("Expected session ID session-1")
	}

	if ctx.maxContextSize != 10000 {
		t.Errorf("Expected max context size 10000")
	}
}

func TestConversationContextAddMessage(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	err := ctx.AddMessage(msg)
	if err != nil {
		t.Errorf("AddMessage failed: %v", err)
	}

	if ctx.GetMessageCount() != 1 {
		t.Errorf("Expected 1 message")
	}

	if ctx.GetTotalTokens() != 5 {
		t.Errorf("Expected 5 tokens")
	}
}

func TestConversationContextGetMessages(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)

	msg1 := &Message{ID: "msg-1", Role: "user", Content: "Hello", Tokens: 5, Timestamp: time.Now()}
	msg2 := &Message{ID: "msg-2", Role: "assistant", Content: "Hi", Tokens: 3, Timestamp: time.Now()}

	ctx.AddMessage(msg1)
	ctx.AddMessage(msg2)

	messages := ctx.GetMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages")
	}
}

func TestConversationContextTruncate(t *testing.T) {
	ctx := NewConversationContext("session-1", 100, 100)

	// 添加消息使总 token 数接近限制
	for i := 0; i < 20; i++ {
		msg := &Message{
			ID:        "msg-" + string(rune(i)),
			Role:      "user",
			Content:   "Message",
			Tokens:    5,
			Timestamp: time.Now(),
		}
		_ = ctx.AddMessage(msg)
	}

	// 添加一个大消息，应该触发截断
	largeMsg := &Message{
		ID:        "large",
		Role:      "user",
		Content:   "Large message",
		Tokens:    50,
		Timestamp: time.Now(),
	}

	err := ctx.AddMessage(largeMsg)
	if err != nil {
		t.Errorf("AddMessage failed: %v", err)
	}

	// 检查总 token 数不超过限制
	if ctx.GetTotalTokens() > 100 {
		t.Errorf("Expected total tokens <= 100, got %d", ctx.GetTotalTokens())
	}
}

func TestConversationContextMaxRounds(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 5)

	// 添加 5 轮对话（10 条消息）
	for i := 0; i < 5; i++ {
		userMsg := &Message{
			ID:        "user-" + string(rune(i)),
			Role:      "user",
			Content:   "Message",
			Tokens:    5,
			Timestamp: time.Now(),
		}
		assistantMsg := &Message{
			ID:        "assistant-" + string(rune(i)),
			Role:      "assistant",
			Content:   "Response",
			Tokens:    5,
			Timestamp: time.Now(),
		}

		ctx.AddMessage(userMsg)
		ctx.AddMessage(assistantMsg)
	}

	if ctx.IsMaxRoundsReached() {
		t.Errorf("Expected max rounds not reached")
	}

	// 再添加一条 assistant 消息应该达到最大轮数
	extraMsg := &Message{
		ID:        "extra",
		Role:      "assistant",
		Content:   "Extra",
		Tokens:    5,
		Timestamp: time.Now(),
	}
	ctx.AddMessage(extraMsg)

	if !ctx.IsMaxRoundsReached() {
		t.Errorf("Expected max rounds reached")
	}
}

func TestConversationContextClear(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg)
	ctx.Clear()

	if ctx.GetMessageCount() != 0 {
		t.Errorf("Expected 0 messages after clear")
	}

	if ctx.GetTotalTokens() != 0 {
		t.Errorf("Expected 0 tokens after clear")
	}
}

func TestConversationContextStatistics(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg)

	stats := ctx.GetStatistics()
	if stats == nil {
		t.Errorf("Statistics should not be nil")
	}

	if msgCount, ok := stats["message_count"].(int64); !ok || msgCount != 1 {
		t.Errorf("Expected message count 1")
	}
}

func TestContextManagerCreateContext(t *testing.T) {
	manager := NewContextManager(10000, 100)

	ctx, err := manager.CreateContext("session-1")
	if err != nil {
		t.Errorf("CreateContext failed: %v", err)
	}

	if ctx == nil {
		t.Errorf("Expected context to be returned")
	}
}

func TestContextManagerGetContext(t *testing.T) {
	manager := NewContextManager(10000, 100)
	manager.CreateContext("session-1")

	ctx, err := manager.GetContext("session-1")
	if err != nil {
		t.Errorf("GetContext failed: %v", err)
	}

	if ctx.SessionID != "session-1" {
		t.Errorf("Expected session ID session-1")
	}
}

func TestContextManagerDeleteContext(t *testing.T) {
	manager := NewContextManager(10000, 100)
	manager.CreateContext("session-1")

	err := manager.DeleteContext("session-1")
	if err != nil {
		t.Errorf("DeleteContext failed: %v", err)
	}

	_, err = manager.GetContext("session-1")
	if err == nil {
		t.Errorf("Expected error for deleted context")
	}
}

func TestContextManagerSummarize(t *testing.T) {
	manager := NewContextManager(10000, 100)
	ctx, _ := manager.CreateContext("session-1")

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}
	ctx.AddMessage(msg)

	summary, err := manager.SummarizeContext("session-1")
	if err != nil {
		t.Errorf("SummarizeContext failed: %v", err)
	}

	if summary == "" {
		t.Errorf("Expected non-empty summary")
	}
}

func BenchmarkAddMessage(b *testing.B) {
	ctx := NewConversationContext("session-1", 1000000, 1000)
	msg := &Message{
		ID:        "msg",
		Role:      "user",
		Content:   "Message",
		Tokens:    10,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.AddMessage(msg)
	}
}

func BenchmarkGetMessages(b *testing.B) {
	ctx := NewConversationContext("session-1", 1000000, 1000)
	for i := 0; i < 100; i++ {
		msg := &Message{
			ID:        "msg-" + string(rune(i)),
			Role:      "user",
			Content:   "Message",
			Tokens:    10,
			Timestamp: time.Now(),
		}
		_ = ctx.AddMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.GetMessages()
	}
}
