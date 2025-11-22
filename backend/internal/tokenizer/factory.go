package tokenizer

import (
	"fmt"
	"strings"
	"sync"
)

// TokenizerFactory Token计数器工厂
type TokenizerFactory struct {
	tiktokenizer *TiktokenTokenizer
	genericCache map[string]*GenericTokenizer
	mu           sync.RWMutex
}

// NewTokenizerFactory 创建Token计数器工厂
func NewTokenizerFactory() (*TokenizerFactory, error) {
	tiktoken, err := NewTiktokenTokenizer()
	if err != nil {
		return nil, fmt.Errorf("failed to create tiktoken tokenizer: %w", err)
	}

	return &TokenizerFactory{
		tiktokenizer: tiktoken,
		genericCache: make(map[string]*GenericTokenizer),
	}, nil
}

// GetTokenizer 获取指定模型的Token计数器
func (f *TokenizerFactory) GetTokenizer(model string) (Tokenizer, error) {
	// 判断模型类型
	if f.isOpenAIModel(model) {
		return f.tiktokenizer, nil
	}

	// 使用通用计数器
	return f.getGenericTokenizer(model), nil
}

// CreateStreamCounter 创建流式Token计数器
func (f *TokenizerFactory) CreateStreamCounter(model string) (StreamTokenCounter, error) {
	if f.isOpenAIModel(model) {
		return NewStreamTokenCounter(model)
	}

	// 对于非OpenAI模型，也可以使用基于tiktoken的流式计数器
	// 因为它们的估算相对准确
	return NewStreamTokenCounter(model)
}

// CreateBatchStreamCounter 创建批量流式Token计数器
func (f *TokenizerFactory) CreateBatchStreamCounter() *BatchStreamTokenCounter {
	return NewBatchStreamTokenCounter()
}

// isOpenAIModel 判断是否为OpenAI模型
func (f *TokenizerFactory) isOpenAIModel(model string) bool {
	openAIModels := []string{
		"gpt-4",
		"gpt-3.5",
		"text-embedding",
		"text-davinci",
		"text-curie",
		"text-babbage",
		"text-ada",
	}

	modelLower := strings.ToLower(model)
	for _, prefix := range openAIModels {
		if strings.HasPrefix(modelLower, prefix) {
			return true
		}
	}

	return false
}

// getGenericTokenizer 获取通用计数器（带缓存）
func (f *TokenizerFactory) getGenericTokenizer(model string) *GenericTokenizer {
	f.mu.RLock()
	tokenizer, exists := f.genericCache[model]
	f.mu.RUnlock()

	if exists {
		return tokenizer
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// 双重检查
	if tokenizer, exists := f.genericCache[model]; exists {
		return tokenizer
	}

	tokenizer = NewGenericTokenizer(model)
	f.genericCache[model] = tokenizer
	return tokenizer
}

// Close 关闭工厂并释放资源
func (f *TokenizerFactory) Close() error {
	if f.tiktokenizer != nil {
		return f.tiktokenizer.Close()
	}
	return nil
}

// GlobalFactory 全局Token计数器工厂
var (
	globalFactory     *TokenizerFactory
	globalFactoryOnce sync.Once
	globalFactoryErr  error
)

// GetGlobalFactory 获取全局Token计数器工厂
func GetGlobalFactory() (*TokenizerFactory, error) {
	globalFactoryOnce.Do(func() {
		globalFactory, globalFactoryErr = NewTokenizerFactory()
	})
	return globalFactory, globalFactoryErr
}

// CountTokensQuick 快速Token计数（使用全局工厂）
func CountTokensQuick(model string, messages []Message) (int, error) {
	factory, err := GetGlobalFactory()
	if err != nil {
		return 0, err
	}

	tokenizer, err := factory.GetTokenizer(model)
	if err != nil {
		return 0, err
	}

	return tokenizer.CountMessages(nil, messages, model)
}
