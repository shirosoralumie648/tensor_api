package model

import "time"

// QuotaType 配额类型
const (
	QuotaTypeByToken = 0 // 按量计费（Token）
	QuotaTypeByCall  = 1 // 按次计费
)

// ModelPricing 模型定价
type ModelPricing struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	Model           string     `gorm:"size:100;not null;index" json:"model"`
	Group           string     `gorm:"size:64;default:'default';index" json:"group"`
	QuotaType       int        `gorm:"default:0" json:"quota_type"`         // 0:按量 1:按次
	ModelPrice      *float64   `json:"model_price"`                         // 按次价格
	ModelRatio      *float64   `json:"model_ratio"`                         // Token倍率
	CompletionRatio float64    `gorm:"default:1.0" json:"completion_ratio"` // 输出Token倍率
	GroupRatio      float64    `gorm:"default:1.0" json:"group_ratio"`      // 分组倍率
	VendorID        string     `gorm:"size:50;index" json:"vendor_id"`      // 供应商ID
	Enabled         bool       `gorm:"default:true;index" json:"enabled"`
	Description     string     `gorm:"type:text" json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at"`
}

func (ModelPricing) TableName() string {
	return "model_pricing"
}

// CalculateQuota 计算配额
func (mp *ModelPricing) CalculateQuota(promptTokens, completionTokens int) int {
	if mp.QuotaType == QuotaTypeByCall {
		// 按次计费
		if mp.ModelPrice != nil {
			return int(*mp.ModelPrice * mp.GroupRatio)
		}
		return 0
	}

	// 按量计费
	if mp.ModelRatio == nil {
		return 0
	}

	quota := float64(promptTokens) * (*mp.ModelRatio)
	quota += float64(completionTokens) * (*mp.ModelRatio) * mp.CompletionRatio
	quota *= mp.GroupRatio

	return int(quota)
}
