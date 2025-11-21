package billing

import (
	"testing"
)

func TestQuotaManagerCreateUserQuota(t *testing.T) {
	manager := NewQuotaManager()

	if err := manager.CreateUserQuota("user-1", 100.0); err != nil {
		t.Errorf("CreateUserQuota failed: %v", err)
	}

	quota, err := manager.GetUserQuota("user-1")
	if err != nil {
		t.Errorf("GetUserQuota failed: %v", err)
	}

	if quota.TotalQuota != 100.0 {
		t.Errorf("Expected quota 100.0, got %.2f", quota.TotalQuota)
	}

	if quota.AvailableQuota != 100.0 {
		t.Errorf("Expected available quota 100.0, got %.2f", quota.AvailableQuota)
	}
}

func TestQuotaManagerPreDeduct(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 预扣费 50
	if err := manager.PreDeduct("user-1", "rec-1", 50.0, "Test pre-deduction"); err != nil {
		t.Errorf("PreDeduct failed: %v", err)
	}

	quota, _ := manager.GetUserQuota("user-1")

	// 可用配额应该是 50
	if quota.GetAvailable() != 50.0 {
		t.Errorf("Expected available quota 50.0, got %.2f", quota.GetAvailable())
	}

	// 冻结配额应该是 50
	quota.mu.RLock()
	frozen := quota.FrozenQuota
	quota.mu.RUnlock()

	if frozen != 50.0 {
		t.Errorf("Expected frozen quota 50.0, got %.2f", frozen)
	}
}

func TestQuotaManagerConfirmDeduction(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 预扣费 50
	manager.PreDeduct("user-1", "rec-1", 50.0, "Test")

	// 确认扣费 40（退款 10）
	if err := manager.ConfirmDeduction("user-1", "rec-1", 40.0); err != nil {
		t.Errorf("ConfirmDeduction failed: %v", err)
	}

	quota, _ := manager.GetUserQuota("user-1")

	// 可用配额应该是 60（100 - 40）
	if quota.GetAvailable() != 60.0 {
		t.Errorf("Expected available quota 60.0, got %.2f", quota.GetAvailable())
	}

	// 已使用配额应该是 40
	quota.mu.RLock()
	used := quota.UsedQuota
	quota.mu.RUnlock()

	if used != 40.0 {
		t.Errorf("Expected used quota 40.0, got %.2f", used)
	}
}

func TestQuotaManagerRefund(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 预扣费 50
	manager.PreDeduct("user-1", "rec-1", 50.0, "Test")

	// 退款
	if err := manager.RefundDeduction("user-1", "rec-1"); err != nil {
		t.Errorf("RefundDeduction failed: %v", err)
	}

	quota, _ := manager.GetUserQuota("user-1")

	// 可用配额应该回到 100
	if quota.GetAvailable() != 100.0 {
		t.Errorf("Expected available quota 100.0, got %.2f", quota.GetAvailable())
	}
}

func TestQuotaManagerRecharge(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 充值 50
	if err := manager.Recharge("user-1", 50.0); err != nil {
		t.Errorf("Recharge failed: %v", err)
	}

	quota, _ := manager.GetUserQuota("user-1")

	// 总配额应该是 150
	if quota.TotalQuota != 150.0 {
		t.Errorf("Expected total quota 150.0, got %.2f", quota.TotalQuota)
	}

	// 可用配额应该是 150
	if quota.GetAvailable() != 150.0 {
		t.Errorf("Expected available quota 150.0, got %.2f", quota.GetAvailable())
	}
}

func TestQuotaManagerInsufficientQuota(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 尝试预扣费超过可用配额
	err := manager.PreDeduct("user-1", "rec-1", 150.0, "Test")

	if err == nil {
		t.Errorf("Expected error for insufficient quota")
	}
}

func TestBillingRecord(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 创建预扣费记录
	manager.PreDeduct("user-1", "rec-1", 50.0, "Test pre-deduction")

	// 获取记录
	record, err := manager.GetBillingRecord("rec-1")
	if err != nil {
		t.Errorf("GetBillingRecord failed: %v", err)
	}

	if record.Operation != OpPreDeduction {
		t.Errorf("Expected operation PreDeduction")
	}

	if record.Amount != 50.0 {
		t.Errorf("Expected amount 50.0, got %.2f", record.Amount)
	}

	if record.Status != "pending" {
		t.Errorf("Expected status pending, got %s", record.Status)
	}
}

func TestBillingTransaction(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 创建事务
	tx := NewBillingTransaction(manager, "user-1")

	// 添加操作
	tx.PreDeduct("rec-1", 50.0, "Test transaction")
	tx.Confirm("rec-1", 40.0)

	// 提交事务
	if err := tx.Commit(); err != nil {
		t.Errorf("Commit failed: %v", err)
	}

	quota, _ := manager.GetUserQuota("user-1")

	// 可用配额应该是 60
	if quota.GetAvailable() != 60.0 {
		t.Errorf("Expected available quota 60.0, got %.2f", quota.GetAvailable())
	}
}

func TestQuotaManagerStatistics(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 执行一些操作
	manager.PreDeduct("user-1", "rec-1", 50.0, "Test 1")
	manager.ConfirmDeduction("user-1", "rec-1", 40.0)

	manager.PreDeduct("user-1", "rec-2", 30.0, "Test 2")
	manager.RefundDeduction("user-1", "rec-2")

	manager.Recharge("user-1", 50.0)

	stats := manager.GetStatistics()

	if deductions, ok := stats["total_deductions"].(int64); !ok || deductions != 1 {
		t.Errorf("Expected 1 deduction")
	}

	if refunds, ok := stats["total_refunds"].(int64); !ok || refunds != 1 {
		t.Errorf("Expected 1 refund")
	}

	if recharges, ok := stats["total_recharges"].(int64); !ok || recharges != 1 {
		t.Errorf("Expected 1 recharge")
	}
}

func TestRefundAccuracy(t *testing.T) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 100.0)

	// 执行一些操作
	manager.PreDeduct("user-1", "rec-1", 50.0, "Test 1")
	manager.ConfirmDeduction("user-1", "rec-1", 50.0)

	manager.PreDeduct("user-1", "rec-2", 30.0, "Test 2")
	manager.RefundDeduction("user-1", "rec-2")

	accuracy := manager.GetRefundAccuracy()

	// 准确率应该是 50% (1 确认 / 2 总操作)
	if accuracy < 49.9 || accuracy > 50.1 {
		t.Errorf("Expected accuracy ~50%%, got %.2f%%", accuracy)
	}
}

func TestUserQuotaFreeze(t *testing.T) {
	quota := &UserQuota{
		UserID:         "user-1",
		TotalQuota:     100.0,
		AvailableQuota: 100.0,
	}

	// 冻结 50
	if err := quota.Freeze(50.0); err != nil {
		t.Errorf("Freeze failed: %v", err)
	}

	if quota.AvailableQuota != 50.0 {
		t.Errorf("Expected available 50.0, got %.2f", quota.AvailableQuota)
	}

	if quota.FrozenQuota != 50.0 {
		t.Errorf("Expected frozen 50.0, got %.2f", quota.FrozenQuota)
	}
}

func BenchmarkQuotaManagerPreDeduct(b *testing.B) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 1000000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.PreDeduct("user-1", "rec-"+string(rune(i)), 1.0, "Test")
	}
}

func BenchmarkQuotaManagerConfirm(b *testing.B) {
	manager := NewQuotaManager()
	manager.CreateUserQuota("user-1", 1000000.0)

	// 预先创建一些记录
	for i := 0; i < b.N; i++ {
		recordID := "rec-" + string(rune(i))
		manager.PreDeduct("user-1", recordID, 1.0, "Test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recordID := "rec-" + string(rune(i))
		_ = manager.ConfirmDeduction("user-1", recordID, 0.9)
	}
}

