package model

import "time"

// ChannelAbility 渠道能力（渠道支持的模型）
type ChannelAbility struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	ChannelID int       `gorm:"not null;index" json:"channel_id"`
	Model     string    `gorm:"size:100;not null;index" json:"model"`
	Group     string    `gorm:"size:64;default:'default';index" json:"group"`
	Enabled   bool      `gorm:"default:true;index" json:"enabled"`
	Priority  int64     `gorm:"default:0" json:"priority"`
	Weight    int       `gorm:"default:1" json:"weight"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Channel *Channel `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
}

func (ChannelAbility) TableName() string {
	return "channel_abilities"
}
