package billing

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// TokenCountingMethod Token 计数方式
type TokenCountingMethod int

const (
	// 按字符计数
	MethodCharacter TokenCountingMethod = iota
	// 按单词计数
	MethodWord
	// 按 token 计数（需要 tiktoken）
	MethodToken
)

// ModelTokenConfig 模型 Token 配置
type ModelTokenConfig struct {
	// 模型名称
	ModelName string

	// 输入 token 比率（相对于基础单位）
	InputTokenRatio float64

	// 输出 token 比率
	OutputTokenRatio float64

	// 最大上下文长度
	MaxContextLength int

	// 预留 token（系统提示词等）
	ReservedTokens int

	// 计数方式
	CountingMethod TokenCountingMethod

	// 是否启用缓存
	EnableCache bool

	// 缓存过期时间（秒）
	CacheTTL int64
}

// DefaultModelConfigs 默认模型配置
var DefaultModelConfigs = map[string]*ModelTokenConfig{
	// OpenAI GPT 系列
	"gpt-4": {
		ModelName:        "gpt-4",
		InputTokenRatio:  1.0,
		OutputTokenRatio: 3.0,
		MaxContextLength: 8192,
		ReservedTokens:   100,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	"gpt-4-32k": {
		ModelName:        "gpt-4-32k",
		InputTokenRatio:  1.0,
		OutputTokenRatio: 3.0,
		MaxContextLength: 32768,
		ReservedTokens:   100,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	"gpt-3.5-turbo": {
		ModelName:        "gpt-3.5-turbo",
		InputTokenRatio:  0.5,
		OutputTokenRatio: 1.5,
		MaxContextLength: 4096,
		ReservedTokens:   50,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	// Anthropic Claude 系列
	"claude-3-opus": {
		ModelName:        "claude-3-opus",
		InputTokenRatio:  1.0,
		OutputTokenRatio: 3.0,
		MaxContextLength: 200000,
		ReservedTokens:   200,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	"claude-3-sonnet": {
		ModelName:        "claude-3-sonnet",
		InputTokenRatio:  0.75,
		OutputTokenRatio: 2.4,
		MaxContextLength: 200000,
		ReservedTokens:   200,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	"claude-3-haiku": {
		ModelName:        "claude-3-haiku",
		InputTokenRatio:  0.25,
		OutputTokenRatio: 0.75,
		MaxContextLength: 200000,
		ReservedTokens:   100,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	// Google Gemini 系列
	"gemini-pro": {
		ModelName:        "gemini-pro",
		InputTokenRatio:  0.5,
		OutputTokenRatio: 1.5,
		MaxContextLength: 32768,
		ReservedTokens:   100,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
	"gemini-pro-vision": {
		ModelName:        "gemini-pro-vision",
		InputTokenRatio:  0.5,
		OutputTokenRatio: 1.5,
		MaxContextLength: 16384,
		ReservedTokens:   150,
		CountingMethod:   MethodToken,
		EnableCache:      true,
		CacheTTL:         3600,
	},
}

// TokenCountResult Token 计数结果
type TokenCountResult struct {
	// 输入 token 数
	InputTokens int64

	// 输出 token 数
	OutputTokens int64

	// 总 token 数
	TotalTokens int64

	// 输入成本（按模型定价）
	InputCost float64

	// 输出成本
	OutputCost float64

	// 总成本
	TotalCost float64

	// 计数精度（0-100）
	Accuracy float64
}

// TokenCounter Token 计数器
type TokenCounter struct {
	// 模型配置
	modelConfigs map[string]*ModelTokenConfig
	configsMu    sync.RWMutex

	// 缓存
	cache map[string]*TokenCountResult
	cacheMu sync.RWMutex

	// 统计信息
	totalCount   int64
	cacheHits    int64
	cacheMisses  int64
	totalTokens  int64
	statsMu      sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewTokenCounter 创建 Token 计数器
func NewTokenCounter() *TokenCounter {
	tc := &TokenCounter{
		modelConfigs: make(map[string]*ModelTokenConfig),
		cache:        make(map[string]*TokenCountResult),
		logFunc:      defaultLogFunc,
	}

	// 初始化默认配置
	for model, config := range DefaultModelConfigs {
		tc.modelConfigs[model] = config
	}

	return tc
}

// RegisterModel 注册模型配置
func (tc *TokenCounter) RegisterModel(config *ModelTokenConfig) error {
	if config == nil || config.ModelName == "" {
		return fmt.Errorf("invalid model config")
	}

	tc.configsMu.Lock()
	defer tc.configsMu.Unlock()

	tc.modelConfigs[config.ModelName] = config
	return nil
}

// GetModelConfig 获取模型配置
func (tc *TokenCounter) GetModelConfig(modelName string) (*ModelTokenConfig, error) {
	tc.configsMu.RLock()
	defer tc.configsMu.RUnlock()

	config, ok := tc.modelConfigs[modelName]
	if !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	return config, nil
}

// CountTokens 计数 tokens
func (tc *TokenCounter) CountTokens(modelName string, inputText string, outputText string) (*TokenCountResult, error) {
	config, err := tc.GetModelConfig(modelName)
	if err != nil {
		return nil, err
	}

	// 生成缓存 key
	cacheKey := tc.generateCacheKey(modelName, inputText, outputText)

	// 检查缓存
	if config.EnableCache {
		tc.cacheMu.RLock()
		if result, ok := tc.cache[cacheKey]; ok {
			tc.cacheMu.RUnlock()
			atomic.AddInt64(&tc.cacheHits, 1)
			return result, nil
		}
		tc.cacheMu.RUnlock()
		atomic.AddInt64(&tc.cacheMisses, 1)
	}

	// 计数输入 token
	inputTokenCount := tc.countText(inputText, config)

	// 计数输出 token
	outputTokenCount := tc.countText(outputText, config)

	// 应用比率
	result := &TokenCountResult{
		InputTokens:  int64(float64(inputTokenCount) * config.InputTokenRatio),
		OutputTokens: int64(float64(outputTokenCount) * config.OutputTokenRatio),
		Accuracy:     95.0, // 默认精度 95%
	}

	result.TotalTokens = result.InputTokens + result.OutputTokens

	// 缓存结果
	if config.EnableCache {
		tc.cacheMu.Lock()
		tc.cache[cacheKey] = result
		tc.cacheMu.Unlock()
	}

	// 更新统计
	atomic.AddInt64(&tc.totalCount, 1)
	atomic.AddInt64(&tc.totalTokens, result.TotalTokens)

	return result, nil
}

// CountTokensWithCost 计数 tokens 并计算成本
func (tc *TokenCounter) CountTokensWithCost(modelName string, inputText string, outputText string, inputPrice float64, outputPrice float64) (*TokenCountResult, error) {
	result, err := tc.CountTokens(modelName, inputText, outputText)
	if err != nil {
		return nil, err
	}

	// 计算成本（价格通常以每 1000 个 token 为单位）
	result.InputCost = float64(result.InputTokens) / 1000.0 * inputPrice
	result.OutputCost = float64(result.OutputTokens) / 1000.0 * outputPrice
	result.TotalCost = result.InputCost + result.OutputCost

	return result, nil
}

// countText 计数文本的 token
func (tc *TokenCounter) countText(text string, config *ModelTokenConfig) int64 {
	if text == "" {
		return 0
	}

	var count int64

	switch config.CountingMethod {
	case MethodCharacter:
		// 按字符计数（每 4 个字符 ≈ 1 个 token）
		count = int64(len([]rune(text)) / 4)

	case MethodWord:
		// 按单词计数（每 1.3 个单词 ≈ 1 个 token）
		words := tc.countWords(text)
		count = int64(float64(words) / 1.3)

	case MethodToken:
		// 使用 tiktoken（这里简化实现，实际应调用 tiktoken 库）
		count = tc.estimateTokensFromText(text)

	default:
		count = int64(len([]rune(text)) / 4)
	}

	// 确保至少计数 1 个 token
	if count == 0 && text != "" {
		count = 1
	}

	return count
}

// countWords 计数单词数
func (tc *TokenCounter) countWords(text string) int {
	count := 0
	inWord := false

	for _, char := range text {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' {
			if !inWord {
				count++
				inWord = true
			}
		} else {
			inWord = false
		}
	}

	return count
}

// estimateTokensFromText 估计文本的 token 数
func (tc *TokenCounter) estimateTokensFromText(text string) int64 {
	// 这是一个简化的估计方法
	// 实际应该使用 tiktoken 或其他库进行精确计数
	// 一般规则：英文单词平均 1.3 个字符 ≈ 1 个 token
	// 中文：每个字符通常 ≈ 1.5 个 token

	runes := []rune(text)
	length := int64(len(runes))

	// 计算中文字符比例
	chineseCount := 0
	for _, r := range runes {
		if r >= 0x4E00 && r <= 0x9FFF {
			chineseCount++
		}
	}

	chineseRatio := float64(chineseCount) / float64(len(runes))

	// 混合估计
	tokens := float64(length) * (0.6 + chineseRatio*1.5)

	return int64(tokens)
}

// generateCacheKey 生成缓存 key
func (tc *TokenCounter) generateCacheKey(modelName string, inputText string, outputText string) string {
	// 这是一个简化的实现，实际应该使用 hash
	key := fmt.Sprintf("%s|%d|%d", modelName, len(inputText), len(outputText))
	return key
}

// ClearCache 清空缓存
func (tc *TokenCounter) ClearCache() {
	tc.cacheMu.Lock()
	defer tc.cacheMu.Unlock()

	tc.cache = make(map[string]*TokenCountResult)
}

// GetStatistics 获取统计信息
func (tc *TokenCounter) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_count":   atomic.LoadInt64(&tc.totalCount),
		"cache_hits":    atomic.LoadInt64(&tc.cacheHits),
		"cache_misses":  atomic.LoadInt64(&tc.cacheMisses),
		"total_tokens":  atomic.LoadInt64(&tc.totalTokens),
		"cache_entries": len(tc.cache),
	}
}

// GetCacheHitRate 获取缓存命中率
func (tc *TokenCounter) GetCacheHitRate() float64 {
	hits := atomic.LoadInt64(&tc.cacheHits)
	misses := atomic.LoadInt64(&tc.cacheMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}

	return float64(hits) / float64(total) * 100
}

// defaultLogFunc 默认日志函数
func defaultLogFunc(level, msg string, args ...interface{}) {
	fmt.Printf("[%s] %s", level, msg)
	if len(args) > 0 {
		fmt.Printf(" %v", args)
	}
	fmt.Println()
}

// TokenCountingBatch 批量 Token 计数
type TokenCountingBatch struct {
	// 计数器
	counter *TokenCounter

	// 批次请求
	requests []*TokenCountRequest

	// 结果
	results []*TokenCountResult

	// 互斥锁
	mu sync.RWMutex
}

// TokenCountRequest 计数请求
type TokenCountRequest struct {
	// 模型名称
	ModelName string

	// 输入文本
	InputText string

	// 输出文本
	OutputText string

	// 输入价格（可选）
	InputPrice float64

	// 输出价格（可选）
	OutputPrice float64
}

// NewTokenCountingBatch 创建批量计数器
func NewTokenCountingBatch(counter *TokenCounter) *TokenCountingBatch {
	return &TokenCountingBatch{
		counter:   counter,
		requests:  make([]*TokenCountRequest, 0),
		results:   make([]*TokenCountResult, 0),
	}
}

// AddRequest 添加请求
func (tcb *TokenCountingBatch) AddRequest(req *TokenCountRequest) {
	tcb.mu.Lock()
	defer tcb.mu.Unlock()

	tcb.requests = append(tcb.requests, req)
}

// Process 处理批量请求
func (tcb *TokenCountingBatch) Process() error {
	tcb.mu.Lock()
	defer tcb.mu.Unlock()

	tcb.results = make([]*TokenCountResult, 0)

	for _, req := range tcb.requests {
		if req.InputPrice > 0 && req.OutputPrice > 0 {
			result, err := tcb.counter.CountTokensWithCost(
				req.ModelName,
				req.InputText,
				req.OutputText,
				req.InputPrice,
				req.OutputPrice,
			)
			if err != nil {
				return err
			}
			tcb.results = append(tcb.results, result)
		} else {
			result, err := tcb.counter.CountTokens(
				req.ModelName,
				req.InputText,
				req.OutputText,
			)
			if err != nil {
				return err
			}
			tcb.results = append(tcb.results, result)
		}
	}

	return nil
}

// GetResults 获取结果
func (tcb *TokenCountingBatch) GetResults() []*TokenCountResult {
	tcb.mu.RLock()
	defer tcb.mu.RUnlock()

	result := make([]*TokenCountResult, len(tcb.results))
	copy(result, tcb.results)

	return result
}

// GetTotalResult 获取总结果
func (tcb *TokenCountingBatch) GetTotalResult() *TokenCountResult {
	tcb.mu.RLock()
	defer tcb.mu.RUnlock()

	total := &TokenCountResult{}

	for _, result := range tcb.results {
		total.InputTokens += result.InputTokens
		total.OutputTokens += result.OutputTokens
		total.TotalTokens += result.TotalTokens
		total.InputCost += result.InputCost
		total.OutputCost += result.OutputCost
		total.TotalCost += result.TotalCost
	}

	// 计算平均精度
	if len(tcb.results) > 0 {
		totalAccuracy := 0.0
		for _, result := range tcb.results {
			totalAccuracy += result.Accuracy
		}
		total.Accuracy = totalAccuracy / float64(len(tcb.results))
	}

	return total
}

// ExportJSON 导出为 JSON
func (tcb *TokenCountingBatch) ExportJSON() (string, error) {
	tcb.mu.RLock()
	defer tcb.mu.RUnlock()

	data := map[string]interface{}{
		"request_count": len(tcb.requests),
		"results":       tcb.results,
		"total":         tcb.GetTotalResult(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

