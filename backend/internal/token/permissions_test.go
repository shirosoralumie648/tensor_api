package token

import (
	"testing"
)

// MockPermissionStore 模拟权限存储
type MockPermissionStore struct {
	permissions map[string]*TokenPermissions
}

func NewMockPermissionStore() *MockPermissionStore {
	return &MockPermissionStore{
		permissions: make(map[string]*TokenPermissions),
	}
}

func (m *MockPermissionStore) Create(permissions *TokenPermissions) error {
	m.permissions[permissions.TokenID] = permissions
	return nil
}

func (m *MockPermissionStore) Update(permissions *TokenPermissions) error {
	m.permissions[permissions.TokenID] = permissions
	return nil
}

func (m *MockPermissionStore) Get(id string) (*TokenPermissions, error) {
	if perms, exists := m.permissions[id]; exists {
		return perms, nil
	}
	return nil, ErrPermissionsNotFound
}

func (m *MockPermissionStore) GetByToken(tokenID string) (*TokenPermissions, error) {
	if perms, exists := m.permissions[tokenID]; exists {
		return perms, nil
	}
	return nil, ErrPermissionsNotFound
}

func (m *MockPermissionStore) Delete(id string) error {
	delete(m.permissions, id)
	return nil
}

var ErrPermissionsNotFound = &struct{ msg string }{msg: "permissions not found"}

func TestCreatePermissions(t *testing.T) {
	store := NewMockPermissionStore()
	manager := NewPermissionManager(store)

	perms, err := manager.CreatePermissions("token123")
	if err != nil {
		t.Fatalf("CreatePermissions failed: %v", err)
	}

	if perms == nil {
		t.Error("Expected permissions to be created")
	}

	if perms.TokenID != "token123" {
		t.Errorf("Expected token ID token123, got %s", perms.TokenID)
	}
}

func TestSetModelWhitelist(t *testing.T) {
	store := NewMockPermissionStore()
	manager := NewPermissionManager(store)

	manager.CreatePermissions("token123")

	err := manager.SetModelWhitelist("token123", []string{"gpt-4", "claude-3"})
	if err != nil {
		t.Fatalf("SetModelWhitelist failed: %v", err)
	}

	perms, _ := manager.GetPermissions("token123")
	if len(perms.ModelWhitelist) != 2 {
		t.Errorf("Expected 2 models, got %d", len(perms.ModelWhitelist))
	}
}

func TestCheckModelPermission(t *testing.T) {
	store := NewMockPermissionStore()
	manager := NewPermissionManager(store)

	manager.CreatePermissions("token123")
	manager.SetModelWhitelist("token123", []string{"gpt-4", "claude-3"})

	allowed, err := manager.CheckModelPermission("token123", "gpt-4")
	if err != nil {
		t.Fatalf("CheckModelPermission failed: %v", err)
	}

	if !allowed {
		t.Error("Expected gpt-4 to be allowed")
	}

	allowed, err = manager.CheckModelPermission("token123", "gemini")
	if err != nil {
		t.Fatalf("CheckModelPermission failed: %v", err)
	}

	if allowed {
		t.Error("Expected gemini to be denied")
	}
}

func TestCheckIPPermission(t *testing.T) {
	store := NewMockPermissionStore()
	manager := NewPermissionManager(store)

	manager.CreatePermissions("token123")
	manager.SetIPWhitelist("token123", []string{"192.168.1.1", "10.0.0.0/8"})

	allowed, err := manager.CheckIPPermission("token123", "192.168.1.1")
	if err != nil {
		t.Fatalf("CheckIPPermission failed: %v", err)
	}

	if !allowed {
		t.Error("Expected 192.168.1.1 to be allowed")
	}

	allowed, err = manager.CheckIPPermission("token123", "10.1.1.1")
	if err != nil {
		t.Fatalf("CheckIPPermission failed: %v", err)
	}

	if !allowed {
		t.Error("Expected 10.1.1.1 to be allowed (CIDR)")
	}
}

func TestSetRateLimit(t *testing.T) {
	store := NewMockPermissionStore()
	manager := NewPermissionManager(store)

	manager.CreatePermissions("token123")
	err := manager.SetRateLimit("token123", 500)
	if err != nil {
		t.Fatalf("SetRateLimit failed: %v", err)
	}

	perms, _ := manager.GetPermissions("token123")
	if perms.RateLimit != 500 {
		t.Errorf("Expected rate limit 500, got %d", perms.RateLimit)
	}
}

func BenchmarkCheckModelPermission(b *testing.B) {
	store := NewMockPermissionStore()
	manager := NewPermissionManager(store)

	manager.CreatePermissions("token123")
	manager.SetModelWhitelist("token123", []string{"gpt-4", "claude-3"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CheckModelPermission("token123", "gpt-4")
	}
}


