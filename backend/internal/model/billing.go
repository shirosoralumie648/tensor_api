package model

import (
	"time"

	"github.com/google/uuid"
)

// BillingLog 计费日志
type BillingLog struct {
	ID           int        `gorm:"primaryKey" json:"id"`
	UserID       int        `gorm:"index" json:"user_id"`              // 用户 ID
	SessionID    *uuid.UUID `gorm:"type:uuid;index" json:"session_id"` // 会话 ID（可选）
	MessageID    *uuid.UUID `gorm:"type:uuid;index" json:"message_id"` // 消息 ID（可选）
	Model        string     `gorm:"size:100;index" json:"model"`       // 模型名称
	InputTokens  int        `json:"input_tokens"`                      // 输入 Token 数
	OutputTokens int        `json:"output_tokens"`                     // 输出 Token 数
	TotalTokens  int        `json:"total_tokens"`                      // 总 Token 数
	Cost         int64      `json:"cost"`                              // 费用（分）
	CostUSD      float64    `json:"cost_usd"`                          // 费用（美元）
	Status       int        `gorm:"default:1" json:"status"`           // 状态: 1=已记录 2=已计费 3=已退款
	ErrorMessage string     `gorm:"type:text" json:"error_message"`    // 错误消息
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

func (BillingLog) TableName() string {
	return "billing_logs"
}

// PricingPlan 定价计划
type PricingPlan struct {
	ID          int    `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	// 针对不同模型的定价信息存储在 model_prices 表中
	Active    bool       `gorm:"default:true" json:"active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (PricingPlan) TableName() string {
	return "pricing_plans"
}

// QuotaLog 额度变更日志
type QuotaLog struct {
	ID            int        `gorm:"primaryKey" json:"id"`
	UserID        int        `gorm:"index" json:"user_id"`
	OperationType string     `gorm:"size:50" json:"operation_type"` // recharge, deduct, refund, adjust
	Amount        int64      `json:"amount"`                        // 变更额度（分）
	Reason        string     `gorm:"type:text" json:"reason"`       // 变更原因
	BillingLogID  *int       `gorm:"index" json:"billing_log_id"`   // 关联的计费日志
	BalanceBefore int64      `json:"balance_before"`                // 变更前余额
	BalanceAfter  int64      `json:"balance_after"`                 // 变更后余额
	CreatedAt     time.Time  `json:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

func (QuotaLog) TableName() string {
	return "quota_logs"
}

// Invoice 已在 billing_models.go 中定义
