package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/internal/repository"
)

// BillingService 计费服务
type BillingService struct {
	billingRepo    *repository.BillingLogRepository
	quotaLogRepo   *repository.QuotaLogRepository
	invoiceRepo    *repository.InvoiceRepository
	pricingRepo    *repository.PricingPlanRepository
	userRepo       *repository.UserRepository
	modelPriceRepo *repository.ModelPriceRepository
}

// NewBillingService 创建计费服务
func NewBillingService() *BillingService {
	return &BillingService{
		billingRepo:    repository.NewBillingLogRepository(),
		quotaLogRepo:   repository.NewQuotaLogRepository(),
		invoiceRepo:    repository.NewInvoiceRepository(),
		pricingRepo:    repository.NewPricingPlanRepository(),
		userRepo:       repository.NewUserRepository(),
		modelPriceRepo: repository.NewModelPriceRepository(),
	}
}

// CalculateCost 计算费用（返回分）
func (s *BillingService) CalculateCost(ctx context.Context, modelName string, inputTokens, outputTokens int) (int64, float64, error) {
	// 从 model_prices 表查询价格
	// 这里假设存在默认的渠道 ID 或者从某个配置中获取
	// 简化实现：直接从价格表中查询最便宜的价格
	price, err := s.modelPriceRepo.FindByModel(ctx, modelName)
	if err != nil {
		return 0, 0, err
	}
	if price == nil {
		return 0, 0, fmt.Errorf("pricing not found for model: %s", modelName)
	}

	// 计算费用
	// 输入：price.InputPrice（美元/1K tokens）
	// 输出：price.OutputPrice（美元/1K tokens）
	inputCostUSD := float64(inputTokens) / 1000.0 * price.InputPrice
	outputCostUSD := float64(outputTokens) / 1000.0 * price.OutputPrice
	totalCostUSD := inputCostUSD + outputCostUSD

	// 转换为分（1 美元 = 100 分，这里用最小单位）
	// 实际应该是 1 美元 = 100 分，但为了避免浮点精度问题，我们使用一个系数
	totalCostCents := int64(totalCostUSD * 10000) // 单位：0.0001 美元，即 1e-4 美元

	return totalCostCents, totalCostUSD, nil
}

// Charge 扣费并记录日志
func (s *BillingService) Charge(ctx context.Context, userID int, sessionID, messageID uuid.UUID, modelName string, inputTokens, outputTokens int) (*model.BillingLog, error) {
	// 1. 计算费用
	cost, costUSD, err := s.CalculateCost(ctx, modelName, inputTokens, outputTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cost: %w", err)
	}

	// 2. 检查用户额度是否足够
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %d", userID)
	}

	if user.Quota < cost {
		return nil, fmt.Errorf("insufficient quota: have %d, need %d", user.Quota, cost)
	}

	// 3. 创建计费日志
	log := &model.BillingLog{
		UserID:       userID,
		SessionID:    &sessionID,
		MessageID:    &messageID,
		Model:        modelName,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		Cost:         cost,
		CostUSD:      costUSD,
		Status:       1, // 已记录
	}

	if err := s.billingRepo.Create(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create billing log: %w", err)
	}

	// 4. 扣减用户额度
	if err := s.userRepo.DeductQuota(ctx, userID, cost); err != nil {
		return nil, fmt.Errorf("failed to deduct quota: %w", err)
	}

	// 5. 记录额度变更日志
	quotaLog := &model.QuotaLog{
		UserID:        userID,
		OperationType: "deduct",
		Amount:        cost,
		Reason:        fmt.Sprintf("AI service usage: %s", modelName),
		BillingLogID:  &log.ID,
		BalanceBefore: user.Quota,
		BalanceAfter:  user.Quota - cost,
	}

	if err := s.quotaLogRepo.Create(ctx, quotaLog); err != nil {
		// 只记录日志，不影响主流程
		return log, nil
	}

	// 6. 更新计费日志状态为已计费
	_ = s.billingRepo.UpdateStatus(ctx, log.ID, 2)

	return log, nil
}

// GetBillingHistory 获取用户计费历史
func (s *BillingService) GetBillingHistory(ctx context.Context, userID int, page int, pageSize int) ([]*model.BillingLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.billingRepo.FindByUserID(ctx, userID, pageSize, offset)
}

// GetQuotaHistory 获取用户额度变更历史
func (s *BillingService) GetQuotaHistory(ctx context.Context, userID int, page int, pageSize int) ([]*model.QuotaLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.quotaLogRepo.FindByUserID(ctx, userID, pageSize, offset)
}

// GetInvoices 获取用户发票列表
func (s *BillingService) GetInvoices(ctx context.Context, userID int, page int, pageSize int) ([]*model.Invoice, int64, error) {
	offset := (page - 1) * pageSize
	return s.invoiceRepo.FindByUserID(ctx, userID, pageSize, offset)
}

// GenerateInvoice 生成发票
func (s *BillingService) GenerateInvoice(ctx context.Context, userID int) (*model.Invoice, error) {
	// 获取本月的计费记录
	// 这是一个简化版本，实际应该按月份查询
	logs, total, err := s.billingRepo.FindByUserID(ctx, userID, 1000, 0)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return nil, fmt.Errorf("no billing logs found for user %d", userID)
	}

	// 计算总费用
	var totalCost int64
	var totalUSD float64
	for _, log := range logs {
		totalCost += log.Cost
		totalUSD += log.CostUSD
	}

	// 创建发票
	invoice := &model.Invoice{
		UserID:    userID,
		InvoiceNo: fmt.Sprintf("INV-%d-%d", userID, len(logs)),
		TotalCost: totalCost,
		TotalUSD:  totalUSD,
		ItemCount: len(logs),
		Status:    1, // 未支付
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}

// RefundCharge 退款
func (s *BillingService) RefundCharge(ctx context.Context, billingLogID int) error {
	// 查询计费记录
	log, err := s.billingRepo.FindByID(ctx, billingLogID)
	if err != nil {
		return err
	}
	if log == nil {
		return fmt.Errorf("billing log not found: %d", billingLogID)
	}

	// 退款金额
	refundAmount := log.Cost

	// 恢复用户额度
	if err := s.userRepo.AddQuota(ctx, log.UserID, refundAmount); err != nil {
		return fmt.Errorf("failed to add quota: %w", err)
	}

	// 记录额度变更日志
	quotaLog := &model.QuotaLog{
		UserID:        log.UserID,
		OperationType: "refund",
		Amount:        refundAmount,
		Reason:        "Charge refund for billing log " + fmt.Sprintf("%d", billingLogID),
		BillingLogID:  &billingLogID,
	}

	// 获取当前用户额度
	user, err := s.userRepo.FindByID(ctx, log.UserID)
	if err != nil {
		return err
	}
	if user != nil {
		quotaLog.BalanceAfter = user.Quota
		quotaLog.BalanceBefore = user.Quota - refundAmount
	}

	if err := s.quotaLogRepo.Create(ctx, quotaLog); err != nil {
		return fmt.Errorf("failed to create quota log: %w", err)
	}

	// 更新计费日志状态为已退款
	return s.billingRepo.UpdateStatus(ctx, billingLogID, 3)
}

