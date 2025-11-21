package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// TokenStatus Token 状态
type TokenStatus string

const (
	StatusActive    TokenStatus = "active"
	StatusRotating  TokenStatus = "rotating"
	StatusRotated   TokenStatus = "rotated"
	StatusExpired   TokenStatus = "expired"
	StatusRevoked   TokenStatus = "revoked"
)

// TokenType Token 类型
type TokenType string

const (
	TypeAPI      TokenType = "api"
	TypeWeb      TokenType = "web"
	TypeInternal TokenType = "internal"
	TypeTemporal TokenType = "temporal"
)

// TokenMetadata Token 元数据
type TokenMetadata struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Type         TokenType         `json:"type"`
	Status       TokenStatus       `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    time.Time         `json:"expires_at"`
	LastUsedAt   time.Time         `json:"last_used_at"`
	LastRotatedAt time.Time        `json:"last_rotated_at"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Tags         map[string]string `json:"tags"`
	UsageCount   int64             `json:"usage_count"`
	RotationCount int64            `json:"rotation_count"`
}

// TokenLifecycleManager Token 生命周期管理器
type TokenLifecycleManager struct {
	mu       sync.RWMutex
	tokens   map[string]*TokenMetadata
	store    TokenStore
	rotationPolicy *RotationPolicy
}

// RotationPolicy 轮换策略
type RotationPolicy struct {
	// 轮换周期
	RotationInterval time.Duration
	// 警告期（过期前多久警告）
	WarningPeriod time.Duration
	// 最大轮换次数
	MaxRotationCount int64
	// 启用自动轮换
	AutoRotationEnabled bool
}

// TokenStore Token 存储接口
type TokenStore interface {
	Create(metadata *TokenMetadata) error
	Update(metadata *TokenMetadata) error
	Get(id string) (*TokenMetadata, error)
	Delete(id string) error
	List(userID string) ([]*TokenMetadata, error)
	Expire(id string) error
	Rotate(oldID string, newMetadata *TokenMetadata) error
}

// DefaultRotationPolicy 默认轮换策略
var DefaultRotationPolicy = &RotationPolicy{
	RotationInterval:    30 * 24 * time.Hour, // 30 天
	WarningPeriod:       7 * 24 * time.Hour,  // 7 天
	MaxRotationCount:    100,
	AutoRotationEnabled: true,
}

// NewTokenLifecycleManager 创建 Token 生命周期管理器
func NewTokenLifecycleManager(store TokenStore, policy *RotationPolicy) *TokenLifecycleManager {
	if policy == nil {
		policy = DefaultRotationPolicy
	}

	return &TokenLifecycleManager{
		tokens:         make(map[string]*TokenMetadata),
		store:          store,
		rotationPolicy: policy,
	}
}

// GenerateToken 生成 Token
func (tm *TokenLifecycleManager) GenerateToken(userID string, tokenType TokenType, name string, validityDays int) (string, *TokenMetadata, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %v", err)
	}

	tokenID := hex.EncodeToString(tokenBytes)

	now := time.Now()
	metadata := &TokenMetadata{
		ID:        tokenID,
		UserID:    userID,
		Type:      tokenType,
		Status:    StatusActive,
		CreatedAt: now,
		ExpiresAt: now.AddDate(0, 0, validityDays),
		Name:      name,
		Tags:      make(map[string]string),
	}

	if err := tm.store.Create(metadata); err != nil {
		return "", nil, fmt.Errorf("failed to store token: %v", err)
	}

	tm.mu.Lock()
	tm.tokens[tokenID] = metadata
	tm.mu.Unlock()

	return tokenID, metadata, nil
}

// GetToken 获取 Token 元数据
func (tm *TokenLifecycleManager) GetToken(tokenID string) (*TokenMetadata, error) {
	tm.mu.RLock()
	metadata, exists := tm.tokens[tokenID]
	tm.mu.RUnlock()

	if exists {
		// 检查是否已过期
		if time.Now().After(metadata.ExpiresAt) && metadata.Status != StatusExpired {
			tm.Expire(tokenID)
			return nil, fmt.Errorf("token expired")
		}
		return metadata, nil
	}

	// 从存储中加载
	metadata, err := tm.store.Get(tokenID)
	if err != nil {
		return nil, err
	}

	tm.mu.Lock()
	tm.tokens[tokenID] = metadata
	tm.mu.Unlock()

	return metadata, nil
}

// RecordUsage 记录 Token 使用
func (tm *TokenLifecycleManager) RecordUsage(tokenID string) error {
	metadata, err := tm.GetToken(tokenID)
	if err != nil {
		return err
	}

	if metadata.Status != StatusActive {
		return fmt.Errorf("token not active, status: %s", metadata.Status)
	}

	metadata.UsageCount++
	metadata.LastUsedAt = time.Now()

	tm.mu.Lock()
	tm.tokens[tokenID] = metadata
	tm.mu.Unlock()

	return tm.store.Update(metadata)
}

// RotateToken 轮换 Token
func (tm *TokenLifecycleManager) RotateToken(tokenID string, newValidityDays int) (string, *TokenMetadata, error) {
	oldMetadata, err := tm.GetToken(tokenID)
	if err != nil {
		return "", nil, err
	}

	if oldMetadata.RotationCount >= tm.rotationPolicy.MaxRotationCount {
		return "", nil, fmt.Errorf("token rotation count exceeded")
	}

	// 生成新 Token
	newTokenBytes := make([]byte, 32)
	if _, err := rand.Read(newTokenBytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate new token: %v", err)
	}

	newTokenID := hex.EncodeToString(newTokenBytes)

	now := time.Now()
	newMetadata := &TokenMetadata{
		ID:           newTokenID,
		UserID:       oldMetadata.UserID,
		Type:         oldMetadata.Type,
		Status:       StatusActive,
		CreatedAt:    now,
		ExpiresAt:    now.AddDate(0, 0, newValidityDays),
		Name:         oldMetadata.Name,
		Description:  oldMetadata.Description,
		Tags:         oldMetadata.Tags,
		RotationCount: oldMetadata.RotationCount + 1,
	}

	// 标记旧 Token 为已轮换
	oldMetadata.Status = StatusRotated
	oldMetadata.LastRotatedAt = now

	if err := tm.store.Rotate(tokenID, newMetadata); err != nil {
		return "", nil, err
	}

	tm.mu.Lock()
	tm.tokens[tokenID] = oldMetadata
	tm.tokens[newTokenID] = newMetadata
	tm.mu.Unlock()

	return newTokenID, newMetadata, nil
}

// Expire 标记 Token 为过期
func (tm *TokenLifecycleManager) Expire(tokenID string) error {
	metadata, err := tm.GetToken(tokenID)
	if err != nil {
		return err
	}

	metadata.Status = StatusExpired

	tm.mu.Lock()
	tm.tokens[tokenID] = metadata
	tm.mu.Unlock()

	return tm.store.Update(metadata)
}

// Revoke 撤销 Token
func (tm *TokenLifecycleManager) Revoke(tokenID string) error {
	metadata, err := tm.GetToken(tokenID)
	if err != nil {
		return err
	}

	metadata.Status = StatusRevoked

	tm.mu.Lock()
	tm.tokens[tokenID] = metadata
	tm.mu.Unlock()

	return tm.store.Update(metadata)
}

// ListTokens 列出用户的所有 Token
func (tm *TokenLifecycleManager) ListTokens(userID string) ([]*TokenMetadata, error) {
	return tm.store.List(userID)
}

// GetExpiringTokens 获取即将过期的 Token
func (tm *TokenLifecycleManager) GetExpiringTokens(userID string) ([]*TokenMetadata, error) {
	tokens, err := tm.store.List(userID)
	if err != nil {
		return nil, err
	}

	warningTime := time.Now().Add(tm.rotationPolicy.WarningPeriod)
	var expiringTokens []*TokenMetadata

	for _, token := range tokens {
		if token.Status == StatusActive && token.ExpiresAt.Before(warningTime) {
			expiringTokens = append(expiringTokens, token)
		}
	}

	return expiringTokens, nil
}

// CheckAutoRotation 检查自动轮换
func (tm *TokenLifecycleManager) CheckAutoRotation(userID string) error {
	if !tm.rotationPolicy.AutoRotationEnabled {
		return nil
	}

	tokens, err := tm.store.List(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, token := range tokens {
		if token.Status != StatusActive {
			continue
		}

		// 检查是否需要轮换
		if token.LastRotatedAt.Add(tm.rotationPolicy.RotationInterval).Before(now) {
			_, _, err := tm.RotateToken(token.ID, 30)
			if err != nil {
				return err
			}
		}
	}

	return nil
}


