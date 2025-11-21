package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Session struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           int            `gorm:"not null;index" json:"user_id"`
	AgentID          *int           `json:"agent_id"`
	GroupID          *uuid.UUID     `gorm:"type:uuid" json:"group_id"`
	Title            string         `gorm:"size:200" json:"title"`
	Description      string         `gorm:"type:text" json:"description"`
	Pinned           bool           `gorm:"default:false" json:"pinned"`
	Archived         bool           `gorm:"default:false" json:"archived"`
	Model            string         `gorm:"size:100" json:"model"`
	Temperature      float64        `gorm:"default:0.7" json:"temperature"`
	TopP             float64        `gorm:"default:1.0" json:"top_p"`
	MaxTokens        *int           `json:"max_tokens"`
	SystemRole       string         `gorm:"type:text" json:"system_role"`
	ContextLength    int            `gorm:"default:4" json:"context_length"`
	PluginIDs        pq.Int64Array  `gorm:"type:int[]" json:"plugin_ids"`
	KnowledgeBaseIDs pq.Int64Array  `gorm:"type:int[]" json:"knowledge_base_ids"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Session) TableName() string {
	return "sessions"
}

