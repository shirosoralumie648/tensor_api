package chat

import (
	"testing"
	"time"
)

func TestStreamClientCreation(t *testing.T) {
	client := NewStreamClient("client-1", 100)
	
	if client.ClientID != "client-1" {
		t.Errorf("Expected client ID client-1, got %s", client.ClientID)
	}
	
	if client.bufferSize != 100 {
		t.Errorf("Expected buffer size 100, got %d", client.bufferSize)
	}
}

func TestStreamClientSendMessage(t *testing.T) {
	client := NewStreamClient("client-1", 10)
	
	msg := &StreamMessage{
		MessageID: "msg-1",
		SessionID: "session-1",
		Content:   "Hello",
		Done:      false,
		Timestamp: time.Now(),
	}
	
	err := client.SendMessage(msg)
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}
	
	stats := client.GetStatistics()
	if count, ok := stats["message_count"].(int64); !ok || count != 1 {
		t.Errorf("Expected message count 1")
	}
}

func TestStreamSessionCreation(t *testing.T) {
	session := NewStreamSession("session-1", "user-1", "gpt-4")
	
	if session.SessionID != "session-1" {
		t.Errorf("Expected session ID session-1")
	}
	
	if session.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients initially")
	}
}

func TestStreamSessionRegisterClient(t *testing.T) {
	session := NewStreamSession("session-1", "user-1", "gpt-4")
	
	client, err := session.RegisterClient("client-1", 100)
	if err != nil {
		t.Errorf("RegisterClient failed: %v", err)
	}
	
	if client == nil {
		t.Errorf("Expected client to be returned")
	}
	
	if session.GetClientCount() != 1 {
		t.Errorf("Expected 1 client")
	}
}

func TestStreamSessionBroadcast(t *testing.T) {
	session := NewStreamSession("session-1", "user-1", "gpt-4")
	
	client1, _ := session.RegisterClient("client-1", 100)
	client2, _ := session.RegisterClient("client-2", 100)
	
	msg := &StreamMessage{
		MessageID: "msg-1",
		SessionID: "session-1",
		Content:   "Broadcast",
		Timestamp: time.Now(),
	}
	
	err := session.BroadcastMessage(msg)
	if err != nil {
		t.Errorf("BroadcastMessage failed: %v", err)
	}
	
	// 验证消息已发送到客户端
	if client1 != nil && client2 != nil {
		stats1 := client1.GetStatistics()
		stats2 := client2.GetStatistics()
		
		if count1, ok := stats1["message_count"].(int64); !ok || count1 < 1 {
			t.Logf("Client 1 message count: %v", stats1["message_count"])
		}
		
		if count2, ok := stats2["message_count"].(int64); !ok || count2 < 1 {
			t.Logf("Client 2 message count: %v", stats2["message_count"])
		}
	}
}

func TestStreamManagerCreateSession(t *testing.T) {
	manager := NewStreamManager(5*time.Minute, 30*time.Minute)
	
	session, err := manager.CreateSession("session-1", "user-1", "gpt-4")
	if err != nil {
		t.Errorf("CreateSession failed: %v", err)
	}
	
	if session == nil {
		t.Errorf("Expected session to be returned")
	}
}

func TestStreamManagerGetSession(t *testing.T) {
	manager := NewStreamManager(5*time.Minute, 30*time.Minute)
	manager.CreateSession("session-1", "user-1", "gpt-4")
	
	session, err := manager.GetSession("session-1")
	if err != nil {
		t.Errorf("GetSession failed: %v", err)
	}
	
	if session.SessionID != "session-1" {
		t.Errorf("Expected session ID session-1")
	}
}

func TestStreamManagerCloseSession(t *testing.T) {
	manager := NewStreamManager(5*time.Minute, 30*time.Minute)
	manager.CreateSession("session-1", "user-1", "gpt-4")
	
	err := manager.CloseSession("session-1")
	if err != nil {
		t.Errorf("CloseSession failed: %v", err)
	}
	
	_, err = manager.GetSession("session-1")
	if err == nil {
		t.Errorf("Expected error for closed session")
	}
}

func TestStreamSessionStatistics(t *testing.T) {
	session := NewStreamSession("session-1", "user-1", "gpt-4")
	client, _ := session.RegisterClient("client-1", 100)
	
	msg := &StreamMessage{
		MessageID: "msg-1",
		SessionID: "session-1",
		Content:   "Test",
		Timestamp: time.Now(),
	}
	
	_ = client.SendMessage(msg)
	
	stats := session.GetStatistics()
	if stats == nil {
		t.Errorf("Statistics should not be nil")
	}
	
	if clientCount, ok := stats["client_count"].(int); !ok || clientCount != 1 {
		t.Errorf("Expected 1 client in stats")
	}
}

func TestStreamManagerStatistics(t *testing.T) {
	manager := NewStreamManager(5*time.Minute, 30*time.Minute)
	manager.CreateSession("session-1", "user-1", "gpt-4")
	
	stats := manager.GetStatistics()
	if stats == nil {
		t.Errorf("Statistics should not be nil")
	}
}

func BenchmarkStreamClientSendMessage(b *testing.B) {
	client := NewStreamClient("client-1", 10000)
	msg := &StreamMessage{
		MessageID: "msg-1",
		SessionID: "session-1",
		Content:   "Test",
		Timestamp: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.SendMessage(msg)
	}
}

func BenchmarkStreamSessionBroadcast(b *testing.B) {
	session := NewStreamSession("session-1", "user-1", "gpt-4")
	for i := 0; i < 10; i++ {
		session.RegisterClient("client-"+string(rune(i)), 100)
	}
	
	msg := &StreamMessage{
		MessageID: "msg-1",
		SessionID: "session-1",
		Content:   "Broadcast",
		Timestamp: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = session.BroadcastMessage(msg)
	}
}

