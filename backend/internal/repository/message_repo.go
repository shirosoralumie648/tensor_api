package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		db: database.DB,
	}
}

// Create 创建消息
func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// FindByID 根据 ID 查询消息
func (r *MessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// FindBySessionID 根据会话 ID 查询所有消息
func (r *MessageRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, page, pageSize int) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	query := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at ASC")

	// 统计总数
	if err := query.Model(&model.Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetContextMessages 获取最近的 N 条上下文消息
func (r *MessageRepository) GetContextMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*model.Message, error) {
	var messages []*model.Message

	err := r.db.WithContext(ctx).
		Where("session_id = ? AND status = 1", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 反转顺序（从旧到新）
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - 1 - i
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Update 更新消息
func (r *MessageRepository) Update(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// UpdateStatus 更新消息状态
func (r *MessageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status int) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).
		Where("id = ?", id).
		Update("status", status).Error
}

