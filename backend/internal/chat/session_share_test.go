package chat

import (
	"testing"
	"time"
)

func TestShareLinkCreation(t *testing.T) {
	link := &ShareLink{
		ID:         "link-1",
		LinkCode:   "abc123",
		SessionID:  "session-1",
		SharerID:   1,
		Name:       "Share",
		Permission: PermissionView,
		CreatedAt:  time.Now(),
		Enabled:    true,
	}

	if link.IsExpired() {
		t.Errorf("Expected link not to be expired")
	}

	if !link.IsAccessible() {
		t.Errorf("Expected link to be accessible")
	}
}

func TestShareLinkExpiry(t *testing.T) {
	expiredTime := time.Now().Add(-1 * time.Hour)
	link := &ShareLink{
		ID:        "link-1",
		LinkCode:  "abc123",
		ExpiresAt: &expiredTime,
		Enabled:   true,
	}

	if !link.IsExpired() {
		t.Errorf("Expected link to be expired")
	}
}

func TestShareLinkAccessCount(t *testing.T) {
	link := &ShareLink{
		ID:        "link-1",
		LinkCode:  "abc123",
		Enabled:   true,
	}

	link.IncrementAccessCount()
	link.IncrementAccessCount()

	if link.GetAccessCount() != 2 {
		t.Errorf("Expected access count 2")
	}
}

func TestShareManagerCreateLink(t *testing.T) {
	sm := NewShareManager()

	link, err := sm.CreateShareLink("session-1", 1, "Share", "Description", PermissionView, nil)
	if err != nil {
		t.Errorf("CreateShareLink failed: %v", err)
	}

	if link.Permission != PermissionView {
		t.Errorf("Expected PermissionView")
	}
}

func TestShareManagerGetLink(t *testing.T) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	retrieved, err := sm.GetShareLink(link.ID)
	if err != nil {
		t.Errorf("GetShareLink failed: %v", err)
	}

	if retrieved.Name != "Share" {
		t.Errorf("Expected Share name")
	}
}

func TestShareManagerGetLinkByCode(t *testing.T) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	retrieved, err := sm.GetShareLinkByCode(link.LinkCode)
	if err != nil {
		t.Errorf("GetShareLinkByCode failed: %v", err)
	}

	if retrieved.ID != link.ID {
		t.Errorf("Expected same link ID")
	}
}

func TestShareManagerAccessLink(t *testing.T) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	access, err := sm.AccessShareLink(link.ID, "192.168.1.1", "Mozilla", nil)
	if err != nil {
		t.Errorf("AccessShareLink failed: %v", err)
	}

	if access.VisitorIP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1")
	}

	if link.GetAccessCount() != 1 {
		t.Errorf("Expected access count 1")
	}
}

func TestShareManagerDisableLink(t *testing.T) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	err := sm.DisableShareLink(link.ID)
	if err != nil {
		t.Errorf("DisableShareLink failed: %v", err)
	}

	retrieved, _ := sm.GetShareLink(link.ID)
	if retrieved.IsAccessible() {
		t.Errorf("Expected link not to be accessible")
	}
}

func TestShareManagerDeleteLink(t *testing.T) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	err := sm.DeleteShareLink(link.ID)
	if err != nil {
		t.Errorf("DeleteShareLink failed: %v", err)
	}

	_, err = sm.GetShareLink(link.ID)
	if err == nil {
		t.Errorf("Expected error after deletion")
	}
}

func TestShareManagerGetSessionShares(t *testing.T) {
	sm := NewShareManager()

	sm.CreateShareLink("session-1", 1, "Share1", "", PermissionView, nil)
	sm.CreateShareLink("session-1", 1, "Share2", "", PermissionComment, nil)

	shares := sm.GetSessionShares("session-1")

	if len(shares) != 2 {
		t.Errorf("Expected 2 shares for session")
	}
}

func TestShareManagerGetAccesses(t *testing.T) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	sm.AccessShareLink(link.ID, "192.168.1.1", "Mozilla", nil)
	sm.AccessShareLink(link.ID, "192.168.1.2", "Chrome", nil)

	accesses := sm.GetShareAccesses(link.ID)

	if len(accesses) != 2 {
		t.Errorf("Expected 2 accesses")
	}
}

func TestShareManagerStatistics(t *testing.T) {
	sm := NewShareManager()

	sm.CreateShareLink("session-1", 1, "Share1", "", PermissionView, nil)
	sm.CreateShareLink("session-1", 1, "Share2", "", PermissionComment, nil)

	stats := sm.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}

	if totalShares, ok := stats["total_shares"].(int64); !ok || totalShares != 2 {
		t.Errorf("Expected 2 total shares")
	}
}

func TestPermissionValidatorCanPerform(t *testing.T) {
	pv := NewPermissionValidator()

	tests := []struct {
		permission SharePermission
		action     string
		expected   bool
	}{
		{PermissionView, "view", true},
		{PermissionView, "comment", false},
		{PermissionComment, "view", true},
		{PermissionComment, "comment", true},
		{PermissionComment, "edit", false},
		{PermissionEdit, "edit", true},
		{PermissionAdmin, "admin", true},
	}

	for _, test := range tests {
		result := pv.CanPerform(test.permission, test.action)
		if result != test.expected {
			t.Errorf("CanPerform(%s, %s) = %v, expected %v", test.permission, test.action, result, test.expected)
		}
	}
}

func TestPermissionValidatorGetRequired(t *testing.T) {
	pv := NewPermissionValidator()

	tests := []struct {
		action   string
		expected SharePermission
	}{
		{"view", PermissionView},
		{"comment", PermissionComment},
		{"edit", PermissionEdit},
		{"admin", PermissionAdmin},
	}

	for _, test := range tests {
		result := pv.GetRequiredPermission(test.action)
		if result != test.expected {
			t.Errorf("GetRequiredPermission(%s) = %s, expected %s", test.action, result, test.expected)
		}
	}
}

func TestShareLinkExpiration(t *testing.T) {
	sm := NewShareManager()

	expiresIn := 1 * time.Second
	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, &expiresIn)

	if link.IsExpired() {
		t.Errorf("Expected link not to be expired immediately")
	}

	time.Sleep(2 * time.Second)

	if !link.IsExpired() {
		t.Errorf("Expected link to be expired after duration")
	}
}

func TestShareManagerPurgeExpired(t *testing.T) {
	sm := NewShareManager()

	expiresIn := 100 * time.Millisecond
	link1, _ := sm.CreateShareLink("session-1", 1, "Share1", "", PermissionView, &expiresIn)

	link2, _ := sm.CreateShareLink("session-1", 1, "Share2", "", PermissionView, nil)

	time.Sleep(200 * time.Millisecond)

	purged := sm.PurgeExpiredShares()

	if purged != 1 {
		t.Errorf("Expected 1 purged link")
	}

	_, err := sm.GetShareLink(link1.ID)
	if err == nil {
		t.Errorf("Expected expired link to be purged")
	}

	_, err = sm.GetShareLink(link2.ID)
	if err != nil {
		t.Errorf("Expected permanent link to remain")
	}
}

func BenchmarkCreateShareLink(b *testing.B) {
	sm := NewShareManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)
	}
}

func BenchmarkAccessShareLink(b *testing.B) {
	sm := NewShareManager()

	link, _ := sm.CreateShareLink("session-1", 1, "Share", "", PermissionView, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.AccessShareLink(link.ID, "192.168.1.1", "Mozilla", nil)
	}
}

func BenchmarkValidatePermission(b *testing.B) {
	pv := NewPermissionValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pv.CanPerform(PermissionEdit, "edit")
	}
}

