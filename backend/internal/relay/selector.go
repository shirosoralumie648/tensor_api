package relay

import (
	"context"
	"errors"
	"math/rand"

	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/internal/repository"
)

// ChannelSelector 渠道选择器
type ChannelSelector struct {
	channelRepo *repository.ChannelRepository
}

// NewChannelSelector 创建渠道选择器
func NewChannelSelector() *ChannelSelector {
	return &ChannelSelector{
		channelRepo: repository.NewChannelRepository(),
	}
}

// SelectChannel 根据模型选择一个可用渠道
// 实现权重负载均衡：每个渠道的权重越高，被选中的概率越大
func (s *ChannelSelector) SelectChannel(ctx context.Context, modelName string) (*model.Channel, error) {
	// 查询支持该模型的所有可用渠道
	channels, err := s.channelRepo.FindByModel(ctx, modelName)
	if err != nil {
		return nil, err
	}

	if len(channels) == 0 {
		return nil, errors.New("no available channel for model: " + modelName)
	}

	// 如果只有一个渠道，直接返回
	if len(channels) == 1 {
		return channels[0], nil
	}

	// 按权重随机选择
	return selectByWeight(channels), nil
}

// selectByWeight 按权重随机选择一个渠道
func selectByWeight(channels []*model.Channel) *model.Channel {
	// 计算总权重
	totalWeight := 0
	for _, ch := range channels {
		if ch.Weight <= 0 {
			ch.Weight = 1
		}
		totalWeight += ch.Weight
	}

	// 生成随机数
	randWeight := rand.Intn(totalWeight)

	// 按权重选择
	currentWeight := 0
	for _, ch := range channels {
		currentWeight += ch.Weight
		if randWeight < currentWeight {
			return ch
		}
	}

	// 降级：返回第一个渠道
	return channels[0]
}

// SelectChannelByName 根据名称直接选择渠道
func (s *ChannelSelector) SelectChannelByName(ctx context.Context, channelName string) (*model.Channel, error) {
	channel, err := s.channelRepo.FindByName(ctx, channelName)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("channel not found: " + channelName)
	}
	return channel, nil
}

// GetAllChannels 获取所有可用渠道
func (s *ChannelSelector) GetAllChannels(ctx context.Context) ([]*model.Channel, error) {
	return s.channelRepo.GetAll(ctx)
}

