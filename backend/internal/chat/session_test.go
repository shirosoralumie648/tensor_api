package chat

import (
	"testing"
	"time"
)

func TestSessionCreation(t *testing.T) {
	session := NewSession("session-1", 1, "Test Session")

	if session.ID != "session-1" {
		t.Errorf("Expected session-1")
	}

	if session.UserID != 1 {
		t.Errorf("Expected user ID 1")
	}

	if !session.IsActive() {
		t.Errorf("Expected session to be active")
	}
}

func TestSessionAddTag(t *testing.T) {
	session := NewSession("session-1", 1, "Test Session")

	session.AddTag("important")

	if !session.HasTag("important") {
		t.Errorf("Expected important tag")
	}
}

func TestSessionRemoveTag(t *testing.T) {
	session := NewSession("session-1", 1, "Test Session")

	session.AddTag("important")
	session.RemoveTag("important")

	if session.HasTag("important") {
		t.Errorf("Expected tag to be removed")
	}
}

func TestSessionSetMetadata(t *testing.T) {
	session := NewSession("session-1", 1, "Test Session")

	session.SetMetadata("key", "value")

	value, exists := session.GetMetadata("key")
	if !exists || value != "value" {
		t.Errorf("Expected metadata value")
	}
}

func TestSessionUpdateTitle(t *testing.T) {
	session := NewSession("session-1", 1, "Original Title")

	session.UpdateTitle("New Title")

	if session.Title != "New Title" {
		t.Errorf("Expected New Title")
	}
}

func TestSessionIncrementMessageCount(t *testing.T) {
	session := NewSession("session-1", 1, "Test Session")

	session.IncrementMessageCount()
	session.IncrementMessageCount()

	if session.MessageCount != 2 {
		t.Errorf("Expected message count 2")
	}
}

func TestSessionAddTokens(t *testing.T) {
	session := NewSession("session-1", 1, "Test Session")

	session.AddTokens(100)
	session.AddTokens(50)

	if session.TotalTokens != 150 {
		t.Errorf("Expected total tokens 150")
	}
}

func TestSessionManagerCreateSession(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test Session")

	err := sm.CreateSession(session)
	if err != nil {
		t.Errorf("CreateSession failed: %v", err)
	}
}

func TestSessionManagerGetSession(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test Session")
	sm.CreateSession(session)

	retrieved, err := sm.GetSession("session-1")
	if err != nil {
		t.Errorf("GetSession failed: %v", err)
	}

	if retrieved.Title != "Test Session" {
		t.Errorf("Expected Test Session")
	}
}

func TestSessionManagerUpdateSession(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Original")
	sm.CreateSession(session)

	err := sm.UpdateSession("session-1", "Updated", "New Description")
	if err != nil {
		t.Errorf("UpdateSession failed: %v", err)
	}

	retrieved, _ := sm.GetSession("session-1")
	if retrieved.Title != "Updated" {
		t.Errorf("Expected Updated title")
	}
}

func TestSessionManagerSoftDelete(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test Session")
	sm.CreateSession(session)

	err := sm.SoftDeleteSession("session-1")
	if err != nil {
		t.Errorf("SoftDeleteSession failed: %v", err)
	}

	_, err = sm.GetSession("session-1")
	if err == nil {
		t.Errorf("Expected error when getting deleted session")
	}
}

func TestSessionManagerRestoreSession(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test Session")
	sm.CreateSession(session)

	sm.SoftDeleteSession("session-1")

	err := sm.RestoreSession("session-1")
	if err != nil {
		t.Errorf("RestoreSession failed: %v", err)
	}

	retrieved, err := sm.GetSession("session-1")
	if err != nil {
		t.Errorf("Expected to retrieve restored session")
	}

	if !retrieved.IsActive() {
		t.Errorf("Expected restored session to be active")
	}
}

func TestSessionManagerGetUserSessions(t *testing.T) {
	sm := NewSessionManager()

	for i := 0; i < 3; i++ {
		session := NewSession("session-"+string(rune(i)), 1, "Test")
		sm.CreateSession(session)
	}

	sessions := sm.GetUserSessions(1)

	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}

func TestSessionManagerSearchByTag(t *testing.T) {
	sm := NewSessionManager()

	session1 := NewSession("session-1", 1, "Test")
	session1.AddTag("work")
	sm.CreateSession(session1)

	session2 := NewSession("session-2", 1, "Test")
	session2.AddTag("personal")
	sm.CreateSession(session2)

	workSessions := sm.SearchByTag(1, "work")

	if len(workSessions) != 1 {
		t.Errorf("Expected 1 work session")
	}
}

func TestSessionManagerSearchByTitle(t *testing.T) {
	sm := NewSessionManager()

	session1 := NewSession("session-1", 1, "Project Alpha")
	sm.CreateSession(session1)

	session2 := NewSession("session-2", 1, "Project Beta")
	sm.CreateSession(session2)

	results := sm.SearchByTitle(1, "Alpha")

	if len(results) > 0 {
		t.Logf("Found %d sessions with keyword", len(results))
	}
}

func TestSessionManagerGetDeletedSessions(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test")
	sm.CreateSession(session)

	sm.SoftDeleteSession("session-1")

	deletedSessions := sm.GetDeletedSessions(1)

	if len(deletedSessions) != 1 {
		t.Errorf("Expected 1 deleted session")
	}
}

func TestSessionManagerPermanentlyDelete(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test")
	sm.CreateSession(session)

	err := sm.PermanentlyDeleteSession("session-1")
	if err != nil {
		t.Errorf("PermanentlyDeleteSession failed: %v", err)
	}

	_, err = sm.GetSessionIncludingDeleted("session-1")
	if err == nil {
		t.Errorf("Expected error when getting permanently deleted session")
	}
}

func TestSessionManagerGetStatistics(t *testing.T) {
	sm := NewSessionManager()

	for i := 0; i < 2; i++ {
		session := NewSession("session-"+string(rune(i)), 1, "Test")
		sm.CreateSession(session)
	}

	stats := sm.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if activeCount, ok := stats["active_count"].(int); !ok || activeCount != 2 {
		t.Errorf("Expected active_count to be 2")
	}
}

func TestSessionManagerPurgeOldDeletedSessions(t *testing.T) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test")
	sm.CreateSession(session)

	// 软删除会话
	sm.SoftDeleteSession("session-1")

	// 手动设置删除时间为45天前
	s, _ := sm.GetSessionIncludingDeleted("session-1")
	oldTime := time.Now().AddDate(0, 0, -45)
	s.mu.Lock()
	s.DeletedAt = &oldTime
	s.mu.Unlock()

	// 清理30天前的会话
	purgedCount := sm.PurgeOldDeletedSessions(30)

	if purgedCount != 1 {
		t.Errorf("Expected 1 purged session")
	}

	_, err := sm.GetSessionIncludingDeleted("session-1")
	if err == nil {
		t.Errorf("Expected purged session to be removed")
	}
}

func BenchmarkSessionCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSession("session-"+string(rune(i)), 1, "Test")
	}
}

func BenchmarkSessionManagerCreateSession(b *testing.B) {
	sm := NewSessionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := NewSession("session-"+string(rune(i)), 1, "Test")
		_ = sm.CreateSession(session)
	}
}

func BenchmarkSessionManagerGetSession(b *testing.B) {
	sm := NewSessionManager()

	session := NewSession("session-1", 1, "Test")
	sm.CreateSession(session)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.GetSession("session-1")
	}
}

func BenchmarkSessionManagerSearchByTitle(b *testing.B) {
	sm := NewSessionManager()

	for i := 0; i < 100; i++ {
		s := NewSession("session-"+string(rune(i)), 1, "Test Session")
		sm.CreateSession(s)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.SearchByTitle(1, "Test")
	}
}

