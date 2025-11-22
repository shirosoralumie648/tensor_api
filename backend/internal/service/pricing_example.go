package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// ExamplePricingService 定价服务使用示例
func ExamplePricingService(db *gorm.DB) {
	ctx := context.Background()

	// 创建定价服务
	pricingService := NewPricingService(db)

	// 1. 列出所有定价
	enabled := true
	pricings, _ := pricingService.ListPricing(ctx, &enabled)
	fmt.Printf("找到 %d 个启用的定价\n", len(pricings))

	// 2. 获取特定模型的定价
	pricing, _ := pricingService.GetPricing(ctx, "gpt-4o", "default")
	if pricing != nil {
		fmt.Printf("GPT-4o定价: 模型倍率=%.2f, 完成倍率=%.2f\n",
			*pricing.ModelRatio, pricing.CompletionRatio)
	}

	// 3. 创建新定价
	ratio := 20.0
	newPricing := &model.ModelPricing{
		Model:           "gpt-4o-pro",
		Group:           "default",
		QuotaType:       model.QuotaTypeByToken,
		ModelRatio:      &ratio,
		CompletionRatio: 2.0,
		GroupRatio:      1.0,
		VendorID:        "openai",
		Enabled:         true,
		Description:     "GPT-4o Pro模型定价",
	}

	if err := pricingService.CreatePricing(ctx, newPricing); err != nil {
		fmt.Printf("创建定价失败: %v\n", err)
	}

	// 4. 计算配额
	quota, _ := pricingService.CalculateQuota(ctx, "gpt-4o", "default", 100, 500)
	fmt.Printf("GPT-4o (100+500 tokens) 配额: %d\n", quota)

	// 5. VIP用户计算配额（享受折扣）
	vipQuota, _ := pricingService.CalculateQuota(ctx, "gpt-4o", "vip", 100, 500)
	fmt.Printf("VIP用户配额: %d (折扣后)\n", vipQuota)

	// 6. 更新定价
	newRatio := 18.0
	updatePricing := &model.ModelPricing{
		ModelRatio: &newRatio,
	}
	pricingService.UpdatePricing(ctx, newPricing.ID, updatePricing)

	// 7. 设置自定义分组倍率
	pricingService.SetGroupRatio("enterprise", 0.5) // 企业用户5折

	// 8. 刷新缓存
	pricingService.RefreshCache(ctx)

	// 9. 获取分组倍率
	ratio1 := pricingService.GetGroupRatio("vip")
	fmt.Printf("VIP用户倍率: %.2f\n", ratio1)
}

// ExamplePricingModes 演示两种计费模式
func ExamplePricingModes(db *gorm.DB) {
	ctx := context.Background()
	pricingService := NewPricingService(db)

	// 模式1: 按量计费（基于Token数量）
	tokenRatio := 15.0
	tokenPricing := &model.ModelPricing{
		Model:           "gpt-4o",
		Group:           "default",
		QuotaType:       model.QuotaTypeByToken, // 按量
		ModelRatio:      &tokenRatio,
		CompletionRatio: 2.0,
		GroupRatio:      1.0,
		Enabled:         true,
	}

	// 计算配额：(100 * 15 + 500 * 15 * 2) * 1.0 = 16500
	quota1, _ := pricingService.CalculateQuota(ctx, "gpt-4o", "default", 100, 500)
	fmt.Printf("按量计费模式配额: %d\n", quota1)

	// 模式2: 按次计费（固定价格）
	callPrice := 10.0
	callPricing := &model.ModelPricing{
		Model:      "dalle-3",
		Group:      "default",
		QuotaType:  model.QuotaTypeByCall, // 按次
		ModelPrice: &callPrice,
		GroupRatio: 1.0,
		Enabled:    true,
	}

	pricingService.CreatePricing(ctx, tokenPricing)
	pricingService.CreatePricing(ctx, callPricing)

	// 按次计费不管token多少，都是固定价格
	quota2, _ := pricingService.CalculateQuota(ctx, "dalle-3", "default", 0, 0)
	fmt.Printf("按次计费模式配额: %d\n", quota2)
}

// ExampleUserGroups 演示用户分组定价
func ExampleUserGroups(db *gorm.DB) {
	ctx := context.Background()
	pricingService := NewPricingService(db)

	model := "gpt-4o"
	promptTokens := 100
	completionTokens := 500

	// 普通用户
	defaultQuota, _ := pricingService.CalculateQuota(ctx, model, "default", promptTokens, completionTokens)

	// VIP用户（8折）
	vipQuota, _ := pricingService.CalculateQuota(ctx, model, "vip", promptTokens, completionTokens)

	// Premium用户（6折）
	premiumQuota, _ := pricingService.CalculateQuota(ctx, model, "premium", promptTokens, completionTokens)

	// 免费用户（1.5倍）
	freeQuota, _ := pricingService.CalculateQuota(ctx, model, "free", promptTokens, completionTokens)

	fmt.Printf("普通用户: %d\n", defaultQuota)
	fmt.Printf("VIP用户: %d (%.0f%%)\n", vipQuota, float64(vipQuota)/float64(defaultQuota)*100)
	fmt.Printf("Premium用户: %d (%.0f%%)\n", premiumQuota, float64(premiumQuota)/float64(defaultQuota)*100)
	fmt.Printf("免费用户: %d (%.0f%%)\n", freeQuota, float64(freeQuota)/float64(defaultQuota)*100)
}
