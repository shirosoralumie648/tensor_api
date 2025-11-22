package quota

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// DefaultQuotaService 默认配额服务实现
type DefaultQuotaService struct {
	db         *gorm.DB
	cache      QuotaCache
	calculator QuotaCalculator
}

// NewDefaultQuotaService 创建默认配额服务
func NewDefaultQuotaService(db *gorm.DB, cache QuotaCache, calculator QuotaCalculator) *DefaultQuotaService {
	return &DefaultQuotaService{
		db:         db,
		cache:      cache,
		calculator: calculator,
	}
}

// PreConsumeQuota 预扣费
func (s *DefaultQuotaService) PreConsumeQuota(req *PreConsumeRequest) (*PreConsumeResponse, error) {
	// 1. 获取用户余额
	balance, err := s.GetUserBalance(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	// 2. 检查余额是否充足
	if balance < req.EstimatedQuota {
		return nil, fmt.Errorf("insufficient balance: have %.2f, need %.2f", balance, req.EstimatedQuota)
	}

	// 3. 信任用户优化：如果余额足够（超过阈值），不实际预扣费
	if req.TrustThreshold > 0 && balance >= req.TrustThreshold {
		// 只记录预扣费记录，不实际扣费
		record := &PreConsumedRecord{
			RequestID:    req.RequestID,
			UserID:       req.UserID,
			Quota:        0, // 未实际扣费
			PromptTokens: req.PromptTokens,
			MaxTokens:    req.MaxTokens,
			Model:        req.Model,
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		}

		if err := s.cache.SetPreConsumed(record); err != nil {
			return nil, fmt.Errorf("failed to cache pre-consumed record: %w", err)
		}

		return &PreConsumeResponse{
			PreConsumed:      false,
			PreConsumedQuota: 0,
			RemainingBalance: balance,
		}, nil
	}

	// 4. 实际预扣费（余额不足或未启用信任）
	if err := s.deductQuota(req.UserID, req.EstimatedQuota); err != nil {
		return nil, fmt.Errorf("failed to deduct quota: %w", err)
	}

	// 5. 记录预扣费
	record := &PreConsumedRecord{
		RequestID:    req.RequestID,
		UserID:       req.UserID,
		Quota:        req.EstimatedQuota,
		PromptTokens: req.PromptTokens,
		MaxTokens:    req.MaxTokens,
		Model:        req.Model,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(15 * time.Minute),
	}

	if err := s.cache.SetPreConsumed(record); err != nil {
		// 缓存失败，回滚预扣费
		_ = s.refundQuota(req.UserID, req.EstimatedQuota)
		return nil, fmt.Errorf("failed to cache pre-consumed record: %w", err)
	}

	return &PreConsumeResponse{
		PreConsumed:      true,
		PreConsumedQuota: req.EstimatedQuota,
		RemainingBalance: balance - req.EstimatedQuota,
	}, nil
}

// ReturnPreConsumedQuota 归还预扣费（请求失败时调用）
func (s *DefaultQuotaService) ReturnPreConsumedQuota(requestID string, userID int) error {
	// 1. 获取预扣费记录
	record, err := s.cache.GetPreConsumed(requestID)
	if err != nil {
		// 记录不存在可能是已经处理过了，不报错
		return nil
	}

	// 2. 如果实际预扣了费用，退还
	if record.Quota > 0 {
		if err := s.refundQuota(userID, record.Quota); err != nil {
			return fmt.Errorf("failed to refund quota: %w", err)
		}
	}

	// 3. 删除预扣费记录
	if err := s.cache.DeletePreConsumed(requestID); err != nil {
		return fmt.Errorf("failed to delete pre-consumed record: %w", err)
	}

	return nil
}

// PostConsumeQuota 后扣费（实际消费调整）
func (s *DefaultQuotaService) PostConsumeQuota(req *PostConsumeRequest) error {
	// 1. 获取预扣费记录
	record, err := s.cache.GetPreConsumed(req.RequestID)
	if err != nil {
		// 没有预扣费记录，直接扣费
		if err := s.deductQuota(req.UserID, req.ActualQuota); err != nil {
			return fmt.Errorf("failed to deduct quota: %w", err)
		}
	} else {
		// 2. 计算差额
		diff := req.ActualQuota - record.Quota

		if diff > 0 {
			// 实际消费 > 预扣，补扣差额
			if err := s.deductQuota(req.UserID, diff); err != nil {
				return fmt.Errorf("failed to deduct additional quota: %w", err)
			}
		} else if diff < 0 {
			// 实际消费 < 预扣，退还多余
			if err := s.refundQuota(req.UserID, -diff); err != nil {
				return fmt.Errorf("failed to refund excess quota: %w", err)
			}
		}

		// 3. 删除预扣费记录
		_ = s.cache.DeletePreConsumed(req.RequestID)
	}

	// 4. 记录消费日志
	log := &model.UnifiedLog{
		UserID:           req.UserID,
		ChannelID:        req.ChannelID,
		LogType:          1, // 1:消费
		ModelName:        req.Model,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		Quota:            int(req.ActualQuota),
		IsStream:         req.IsStream,
		UseTime:          int(req.ResponseTime),
		CreatedAt:        time.Now(),
	}

	if err := s.db.Create(log).Error; err != nil {
		// 日志记录失败不影响主流程
		fmt.Printf("failed to create consume log: %v\n", err)
	}

	// 5. 失效用户余额缓存
	_ = s.cache.InvalidateUserBalance(req.UserID)

	return nil
}

// RefundQuota 退款
func (s *DefaultQuotaService) RefundQuota(req *RefundRequest) error {
	if err := s.refundQuota(req.UserID, req.Quota); err != nil {
		return fmt.Errorf("failed to refund quota: %w", err)
	}

	// 记录退款日志
	log := &model.UnifiedLog{
		UserID:    req.UserID,
		LogType:   2, // 2:退款
		Quota:     int(req.Quota),
		CreatedAt: time.Now(),
	}

	_ = s.db.Create(log)
	_ = s.cache.InvalidateUserBalance(req.UserID)

	return nil
}

// GetUserBalance 获取用户余额
func (s *DefaultQuotaService) GetUserBalance(userID int) (float64, error) {
	// 1. 尝试从缓存获取
	if balance, exists, err := s.cache.GetUserBalance(userID); err == nil && exists {
		return balance, nil
	}

	// 2. 从数据库获取
	var user model.User
	if err := s.db.Select("quota").First(&user, userID).Error; err != nil {
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	// 3. 更新缓存
	_ = s.cache.SetUserBalance(userID, float64(user.Quota))

	return float64(user.Quota), nil
}

// GetPreConsumedRecord 获取预扣费记录
func (s *DefaultQuotaService) GetPreConsumedRecord(requestID string) (*PreConsumedRecord, error) {
	return s.cache.GetPreConsumed(requestID)
}

// deductQuota 扣除配额（内部方法）
func (s *DefaultQuotaService) deductQuota(userID int, quota float64) error {
	result := s.db.Model(&model.User{}).
		Where("id = ? AND quota >= ?", userID, int(quota)).
		Update("quota", gorm.Expr("quota - ?", int(quota)))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("insufficient quota or user not found")
	}

	return nil
}

// refundQuota 退还配额（内部方法）
func (s *DefaultQuotaService) refundQuota(userID int, quota float64) error {
	return s.db.Model(&model.User{}).
		Where("id = ?", userID).
		Update("quota", gorm.Expr("quota + ?", int(quota))).Error
}
