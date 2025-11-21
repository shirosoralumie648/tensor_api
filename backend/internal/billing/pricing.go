package billing

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// PricingType 定价类型
type PricingType int

const (
	// 按 token 数计费
	PricingByToken PricingType = iota
	// 按次数计费
	PricingByRequest
	// 按时间计费
	PricingByTime
)

// String 返回定价类型的字符串表示
func (pt PricingType) String() string {
	switch pt {
	case PricingByToken:
		return "by_token"
	case PricingByRequest:
		return "by_request"
	case PricingByTime:
		return "by_time"
	default:
		return "unknown"
	}
}

// ModelPrice 模型价格
type ModelPrice struct {
	// 模型名称
	ModelName string

	// 输入价格（美元/1K tokens 或 美元/次）
	InputPrice float64

	// 输出价格（美元/1K tokens 或 美元/次）
	OutputPrice float64

	// 定价类型
	PricingType PricingType

	// 最小费用
	MinPrice float64

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time

	// 版本号
	Version int

	// 备注
	Remark string
}

// PriceGroup 价格组（分组定价）
type PriceGroup struct {
	// 组 ID
	GroupID string

	// 组名称
	GroupName string

	// 组内模型列表
	Models []string

	// 倍率（相对于基础价格）
	Multiplier float64

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time

	// 备注
	Remark string
}

// PricingStrategy 定价策略
type PricingStrategy struct {
	// 策略 ID
	StrategyID string

	// 策略名称
	StrategyName string

	// 适用的模型列表
	ApplicableModels []string

	// 基础价格组
	BasePrices map[string]*ModelPrice

	// 分组列表
	Groups []*PriceGroup

	// 有效期开始
	EffectiveFrom time.Time

	// 有效期结束
	EffectiveTo *time.Time

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time

	// 备注
	Remark string

	// 互斥锁
	mu sync.RWMutex
}

// PricingManager 定价管理器
type PricingManager struct {
	// 模型价格映射
	modelPrices map[string]*ModelPrice
	pricesMu    sync.RWMutex

	// 价格组映射
	priceGroups map[string]*PriceGroup
	groupsMu    sync.RWMutex

	// 定价策略映射
	strategies map[string]*PricingStrategy
	strategiesMu sync.RWMutex

	// 活跃策略
	activeStrategy *PricingStrategy
	activeMu       sync.RWMutex

	// 定价历史（用于跟踪价格变化）
	priceHistory map[string][]*PriceHistoryRecord
	historyMu    sync.RWMutex

	// 统计信息
	updateCount int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// PriceHistoryRecord 价格历史记录
type PriceHistoryRecord struct {
	// 模型名称
	ModelName string

	// 旧价格
	OldPrice *ModelPrice

	// 新价格
	NewPrice *ModelPrice

	// 变更时间
	ChangedAt time.Time

	// 变更原因
	Reason string

	// 版本号
	Version int
}

// NewPricingManager 创建定价管理器
func NewPricingManager() *PricingManager {
	return &PricingManager{
		modelPrices:  make(map[string]*ModelPrice),
		priceGroups:  make(map[string]*PriceGroup),
		strategies:   make(map[string]*PricingStrategy),
		priceHistory: make(map[string][]*PriceHistoryRecord),
		logFunc:      defaultLogFunc,
	}
}

// RegisterModelPrice 注册模型价格
func (pm *PricingManager) RegisterModelPrice(modelName string, inputPrice, outputPrice float64, pricingType PricingType) error {
	pm.pricesMu.Lock()
	defer pm.pricesMu.Unlock()

	if _, exists := pm.modelPrices[modelName]; exists {
		return fmt.Errorf("model price for %s already registered", modelName)
	}

	pm.modelPrices[modelName] = &ModelPrice{
		ModelName:   modelName,
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		PricingType: pricingType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Version:     1,
	}

	pm.logFunc("info", fmt.Sprintf("Registered price for model %s (input: %.6f, output: %.6f)", modelName, inputPrice, outputPrice))

	return nil
}

// UpdateModelPrice 更新模型价格
func (pm *PricingManager) UpdateModelPrice(modelName string, inputPrice, outputPrice float64, reason string) error {
	pm.pricesMu.Lock()
	price, exists := pm.modelPrices[modelName]
	if !exists {
		pm.pricesMu.Unlock()
		return fmt.Errorf("model price for %s not found", modelName)
	}

	// 保存历史记录
	oldPrice := &ModelPrice{
		ModelName:   price.ModelName,
		InputPrice:  price.InputPrice,
		OutputPrice: price.OutputPrice,
		PricingType: price.PricingType,
		MinPrice:    price.MinPrice,
		Version:     price.Version,
	}

	newPrice := &ModelPrice{
		ModelName:   modelName,
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		PricingType: price.PricingType,
		MinPrice:    price.MinPrice,
		Version:     price.Version + 1,
	}

	// 更新价格
	price.InputPrice = inputPrice
	price.OutputPrice = outputPrice
	price.UpdatedAt = time.Now()
	price.Version++

	pm.pricesMu.Unlock()

	// 记录历史
	pm.historyMu.Lock()
	if pm.priceHistory[modelName] == nil {
		pm.priceHistory[modelName] = make([]*PriceHistoryRecord, 0)
	}
	pm.priceHistory[modelName] = append(pm.priceHistory[modelName], &PriceHistoryRecord{
		ModelName: modelName,
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
		ChangedAt: time.Now(),
		Reason:    reason,
		Version:   newPrice.Version,
	})
	pm.historyMu.Unlock()

	atomic.AddInt64(&pm.updateCount, 1)

	pm.logFunc("info", fmt.Sprintf("Updated price for model %s (input: %.6f -> %.6f, output: %.6f -> %.6f)", 
		modelName, oldPrice.InputPrice, inputPrice, oldPrice.OutputPrice, outputPrice))

	return nil
}

// GetModelPrice 获取模型价格
func (pm *PricingManager) GetModelPrice(modelName string) (*ModelPrice, error) {
	pm.pricesMu.RLock()
	defer pm.pricesMu.RUnlock()

	price, ok := pm.modelPrices[modelName]
	if !ok {
		return nil, fmt.Errorf("model price for %s not found", modelName)
	}

	return price, nil
}

// CreatePriceGroup 创建价格组
func (pm *PricingManager) CreatePriceGroup(groupID, groupName string, models []string, multiplier float64) error {
	pm.groupsMu.Lock()
	defer pm.groupsMu.Unlock()

	if _, exists := pm.priceGroups[groupID]; exists {
		return fmt.Errorf("price group %s already exists", groupID)
	}

	pm.priceGroups[groupID] = &PriceGroup{
		GroupID:    groupID,
		GroupName:  groupName,
		Models:     models,
		Multiplier: multiplier,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	pm.logFunc("info", fmt.Sprintf("Created price group %s (%s) with multiplier %.2f for %d models", groupID, groupName, multiplier, len(models)))

	return nil
}

// GetPriceGroup 获取价格组
func (pm *PricingManager) GetPriceGroup(groupID string) (*PriceGroup, error) {
	pm.groupsMu.RLock()
	defer pm.groupsMu.RUnlock()

	group, ok := pm.priceGroups[groupID]
	if !ok {
		return nil, fmt.Errorf("price group %s not found", groupID)
	}

	return group, nil
}

// CreatePricingStrategy 创建定价策略
func (pm *PricingManager) CreatePricingStrategy(strategyID, strategyName string, applicableModels []string) error {
	pm.strategiesMu.Lock()
	defer pm.strategiesMu.Unlock()

	if _, exists := pm.strategies[strategyID]; exists {
		return fmt.Errorf("pricing strategy %s already exists", strategyID)
	}

	pm.strategies[strategyID] = &PricingStrategy{
		StrategyID:       strategyID,
		StrategyName:     strategyName,
		ApplicableModels: applicableModels,
		BasePrices:       make(map[string]*ModelPrice),
		Groups:           make([]*PriceGroup, 0),
		EffectiveFrom:    time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	pm.logFunc("info", fmt.Sprintf("Created pricing strategy %s (%s) for %d models", strategyID, strategyName, len(applicableModels)))

	return nil
}

// ActivatePricingStrategy 激活定价策略
func (pm *PricingManager) ActivatePricingStrategy(strategyID string) error {
	pm.strategiesMu.RLock()
	strategy, exists := pm.strategies[strategyID]
	pm.strategiesMu.RUnlock()

	if !exists {
		return fmt.Errorf("pricing strategy %s not found", strategyID)
	}

	pm.activeMu.Lock()
	pm.activeStrategy = strategy
	pm.activeMu.Unlock()

	pm.logFunc("info", fmt.Sprintf("Activated pricing strategy %s (%s)", strategyID, strategy.StrategyName))

	return nil
}

// GetActiveStrategy 获取活跃策略
func (pm *PricingManager) GetActiveStrategy() *PricingStrategy {
	pm.activeMu.RLock()
	defer pm.activeMu.RUnlock()

	return pm.activeStrategy
}

// CalculatePrice 计算价格
func (pm *PricingManager) CalculatePrice(modelName string, inputTokens, outputTokens int64) (float64, error) {
	price, err := pm.GetModelPrice(modelName)
	if err != nil {
		return 0, err
	}

	var totalCost float64

	if price.PricingType == PricingByToken {
		// 按 token 计费
		inputCost := float64(inputTokens) / 1000.0 * price.InputPrice
		outputCost := float64(outputTokens) / 1000.0 * price.OutputPrice
		totalCost = inputCost + outputCost
	} else if price.PricingType == PricingByRequest {
		// 按次计费
		totalCost = price.InputPrice
	}

	// 应用最小费用
	if totalCost < price.MinPrice && price.MinPrice > 0 {
		totalCost = price.MinPrice
	}

	return totalCost, nil
}

// CalculatePriceWithGroup 计算带分组倍率的价格
func (pm *PricingManager) CalculatePriceWithGroup(modelName string, groupID string, inputTokens, outputTokens int64) (float64, error) {
	basePrice, err := pm.CalculatePrice(modelName, inputTokens, outputTokens)
	if err != nil {
		return 0, err
	}

	group, err := pm.GetPriceGroup(groupID)
	if err != nil {
		return 0, err
	}

	// 检查模型是否在组内
	isInGroup := false
	for _, m := range group.Models {
		if m == modelName {
			isInGroup = true
			break
		}
	}

	if !isInGroup {
		return 0, fmt.Errorf("model %s not in group %s", modelName, groupID)
	}

	return basePrice * group.Multiplier, nil
}

// GetPriceHistory 获取价格历史
func (pm *PricingManager) GetPriceHistory(modelName string) []*PriceHistoryRecord {
	pm.historyMu.RLock()
	defer pm.historyMu.RUnlock()

	if history, exists := pm.priceHistory[modelName]; exists {
		return history
	}

	return make([]*PriceHistoryRecord, 0)
}

// GetAllModelPrices 获取所有模型价格
func (pm *PricingManager) GetAllModelPrices() map[string]*ModelPrice {
	pm.pricesMu.RLock()
	defer pm.pricesMu.RUnlock()

	// 复制映射以避免外部修改
	result := make(map[string]*ModelPrice)
	for k, v := range pm.modelPrices {
		result[k] = v
	}

	return result
}

// GetAllPriceGroups 获取所有价格组
func (pm *PricingManager) GetAllPriceGroups() map[string]*PriceGroup {
	pm.groupsMu.RLock()
	defer pm.groupsMu.RUnlock()

	// 复制映射以避免外部修改
	result := make(map[string]*PriceGroup)
	for k, v := range pm.priceGroups {
		result[k] = v
	}

	return result
}

// GetStatistics 获取统计信息
func (pm *PricingManager) GetStatistics() map[string]interface{} {
	pm.pricesMu.RLock()
	modelCount := len(pm.modelPrices)
	pm.pricesMu.RUnlock()

	pm.groupsMu.RLock()
	groupCount := len(pm.priceGroups)
	pm.groupsMu.RUnlock()

	pm.strategiesMu.RLock()
	strategyCount := len(pm.strategies)
	pm.strategiesMu.RUnlock()

	pm.historyMu.RLock()
	historyCount := 0
	for _, records := range pm.priceHistory {
		historyCount += len(records)
	}
	pm.historyMu.RUnlock()

	return map[string]interface{}{
		"model_count":    modelCount,
		"group_count":    groupCount,
		"strategy_count": strategyCount,
		"history_count":  historyCount,
		"update_count":   atomic.LoadInt64(&pm.updateCount),
	}
}

// PricingCache 定价缓存（优化查询性能）
type PricingCache struct {
	// 管理器
	manager *PricingManager

	// 缓存
	cache sync.Map // map[string]interface{}

	// 缓存 TTL
	ttl time.Duration

	// 最后更新时间
	lastUpdate map[string]time.Time
	updateMu   sync.RWMutex

	// 缓存命中统计
	hits   int64
	misses int64
}

// NewPricingCache 创建定价缓存
func NewPricingCache(manager *PricingManager, ttl time.Duration) *PricingCache {
	return &PricingCache{
		manager:    manager,
		ttl:        ttl,
		lastUpdate: make(map[string]time.Time),
	}
}

// GetPrice 获取价格（带缓存）
func (pc *PricingCache) GetPrice(modelName string) (*ModelPrice, error) {
	cacheKey := "price:" + modelName

	// 检查缓存
	if cached, ok := pc.cache.Load(cacheKey); ok {
		// 检查缓存是否过期
		pc.updateMu.RLock()
		lastUpdate, exists := pc.lastUpdate[cacheKey]
		pc.updateMu.RUnlock()

		if exists && time.Since(lastUpdate) < pc.ttl {
			atomic.AddInt64(&pc.hits, 1)
			return cached.(*ModelPrice), nil
		}
	}

	// 缓存未命中或已过期，从管理器获取
	price, err := pc.manager.GetModelPrice(modelName)
	if err != nil {
		atomic.AddInt64(&pc.misses, 1)
		return nil, err
	}

	// 更新缓存
	pc.cache.Store(cacheKey, price)
	pc.updateMu.Lock()
	pc.lastUpdate[cacheKey] = time.Now()
	pc.updateMu.Unlock()

	atomic.AddInt64(&pc.misses, 1)

	return price, nil
}

// CalculatePrice 计算价格（带缓存）
func (pc *PricingCache) CalculatePrice(modelName string, inputTokens, outputTokens int64) (float64, error) {
	// 获取价格（使用缓存）
	price, err := pc.GetPrice(modelName)
	if err != nil {
		return 0, err
	}

	var totalCost float64

	if price.PricingType == PricingByToken {
		inputCost := float64(inputTokens) / 1000.0 * price.InputPrice
		outputCost := float64(outputTokens) / 1000.0 * price.OutputPrice
		totalCost = inputCost + outputCost
	} else if price.PricingType == PricingByRequest {
		totalCost = price.InputPrice
	}

	if totalCost < price.MinPrice && price.MinPrice > 0 {
		totalCost = price.MinPrice
	}

	return totalCost, nil
}

// GetCacheHitRate 获取缓存命中率
func (pc *PricingCache) GetCacheHitRate() float64 {
	hits := atomic.LoadInt64(&pc.hits)
	misses := atomic.LoadInt64(&pc.misses)
	total := hits + misses

	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total) * 100.0
}

// ClearCache 清空缓存
func (pc *PricingCache) ClearCache() {
	pc.cache.Range(func(key, value interface{}) bool {
		pc.cache.Delete(key)
		return true
	})

	pc.updateMu.Lock()
	pc.lastUpdate = make(map[string]time.Time)
	pc.updateMu.Unlock()

	atomic.StoreInt64(&pc.hits, 0)
	atomic.StoreInt64(&pc.misses, 0)
}

// defaultLogFunc 默认日志函数
func defaultLogFunc(level, msg string, args ...interface{}) {
	// 这里可以集成实际的日志系统
	fmt.Printf("[%s] %s\n", level, msg)
}

