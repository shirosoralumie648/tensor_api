package model

import "time"

// LogType 日志类型
const (
	LogTypeTopup   = 1 // 充值
	LogTypeConsume = 2 // 消费
	LogTypeManage  = 3 // 管理操作
	LogTypeSystem  = 4 // 系统日志
	LogTypeError   = 5 // 错误日志
	LogTypeRefund  = 6 // 退款
)

// UnifiedLog 统一日志
type UnifiedLog struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	UserID           int       `gorm:"not null;index" json:"user_id"`
	Username         string    `gorm:"size:100" json:"username"`
	TokenID          int       `json:"token_id"`
	TokenName        string    `gorm:"size:100" json:"token_name"`
	ChannelID        int       `gorm:"index" json:"channel_id"`
	ChannelName      string    `gorm:"size:100" json:"channel_name"`
	LogType          int       `gorm:"not null;index" json:"log_type"`
	ModelName        string    `gorm:"size:100;index" json:"model_name"`
	Content          string    `gorm:"type:text" json:"content"`
	Quota            int       `json:"quota"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	UseTime          int       `json:"use_time"` // 毫秒
	IsStream         bool      `json:"is_stream"`
	Group            string    `gorm:"size:64" json:"group"`
	IP               string    `gorm:"size:45" json:"ip"`
	UserAgent        string    `gorm:"type:text" json:"user_agent"`
	RequestID        string    `gorm:"size:100;index" json:"request_id"`
	Other            string    `gorm:"type:jsonb" json:"other"`
	CreatedAt        time.Time `gorm:"index" json:"created_at"`
}

func (UnifiedLog) TableName() string {
	return "unified_logs"
}
