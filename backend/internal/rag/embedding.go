package rag

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Embedding 向量表示
type Embedding struct {
	// ID
	ID string `json:"id"`

	// 块 ID
	ChunkID string `json:"chunk_id"`

	// 向量数据（1536维 for ada-002）
	Vector []float32 `json:"vector"`

	// 模型名称
	Model string `json:"model"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// Token 消耗
	TokensUsed int `json:"tokens_used"`
}

// EmbeddingModel 向量模型配置
type EmbeddingModel struct {
	// 模型名称
	Name string

	// 维度
	Dimension int

	// 每个请求的最大 Token 数
	MaxTokens int

	// 价格（每 1M tokens）
	PricePerMillion float64
}

// 预定义模型
var (
	ModelAdaV2 = &EmbeddingModel{
		Name:            "text-embedding-ada-002",
		Dimension:       1536,
		MaxTokens:       8191,
		PricePerMillion: 0.0001,
	}

	Model3Small = &EmbeddingModel{
		Name:            "text-embedding-3-small",
		Dimension:       512,
		MaxTokens:       8191,
		PricePerMillion: 0.02,
	}

	Model3Large = &EmbeddingModel{
		Name:            "text-embedding-3-large",
		Dimension:       3072,
		MaxTokens:       8191,
		PricePerMillion: 0.13,
	}
)

// EmbeddingService 向量化服务
type EmbeddingService struct {
	// 模型配置
	model *EmbeddingModel

	// API 客户端
	apiClient EmbeddingAPIClient

	// 缓存
	cache *EmbeddingCache

	// 统计
	totalEmbeddings int64
	totalTokensUsed int64
	totalCost       float64
	costMu          sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// defaultLogFunc 默认日志函数
func defaultLogFuncEmb(level, msg string, args ...interface{}) {
	// 默认实现：忽略日志
}

// EmbeddingAPIClient Embedding API 客户端接口
type EmbeddingAPIClient interface {
	// Embed 获取单个文本的向量
	Embed(ctx context.Context, text string) ([]float32, int, error)

	// EmbedBatch 批量获取向量
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, int, error)
}

// MockEmbeddingClient 模拟 API 客户端（用于测试）
type MockEmbeddingClient struct {
	model *EmbeddingModel
}

// Embed 生成向量（模拟）
func (mec *MockEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, int, error) {
	vec := make([]float32, mec.model.Dimension)
	// 简单的哈希算法生成确定性向量
	hash := uint32(0)
	for i, c := range text {
		hash = hash*31 + uint32(c)
		vec[i%mec.model.Dimension] = float32(hash) / float32(^uint32(0))
	}

	// 规范化向量
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	norm = float32(1.0) / float32(len(vec))

	for i := range vec {
		vec[i] *= norm
	}

	// 估计 Token 数
	tokens := len(text) / 4

	return vec, tokens, nil
}

// EmbedBatch 批量生成向量
func (mec *MockEmbeddingClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, int, error) {
	var result [][]float32
	totalTokens := 0

	for _, text := range texts {
		vec, tokens, err := mec.Embed(ctx, text)
		if err != nil {
			return nil, 0, err
		}

		result = append(result, vec)
		totalTokens += tokens
	}

	return result, totalTokens, nil
}

// NewEmbeddingService 创建向量化服务
func NewEmbeddingService(model *EmbeddingModel, client EmbeddingAPIClient) *EmbeddingService {
	return &EmbeddingService{
		model:     model,
		apiClient: client,
		cache:     NewEmbeddingCache(10000), // 10000 条缓存
		logFunc:   defaultLogFuncEmb,
	}
}

// Embed 获取单条文本的向量
func (es *EmbeddingService) Embed(ctx context.Context, text string) (*Embedding, error) {
	// 检查缓存
	cached, exists := es.cache.Get(text, es.model.Name)
	if exists {
		return cached, nil
	}

	// 调用 API
	vector, tokens, err := es.apiClient.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// 创建 Embedding
	embedding := &Embedding{
		ID:         fmt.Sprintf("emb-%d", time.Now().UnixNano()),
		Vector:     vector,
		Model:      es.model.Name,
		CreatedAt:  time.Now(),
		TokensUsed: tokens,
	}

	// 缓存结果
	es.cache.Set(text, es.model.Name, embedding)

	// 更新统计
	atomic.AddInt64(&es.totalEmbeddings, 1)
	atomic.AddInt64(&es.totalTokensUsed, int64(tokens))

	es.costMu.Lock()
	es.totalCost += float64(tokens) * es.model.PricePerMillion / 1000000
	es.costMu.Unlock()

	es.logFunc("debug", fmt.Sprintf("Embedded text (tokens: %d)", tokens))

	return embedding, nil
}

// EmbedBatch 批量获取向量
func (es *EmbeddingService) EmbedBatch(ctx context.Context, texts []string) ([]*Embedding, error) {
	// 分离缓存命中和未命中的文本
	var toFetch []string
	var toFetchIndices []int
	var result []*Embedding

	for i, text := range texts {
		if cached, exists := es.cache.Get(text, es.model.Name); exists {
			result = append(result, cached)
		} else {
			toFetch = append(toFetch, text)
			toFetchIndices = append(toFetchIndices, i)
		}
	}

	// 如果没有需要获取的文本，直接返回
	if len(toFetch) == 0 {
		return result, nil
	}

	// 调用 API 批量获取向量
	vectors, tokens, err := es.apiClient.EmbedBatch(ctx, toFetch)
	if err != nil {
		return nil, err
	}

	// 创建 Embedding 对象
	for i, vector := range vectors {
		embedding := &Embedding{
			ID:         fmt.Sprintf("emb-%d", time.Now().UnixNano()),
			Vector:     vector,
			Model:      es.model.Name,
			CreatedAt:  time.Now(),
			TokensUsed: tokens / len(vectors),
		}

		// 缓存
		es.cache.Set(toFetch[i], es.model.Name, embedding)

		// 插入到正确位置
		result = append(result, embedding)
	}

	// 更新统计
	atomic.AddInt64(&es.totalEmbeddings, int64(len(vectors)))
	atomic.AddInt64(&es.totalTokensUsed, int64(tokens))

	es.costMu.Lock()
	es.totalCost += float64(tokens) * es.model.PricePerMillion / 1000000
	es.costMu.Unlock()

	es.logFunc("info", fmt.Sprintf("Batch embedded %d texts (tokens: %d)", len(toFetch), tokens))

	return result, nil
}

// GetStatistics 获取统计信息
func (es *EmbeddingService) GetStatistics() map[string]interface{} {
	es.costMu.RLock()
	defer es.costMu.RUnlock()

	return map[string]interface{}{
		"total_embeddings": atomic.LoadInt64(&es.totalEmbeddings),
		"total_tokens":     atomic.LoadInt64(&es.totalTokensUsed),
		"total_cost":       es.totalCost,
		"model":            es.model.Name,
		"cache_size":       es.cache.Size(),
	}
}

// EmbeddingCache 向量缓存
type EmbeddingCache struct {
	// 缓存数据
	data map[string]*Embedding

	// 互斥锁
	mu sync.RWMutex

	// 最大容量
	maxSize int

	// 访问计数
	accessCount map[string]int64
	countMu     sync.RWMutex
}

// NewEmbeddingCache 创建缓存
func NewEmbeddingCache(maxSize int) *EmbeddingCache {
	return &EmbeddingCache{
		data:        make(map[string]*Embedding),
		maxSize:     maxSize,
		accessCount: make(map[string]int64),
	}
}

// Get 获取缓存
func (ec *EmbeddingCache) Get(text, model string) (*Embedding, bool) {
	key := fmt.Sprintf("%s:%s", text, model)

	ec.mu.RLock()
	embedding, exists := ec.data[key]
	ec.mu.RUnlock()

	if exists {
		ec.countMu.Lock()
		ec.accessCount[key]++
		ec.countMu.Unlock()
	}

	return embedding, exists
}

// Set 设置缓存
func (ec *EmbeddingCache) Set(text, model string, embedding *Embedding) {
	key := fmt.Sprintf("%s:%s", text, model)

	ec.mu.Lock()
	defer ec.mu.Unlock()

	// 如果缓存满了，删除访问最少的项
	if len(ec.data) >= ec.maxSize {
		ec.evictLRU()
	}

	ec.data[key] = embedding

	ec.countMu.Lock()
	ec.accessCount[key] = 1
	ec.countMu.Unlock()
}

// evictLRU 删除最少使用的项
func (ec *EmbeddingCache) evictLRU() {
	var minKey string
	var minCount int64 = 1<<63 - 1

	ec.countMu.RLock()
	for key, count := range ec.accessCount {
		if count < minCount {
			minCount = count
			minKey = key
		}
	}
	ec.countMu.RUnlock()

	if minKey != "" {
		delete(ec.data, minKey)

		ec.countMu.Lock()
		delete(ec.accessCount, minKey)
		ec.countMu.Unlock()
	}
}

// Size 获取缓存大小
func (ec *EmbeddingCache) Size() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return len(ec.data)
}

// Clear 清空缓存
func (ec *EmbeddingCache) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.data = make(map[string]*Embedding)

	ec.countMu.Lock()
	ec.accessCount = make(map[string]int64)
	ec.countMu.Unlock()
}

// VectorStore 向量存储接口
type VectorStore interface {
	// 保存 embedding
	SaveEmbedding(ctx context.Context, embedding *Embedding) error

	// 批量保存 embedding
	SaveEmbeddings(ctx context.Context, embeddings []*Embedding) error

	// 按 ID 检索
	GetEmbedding(ctx context.Context, id string) (*Embedding, error)

	// 相似度搜索
	Search(ctx context.Context, query []float32, topK int) ([]*Embedding, error)

	// 删除 embedding
	DeleteEmbedding(ctx context.Context, id string) error

	// 按块 ID 删除
	DeleteByChunkID(ctx context.Context, chunkID string) error
}

// InMemoryVectorStore 内存向量存储（用于测试）
type InMemoryVectorStore struct {
	// 向量数据
	embeddings map[string]*Embedding

	// 块 ID 索引
	chunkIndex map[string][]string

	// 互斥锁
	mu sync.RWMutex
}

// NewInMemoryVectorStore 创建内存存储
func NewInMemoryVectorStore() *InMemoryVectorStore {
	return &InMemoryVectorStore{
		embeddings: make(map[string]*Embedding),
		chunkIndex: make(map[string][]string),
	}
}

// SaveEmbedding 保存 embedding
func (ivs *InMemoryVectorStore) SaveEmbedding(ctx context.Context, embedding *Embedding) error {
	ivs.mu.Lock()
	defer ivs.mu.Unlock()

	ivs.embeddings[embedding.ID] = embedding

	if embedding.ChunkID != "" {
		ivs.chunkIndex[embedding.ChunkID] = append(ivs.chunkIndex[embedding.ChunkID], embedding.ID)
	}

	return nil
}

// SaveEmbeddings 批量保存
func (ivs *InMemoryVectorStore) SaveEmbeddings(ctx context.Context, embeddings []*Embedding) error {
	for _, embedding := range embeddings {
		if err := ivs.SaveEmbedding(ctx, embedding); err != nil {
			return err
		}
	}

	return nil
}

// GetEmbedding 获取 embedding
func (ivs *InMemoryVectorStore) GetEmbedding(ctx context.Context, id string) (*Embedding, error) {
	ivs.mu.RLock()
	defer ivs.mu.RUnlock()

	embedding, exists := ivs.embeddings[id]
	if !exists {
		return nil, fmt.Errorf("embedding not found: %s", id)
	}

	return embedding, nil
}

// Search 搜索相似的向量
func (ivs *InMemoryVectorStore) Search(ctx context.Context, query []float32, topK int) ([]*Embedding, error) {
	ivs.mu.RLock()
	defer ivs.mu.RUnlock()

	type result struct {
		embedding  *Embedding
		similarity float32
	}

	var results []result

	// 计算相似度
	for _, embedding := range ivs.embeddings {
		similarity := cosineSimilarity(query, embedding.Vector)
		results = append(results, result{embedding, similarity})
	}

	// 排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].similarity > results[i].similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 返回 Top K
	var topResults []*Embedding
	for i := 0; i < topK && i < len(results); i++ {
		topResults = append(topResults, results[i].embedding)
	}

	return topResults, nil
}

// DeleteEmbedding 删除 embedding
func (ivs *InMemoryVectorStore) DeleteEmbedding(ctx context.Context, id string) error {
	ivs.mu.Lock()
	defer ivs.mu.Unlock()

	embedding, exists := ivs.embeddings[id]
	if exists {
		delete(ivs.embeddings, id)

		// 从块索引中移除
		if embedding.ChunkID != "" {
			ids := ivs.chunkIndex[embedding.ChunkID]
			for i, cid := range ids {
				if cid == id {
					ivs.chunkIndex[embedding.ChunkID] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
	}

	return nil
}

// DeleteByChunkID 按块 ID 删除
func (ivs *InMemoryVectorStore) DeleteByChunkID(ctx context.Context, chunkID string) error {
	ivs.mu.Lock()
	defer ivs.mu.Unlock()

	ids, exists := ivs.chunkIndex[chunkID]
	if exists {
		for _, id := range ids {
			delete(ivs.embeddings, id)
		}

		delete(ivs.chunkIndex, chunkID)
	}

	return nil
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / float32(len(a))
}
