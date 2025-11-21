package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int            `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	DisplayName  string         `gorm:"size:100" json:"display_name"`
	AvatarURL    string         `gorm:"type:text" json:"avatar_url"`
	Role         int            `gorm:"default:1" json:"role"`
	Quota        int64          `gorm:"default:0" json:"quota"`
	TotalQuota   int64          `gorm:"default:0" json:"total_quota"`
	UsedQuota    int64          `gorm:"default:0" json:"used_quota"`
	InviteCode   string         `gorm:"uniqueIndex;size:20" json:"invite_code"`
	InvitedBy    *int           `json:"invited_by"`
	Status       int            `gorm:"default:1" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"size:50" json:"last_login_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type UserSettings struct {
	UserID       int       `gorm:"primaryKey" json:"user_id"`
	Language     string    `gorm:"size:10;default:zh-CN" json:"language"`
	Theme        string    `gorm:"size:20;default:auto" json:"theme"`
	FontSize     int       `gorm:"default:14" json:"font_size"`
	TTSEnabled   bool      `gorm:"default:false" json:"tts_enabled"`
	TTSVoice     string    `gorm:"size:50" json:"tts_voice"`
	TTSSpeed     float64   `gorm:"default:1.0" json:"tts_speed"`
	STTEnabled   bool      `gorm:"default:false" json:"stt_enabled"`
	SendKey      string    `gorm:"size:20;default:Enter" json:"send_key"`
	CustomConfig string    `gorm:"type:jsonb" json:"custom_config"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}


