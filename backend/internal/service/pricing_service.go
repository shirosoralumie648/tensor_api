package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// PricingService 定价服务接口
type PricingService interface {
	// GetPricing 获取模型定价
	GetPricing(ctx context.Context, modelName, group string) (*model.ModelPricing, error)

	// ListPricing 列出所有定价
	ListPricing(ctx context.Context, enabled *bool) ([]*model.ModelPricing, error)

	// CreatePricing 创建定价
	CreatePricing(ctx context.Context, pricing *model.ModelPricing) error

	// UpdatePricing 更新定价
	UpdatePricing(ctx context.Context, id int, pricing *model.ModelPricing) error

	// DeletePricing 删除定价（软删除）
	DeletePricing(ctx context.Context, id int) error

	// CalculateQuota 计算配额
	CalculateQuota(ctx context.Context, modelName, group string, promptTokens, completionTokens int) (int, error)

	// RefreshCache 刷新缓存
	RefreshCache(ctx context.Context) error
}

// DefaultPricingService 默认定价服务实现
type DefaultPricingService struct {
	db          *gorm.DB
	cache       map[string]*model.ModelPricing
	mu          sync.RWMutex
	groupRatios map[string]float64 // 用户分组倍率
}

// NewPricingService 创建定价服务
func NewPricingService(db *gorm.DB) *DefaultPricingService {
	service := &DefaultPricingService{
		db:    db,
		cache: make(map[string]*model.ModelPricing),
		groupRatios: map[string]float64{
			"default": 1.0,
			"vip":     0.8, // VIP用户8折
			"premium": 0.6, // Premium用户6折
			"free":    1.5, // 免费用户1.5倍
		},
	}

	// 初始加载缓存
	service.RefreshCache(context.Background())

	return service
}

// GetPricing 获取模型定价
func (s *DefaultPricingService) GetPricing(ctx context.Context, modelName, group string) (*model.ModelPricing, error) {
	// 构建缓存key
	cacheKey := fmt.Sprintf("%s_%s", modelName, group)

	// 尝试从缓存获取
	s.mu.RLock()
	if pricing, exists := s.cache[cacheKey]; exists {
		s.mu.RUnlock()
		return pricing, nil
	}
	s.mu.RUnlock()

	// 从数据库查询
	var pricing model.ModelPricing
	err := s.db.WithContext(ctx).
		Where("model = ? AND \"group\" = ? AND enabled = ? AND deleted_at IS NULL", modelName, group, true).
		First(&pricing).Error

	if err == gorm.ErrRecordNotFound {
		// 尝试获取default分组的定价
		if group != "default" {
			return s.GetPricing(ctx, modelName, "default")
		}
		return nil, fmt.Errorf("pricing not found for model: %s", modelName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get pricing: %w", err)
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[cacheKey] = &pricing
	s.mu.Unlock()

	return &pricing, nil
}

// ListPricing 列出所有定价
func (s *DefaultPricingService) ListPricing(ctx context.Context, enabled *bool) ([]*model.ModelPricing, error) {
	var pricings []*model.ModelPricing

	query := s.db.WithContext(ctx).Where("deleted_at IS NULL")

	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Order("model ASC, \"group\" ASC").Find(&pricings).Error; err != nil {
		return nil, fmt.Errorf("failed to list pricing: %w", err)
	}

	return pricings, nil
}

// CreatePricing 创建定价
func (s *DefaultPricingService) CreatePricing(ctx context.Context, pricing *model.ModelPricing) error {
	// 检查是否已存在
	var existing model.ModelPricing
	err := s.db.WithContext(ctx).
		Where("model = ? AND \"group\" = ? AND deleted_at IS NULL", pricing.Model, pricing.Group).
		First(&existing).Error

	if err == nil {
		return fmt.Errorf("pricing already exists for model %s and group %s", pricing.Model, pricing.Group)
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing pricing: %w", err)
	}

	// 创建新定价
	if err := s.db.WithContext(ctx).Create(pricing).Error; err != nil {
		return fmt.Errorf("failed to create pricing: %w", err)
	}

	// 刷新缓存
	s.RefreshCache(ctx)

	return nil
}

// UpdatePricing 更新定价
func (s *DefaultPricingService) UpdatePricing(ctx context.Context, id int, pricing *model.ModelPricing) error {
	// 更新时间
	pricing.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).
		Model(&model.ModelPricing{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(pricing).Error; err != nil {
		return fmt.Errorf("failed to update pricing: %w", err)
	}

	// 刷新缓存
	s.RefreshCache(ctx)

	return nil
}

// DeletePricing 删除定价（软删除）
func (s *DefaultPricingService) DeletePricing(ctx context.Context, id int) error {
	now := time.Now()
	if err := s.db.WithContext(ctx).
		Model(&model.ModelPricing{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error; err != nil {
		return fmt.Errorf("failed to delete pricing: %w", err)
	}

	// 刷新缓存
	s.RefreshCache(ctx)

	return nil
}

// CalculateQuota 计算配额
func (s *DefaultPricingService) CalculateQuota(ctx context.Context, modelName, group string, promptTokens, completionTokens int) (int, error) {
	// 1. 获取定价
	pricing, err := s.GetPricing(ctx, modelName, group)
	if err != nil {
		return 0, err
	}

	// 2. 计算基础配额
	baseQuota := pricing.CalculateQuota(promptTokens, completionTokens)

	// 3. 应用分组倍率（如果定价中没有设置）
	if pricing.GroupRatio == 1.0 {
		if ratio, exists := s.groupRatios[group]; exists {
			baseQuota = int(float64(baseQuota) * ratio)
		}
	}

	return baseQuota, nil
}

// RefreshCache 刷新缓存
func (s *DefaultPricingService) RefreshCache(ctx context.Context) error {
	var pricings []model.ModelPricing
	if err := s.db.WithContext(ctx).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&pricings).Error; err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空并重建缓存
	s.cache = make(map[string]*model.ModelPricing)
	for i := range pricings {
		cacheKey := fmt.Sprintf("%s_%s", pricings[i].Model, pricings[i].Group)
		s.cache[cacheKey] = &pricings[i]
	}

	return nil
}

// SetGroupRatio 设置分组倍率
func (s *DefaultPricingService) SetGroupRatio(group string, ratio float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groupRatios[group] = ratio
}

// GetGroupRatio 获取分组倍率
func (s *DefaultPricingService) GetGroupRatio(group string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ratio, exists := s.groupRatios[group]; exists {
		return ratio
	}
	return 1.0
}
