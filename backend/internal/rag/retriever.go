package rag

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SearchResult 搜索结果
type SearchResult struct {
	// 块 ID
	ChunkID string `json:"chunk_id"`

	// 内容
	Content string `json:"content"`

	// 相似度分数（0-1）
	Score float32 `json:"score"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata"`

	// 检索方法
	Method string `json:"method"` // vector, bm25, hybrid

	// 排名位置
	Rank int `json:"rank"`
}

// BM25 BM25 算法参数
type BM25 struct {
	// k1 参数（通常 1.5）
	K1 float32

	// b 参数（通常 0.75）
	B float32

	// 平均文档长度
	avgDocLength float32

	// IDF 缓存
	idfCache map[string]float32
	cacheMu  sync.RWMutex
}

// NewBM25 创建 BM25
func NewBM25() *BM25 {
	return &BM25{
		K1:       1.5,
		B:        0.75,
		idfCache: make(map[string]float32),
	}
}

// calculateIDF 计算 IDF
func (bm25 *BM25) calculateIDF(term string, docCount, docFreq int) float32 {
	if docFreq == 0 {
		return 0
	}

	numerator := float64(docCount - docFreq + 1)
	denominator := float64(docFreq + 1)
	idf := float32(math.Log(numerator / denominator))
	return idf
}

// Retriever 语义检索引擎
type Retriever struct {
	// 向量存储
	vectorStore VectorStore

	// 向量化服务
	embeddingService *EmbeddingService

	// 块存储（简单的内存存储）
	chunkStore map[string]*Chunk
	chunkMu    sync.RWMutex

	// BM25 算法
	bm25 *BM25

	// 搜索统计
	totalSearches  int64
	totalVectorOps int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// defaultLogFuncRet 默认日志函数
func defaultLogFuncRet(level, msg string, args ...interface{}) {
	// 默认实现：忽略日志
}

// NewRetriever 创建检索引擎
func NewRetriever(vectorStore VectorStore, embeddingService *EmbeddingService) *Retriever {
	return &Retriever{
		vectorStore:      vectorStore,
		embeddingService: embeddingService,
		chunkStore:       make(map[string]*Chunk),
		bm25:             NewBM25(),
		logFunc:          defaultLogFuncRet,
	}
}

// IndexChunk 索引块
func (r *Retriever) IndexChunk(ctx context.Context, chunk *Chunk) error {
	// 生成向量
	embedding, err := r.embeddingService.Embed(ctx, chunk.Content)
	if err != nil {
		return err
	}

	embedding.ChunkID = chunk.ID

	// 保存到向量存储
	if err := r.vectorStore.SaveEmbedding(ctx, embedding); err != nil {
		return err
	}

	// 保存块到本地存储
	r.chunkMu.Lock()
	r.chunkStore[chunk.ID] = chunk
	r.chunkMu.Unlock()

	atomic.AddInt64(&r.totalVectorOps, 1)

	r.logFunc("debug", fmt.Sprintf("Indexed chunk %s", chunk.ID))

	return nil
}

// IndexChunks 批量索引块
func (r *Retriever) IndexChunks(ctx context.Context, chunks []*Chunk) error {
	// 批量生成向量
	var texts []string
	for _, chunk := range chunks {
		texts = append(texts, chunk.Content)
	}

	embeddings, err := r.embeddingService.EmbedBatch(ctx, texts)
	if err != nil {
		return err
	}

	// 关联块 ID
	for i, embedding := range embeddings {
		embedding.ChunkID = chunks[i].ID
	}

	// 批量保存到向量存储
	if err := r.vectorStore.SaveEmbeddings(ctx, embeddings); err != nil {
		return err
	}

	// 保存块到本地存储
	r.chunkMu.Lock()
	for _, chunk := range chunks {
		r.chunkStore[chunk.ID] = chunk
	}
	r.chunkMu.Unlock()

	atomic.AddInt64(&r.totalVectorOps, int64(len(chunks)))

	r.logFunc("info", fmt.Sprintf("Indexed %d chunks", len(chunks)))

	return nil
}

// VectorSearch 向量搜索
func (r *Retriever) VectorSearch(ctx context.Context, query string, topK int) ([]*SearchResult, error) {
	start := time.Now()

	// 生成查询向量
	queryEmbedding, err := r.embeddingService.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	// 搜索相似向量
	embeddings, err := r.vectorStore.Search(ctx, queryEmbedding.Vector, topK)
	if err != nil {
		return nil, err
	}

	// 转换为搜索结果
	var results []*SearchResult

	for i, embedding := range embeddings {
		r.chunkMu.RLock()
		chunk, exists := r.chunkStore[embedding.ChunkID]
		r.chunkMu.RUnlock()

		if !exists {
			continue
		}

		// 计算相似度
		similarity := cosineSimilarity(queryEmbedding.Vector, embedding.Vector)

		results = append(results, &SearchResult{
			ChunkID:  embedding.ChunkID,
			Content:  chunk.Content,
			Score:    similarity,
			Method:   "vector",
			Rank:     i + 1,
			Metadata: chunk.Metadata,
		})
	}

	atomic.AddInt64(&r.totalSearches, 1)

	r.logFunc("debug", fmt.Sprintf("Vector search completed in %v", time.Since(start)))

	return results, nil
}

// BM25Search BM25 搜索
func (r *Retriever) BM25Search(ctx context.Context, query string, topK int) ([]*SearchResult, error) {
	start := time.Now()

	// 分词
	queryTerms := strings.Fields(strings.ToLower(query))

	var results []*SearchResult

	r.chunkMu.RLock()
	defer r.chunkMu.RUnlock()

	// 计算每个块的 BM25 分数
	type scoreItem struct {
		chunkID string
		chunk   *Chunk
		score   float32
	}

	var scores []scoreItem
	docCount := len(r.chunkStore)

	for id, chunk := range r.chunkStore {
		score := r.calculateBM25Score(chunk.Content, queryTerms, docCount)

		if score > 0 {
			scores = append(scores, scoreItem{id, chunk, score})
		}
	}

	// 排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 返回 Top K
	for i := 0; i < topK && i < len(scores); i++ {
		item := scores[i]

		results = append(results, &SearchResult{
			ChunkID:  item.chunkID,
			Content:  item.chunk.Content,
			Score:    item.score,
			Method:   "bm25",
			Rank:     i + 1,
			Metadata: item.chunk.Metadata,
		})
	}

	atomic.AddInt64(&r.totalSearches, 1)

	r.logFunc("debug", fmt.Sprintf("BM25 search completed in %v", time.Since(start)))

	return results, nil
}

// calculateBM25Score 计算 BM25 分数
func (r *Retriever) calculateBM25Score(doc string, queryTerms []string, docCount int) float32 {
	docTerms := strings.Fields(strings.ToLower(doc))
	docLength := len(docTerms)

	if docLength == 0 {
		return 0
	}

	score := float32(0)

	for _, term := range queryTerms {
		// 计算 term 在文档中的频率
		termFreq := 0
		for _, docTerm := range docTerms {
			if docTerm == term {
				termFreq++
			}
		}

		if termFreq == 0 {
			continue
		}

		// 计算 IDF
		docFreq := 0
		r.chunkMu.RLock()
		for _, chunk := range r.chunkStore {
			if strings.Contains(strings.ToLower(chunk.Content), term) {
				docFreq++
			}
		}
		r.chunkMu.RUnlock()

		idf := r.bm25.calculateIDF(term, docCount, docFreq)

		// BM25 公式
		numerator := float32(termFreq) * (r.bm25.K1 + 1)
		denominator := float32(termFreq) + r.bm25.K1*(1-r.bm25.B+r.bm25.B*float32(docLength)/r.bm25.avgDocLength)

		score += idf * numerator / denominator
	}

	return score
}

// HybridSearch 混合搜索（向量 + BM25）
func (r *Retriever) HybridSearch(ctx context.Context, query string, topK int, vectorWeight float32) ([]*SearchResult, error) {
	// 执行向量搜索
	vectorResults, err := r.VectorSearch(ctx, query, topK*2)
	if err != nil {
		return nil, err
	}

	// 执行 BM25 搜索
	bm25Results, err := r.BM25Search(ctx, query, topK*2)
	if err != nil {
		return nil, err
	}

	// 合并结果
	scoreMap := make(map[string]float32)
	resultMap := make(map[string]*SearchResult)

	// 添加向量搜索结果
	for _, result := range vectorResults {
		score := result.Score * vectorWeight
		scoreMap[result.ChunkID] = score
		resultMap[result.ChunkID] = result
	}

	// 添加 BM25 结果
	for _, result := range bm25Results {
		bm25Weight := 1 - vectorWeight
		score := result.Score * bm25Weight
		if existing, exists := scoreMap[result.ChunkID]; exists {
			scoreMap[result.ChunkID] = existing + score
		} else {
			scoreMap[result.ChunkID] = score
			resultMap[result.ChunkID] = result
		}
	}

	// 排序
	type scoreItem struct {
		chunkID string
		score   float32
	}

	var scores []scoreItem
	for id, score := range scoreMap {
		scores = append(scores, scoreItem{id, score})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 返回 Top K
	var results []*SearchResult
	for i := 0; i < topK && i < len(scores); i++ {
		result := resultMap[scores[i].chunkID]
		result.Score = scores[i].score
		result.Method = "hybrid"
		result.Rank = i + 1
		results = append(results, result)
	}

	r.logFunc("debug", fmt.Sprintf("Hybrid search completed with %d results", len(results)))

	return results, nil
}

// DeleteChunk 删除块
func (r *Retriever) DeleteChunk(ctx context.Context, chunkID string) error {
	// 从向量存储删除
	if err := r.vectorStore.DeleteByChunkID(ctx, chunkID); err != nil {
		return err
	}

	// 从本地存储删除
	r.chunkMu.Lock()
	delete(r.chunkStore, chunkID)
	r.chunkMu.Unlock()

	return nil
}

// GetStatistics 获取统计信息
func (r *Retriever) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_searches":     atomic.LoadInt64(&r.totalSearches),
		"total_vector_ops":   atomic.LoadInt64(&r.totalVectorOps),
		"indexed_chunks":     len(r.chunkStore),
	}
}

// RerankingService 重排序服务
type RerankingService struct {
	// 重排序模型名称
	modelName string

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewRerankingService 创建重排序服务
func NewRerankingService(modelName string) *RerankingService {
	return &RerankingService{
		modelName: modelName,
		logFunc:   defaultLogFuncRet,
	}
}

// Rerank 重排序结果
func (rs *RerankingService) Rerank(query string, results []*SearchResult, topK int) []*SearchResult {
	if len(results) <= topK {
		return results
	}

	// 简单的重排序：基于查询和内容的关键词重叠
	type scoreItem struct {
		result *SearchResult
		score  float32
	}

	var scores []scoreItem
	queryTerms := make(map[string]bool)
	for _, term := range strings.Fields(strings.ToLower(query)) {
		queryTerms[term] = true
	}

	for _, result := range results {
		contentTerms := strings.Fields(strings.ToLower(result.Content))
		overlap := 0

		for _, term := range contentTerms {
			if queryTerms[term] {
				overlap++
			}
		}

		rerankScore := result.Score * 0.7
		rerankScore += float32(overlap) * 0.3

		scores = append(scores, scoreItem{result, rerankScore})
	}

	// 排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 返回 Top K
	var reranked []*SearchResult
	for i := 0; i < topK && i < len(scores); i++ {
		scores[i].result.Rank = i + 1
		reranked = append(reranked, scores[i].result)
	}

	rs.logFunc("debug", fmt.Sprintf("Reranked %d results", len(reranked)))

	return reranked
}

