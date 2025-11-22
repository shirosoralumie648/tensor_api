package selector

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

var roundRobinIndex int64

// selectByWeight 权重选择策略
func (s *DefaultChannelSelector) selectByWeight(ctx context.Context, channels []*model.Channel, req *SelectRequest) (*model.Channel, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels available")
	}

	// 计算总权重
	totalWeight := 0
	for _, ch := range channels {
		if ch.Weight > 0 {
			totalWeight += ch.Weight
		}
	}

	if totalWeight == 0 {
		// 如果没有权重，随机选择
		return channels[rand.Intn(len(channels))], nil
	}

	// 根据权重随机选择
	r := rand.Intn(totalWeight)
	sum := 0

	for _, ch := range channels {
		if ch.Weight <= 0 {
			continue
		}

		sum += ch.Weight
		if r < sum {
			return ch, nil
		}
	}

	// 理论上不会到这里，保险起见返回最后一个
	return channels[len(channels)-1], nil
}

// selectByPriority 优先级选择策略
func (s *DefaultChannelSelector) selectByPriority(ctx context.Context, channels []*model.Channel, req *SelectRequest) (*model.Channel, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels available")
	}

	// 找到最高优先级
	maxPriority := channels[0].Priority
	for _, ch := range channels {
		if ch.Priority > maxPriority {
			maxPriority = ch.Priority
		}
	}

	// 收集所有最高优先级的渠道
	topChannels := make([]*model.Channel, 0)
	for _, ch := range channels {
		if ch.Priority == maxPriority {
			topChannels = append(topChannels, ch)
		}
	}

	// 如果只有一个最高优先级渠道，直接返回
	if len(topChannels) == 1 {
		return topChannels[0], nil
	}

	// 如果有多个，使用权重选择
	return s.selectByWeight(ctx, topChannels, req)
}

// selectByRoundRobin 轮询选择策略
func (s *DefaultChannelSelector) selectByRoundRobin(ctx context.Context, channels []*model.Channel, req *SelectRequest) (*model.Channel, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels available")
	}

	// 原子递增索引
	index := atomic.AddInt64(&roundRobinIndex, 1)
	selectedIndex := int(index) % len(channels)

	return channels[selectedIndex], nil
}

// selectByLowestLatency 最低延迟选择策略
func (s *DefaultChannelSelector) selectByLowestLatency(ctx context.Context, channels []*model.Channel, req *SelectRequest) (*model.Channel, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels available")
	}

	// 获取所有渠道的统计信息
	var bestChannel *model.Channel
	var bestLatency int64 = -1

	for _, ch := range channels {
		stats, err := s.stats.GetStats(ctx, ch.ID)
		if err != nil {
			continue
		}

		// 如果没有统计数据，使用 ResponseTime 字段
		latency := int64(ch.ResponseTime)
		if stats.TotalRequests > 0 && stats.AvgResponseTime > 0 {
			latency = stats.AvgResponseTime.Milliseconds()
		}

		// 第一个渠道或者找到更低延迟的渠道
		if bestLatency == -1 || (latency > 0 && latency < bestLatency) {
			bestChannel = ch
			bestLatency = latency
		}
	}

	// 如果所有渠道都没有延迟数据，退化为权重选择
	if bestChannel == nil {
		return s.selectByWeight(ctx, channels, req)
	}

	return bestChannel, nil
}

// selectByRandom 随机选择策略
func (s *DefaultChannelSelector) selectByRandom(ctx context.Context, channels []*model.Channel, req *SelectRequest) (*model.Channel, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels available")
	}

	// 纯随机选择
	index := rand.Intn(len(channels))
	return channels[index], nil
}
