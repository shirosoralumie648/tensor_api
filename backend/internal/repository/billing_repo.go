package repository

import (
	"context"
	"errors"

	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/gorm"
)

// BillingLogRepository 计费日志仓储
type BillingLogRepository struct {
	db *gorm.DB
}

func NewBillingLogRepository() *BillingLogRepository {
	return &BillingLogRepository{
		db: database.DB,
	}
}

// Create 创建计费日志
func (r *BillingLogRepository) Create(ctx context.Context, log *model.BillingLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID 根据 ID 查询
func (r *BillingLogRepository) FindByID(ctx context.Context, id int) (*model.BillingLog, error) {
	var log model.BillingLog
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// FindByUserID 查询用户的计费记录
func (r *BillingLogRepository) FindByUserID(ctx context.Context, userID int, limit int, offset int) ([]*model.BillingLog, int64, error) {
	var logs []*model.BillingLog
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID)

	if err := query.Model(&model.BillingLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// UpdateStatus 更新计费状态
func (r *BillingLogRepository) UpdateStatus(ctx context.Context, id int, status int) error {
	return r.db.WithContext(ctx).Model(&model.BillingLog{}).Where("id = ?", id).Update("status", status).Error
}

// QuotaLogRepository 额度变更日志仓储
type QuotaLogRepository struct {
	db *gorm.DB
}

func NewQuotaLogRepository() *QuotaLogRepository {
	return &QuotaLogRepository{
		db: database.DB,
	}
}

// Create 创建额度日志
func (r *QuotaLogRepository) Create(ctx context.Context, log *model.QuotaLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByUserID 查询用户的额度变更记录
func (r *QuotaLogRepository) FindByUserID(ctx context.Context, userID int, limit int, offset int) ([]*model.QuotaLog, int64, error) {
	var logs []*model.QuotaLog
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID)

	if err := query.Model(&model.QuotaLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// InvoiceRepository 发票仓储
type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{
		db: database.DB,
	}
}

// Create 创建发票
func (r *InvoiceRepository) Create(ctx context.Context, invoice *model.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// FindByID 根据 ID 查询发票
func (r *InvoiceRepository) FindByID(ctx context.Context, id int) (*model.Invoice, error) {
	var invoice model.Invoice
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&invoice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// FindByUserID 查询用户的发票
func (r *InvoiceRepository) FindByUserID(ctx context.Context, userID int, limit int, offset int) ([]*model.Invoice, int64, error) {
	var invoices []*model.Invoice
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID)

	if err := query.Model(&model.Invoice{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// UpdateStatus 更新发票状态
func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id int, status int) error {
	return r.db.WithContext(ctx).Model(&model.Invoice{}).Where("id = ?", id).Update("status", status).Error
}

// PricingPlanRepository 定价计划仓储
type PricingPlanRepository struct {
	db *gorm.DB
}

func NewPricingPlanRepository() *PricingPlanRepository {
	return &PricingPlanRepository{
		db: database.DB,
	}
}

// GetActivePlan 获取活跃的定价计划
func (r *PricingPlanRepository) GetActivePlan(ctx context.Context) (*model.PricingPlan, error) {
	var plan model.PricingPlan
	err := r.db.WithContext(ctx).Where("active = ? AND deleted_at IS NULL", true).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

