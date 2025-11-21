package token

import (
	"testing"
	"time"
)

// MockTokenStore 模拟 Token 存储
type MockTokenStore struct {
	tokens map[string]*TokenMetadata
}

func NewMockTokenStore() *MockTokenStore {
	return &MockTokenStore{
		tokens: make(map[string]*TokenMetadata),
	}
}

func (m *MockTokenStore) Create(metadata *TokenMetadata) error {
	m.tokens[metadata.ID] = metadata
	return nil
}

func (m *MockTokenStore) Update(metadata *TokenMetadata) error {
	m.tokens[metadata.ID] = metadata
	return nil
}

func (m *MockTokenStore) Get(id string) (*TokenMetadata, error) {
	if token, exists := m.tokens[id]; exists {
		return token, nil
	}
	return nil, ErrTokenNotFound
}

func (m *MockTokenStore) Delete(id string) error {
	delete(m.tokens, id)
	return nil
}

func (m *MockTokenStore) List(userID string) ([]*TokenMetadata, error) {
	var tokens []*TokenMetadata
	for _, token := range m.tokens {
		if token.UserID == userID {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (m *MockTokenStore) Expire(id string) error {
	if token, exists := m.tokens[id]; exists {
		token.Status = StatusExpired
		return nil
	}
	return ErrTokenNotFound
}

func (m *MockTokenStore) Rotate(oldID string, newMetadata *TokenMetadata) error {
	if token, exists := m.tokens[oldID]; exists {
		token.Status = StatusRotated
		m.tokens[newMetadata.ID] = newMetadata
		return nil
	}
	return ErrTokenNotFound
}

var ErrTokenNotFound = &struct{ msg string }{msg: "token not found"}

func TestGenerateToken(t *testing.T) {
	store := NewMockTokenStore()
	manager := NewTokenLifecycleManager(store, DefaultRotationPolicy)

	token, metadata, err := manager.GenerateToken("user123", TypeAPI, "test-token", 30)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	if metadata == nil {
		t.Error("Expected metadata to be returned")
	}

	if metadata.Status != StatusActive {
		t.Errorf("Expected status Active, got %s", metadata.Status)
	}
}

func TestRecordUsage(t *testing.T) {
	store := NewMockTokenStore()
	manager := NewTokenLifecycleManager(store, DefaultRotationPolicy)

	token, _, err := manager.GenerateToken("user123", TypeAPI, "test-token", 30)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	err = manager.RecordUsage(token)
	if err != nil {
		t.Fatalf("RecordUsage failed: %v", err)
	}

	metadata, err := manager.GetToken(token)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if metadata.UsageCount != 1 {
		t.Errorf("Expected usage count 1, got %d", metadata.UsageCount)
	}
}

func TestRotateToken(t *testing.T) {
	store := NewMockTokenStore()
	manager := NewTokenLifecycleManager(store, DefaultRotationPolicy)

	token, _, err := manager.GenerateToken("user123", TypeAPI, "test-token", 30)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	newToken, newMetadata, err := manager.RotateToken(token, 30)
	if err != nil {
		t.Fatalf("RotateToken failed: %v", err)
	}

	if newToken == token {
		t.Error("Expected new token to be different")
	}

	if newMetadata.Status != StatusActive {
		t.Errorf("Expected new token status Active, got %s", newMetadata.Status)
	}

	oldMetadata, _ := manager.GetToken(token)
	if oldMetadata.Status != StatusRotated {
		t.Errorf("Expected old token status Rotated, got %s", oldMetadata.Status)
	}
}

func TestExpireToken(t *testing.T) {
	store := NewMockTokenStore()
	manager := NewTokenLifecycleManager(store, DefaultRotationPolicy)

	token, _, err := manager.GenerateToken("user123", TypeAPI, "test-token", 30)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	err = manager.Expire(token)
	if err != nil {
		t.Fatalf("Expire failed: %v", err)
	}

	metadata, _ := manager.GetToken(token)
	if metadata.Status != StatusExpired {
		t.Errorf("Expected status Expired, got %s", metadata.Status)
	}
}

func TestRevokeToken(t *testing.T) {
	store := NewMockTokenStore()
	manager := NewTokenLifecycleManager(store, DefaultRotationPolicy)

	token, _, err := manager.GenerateToken("user123", TypeAPI, "test-token", 30)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	err = manager.Revoke(token)
	if err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	metadata, _ := manager.GetToken(token)
	if metadata.Status != StatusRevoked {
		t.Errorf("Expected status Revoked, got %s", metadata.Status)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	store := NewMockTokenStore()
	manager := NewTokenLifecycleManager(store, DefaultRotationPolicy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GenerateToken("user123", TypeAPI, "test-token", 30)
	}
}


