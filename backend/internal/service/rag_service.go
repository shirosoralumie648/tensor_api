package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
	"github.com/shirosoralumie648/Oblivious/backend/pkg/logger"
	"go.uber.org/zap"
)

// RAGService 处理知识库和 RAG 相关的业务逻辑
type RAGService struct {
	kbRepo        *repository.KnowledgeBaseRepository
	embeddingURL  string // Embedding API URL
	embeddingKey  string // Embedding API Key
	embeddingModel string // 使用的 Embedding 模型
}

// NewRAGService 创建新的 RAG Service
func NewRAGService(embeddingURL, embeddingKey string) *RAGService {
	if embeddingURL == "" {
		embeddingURL = "http://localhost:8083/v1/embeddings" // 默认使用本地中转服务
	}

	return &RAGService{
		kbRepo:         repository.NewKnowledgeBaseRepository(),
		embeddingURL:   embeddingURL,
		embeddingKey:   embeddingKey,
		embeddingModel: "text-embedding-3-small",
	}
}

// CreateKnowledgeBaseRequest 创建知识库的请求
type CreateKnowledgeBaseRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	EmbeddingModel string `json:"embedding_model"`
	ChunkSize      int    `json:"chunk_size"`
	ChunkOverlap   int    `json:"chunk_overlap"`
}

// CreateKnowledgeBase 创建知识库
func (s *RAGService) CreateKnowledgeBase(ctx context.Context, userID int, req *CreateKnowledgeBaseRequest) (*model.KnowledgeBase, error) {
	// 设置默认值
	chunkSize := req.ChunkSize
	if chunkSize == 0 {
		chunkSize = 512
	}
	chunkOverlap := req.ChunkOverlap
	if chunkOverlap == 0 {
		chunkOverlap = 50
	}

	embeddingModel := req.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	kb := &model.KnowledgeBase{
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		EmbeddingModel: embeddingModel,
		ChunkSize:      chunkSize,
		ChunkOverlap:   chunkOverlap,
		Status:         1,
	}

	if err := s.kbRepo.CreateKB(ctx, kb); err != nil {
		return nil, err
	}

	return kb, nil
}

// GetKnowledgeBase 获取知识库详情
func (s *RAGService) GetKnowledgeBase(ctx context.Context, id int, userID int) (*model.KnowledgeBase, error) {
	kb, err := s.kbRepo.FindKBByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if kb == nil {
		return nil, fmt.Errorf("knowledge base not found")
	}

	// 检查权限
	if kb.UserID != userID {
		return nil, fmt.Errorf("permission denied")
	}

	return kb, nil
}

// ListKnowledgeBases 获取用户的知识库列表
func (s *RAGService) ListKnowledgeBases(ctx context.Context, userID int, page, pageSize int) ([]*model.KnowledgeBase, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	kbs, total, err := s.kbRepo.FindKBsByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return kbs, total, nil
}

// DeleteKnowledgeBase 删除知识库
func (s *RAGService) DeleteKnowledgeBase(ctx context.Context, id int, userID int) error {
	kb, err := s.GetKnowledgeBase(ctx, id, userID)
	if err != nil {
		return err
	}

	if kb == nil {
		return fmt.Errorf("knowledge base not found")
	}

	return s.kbRepo.DeleteKB(ctx, id)
}

// GetTextEmbedding 调用 API 获取文本的向量表示
func (s *RAGService) GetTextEmbedding(ctx context.Context, text string) ([]float64, error) {
	// 构建请求
	reqBody := map[string]interface{}{
		"model": s.embeddingModel,
		"input": text,
	}

	jsonBody, _ := json.Marshal(reqBody)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.embeddingURL, bytes.NewReader(jsonBody))
	if err != nil {
		logger.Error("Failed to create embedding request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.embeddingKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.embeddingKey))
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to call embedding API", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error("Embedding API error", 
			zap.Int("status", resp.StatusCode), 
			zap.String("body", string(respBody)))
		return nil, fmt.Errorf("embedding API error: status=%d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error("Failed to parse embedding response", zap.Error(err))
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return result.Data[0].Embedding, nil
}

// ChunkText 将文本分块
func (s *RAGService) ChunkText(text string, chunkSize, chunkOverlap int) []string {
	// 简单的分块实现：按字符数分块
	chunks := []string{}
	runes := []rune(text)
	totalLen := len(runes)

	if chunkSize <= 0 {
		chunkSize = 512
	}
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}

	// 确保 overlap 小于 chunk size
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 2
	}

	for i := 0; i < totalLen; i += chunkSize - chunkOverlap {
		end := i + chunkSize
		if end > totalLen {
			end = totalLen
		}

		chunk := string(runes[i:end])
		chunks = append(chunks, chunk)

		// 最后一个块，停止
		if end == totalLen {
			break
		}
	}

	return chunks
}

// UploadDocument 上传和处理文档
func (s *RAGService) UploadDocument(ctx context.Context, userID int, kbID int, title string, fileContent string) (*model.Document, error) {
	// 获取知识库并检查权限
	kb, err := s.GetKnowledgeBase(ctx, kbID, userID)
	if err != nil {
		return nil, err
	}

	if kb == nil {
		return nil, fmt.Errorf("knowledge base not found")
	}

	// 创建文档记录
	doc := &model.Document{
		ID:              uuid.New(),
		KnowledgeBaseID: kbID,
		Title:           title,
		Status:          model.DocumentStatusPending,
		FileSize:        int64(len(fileContent)),
	}

	if err := s.kbRepo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}

	// 异步处理文档（在生产环境中应该使用消息队列）
	go s.processDocumentAsync(context.Background(), doc.ID, kbID, fileContent, kb)

	return doc, nil
}

// processDocumentAsync 异步处理文档
func (s *RAGService) processDocumentAsync(ctx context.Context, docID uuid.UUID, kbID int, content string, kb *model.KnowledgeBase) {
	// 更新文档状态为处理中
	doc, _ := s.kbRepo.FindDocumentByID(ctx, docID)
	if doc == nil {
		return
	}
	doc.Status = model.DocumentStatusProcessing
	now := time.Now()
	doc.ProcessingStartedAt = &now
	s.kbRepo.UpdateDocument(ctx, doc)

	// 文本分块
	chunks := s.ChunkText(content, kb.ChunkSize, kb.ChunkOverlap)

	// 创建文本块并获取向量表示
	documentChunks := make([]*model.DocumentChunk, 0)

	for i, chunkText := range chunks {
		// 获取向量
		embedding, err := s.GetTextEmbedding(ctx, chunkText)
		if err != nil {
			logger.Error("Failed to get embedding for chunk",
				zap.Error(err),
				zap.Int("chunk_index", i),
				zap.String("doc_id", docID.String()))
			doc.Status = model.DocumentStatusFailed
			doc.ErrorMessage = fmt.Sprintf("Failed to embed chunk %d: %v", i, err)
			s.kbRepo.UpdateDocument(ctx, doc)
			return
		}

		// 转换为 pq.Float64Array
		chunk := &model.DocumentChunk{
			ID:        uuid.New(),
			DocumentID: docID,
			Content:   chunkText,
			Embedding: embedding,
			Metadata:  fmt.Sprintf(`{"chunk_index": %d, "total_chunks": %d}`, i, len(chunks)),
		}

		documentChunks = append(documentChunks, chunk)
	}

	// 批量保存文本块
	if err := s.kbRepo.CreateChunks(ctx, documentChunks); err != nil {
		logger.Error("Failed to create chunks", zap.Error(err))
		doc.Status = model.DocumentStatusFailed
		doc.ErrorMessage = fmt.Sprintf("Failed to save chunks: %v", err)
		s.kbRepo.UpdateDocument(ctx, doc)
		return
	}

	// 更新文档状态和统计
	doc.Status = model.DocumentStatusCompleted
	doc.ChunkCount = len(documentChunks)
	now = time.Now()
	doc.ProcessingCompletedAt = &now
	s.kbRepo.UpdateDocument(ctx, doc)

	// 更新知识库的统计信息
	s.kbRepo.IncrementDocumentCount(ctx, kbID)
	s.kbRepo.IncrementTotalChunks(ctx, kbID, len(documentChunks))

	logger.Info("Document processed successfully",
		zap.String("doc_id", docID.String()),
		zap.Int("chunks", len(documentChunks)))
}

// SearchDocuments 搜索知识库中的文档
func (s *RAGService) SearchDocuments(ctx context.Context, userID int, kbID int, query string, limit int) ([]*model.KBSearchResult, error) {
	// 获取知识库并检查权限
	_, err := s.GetKnowledgeBase(ctx, kbID, userID)
	if err != nil {
		return nil, err
	}

	// 获取查询文本的向量表示
	queryEmbedding, err := s.GetTextEmbedding(ctx, query)
	if err != nil {
		logger.Error("Failed to get query embedding", zap.Error(err))
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// 执行向量相似度搜索
	results, err := s.kbRepo.SearchChunksByVector(ctx, kbID, queryEmbedding, limit)
	if err != nil {
		logger.Error("Failed to search chunks", zap.Error(err))
		return nil, err
	}

	return results, nil
}

// GetDocumentList 获取知识库的文档列表
func (s *RAGService) GetDocumentList(ctx context.Context, userID int, kbID int, page, pageSize int) ([]*model.Document, int64, error) {
	// 检查权限
	_, err := s.GetKnowledgeBase(ctx, kbID, userID)
	if err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	docs, total, err := s.kbRepo.FindDocumentsByKBID(ctx, kbID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// DeleteDocument 删除文档
func (s *RAGService) DeleteDocument(ctx context.Context, userID int, kbID int, docID uuid.UUID) error {
	// 检查权限
	_, err := s.GetKnowledgeBase(ctx, kbID, userID)
	if err != nil {
		return err
	}

	// 删除相关的文本块
	if err := s.kbRepo.DeleteChunksByDocumentID(ctx, docID); err != nil {
		return err
	}

	// 删除文档
	if err := s.kbRepo.DeleteDocument(ctx, docID); err != nil {
		return err
	}

	return nil
}

// BuildRAGContext 为对话构建 RAG 上下文
func (s *RAGService) BuildRAGContext(ctx context.Context, userID int, kbID int, userQuery string, limit int) (string, error) {
	// 搜索相关文档
	results, err := s.SearchDocuments(ctx, userID, kbID, userQuery, limit)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	// 构建上下文
	var contextBuilder strings.Builder
	contextBuilder.WriteString("相关信息：\n\n")

	for i, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("【%s】(来源：%s)\n", result.DocumentTitle, result.DocumentID.String()[:8]))
		contextBuilder.WriteString(result.Content)
		contextBuilder.WriteString("\n\n")

		// 只使用前 limit 个结果
		if i >= limit-1 {
			break
		}
	}

	return contextBuilder.String(), nil
}

