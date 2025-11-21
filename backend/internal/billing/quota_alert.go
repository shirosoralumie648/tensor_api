package billing

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AlertLevel 预警等级
type AlertLevel int

const (
	// 正常
	AlertLevelNormal AlertLevel = iota
	// 警告（使用量达到 70%）
	AlertLevelWarning
	// 严重（使用量达到 90%）
	AlertLevelCritical
	// 已耗尽（配额用尽）
	AlertLevelExhausted
)

// String 返回警告等级的字符串表示
func (al AlertLevel) String() string {
	switch al {
	case AlertLevelNormal:
		return "normal"
	case AlertLevelWarning:
		return "warning"
	case AlertLevelCritical:
		return "critical"
	case AlertLevelExhausted:
		return "exhausted"
	default:
		return "unknown"
	}
}

// AlertRule 预警规则
type AlertRule struct {
	// 规则 ID
	RuleID string

	// 用户 ID
	UserID string

	// 触发阈值（百分比，0-100）
	Threshold float64

	// 预警等级
	Level AlertLevel

	// 是否启用
	Enabled bool

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time

	// 备注
	Remark string
}

// QuotaAlert 配额警告
type QuotaAlert struct {
	// 警告 ID
	AlertID string

	// 用户 ID
	UserID string

	// 警告等级
	Level AlertLevel

	// 使用率（百分比）
	UsageRate float64

	// 剩余配额
	RemainingQuota float64

	// 已使用配额
	UsedQuota float64

	// 总配额
	TotalQuota float64

	// 触发时间
	TriggeredAt time.Time

	// 是否已处理
	Handled bool

	// 处理时间
	HandledAt *time.Time

	// 消息
	Message string
}

// QuotaExpiryPolicy 配额有效期策略
type QuotaExpiryPolicy struct {
	// 策略 ID
	PolicyID string

	// 用户 ID
	UserID string

	// 配额获得日期
	AcquiredAt time.Time

	// 有效期天数（0 表示永久）
	ExpiryDays int

	// 过期时间
	ExpiresAt *time.Time

	// 是否启用
	Enabled bool

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time
}

// AlertManager 预警管理器
type AlertManager struct {
	// 预警规则映射
	rules map[string]*AlertRule
	rulesMu sync.RWMutex

	// 用户预警列表
	userAlerts map[string][]*QuotaAlert
	alertsMu sync.RWMutex

	// 配额有效期策略
	expiryPolicies map[string]*QuotaExpiryPolicy
	expiryMu sync.RWMutex

	// 配额管理器
	quotaManager *QuotaManager

	// 预警回调函数
	callbacks map[AlertLevel][]func(*QuotaAlert)
	callbacksMu sync.RWMutex

	// 统计信息
	alertCount int64
	handledCount int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewAlertManager 创建预警管理器
func NewAlertManager(quotaManager *QuotaManager) *AlertManager {
	return &AlertManager{
		rules:           make(map[string]*AlertRule),
		userAlerts:      make(map[string][]*QuotaAlert),
		expiryPolicies:  make(map[string]*QuotaExpiryPolicy),
		quotaManager:    quotaManager,
		callbacks:       make(map[AlertLevel][]func(*QuotaAlert)),
		logFunc:         defaultLogFunc,
	}
}

// CreateAlertRule 创建预警规则
func (am *AlertManager) CreateAlertRule(ruleID, userID string, threshold float64, level AlertLevel) error {
	am.rulesMu.Lock()
	defer am.rulesMu.Unlock()

	if _, exists := am.rules[ruleID]; exists {
		return fmt.Errorf("alert rule %s already exists", ruleID)
	}

	if threshold < 0 || threshold > 100 {
		return fmt.Errorf("threshold must be between 0 and 100, got %.2f", threshold)
	}

	am.rules[ruleID] = &AlertRule{
		RuleID:    ruleID,
		UserID:    userID,
		Threshold: threshold,
		Level:     level,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	am.logFunc("info", fmt.Sprintf("Created alert rule %s for user %s (threshold: %.2f%%)", ruleID, userID, threshold))

	return nil
}

// GetAlertRule 获取预警规则
func (am *AlertManager) GetAlertRule(ruleID string) (*AlertRule, error) {
	am.rulesMu.RLock()
	defer am.rulesMu.RUnlock()

	rule, ok := am.rules[ruleID]
	if !ok {
		return nil, fmt.Errorf("alert rule %s not found", ruleID)
	}

	return rule, nil
}

// DisableAlertRule 禁用预警规则
func (am *AlertManager) DisableAlertRule(ruleID string) error {
	am.rulesMu.Lock()
	defer am.rulesMu.Unlock()

	rule, ok := am.rules[ruleID]
	if !ok {
		return fmt.Errorf("alert rule %s not found", ruleID)
	}

	rule.Enabled = false
	rule.UpdatedAt = time.Now()

	am.logFunc("info", fmt.Sprintf("Disabled alert rule %s", ruleID))

	return nil
}

// CheckQuotaUsage 检查配额使用情况并触发预警
func (am *AlertManager) CheckQuotaUsage(userID string) error {
	quota, err := am.quotaManager.GetUserQuota(userID)
	if err != nil {
		return err
	}

	quota.mu.RLock()
	used := quota.UsedQuota
	total := quota.TotalQuota
	available := quota.AvailableQuota
	quota.mu.RUnlock()

	usageRate := (used / total) * 100.0

	// 获取用户的预警规则
	am.rulesMu.RLock()
	var userRules []*AlertRule
	for _, rule := range am.rules {
		if rule.UserID == userID && rule.Enabled {
			userRules = append(userRules, rule)
		}
	}
	am.rulesMu.RUnlock()

	// 检查是否触发预警
	for _, rule := range userRules {
		if usageRate >= rule.Threshold {
			alert := &QuotaAlert{
				AlertID:        fmt.Sprintf("alert-%s-%d", userID, time.Now().UnixNano()),
				UserID:         userID,
				Level:          rule.Level,
				UsageRate:      usageRate,
				RemainingQuota: available,
				UsedQuota:      used,
				TotalQuota:     total,
				TriggeredAt:    time.Now(),
				Message:        fmt.Sprintf("Quota usage reached %.2f%%", usageRate),
			}

			// 记录警告
			am.alertsMu.Lock()
			if am.userAlerts[userID] == nil {
				am.userAlerts[userID] = make([]*QuotaAlert, 0)
			}
			am.userAlerts[userID] = append(am.userAlerts[userID], alert)
			am.alertsMu.Unlock()

			atomic.AddInt64(&am.alertCount, 1)

			// 触发回调
			am.triggerCallbacks(alert)

			am.logFunc("warn", fmt.Sprintf("Alert triggered for user %s: usage %.2f%% (level: %s)", userID, usageRate, rule.Level.String()))
		}
	}

	return nil
}

// RegisterCallback 注册预警回调
func (am *AlertManager) RegisterCallback(level AlertLevel, callback func(*QuotaAlert)) {
	am.callbacksMu.Lock()
	defer am.callbacksMu.Unlock()

	if am.callbacks[level] == nil {
		am.callbacks[level] = make([]func(*QuotaAlert), 0)
	}

	am.callbacks[level] = append(am.callbacks[level], callback)
}

// triggerCallbacks 触发预警回调
func (am *AlertManager) triggerCallbacks(alert *QuotaAlert) {
	am.callbacksMu.RLock()
	callbacks, ok := am.callbacks[alert.Level]
	am.callbacksMu.RUnlock()

	if !ok {
		return
	}

	for _, callback := range callbacks {
		go callback(alert)
	}
}

// HandleAlert 标记预警为已处理
func (am *AlertManager) HandleAlert(alertID string) error {
	am.alertsMu.Lock()
	defer am.alertsMu.Unlock()

	for _, alerts := range am.userAlerts {
		for _, alert := range alerts {
			if alert.AlertID == alertID {
				alert.Handled = true
				alert.HandledAt = timePtr(time.Now())
				atomic.AddInt64(&am.handledCount, 1)
				return nil
			}
		}
	}

	return fmt.Errorf("alert %s not found", alertID)
}

// GetUserAlerts 获取用户的预警列表
func (am *AlertManager) GetUserAlerts(userID string) []*QuotaAlert {
	am.alertsMu.RLock()
	defer am.alertsMu.RUnlock()

	if alerts, ok := am.userAlerts[userID]; ok {
		return alerts
	}

	return make([]*QuotaAlert, 0)
}

// SetExpiryPolicy 设置配额有效期策略
func (am *AlertManager) SetExpiryPolicy(userID string, expiryDays int) error {
	am.expiryMu.Lock()
	defer am.expiryMu.Unlock()

	policyID := fmt.Sprintf("policy-%s", userID)

	expiresAt := time.Now().AddDate(0, 0, expiryDays)

	am.expiryPolicies[policyID] = &QuotaExpiryPolicy{
		PolicyID:   policyID,
		UserID:     userID,
		AcquiredAt: time.Now(),
		ExpiryDays: expiryDays,
		ExpiresAt:  &expiresAt,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	am.logFunc("info", fmt.Sprintf("Set expiry policy for user %s: %d days", userID, expiryDays))

	return nil
}

// CheckQuotaExpiry 检查配额是否过期
func (am *AlertManager) CheckQuotaExpiry(userID string) (bool, *time.Time, error) {
	am.expiryMu.RLock()
	defer am.expiryMu.RUnlock()

	policyID := fmt.Sprintf("policy-%s", userID)
	policy, ok := am.expiryPolicies[policyID]

	if !ok {
		return false, nil, nil // 无过期策略
	}

	if !policy.Enabled {
		return false, nil, nil
	}

	if policy.ExpiryDays == 0 {
		return false, nil, nil // 永久有效
	}

	if policy.ExpiresAt != nil && time.Now().After(*policy.ExpiresAt) {
		return true, policy.ExpiresAt, nil
	}

	return false, policy.ExpiresAt, nil
}

// AutoRechargeManager 自动充值管理器
type AutoRechargeManager struct {
	// 配置映射
	configs map[string]*AutoRechargeConfig
	configsMu sync.RWMutex

	// 充值历史
	history map[string][]*RechargeRecord
	historyMu sync.RWMutex

	// 配额管理器
	quotaManager *QuotaManager

	// 统计信息
	rechargeCount int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// AutoRechargeConfig 自动充值配置
type AutoRechargeConfig struct {
	// 用户 ID
	UserID string

	// 触发阈值（百分比）
	TriggerThreshold float64

	// 充值金额
	RechargeAmount float64

	// 最大充值次数/周期
	MaxRechargePerPeriod int

	// 周期（天）
	PeriodDays int

	// 是否启用
	Enabled bool

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time
}

// RechargeRecord 充值记录
type RechargeRecord struct {
	// 记录 ID
	RecordID string

	// 用户 ID
	UserID string

	// 充值金额
	Amount float64

	// 触发原因
	Reason string

	// 充值时间
	CreatedAt time.Time
}

// NewAutoRechargeManager 创建自动充值管理器
func NewAutoRechargeManager(quotaManager *QuotaManager) *AutoRechargeManager {
	return &AutoRechargeManager{
		configs:      make(map[string]*AutoRechargeConfig),
		history:      make(map[string][]*RechargeRecord),
		quotaManager: quotaManager,
		logFunc:      defaultLogFunc,
	}
}

// CreateAutoRechargeConfig 创建自动充值配置
func (arm *AutoRechargeManager) CreateAutoRechargeConfig(userID string, triggerThreshold, rechargeAmount float64, maxRechargePerPeriod, periodDays int) error {
	arm.configsMu.Lock()
	defer arm.configsMu.Unlock()

	if triggerThreshold < 0 || triggerThreshold > 100 {
		return fmt.Errorf("trigger threshold must be between 0 and 100")
	}

	arm.configs[userID] = &AutoRechargeConfig{
		UserID:                  userID,
		TriggerThreshold:        triggerThreshold,
		RechargeAmount:          rechargeAmount,
		MaxRechargePerPeriod:    maxRechargePerPeriod,
		PeriodDays:              periodDays,
		Enabled:                 true,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	arm.logFunc("info", fmt.Sprintf("Created auto-recharge config for user %s (threshold: %.2f%%, amount: %.2f)", userID, triggerThreshold, rechargeAmount))

	return nil
}

// CheckAndRecharge 检查并执行自动充值
func (arm *AutoRechargeManager) CheckAndRecharge(userID string) error {
	arm.configsMu.RLock()
	config, ok := arm.configs[userID]
	arm.configsMu.RUnlock()

	if !ok || !config.Enabled {
		return nil
	}

	quota, err := arm.quotaManager.GetUserQuota(userID)
	if err != nil {
		return err
	}

	quota.mu.RLock()
	used := quota.UsedQuota
	total := quota.TotalQuota
	quota.mu.RUnlock()

	usageRate := (used / total) * 100.0

	if usageRate >= config.TriggerThreshold {
		// 检查周期内的充值次数
		arm.historyMu.RLock()
		records := arm.history[userID]
		arm.historyMu.RUnlock()

		// 统计最近周期内的充值次数
		rechargeCount := 0
		cutoffTime := time.Now().AddDate(0, 0, -config.PeriodDays)

		for _, record := range records {
			if record.CreatedAt.After(cutoffTime) {
				rechargeCount++
			}
		}

		if rechargeCount < config.MaxRechargePerPeriod {
			// 执行充值
			err := arm.quotaManager.Recharge(userID, config.RechargeAmount)
			if err != nil {
				return err
			}

			// 记录充值
			record := &RechargeRecord{
				RecordID:  fmt.Sprintf("recharge-%s-%d", userID, time.Now().UnixNano()),
				UserID:    userID,
				Amount:    config.RechargeAmount,
				Reason:    fmt.Sprintf("Auto-recharge triggered at usage %.2f%%", usageRate),
				CreatedAt: time.Now(),
			}

			arm.historyMu.Lock()
			if arm.history[userID] == nil {
				arm.history[userID] = make([]*RechargeRecord, 0)
			}
			arm.history[userID] = append(arm.history[userID], record)
			arm.historyMu.Unlock()

			atomic.AddInt64(&arm.rechargeCount, 1)

			arm.logFunc("info", fmt.Sprintf("Auto-recharged user %s with %.2f (usage: %.2f%%)", userID, config.RechargeAmount, usageRate))

			return nil
		}

		arm.logFunc("warn", fmt.Sprintf("Auto-recharge limit reached for user %s", userID))
		return fmt.Errorf("auto-recharge limit reached for this period")
	}

	return nil
}

// GetRechargeHistory 获取充值历史
func (arm *AutoRechargeManager) GetRechargeHistory(userID string) []*RechargeRecord {
	arm.historyMu.RLock()
	defer arm.historyMu.RUnlock()

	if records, ok := arm.history[userID]; ok {
		return records
	}

	return make([]*RechargeRecord, 0)
}

// GetStatistics 获取统计信息
func (arm *AutoRechargeManager) GetStatistics() map[string]interface{} {
	arm.configsMu.RLock()
	configCount := len(arm.configs)
	arm.configsMu.RUnlock()

	arm.historyMu.RLock()
	historyCount := 0
	for _, records := range arm.history {
		historyCount += len(records)
	}
	arm.historyMu.RUnlock()

	return map[string]interface{}{
		"config_count":    configCount,
		"history_count":   historyCount,
		"recharge_count":  atomic.LoadInt64(&arm.rechargeCount),
	}
}

