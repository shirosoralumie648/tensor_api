package model

import "time"

// Channel 代表一个 AI 服务渠道（例如 OpenAI、Claude、Gemini 等）
type Channel struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;uniqueIndex" json:"name"`           // 渠道名称（e.g., "openai", "claude", "gemini"）
	Type          string    `gorm:"size:50" json:"type"`                        // 渠道类型
	APIKey        string    `gorm:"size:500" json:"api_key"`                    // API 密钥（加密存储）
	BaseURL       string    `gorm:"size:500" json:"base_url"`                   // 基础 URL
	Weight        int       `gorm:"default:1" json:"weight"`                    // 权重（用于负载均衡）
	MaxRateLimit  int       `json:"max_rate_limit"`                             // 最大请求速率（请求/分钟）
	ModelMapping  string    `gorm:"type:jsonb" json:"model_mapping"`            // 模型映射（JSON格式）
	SupportModels string    `gorm:"type:text" json:"support_models"`            // 支持的模型列表（逗号分隔）
	Enabled       bool      `gorm:"default:true" json:"enabled"`                // 是否启用
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at"`
}

func (Channel) TableName() string {
	return "channels"
}

// ModelPrice 代表不同模型的价格配置
type ModelPrice struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	ChannelID    int       `gorm:"index" json:"channel_id"`
	Model        string    `gorm:"size:100;index" json:"model"`
	InputPrice   float64   `json:"input_price"`    // 输入价格（每 1000 tokens）
	OutputPrice  float64   `json:"output_price"`   // 输出价格（每 1000 tokens）
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (ModelPrice) TableName() string {
	return "model_prices"
}

