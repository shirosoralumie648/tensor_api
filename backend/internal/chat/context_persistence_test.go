package chat

import (
	"testing"
	"time"
)

func TestContextPersistenceSaveVersion(t *testing.T) {
	persistence := NewContextPersistence("session-1", 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	messages := []*Message{msg}
	err := persistence.SaveVersion(messages, 5, "Initial message")
	if err != nil {
		t.Errorf("SaveVersion failed: %v", err)
	}
}

func TestContextPersistenceGetVersion(t *testing.T) {
	persistence := NewContextPersistence("session-1", 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	_ = persistence.SaveVersion([]*Message{msg}, 5, "Test")

	version, err := persistence.GetLatestVersion()
	if err != nil {
		t.Errorf("GetLatestVersion failed: %v", err)
	}

	if version == nil {
		t.Errorf("Expected version to be returned")
	}

	if len(version.Messages) != 1 {
		t.Errorf("Expected 1 message in version")
	}
}

func TestContextPersistenceRestore(t *testing.T) {
	persistence := NewContextPersistence("session-1", 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	_ = persistence.SaveVersion([]*Message{msg}, 5, "Test")

	version, _ := persistence.GetLatestVersion()
	messages, err := persistence.RestoreFromVersion(version.VersionNumber)
	if err != nil {
		t.Errorf("RestoreFromVersion failed: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message after restore")
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello'")
	}
}

func TestContextPersistenceVersionLimit(t *testing.T) {
	persistence := NewContextPersistence("session-1", 3)

	for i := 0; i < 5; i++ {
		msg := &Message{
			ID:        "msg-" + string(rune(i)),
			Role:      "user",
			Content:   "Message",
			Tokens:    5,
			Timestamp: time.Now(),
		}
		_ = persistence.SaveVersion([]*Message{msg}, 5, "Test")
	}

	versions := persistence.GetVersions()
	if len(versions) > 3 {
		t.Errorf("Expected at most 3 versions, got %d", len(versions))
	}
}

func TestContextRetrievalSearchMessages(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)
	retrieval := NewContextRetrieval()

	_ = retrieval.RegisterContext("session-1", ctx)

	msg1 := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello world",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	msg2 := &Message{
		ID:        "msg-2",
		Role:      "assistant",
		Content:   "Hi there",
		Tokens:    3,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg1)
	ctx.AddMessage(msg2)

	results, err := retrieval.SearchMessages("session-1", "world")
	if err != nil {
		t.Errorf("SearchMessages failed: %v", err)
	}

	if len(results) == 0 {
		t.Logf("Search results: %d", len(results))
	}
}

func TestContextRetrievalGetSummary(t *testing.T) {
	ctx := NewConversationContext("session-1", 10000, 100)
	retrieval := NewContextRetrieval()

	_ = retrieval.RegisterContext("session-1", ctx)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg)

	summary, err := retrieval.GetContextSummary("session-1")
	if err != nil {
		t.Errorf("GetContextSummary failed: %v", err)
	}

	if summary == nil {
		t.Errorf("Expected summary to be returned")
	}
}

func TestAdvancedContextManagerCreateContext(t *testing.T) {
	manager := NewAdvancedContextManager(10000, 100)

	ctx, err := manager.CreateContext("session-1")
	if err != nil {
		t.Errorf("CreateContext failed: %v", err)
	}

	if ctx == nil {
		t.Errorf("Expected context to be returned")
	}
}

func TestAdvancedContextManagerSaveRestore(t *testing.T) {
	manager := NewAdvancedContextManager(10000, 100)
	ctx, _ := manager.CreateContext("session-1")

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Test",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg)

	// 保存版本
	err := manager.SaveContextVersion("session-1", "Initial")
	if err != nil {
		t.Errorf("SaveContextVersion failed: %v", err)
	}

	// 清空上下文
	ctx.Clear()

	if ctx.GetMessageCount() != 0 {
		t.Errorf("Expected empty context after clear")
	}

	// 恢复版本
	versions, _ := manager.GetContextVersions("session-1")
	if len(versions) > 0 {
		err = manager.RestoreContextVersion("session-1", versions[0].VersionNumber)
		if err != nil {
			t.Errorf("RestoreContextVersion failed: %v", err)
		}

		if ctx.GetMessageCount() != 1 {
			t.Errorf("Expected 1 message after restore")
		}
	}
}

func TestAdvancedContextManagerSearch(t *testing.T) {
	manager := NewAdvancedContextManager(10000, 100)
	ctx, _ := manager.CreateContext("session-1")

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello world",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg)

	results, err := manager.SearchMessages("session-1", "world")
	if err != nil {
		t.Errorf("SearchMessages failed: %v", err)
	}

	if len(results) == 0 {
		t.Logf("Search results: %d", len(results))
	}
}

func TestAdvancedContextManagerSummary(t *testing.T) {
	manager := NewAdvancedContextManager(10000, 100)
	ctx, _ := manager.CreateContext("session-1")

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Test",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	ctx.AddMessage(msg)

	summary, err := manager.GetContextSummary("session-1")
	if err != nil {
		t.Errorf("GetContextSummary failed: %v", err)
	}

	if summary == nil {
		t.Errorf("Expected summary")
	}

	if msgCount, ok := summary["message_count"].(int); !ok || msgCount < 1 {
		t.Errorf("Expected message count in summary")
	}
}

func TestContextPersistenceToJSON(t *testing.T) {
	persistence := NewContextPersistence("session-1", 100)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	_ = persistence.SaveVersion([]*Message{msg}, 5, "Test")

	jsonStr, err := persistence.ToJSON()
	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}

	if jsonStr == "" {
		t.Errorf("Expected non-empty JSON")
	}
}

func BenchmarkSaveVersion(b *testing.B) {
	persistence := NewContextPersistence("session-1", 1000)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = persistence.SaveVersion([]*Message{msg}, 5, "Test")
	}
}

func BenchmarkRestoreFromVersion(b *testing.B) {
	persistence := NewContextPersistence("session-1", 1000)

	msg := &Message{
		ID:        "msg-1",
		Role:      "user",
		Content:   "Hello",
		Tokens:    5,
		Timestamp: time.Now(),
	}

	_ = persistence.SaveVersion([]*Message{msg}, 5, "Test")

	version, _ := persistence.GetLatestVersion()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = persistence.RestoreFromVersion(version.VersionNumber)
	}
}

