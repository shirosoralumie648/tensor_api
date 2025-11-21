package service

import (
	"context"
	"fmt"
	"time"

	"github.com/oblivious/backend/internal/model"
	"gorm.io/gorm"
)

// AdvancedBillingService 高级计费服务
type AdvancedBillingService struct {
	db *gorm.DB
}

// NewAdvancedBillingService 创建高级计费服务
func NewAdvancedBillingService(db *gorm.DB) *AdvancedBillingService {
	return &AdvancedBillingService{db: db}
}

// RecordUsage 记录 Token 使用
func (s *AdvancedBillingService) RecordUsage(ctx context.Context, record *model.BillingRecord) error {
	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	// 更新订阅的 Token 使用统计
	if record.SubscriptionID != "" {
		if err := s.db.WithContext(ctx).
			Model(&model.Subscription{}).
			Where("id = ?", record.SubscriptionID).
			Update("tokens_used", gorm.Expr("tokens_used + ?", record.TotalTokens)).Error; err != nil {
			// 只记录日志，不影响主流程
			fmt.Printf("failed to update subscription tokens: %v\n", err)
		}
	}

	return nil
}

// CalculateCost 计算 Token 成本
func (s *AdvancedBillingService) CalculateCost(ctx context.Context, userID, model string, promptTokens, completionTokens int) (float32, error) {
	// 获取用户订阅信息
	var subscription model.Subscription
	if err := s.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ? AND status = ?", userID, "active").
		First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("no active subscription found for user %s", userID)
		}
		return 0, err
	}

	// 获取模型价格
	var modelPrice model.ModelPrice
	if err := s.db.WithContext(ctx).
		Where("model_id = ? AND is_active = ?", model, true).
		First(&modelPrice).Error; err != nil {
		return 0, fmt.Errorf("model price not found for %s", model)
	}

	// 计算成本
	promptCost := float32(promptTokens) / 1000 * modelPrice.PromptPricePerK
	completionCost := float32(completionTokens) / 1000 * modelPrice.CompletionPricePerK
	baseCost := promptCost + completionCost

	// 应用订阅折扣
	discountRate := float32(1.0)
	switch subscription.Plan.Name {
	case "pro":
		discountRate = 0.9 // 9 折
	case "enterprise":
		discountRate = 0.8 // 8 折
	}

	return baseCost * discountRate, nil
}

// ApplyCoupon 应用优惠券
func (s *AdvancedBillingService) ApplyCoupon(ctx context.Context, userID, couponCode string) (*model.Coupon, error) {
	// 查找优惠券
	var coupon model.Coupon
	if err := s.db.WithContext(ctx).
		Where("code = ?", couponCode).
		First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("coupon not found: %s", couponCode)
		}
		return nil, err
	}

	// 检查优惠券状态
	if !coupon.IsActive {
		return nil, fmt.Errorf("coupon is not active")
	}

	// 检查过期时间
	if time.Now().After(coupon.ExpiresAt) {
		return nil, fmt.Errorf("coupon has expired")
	}

	// 检查使用次数
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		return nil, fmt.Errorf("coupon max uses reached")
	}

	// 检查用户使用次数
	var userUsageCount int64
	if err := s.db.WithContext(ctx).
		Model(&model.CouponUsage{}).
		Where("coupon_id = ? AND user_id = ?", coupon.ID, userID).
		Count(&userUsageCount).Error; err != nil {
		return nil, err
	}

	if userUsageCount >= int64(coupon.MaxUsagePerUser) {
		return nil, fmt.Errorf("coupon max uses per user reached")
	}

	// 更新优惠券使用次数
	if err := s.db.WithContext(ctx).
		Model(&coupon).
		Update("used_count", gorm.Expr("used_count + ?", 1)).Error; err != nil {
		return nil, err
	}

	// 记录使用
	usage := &model.CouponUsage{
		CouponID:       coupon.ID,
		UserID:         userID,
		DiscountAmount: calculateDiscountAmount(&coupon),
		UsedAt:         time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(usage).Error; err != nil {
		return nil, err
	}

	return &coupon, nil
}

// CreateInvoice 生成发票
func (s *AdvancedBillingService) CreateInvoice(ctx context.Context, userID string, month string) (*model.Invoice, error) {
	// 查询该月的所有计费记录
	var records []model.BillingRecord
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND billing_month = ?", userID, month).
		Find(&records).Error; err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no billing records found for user %s in month %s", userID, month)
	}

	// 计算总金额和汇总信息
	var totalAmount float32
	invoiceItems := make([]model.InvoiceItem, 0)
	modelUsageMap := make(map[string]*model.InvoiceItem)

	for _, record := range records {
		totalAmount += record.FinalCost

		// 按模型汇总
		if item, ok := modelUsageMap[record.Model]; ok {
			item.Quantity += record.TotalTokens
			item.Amount += record.FinalCost
		} else {
			// 获取模型价格
			var modelPrice model.ModelPrice
			s.db.WithContext(ctx).
				Where("model_id = ?", record.Model).
				First(&modelPrice)

			invoiceItems = append(invoiceItems, model.InvoiceItem{
				Model:       record.Model,
				Quantity:    record.TotalTokens,
				UnitPrice:   modelPrice.PromptPricePerK,
				Amount:      record.FinalCost,
				Description: fmt.Sprintf("AI API usage - %s", record.Model),
			})

			modelUsageMap[record.Model] = &invoiceItems[len(invoiceItems)-1]
		}
	}

	// 创建发票
	invoiceNumber := generateInvoiceNumber(userID, month)
	now := time.Now()
	dueDate := now.AddDate(0, 0, 30) // 30 天后到期

	invoiceJSON := model.JSONArray{}
	for _, item := range invoiceItems {
		invoiceJSON = append(invoiceJSON, item)
	}

	invoice := &model.Invoice{
		ID:            generateID(),
		UserID:        userID,
		BillingMonth:  month,
		Status:        "issued",
		Amount:        totalAmount,
		FinalAmount:   totalAmount,
		Currency:      "USD",
		IssuedAt:      &now,
		DueDate:       &dueDate,
		InvoiceNumber: invoiceNumber,
		Items:         invoiceJSON,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// GetUserBillingStats 获取用户计费统计
func (s *AdvancedBillingService) GetUserBillingStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取当月使用统计
	var currentMonthCost float32
	var currentMonthTokens int64
	month := time.Now().Format("2006-01")

	if err := s.db.WithContext(ctx).
		Model(&model.BillingRecord{}).
		Where("user_id = ? AND billing_month = ?", userID, month).
		Select("SUM(final_cost) as cost, SUM(total_tokens) as tokens").
		Row().
		Scan(&currentMonthCost, &currentMonthTokens); err != nil {
		return nil, err
	}

	stats["current_month_cost"] = currentMonthCost
	stats["current_month_tokens"] = currentMonthTokens

	// 获取用户订阅信息
	var subscription model.Subscription
	if err := s.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ? AND status = ?", userID, "active").
		First(&subscription).Error; err == nil {
		stats["subscription_plan"] = subscription.Plan.Name
		stats["tokens_quota"] = subscription.Plan.MonthlyQuota
		stats["tokens_used"] = subscription.TokensUsed
		stats["tokens_remaining"] = subscription.Plan.MonthlyQuota - subscription.TokensUsed
	}

	// 获取总消费（生命周期）
	var totalCost float32
	s.db.WithContext(ctx).
		Model(&model.BillingRecord{}).
		Where("user_id = ?", userID).
		Select("SUM(final_cost)").
		Row().
		Scan(&totalCost)

	stats["total_cost"] = totalCost

	return stats, nil
}

// CheckQuotaWarning 检查配额警告
func (s *AdvancedBillingService) CheckQuotaWarning(ctx context.Context, userID string) (bool, error) {
	// 获取用户设置
	var settings model.BillingSettings
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	if !settings.EnableAlerts || settings.AlertThreshold == 0 {
		return false, nil
	}

	// 获取当月 Token 使用
	var subscription model.Subscription
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		First(&subscription).Error; err != nil {
		return false, err
	}

	// 检查是否低于警告阈值
	remainingTokens := subscription.Plan.MonthlyQuota - subscription.TokensUsed
	return remainingTokens <= settings.AlertThreshold, nil
}

// AutoTopup 自动充值
func (s *AdvancedBillingService) AutoTopup(ctx context.Context, userID string) error {
	// 获取用户设置
	var settings model.BillingSettings
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND auto_topup = ?", userID, true).
		First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // 用户未启用自动充值
		}
		return err
	}

	// 检查是否需要自动充值
	var subscription model.Subscription
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		First(&subscription).Error; err != nil {
		return err
	}

	remainingTokens := subscription.Plan.MonthlyQuota - subscription.TokensUsed
	if remainingTokens > settings.AutoTopupThreshold {
		return nil // 不需要充值
	}

	// 执行充值（这里简化处理，实际应与支付系统集成）
	// TODO: 调用支付系统进行实际扣款

	// 记录充值
	record := &model.BillingRecord{
		ID:           generateID(),
		UserID:       userID,
		Model:        "topup",
		Provider:     "system",
		FinalCost:    settings.AutoTopupAmount,
		Status:       "completed",
		BillingMonth: time.Now().Format("2006-01"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to record topup: %w", err)
	}

	return nil
}

// 辅助函数

func calculateDiscountAmount(coupon *model.Coupon) float32 {
	switch coupon.Type {
	case "percentage":
		return coupon.Value / 100
	case "fixed":
		return coupon.Value
	default:
		return 0
	}
}

func generateInvoiceNumber(userID, month string) string {
	return fmt.Sprintf("INV-%s-%s", userID, month)
}

func generateID() string {
	// 使用 UUID 生成唯一 ID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
