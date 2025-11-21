package token

import (
	"testing"
)

func TestEncryptDecryptToken(t *testing.T) {
	config := &SecurityConfig{
		EncryptionKey:        "0123456789abcdef0123456789abcdef", // 32 字节
		AnomalyDetection:     true,
		LeakageMonitoring:    true,
		FailureThreshold:     5,
		MaxConsecutiveErrors: 3,
	}

	manager, err := NewTokenSecurityManager(config)
	if err != nil {
		t.Fatalf("NewTokenSecurityManager failed: %v", err)
	}

	plainToken := "test_token_123456789"

	encrypted, err := manager.EncryptToken(plainToken)
	if err != nil {
		t.Fatalf("EncryptToken failed: %v", err)
	}

	decrypted, err := manager.DecryptToken(encrypted)
	if err != nil {
		t.Fatalf("DecryptToken failed: %v", err)
	}

	if decrypted != plainToken {
		t.Errorf("Expected %s, got %s", plainToken, decrypted)
	}
}

func TestRecordFailure(t *testing.T) {
	config := &SecurityConfig{
		EncryptionKey:        "0123456789abcdef0123456789abcdef",
		AnomalyDetection:     true,
		LeakageMonitoring:    true,
		FailureThreshold:     5,
		MaxConsecutiveErrors: 3,
	}

	manager, err := NewTokenSecurityManager(config)
	if err != nil {
		t.Fatalf("NewTokenSecurityManager failed: %v", err)
	}

	tokenID := "token123"

	err = manager.RecordFailure(tokenID)
	if err != nil {
		t.Fatalf("RecordFailure failed: %v", err)
	}

	isLocked := manager.IsTokenLocked(tokenID)
	if isLocked {
		t.Error("Expected token not to be locked after 1 failure")
	}

	// 记录更多失败
	for i := 0; i < 2; i++ {
		manager.RecordFailure(tokenID)
	}

	isLocked = manager.IsTokenLocked(tokenID)
	if !isLocked {
		t.Error("Expected token to be locked after 3 failures")
	}
}

func TestDetectAnomaly(t *testing.T) {
	config := &SecurityConfig{
		EncryptionKey:        "0123456789abcdef0123456789abcdef",
		AnomalyDetection:     true,
		LeakageMonitoring:    false,
		FailureThreshold:     5,
		MaxConsecutiveErrors: 3,
	}

	manager, err := NewTokenSecurityManager(config)
	if err != nil {
		t.Fatalf("NewTokenSecurityManager failed: %v", err)
	}

	tokenID := "token123"

	// 首次请求不应该标记为异常
	isAnomaly, err := manager.DetectAnomaly(tokenID, "gpt-4", 1000.0)
	if err != nil {
		t.Fatalf("DetectAnomaly failed: %v", err)
	}

	if isAnomaly {
		t.Error("Expected first request not to be anomaly")
	}

	// 大幅不同的请求应该被标记为异常
	isAnomaly, err = manager.DetectAnomaly(tokenID, "claude-3", 5000.0)
	if err != nil {
		t.Fatalf("DetectAnomaly failed: %v", err)
	}

	if !isAnomaly {
		t.Error("Expected large request to be anomaly")
	}
}

func TestReportSuspiciousActivity(t *testing.T) {
	config := &SecurityConfig{
		EncryptionKey:        "0123456789abcdef0123456789abcdef",
		AnomalyDetection:     true,
		LeakageMonitoring:    true,
		FailureThreshold:     5,
		MaxConsecutiveErrors: 3,
	}

	manager, err := NewTokenSecurityManager(config)
	if err != nil {
		t.Fatalf("NewTokenSecurityManager failed: %v", err)
	}

	err = manager.ReportSuspiciousActivity("token123", "external_db", "Token found in public breach", "high")
	if err != nil {
		t.Fatalf("ReportSuspiciousActivity failed: %v", err)
	}
}

func TestGetSecurityStatus(t *testing.T) {
	config := &SecurityConfig{
		EncryptionKey:        "0123456789abcdef0123456789abcdef",
		AnomalyDetection:     true,
		LeakageMonitoring:    true,
		FailureThreshold:     5,
		MaxConsecutiveErrors: 3,
	}

	manager, err := NewTokenSecurityManager(config)
	if err != nil {
		t.Fatalf("NewTokenSecurityManager failed: %v", err)
	}

	tokenID := "token123"

	status := manager.GetSecurityStatus(tokenID)
	if status["token_id"] != tokenID {
		t.Errorf("Expected token_id %s, got %v", tokenID, status["token_id"])
	}

	if status["is_locked"] != false {
		t.Error("Expected is_locked to be false")
	}
}

func BenchmarkEncryptToken(b *testing.B) {
	config := &SecurityConfig{
		EncryptionKey: "0123456789abcdef0123456789abcdef",
	}

	manager, _ := NewTokenSecurityManager(config)

	plainToken := "test_token_123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.EncryptToken(plainToken)
	}
}

func BenchmarkDecryptToken(b *testing.B) {
	config := &SecurityConfig{
		EncryptionKey: "0123456789abcdef0123456789abcdef",
	}

	manager, _ := NewTokenSecurityManager(config)

	plainToken := "test_token_123456789"
	encrypted, _ := manager.EncryptToken(plainToken)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.DecryptToken(encrypted)
	}
}


