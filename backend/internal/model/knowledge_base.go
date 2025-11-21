package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// KnowledgeBase 代表一个知识库
type KnowledgeBase struct {
	ID              int       `gorm:"primaryKey" json:"id"`
	UserID          int       `gorm:"index" json:"user_id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	Description     string    `gorm:"type:text" json:"description"`
	EmbeddingModel  string    `gorm:"size:100;default:text-embedding-3-small" json:"embedding_model"`
	ChunkSize       int       `gorm:"default:512" json:"chunk_size"`
	ChunkOverlap    int       `gorm:"default:50" json:"chunk_overlap"`
	DocumentCount   int       `gorm:"default:0" json:"document_count"`
	TotalChunks     int       `gorm:"default:0" json:"total_chunks"`
	Status          int       `gorm:"default:1" json:"status"` // 1: 启用, 2: 禁用
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// Document 代表知识库中的一个文档
type Document struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	KnowledgeBaseID     int       `gorm:"index" json:"kb_id"`
	Title               string    `gorm:"size:200;not null" json:"title"`
	FileURL             string    `gorm:"type:text" json:"file_url"`
	FileType            string    `gorm:"size:50" json:"file_type"` // pdf, txt, markdown, docx
	FileSize            int64     `json:"file_size"`
	Status              int       `gorm:"default:1" json:"status"` // 1: 待处理, 2: 处理中, 3: 完成, 4: 失败
	ChunkCount          int       `gorm:"default:0" json:"chunk_count"`
	ErrorMessage        string    `gorm:"type:text" json:"error_message"`
	ProcessingStartedAt *time.Time `json:"processing_started_at"`
	ProcessingCompletedAt *time.Time `json:"processing_completed_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	DeletedAt           *time.Time `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// DocumentChunk 代表文档的一个文本块
type DocumentChunk struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentID  uuid.UUID     `gorm:"type:uuid;index" json:"document_id"`
	Content     string        `gorm:"type:text;not null" json:"content"`
	Embedding   pq.Float64Array `gorm:"type:vector(1536)" json:"embedding"` // OpenAI embedding 维度
	Metadata    string        `gorm:"type:jsonb" json:"metadata"`            // 存储位置、页码等信息
	CreatedAt   time.Time     `json:"created_at"`
}

// TableName 指定表名
func (DocumentChunk) TableName() string {
	return "chunks"
}

// KBSearchResult 知识库搜索结果
type KBSearchResult struct {
	ChunkID     uuid.UUID `json:"chunk_id"`
	DocumentID  uuid.UUID `json:"document_id"`
	DocumentTitle string   `json:"document_title"`
	Content     string    `json:"content"`
	Similarity  float64   `json:"similarity"`
	Metadata    string    `json:"metadata"`
}

// ProcessingStatus 处理状态常量
const (
	DocumentStatusPending    = 1
	DocumentStatusProcessing = 2
	DocumentStatusCompleted  = 3
	DocumentStatusFailed     = 4
)

// FileTypeConstants 文件类型常量
const (
	FileTypePDF      = "pdf"
	FileTypeTXT      = "txt"
	FileTypeMarkdown = "markdown"
	FileTypeDocx     = "docx"
)

