package relay

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// ChannelSelectorStrategy 渠道选择策略
type ChannelSelectorStrategy int

const (
	// 随机选择
	SelectorStrategyRandom ChannelSelectorStrategy = iota
	// 轮询
	SelectorStrategyRoundRobin
	// 加权轮询
	SelectorStrategyWeightedRoundRobin
	// 最少连接
	SelectorStrategyLeastConnection
	// 最低延迟
	SelectorStrategyLowestLatency
)

// String 返回策略的字符串表示
func (s ChannelSelectorStrategy) String() string {
	switch s {
	case SelectorStrategyRandom:
		return "random"
	case SelectorStrategyRoundRobin:
		return "round_robin"
	case SelectorStrategyWeightedRoundRobin:
		return "weighted_round_robin"
	case SelectorStrategyLeastConnection:
		return "least_connection"
	case SelectorStrategyLowestLatency:
		return "lowest_latency"
	default:
		return "unknown"
	}
}

// ChannelSelector 渠道选择器
type ChannelSelector struct {
	// 缓存管理器
	cache *ChannelCache

	// 选择策略
	strategy ChannelSelectorStrategy

	// 轮询计数器（用于轮询和加权轮询）
	roundRobinCounter int64

	// 通配符匹配规则
	wildcardRules map[string]*WildcardRule
	wildcardMu    sync.RWMutex

	// 统计信息
	totalSelections  int64
	statisticsCache  map[string]*SelectorStatistics
	statisticsMu     sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// WildcardRule 通配符规则
type WildcardRule struct {
	// 规则 ID
	ID string

	// 模式（如 "gpt-*", "claude-*"）
	Pattern string

	// 匹配的渠道类型
	ChannelType string

	// 优先渠道 ID 列表
	PriorityChannels []string

	// 规则权重
	Weight int

	// 是否启用
	Enabled bool
}

// SelectorStatistics 选择器统计
type SelectorStatistics struct {
	// 选择次数
	SelectionCount int64

	// 选择成功次数
	SuccessCount int64

	// 选择失败次数
	FailureCount int64

	// 平均选择时间（微秒）
	AvgSelectionTime float64

	// 最后选择时间
	LastSelectionTime int64
}

// NewChannelSelector 创建新的渠道选择器
func NewChannelSelector(cache *ChannelCache, strategy ChannelSelectorStrategy) *ChannelSelector {
	return &ChannelSelector{
		cache:            cache,
		strategy:         strategy,
		roundRobinCounter: 0,
		wildcardRules:    make(map[string]*WildcardRule),
		statisticsCache:  make(map[string]*SelectorStatistics),
		logFunc:          defaultLogFunc,
	}
}

// SelectChannel 选择一个渠道
func (cs *ChannelSelector) SelectChannel(options *ChannelSelectOptions) (*Channel, error) {
	if options == nil {
		return nil, fmt.Errorf("select options cannot be nil")
	}

	// 获取候选渠道
	candidates := cs.getCandidates(options)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available channels found")
	}

	// 根据策略选择
	var selected *Channel
	switch cs.strategy {
	case SelectorStrategyRandom:
		selected = cs.selectRandom(candidates)

	case SelectorStrategyRoundRobin:
		selected = cs.selectRoundRobin(candidates)

	case SelectorStrategyWeightedRoundRobin:
		selected = cs.selectWeightedRoundRobin(candidates)

	case SelectorStrategyLeastConnection:
		selected = cs.selectLeastConnection(candidates)

	case SelectorStrategyLowestLatency:
		selected = cs.selectLowestLatency(candidates)

	default:
		return nil, fmt.Errorf("unknown strategy: %d", cs.strategy)
	}

	// 记录统计
	cs.recordSelection(selected)

	return selected, nil
}

// getCandidates 获取候选渠道
func (cs *ChannelSelector) getCandidates(options *ChannelSelectOptions) []*Channel {
	// 先应用通配符规则
	model := options.Model
	if rule, ok := cs.matchWildcardRule(model); ok && rule.Enabled && len(rule.PriorityChannels) > 0 {
		// 使用优先渠道
		candidates := make([]*Channel, 0)
		for _, chID := range rule.PriorityChannels {
			if ch, err := cs.cache.GetChannel(chID); err == nil && ch.IsAvailable() {
				candidates = append(candidates, ch)
			}
		}
		if len(candidates) > 0 {
			return candidates
		}
	}

	// 创建过滤条件
	filter := &ChannelFilter{
		Type:            options.ChannelType,
		Model:           model,
		Region:          options.Region,
		MinAvailability: options.MinAvailability,
		OnlyEnabled:     true,
	}

	// 获取所有匹配的渠道
	return cs.cache.FilterChannels(filter)
}

// selectRandom 随机选择
func (cs *ChannelSelector) selectRandom(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}
	idx := rand.Intn(len(candidates))
	return candidates[idx]
}

// selectRoundRobin 轮询选择
func (cs *ChannelSelector) selectRoundRobin(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	counter := atomic.AddInt64(&cs.roundRobinCounter, 1)
	idx := int(counter) % len(candidates)
	return candidates[idx]
}

// selectWeightedRoundRobin 加权轮询选择
func (cs *ChannelSelector) selectWeightedRoundRobin(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0
	for _, ch := range candidates {
		totalWeight += ch.Weight
	}

	if totalWeight == 0 {
		return cs.selectRandom(candidates)
	}

	// 根据权重随机选择
	target := rand.Intn(totalWeight)
	current := 0

	for _, ch := range candidates {
		current += ch.Weight
		if target < current {
			return ch
		}
	}

	return candidates[len(candidates)-1]
}

// selectLeastConnection 最少连接选择
func (cs *ChannelSelector) selectLeastConnection(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	// 按当前并发数排序
	sort.Slice(candidates, func(i, j int) bool {
		iConcurrency := atomic.LoadInt64(&candidates[i].Metrics.CurrentConcurrency)
		jConcurrency := atomic.LoadInt64(&candidates[j].Metrics.CurrentConcurrency)
		return iConcurrency < jConcurrency
	})

	return candidates[0]
}

// selectLowestLatency 最低延迟选择
func (cs *ChannelSelector) selectLowestLatency(candidates []*Channel) *Channel {
	if len(candidates) == 0 {
		return nil
	}

	// 按平均延迟排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Metrics.AvgLatency < candidates[j].Metrics.AvgLatency
	})

	return candidates[0]
}

// AddWildcardRule 添加通配符规则
func (cs *ChannelSelector) AddWildcardRule(rule *WildcardRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	cs.wildcardMu.Lock()
	defer cs.wildcardMu.Unlock()

	cs.wildcardRules[rule.ID] = rule
	return nil
}

// RemoveWildcardRule 移除通配符规则
func (cs *ChannelSelector) RemoveWildcardRule(ruleID string) error {
	cs.wildcardMu.Lock()
	defer cs.wildcardMu.Unlock()

	if _, ok := cs.wildcardRules[ruleID]; !ok {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	delete(cs.wildcardRules, ruleID)
	return nil
}

// matchWildcardRule 匹配通配符规则
func (cs *ChannelSelector) matchWildcardRule(model string) (*WildcardRule, bool) {
	cs.wildcardMu.RLock()
	defer cs.wildcardMu.RUnlock()

	// 按优先级查找规则
	var matched *WildcardRule
	var maxWeight int

	for _, rule := range cs.wildcardRules {
		if !rule.Enabled {
			continue
		}

		if cs.matchPattern(model, rule.Pattern) {
			if rule.Weight > maxWeight {
				matched = rule
				maxWeight = rule.Weight
			}
		}
	}

	return matched, matched != nil
}

// matchPattern 模式匹配（支持通配符 *）
func (cs *ChannelSelector) matchPattern(text, pattern string) bool {
	// 简单的通配符匹配
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return text == pattern
	}

	// 检查前缀
	if parts[0] != "" && !strings.HasPrefix(text, parts[0]) {
		return false
	}

	// 检查后缀
	if parts[len(parts)-1] != "" && !strings.HasSuffix(text, parts[len(parts)-1]) {
		return false
	}

	// 检查中间部分
	pos := len(parts[0])
	for i := 1; i < len(parts)-1; i++ {
		if parts[i] == "" {
			continue
		}
		idx := strings.Index(text[pos:], parts[i])
		if idx == -1 {
			return false
		}
		pos += idx + len(parts[i])
	}

	return true
}

// recordSelection 记录选择
func (cs *ChannelSelector) recordSelection(ch *Channel) {
	atomic.AddInt64(&cs.totalSelections, 1)

	cs.statisticsMu.Lock()
	defer cs.statisticsMu.Unlock()

	if stats, ok := cs.statisticsCache[ch.ID]; ok {
		atomic.AddInt64(&stats.SelectionCount, 1)
	} else {
		cs.statisticsCache[ch.ID] = &SelectorStatistics{
			SelectionCount: 1,
		}
	}
}

// GetStatistics 获取统计信息
func (cs *ChannelSelector) GetStatistics() map[string]interface{} {
	cs.statisticsMu.RLock()
	defer cs.statisticsMu.RUnlock()

	stats := make(map[string]*SelectorStatistics)
	for k, v := range cs.statisticsCache {
		stats[k] = v
	}

	return map[string]interface{}{
		"strategy":           cs.strategy.String(),
		"total_selections":   atomic.LoadInt64(&cs.totalSelections),
		"channel_statistics": stats,
	}
}

// ChannelSelectOptions 选择选项
type ChannelSelectOptions struct {
	// 渠道类型
	ChannelType string

	// 模型名称
	Model string

	// 地理位置
	Region string

	// 最小可用性（百分比）
	MinAvailability float64

	// 首选渠道 ID
	PreferredChannelID string

	// 排除的渠道 ID
	ExcludedChannelIDs map[string]bool
}

// ChannelSelectorManager 渠道选择器管理器
type ChannelSelectorManager struct {
	// 缓存
	cache *ChannelCache

	// 默认选择器
	defaultSelector *ChannelSelector

	// 按类型的选择器
	selectors map[string]*ChannelSelector
	selectorsMu sync.RWMutex

	// 规则管理器
	ruleManager *WildcardRuleManager

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewChannelSelectorManager 创建选择器管理器
func NewChannelSelectorManager(cache *ChannelCache) *ChannelSelectorManager {
	defaultSelector := NewChannelSelector(cache, SelectorStrategyWeightedRoundRobin)

	return &ChannelSelectorManager{
		cache:           cache,
		defaultSelector: defaultSelector,
		selectors:       make(map[string]*ChannelSelector),
		ruleManager:     NewWildcardRuleManager(),
		logFunc:         defaultLogFunc,
	}
}

// SelectChannel 选择渠道
func (csm *ChannelSelectorManager) SelectChannel(options *ChannelSelectOptions) (*Channel, error) {
	if options == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	selector := csm.getSelector(options.ChannelType)
	return selector.SelectChannel(options)
}

// getSelector 获取对应的选择器
func (csm *ChannelSelectorManager) getSelector(channelType string) *ChannelSelector {
	csm.selectorsMu.RLock()
	defer csm.selectorsMu.RUnlock()

	if selector, ok := csm.selectors[channelType]; ok {
		return selector
	}

	return csm.defaultSelector
}

// RegisterSelector 注册选择器
func (csm *ChannelSelectorManager) RegisterSelector(channelType string, selector *ChannelSelector) {
	csm.selectorsMu.Lock()
	defer csm.selectorsMu.Unlock()

	csm.selectors[channelType] = selector
}

// SetStrategy 设置策略
func (csm *ChannelSelectorManager) SetStrategy(channelType string, strategy ChannelSelectorStrategy) {
	csm.selectorsMu.Lock()
	defer csm.selectorsMu.Unlock()

	if selector, ok := csm.selectors[channelType]; ok {
		selector.strategy = strategy
	} else {
		selector := NewChannelSelector(csm.cache, strategy)
		csm.selectors[channelType] = selector
	}
}

// GetRuleManager 获取规则管理器
func (csm *ChannelSelectorManager) GetRuleManager() *WildcardRuleManager {
	return csm.ruleManager
}

// WildcardRuleManager 通配符规则管理器
type WildcardRuleManager struct {
	rules   map[string]*WildcardRule
	rulesMu sync.RWMutex
}

// NewWildcardRuleManager 创建规则管理器
func NewWildcardRuleManager() *WildcardRuleManager {
	return &WildcardRuleManager{
		rules: make(map[string]*WildcardRule),
	}
}

// AddRule 添加规则
func (wrm *WildcardRuleManager) AddRule(rule *WildcardRule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Pattern == "" {
		return fmt.Errorf("rule pattern is required")
	}

	wrm.rulesMu.Lock()
	defer wrm.rulesMu.Unlock()

	wrm.rules[rule.ID] = rule
	return nil
}

// RemoveRule 移除规则
func (wrm *WildcardRuleManager) RemoveRule(ruleID string) error {
	wrm.rulesMu.Lock()
	defer wrm.rulesMu.Unlock()

	if _, ok := wrm.rules[ruleID]; !ok {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	delete(wrm.rules, ruleID)
	return nil
}

// GetRule 获取规则
func (wrm *WildcardRuleManager) GetRule(ruleID string) (*WildcardRule, error) {
	wrm.rulesMu.RLock()
	defer wrm.rulesMu.RUnlock()

	if rule, ok := wrm.rules[ruleID]; ok {
		return rule, nil
	}

	return nil, fmt.Errorf("rule %s not found", ruleID)
}

// GetAllRules 获取所有规则
func (wrm *WildcardRuleManager) GetAllRules() []*WildcardRule {
	wrm.rulesMu.RLock()
	defer wrm.rulesMu.RUnlock()

	rules := make([]*WildcardRule, 0, len(wrm.rules))
	for _, rule := range wrm.rules {
		rules = append(rules, rule)
	}

	return rules
}

// ApplyRulesToSelector 应用规则到选择器
func (wrm *WildcardRuleManager) ApplyRulesToSelector(selector *ChannelSelector) error {
	wrm.rulesMu.RLock()
	defer wrm.rulesMu.RUnlock()

	for _, rule := range wrm.rules {
		if err := selector.AddWildcardRule(rule); err != nil {
			return err
		}
	}

	return nil
}

// defaultLogFunc 默认日志函数
func defaultLogFunc(level, msg string, args ...interface{}) {
	fmt.Printf("[%s] %s", level, msg)
	if len(args) > 0 {
		fmt.Printf(" %v", args)
	}
	fmt.Println()
}

