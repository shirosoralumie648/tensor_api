package selector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/gorm"
)

// ChannelCache 渠道缓存
type ChannelCache struct {
	db              *gorm.DB
	channels        map[string][]*model.Channel // key: model, value: channels
	allChannels     []*model.Channel
	mu              sync.RWMutex
	ttl             time.Duration
	lastRefreshTime time.Time
}

// NewChannelCache 创建渠道缓存
func NewChannelCache(db *gorm.DB, ttl time.Duration) *ChannelCache {
	if ttl == 0 {
		ttl = 5 * time.Minute // 默认5分钟
	}

	cache := &ChannelCache{
		db:       db,
		channels: make(map[string][]*model.Channel),
		ttl:      ttl,
	}

	return cache
}

// GetAvailableChannels 获取可用渠道
func (c *ChannelCache) GetAvailableChannels(ctx context.Context, modelName string) ([]*model.Channel, error) {
	c.mu.RLock()
	needRefresh := time.Since(c.lastRefreshTime) > c.ttl
	c.mu.RUnlock()

	// 如果缓存过期，刷新缓存
	if needRefresh {
		if err := c.Refresh(ctx); err != nil {
			return nil, err
		}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	channels, exists := c.channels[modelName]
	if !exists || len(channels) == 0 {
		return nil, fmt.Errorf("no channels found for model: %s", modelName)
	}

	// 返回副本避免并发修改
	result := make([]*model.Channel, len(channels))
	for i, ch := range channels {
		result[i] = ch
	}

	return result, nil
}

// Refresh 刷新缓存
func (c *ChannelCache) Refresh(ctx context.Context) error {
	// 从数据库加载所有启用的渠道
	var channels []model.Channel
	if err := c.db.WithContext(ctx).
		Where("enabled = ? AND status = ? AND deleted_at IS NULL", true, 0).
		Order("priority DESC, weight DESC").
		Find(&channels).Error; err != nil {
		return fmt.Errorf("failed to load channels: %w", err)
	}

	// 构建新的缓存
	newChannels := make(map[string][]*model.Channel)
	allChannels := make([]*model.Channel, 0, len(channels))

	for i := range channels {
		ch := &channels[i]
		allChannels = append(allChannels, ch)

		// 解析支持的模型列表
		models := ch.GetSupportedModels()
		for _, modelName := range models {
			if _, exists := newChannels[modelName]; !exists {
				newChannels[modelName] = make([]*model.Channel, 0)
			}
			newChannels[modelName] = append(newChannels[modelName], ch)
		}
	}

	// 更新缓存
	c.mu.Lock()
	c.channels = newChannels
	c.allChannels = allChannels
	c.lastRefreshTime = time.Now()
	c.mu.Unlock()

	return nil
}

// GetAllChannels 获取所有渠道
func (c *ChannelCache) GetAllChannels() []*model.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*model.Channel, len(c.allChannels))
	copy(result, c.allChannels)

	return result
}

// InvalidateChannel 使指定渠道的缓存失效
func (c *ChannelCache) InvalidateChannel(channelID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 从所有模型的缓存中移除该渠道
	for modelName, channels := range c.channels {
		filtered := make([]*model.Channel, 0)
		for _, ch := range channels {
			if ch.ID != channelID {
				filtered = append(filtered, ch)
			}
		}
		c.channels[modelName] = filtered
	}

	// 从全部渠道中移除
	filtered := make([]*model.Channel, 0)
	for _, ch := range c.allChannels {
		if ch.ID != channelID {
			filtered = append(filtered, ch)
		}
	}
	c.allChannels = filtered
}

// GetTTL 获取缓存过期时间
func (c *ChannelCache) GetTTL() time.Duration {
	return c.ttl
}

// SetTTL 设置缓存过期时间
func (c *ChannelCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}
