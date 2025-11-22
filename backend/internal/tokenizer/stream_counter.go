package tokenizer

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkoukk/tiktoken-go"
)

// DefaultStreamTokenCounter 默认流式Token计数器
type DefaultStreamTokenCounter struct {
	model      string
	encoder    *tiktoken.Tiktoken
	buffer     strings.Builder
	tokenCount int64
	mu         sync.Mutex
	finalized  bool
}

// NewStreamTokenCounter 创建流式Token计数器
func NewStreamTokenCounter(model string) (*DefaultStreamTokenCounter, error) {
	encoder, err := tiktoken.EncodingForModel(model)
	if err != nil {
		// 如果模型不支持，使用默认编码
		encoder, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return nil, err
		}
	}

	return &DefaultStreamTokenCounter{
		model:   model,
		encoder: encoder,
	}, nil
}

// AddChunk 添加流式数据块
func (c *DefaultStreamTokenCounter) AddChunk(chunk string) error {
	if c.finalized {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 累积到缓冲区
	c.buffer.WriteString(chunk)

	// 每累积一定长度就计算一次（避免频繁计算）
	if c.buffer.Len() > 100 {
		text := c.buffer.String()
		tokens := c.encoder.Encode(text, nil, nil)
		atomic.StoreInt64(&c.tokenCount, int64(len(tokens)))
	}

	return nil
}

// GetCurrentCount 获取当前Token计数
func (c *DefaultStreamTokenCounter) GetCurrentCount() int {
	return int(atomic.LoadInt64(&c.tokenCount))
}

// Reset 重置计数器
func (c *DefaultStreamTokenCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer.Reset()
	atomic.StoreInt64(&c.tokenCount, 0)
	c.finalized = false
}

// Finalize 完成计数（返回最终结果）
func (c *DefaultStreamTokenCounter) Finalize() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.finalized {
		return int(atomic.LoadInt64(&c.tokenCount))
	}

	// 计算最终的token数
	text := c.buffer.String()
	if text != "" {
		tokens := c.encoder.Encode(text, nil, nil)
		atomic.StoreInt64(&c.tokenCount, int64(len(tokens)))
	}

	c.finalized = true
	return int(atomic.LoadInt64(&c.tokenCount))
}

// BatchStreamTokenCounter 批量流式Token计数器（支持多个流同时计数）
type BatchStreamTokenCounter struct {
	counters map[string]*DefaultStreamTokenCounter
	mu       sync.RWMutex
}

// NewBatchStreamTokenCounter 创建批量流式Token计数器
func NewBatchStreamTokenCounter() *BatchStreamTokenCounter {
	return &BatchStreamTokenCounter{
		counters: make(map[string]*DefaultStreamTokenCounter),
	}
}

// CreateCounter 创建一个新的计数器
func (b *BatchStreamTokenCounter) CreateCounter(id string, model string) error {
	counter, err := NewStreamTokenCounter(model)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.counters[id] = counter
	return nil
}

// AddChunk 向指定计数器添加数据块
func (b *BatchStreamTokenCounter) AddChunk(id string, chunk string) error {
	b.mu.RLock()
	counter, exists := b.counters[id]
	b.mu.RUnlock()

	if !exists {
		return nil // 忽略不存在的计数器
	}

	return counter.AddChunk(chunk)
}

// GetCount 获取指定计数器的当前计数
func (b *BatchStreamTokenCounter) GetCount(id string) int {
	b.mu.RLock()
	counter, exists := b.counters[id]
	b.mu.RUnlock()

	if !exists {
		return 0
	}

	return counter.GetCurrentCount()
}

// Finalize 完成指定计数器的计数
func (b *BatchStreamTokenCounter) Finalize(id string) int {
	b.mu.RLock()
	counter, exists := b.counters[id]
	b.mu.RUnlock()

	if !exists {
		return 0
	}

	return counter.Finalize()
}

// Remove 移除指定计数器
func (b *BatchStreamTokenCounter) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.counters, id)
}

// Clear 清空所有计数器
func (b *BatchStreamTokenCounter) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.counters = make(map[string]*DefaultStreamTokenCounter)
}

// Count 返回当前计数器数量
func (b *BatchStreamTokenCounter) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.counters)
}
