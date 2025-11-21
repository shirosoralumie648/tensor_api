package billing

import (
	"testing"
	"time"
)

func TestAlertRuleCreation(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	alertManager := NewAlertManager(quotaManager)

	err := alertManager.CreateAlertRule("rule-1", "user-1", 70.0, AlertLevelWarning)
	if err != nil {
		t.Errorf("CreateAlertRule failed: %v", err)
	}

	rule, err := alertManager.GetAlertRule("rule-1")
	if err != nil {
		t.Errorf("GetAlertRule failed: %v", err)
	}

	if rule.Threshold != 70.0 || rule.Level != AlertLevelWarning {
		t.Errorf("Rule data mismatch")
	}
}

func TestAlertRuleValidation(t *testing.T) {
	quotaManager := NewQuotaManager()
	alertManager := NewAlertManager(quotaManager)

	// 测试无效的阈值
	err := alertManager.CreateAlertRule("rule-1", "user-1", 150.0, AlertLevelWarning)
	if err == nil {
		t.Errorf("Expected error for invalid threshold")
	}
}

func TestCheckQuotaUsage(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)
	quotaManager.PreDeduct("user-1", "req-1", 75.0, "Test")

	alertManager := NewAlertManager(quotaManager)
	alertManager.CreateAlertRule("rule-1", "user-1", 70.0, AlertLevelWarning)

	// 检查配额使用
	err := alertManager.CheckQuotaUsage("user-1")
	if err != nil {
		t.Errorf("CheckQuotaUsage failed: %v", err)
	}

	// 应该触发预警
	alerts := alertManager.GetUserAlerts("user-1")
	if len(alerts) == 0 {
		t.Errorf("Expected at least 1 alert")
	}
}

func TestDisableAlertRule(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	alertManager := NewAlertManager(quotaManager)
	alertManager.CreateAlertRule("rule-1", "user-1", 70.0, AlertLevelWarning)

	// 禁用规则
	err := alertManager.DisableAlertRule("rule-1")
	if err != nil {
		t.Errorf("DisableAlertRule failed: %v", err)
	}

	rule, _ := alertManager.GetAlertRule("rule-1")
	if rule.Enabled {
		t.Errorf("Rule should be disabled")
	}
}

func TestQuotaExpiryPolicy(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	alertManager := NewAlertManager(quotaManager)

	// 设置配额有效期为 30 天
	err := alertManager.SetExpiryPolicy("user-1", 30)
	if err != nil {
		t.Errorf("SetExpiryPolicy failed: %v", err)
	}

	// 检查配额是否过期
	expired, expiresAt, err := alertManager.CheckQuotaExpiry("user-1")
	if err != nil {
		t.Errorf("CheckQuotaExpiry failed: %v", err)
	}

	if expired {
		t.Errorf("Quota should not be expired yet")
	}

	if expiresAt == nil {
		t.Errorf("Expected expiry time")
	}
}

func TestAutoRechargeConfig(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	autoRechargeManager := NewAutoRechargeManager(quotaManager)

	// 创建自动充值配置
	err := autoRechargeManager.CreateAutoRechargeConfig("user-1", 80.0, 50.0, 5, 7)
	if err != nil {
		t.Errorf("CreateAutoRechargeConfig failed: %v", err)
	}
}

func TestAutoRechargeValidation(t *testing.T) {
	quotaManager := NewQuotaManager()
	autoRechargeManager := NewAutoRechargeManager(quotaManager)

	// 测试无效的触发阈值
	err := autoRechargeManager.CreateAutoRechargeConfig("user-1", 150.0, 50.0, 5, 7)
	if err == nil {
		t.Errorf("Expected error for invalid threshold")
	}
}

func TestAutoRecharge(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	autoRechargeManager := NewAutoRechargeManager(quotaManager)
	autoRechargeManager.CreateAutoRechargeConfig("user-1", 50.0, 50.0, 5, 7)

	// 模拟使用 60% 配额
	quotaManager.PreDeduct("user-1", "req-1", 60.0, "Test")

	// 执行自动充值检查
	err := autoRechargeManager.CheckAndRecharge("user-1")
	if err != nil {
		t.Errorf("CheckAndRecharge failed: %v", err)
	}

	// 检查充值历史
	history := autoRechargeManager.GetRechargeHistory("user-1")
	if len(history) == 0 {
		t.Errorf("Expected at least 1 recharge record")
	}
}

func TestAlertCallback(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	alertManager := NewAlertManager(quotaManager)
	alertManager.CreateAlertRule("rule-1", "user-1", 70.0, AlertLevelWarning)

	callbackCalled := false

	// 注册回调
	alertManager.RegisterCallback(AlertLevelWarning, func(alert *QuotaAlert) {
		callbackCalled = true
	})

	// 预扣费 75
	quotaManager.PreDeduct("user-1", "req-1", 75.0, "Test")

	// 检查配额使用
	alertManager.CheckQuotaUsage("user-1")

	// 等待异步回调
	time.Sleep(100 * time.Millisecond)

	if !callbackCalled {
		t.Errorf("Callback should have been called")
	}
}

func TestHandleAlert(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)
	quotaManager.PreDeduct("user-1", "req-1", 75.0, "Test")

	alertManager := NewAlertManager(quotaManager)
	alertManager.CreateAlertRule("rule-1", "user-1", 70.0, AlertLevelWarning)

	// 检查配额使用并触发预警
	alertManager.CheckQuotaUsage("user-1")

	alerts := alertManager.GetUserAlerts("user-1")
	if len(alerts) == 0 {
		t.Errorf("Expected at least 1 alert")
	}

	// 标记预警为已处理
	alertID := alerts[0].AlertID
	err := alertManager.HandleAlert(alertID)
	if err != nil {
		t.Errorf("HandleAlert failed: %v", err)
	}

	// 验证预警已处理
	alerts = alertManager.GetUserAlerts("user-1")
	if !alerts[0].Handled {
		t.Errorf("Alert should be marked as handled")
	}
}

func TestRechargeHistoryTracking(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	autoRechargeManager := NewAutoRechargeManager(quotaManager)
	autoRechargeManager.CreateAutoRechargeConfig("user-1", 50.0, 50.0, 5, 7)

	// 执行多次充值
	quotaManager.PreDeduct("user-1", "req-1", 60.0, "Test")
	autoRechargeManager.CheckAndRecharge("user-1")

	quotaManager.PreDeduct("user-1", "req-2", 60.0, "Test")
	autoRechargeManager.CheckAndRecharge("user-1")

	history := autoRechargeManager.GetRechargeHistory("user-1")
	if len(history) < 2 {
		t.Errorf("Expected at least 2 recharge records")
	}
}

func TestAlertLevelString(t *testing.T) {
	levels := map[AlertLevel]string{
		AlertLevelNormal:    "normal",
		AlertLevelWarning:   "warning",
		AlertLevelCritical:  "critical",
		AlertLevelExhausted: "exhausted",
	}

	for level, expected := range levels {
		if level.String() != expected {
			t.Errorf("Expected %s, got %s", expected, level.String())
		}
	}
}

func TestAlertStatistics(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	alertManager := NewAlertManager(quotaManager)

	stats := alertManager.GetStatistics()
	if alertCount, ok := stats["alert_count"].(int64); !ok || alertCount < 0 {
		t.Logf("Alert stats: %v", stats)
	}
}

func TestAutoRechargeStatistics(t *testing.T) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 100.0)

	autoRechargeManager := NewAutoRechargeManager(quotaManager)
	autoRechargeManager.CreateAutoRechargeConfig("user-1", 50.0, 50.0, 5, 7)

	stats := autoRechargeManager.GetStatistics()
	if configCount, ok := stats["config_count"].(int); !ok || configCount != 1 {
		t.Errorf("Expected 1 config")
	}
}

func BenchmarkCheckQuotaUsage(b *testing.B) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 1000000.0)
	quotaManager.PreDeduct("user-1", "req-1", 500000.0, "Test")

	alertManager := NewAlertManager(quotaManager)
	alertManager.CreateAlertRule("rule-1", "user-1", 50.0, AlertLevelWarning)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = alertManager.CheckQuotaUsage("user-1")
	}
}

func BenchmarkAutoRecharge(b *testing.B) {
	quotaManager := NewQuotaManager()
	quotaManager.CreateUserQuota("user-1", 1000000.0)

	autoRechargeManager := NewAutoRechargeManager(quotaManager)
	autoRechargeManager.CreateAutoRechargeConfig("user-1", 50.0, 50000.0, 1000, 7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = autoRechargeManager.CheckAndRecharge("user-1")
	}
}

