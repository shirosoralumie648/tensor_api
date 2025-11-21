package relay

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	// 检查间隔
	Interval time.Duration

	// 检查超时
	Timeout time.Duration

	// 最大重试次数
	MaxRetries int

	// 健康判定的成功率阈值
	HealthyThreshold float64

	// 降级判定的成功率阈值
	DegradedThreshold float64

	// 连续失败触发不可用的次数
	MaxConsecutiveFailures int

	// 恢复检查间隔（用于不可用的渠道）
	RecoveryInterval time.Duration

	// 健康检查端点
	HealthCheckEndpoint string
}

// DefaultHealthCheckConfig 默认配置
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Interval:                   5 * time.Minute,
		Timeout:                    10 * time.Second,
		MaxRetries:                 3,
		HealthyThreshold:           95.0,
		DegradedThreshold:          50.0,
		MaxConsecutiveFailures:     5,
		RecoveryInterval:           1 * time.Minute,
		HealthCheckEndpoint:        "/health",
	}
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	// 渠道 ID
	ChannelID string

	// 检查时间
	CheckTime time.Time

	// 是否健康
	Healthy bool

	// 成功率
	SuccessRate float64

	// 延迟（毫秒）
	Latency int64

	// 错误信息
	Error string

	// 状态
	Status ChannelStatus
}

// HealthChecker 健康检查器
type HealthChecker struct {
	// 缓存
	cache *ChannelCache

	// 配置
	config *HealthCheckConfig

	// HTTP 客户端
	httpClient *http.Client

	// 检查状态（按渠道 ID）
	checkStates map[string]*channelCheckState
	statesMu    sync.RWMutex

	// 停止信号
	stopCh chan struct{}

	// 等待组
	wg sync.WaitGroup

	// 统计信息
	totalChecks  int64
	successCount int64
	failureCount int64
	statsMu      sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})

	// 结果回调
	resultCallback func(*HealthCheckResult)
}

// channelCheckState 渠道检查状态
type channelCheckState struct {
	// 上次检查时间
	lastCheckTime time.Time

	// 连续失败计数
	consecutiveFailures int64

	// 是否处于恢复模式
	inRecovery bool

	// 上次状态
	lastStatus ChannelStatus

	// 互斥锁
	mu sync.RWMutex
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(cache *ChannelCache, config *HealthCheckConfig) *HealthChecker {
	if config == nil {
		config = DefaultHealthCheckConfig()
	}

	return &HealthChecker{
		cache:          cache,
		config:         config,
		httpClient:     &http.Client{Timeout: config.Timeout},
		checkStates:    make(map[string]*channelCheckState),
		stopCh:         make(chan struct{}),
		logFunc:        defaultLogFunc,
		resultCallback: nil,
	}
}

// Start 启动健康检查
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		hc.run()
	}()
	hc.logFunc("info", "Health checker started")
}

// Stop 停止健康检查
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
	hc.wg.Wait()
	hc.logFunc("info", "Health checker stopped")
}

// run 运行检查循环
func (hc *HealthChecker) run() {
	// 立即执行第一次检查
	hc.checkAll()

	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.checkAll()
		}
	}
}

// checkAll 检查所有渠道
func (hc *HealthChecker) checkAll() {
	channels := hc.cache.GetAllChannels()

	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func(channel *Channel) {
			defer wg.Done()
			hc.checkChannel(channel)
		}(ch)
	}

	wg.Wait()
}

// checkChannel 检查单个渠道
func (hc *HealthChecker) checkChannel(ch *Channel) {
	state := hc.getOrCreateState(ch.ID)

	// 检查是否需要进行检查
	if !hc.shouldCheck(state) {
		return
	}

	// 执行检查
	result := hc.performCheck(ch)
	result.CheckTime = time.Now()

	// 更新状态
	hc.updateChannelStatus(ch, result, state)

	// 记录统计
	atomic.AddInt64(&hc.totalChecks, 1)
	if result.Healthy {
		atomic.AddInt64(&hc.successCount, 1)
	} else {
		atomic.AddInt64(&hc.failureCount, 1)
	}

	// 调用回调
	if hc.resultCallback != nil {
		hc.resultCallback(result)
	}

	hc.logFunc("info", fmt.Sprintf("Health check for %s: %v (status: %s, latency: %dms)",
		ch.ID, result.Healthy, result.Status, result.Latency))
}

// shouldCheck 是否应该检查
func (hc *HealthChecker) shouldCheck(state *channelCheckState) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()

	now := time.Now()

	// 如果在恢复模式，使用较短的恢复间隔
	if state.inRecovery {
		return now.Sub(state.lastCheckTime) >= hc.config.RecoveryInterval
	}

	// 否则使用标准间隔
	return now.Sub(state.lastCheckTime) >= hc.config.Interval
}

// performCheck 执行检查
func (hc *HealthChecker) performCheck(ch *Channel) *HealthCheckResult {
	result := &HealthCheckResult{
		ChannelID: ch.ID,
		Status:    ch.Status,
	}

	// 构建检查 URL
	checkURL := ch.BaseURL + hc.config.HealthCheckEndpoint
	if checkURL == hc.config.HealthCheckEndpoint {
		// 如果没有配置 BaseURL，使用默认的健康检查逻辑
		result.Healthy = hc.defaultHealthCheck(ch)
		result.SuccessRate = ch.Metrics.GetSuccessRate()
		result.Status = hc.determineStatus(result.SuccessRate)
		return result
	}

	// 执行 HTTP 请求
	start := time.Now()
	resp, err := hc.httpClient.Get(checkURL)
	latency := time.Since(start)

	result.Latency = latency.Milliseconds()

	if err != nil {
		result.Healthy = false
		result.Error = err.Error()
		result.Status = ChannelStatusUnavailable
		return result
	}

	defer resp.Body.Close()

	// 检查响应状态码
	result.Healthy = resp.StatusCode >= 200 && resp.StatusCode < 300
	result.SuccessRate = ch.Metrics.GetSuccessRate()

	if result.Healthy {
		result.Status = hc.determineStatus(result.SuccessRate)
	} else {
		result.Status = ChannelStatusUnavailable
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result
}

// defaultHealthCheck 默认健康检查逻辑
func (hc *HealthChecker) defaultHealthCheck(ch *Channel) bool {
	// 基于成功率判定
	successRate := ch.Metrics.GetSuccessRate()
	return successRate >= hc.config.HealthyThreshold
}

// determineStatus 确定渠道状态
func (hc *HealthChecker) determineStatus(successRate float64) ChannelStatus {
	if successRate >= hc.config.HealthyThreshold {
		return ChannelStatusHealthy
	}
	if successRate >= hc.config.DegradedThreshold {
		return ChannelStatusDegraded
	}
	return ChannelStatusUnavailable
}

// updateChannelStatus 更新渠道状态
func (hc *HealthChecker) updateChannelStatus(ch *Channel, result *HealthCheckResult, state *channelCheckState) {
	state.mu.Lock()
	defer state.mu.Unlock()

	state.lastCheckTime = time.Now()

	// 如果检查失败
	if !result.Healthy {
		atomic.AddInt64(&state.consecutiveFailures, 1)

		// 如果连续失败过多，标记为不可用
		if atomic.LoadInt64(&state.consecutiveFailures) >= int64(hc.config.MaxConsecutiveFailures) {
			result.Status = ChannelStatusUnavailable
			state.inRecovery = true
		}

		// 记录渠道失败
		ch.RecordFailure()
		return
	}

	// 检查成功，重置连续失败计数
	atomic.StoreInt64(&state.consecutiveFailures, 0)

	// 如果之前是不可用状态，现在恢复
	if state.lastStatus == ChannelStatusUnavailable && result.Status != ChannelStatusUnavailable {
		state.inRecovery = false
		hc.logFunc("info", fmt.Sprintf("Channel %s recovered from unavailable", ch.ID))
	}

	// 更新渠道状态
	ch.SetStatus(result.Status)
	state.lastStatus = result.Status

	// 记录渠道成功
	ch.RecordSuccess(result.Latency)
}

// GetStatistics 获取统计信息
func (hc *HealthChecker) GetStatistics() map[string]interface{} {
	hc.statsMu.RLock()
	defer hc.statsMu.RUnlock()

	total := atomic.LoadInt64(&hc.totalChecks)
	success := atomic.LoadInt64(&hc.successCount)
	failure := atomic.LoadInt64(&hc.failureCount)

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_checks":    total,
		"success_count":   success,
		"failure_count":   failure,
		"success_rate":    successRate,
		"interval":        hc.config.Interval.String(),
		"timeout":         hc.config.Timeout.String(),
		"check_states":    len(hc.checkStates),
	}
}

// getOrCreateState 获取或创建检查状态
func (hc *HealthChecker) getOrCreateState(channelID string) *channelCheckState {
	hc.statesMu.Lock()
	defer hc.statesMu.Unlock()

	if state, ok := hc.checkStates[channelID]; ok {
		return state
	}

	state := &channelCheckState{
		lastCheckTime:       time.Now().Add(-hc.config.Interval),
		consecutiveFailures: 0,
		inRecovery:          false,
	}
	hc.checkStates[channelID] = state

	return state
}

// SetResultCallback 设置结果回调
func (hc *HealthChecker) SetResultCallback(callback func(*HealthCheckResult)) {
	hc.resultCallback = callback
}

// HealthCheckManager 健康检查管理器
type HealthCheckManager struct {
	// 健康检查器映射
	checkers map[string]*HealthChecker
	checkersMu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewHealthCheckManager 创建管理器
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		checkers: make(map[string]*HealthChecker),
		logFunc:  defaultLogFunc,
	}
}

// RegisterChecker 注册检查器
func (hcm *HealthCheckManager) RegisterChecker(name string, checker *HealthChecker) {
	hcm.checkersMu.Lock()
	defer hcm.checkersMu.Unlock()

	if _, exists := hcm.checkers[name]; exists {
		hcm.logFunc("warn", fmt.Sprintf("Checker %s already exists, overwriting", name))
	}

	hcm.checkers[name] = checker
}

// StartAll 启动所有检查器
func (hcm *HealthCheckManager) StartAll() {
	hcm.checkersMu.RLock()
	defer hcm.checkersMu.RUnlock()

	for name, checker := range hcm.checkers {
		checker.Start()
		hcm.logFunc("info", fmt.Sprintf("Started checker: %s", name))
	}
}

// StopAll 停止所有检查器
func (hcm *HealthCheckManager) StopAll() {
	hcm.checkersMu.RLock()
	defer hcm.checkersMu.RUnlock()

	for name, checker := range hcm.checkers {
		checker.Stop()
		hcm.logFunc("info", fmt.Sprintf("Stopped checker: %s", name))
	}
}

// GetChecker 获取检查器
func (hcm *HealthCheckManager) GetChecker(name string) *HealthChecker {
	hcm.checkersMu.RLock()
	defer hcm.checkersMu.RUnlock()

	return hcm.checkers[name]
}

// GetAllStatistics 获取所有统计信息
func (hcm *HealthCheckManager) GetAllStatistics() map[string]map[string]interface{} {
	hcm.checkersMu.RLock()
	defer hcm.checkersMu.RUnlock()

	result := make(map[string]map[string]interface{})
	for name, checker := range hcm.checkers {
		result[name] = checker.GetStatistics()
	}

	return result
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	// 渠道 ID
	channelID string

	// 状态
	state CircuitState

	// 失败次数
	failureCount int64

	// 成功次数
	successCount int64

	// 失败阈值
	failureThreshold int64

	// 成功阈值
	successThreshold int64

	// 超时时间
	timeout time.Duration

	// 最后状态变更时间
	lastStateChangeTime time.Time

	// 互斥锁
	mu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// CircuitState 断路器状态
type CircuitState int

const (
	// 关闭（正常）
	CircuitClosed CircuitState = iota
	// 打开（熔断）
	CircuitOpen
	// 半开（尝试恢复）
	CircuitHalfOpen
)

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(channelID string, failureThreshold, successThreshold int64, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		channelID:           channelID,
		state:               CircuitClosed,
		failureCount:        0,
		successCount:        0,
		failureThreshold:    failureThreshold,
		successThreshold:    successThreshold,
		timeout:             timeout,
		lastStateChangeTime: time.Now(),
		logFunc:             defaultLogFunc,
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		// 保持关闭状态
		cb.failureCount = 0

	case CircuitOpen:
		// 如果已超时，转换到半开状态
		if time.Since(cb.lastStateChangeTime) >= cb.timeout {
			cb.state = CircuitHalfOpen
			cb.successCount = 1
			cb.logFunc("info", fmt.Sprintf("Circuit breaker %s opened -> half-open", cb.channelID))
		}

	case CircuitHalfOpen:
		// 累计成功计数
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = CircuitClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.lastStateChangeTime = time.Now()
			cb.logFunc("info", fmt.Sprintf("Circuit breaker %s half-open -> closed", cb.channelID))
		}
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		// 累计失败计数
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = CircuitOpen
			cb.lastStateChangeTime = time.Now()
			cb.logFunc("info", fmt.Sprintf("Circuit breaker %s closed -> open", cb.channelID))
		}

	case CircuitHalfOpen:
		// 直接打开
		cb.state = CircuitOpen
		cb.lastStateChangeTime = time.Now()
		cb.logFunc("info", fmt.Sprintf("Circuit breaker %s half-open -> open", cb.channelID))
	}
}

// IsAvailable 是否可用
func (cb *CircuitBreaker) IsAvailable() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state != CircuitOpen
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

