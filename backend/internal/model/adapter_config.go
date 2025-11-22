package model

import "time"

// AdapterConfig 适配器配置
type AdapterConfig struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"size:100;uniqueIndex" json:"name"`
	Type            string     `gorm:"size:50" json:"type"`
	Version         string     `gorm:"size:20;default:'v1.0.0'" json:"version"`
	HandlerClass    string     `gorm:"size:200" json:"handler_class"`
	SupportedModels string     `gorm:"type:text" json:"supported_models"`
	DefaultConfig   string     `gorm:"type:jsonb;default:'{}'" json:"default_config"`
	Enabled         bool       `gorm:"default:true" json:"enabled"`
	Description     string     `gorm:"type:text" json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at"`
}

func (AdapterConfig) TableName() string {
	return "adapter_configs"
}
