package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// Agent 代表一个 AI 助手
type Agent struct {
	ID           int            `gorm:"primaryKey" json:"id"`
	UserID       *int           `gorm:"index" json:"user_id"`
	Identifier   string         `gorm:"uniqueIndex;size:100" json:"identifier"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	Avatar       string         `gorm:"type:text" json:"avatar"`
	Description  string         `gorm:"type:text" json:"description"`
	Category     string         `gorm:"size:50" json:"category"`
	SystemRole   string         `gorm:"type:text" json:"system_role"`
	Model        string         `gorm:"size:100" json:"model"`
	Temperature  float64        `gorm:"default:0.7" json:"temperature"`
	TopP         float64        `gorm:"default:1.0" json:"top_p"`
	MaxTokens    *int           `json:"max_tokens"`
	Tools        json.RawMessage `gorm:"type:jsonb" json:"tools"`
	PluginIDs    pq.Int64Array  `gorm:"type:integer[]" json:"plugin_ids"`
	KnowledgeBaseIDs pq.Int64Array `gorm:"type:integer[]" json:"knowledge_base_ids"`
	IsPublic     bool           `gorm:"default:false" json:"is_public"`
	IsFeatured   bool           `gorm:"default:false" json:"is_featured"`
	Views        int            `gorm:"default:0" json:"views"`
	Likes        int            `gorm:"default:0" json:"likes"`
	Forks        int            `gorm:"default:0" json:"forks"`
	Status       int            `gorm:"default:1" json:"status"` // 1: 启用, 2: 禁用
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    *time.Time     `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (Agent) TableName() string {
	return "agents"
}

// AgentFork 代表助手的 Fork 记录（当用户复制他人的助手时）
type AgentFork struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	OriginalID   int       `gorm:"index" json:"original_id"`
	UserID       int       `gorm:"index" json:"user_id"`
	ForkName     string    `gorm:"size:100" json:"fork_name"`
	Description  string    `gorm:"type:text" json:"description"`
	IsPublic     bool      `gorm:"default:false" json:"is_public"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at"`
}

// TableName 指定表名
func (AgentFork) TableName() string {
	return "agent_forks"
}

// AgentUsage 记录助手的使用统计
type AgentUsage struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	AgentID      int       `gorm:"index" json:"agent_id"`
	UserID       int       `gorm:"index" json:"user_id"`
	SessionID    string    `json:"session_id"`
	MessageCount int       `json:"message_count"`
	TokenCount   int       `json:"token_count"`
	Cost         float64   `json:"cost"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (AgentUsage) TableName() string {
	return "agent_usages"
}

// AgentToolConfig 表示工具配置（序列化后的 JSON）
type AgentToolConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Value 实现 driver.Valuer 接口
func (a AgentToolConfig) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *AgentToolConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &a)
}

