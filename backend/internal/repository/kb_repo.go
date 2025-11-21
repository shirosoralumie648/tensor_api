package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/oblivious/backend/internal/database"
	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// KnowledgeBaseRepository 处理知识库相关的数据库操作
type KnowledgeBaseRepository struct {
	db *gorm.DB
}

// NewKnowledgeBaseRepository 创建新的 KnowledgeBase Repository
func NewKnowledgeBaseRepository() *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{
		db: database.DB,
	}
}

// CreateKB 创建知识库
func (r *KnowledgeBaseRepository) CreateKB(ctx context.Context, kb *model.KnowledgeBase) error {
	if err := r.db.WithContext(ctx).Create(kb).Error; err != nil {
		logger.Error("Failed to create knowledge base", zap.Error(err))
		return err
	}
	return nil
}

// FindKBByID 根据 ID 获取知识库
func (r *KnowledgeBaseRepository) FindKBByID(ctx context.Context, id int) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&kb).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to find knowledge base", zap.Error(err))
		return nil, err
	}
	return &kb, nil
}

// FindKBsByUserID 获取用户的知识库列表
func (r *KnowledgeBaseRepository) FindKBsByUserID(ctx context.Context, userID int, page, pageSize int) ([]*model.KnowledgeBase, int64, error) {
	var kbs []*model.KnowledgeBase
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID)

	if err := query.Model(&model.KnowledgeBase{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count knowledge bases", zap.Error(err))
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&kbs).Error; err != nil {
		logger.Error("Failed to find knowledge bases", zap.Error(err))
		return nil, 0, err
	}

	return kbs, total, nil
}

// UpdateKB 更新知识库
func (r *KnowledgeBaseRepository) UpdateKB(ctx context.Context, kb *model.KnowledgeBase) error {
	if err := r.db.WithContext(ctx).Save(kb).Error; err != nil {
		logger.Error("Failed to update knowledge base", zap.Error(err))
		return err
	}
	return nil
}

// DeleteKB 软删除知识库
func (r *KnowledgeBaseRepository) DeleteKB(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.KnowledgeBase{}).Error; err != nil {
		logger.Error("Failed to delete knowledge base", zap.Error(err))
		return err
	}
	return nil
}

// CreateDocument 创建文档
func (r *KnowledgeBaseRepository) CreateDocument(ctx context.Context, doc *model.Document) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		logger.Error("Failed to create document", zap.Error(err))
		return err
	}
	return nil
}

// FindDocumentByID 根据 ID 获取文档
func (r *KnowledgeBaseRepository) FindDocumentByID(ctx context.Context, id uuid.UUID) (*model.Document, error) {
	var doc model.Document
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to find document", zap.Error(err))
		return nil, err
	}
	return &doc, nil
}

// FindDocumentsByKBID 获取知识库的文档列表
func (r *KnowledgeBaseRepository) FindDocumentsByKBID(ctx context.Context, kbID int, page, pageSize int) ([]*model.Document, int64, error) {
	var docs []*model.Document
	var total int64

	query := r.db.WithContext(ctx).Where("kb_id = ? AND deleted_at IS NULL", kbID)

	if err := query.Model(&model.Document{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count documents", zap.Error(err))
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		logger.Error("Failed to find documents", zap.Error(err))
		return nil, 0, err
	}

	return docs, total, nil
}

// UpdateDocument 更新文档
func (r *KnowledgeBaseRepository) UpdateDocument(ctx context.Context, doc *model.Document) error {
	if err := r.db.WithContext(ctx).Save(doc).Error; err != nil {
		logger.Error("Failed to update document", zap.Error(err))
		return err
	}
	return nil
}

// DeleteDocument 删除文档
func (r *KnowledgeBaseRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Document{}).Error; err != nil {
		logger.Error("Failed to delete document", zap.Error(err))
		return err
	}
	return nil
}

// CreateChunk 创建文本块
func (r *KnowledgeBaseRepository) CreateChunk(ctx context.Context, chunk *model.DocumentChunk) error {
	if err := r.db.WithContext(ctx).Create(chunk).Error; err != nil {
		logger.Error("Failed to create chunk", zap.Error(err))
		return err
	}
	return nil
}

// CreateChunks 批量创建文本块
func (r *KnowledgeBaseRepository) CreateChunks(ctx context.Context, chunks []*model.DocumentChunk) error {
	if err := r.db.WithContext(ctx).CreateInBatches(chunks, 100).Error; err != nil {
		logger.Error("Failed to create chunks", zap.Error(err))
		return err
	}
	return nil
}

// SearchChunksByVector 根据向量进行相似度搜索
func (r *KnowledgeBaseRepository) SearchChunksByVector(ctx context.Context, kbID int, embedding []float64, limit int) ([]*model.KBSearchResult, error) {
	var results []*model.KBSearchResult

	// 使用 pgvector 的余弦相似度搜索
	query := `
		SELECT 
			c.id as chunk_id,
			c.document_id,
			d.title as document_title,
			c.content,
			1 - (c.embedding <=> $1::vector) as similarity,
			c.metadata
		FROM chunks c
		INNER JOIN documents d ON c.document_id = d.id
		INNER JOIN knowledge_bases kb ON d.kb_id = kb.id
		WHERE kb.id = $2 AND d.deleted_at IS NULL
		ORDER BY c.embedding <=> $1::vector
		LIMIT $3
	`

	if err := r.db.WithContext(ctx).Raw(query, embedding, kbID, limit).Scan(&results).Error; err != nil {
		logger.Error("Failed to search chunks by vector", zap.Error(err))
		return nil, err
	}

	return results, nil
}

// GetChunksByDocumentID 获取文档的所有文本块
func (r *KnowledgeBaseRepository) GetChunksByDocumentID(ctx context.Context, docID uuid.UUID) ([]*model.DocumentChunk, error) {
	var chunks []*model.DocumentChunk

	if err := r.db.WithContext(ctx).
		Where("document_id = ?", docID).
		Order("created_at ASC").
		Find(&chunks).Error; err != nil {
		logger.Error("Failed to get chunks by document ID", zap.Error(err))
		return nil, err
	}

	return chunks, nil
}

// DeleteChunksByDocumentID 删除文档的所有文本块
func (r *KnowledgeBaseRepository) DeleteChunksByDocumentID(ctx context.Context, docID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("document_id = ?", docID).Delete(&model.DocumentChunk{}).Error; err != nil {
		logger.Error("Failed to delete chunks by document ID", zap.Error(err))
		return err
	}
	return nil
}

// IncrementDocumentCount 增加知识库的文档计数
func (r *KnowledgeBaseRepository) IncrementDocumentCount(ctx context.Context, kbID int) error {
	if err := r.db.WithContext(ctx).
		Model(&model.KnowledgeBase{}).
		Where("id = ?", kbID).
		Update("document_count", gorm.Expr("document_count + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment document count", zap.Error(err))
		return err
	}
	return nil
}

// IncrementTotalChunks 增加知识库的文本块计数
func (r *KnowledgeBaseRepository) IncrementTotalChunks(ctx context.Context, kbID int, count int) error {
	if err := r.db.WithContext(ctx).
		Model(&model.KnowledgeBase{}).
		Where("id = ?", kbID).
		Update("total_chunks", gorm.Expr("total_chunks + ?", count)).Error; err != nil {
		logger.Error("Failed to increment total chunks", zap.Error(err))
		return err
	}
	return nil
}
