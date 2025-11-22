package quota

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ExampleUsage 配额服务使用示例
func ExampleUsage(db *gorm.DB, redisClient *redis.Client) {
	// 1. 创建组件
	calculator := NewDefaultQuotaCalculator(db)
	cache := NewRedisQuotaCache(redisClient)
	quotaService := NewDefaultQuotaService(db, cache, calculator)

	// 2. 预扣费
	preReq := &PreConsumeRequest{
		RequestID:      "req-12345",
		UserID:         1,
		Model:          "gpt-4o",
		PromptTokens:   100,
		MaxTokens:      1000,
		EstimatedQuota: 500,
		TrustThreshold: 10000, // 余额>10000不预扣
	}

	preResp, err := quotaService.PreConsumeQuota(preReq)
	if err != nil {
		fmt.Printf("预扣费失败: %v\n", err)
		return
	}

	fmt.Printf("预扣费结果: 已预扣=%v, 金额=%.2f\n",
		preResp.PreConsumed, preResp.PreConsumedQuota)

	// 3. 调用AI API...
	// response, err := callAIAPI(...)

	// 4. 后扣费（根据实际使用）
	postReq := &PostConsumeRequest{
		RequestID:        "req-12345",
		UserID:           1,
		ChannelID:        5,
		Model:            "gpt-4o",
		PromptTokens:     100,
		CompletionTokens: 350, // 实际生成350 tokens
		ActualQuota:      200, // 实际消费200积分
		IsStream:         false,
		ResponseTime:     500,
	}

	if err := quotaService.PostConsumeQuota(postReq); err != nil {
		fmt.Printf("后扣费失败: %v\n", err)
		return
	}

	fmt.Println("后扣费成功")

	// 5. 如果失败，归还预扣费
	// if callFailed {
	//     quotaService.ReturnPreConsumedQuota("req-12345", 1)
	// }
}

// ExampleStreamUsage 流式请求使用示例
func ExampleStreamUsage(db *gorm.DB, redisClient *redis.Client) {
	calculator := NewDefaultQuotaCalculator(db)
	cache := NewRedisQuotaCache(redisClient)
	quotaService := NewDefaultQuotaService(db, cache, calculator)

	// 1. 预扣费（估算最大token）
	preReq := &PreConsumeRequest{
		RequestID:      "stream-req-001",
		UserID:         1,
		Model:          "gpt-4o",
		PromptTokens:   200,
		MaxTokens:      2000,
		EstimatedQuota: 1000,
		TrustThreshold: 5000,
	}

	_, err := quotaService.PreConsumeQuota(preReq)
	if err != nil {
		fmt.Printf("预扣费失败: %v\n", err)
		return
	}

	// 2. 流式调用
	// 使用 tokenizer.StreamTokenCounter 实时累积token
	// ...

	// 3. 流式结束后，根据实际token后扣费
	postReq := &PostConsumeRequest{
		RequestID:        "stream-req-001",
		UserID:           1,
		ChannelID:        5,
		Model:            "gpt-4o",
		PromptTokens:     200,
		CompletionTokens: 850, // 流式实际生成850 tokens
		ActualQuota:      450,
		IsStream:         true,
		ResponseTime:     5000,
	}

	quotaService.PostConsumeQuota(postReq)
}

// ExampleCalculator 配额计算器使用示例
func ExampleCalculator(db *gorm.DB) {
	calc := NewDefaultQuotaCalculator(db)

	// 计算实际配额
	quota, _ := calc.CalculateQuota("gpt-4o", 100, 500)
	fmt.Printf("GPT-4o (100+500 tokens) 配额: %.2f\n", quota)

	// 估算最大配额（用于预扣费）
	maxQuota, _ := calc.EstimateMaxQuota("gpt-4o", 100, 2000)
	fmt.Printf("GPT-4o 最大预估配额: %.2f\n", maxQuota)

	// 切换用户分组（VIP用户可能有不同定价）
	calc.SetGroup("vip")
	vipQuota, _ := calc.CalculateQuota("gpt-4o", 100, 500)
	fmt.Printf("VIP用户配额: %.2f\n", vipQuota)

	// 刷新价格缓存
	calc.RefreshCache()
}
