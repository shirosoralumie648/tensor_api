package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// PricingModel 定义计费方式
type PricingModel string

const (
	PricingPay    PricingModel = "pay_as_you_go" // 按量计费
	PricingUsage  PricingModel = "usage_based"   // 用量基础
	PricingTiered PricingModel = "tiered"        // 分层计费
)

// SubscriptionPlan 订阅计划
type SubscriptionPlan struct {
	ID                 string    `gorm:"primaryKey"`
	Name               string    `gorm:"index"`                                // 计划名称: basic, pro, enterprise
	Description        string                                                   // 计划描述
	MonthlyQuota       int64     `gorm:"index"`                                // 月度 Token 配额
	MonthlyPrice       float32   `gorm:"index"`                                // 月度价格 (USD)
	ExtraTokenPrice    float32   `gorm:"index"`                                // 超出配额后的价格 (每 1000 tokens)
	SupportLevel       string    // 支持级别: basic, priority, enterprise
	Features           JSONArray `gorm:"type:json"`                            // 功能特性列表
	MaxConcurrentReq   int       `gorm:"index"`                                // 最大并发请求数
	RateLimitPerMin    int       `gorm:"index"`                                // 每分钟速率限制
	IsActive           bool      `gorm:"index;default:true"`                   // 是否活跃
	DisplayOrder       int       `gorm:"index"`                                // 显示顺序
	CreatedAt          time.Time `gorm:"autoCreateTime;index"`                 // 创建时间
	UpdatedAt          time.Time `gorm:"autoUpdateTime;index"`                 // 更新时间
}

// Subscription 用户订阅
type Subscription struct {
	ID              string    `gorm:"primaryKey"`
	UserID          string    `gorm:"index;not null"`                         // 用户 ID
	PlanID          string    `gorm:"index;not null"`                         // 订阅计划 ID
	Plan            SubscriptionPlan `gorm:"foreignKey:PlanID;references:ID"` // 关联计划
	Status          string    `gorm:"index;default:active"`                   // 状态: active, cancelled, expired
	BillingCycle    string    `gorm:"index"`                                  // 计费周期: monthly, yearly
	CurrentPeriodStart time.Time `gorm:"index"`                               // 当前周期开始
	CurrentPeriodEnd   time.Time `gorm:"index"`                               // 当前周期结束
	TokensUsed      int64     `gorm:"default:0"`                              // 本周期已使用 Token
	AutoRenew       bool      `gorm:"default:true"`                           // 自动续期
	NextBillingDate *time.Time`gorm:"index"`                                  // 下一个计费日期
	CancelledAt     *time.Time`gorm:"index"`                                  // 取消日期
	CreatedAt       time.Time `gorm:"autoCreateTime;index"`                   // 创建时间
	UpdatedAt       time.Time `gorm:"autoUpdateTime;index"`                   // 更新时间
}

// BillingRecord 计费记录
type BillingRecord struct {
	ID                 string    `gorm:"primaryKey"`
	UserID             string    `gorm:"index;not null"`                      // 用户 ID
	SubscriptionID     string    `gorm:"index"`                               // 订阅 ID
	Model              string    `gorm:"index;not null"`                      // 模型名称
	Provider           string    `gorm:"index"`                               // AI 提供商
	RequestID          string    `gorm:"index;uniqueIndex:idx_request"`       // 请求 ID (防止重复计费)
	PromptTokens       int64     `gorm:"not null"`                            // 输入 Token 数
	CompletionTokens   int64     `gorm:"not null"`                            // 输出 Token 数
	TotalTokens        int64     `gorm:"index;not null"`                      // 总 Token 数
	CostUSD            float32   `gorm:"not null"`                            // 成本 (USD)
	DiscountRate       float32   `gorm:"default:1.0"`                         // 折扣率 (0-1)
	DiscountAmount     float32   `gorm:"default:0"`                           // 折扣金额
	FinalCost          float32   `gorm:"index;not null"`                      // 最终成本
	CouponID           string    `gorm:"index"`                               // 优惠券 ID
	Status             string    `gorm:"index;default:completed"`             // 状态: completed, refunded, pending
	BillingMonth       string    `gorm:"index"`                               // 计费月份 (YYYY-MM)
	APIEndpoint        string                                                   // API 端点
	Metadata           JSONArray `gorm:"type:json"`                           // 额外元数据
	CreatedAt          time.Time `gorm:"autoCreateTime;index:idx_created_date"` // 创建时间
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`                      // 更新时间
}

// Invoice 发票
type Invoice struct {
	ID                 string    `gorm:"primaryKey"`
	UserID             string    `gorm:"index;not null"`                      // 用户 ID
	BillingMonth       string    `gorm:"index;not null"`                      // 计费月份 (YYYY-MM)
	Status             string    `gorm:"index;default:draft"`                 // 状态: draft, issued, sent, paid, overdue, cancelled
	BillingAddress     string                                                   // 账单地址
	Amount             float32   `gorm:"not null"`                            // 总金额
	DiscountAmount     float32   `gorm:"default:0"`                           // 折扣总额
	TaxAmount          float32   `gorm:"default:0"`                           // 税费
	FinalAmount        float32   `gorm:"not null"`                            // 最终金额
	Currency           string    `gorm:"default:USD"`                         // 货币
	IssuedAt           *time.Time`gorm:"index"`                               // 发票签发时间
	DueDate            *time.Time`gorm:"index"`                               // 到期日期
	PaidAt             *time.Time`gorm:"index"`                               // 支付时间
	PaymentMethod      string                                                   // 支付方式: credit_card, wire_transfer, etc.
	TransactionID      string    `gorm:"index"`                               // 交易 ID
	InvoiceNumber      string    `gorm:"index;uniqueIndex"`                   // 发票编号
	Items              JSONArray `gorm:"type:json;not null"`                  // 发票项目
	Notes              string                                                   // 备注
	Metadata           JSONArray `gorm:"type:json"`                           // 额外信息
	CreatedAt          time.Time `gorm:"autoCreateTime;index"`                // 创建时间
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`                      // 更新时间
}

// InvoiceItem 发票项目
type InvoiceItem struct {
	Model            string  `json:"model"`             // 模型名称
	Quantity         int64   `json:"quantity"`          // 数量 (Token 数)
	UnitPrice        float32 `json:"unit_price"`        // 单价 (每 1000 tokens)
	Amount           float32 `json:"amount"`            // 金额
	Description      string  `json:"description"`       // 描述
}

// Coupon 优惠券
type Coupon struct {
	ID                 string    `gorm:"primaryKey"`
	Code               string    `gorm:"index;uniqueIndex;not null"`          // 优惠券代码
	Type               string    `gorm:"index;not null"`                      // 类型: percentage, fixed, free_tokens
	Value              float32   `gorm:"not null"`                            // 值 (百分比或金额)
	MaxUses            int64     `gorm:"not null"`                            // 最大使用次数 (0 为无限)
	UsedCount          int64     `gorm:"default:0;index"`                     // 已使用次数
	MaxUsagePerUser    int       `gorm:"default:1"`                           // 每用户最大使用次数
	MinPurchaseAmount  float32   `gorm:"default:0"`                           // 最小购买金额
	ValidFrom          time.Time `gorm:"index;not null"`                      // 生效时间
	ExpiresAt          time.Time `gorm:"index;not null"`                      // 过期时间
	IsActive           bool      `gorm:"index;default:true"`                  // 是否活跃
	Description        string                                                   // 描述
	CreatedBy          string    `gorm:"index"`                               // 创建者
	ApplicablePlans    JSONArray `gorm:"type:json"`                           // 适用计划列表 (空表示全部)
	Metadata           JSONArray `gorm:"type:json"`                           // 额外信息
	CreatedAt          time.Time `gorm:"autoCreateTime;index"`                // 创建时间
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`                      // 更新时间
}

// CouponUsage 优惠券使用记录
type CouponUsage struct {
	ID                 string    `gorm:"primaryKey"`
	CouponID           string    `gorm:"index;not null"`                      // 优惠券 ID
	UserID             string    `gorm:"index;not null"`                      // 用户 ID
	DiscountAmount     float32   `gorm:"not null"`                            // 折扣金额
	UsedAt             time.Time `gorm:"autoCreateTime;index"`                // 使用时间
}

// ModelPrice 模型价格
type ModelPrice struct {
	ID                 string    `gorm:"primaryKey"`
	ModelID            string    `gorm:"index;uniqueIndex;not null"`          // 模型 ID
	Provider           string    `gorm:"index;not null"`                      // 提供商
	PromptPricePerK    float32   `gorm:"not null"`                            // 每 1000 输入 tokens 的价格
	CompletionPricePerK float32  `gorm:"not null"`                            // 每 1000 输出 tokens 的价格
	EffectiveFrom      time.Time `gorm:"index"`                               // 生效日期
	EffectiveTo        *time.Time`gorm:"index"`                               // 失效日期
	IsActive           bool      `gorm:"index;default:true"`                  // 是否活跃
	UpdatedBy          string                                                   // 更新者
	Metadata           JSONArray `gorm:"type:json"`                           // 额外信息
	CreatedAt          time.Time `gorm:"autoCreateTime;index"`                // 创建时间
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`                      // 更新时间
}

// BillingSettings 用户计费设置
type BillingSettings struct {
	ID                 string    `gorm:"primaryKey"`
	UserID             string    `gorm:"index;uniqueIndex;not null"`          // 用户 ID
	BillingEmail       string    `gorm:"index"`                               // 计费邮箱
	BillingAddress     string                                                   // 计费地址
	TaxID              string                                                   // 税号
	Currency           string    `gorm:"default:USD"`                         // 偏好货币
	AutoTopup          bool      `gorm:"default:true"`                        // 自动充值
	AutoTopupThreshold int64     `gorm:"default:0"`                           // 自动充值触发门槛 (剩余 token 数)
	AutoTopupAmount    float32   `gorm:"default:100"`                         // 自动充值金额 (USD)
	InvoiceFrequency   string    `gorm:"default:monthly"`                     // 发票频率: monthly, quarterly, annually
	SendInvoiceEmail   bool      `gorm:"default:true"`                        // 是否发送发票邮件
	AlertThreshold     int64     `gorm:"default:0"`                           // 告警阈值 (剩余 token 数)
	EnableAlerts       bool      `gorm:"default:true"`                        // 是否启用告警
	Metadata           JSONArray `gorm:"type:json"`                           // 额外信息
	CreatedAt          time.Time `gorm:"autoCreateTime;index"`                // 创建时间
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`                      // 更新时间
}

// JSONArray 用于存储 JSON 数组
type JSONArray []interface{}

// Value 实现 driver.Valuer 接口
func (ja JSONArray) Value() (driver.Value, error) {
	return json.Marshal(ja)
}

// Scan 实现 sql.Scanner 接口
func (ja *JSONArray) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &ja)
}

