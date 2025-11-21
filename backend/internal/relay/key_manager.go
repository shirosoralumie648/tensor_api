package relay

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// KeyRotationStrategy 密钥轮询策略
type KeyRotationStrategy int

const (
	// 随机选择
	KeyStrategyRandom KeyRotationStrategy = iota
	// 轮询
	KeyStrategyRoundRobin
	// 故障感知轮询
	KeyStrategyFailureAware
)

// String 返回策略的字符串表示
func (s KeyRotationStrategy) String() string {
	switch s {
	case KeyStrategyRandom:
		return "random"
	case KeyStrategyRoundRobin:
		return "round_robin"
	case KeyStrategyFailureAware:
		return "failure_aware"
	default:
		return "unknown"
	}
}

// APIKey API 密钥
type APIKey struct {
	// 密钥 ID
	ID string

	// API 密钥值
	Key string

	// 密钥类型
	Type string

	// 优先级
	Priority int

	// 权重
	Weight int

	// 是否启用
	Enabled bool

	// 使用次数
	UsageCount int64

	// 成功次数
	SuccessCount int64

	// 失败次数
	FailureCount int64

	// 最后使用时间
	LastUsedAt int64

	// 最后失败时间
	LastFailureAt int64

	// 连续失败计数
	ConsecutiveFailures int64

	// 创建时间
	CreatedAt time.Time

	// 过期时间
	ExpiresAt *time.Time

	// 配额限制
	QuotaLimit *int64

	// 当前使用量
	CurrentUsage int64
}

// NewAPIKey 创建新的 API 密钥
func NewAPIKey(id, key, keyType string) *APIKey {
	return &APIKey{
		ID:        id,
		Key:       key,
		Type:      keyType,
		Priority:  100,
		Weight:    1,
		Enabled:   true,
		CreatedAt: time.Now(),
	}
}

// IsValid 是否有效
func (k *APIKey) IsValid() bool {
	if !k.Enabled {
		return false
	}

	// 检查过期时间
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}

	// 检查配额
	if k.QuotaLimit != nil && atomic.LoadInt64(&k.CurrentUsage) >= *k.QuotaLimit {
		return false
	}

	return true
}

// GetSuccessRate 获取成功率
func (k *APIKey) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&k.UsageCount)
	if total == 0 {
		return 0
	}
	success := atomic.LoadInt64(&k.SuccessCount)
	return float64(success) / float64(total) * 100
}

// RecordUsage 记录使用
func (k *APIKey) RecordUsage(success bool, quota int64) {
	atomic.AddInt64(&k.UsageCount, 1)
	if success {
		atomic.AddInt64(&k.SuccessCount, 1)
		atomic.StoreInt64(&k.ConsecutiveFailures, 0)
	} else {
		atomic.AddInt64(&k.FailureCount, 1)
		atomic.AddInt64(&k.ConsecutiveFailures, 1)
		atomic.StoreInt64(&k.LastFailureAt, time.Now().Unix())
	}
	atomic.StoreInt64(&k.LastUsedAt, time.Now().Unix())
	if quota > 0 {
		atomic.AddInt64(&k.CurrentUsage, quota)
	}
}

// KeyManager 密钥管理器
type KeyManager struct {
	// 策略
	strategy KeyRotationStrategy

	// 密钥池（按渠道分组）
	keys map[string][]*APIKey
	keysMu sync.RWMutex

	// 轮询计数器
	roundRobinCounter int64

	// 统计信息
	totalSelections      int64
	successfulSelections int64
	failedSelections     int64
	statsLock            sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewKeyManager 创建密钥管理器
func NewKeyManager(strategy KeyRotationStrategy) *KeyManager {
	return &KeyManager{
		strategy:           strategy,
		keys:               make(map[string][]*APIKey),
		roundRobinCounter:  0,
		logFunc:            defaultLogFunc,
	}
}

// AddKey 添加密钥
func (km *KeyManager) AddKey(channelType string, key *APIKey) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if key.ID == "" {
		return fmt.Errorf("key ID is required")
	}

	km.keysMu.Lock()
	defer km.keysMu.Unlock()

	km.keys[channelType] = append(km.keys[channelType], key)
	return nil
}

// RemoveKey 移除密钥
func (km *KeyManager) RemoveKey(channelType, keyID string) error {
	km.keysMu.Lock()
	defer km.keysMu.Unlock()

	keys, ok := km.keys[channelType]
	if !ok {
		return fmt.Errorf("no keys for channel type %s", channelType)
	}

	for i, k := range keys {
		if k.ID == keyID {
			km.keys[channelType] = append(keys[:i], keys[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("key %s not found", keyID)
}

// SelectKey 选择一个密钥
func (km *KeyManager) SelectKey(channelType string) (*APIKey, error) {
	km.keysMu.RLock()
	keys, ok := km.keys[channelType]
	km.keysMu.RUnlock()

	if !ok || len(keys) == 0 {
		atomic.AddInt64(&km.failedSelections, 1)
		return nil, fmt.Errorf("no keys available for %s", channelType)
	}

	// 获取可用的密钥
	available := make([]*APIKey, 0)
	for _, k := range keys {
		if k.IsValid() {
			available = append(available, k)
		}
	}

	if len(available) == 0 {
		atomic.AddInt64(&km.failedSelections, 1)
		return nil, fmt.Errorf("no valid keys available for %s", channelType)
	}

	// 根据策略选择
	var selected *APIKey
	switch km.strategy {
	case KeyStrategyRandom:
		selected = km.selectRandom(available)

	case KeyStrategyRoundRobin:
		selected = km.selectRoundRobin(available)

	case KeyStrategyFailureAware:
		selected = km.selectFailureAware(available)

	default:
		return nil, fmt.Errorf("unknown strategy: %d", km.strategy)
	}

	if selected == nil {
		atomic.AddInt64(&km.failedSelections, 1)
		return nil, fmt.Errorf("selection failed for %s", channelType)
	}

	atomic.AddInt64(&km.totalSelections, 1)
	atomic.AddInt64(&km.successfulSelections, 1)

	return selected, nil
}

// selectRandom 随机选择
func (km *KeyManager) selectRandom(keys []*APIKey) *APIKey {
	if len(keys) == 0 {
		return nil
	}
	idx := randInt(len(keys))
	return keys[idx]
}

// selectRoundRobin 轮询选择
func (km *KeyManager) selectRoundRobin(keys []*APIKey) *APIKey {
	if len(keys) == 0 {
		return nil
	}
	counter := atomic.AddInt64(&km.roundRobinCounter, 1)
	idx := int(counter) % len(keys)
	return keys[idx]
}

// selectFailureAware 故障感知轮询
func (km *KeyManager) selectFailureAware(keys []*APIKey) *APIKey {
	if len(keys) == 0 {
		return nil
	}

	// 计算每个密钥的得分（成功率 * 权重）
	type keyScore struct {
		key   *APIKey
		score float64
	}

	scores := make([]keyScore, 0)
	totalWeight := 0

	for _, k := range keys {
		successRate := k.GetSuccessRate()
		// 如果连续失败过多，降权
		consecutiveFailures := atomic.LoadInt64(&k.ConsecutiveFailures)
		if consecutiveFailures > 0 {
			successRate = successRate * 0.5 // 降权 50%
		}

		score := successRate * float64(k.Weight)
		scores = append(scores, keyScore{k, score})
		totalWeight += k.Weight
	}

	if totalWeight == 0 {
		return km.selectRandom(keys)
	}

	// 按得分比例选择
	target := randFloat() * float64(totalWeight)
	current := 0.0

	for _, ks := range scores {
		current += ks.score
		if target < current {
			return ks.key
		}
	}

	return keys[len(keys)-1]
}

// GetStatistics 获取统计信息
func (km *KeyManager) GetStatistics() map[string]interface{} {
	km.statsLock.RLock()
	defer km.statsLock.RUnlock()

	total := atomic.LoadInt64(&km.totalSelections)
	successful := atomic.LoadInt64(&km.successfulSelections)
	failed := atomic.LoadInt64(&km.failedSelections)

	successRate := 0.0
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	return map[string]interface{}{
		"strategy":                km.strategy.String(),
		"total_selections":        total,
		"successful_selections":   successful,
		"failed_selections":       failed,
		"success_rate":            successRate,
	}
}

// SetStrategy 设置策略
func (km *KeyManager) SetStrategy(strategy KeyRotationStrategy) {
	km.strategy = strategy
}

// GetKeys 获取所有密钥
func (km *KeyManager) GetKeys(channelType string) []*APIKey {
	km.keysMu.RLock()
	defer km.keysMu.RUnlock()

	if keys, ok := km.keys[channelType]; ok {
		// 返回副本
		result := make([]*APIKey, len(keys))
		copy(result, keys)
		return result
	}

	return make([]*APIKey, 0)
}

// KeyPool 密钥池（支持多个渠道）
type KeyPool struct {
	// 渠道管理器映射
	managers map[string]*KeyManager
	managersMu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewKeyPool 创建密钥池
func NewKeyPool() *KeyPool {
	return &KeyPool{
		managers: make(map[string]*KeyManager),
		logFunc:  defaultLogFunc,
	}
}

// RegisterChannelType 注册渠道类型
func (kp *KeyPool) RegisterChannelType(channelType string, strategy KeyRotationStrategy) {
	kp.managersMu.Lock()
	defer kp.managersMu.Unlock()

	if _, exists := kp.managers[channelType]; !exists {
		kp.managers[channelType] = NewKeyManager(strategy)
	}
}

// AddKey 添加密钥
func (kp *KeyPool) AddKey(channelType string, key *APIKey) error {
	kp.managersMu.RLock()
	manager, ok := kp.managers[channelType]
	kp.managersMu.RUnlock()

	if !ok {
		return fmt.Errorf("channel type %s not registered", channelType)
	}

	return manager.AddKey(channelType, key)
}

// SelectKey 选择密钥
func (kp *KeyPool) SelectKey(channelType string) (*APIKey, error) {
	kp.managersMu.RLock()
	manager, ok := kp.managers[channelType]
	kp.managersMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("channel type %s not registered", channelType)
	}

	return manager.SelectKey(channelType)
}

// SetStrategy 设置策略
func (kp *KeyPool) SetStrategy(channelType string, strategy KeyRotationStrategy) error {
	kp.managersMu.RLock()
	manager, ok := kp.managers[channelType]
	kp.managersMu.RUnlock()

	if !ok {
		return fmt.Errorf("channel type %s not registered", channelType)
	}

	manager.SetStrategy(strategy)
	return nil
}

// GetAllStatistics 获取所有统计信息
func (kp *KeyPool) GetAllStatistics() map[string]map[string]interface{} {
	kp.managersMu.RLock()
	defer kp.managersMu.RUnlock()

	result := make(map[string]map[string]interface{})
	for channelType, manager := range kp.managers {
		result[channelType] = manager.GetStatistics()
	}

	return result
}

// 辅助函数
func randInt(n int) int {
	return int(atomic.AddInt64(&globalRandSeed, 1)) % n
}

func randFloat() float64 {
	return float64(atomic.AddInt64(&globalRandSeed, 1)) / float64(1<<63 - 1)
}

var globalRandSeed int64 = int64(time.Now().UnixNano())

