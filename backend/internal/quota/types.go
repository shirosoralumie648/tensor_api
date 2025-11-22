package quota

import (
	"time"
)

// QuotaType 配额类型
type QuotaType int

const (
	QuotaTypeBalance QuotaType = iota // 余额配额
	QuotaTypeToken                    // Token配额
)

// PreConsumeRequest 预扣费请求
type PreConsumeRequest struct {
	RequestID      string  `json:"request_id"`      // 请求ID（用于幂等性）
	UserID         int     `json:"user_id"`         // 用户ID
	Model          string  `json:"model"`           // 模型名称
	PromptTokens   int     `json:"prompt_tokens"`   // 预估Prompt Tokens
	MaxTokens      int     `json:"max_tokens"`      // 最大生成Tokens
	EstimatedQuota float64 `json:"estimated_quota"` // 预估配额
	TrustThreshold float64 `json:"trust_threshold"` // 信任阈值（余额大于此值不预扣）
}

// PreConsumeResponse 预扣费响应
type PreConsumeResponse struct {
	PreConsumed      bool    `json:"pre_consumed"`       // 是否执行了预扣费
	PreConsumedQuota float64 `json:"pre_consumed_quota"` // 预扣费金额
	RemainingBalance float64 `json:"remaining_balance"`  // 剩余余额
}

// PostConsumeRequest 后扣费请求
type PostConsumeRequest struct {
	RequestID        string  `json:"request_id"`        // 请求ID
	UserID           int     `json:"user_id"`           // 用户ID
	ChannelID        int     `json:"channel_id"`        // 渠道ID
	Model            string  `json:"model"`             // 模型名称
	PromptTokens     int     `json:"prompt_tokens"`     // 实际Prompt Tokens
	CompletionTokens int     `json:"completion_tokens"` // 实际Completion Tokens
	TotalTokens      int     `json:"total_tokens"`      // 总Tokens
	ActualQuota      float64 `json:"actual_quota"`      // 实际配额消耗
	IsStream         bool    `json:"is_stream"`         // 是否流式
	ResponseTime     int64   `json:"response_time"`     // 响应时间（毫秒）
}

// RefundRequest 退款请求
type RefundRequest struct {
	RequestID string  `json:"request_id"` // 请求ID
	UserID    int     `json:"user_id"`    // 用户ID
	Quota     float64 `json:"quota"`      // 退款金额
	Reason    string  `json:"reason"`     // 退款原因
}

// PreConsumedRecord 预扣费记录
type PreConsumedRecord struct {
	RequestID    string    `json:"request_id"`
	UserID       int       `json:"user_id"`
	Quota        float64   `json:"quota"`
	PromptTokens int       `json:"prompt_tokens"`
	MaxTokens    int       `json:"max_tokens"`
	Model        string    `json:"model"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// QuotaService 配额服务接口
type QuotaService interface {
	// PreConsumeQuota 预扣费
	PreConsumeQuota(req *PreConsumeRequest) (*PreConsumeResponse, error)

	// ReturnPreConsumedQuota 归还预扣费（失败时调用）
	ReturnPreConsumedQuota(requestID string, userID int) error

	// PostConsumeQuota 后扣费（实际消费调整）
	PostConsumeQuota(req *PostConsumeRequest) error

	// RefundQuota 退款
	RefundQuota(req *RefundRequest) error

	// GetUserBalance 获取用户余额
	GetUserBalance(userID int) (float64, error)

	// GetPreConsumedRecord 获取预扣费记录
	GetPreConsumedRecord(requestID string) (*PreConsumedRecord, error)
}

// QuotaCache 配额缓存接口
type QuotaCache interface {
	// SetPreConsumed 设置预扣费记录
	SetPreConsumed(record *PreConsumedRecord) error

	// GetPreConsumed 获取预扣费记录
	GetPreConsumed(requestID string) (*PreConsumedRecord, error)

	// DeletePreConsumed 删除预扣费记录
	DeletePreConsumed(requestID string) error

	// GetUserBalance 获取用户余额缓存
	GetUserBalance(userID int) (float64, bool, error)

	// SetUserBalance 设置用户余额缓存
	SetUserBalance(userID int, balance float64) error

	// InvalidateUserBalance 失效用户余额缓存
	InvalidateUserBalance(userID int) error
}

// QuotaCalculator 配额计算器
type QuotaCalculator interface {
	// CalculateQuota 计算配额（基于Token数量和模型定价）
	CalculateQuota(model string, promptTokens, completionTokens int) (float64, error)

	// EstimateMaxQuota 估算最大配额（用于预扣费）
	EstimateMaxQuota(model string, promptTokens, maxTokens int) (float64, error)
}
