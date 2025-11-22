package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
)

// ChannelService 渠道服务
type ChannelService struct {
	repo *repository.ChannelRepository
}

// NewChannelService 创建渠道服务
func NewChannelService(repo *repository.ChannelRepository) *ChannelService {
	return &ChannelService{
		repo: repo,
	}
}

// Create 创建渠道
func (s *ChannelService) Create(ctx context.Context, channel *model.Channel) error {
	// 基础验证
	if channel.Name == "" {
		return fmt.Errorf("channel name is required")
	}
	if channel.Type == "" {
		return fmt.Errorf("channel type is required")
	}
	if channel.APIKey == "" {
		return fmt.Errorf("api key is required")
	}

	// 设置默认值
	if channel.BaseURL == "" {
		// 根据类型设置默认 BaseURL
		switch channel.Type {
		case "openai":
			channel.BaseURL = "https://api.openai.com"
		case "anthropic":
			channel.BaseURL = "https://api.anthropic.com"
		}
	}

	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	channel.Status = model.ChannelStatusEnabled

	return s.repo.Create(ctx, channel)
}

// Update 更新渠道
func (s *ChannelService) Update(ctx context.Context, channel *model.Channel) error {
	// 检查是否存在
	existing, err := s.repo.GetByID(ctx, channel.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("channel not found: %d", channel.ID)
	}

	channel.UpdatedAt = time.Now()
	return s.repo.Update(ctx, channel)
}

// Delete 删除渠道
func (s *ChannelService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// GetByID 获取渠道
func (s *ChannelService) GetByID(ctx context.Context, id int) (*model.Channel, error) {
	return s.repo.GetByID(ctx, id)
}

// List 获取渠道列表
func (s *ChannelService) List(ctx context.Context, page, pageSize int) ([]*model.Channel, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}

// Enable 启用渠道
func (s *ChannelService) Enable(ctx context.Context, id int) error {
	channel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if channel == nil {
		return fmt.Errorf("channel not found: %d", id)
	}

	channel.Status = model.ChannelStatusEnabled
	channel.Enabled = true
	channel.UpdatedAt = time.Now()

	return s.repo.Update(ctx, channel)
}

// Disable 禁用渠道
func (s *ChannelService) Disable(ctx context.Context, id int) error {
	channel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if channel == nil {
		return fmt.Errorf("channel not found: %d", id)
	}

	channel.Status = model.ChannelStatusDisabled
	channel.Enabled = false
	channel.UpdatedAt = time.Now()

	return s.repo.Update(ctx, channel)
}

// GetAllEnabled 获取所有启用的渠道
func (s *ChannelService) GetAllEnabled(ctx context.Context) ([]*model.Channel, error) {
	return s.repo.GetAll(ctx)
}
