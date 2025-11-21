package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/oblivious/backend/internal/database"
	"github.com/oblivious/backend/internal/model"
	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository() *ChannelRepository {
	return &ChannelRepository{
		db: database.DB,
	}
}

// Create 创建渠道
func (r *ChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

// FindByID 根据 ID 获取渠道
func (r *ChannelRepository) FindByID(ctx context.Context, id int) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.WithContext(ctx).Where("id = ? AND enabled = ? AND deleted_at IS NULL", id, true).First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

// FindByName 根据名称获取渠道
func (r *ChannelRepository) FindByName(ctx context.Context, name string) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.WithContext(ctx).Where("name = ? AND enabled = ? AND deleted_at IS NULL", name, true).First(&channel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

// FindByModel 根据模型名称查找支持该模型的所有启用渠道
func (r *ChannelRepository) FindByModel(ctx context.Context, modelName string) ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.WithContext(ctx).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&channels).
		Error

	if err != nil {
		return nil, err
	}

	// 过滤支持该模型的渠道
	var result []*model.Channel
	for _, ch := range channels {
		if ch.SupportModels == "" {
			// 如果没有指定支持的模型，则支持所有模型
			result = append(result, ch)
			continue
		}

		models := strings.Split(ch.SupportModels, ",")
		for _, m := range models {
			if strings.TrimSpace(m) == modelName {
				result = append(result, ch)
				break
			}
		}
	}

	return result, nil
}

// GetAll 获取所有启用的渠道
func (r *ChannelRepository) GetAll(ctx context.Context) ([]*model.Channel, error) {
	var channels []*model.Channel
	err := r.db.WithContext(ctx).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&channels).
		Error
	return channels, err
}

// Update 更新渠道
func (r *ChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

// Delete 软删除渠道
func (r *ChannelRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&model.Channel{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// ModelPriceRepository 模型价格仓储
type ModelPriceRepository struct {
	db *gorm.DB
}

func NewModelPriceRepository() *ModelPriceRepository {
	return &ModelPriceRepository{
		db: database.DB,
	}
}

// Create 创建模型价格
func (r *ModelPriceRepository) Create(ctx context.Context, price *model.ModelPrice) error {
	return r.db.WithContext(ctx).Create(price).Error
}

// FindByChannelAndModel 根据渠道和模型查找价格
func (r *ModelPriceRepository) FindByChannelAndModel(ctx context.Context, channelID int, modelName string) (*model.ModelPrice, error) {
	var price model.ModelPrice
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND model = ? AND deleted_at IS NULL", channelID, modelName).
		First(&price).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &price, nil
}

// FindByModel 根据模型查找价格（返回最便宜的）
func (r *ModelPriceRepository) FindByModel(ctx context.Context, modelName string) (*model.ModelPrice, error) {
	var price model.ModelPrice
	err := r.db.WithContext(ctx).
		Where("model = ? AND deleted_at IS NULL", modelName).
		Order("input_price ASC, output_price ASC").
		First(&price).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &price, nil
}

// Update 更新模型价格
func (r *ModelPriceRepository) Update(ctx context.Context, price *model.ModelPrice) error {
	return r.db.WithContext(ctx).Save(price).Error
}

// Delete 软删除模型价格
func (r *ModelPriceRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&model.ModelPrice{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

