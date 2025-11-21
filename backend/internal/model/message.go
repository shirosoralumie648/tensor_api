package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	SessionID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"session_id"`
	TopicID      *uuid.UUID `gorm:"type:uuid" json:"topic_id"`
	ParentID     *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	Role         string     `gorm:"size:20;not null" json:"role"` // user, assistant, system, tool
	Content      string     `gorm:"type:text;not null" json:"content"`
	Model        string     `gorm:"size:100" json:"model"`
	InputTokens  int        `gorm:"default:0" json:"input_tokens"`
	OutputTokens int        `gorm:"default:0" json:"output_tokens"`
	TotalTokens  int        `gorm:"default:0" json:"total_tokens"`
	Cost         int64      `gorm:"default:0" json:"cost"` // 花费（分）
	Metadata     string     `gorm:"type:jsonb" json:"metadata"`
	Files        string     `gorm:"type:jsonb" json:"files"`
	ToolCalls    string     `gorm:"type:jsonb" json:"tool_calls"`
	Status       int        `gorm:"default:1" json:"status"` // 1: 正常, 2: 错误, 3: 已删除
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Message) TableName() string {
	return "messages"
}

