package ratelimit

import (
	"sync"
	"time"
)

// QuotaType 配额类型
type QuotaType string

const (
	QuotaDaily   QuotaType = "daily"
	QuotaMonthly QuotaType = "monthly"
	QuotaHourly  QuotaType = "hourly"
)

// QuotaManager 配额管理器
type QuotaManager struct {
	mu           sync.RWMutex
	quotas       map[string]*UserQuota
	dailyQuota   int64
	monthlyQuota int64
	hourlyQuota  int64
}

// UserQuota 用户配额
type UserQuota struct {
	UserID       string
	DailyUsed    int64
	MonthlyUsed  int64
	HourlyUsed   int64
	DailyReset   time.Time
	MonthlyReset time.Time
	HourlyReset  time.Time
	Locked       bool
	LockedUntil  time.Time
}

// QuotaRequest 配额请求
type QuotaRequest struct {
	UserID   string
	Cost     int64
	Type     QuotaType
}

// QuotaResponse 配额响应
type QuotaResponse struct {
	Allowed      bool
	Remaining    int64
	RetryAfter   time.Duration
	Message      string
}

// NewQuotaManager 创建配额管理器
func NewQuotaManager(daily, monthly, hourly int64) *QuotaManager {
	return &QuotaManager{
		quotas:       make(map[string]*UserQuota),
		dailyQuota:   daily,
		monthlyQuota: monthly,
		hourlyQuota:  hourly,
	}
}

// CheckQuota 检查配额
func (qm *QuotaManager) CheckQuota(req *QuotaRequest) *QuotaResponse {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[req.UserID]
	if !exists {
		quota = &UserQuota{
			UserID:       req.UserID,
			DailyReset:   time.Now().Add(24 * time.Hour),
			MonthlyReset: time.Now().AddDate(0, 1, 0),
			HourlyReset:  time.Now().Add(time.Hour),
		}
		qm.quotas[req.UserID] = quota
	}

	// 检查锁定状态
	if quota.Locked && time.Now().Before(quota.LockedUntil) {
		return &QuotaResponse{
			Allowed:    false,
			RetryAfter: time.Until(quota.LockedUntil),
			Message:    "Account locked due to quota violation",
		}
	}

	if quota.Locked {
		quota.Locked = false
	}

	// 重置过期的配额
	qm.resetExpiredQuotas(quota)

	// 检查各个类型的配额
	response := &QuotaResponse{Allowed: true}

	// 检查每日配额
	if quota.DailyUsed+req.Cost > qm.dailyQuota {
		response.Allowed = false
		response.Remaining = qm.dailyQuota - quota.DailyUsed
		response.RetryAfter = time.Until(quota.DailyReset)
		response.Message = "Daily quota exceeded"
		qm.lockAccount(quota)
		return response
	}

	// 检查每月配额
	if quota.MonthlyUsed+req.Cost > qm.monthlyQuota {
		response.Allowed = false
		response.Remaining = qm.monthlyQuota - quota.MonthlyUsed
		response.RetryAfter = time.Until(quota.MonthlyReset)
		response.Message = "Monthly quota exceeded"
		qm.lockAccount(quota)
		return response
	}

	// 检查每小时配额
	if quota.HourlyUsed+req.Cost > qm.hourlyQuota {
		response.Allowed = false
		response.Remaining = qm.hourlyQuota - quota.HourlyUsed
		response.RetryAfter = time.Until(quota.HourlyReset)
		response.Message = "Hourly quota exceeded"
		return response
	}

	// 更新使用量
	quota.DailyUsed += req.Cost
	quota.MonthlyUsed += req.Cost
	quota.HourlyUsed += req.Cost

	response.Remaining = qm.dailyQuota - quota.DailyUsed

	return response
}

// resetExpiredQuotas 重置过期的配额
func (qm *QuotaManager) resetExpiredQuotas(quota *UserQuota) {
	now := time.Now()

	if now.After(quota.HourlyReset) {
		quota.HourlyUsed = 0
		quota.HourlyReset = now.Add(time.Hour)
	}

	if now.After(quota.DailyReset) {
		quota.DailyUsed = 0
		quota.DailyReset = now.Add(24 * time.Hour)
	}

	if now.After(quota.MonthlyReset) {
		quota.MonthlyUsed = 0
		quota.MonthlyReset = now.AddDate(0, 1, 0)
	}
}

// lockAccount 锁定账户
func (qm *QuotaManager) lockAccount(quota *UserQuota) {
	quota.Locked = true
	quota.LockedUntil = time.Now().Add(time.Hour) // 锁定 1 小时
}

// GetQuotaStatus 获取配额状态
func (qm *QuotaManager) GetQuotaStatus(userID string) map[string]interface{} {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		return map[string]interface{}{
			"daily":   map[string]int64{"used": 0, "quota": qm.dailyQuota},
			"monthly": map[string]int64{"used": 0, "quota": qm.monthlyQuota},
			"hourly":  map[string]int64{"used": 0, "quota": qm.hourlyQuota},
			"locked":  false,
		}
	}

	return map[string]interface{}{
		"daily": map[string]interface{}{
			"used":      quota.DailyUsed,
			"quota":     qm.dailyQuota,
			"remaining": qm.dailyQuota - quota.DailyUsed,
			"reset":     quota.DailyReset,
		},
		"monthly": map[string]interface{}{
			"used":      quota.MonthlyUsed,
			"quota":     qm.monthlyQuota,
			"remaining": qm.monthlyQuota - quota.MonthlyUsed,
			"reset":     quota.MonthlyReset,
		},
		"hourly": map[string]interface{}{
			"used":      quota.HourlyUsed,
			"quota":     qm.hourlyQuota,
			"remaining": qm.hourlyQuota - quota.HourlyUsed,
			"reset":     quota.HourlyReset,
		},
		"locked":       quota.Locked,
		"locked_until": quota.LockedUntil,
	}
}

// ResetUserQuota 重置用户配额
func (qm *QuotaManager) ResetUserQuota(userID string, quotaType QuotaType) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		return
	}

	switch quotaType {
	case QuotaDaily:
		quota.DailyUsed = 0
		quota.DailyReset = time.Now().Add(24 * time.Hour)
	case QuotaMonthly:
		quota.MonthlyUsed = 0
		quota.MonthlyReset = time.Now().AddDate(0, 1, 0)
	case QuotaHourly:
		quota.HourlyUsed = 0
		quota.HourlyReset = time.Now().Add(time.Hour)
	}
}

// SetUserQuota 设置用户配额
func (qm *QuotaManager) SetUserQuota(userID string, daily, monthly, hourly int64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		quota = &UserQuota{
			UserID:       userID,
			DailyReset:   time.Now().Add(24 * time.Hour),
			MonthlyReset: time.Now().AddDate(0, 1, 0),
			HourlyReset:  time.Now().Add(time.Hour),
		}
		qm.quotas[userID] = quota
	}

	// 更新全局配额（可以根据需要改为单用户配额）
	if daily > 0 {
		qm.dailyQuota = daily
	}
	if monthly > 0 {
		qm.monthlyQuota = monthly
	}
	if hourly > 0 {
		qm.hourlyQuota = hourly
	}
}

// GetRemainingQuota 获取剩余配额
func (qm *QuotaManager) GetRemainingQuota(userID string) (daily, monthly, hourly int64) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	quota, exists := qm.quotas[userID]
	if !exists {
		return qm.dailyQuota, qm.monthlyQuota, qm.hourlyQuota
	}

	return qm.dailyQuota - quota.DailyUsed,
		qm.monthlyQuota - quota.MonthlyUsed,
		qm.hourlyQuota - quota.HourlyUsed
}

