package model

import (
	"database/sql"
	"time"
)

// TokenStatus Token 状态枚举
type TokenStatus int

const (
	TokenStatusNormal     TokenStatus = 1 // 正常
	TokenStatusExhausted  TokenStatus = 2 // 已耗尽
	TokenStatusDisabled   TokenStatus = 3 // 已禁用
	TokenStatusExpired    TokenStatus = 4 // 已过期
	TokenStatusDeleted    TokenStatus = 5 // 已删除（软删除）
)

// TokenStatusName Token 状态名称映射
var TokenStatusName = map[TokenStatus]string{
	TokenStatusNormal:    "正常",
	TokenStatusExhausted: "已耗尽",
	TokenStatusDisabled:  "已禁用",
	TokenStatusExpired:   "已过期",
	TokenStatusDeleted:   "已删除",
}

// Token API Token 模型
type Token struct {
	ID               int
	UserID           int
	TokenHash        string
	Name             string
	Description      sql.NullString
	Status           TokenStatus
	QuotaLimit       sql.NullInt64
	QuotaUsed        int64
	CreatedAt        time.Time
	ExpireAt         sql.NullTime
	RenewedAt        sql.NullTime
	DeletedAt        sql.NullTime
	LastUsedAt       sql.NullTime
	IPWhitelist      []string
	ModelWhitelist   []string
	Metadata         map[string]interface{}
	UpdatedAt        time.Time
}

// TokenAuditLog Token 审计日志
type TokenAuditLog struct {
	ID        int64
	UserID    int
	TokenID   int
	Operation string
	OldStatus *TokenStatus
	NewStatus *TokenStatus
	Details   map[string]interface{}
	CreatedAt time.Time
	IPAddress sql.NullString
	UserAgent sql.NullString
}

// TokenRenewalLog Token 续期日志
type TokenRenewalLog struct {
	ID            int64
	TokenID       int
	OldExpireAt   sql.NullTime
	NewExpireAt   sql.NullTime
	RenewalReason string
	CreatedAt     time.Time
}

// TokenQuotaThreshold Token 配额预警阈值
type TokenQuotaThreshold struct {
	ID               int
	UserID           int
	TokenID          int
	ThresholdPercent int
	IsWarned         bool
	WarnedAt         sql.NullTime
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsValid 检查 Token 是否有效
func (t *Token) IsValid() bool {
	// 状态必须是正常
	if t.Status != TokenStatusNormal {
		return false
	}

	// 检查是否已删除
	if t.DeletedAt.Valid {
		return false
	}

	// 检查是否已过期
	if t.ExpireAt.Valid && t.ExpireAt.Time.Before(time.Now()) {
		return false
	}

	// 检查配额
	if t.QuotaLimit.Valid && t.QuotaUsed >= t.QuotaLimit.Int64 {
		return false
	}

	return true
}

// GetRemainingQuota 获取剩余配额
func (t *Token) GetRemainingQuota() int64 {
	if !t.QuotaLimit.Valid {
		return -1 // 无限额
	}
	remaining := t.QuotaLimit.Int64 - t.QuotaUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetQuotaPercentage 获取配额使用百分比（0-100）
func (t *Token) GetQuotaPercentage() float64 {
	if !t.QuotaLimit.Valid || t.QuotaLimit.Int64 == 0 {
		return 0
	}
	return float64(t.QuotaUsed) / float64(t.QuotaLimit.Int64) * 100
}

// IsExpiringSoon 检查 Token 是否即将过期（在 7 天内）
func (t *Token) IsExpiringSoon() bool {
	if !t.ExpireAt.Valid {
		return false
	}
	sevenDaysLater := time.Now().AddDate(0, 0, 7)
	return t.ExpireAt.Time.Before(sevenDaysLater) && t.ExpireAt.Time.After(time.Now())
}

// CanRenew 检查 Token 是否可以续期
func (t *Token) CanRenew() bool {
	// 必须处于正常状态
	if t.Status != TokenStatusNormal {
		return false
	}

	// 必须在有效期内或即将过期（7 天内）
	if !t.ExpireAt.Valid {
		return false
	}

	if t.ExpireAt.Time.Before(time.Now()) {
		return false
	}

	return t.IsExpiringSoon()
}

// GetStatusName 获取状态名称
func (s TokenStatus) String() string {
	if name, ok := TokenStatusName[s]; ok {
		return name
	}
	return "未知"
}

// ValidateIPAddress 验证 IP 地址是否在白名单中
func (t *Token) ValidateIPAddress(ip string) bool {
	if len(t.IPWhitelist) == 0 {
		return true // 没有 IP 限制
	}

	for _, whitelistIP := range t.IPWhitelist {
		if whitelistIP == ip || whitelistIP == "*" {
			return true
		}
	}

	return false
}

// ValidateModel 验证模型是否在白名单中
func (t *Token) ValidateModel(model string) bool {
	if len(t.ModelWhitelist) == 0 {
		return true // 没有模型限制
	}

	for _, whitelistModel := range t.ModelWhitelist {
		if whitelistModel == model || whitelistModel == "*" {
			return true
		}
	}

	return false
}

// UseQuota 使用配额
func (t *Token) UseQuota(amount int64) error {
	if !t.IsValid() {
		return ErrTokenInvalid
	}

	if t.QuotaLimit.Valid && t.QuotaUsed+amount > t.QuotaLimit.Int64 {
		return ErrQuotaExceeded
	}

	t.QuotaUsed += amount
	return nil
}

// RefundQuota 退款配额
func (t *Token) RefundQuota(amount int64) error {
	if t.QuotaUsed < amount {
		return ErrInvalidRefundAmount
	}

	t.QuotaUsed -= amount
	return nil
}

// TokenOperationType Token 操作类型
type TokenOperationType string

const (
	TokenOpCreate   TokenOperationType = "create"
	TokenOpUpdate   TokenOperationType = "update"
	TokenOpDelete   TokenOperationType = "delete"
	TokenOpRenew    TokenOperationType = "renew"
	TokenOpDisable  TokenOperationType = "disable"
	TokenOpEnable   TokenOperationType = "enable"
	TokenOpExpire   TokenOperationType = "expire"
	TokenOpUseQuota TokenOperationType = "use_quota"
)

