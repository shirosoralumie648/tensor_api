package rag

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// EnhancedPrompt RAG 增强提示词
type EnhancedPrompt struct {
	// 原始查询
	OriginalQuery string `json:"original_query"`

	// 检索结果
	SearchResults []*SearchResult `json:"search_results"`

	// 增强提示词
	EnhancedPrompt string `json:"enhanced_prompt"`

	// 引用信息
	Citations []*Citation `json:"citations"`

	// 质量评分
	QualityScore float32 `json:"quality_score"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 使用的检索方法
	RetrievalMethod string `json:"retrieval_method"`

	// Token 消耗
	TokensUsed int `json:"tokens_used"`
}

// Citation 引用信息
type Citation struct {
	// 引用 ID
	ID string `json:"id"`

	// 来源文档名称
	SourceName string `json:"source_name"`

	// 页码或位置
	Page int `json:"page"`

	// 内容摘录
	Content string `json:"content"`

	// 相关性分数
	Relevance float32 `json:"relevance"`
}

// RAGConfig RAG 配置
type RAGConfig struct {
	// 启用 RAG
	Enabled bool

	// 检索方法
	RetrievalMethod string // vector, bm25, hybrid

	// Top K 检索数量
	TopK int

	// 混合搜索向量权重
	VectorWeight float32

	// 最小相关性阈值
	MinRelevance float32

	// 启用自动触发
	AutoTrigger bool

	// 自动触发的最小查询长度
	MinQueryLength int

	// 提示词模板
	PromptTemplate string

	// 启用结果重排
	EnableReranking bool

	// 最大上下文长度
	MaxContextLength int
}

// DefaultRAGConfig 默认配置
func DefaultRAGConfig() *RAGConfig {
	return &RAGConfig{
		Enabled:            true,
		RetrievalMethod:    "hybrid",
		TopK:               5,
		VectorWeight:       0.7,
		MinRelevance:       0.3,
		AutoTrigger:        true,
		MinQueryLength:     10,
		EnableReranking:    true,
		MaxContextLength:   4000,
		PromptTemplate:     defaultPromptTemplate,
	}
}

// 默认提示词模板
const defaultPromptTemplate = `根据以下参考信息回答用户问题。如果参考信息中没有相关内容，请说明您无法根据提供的信息回答。

=== 参考信息 ===
%s

=== 用户问题 ===
%s

请基于参考信息回答，并在答案中标注引用来源（例如：[来源1]）。`

// RAGService RAG 服务
type RAGService struct {
	// 检索引擎
	retriever *Retriever

	// 重排序服务
	reranker *RerankingService

	// 配置
	config *RAGConfig

	// Token 计数器
	tokenCounter *TokenCounter

	// 统计信息
	totalRAGs        int64
	totalRetrievals  int64
	avgQualityScore  float32
	scoresMu         sync.RWMutex
	scores           []float32

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewRAGService 创建 RAG 服务
func NewRAGService(retriever *Retriever, config *RAGConfig) *RAGService {
	if config == nil {
		config = DefaultRAGConfig()
	}

	return &RAGService{
		retriever:    retriever,
		reranker:     NewRerankingService("cross-encoder"),
		config:       config,
		tokenCounter: NewTokenCounter(),
		scores:       make([]float32, 0),
		logFunc:      defaultLogFuncRet,
	}
}

// EnhancePrompt 增强提示词
func (rs *RAGService) EnhancePrompt(ctx context.Context, query string) (*EnhancedPrompt, error) {
	startTime := time.Now()

	// 检查是否应该使用 RAG
	if !rs.shouldUseRAG(query) {
		return nil, fmt.Errorf("RAG not triggered: query too short or RAG disabled")
	}

	// 执行检索
	var results []*SearchResult
	var retrievalMethod string
	var err error

	switch rs.config.RetrievalMethod {
	case "vector":
		results, err = rs.retriever.VectorSearch(ctx, query, rs.config.TopK)
		retrievalMethod = "vector"
	case "bm25":
		results, err = rs.retriever.BM25Search(ctx, query, rs.config.TopK)
		retrievalMethod = "bm25"
	case "hybrid":
		results, err = rs.retriever.HybridSearch(ctx, query, rs.config.TopK, rs.config.VectorWeight)
		retrievalMethod = "hybrid"
	default:
		results, err = rs.retriever.HybridSearch(ctx, query, rs.config.TopK, rs.config.VectorWeight)
		retrievalMethod = "hybrid"
	}

	if err != nil {
		return nil, err
	}

	// 过滤结果
	results = rs.filterResults(results)

	// 重排结果
	if rs.config.EnableReranking && len(results) > 0 {
		results = rs.reranker.Rerank(query, results, rs.config.TopK)
	}

	// 构建引用信息
	citations := rs.buildCitations(results)

	// 生成增强提示词
	enhancedPrompt := rs.generateEnhancedPrompt(query, results)

	// 计算质量评分
	qualityScore := rs.calculateQualityScore(results)

	// 更新统计
	atomic.AddInt64(&rs.totalRAGs, 1)
	atomic.AddInt64(&rs.totalRetrievals, int64(len(results)))
	rs.updateScores(qualityScore)

	// 计算 Token 消耗
	tokensUsed := rs.tokenCounter.Count(enhancedPrompt)

	result := &EnhancedPrompt{
		OriginalQuery:   query,
		SearchResults:   results,
		EnhancedPrompt:  enhancedPrompt,
		Citations:       citations,
		QualityScore:    qualityScore,
		CreatedAt:       time.Now(),
		RetrievalMethod: retrievalMethod,
		TokensUsed:      tokensUsed,
	}

	rs.logFunc("debug", fmt.Sprintf("Enhanced prompt in %v with %d results", time.Since(startTime), len(results)))

	return result, nil
}

// shouldUseRAG 判断是否应该使用 RAG
func (rs *RAGService) shouldUseRAG(query string) bool {
	if !rs.config.Enabled {
		return false
	}

	if !rs.config.AutoTrigger {
		return true
	}

	return len(query) >= rs.config.MinQueryLength
}

// filterResults 过滤结果
func (rs *RAGService) filterResults(results []*SearchResult) []*SearchResult {
	var filtered []*SearchResult

	for _, result := range results {
		if result.Score >= rs.config.MinRelevance {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// buildCitations 构建引用信息
func (rs *RAGService) buildCitations(results []*SearchResult) []*Citation {
	var citations []*Citation

	for i, result := range results {
		// 从元数据中提取信息
		sourceName := fmt.Sprintf("Source %d", i+1)
		if name, ok := result.Metadata["title"].(string); ok {
			sourceName = name
		}

		page := 0
		if p, ok := result.Metadata["page"].(int); ok {
			page = p
		}

		content := result.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}

		citations = append(citations, &Citation{
			ID:         fmt.Sprintf("citation-%d", i+1),
			SourceName: sourceName,
			Page:       page,
			Content:    content,
			Relevance:  result.Score,
		})
	}

	return citations
}

// generateEnhancedPrompt 生成增强提示词
func (rs *RAGService) generateEnhancedPrompt(query string, results []*SearchResult) string {
	// 构建参考信息
	var referenceParts []string

	for i, result := range results {
		// 截断长内容
		content := result.Content
		if len(content) > rs.config.MaxContextLength/len(results) {
			content = content[:rs.config.MaxContextLength/len(results)] + "..."
		}

		sourceName := fmt.Sprintf("Source %d", i+1)
		if name, ok := result.Metadata["title"].(string); ok {
			sourceName = name
		}

		ref := fmt.Sprintf("[%d] %s\n来源: %s, 相关度: %.2f", 
			i+1, content, sourceName, result.Score)
		referenceParts = append(referenceParts, ref)
	}

	references := strings.Join(referenceParts, "\n\n")

	// 应用模板
	enhancedPrompt := fmt.Sprintf(rs.config.PromptTemplate, references, query)

	return enhancedPrompt
}

// calculateQualityScore 计算质量评分
func (rs *RAGService) calculateQualityScore(results []*SearchResult) float32 {
	if len(results) == 0 {
		return 0
	}

	// 基于结果的平均相关度
	var totalScore float32
	for _, result := range results {
		totalScore += result.Score
	}

	avgScore := totalScore / float32(len(results))

	// 考虑结果数量的因素
	relevanceFactor := float32(len(results)) / float32(5)
	if relevanceFactor > 1 {
		relevanceFactor = 1
	}

	qualityScore := avgScore * 0.7 + relevanceFactor*0.3

	return qualityScore
}

// updateScores 更新平均分数
func (rs *RAGService) updateScores(score float32) {
	rs.scoresMu.Lock()
	defer rs.scoresMu.Unlock()

	rs.scores = append(rs.scores, score)

	// 保持最近 1000 个分数
	if len(rs.scores) > 1000 {
		rs.scores = rs.scores[1:]
	}

	// 计算平均值
	if len(rs.scores) > 0 {
		var sum float32
		for _, s := range rs.scores {
			sum += s
		}
		rs.avgQualityScore = sum / float32(len(rs.scores))
	}
}

// GetStatistics 获取统计信息
func (rs *RAGService) GetStatistics() map[string]interface{} {
	rs.scoresMu.RLock()
	defer rs.scoresMu.RUnlock()

	return map[string]interface{}{
		"total_rags":        atomic.LoadInt64(&rs.totalRAGs),
		"total_retrievals":  atomic.LoadInt64(&rs.totalRetrievals),
		"avg_quality_score": rs.avgQualityScore,
		"avg_results_per_rag": func() float32 {
			totalRAGs := atomic.LoadInt64(&rs.totalRAGs)
			if totalRAGs == 0 {
				return 0
			}
			return float32(atomic.LoadInt64(&rs.totalRetrievals)) / float32(totalRAGs)
		}(),
	}
}

// SetConfig 设置配置
func (rs *RAGService) SetConfig(config *RAGConfig) {
	if config != nil {
		rs.config = config
		rs.logFunc("info", "RAG config updated")
	}
}

// QueryResult 查询结果（用于与 AI 集成）
type QueryResult struct {
	// 原始查询
	Query string `json:"query"`

	// RAG 增强的提示词
	EnhancedPrompt *EnhancedPrompt `json:"enhanced_prompt"`

	// AI 响应
	Response string `json:"response"`

	// 是否使用了 RAG
	UsedRAG bool `json:"used_rag"`

	// 完成时间
	CompletedAt time.Time `json:"completed_at"`
}

// RAGEnabledChat RAG 增强的聊天接口
type RAGEnabledChat struct {
	// RAG 服务
	ragService *RAGService

	// 互斥锁
	mu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewRAGEnabledChat 创建 RAG 增强聊天
func NewRAGEnabledChat(ragService *RAGService) *RAGEnabledChat {
	return &RAGEnabledChat{
		ragService: ragService,
		logFunc:    defaultLogFuncRet,
	}
}

// ProcessQuery 处理查询
func (rac *RAGEnabledChat) ProcessQuery(ctx context.Context, query string, useRAG bool) (*QueryResult, error) {
	result := &QueryResult{
		Query:       query,
		UsedRAG:     false,
		CompletedAt: time.Now(),
	}

	// 尝试使用 RAG
	if useRAG {
		enhanced, err := rac.ragService.EnhancePrompt(ctx, query)
		if err == nil {
			result.EnhancedPrompt = enhanced
			result.UsedRAG = true

			rac.logFunc("debug", fmt.Sprintf("Query enhanced with RAG: %d results", len(enhanced.SearchResults)))
		} else {
			rac.logFunc("debug", fmt.Sprintf("RAG not used: %v", err))
		}
	}

	return result, nil
}

// GetEnhancedPromptForAI 获取用于 AI 的增强提示词
func (rac *RAGEnabledChat) GetEnhancedPromptForAI(ctx context.Context, query string) (string, error) {
	enhanced, err := rac.ragService.EnhancePrompt(ctx, query)
	if err != nil {
		// 如果 RAG 失败，返回原始查询
		return query, nil
	}

	return enhanced.EnhancedPrompt, nil
}

// VerifyRAGQuality 验证 RAG 质量
type RAGQualityVerifier struct {
	// RAG 服务
	ragService *RAGService

	// 验证阈值
	minQualityScore float32

	// 统计
	passedCount  int64
	failedCount  int64
	lastVerified time.Time
	verifyMu     sync.RWMutex
}

// NewRAGQualityVerifier 创建质量验证器
func NewRAGQualityVerifier(ragService *RAGService, minScore float32) *RAGQualityVerifier {
	return &RAGQualityVerifier{
		ragService:      ragService,
		minQualityScore: minScore,
	}
}

// Verify 验证质量
func (rqv *RAGQualityVerifier) Verify(enhanced *EnhancedPrompt) bool {
	rqv.verifyMu.Lock()
	defer rqv.verifyMu.Unlock()

	passed := enhanced.QualityScore >= rqv.minQualityScore

	if passed {
		atomic.AddInt64(&rqv.passedCount, 1)
	} else {
		atomic.AddInt64(&rqv.failedCount, 1)
	}

	rqv.lastVerified = time.Now()

	return passed
}

// GetStats 获取统计
func (rqv *RAGQualityVerifier) GetStats() map[string]interface{} {
	rqv.verifyMu.RLock()
	defer rqv.verifyMu.RUnlock()

	total := atomic.LoadInt64(&rqv.passedCount) + atomic.LoadInt64(&rqv.failedCount)

	passRate := float32(0)
	if total > 0 {
		passRate = float32(atomic.LoadInt64(&rqv.passedCount)) / float32(total)
	}

	return map[string]interface{}{
		"passed_count": atomic.LoadInt64(&rqv.passedCount),
		"failed_count": atomic.LoadInt64(&rqv.failedCount),
		"total_count":  total,
		"pass_rate":    passRate,
	}
}

