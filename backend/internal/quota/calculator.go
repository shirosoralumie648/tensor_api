package quota

import (
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// DefaultQuotaCalculator 默认配额计算器
type DefaultQuotaCalculator struct {
	db         *gorm.DB
	priceCache map[string]*model.ModelPricing
	mu         sync.RWMutex
	group      string // 用户分组（默认default）
}

// NewDefaultQuotaCalculator 创建默认配额计算器
func NewDefaultQuotaCalculator(db *gorm.DB) *DefaultQuotaCalculator {
	calc := &DefaultQuotaCalculator{
		db:         db,
		priceCache: make(map[string]*model.ModelPricing),
		group:      "default",
	}

	// 初始加载价格
	calc.refreshPriceCache()

	return calc
}

// CalculateQuota 计算配额（返回积分，内部使用整数计算避免浮点精度问题）
func (c *DefaultQuotaCalculator) CalculateQuota(modelName string, promptTokens, completionTokens int) (float64, error) {
	pricing, err := c.getModelPricing(modelName)
	if err != nil {
		return 0, err
	}

	// 使用ModelPricing内置的计算方法
	quota := pricing.CalculateQuota(promptTokens, completionTokens)
	return float64(quota), nil
}

// EstimateMaxQuota 估算最大配额（用于预扣费）
func (c *DefaultQuotaCalculator) EstimateMaxQuota(modelName string, promptTokens, maxTokens int) (float64, error) {
	// 保守估算：假设生成满 maxTokens
	return c.CalculateQuota(modelName, promptTokens, maxTokens)
}

// getModelPricing 获取模型定价
func (c *DefaultQuotaCalculator) getModelPricing(modelName string) (*model.ModelPricing, error) {
	// 构建缓存key: model_group
	cacheKey := fmt.Sprintf("%s_%s", modelName, c.group)

	c.mu.RLock()
	pricing, exists := c.priceCache[cacheKey]
	c.mu.RUnlock()

	if exists {
		return pricing, nil
	}

	// 从数据库加载
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if pricing, exists := c.priceCache[cacheKey]; exists {
		return pricing, nil
	}

	var dbPricing model.ModelPricing
	if err := c.db.Where("model = ? AND \"group\" = ? AND enabled = ? AND deleted_at IS NULL",
		modelName, c.group, true).
		First(&dbPricing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 使用默认定价
			defaultPricing := c.getDefaultPricing(modelName)
			c.priceCache[cacheKey] = defaultPricing
			return defaultPricing, nil
		}
		return nil, fmt.Errorf("failed to load model pricing: %w", err)
	}

	c.priceCache[cacheKey] = &dbPricing
	return &dbPricing, nil
}

// getDefaultPricing 获取默认定价（当数据库中没有时）
func (c *DefaultQuotaCalculator) getDefaultPricing(modelName string) *model.ModelPricing {
	// 使用按量计费的默认倍率
	// 倍率说明：基于每1000 tokens的积分消耗

	ratio := 15.0          // 默认倍率
	completionRatio := 2.0 // 默认completion倍率

	// 根据模型设置不同倍率
	switch modelName {
	case "gpt-4o":
		ratio = 15.0
		completionRatio = 2.0
	case "gpt-4o-mini":
		ratio = 0.6
		completionRatio = 2.0
	case "gpt-4-turbo", "gpt-4":
		ratio = 30.0
		completionRatio = 2.0
	case "gpt-3.5-turbo":
		ratio = 1.5
		completionRatio = 2.0
	case "claude-3.5-sonnet":
		ratio = 15.0
		completionRatio = 5.0
	case "claude-3-opus":
		ratio = 75.0
		completionRatio = 5.0
	case "gemini-1.5-pro":
		ratio = 7.0
		completionRatio = 2.0
	case "gemini-1.5-flash", "gemini-2.0-flash":
		ratio = 0.35
		completionRatio = 2.0
	}

	return &model.ModelPricing{
		Model:           modelName,
		Group:           c.group,
		QuotaType:       model.QuotaTypeByToken,
		ModelRatio:      &ratio,
		CompletionRatio: completionRatio,
		GroupRatio:      1.0,
		Enabled:         true,
	}
}

// refreshPriceCache 刷新价格缓存
func (c *DefaultQuotaCalculator) refreshPriceCache() error {
	ctx := context.Background()

	var pricings []model.ModelPricing
	if err := c.db.WithContext(ctx).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&pricings).Error; err != nil {
		return fmt.Errorf("failed to refresh price cache: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空并重建缓存
	c.priceCache = make(map[string]*model.ModelPricing)
	for i := range pricings {
		cacheKey := fmt.Sprintf("%s_%s", pricings[i].Model, pricings[i].Group)
		c.priceCache[cacheKey] = &pricings[i]
	}

	return nil
}

// RefreshCache 手动刷新缓存
func (c *DefaultQuotaCalculator) RefreshCache() error {
	return c.refreshPriceCache()
}

// SetGroup 设置用户分组（用于差异化定价）
func (c *DefaultQuotaCalculator) SetGroup(group string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.group = group
}
