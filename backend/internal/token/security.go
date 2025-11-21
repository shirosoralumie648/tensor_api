package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	EncryptionKey       string        // AES-256 密钥 (32 字节)
	AnomalyDetection    bool          // 启用异常检测
	LeakageMonitoring   bool          // 启用泄露监控
	FailureThreshold    int           // 失败次数阈值
	TimeoutDuration     time.Duration // 超时时间
	MaxConsecutiveErrors int          // 最大连续错误
}

// TokenSecurityManager Token 安全管理器
type TokenSecurityManager struct {
	mu                sync.RWMutex
	config            *SecurityConfig
	encryptionCipher  cipher.Block
	anomalyDetector   *AnomalyDetector
	leakageMonitor    *LeakageMonitor
	failureTracker    map[string]*FailureRecord
}

// FailureRecord 失败记录
type FailureRecord struct {
	TokenID           string
	ConsecutiveErrors int
	LastErrorTime     time.Time
	ErrorCount        int64
	LockedUntil       time.Time
}

// AnomalyDetector 异常检测器
type AnomalyDetector struct {
	mu             sync.RWMutex
	usagePatterns  map[string]*UsagePattern
	threshold      float64
}

// UsagePattern 使用模式
type UsagePattern struct {
	TokenID         string
	AvgRequestTime  float64
	AvgRequestSize  float64
	CommonModels    []string
	CommonIPs       []string
	LastUpdate      time.Time
	Anomalies       int64
}

// LeakageMonitor 泄露监控器
type LeakageMonitor struct {
	mu             sync.RWMutex
	publicDatabases []string // 公开数据库列表
	suspiciousTokens map[string]*SuspiciousRecord
}

// SuspiciousRecord 可疑记录
type SuspiciousRecord struct {
	TokenID      string
	DetectedAt   time.Time
	Source       string
	Severity     string // low, medium, high
	Description  string
}

// NewTokenSecurityManager 创建 Token 安全管理器
func NewTokenSecurityManager(config *SecurityConfig) (*TokenSecurityManager, error) {
	// 初始化加密
	block, err := aes.NewCipher([]byte(config.EncryptionKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	anomalyDetector := &AnomalyDetector{
		usagePatterns: make(map[string]*UsagePattern),
		threshold:     0.7,
	}

	leakageMonitor := &LeakageMonitor{
		publicDatabases: []string{},
		suspiciousTokens: make(map[string]*SuspiciousRecord),
	}

	return &TokenSecurityManager{
		config:          config,
		encryptionCipher: block,
		anomalyDetector: anomalyDetector,
		leakageMonitor:  leakageMonitor,
		failureTracker:  make(map[string]*FailureRecord),
	}, nil
}

// EncryptToken 加密 Token
func (tsm *TokenSecurityManager) EncryptToken(token string) (string, error) {
	plaintext := []byte(token)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	stream := cipher.NewCFBEncrypter(tsm.encryptionCipher, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken 解密 Token
func (tsm *TokenSecurityManager) DecryptToken(encryptedToken string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %v", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(tsm.encryptionCipher, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// DetectAnomaly 检测异常
func (tsm *TokenSecurityManager) DetectAnomaly(tokenID string, model string, requestSize float64) (bool, error) {
	if !tsm.config.AnomalyDetection {
		return false, nil
	}

	tsm.anomalyDetector.mu.RLock()
	pattern, exists := tsm.anomalyDetector.usagePatterns[tokenID]
	tsm.anomalyDetector.mu.RUnlock()

	if !exists {
		// 初始化模式
		pattern = &UsagePattern{
			TokenID:        tokenID,
			CommonModels:   []string{model},
			AvgRequestSize: requestSize,
			LastUpdate:     time.Now(),
		}
		tsm.anomalyDetector.mu.Lock()
		tsm.anomalyDetector.usagePatterns[tokenID] = pattern
		tsm.anomalyDetector.mu.Unlock()
		return false, nil
	}

	// 检查异常
	var anomalyScore float64

	// 检查模型异常
	found := false
	for _, m := range pattern.CommonModels {
		if m == model {
			found = true
			break
		}
	}
	if !found && len(pattern.CommonModels) > 0 {
		anomalyScore += 0.3
	}

	// 检查请求大小异常
	if pattern.AvgRequestSize > 0 {
		ratio := requestSize / pattern.AvgRequestSize
		if ratio > 2.0 || ratio < 0.5 {
			anomalyScore += 0.3
		}
	}

	isAnomaly := anomalyScore >= tsm.anomalyDetector.threshold

	if isAnomaly {
		tsm.anomalyDetector.mu.Lock()
		pattern.Anomalies++
		tsm.anomalyDetector.mu.Unlock()
	}

	return isAnomaly, nil
}

// MonitorLeakage 监控泄露
func (tsm *TokenSecurityManager) MonitorLeakage(tokenID string) (bool, error) {
	if !tsm.config.LeakageMonitoring {
		return false, nil
	}

	// 检查泄露数据库
	for _, db := range tsm.leakageMonitor.publicDatabases {
		// 这里应该实现实际的泄露检测逻辑
		// 简化版本仅记录检查时间
	}

	return false, nil
}

// RecordFailure 记录失败
func (tsm *TokenSecurityManager) RecordFailure(tokenID string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	record, exists := tsm.failureTracker[tokenID]
	if !exists {
		record = &FailureRecord{
			TokenID: tokenID,
		}
		tsm.failureTracker[tokenID] = record
	}

	record.ConsecutiveErrors++
	record.ErrorCount++
	record.LastErrorTime = time.Now()

	// 检查是否超过阈值
	if record.ConsecutiveErrors >= tsm.config.MaxConsecutiveErrors {
		record.LockedUntil = time.Now().Add(tsm.config.TimeoutDuration)
	}

	return nil
}

// ResetFailure 重置失败
func (tsm *TokenSecurityManager) ResetFailure(tokenID string) {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if record, exists := tsm.failureTracker[tokenID]; exists {
		record.ConsecutiveErrors = 0
		record.LockedUntil = time.Time{}
	}
}

// IsTokenLocked 检查 Token 是否被锁定
func (tsm *TokenSecurityManager) IsTokenLocked(tokenID string) bool {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	record, exists := tsm.failureTracker[tokenID]
	if !exists {
		return false
	}

	if record.LockedUntil.IsZero() {
		return false
	}

	if time.Now().After(record.LockedUntil) {
		// 解锁
		record.LockedUntil = time.Time{}
		record.ConsecutiveErrors = 0
		return false
	}

	return true
}

// ReportSuspiciousActivity 报告可疑活动
func (tsm *TokenSecurityManager) ReportSuspiciousActivity(tokenID, source, description, severity string) error {
	tsm.leakageMonitor.mu.Lock()
	defer tsm.leakageMonitor.mu.Unlock()

	record := &SuspiciousRecord{
		TokenID:     tokenID,
		DetectedAt:  time.Now(),
		Source:      source,
		Severity:    severity,
		Description: description,
	}

	tsm.leakageMonitor.suspiciousTokens[tokenID] = record

	return nil
}

// GetSecurityStatus 获取安全状态
func (tsm *TokenSecurityManager) GetSecurityStatus(tokenID string) map[string]interface{} {
	tsm.mu.RLock()
	failureRecord, hasFailure := tsm.failureTracker[tokenID]
	tsm.mu.RUnlock()

	status := map[string]interface{}{
		"token_id":    tokenID,
		"is_locked":   tsm.IsTokenLocked(tokenID),
		"last_error":  nil,
		"anomalies":   0,
	}

	if hasFailure {
		status["last_error"] = failureRecord.LastErrorTime
		status["consecutive_errors"] = failureRecord.ConsecutiveErrors
		status["error_count"] = failureRecord.ErrorCount
	}

	tsm.anomalyDetector.mu.RLock()
	if pattern, exists := tsm.anomalyDetector.usagePatterns[tokenID]; exists {
		status["anomalies"] = pattern.Anomalies
	}
	tsm.anomalyDetector.mu.RUnlock()

	return status
}


