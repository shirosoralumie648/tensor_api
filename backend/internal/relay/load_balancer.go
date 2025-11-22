package relay

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ChannelSelectOptions 渠道选择选项
type ChannelSelectOptions struct {
	ChannelType     string
	Model           string
	UserGroup       string
	ExcludeIDs      []int
	Region          string
	MinAvailability float64
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy int

const (
	// 加权轮询
	LBStrategyWeightedRoundRobin LoadBalanceStrategy = iota
	// 随机
	LBStrategyRandom
	// 最少连接
	LBStrategyLeastConnection
	// 最低延迟
	LBStrategyLowestLatency
	// 一致性哈希
	LBStrategyConsistentHash
	// 响应时间加权
	LBStrategyWeightedByLatency
)

// String 返回策略的字符串表示
func (s LoadBalanceStrategy) String() string {
	switch s {
	case LBStrategyWeightedRoundRobin:
		return "weighted_round_robin"
	case LBStrategyRandom:
		return "random"
	case LBStrategyLeastConnection:
		return "least_connection"
	case LBStrategyLowestLatency:
		return "lowest_latency"
	case LBStrategyConsistentHash:
		return "consistent_hash"
	case LBStrategyWeightedByLatency:
		return "weighted_by_latency"
	default:
		return "unknown"
	}
}

// LoadBalancerConfig 负载均衡器配置
type LoadBalancerConfig struct {
	// 策略
	Strategy LoadBalanceStrategy

	// 启用健康检查
	EnableHealthCheck bool

	// 健康检查配置
	HealthCheckConfig *HealthCheckConfig

	// 启用断路器
	EnableCircuitBreaker bool

	// 断路器失败阈值
	CircuitBreakerFailureThreshold int64

	// 断路器成功阈值
	CircuitBreakerSuccessThreshold int64

	// 断路器超时
	CircuitBreakerTimeout time.Duration

	// 重试次数
	MaxRetries int

	// 重试间隔
	RetryInterval time.Duration

	// 是否启用权重自适应
	EnableAdaptiveWeight bool

	// 权重调整间隔
	WeightAdjustInterval time.Duration
}

// DefaultLoadBalancerConfig 默认配置
func DefaultLoadBalancerConfig() *LoadBalancerConfig {
	return &LoadBalancerConfig{
		Strategy:                       LBStrategyWeightedRoundRobin,
		EnableHealthCheck:              true,
		HealthCheckConfig:              DefaultHealthCheckConfig(),
		EnableCircuitBreaker:           true,
		CircuitBreakerFailureThreshold: 5,
		CircuitBreakerSuccessThreshold: 3,
		CircuitBreakerTimeout:          1 * time.Minute,
		MaxRetries:                     3,
		RetryInterval:                  100 * time.Millisecond,
		EnableAdaptiveWeight:           true,
		WeightAdjustInterval:           5 * time.Minute,
	}
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	// 缓存
	cache *ChannelCache

	// 配置
	config *LoadBalancerConfig

	// 健康检查器
	healthChecker *HealthChecker

	// 断路器映射
	circuitBreakers map[string]*CircuitBreaker
	breakersMu      sync.RWMutex

	// 轮询计数器
	roundRobinCounter int64

	// 权重调整定时器
	weightAdjustTicker *time.Ticker
	weightAdjustStopCh chan struct{}

	// 统计信息
	totalRequests int64
	successCount  int64
	failureCount  int64
	statsMu       sync.RWMutex

	// 停止信号
	stopCh chan struct{}

	// 等待组
	wg sync.WaitGroup

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(cache *ChannelCache, config *LoadBalancerConfig) *LoadBalancer {
	if config == nil {
		config = DefaultLoadBalancerConfig()
	}

	lb := &LoadBalancer{
		cache:              cache,
		config:             config,
		circuitBreakers:    make(map[string]*CircuitBreaker),
		roundRobinCounter:  0,
		weightAdjustStopCh: make(chan struct{}),
		logFunc:            defaultLogFunc,
		stopCh:             make(chan struct{}),
	}

	// 创建健康检查器
	if config.EnableHealthCheck {
		lb.healthChecker = NewHealthChecker(cache, config.HealthCheckConfig)
	}

	return lb
}

// Start 启动负载均衡器
func (lb *LoadBalancer) Start() {
	if lb.config.EnableHealthCheck && lb.healthChecker != nil {
		lb.healthChecker.Start()
	}

	if lb.config.EnableAdaptiveWeight {
		lb.wg.Add(1)
		go func() {
			defer lb.wg.Done()
			lb.runWeightAdjustment()
		}()
	}

	lb.logFunc("info", "Load balancer started")
}

// Stop 停止负载均衡器
func (lb *LoadBalancer) Stop() {
	if lb.config.EnableHealthCheck && lb.healthChecker != nil {
		lb.healthChecker.Stop()
	}

	close(lb.stopCh)
	close(lb.weightAdjustStopCh)
	lb.wg.Wait()

	lb.logFunc("info", "Load balancer stopped")
}

// SelectChannel 选择一个渠道
func (lb *LoadBalancer) SelectChannel(options *ChannelSelectOptions) (*Channel, error) {
	// 获取所有可用渠道
	candidates := lb.getAvailableChannels(options)
	if len(candidates) == 0 {
		atomic.AddInt64(&lb.failureCount, 1)
		return nil, fmt.Errorf("no available channels")
	}

	// 根据策略选择
	var selected *Channel
	switch lb.config.Strategy {
	case LBStrategyWeightedRoundRobin:
		selected = lb.selectWeightedRoundRobin(candidates)

	case LBStrategyRandom:
		selected = lb.selectRandom(candidates)

	case LBStrategyLeastConnection:
		selected = lb.selectLeastConnection(candidates)

	case LBStrategyLowestLatency:
		selected = lb.selectLowestLatency(candidates)

	case LBStrategyWeightedByLatency:
		selected = lb.selectWeightedByLatency(candidates)

	case LBStrategyConsistentHash:
		selected = lb.selectConsistentHash(candidates, options.Model)

	default:
		return nil, fmt.Errorf("unknown strategy: %d", lb.config.Strategy)
	}

	if selected == nil {
		atomic.AddInt64(&lb.failureCount, 1)
		return nil, fmt.Errorf("selection failed")
	}

	atomic.AddInt64(&lb.totalRequests, 1)
	atomic.AddInt64(&lb.successCount, 1)

	return selected, nil
}

// getAvailableChannels 获取可用渠道
func (lb *LoadBalancer) getAvailableChannels(options *ChannelSelectOptions) []*Channel {
	filter := &ChannelFilter{
		Type:            options.ChannelType,
		Model:           options.Model,
		Region:          options.Region,
		MinAvailability: options.MinAvailability,
		OnlyEnabled:     true,
	}

	candidates := lb.cache.FilterChannels(filter)

	// 过滤掉被断路器标记为不可用的渠道
	if lb.config.EnableCircuitBreaker {
		filtered := make([]*Channel, 0)
		for _, ch := range candidates {
			if lb.isCircuitBreakerAvailable(ch.ID) {
				filtered = append(filtered, ch)
			}
		}
		return filtered
	}

	return candidates
}

// selectWeightedRoundRobin 加权轮询选择
func (lb *LoadBalancer) selectWeightedRoundRobin(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0
	for _, ch := range candidates {
		totalWeight += ch.Weight
	}

	if totalWeight == 0 {
		return lb.selectRandom(candidates)
	}

	// 根据权重随机选择
	target := randInt(totalWeight)
	current := 0

	for _, ch := range candidates {
		current += ch.Weight
		if target < current {
			return ch
		}
	}

	return candidates[len(candidates)-1]
}

// selectRandom 随机选择
func (lb *LoadBalancer) selectRandom(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}
	idx := randInt(len(candidates))
	return candidates[idx]
}

// selectLeastConnection 最少连接选择
func (lb *LoadBalancer) selectLeastConnection(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		iConcurrency := atomic.LoadInt64(&candidates[i].Metrics.CurrentConcurrency)
		jConcurrency := atomic.LoadInt64(&candidates[j].Metrics.CurrentConcurrency)
		return iConcurrency < jConcurrency
	})

	return candidates[0]
}

// selectLowestLatency 最低延迟选择
func (lb *LoadBalancer) selectLowestLatency(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Metrics.AvgLatency < candidates[j].Metrics.AvgLatency
	})

	return candidates[0]
}

// selectWeightedByLatency 按延迟加权选择
func (lb *LoadBalancer) selectWeightedByLatency(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	// 计算每个渠道的得分（反向延迟 * 权重）
	type channelScore struct {
		channel *Channel
		score   float64
	}

	scores := make([]channelScore, 0)
	maxLatency := 0.0
	totalScore := 0.0

	// 找到最大延迟
	for _, ch := range candidates {
		if ch.Metrics.AvgLatency > maxLatency {
			maxLatency = ch.Metrics.AvgLatency
		}
	}

	if maxLatency == 0 {
		maxLatency = 1
	}

	// 计算得分
	for _, ch := range candidates {
		// 反向延迟（低延迟 = 高分）
		normalizedLatency := 1 - (ch.Metrics.AvgLatency / maxLatency)
		score := normalizedLatency * float64(ch.Weight)
		scores = append(scores, channelScore{ch, score})
		totalScore += score
	}

	if totalScore == 0 {
		return lb.selectRandom(candidates)
	}

	// 按得分比例选择
	target := randFloat() * totalScore
	current := 0.0

	for _, cs := range scores {
		current += cs.score
		if target < current {
			return cs.channel
		}
	}

	return candidates[len(candidates)-1]
}

// selectConsistentHash 一致性哈希选择
func (lb *LoadBalancer) selectConsistentHash(candidates []*Channel, key string) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	if key == "" {
		return lb.selectRandom(candidates)
	}

	// 简单的哈希选择
	hash := hashString(key)
	idx := int(hash) % len(candidates)
	return candidates[idx]
}

// RecordRequest 记录请求
func (lb *LoadBalancer) RecordRequest(channelID string, success bool, latency int64) error {
	ch, err := lb.cache.GetChannel(channelID)
	if err != nil {
		return err
	}

	if success {
		ch.RecordSuccess(latency)
		if lb.config.EnableCircuitBreaker {
			lb.recordCircuitBreakerSuccess(channelID)
		}
	} else {
		ch.RecordFailure()
		if lb.config.EnableCircuitBreaker {
			lb.recordCircuitBreakerFailure(channelID)
		}
	}

	return nil
}

// recordCircuitBreakerSuccess 记录断路器成功
func (lb *LoadBalancer) recordCircuitBreakerSuccess(channelID string) {
	breaker := lb.getOrCreateCircuitBreaker(channelID)
	breaker.RecordSuccess()
}

// recordCircuitBreakerFailure 记录断路器失败
func (lb *LoadBalancer) recordCircuitBreakerFailure(channelID string) {
	breaker := lb.getOrCreateCircuitBreaker(channelID)
	breaker.RecordFailure()
}

// getOrCreateCircuitBreaker 获取或创建断路器
func (lb *LoadBalancer) getOrCreateCircuitBreaker(channelID string) *CircuitBreaker {
	lb.breakersMu.RLock()
	if breaker, ok := lb.circuitBreakers[channelID]; ok {
		lb.breakersMu.RUnlock()
		return breaker
	}
	lb.breakersMu.RUnlock()

	lb.breakersMu.Lock()
	defer lb.breakersMu.Unlock()

	if breaker, ok := lb.circuitBreakers[channelID]; ok {
		return breaker
	}

	breaker := NewCircuitBreaker(
		channelID,
		lb.config.CircuitBreakerFailureThreshold,
		lb.config.CircuitBreakerSuccessThreshold,
		lb.config.CircuitBreakerTimeout,
	)

	lb.circuitBreakers[channelID] = breaker
	return breaker
}

// isCircuitBreakerAvailable 断路器是否可用
func (lb *LoadBalancer) isCircuitBreakerAvailable(channelID string) bool {
	lb.breakersMu.RLock()
	breaker, ok := lb.circuitBreakers[channelID]
	lb.breakersMu.RUnlock()

	if !ok {
		return true
	}

	return breaker.IsAvailable()
}

// SetStrategy 设置策略
func (lb *LoadBalancer) SetStrategy(strategy LoadBalanceStrategy) {
	lb.config.Strategy = strategy
	lb.logFunc("info", fmt.Sprintf("Load balance strategy changed to %s", strategy.String()))
}

// GetStatistics 获取统计信息
func (lb *LoadBalancer) GetStatistics() map[string]interface{} {
	lb.statsMu.RLock()
	defer lb.statsMu.RUnlock()

	total := atomic.LoadInt64(&lb.totalRequests)
	success := atomic.LoadInt64(&lb.successCount)
	failure := atomic.LoadInt64(&lb.failureCount)

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	return map[string]interface{}{
		"strategy":         lb.config.Strategy.String(),
		"total_requests":   total,
		"success_count":    success,
		"failure_count":    failure,
		"success_rate":     successRate,
		"circuit_breakers": len(lb.circuitBreakers),
		"health_check":     lb.config.EnableHealthCheck,
	}
}

// runWeightAdjustment 运行权重调整
func (lb *LoadBalancer) runWeightAdjustment() {
	ticker := time.NewTicker(lb.config.WeightAdjustInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lb.stopCh:
			return
		case <-lb.weightAdjustStopCh:
			return
		case <-ticker.C:
			lb.adjustWeights()
		}
	}
}

// adjustWeights 调整权重
func (lb *LoadBalancer) adjustWeights() {
	channels := lb.cache.GetAllChannels()

	for _, ch := range channels {
		successRate := ch.Metrics.GetSuccessRate()

		// 根据成功率调整权重
		baseWeight := ch.Weight

		if successRate >= 95 {
			// 优秀，增加权重
			ch.Weight = int(math.Ceil(float64(baseWeight) * 1.1))
		} else if successRate >= 80 {
			// 良好，保持
			ch.Weight = baseWeight
		} else if successRate >= 50 {
			// 一般，降低权重
			ch.Weight = int(math.Floor(float64(baseWeight) * 0.9))
		} else {
			// 差，大幅降低
			ch.Weight = int(math.Max(1, float64(baseWeight)*0.5))
		}
	}

	lb.logFunc("info", "Weight adjustment completed")
}

// LoadBalancerManager 负载均衡器管理器
type LoadBalancerManager struct {
	// 负载均衡器映射
	balancers   map[string]*LoadBalancer
	balancersMu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewLoadBalancerManager 创建管理器
func NewLoadBalancerManager() *LoadBalancerManager {
	return &LoadBalancerManager{
		balancers: make(map[string]*LoadBalancer),
		logFunc:   defaultLogFunc,
	}
}

// RegisterBalancer 注册负载均衡器
func (lbm *LoadBalancerManager) RegisterBalancer(name string, balancer *LoadBalancer) {
	lbm.balancersMu.Lock()
	defer lbm.balancersMu.Unlock()

	lbm.balancers[name] = balancer
}

// StartAll 启动所有负载均衡器
func (lbm *LoadBalancerManager) StartAll() {
	lbm.balancersMu.RLock()
	defer lbm.balancersMu.RUnlock()

	for name, balancer := range lbm.balancers {
		balancer.Start()
		lbm.logFunc("info", fmt.Sprintf("Started load balancer: %s", name))
	}
}

// StopAll 停止所有负载均衡器
func (lbm *LoadBalancerManager) StopAll() {
	lbm.balancersMu.RLock()
	defer lbm.balancersMu.RUnlock()

	for name, balancer := range lbm.balancers {
		balancer.Stop()
		lbm.logFunc("info", fmt.Sprintf("Stopped load balancer: %s", name))
	}
}

// GetBalancer 获取负载均衡器
func (lbm *LoadBalancerManager) GetBalancer(name string) *LoadBalancer {
	lbm.balancersMu.RLock()
	defer lbm.balancersMu.RUnlock()

	return lbm.balancers[name]
}

// GetAllStatistics 获取所有统计信息
func (lbm *LoadBalancerManager) GetAllStatistics() map[string]map[string]interface{} {
	lbm.balancersMu.RLock()
	defer lbm.balancersMu.RUnlock()

	result := make(map[string]map[string]interface{})
	for name, balancer := range lbm.balancers {
		result[name] = balancer.GetStatistics()
	}

	return result
}

// hashString 计算字符串哈希值
func hashString(s string) uint64 {
	hash := uint64(5381)
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}
