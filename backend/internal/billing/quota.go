package billing

import (
	"sync"
)

// UserQuota 用户配额结构
type UserQuota struct {
	mu             sync.RWMutex
	UsedQuota      float64
	TotalQuota     float64
	AvailableQuota float64
}

// QuotaManager 配额管理器
type QuotaManager struct {
	mu     sync.RWMutex
	quotas map[string]*UserQuota
}

// NewQuotaManager 创建配额管理器
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quotas: make(map[string]*UserQuota),
	}
}

// GetQuota 获取用户配额
func (qm *QuotaManager) GetQuota(userID string) float64 {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	if q, ok := qm.quotas[userID]; ok {
		q.mu.RLock()
		defer q.mu.RUnlock()
		return q.TotalQuota
	}
	return 0
}

// GetUserQuota 获取用户配额对象
func (qm *QuotaManager) GetUserQuota(userID string) (*UserQuota, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, ok := qm.quotas[userID]; !ok {
		qm.quotas[userID] = &UserQuota{}
	}
	return qm.quotas[userID], nil
}

// SetQuota 设置用户配额
func (qm *QuotaManager) SetQuota(userID string, amount float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, ok := qm.quotas[userID]; !ok {
		qm.quotas[userID] = &UserQuota{}
	}

	q := qm.quotas[userID]
	q.mu.Lock()
	q.TotalQuota = amount
	q.AvailableQuota = q.TotalQuota - q.UsedQuota
	q.mu.Unlock()
}

// CreateUserQuota 创建用户配额
func (qm *QuotaManager) CreateUserQuota(userID string, amount float64) error {
	qm.SetQuota(userID, amount)
	return nil
}

// GetUsage 获取用户已使用额度
func (qm *QuotaManager) GetUsage(userID string) float64 {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	if q, ok := qm.quotas[userID]; ok {
		q.mu.RLock()
		defer q.mu.RUnlock()
		return q.UsedQuota
	}
	return 0
}

// AddUsage 增加使用额度
func (qm *QuotaManager) AddUsage(userID string, amount float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, ok := qm.quotas[userID]; !ok {
		qm.quotas[userID] = &UserQuota{}
	}

	q := qm.quotas[userID]
	q.mu.Lock()
	q.UsedQuota += amount
	q.AvailableQuota = q.TotalQuota - q.UsedQuota
	q.mu.Unlock()
}

// PreDeduct 预扣费
func (qm *QuotaManager) PreDeduct(userID, transactionID string, amount float64, reason string) (float64, error) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, ok := qm.quotas[userID]; !ok {
		qm.quotas[userID] = &UserQuota{}
	}

	q := qm.quotas[userID]
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.AvailableQuota < amount {
		return q.AvailableQuota, nil
	}

	// 这里不实际扣减，只是返回剩余量作为检查
	return q.AvailableQuota - amount, nil
}

// Recharge 充值
func (qm *QuotaManager) Recharge(userID string, amount float64) error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, ok := qm.quotas[userID]; !ok {
		qm.quotas[userID] = &UserQuota{}
	}

	q := qm.quotas[userID]
	q.mu.Lock()
	q.TotalQuota += amount
	q.AvailableQuota = q.TotalQuota - q.UsedQuota
	q.mu.Unlock()

	return nil
}

// ConfirmDeduction 确认扣费
func (qm *QuotaManager) ConfirmDeduction(userID string, transactionID string, amount float64) error {
	qm.AddUsage(userID, amount)
	return nil
}

// HasSufficientQuota 检查配额是否充足
func (qm *QuotaManager) HasSufficientQuota(userID string, estimatedCost float64) bool {
	val, _ := qm.GetUserQuota(userID)
	val.mu.RLock()
	defer val.mu.RUnlock()
	return val.AvailableQuota >= estimatedCost
}
