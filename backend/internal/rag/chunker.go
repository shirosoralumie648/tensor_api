package rag

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
)

// Chunk 文本块
type Chunk struct {
	// 块 ID
	ID string `json:"id"`

	// 文档 ID
	DocumentID string `json:"document_id"`

	// 块索引
	ChunkIndex int `json:"chunk_index"`

	// 内容
	Content string `json:"content"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// Token 数（估计）
	TokenCount int `json:"token_count"`

	// 开始位置
	StartPosition int `json:"start_position"`

	// 结束位置
	EndPosition int `json:"end_position"`
}

// ChunkingStrategy 分块策略
type ChunkingStrategy string

const (
	// 按段落分块
	StrategyParagraph ChunkingStrategy = "paragraph"

	// 按句子分块
	StrategySentence ChunkingStrategy = "sentence"

	// 按固定长度分块
	StrategyFixed ChunkingStrategy = "fixed"

	// 混合分块
	StrategyHybrid ChunkingStrategy = "hybrid"
)

// Chunker 文本分块器
type Chunker struct {
	// 块大小（tokens）
	chunkSize int

	// 块重叠（tokens）
	chunkOverlap int

	// 分块策略
	strategy ChunkingStrategy

	// 分隔符
	separators []string

	// 是否启用智能分块
	smartChunking bool

	// Token计数器
	tokenCounter *TokenCounter

	// 统计信息
	totalChunked int64
	totalTokens  int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// defaultLogFunc 默认日志函数
func defaultLogFunc(level, msg string, args ...interface{}) {
	// 默认实现：忽略日志
}

// NewChunker 创建分块器
func NewChunker(chunkSize, chunkOverlap int, strategy ChunkingStrategy) *Chunker {
	return &Chunker{
		chunkSize:     chunkSize,
		chunkOverlap:  chunkOverlap,
		strategy:      strategy,
		smartChunking: true,
		tokenCounter:  NewTokenCounter(),
		separators: []string{
			"\n\n", // 段落分隔符
			"\n",   // 行分隔符
			"。",    // 中文句号
			"！",    // 中文感叹号
			"？",    // 中文问号
			".",    // 英文句号
			"!",    // 英文感叹号
			"?",    // 英文问号
			" ",    // 空格
		},
		logFunc: defaultLogFunc,
	}
}

// estimateTokens 估计 Token 数
func (c *Chunker) estimateTokens(text string) int {
	// 简单估计：平均每个单词4个字符，1个Token
	words := strings.Fields(text)
	chars := len(text)

	// 对于中文，按字符数估计
	hasChineseChars := false
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hasChineseChars = true
			break
		}
	}

	if hasChineseChars {
		// 中文：大约3个字符 = 1个Token
		return (chars + 2) / 3
	}

	// 英文：单词数 + 标点符号
	return len(words) + strings.Count(text, ".") + strings.Count(text, ",")
}

// splitByStrategy 按策略分割文本
func (c *Chunker) splitByStrategy(text string) []string {
	switch c.strategy {
	case StrategyParagraph:
		return c.splitByParagraph(text)
	case StrategySentence:
		return c.splitBySentence(text)
	case StrategyFixed:
		return c.splitByFixed(text)
	case StrategyHybrid:
		return c.splitByHybrid(text)
	default:
		return c.splitByParagraph(text)
	}
}

// splitByParagraph 按段落分割
func (c *Chunker) splitByParagraph(text string) []string {
	parts := strings.Split(text, "\n\n")

	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// splitBySentence 按句子分割
func (c *Chunker) splitBySentence(text string) []string {
	// 使用多个分隔符
	delimiters := []string{"。", "！", "？", ".", "!", "?", "\n"}

	var result []string
	current := ""

	for _, ch := range text {
		current += string(ch)

		// 检查是否到达任何分隔符
		for _, delim := range delimiters {
			if strings.HasSuffix(current, delim) {
				current = strings.TrimSpace(current)
				if current != "" {
					result = append(result, current)
				}
				current = ""
				break
			}
		}
	}

	if strings.TrimSpace(current) != "" {
		result = append(result, strings.TrimSpace(current))
	}

	return result
}

// splitByFixed 按固定长度分割
func (c *Chunker) splitByFixed(text string) []string {
	chars := 1000 // 固定块大小为1000字符

	var result []string
	for i := 0; i < len(text); i += chars {
		end := i + chars
		if end > len(text) {
			end = len(text)
		}

		chunk := text[i:end]
		if strings.TrimSpace(chunk) != "" {
			result = append(result, chunk)
		}
	}

	return result
}

// splitByHybrid 混合分割
func (c *Chunker) splitByHybrid(text string) []string {
	// 首先按段落分割
	parts := c.splitByParagraph(text)

	var result []string

	for _, part := range parts {
		if c.estimateTokens(part) <= c.chunkSize {
			result = append(result, part)
		} else {
			// 如果段落太大，按句子分割
			sentences := c.splitBySentence(part)
			result = append(result, sentences...)
		}
	}

	return result
}

// mergeChunks 合并块以满足大小要求
func (c *Chunker) mergeChunks(parts []string) []string {
	var result []string
	var current strings.Builder
	currentTokens := 0

	for _, part := range parts {
		partTokens := c.estimateTokens(part)

		// 如果添加此部分会超过限制
		if currentTokens+partTokens > c.chunkSize && currentTokens > 0 {
			// 保存当前块
			chunk := current.String()
			if strings.TrimSpace(chunk) != "" {
				result = append(result, chunk)
			}

			// 开始新块
			current.Reset()
			current.WriteString(part)
			currentTokens = partTokens
		} else {
			// 添加到当前块
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			current.WriteString(part)
			currentTokens += partTokens + 1 // +1 用于换行符
		}
	}

	// 保存最后一个块
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// addOverlap 添加块重叠
func (c *Chunker) addOverlap(chunks []string) []string {
	if c.chunkOverlap <= 0 {
		return chunks
	}

	var result []string

	for i, chunk := range chunks {
		result = append(result, chunk)

		// 如果有下一个块，添加重叠内容
		if i < len(chunks)-1 {
			nextChunk := chunks[i+1]
			overlapSize := c.chunkOverlap

			if len(nextChunk) < overlapSize {
				overlapSize = len(nextChunk)
			}

			overlapContent := nextChunk[:overlapSize]
			overlappedChunk := chunk + "\n" + overlapContent

			result[len(result)-1] = overlappedChunk
		}
	}

	return result
}

// ChunkText 分块文本
func (c *Chunker) ChunkText(text string) ([]*Chunk, error) {
	startTime := time.Now()

	// 分割文本
	parts := c.splitByStrategy(text)

	// 合并块
	merged := c.mergeChunks(parts)

	// 创建 Chunk 对象
	chunks := make([]*Chunk, len(merged))
	position := 0

	for i, content := range merged {
		tokens := c.estimateTokens(content)

		chunks[i] = &Chunk{
			ID:            fmt.Sprintf("chunk-%d-%d", time.Now().UnixNano(), i),
			ChunkIndex:    i,
			Content:       content,
			CreatedAt:     time.Now(),
			TokenCount:    tokens,
			StartPosition: position,
			EndPosition:   position + len(content),
			Metadata: map[string]interface{}{
				"strategy": c.strategy,
				"tokens":   tokens,
			},
		}

		position += len(content)
		atomic.AddInt64(&c.totalTokens, int64(tokens))
	}

	atomic.AddInt64(&c.totalChunked, int64(len(chunks)))

	c.logFunc("info", fmt.Sprintf("Chunked %d text into %d chunks in %v", len(text), len(chunks), time.Since(startTime)))

	return chunks, nil
}

// ChunkDocument 分块文档
func (c *Chunker) ChunkDocument(documentID, title, content string, metadata map[string]interface{}) ([]*Chunk, error) {
	chunks, err := c.ChunkText(content)
	if err != nil {
		return nil, err
	}

	// 添加文档元数据
	for i, chunk := range chunks {
		chunk.DocumentID = documentID
		if chunk.Metadata == nil {
			chunk.Metadata = make(map[string]interface{})
		}
		chunk.Metadata["title"] = title
		chunk.Metadata["document_id"] = documentID

		// 合并传入的元数据
		for k, v := range metadata {
			chunk.Metadata[k] = v
		}
		chunks[i] = chunk
	}

	return chunks, nil
}

// GetStatistics 获取统计信息
func (c *Chunker) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_chunked": atomic.LoadInt64(&c.totalChunked),
		"total_tokens":  atomic.LoadInt64(&c.totalTokens),
		"chunk_size":    c.chunkSize,
		"chunk_overlap": c.chunkOverlap,
		"strategy":      c.strategy,
	}
}

// TokenCounter 简单的Token计数器
type TokenCounter struct{}

// NewTokenCounter 创建Token计数器
func NewTokenCounter() *TokenCounter {
	return &TokenCounter{}
}

// Count 计数Token
func (tc *TokenCounter) Count(text string) int {
	// 简单估计
	words := strings.Fields(text)
	chars := len(text)

	// 检查中文字符
	hasChineseChars := false
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hasChineseChars = true
			break
		}
	}

	if hasChineseChars {
		// 中文：大约3个字符 = 1个Token
		return (chars + 2) / 3
	}

	// 英文：单词数 + 标点符号
	return len(words) + strings.Count(text, ".") + strings.Count(text, ",")
}

// AdvancedChunker 高级分块器
type AdvancedChunker struct {
	// 基础分块器
	base *Chunker

	// 标题检测
	detectTitles bool

	// 代码块检测
	detectCodeBlocks bool

	// 表格检测
	detectTables bool

	// 互斥锁
	mu sync.RWMutex
}

// NewAdvancedChunker 创建高级分块器
func NewAdvancedChunker(chunkSize, chunkOverlap int) *AdvancedChunker {
	return &AdvancedChunker{
		base:             NewChunker(chunkSize, chunkOverlap, StrategyHybrid),
		detectTitles:     true,
		detectCodeBlocks: true,
		detectTables:     true,
	}
}

// ChunkWithStructure 带结构的分块
func (ac *AdvancedChunker) ChunkWithStructure(content string) ([]*Chunk, error) {
	chunks, err := ac.base.ChunkText(content)
	if err != nil {
		return nil, err
	}

	// 检测结构并添加元数据
	for i, chunk := range chunks {
		ac.detectStructure(chunk, i, content)
	}

	return chunks, nil
}

// detectStructure 检测结构
func (ac *AdvancedChunker) detectStructure(chunk *Chunk, index int, fullContent string) {
	if chunk.Metadata == nil {
		chunk.Metadata = make(map[string]interface{})
	}

	// 检测标题
	if ac.detectTitles && strings.HasPrefix(strings.TrimSpace(chunk.Content), "#") {
		chunk.Metadata["is_title"] = true
	}

	// 检测代码块
	if ac.detectCodeBlocks && strings.Contains(chunk.Content, "```") {
		chunk.Metadata["is_code"] = true
	}

	// 检测表格
	if ac.detectTables && strings.Contains(chunk.Content, "|") {
		chunk.Metadata["is_table"] = true
	}

	// 添加位置信息
	chunk.Metadata["index"] = index
}
