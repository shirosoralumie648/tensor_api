package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
)

// ChannelAbilityService 渠道能力服务接口
type ChannelAbilityService interface {
	// SyncFromChannel 从渠道同步能力（创建/更新渠道时调用）
	SyncFromChannel(ctx context.Context, channel *model.Channel) error

	// FindByModelAndGroup 查找指定模型的渠道能力
	FindByModelAndGroup(ctx context.Context, modelName, group string) ([]*model.ChannelAbility, error)

	// GetAvailableChannelsForModel 获取可用于指定模型的渠道列表
	GetAvailableChannelsForModel(ctx context.Context, modelName string) ([]*model.ChannelAbility, error)

	// DeleteByChannel 删除渠道的所有能力
	DeleteByChannel(ctx context.Context, channelID int) error
}

// DefaultChannelAbilityService 默认实现
type DefaultChannelAbilityService struct {
	db                   *gorm.DB
	abilityRepo          repository.ChannelAbilityRepository
	cache                map[string][]*model.ChannelAbility
	mu                   sync.RWMutex
	cacheTTL             time.Duration
	lastCacheRefreshtime map[string]time.Time
}

// NewChannelAbilityService 创建服务
func NewChannelAbilityService(db *gorm.DB, abilityRepo repository.ChannelAbilityRepository) *DefaultChannelAbilityService {
	return &DefaultChannelAbilityService{
		db:                   db,
		abilityRepo:          abilityRepo,
		cache:                make(map[string][]*model.ChannelAbility),
		cacheTTL:             5 * time.Minute,
		lastCacheRefreshtime: make(map[string]time.Time),
	}
}

// SyncFromChannel 从渠道同步能力
func (s *DefaultChannelAbilityService) SyncFromChannel(ctx context.Context, channel *model.Channel) error {
	// 1. 解析渠道支持的模型列表
	models := channel.GetSupportedModels()
	if len(models) == 0 {
		// 如果没有指定模型，不创建能力记录
		return nil
	}

	// 2. 构建能力列表
	abilities := make([]*model.ChannelAbility, 0, len(models))
	for _, modelName := range models {
		ability := &model.ChannelAbility{
			ChannelID: channel.ID,
			Model:     modelName,
			Group:     channel.Group,
			Priority:  channel.Priority,
			Weight:    channel.Weight,
			Enabled:   channel.Enabled,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		abilities = append(abilities, ability)
	}

	// 3. 更新数据库
	if err := s.abilityRepo.UpdateByChannel(ctx, channel.ID, abilities); err != nil {
		return fmt.Errorf("failed to sync abilities: %w", err)
	}

	// 4. 清空缓存
	s.invalidateCache()

	return nil
}

// FindByModelAndGroup 查找指定模型的渠道能力
func (s *DefaultChannelAbilityService) FindByModelAndGroup(ctx context.Context, modelName, group string) ([]*model.ChannelAbility, error) {
	// 构建缓存key
	cacheKey := fmt.Sprintf("%s_%s", modelName, group)

	// 检查缓存
	s.mu.RLock()
	if abilities, exists := s.cache[cacheKey]; exists {
		if time.Since(s.lastCacheRefreshtime[cacheKey]) < s.cacheTTL {
			s.mu.RUnlock()
			return abilities, nil
		}
	}
	s.mu.RUnlock()

	// 从数据库查询
	abilities, err := s.abilityRepo.FindByModelAndGroup(ctx, modelName, group)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[cacheKey] = abilities
	s.lastCacheRefreshtime[cacheKey] = time.Now()
	s.mu.Unlock()

	return abilities, nil
}

// GetAvailableChannelsForModel 获取可用于指定模型的渠道列表
func (s *DefaultChannelAbilityService) GetAvailableChannelsForModel(ctx context.Context, modelName string) ([]*model.ChannelAbility, error) {
	// 使用缓存
	cacheKey := fmt.Sprintf("enabled_%s", modelName)

	s.mu.RLock()
	if abilities, exists := s.cache[cacheKey]; exists {
		if time.Since(s.lastCacheRefreshtime[cacheKey]) < s.cacheTTL {
			s.mu.RUnlock()
			return abilities, nil
		}
	}
	s.mu.RUnlock()

	// 从数据库查询
	abilities, err := s.abilityRepo.FindEnabledByModel(ctx, modelName)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	s.mu.Lock()
	s.cache[cacheKey] = abilities
	s.lastCacheRefreshtime[cacheKey] = time.Now()
	s.mu.Unlock()

	return abilities, nil
}

// DeleteByChannel 删除渠道的所有能力
func (s *DefaultChannelAbilityService) DeleteByChannel(ctx context.Context, channelID int) error {
	if err := s.abilityRepo.DeleteByChannel(ctx, channelID); err != nil {
		return err
	}

	// 清空缓存
	s.invalidateCache()

	return nil
}

// invalidateCache 清空所有缓存
func (s *DefaultChannelAbilityService) invalidateCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string][]*model.ChannelAbility)
	s.lastCacheRefreshtime = make(map[string]time.Time)
}

// ParseSupportedModels 解析支持的模型列表（辅助函数）
func ParseSupportedModels(supportModels string) []string {
	if supportModels == "" {
		return nil
	}

	// 按逗号分割
	models := strings.Split(supportModels, ",")
	result := make([]string, 0, len(models))

	for _, model := range models {
		model = strings.TrimSpace(model)
		if model != "" {
			result = append(result, model)
		}
	}

	return result
}
