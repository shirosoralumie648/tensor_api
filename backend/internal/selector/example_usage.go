package selector

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ExampleUsage 使用示例
func ExampleUsage(db *gorm.DB) {
	ctx := context.Background()

	// 1. 创建组件
	cache := NewChannelCache(db, 5*time.Minute)
	stats := NewStatsManager()
	selector := NewDefaultChannelSelector(db, cache, stats)

	// 2. 初始化缓存
	if err := cache.Refresh(ctx); err != nil {
		panic(err)
	}

	// 3. 选择渠道 - 使用权重策略
	req := &SelectRequest{
		Model:    "gpt-4o",
		Strategy: StrategyWeight,
		UserID:   1,
	}

	result, err := selector.Select(ctx, req)
	if err != nil {
		panic(err)
	}

	// 使用选中的渠道
	channel := result.Channel
	_ = channel // 进行实际的API调用

	// 4. 更新统计信息
	success := true
	responseTime := 500 * time.Millisecond
	selector.UpdateStats(ctx, channel.ID, success, responseTime)

	// 5. 带重试的选择
	resultWithRetry, err := selector.SelectWithRetry(ctx, req, 3)
	if err != nil {
		panic(err)
	}
	_ = resultWithRetry

	// 6. 标记渠道失败
	if err := selector.MarkChannelFailed(ctx, channel.ID, "timeout"); err != nil {
		panic(err)
	}

	// 7. 刷新缓存（热更新）
	if err := selector.RefreshCache(ctx); err != nil {
		panic(err)
	}

	// 8. 获取统计信息
	channelStats, err := selector.GetStats(ctx, channel.ID)
	if err != nil {
		panic(err)
	}
	_ = channelStats
}
