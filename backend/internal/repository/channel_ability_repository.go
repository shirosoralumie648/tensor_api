package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// ChannelAbilityRepository 渠道能力仓储接口
type ChannelAbilityRepository interface {
	// FindByModelAndGroup 根据模型和分组查找渠道能力
	FindByModelAndGroup(ctx context.Context, modelName, group string) ([]*model.ChannelAbility, error)

	// FindByChannel 根据渠道ID查找所有能力
	FindByChannel(ctx context.Context, channelID int) ([]*model.ChannelAbility, error)

	// BatchCreate 批量创建能力
	BatchCreate(ctx context.Context, abilities []*model.ChannelAbility) error

	// UpdateByChannel 更新渠道的所有能力
	UpdateByChannel(ctx context.Context, channelID int, abilities []*model.ChannelAbility) error

	// DeleteByChannel 删除渠道的所有能力
	DeleteByChannel(ctx context.Context, channelID int) error

	// FindEnabledByModel 查找指定模型的所有启用渠道
	FindEnabledByModel(ctx context.Context, modelName string) ([]*model.ChannelAbility, error)
}

// DefaultChannelAbilityRepository 默认实现
type DefaultChannelAbilityRepository struct {
	db *gorm.DB
}

// NewChannelAbilityRepository 创建仓储
func NewChannelAbilityRepository(db *gorm.DB) ChannelAbilityRepository {
	return &DefaultChannelAbilityRepository{db: db}
}

// FindByModelAndGroup 根据模型和分组查找渠道能力
func (r *DefaultChannelAbilityRepository) FindByModelAndGroup(ctx context.Context, modelName, group string) ([]*model.ChannelAbility, error) {
	var abilities []*model.ChannelAbility

	query := r.db.WithContext(ctx).
		Where("model = ? AND enabled = ?", modelName, true)

	if group != "" {
		query = query.Where("\"group\" = ?", group)
	}

	// 使用复合索引优化查询
	if err := query.
		Order("priority DESC, weight DESC").
		Find(&abilities).Error; err != nil {
		return nil, fmt.Errorf("failed to find abilities: %w", err)
	}

	return abilities, nil
}

// FindByChannel 根据渠道ID查找所有能力
func (r *DefaultChannelAbilityRepository) FindByChannel(ctx context.Context, channelID int) ([]*model.ChannelAbility, error) {
	var abilities []*model.ChannelAbility

	if err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Find(&abilities).Error; err != nil {
		return nil, fmt.Errorf("failed to find abilities by channel: %w", err)
	}

	return abilities, nil
}

// BatchCreate 批量创建能力
func (r *DefaultChannelAbilityRepository) BatchCreate(ctx context.Context, abilities []*model.ChannelAbility) error {
	if len(abilities) == 0 {
		return nil
	}

	// 使用事务批量插入
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(abilities, 100).Error; err != nil {
			return fmt.Errorf("failed to batch create abilities: %w", err)
		}
		return nil
	})
}

// UpdateByChannel 更新渠道的所有能力
func (r *DefaultChannelAbilityRepository) UpdateByChannel(ctx context.Context, channelID int, abilities []*model.ChannelAbility) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除旧的能力
		if err := tx.Where("channel_id = ?", channelID).Delete(&model.ChannelAbility{}).Error; err != nil {
			return fmt.Errorf("failed to delete old abilities: %w", err)
		}

		// 2. 插入新的能力
		if len(abilities) > 0 {
			if err := tx.CreateInBatches(abilities, 100).Error; err != nil {
				return fmt.Errorf("failed to create new abilities: %w", err)
			}
		}

		return nil
	})
}

// DeleteByChannel 删除渠道的所有能力
func (r *DefaultChannelAbilityRepository) DeleteByChannel(ctx context.Context, channelID int) error {
	if err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Delete(&model.ChannelAbility{}).Error; err != nil {
		return fmt.Errorf("failed to delete abilities: %w", err)
	}

	return nil
}

// FindEnabledByModel 查找指定模型的所有启用渠道
func (r *DefaultChannelAbilityRepository) FindEnabledByModel(ctx context.Context, modelName string) ([]*model.ChannelAbility, error) {
	var abilities []*model.ChannelAbility

	// 联查channel表确保渠道也是启用的
	if err := r.db.WithContext(ctx).
		Joins("JOIN channels ON channels.id = channel_abilities.channel_id").
		Where("channel_abilities.model = ?", modelName).
		Where("channel_abilities.enabled = ?", true).
		Where("channels.enabled = ?", true).
		Where("channels.status = ?", 0). // 0:正常
		Where("channels.deleted_at IS NULL").
		Order("channel_abilities.priority DESC, channel_abilities.weight DESC").
		Find(&abilities).Error; err != nil {
		return nil, fmt.Errorf("failed to find enabled abilities: %w", err)
	}

	return abilities, nil
}
