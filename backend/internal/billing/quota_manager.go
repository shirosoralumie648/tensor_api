package billing

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// QuotaOperation 配额操作类型
type QuotaOperation int

const (
	// 预扣费
	OpPreDeduction QuotaOperation = iota
	// 后扣费
	OpPostDeduction
	// 退款
	OpRefund
	// 充值
	OpRecharge
)

// String 返回操作类型的字符串表示
func (op QuotaOperation) String() string {
	switch op {
	case OpPreDeduction:
		return "pre_deduction"
	case OpPostDeduction:
		return "post_deduction"
	case OpRefund:
		return "refund"
	case OpRecharge:
		return "recharge"
	default:
		return "unknown"
	}
}

// UserQuota 用户配额
type UserQuota struct {
	// 用户 ID
	UserID string

	// 总配额（美元）
	TotalQuota float64

	// 已使用配额
	UsedQuota float64

	// 冻结配额（预扣费）
	FrozenQuota float64

	// 可用配额
	AvailableQuota float64

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time

	// 互斥锁
	mu sync.RWMutex
}

// GetAvailable 获取可用配额
func (uq *UserQuota) GetAvailable() float64 {
	uq.mu.RLock()
	defer uq.mu.RUnlock()

	return uq.AvailableQuota
}

// Deduct 扣费
func (uq *UserQuota) Deduct(amount float64) error {
	uq.mu.Lock()
	defer uq.mu.Unlock()

	if amount > uq.AvailableQuota {
		return fmt.Errorf("insufficient quota: need %.4f, have %.4f", amount, uq.AvailableQuota)
	}

	uq.AvailableQuota -= amount
	uq.UsedQuota += amount
	uq.UpdatedAt = time.Now()

	return nil
}

// Freeze 冻结配额（预扣费）
func (uq *UserQuota) Freeze(amount float64) error {
	uq.mu.Lock()
	defer uq.mu.Unlock()

	if amount > uq.AvailableQuota {
		return fmt.Errorf("insufficient quota to freeze: need %.4f, have %.4f", amount, uq.AvailableQuota)
	}

	uq.AvailableQuota -= amount
	uq.FrozenQuota += amount
	uq.UpdatedAt = time.Now()

	return nil
}

// Release 释放冻结的配额（退款）
func (uq *UserQuota) Release(amount float64) error {
	uq.mu.Lock()
	defer uq.mu.Unlock()

	if amount > uq.FrozenQuota {
		return fmt.Errorf("cannot release more than frozen: release %.4f, frozen %.4f", amount, uq.FrozenQuota)
	}

	uq.FrozenQuota -= amount
	uq.AvailableQuota += amount
	uq.UpdatedAt = time.Now()

	return nil
}

// Confirm 确认冻结的配额（扣费确认）
func (uq *UserQuota) Confirm(amount float64) error {
	uq.mu.Lock()
	defer uq.mu.Unlock()

	if amount > uq.FrozenQuota {
		return fmt.Errorf("cannot confirm more than frozen: confirm %.4f, frozen %.4f", amount, uq.FrozenQuota)
	}

	uq.FrozenQuota -= amount
	uq.UsedQuota += amount
	uq.UpdatedAt = time.Now()

	return nil
}

// BillingRecord 账单记录
type BillingRecord struct {
	// 记录 ID
	RecordID string

	// 用户 ID
	UserID string

	// 操作类型
	Operation QuotaOperation

	// 金额
	Amount float64

	// 请求 ID
	RequestID string

	// 模型名称
	ModelName string

	// 输入 token 数
	InputTokens int64

	// 输出 token 数
	OutputTokens int64

	// 描述
	Description string

	// 创建时间
	CreatedAt time.Time

	// 完成时间
	CompletedAt *time.Time

	// 状态（pending, confirmed, refunded）
	Status string
}

// QuotaManager 配额管理器
type QuotaManager struct {
	// 用户配额映射
	userQuotas map[string]*UserQuota
	quotasMu   sync.RWMutex

	// 账单记录
	records map[string]*BillingRecord
	recordsMu sync.RWMutex

	// 待确认的预扣费
	pendingDeductions map[string]float64
	pendingMu         sync.RWMutex

	// 统计信息
	totalDeductions   int64
	totalRefunds      int64
	totalRecharges    int64
	statsMu           sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewQuotaManager 创建配额管理器
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		userQuotas:        make(map[string]*UserQuota),
		records:           make(map[string]*BillingRecord),
		pendingDeductions: make(map[string]float64),
		logFunc:           defaultLogFunc,
	}
}

// CreateUserQuota 创建用户配额
func (qm *QuotaManager) CreateUserQuota(userID string, quota float64) error {
	qm.quotasMu.Lock()
	defer qm.quotasMu.Unlock()

	if _, exists := qm.userQuotas[userID]; exists {
		return fmt.Errorf("quota for user %s already exists", userID)
	}

	qm.userQuotas[userID] = &UserQuota{
		UserID:         userID,
		TotalQuota:     quota,
		AvailableQuota: quota,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return nil
}

// GetUserQuota 获取用户配额
func (qm *QuotaManager) GetUserQuota(userID string) (*UserQuota, error) {
	qm.quotasMu.RLock()
	defer qm.quotasMu.RUnlock()

	quota, ok := qm.userQuotas[userID]
	if !ok {
		return nil, fmt.Errorf("quota for user %s not found", userID)
	}

	return quota, nil
}

// PreDeduct 预扣费（冻结配额）
func (qm *QuotaManager) PreDeduct(userID string, recordID string, amount float64, description string) error {
	quota, err := qm.GetUserQuota(userID)
	if err != nil {
		return err
	}

	// 冻结配额
	if err := quota.Freeze(amount); err != nil {
		return err
	}

	// 记录预扣费
	record := &BillingRecord{
		RecordID:    recordID,
		UserID:      userID,
		Operation:   OpPreDeduction,
		Amount:      amount,
		Description: description,
		CreatedAt:   time.Now(),
		Status:      "pending",
	}

	qm.recordsMu.Lock()
	qm.records[recordID] = record
	qm.recordsMu.Unlock()

	// 记录待确认
	qm.pendingMu.Lock()
	qm.pendingDeductions[recordID] = amount
	qm.pendingMu.Unlock()

	qm.logFunc("info", fmt.Sprintf("Pre-deducted %.4f for user %s (record: %s)", amount, userID, recordID))

	return nil
}

// ConfirmDeduction 确认扣费
func (qm *QuotaManager) ConfirmDeduction(userID string, recordID string, actualAmount float64) error {
	quota, err := qm.GetUserQuota(userID)
	if err != nil {
		return err
	}

	// 获取原始冻结金额
	qm.pendingMu.RLock()
	originalAmount, ok := qm.pendingDeductions[recordID]
	qm.pendingMu.RUnlock()

	if !ok {
		return fmt.Errorf("no pending deduction found for record %s", recordID)
	}

	// 确认实际金额
	if err := quota.Confirm(actualAmount); err != nil {
		return err
	}

	// 如果实际金额少于冻结金额，退款差额
	refundAmount := originalAmount - actualAmount
	if refundAmount > 0 {
		if err := quota.Release(refundAmount); err != nil {
			return err
		}
	}

	// 更新记录状态
	qm.recordsMu.Lock()
	if record, ok := qm.records[recordID]; ok {
		record.Status = "confirmed"
		record.Amount = actualAmount
		record.CompletedAt = timePtr(time.Now())
		qm.records[recordID] = record
	}
	qm.recordsMu.Unlock()

	// 移除待确认
	qm.pendingMu.Lock()
	delete(qm.pendingDeductions, recordID)
	qm.pendingMu.Unlock()

	atomic.AddInt64(&qm.totalDeductions, 1)

	qm.logFunc("info", fmt.Sprintf("Confirmed deduction %.4f for user %s (record: %s)", actualAmount, userID, recordID))

	return nil
}

// RefundDeduction 退款预扣费
func (qm *QuotaManager) RefundDeduction(userID string, recordID string) error {
	quota, err := qm.GetUserQuota(userID)
	if err != nil {
		return err
	}

	// 获取冻结金额
	qm.pendingMu.RLock()
	frozenAmount, ok := qm.pendingDeductions[recordID]
	qm.pendingMu.RUnlock()

	if !ok {
		return fmt.Errorf("no pending deduction found for record %s", recordID)
	}

	// 释放冻结的配额
	if err := quota.Release(frozenAmount); err != nil {
		return err
	}

	// 更新记录状态
	qm.recordsMu.Lock()
	if record, ok := qm.records[recordID]; ok {
		record.Status = "refunded"
		record.CompletedAt = timePtr(time.Now())
		qm.records[recordID] = record
	}
	qm.recordsMu.Unlock()

	// 移除待确认
	qm.pendingMu.Lock()
	delete(qm.pendingDeductions, recordID)
	qm.pendingMu.Unlock()

	atomic.AddInt64(&qm.totalRefunds, 1)

	qm.logFunc("info", fmt.Sprintf("Refunded %.4f for user %s (record: %s)", frozenAmount, userID, recordID))

	return nil
}

// Recharge 充值
func (qm *QuotaManager) Recharge(userID string, amount float64) error {
	quota, err := qm.GetUserQuota(userID)
	if err != nil {
		return err
	}

	quota.mu.Lock()
	defer quota.mu.Unlock()

	quota.TotalQuota += amount
	quota.AvailableQuota += amount
	quota.UpdatedAt = time.Now()

	atomic.AddInt64(&qm.totalRecharges, 1)

	qm.logFunc("info", fmt.Sprintf("Recharged %.4f for user %s", amount, userID))

	return nil
}

// GetBillingRecord 获取账单记录
func (qm *QuotaManager) GetBillingRecord(recordID string) (*BillingRecord, error) {
	qm.recordsMu.RLock()
	defer qm.recordsMu.RUnlock()

	record, ok := qm.records[recordID]
	if !ok {
		return nil, fmt.Errorf("billing record %s not found", recordID)
	}

	return record, nil
}

// GetUserBillingRecords 获取用户的账单记录
func (qm *QuotaManager) GetUserBillingRecords(userID string) []*BillingRecord {
	qm.recordsMu.RLock()
	defer qm.recordsMu.RUnlock()

	var records []*BillingRecord
	for _, record := range qm.records {
		if record.UserID == userID {
			records = append(records, record)
		}
	}

	return records
}

// GetStatistics 获取统计信息
func (qm *QuotaManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_deductions":   atomic.LoadInt64(&qm.totalDeductions),
		"total_refunds":      atomic.LoadInt64(&qm.totalRefunds),
		"total_recharges":    atomic.LoadInt64(&qm.totalRecharges),
		"pending_deductions": len(qm.pendingDeductions),
		"total_users":        len(qm.userQuotas),
		"total_records":      len(qm.records),
	}
}

// GetRefundAccuracy 获取退款准确率（确认数 / （确认数 + 退款数））
func (qm *QuotaManager) GetRefundAccuracy() float64 {
	deductions := atomic.LoadInt64(&qm.totalDeductions)
	refunds := atomic.LoadInt64(&qm.totalRefunds)
	total := deductions + refunds

	if total == 0 {
		return 100.0
	}

	return float64(deductions) / float64(total) * 100.0
}

// 辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}

// BillingTransaction 计费事务（支持事务性操作）
type BillingTransaction struct {
	// 管理器
	manager *QuotaManager

	// 用户 ID
	userID string

	// 操作列表
	operations []*transactionOp

	// 互斥锁
	mu sync.RWMutex

	// 是否已提交
	committed bool

	// 是否已回滚
	rolledBack bool
}

// transactionOp 事务操作
type transactionOp struct {
	// 操作类型
	opType string

	// 参数
	params map[string]interface{}
}

// NewBillingTransaction 创建计费事务
func NewBillingTransaction(manager *QuotaManager, userID string) *BillingTransaction {
	return &BillingTransaction{
		manager:    manager,
		userID:     userID,
		operations: make([]*transactionOp, 0),
	}
}

// PreDeduct 添加预扣费操作
func (bt *BillingTransaction) PreDeduct(recordID string, amount float64, description string) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.committed || bt.rolledBack {
		return
	}

	bt.operations = append(bt.operations, &transactionOp{
		opType: "pre_deduct",
		params: map[string]interface{}{
			"recordID":    recordID,
			"amount":      amount,
			"description": description,
		},
	})
}

// Confirm 添加确认操作
func (bt *BillingTransaction) Confirm(recordID string, actualAmount float64) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.committed || bt.rolledBack {
		return
	}

	bt.operations = append(bt.operations, &transactionOp{
		opType: "confirm",
		params: map[string]interface{}{
			"recordID":     recordID,
			"actualAmount": actualAmount,
		},
	})
}

// Commit 提交事务
func (bt *BillingTransaction) Commit() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.committed {
		return fmt.Errorf("transaction already committed")
	}

	if bt.rolledBack {
		return fmt.Errorf("transaction already rolled back")
	}

	// 执行所有操作
	for _, op := range bt.operations {
		switch op.opType {
		case "pre_deduct":
			recordID := op.params["recordID"].(string)
			amount := op.params["amount"].(float64)
			description := op.params["description"].(string)
			if err := bt.manager.PreDeduct(bt.userID, recordID, amount, description); err != nil {
				// 回滚
				bt.Rollback()
				return err
			}

		case "confirm":
			recordID := op.params["recordID"].(string)
			actualAmount := op.params["actualAmount"].(float64)
			if err := bt.manager.ConfirmDeduction(bt.userID, recordID, actualAmount); err != nil {
				// 回滚
				bt.Rollback()
				return err
			}
		}
	}

	bt.committed = true
	return nil
}

// Rollback 回滚事务
func (bt *BillingTransaction) Rollback() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if bt.rolledBack {
		return fmt.Errorf("transaction already rolled back")
	}

	// 执行逆向操作（简化实现）
	for _, op := range bt.operations {
		if op.opType == "pre_deduct" {
			recordID := op.params["recordID"].(string)
			// 这里应该回滚，但简化实现中仅标记
			_ = recordID
		}
	}

	bt.rolledBack = true
	return nil
}

