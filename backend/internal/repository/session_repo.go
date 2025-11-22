package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		db: database.DB,
	}
}

// Create 创建会话
func (r *SessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByID 根据 ID 查询会话
func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// FindByUserID 查询用户的所有会话
func (r *SessionRepository) FindByUserID(ctx context.Context, userID int, page, pageSize int) ([]*model.Session, int64, error) {
	var sessions []*model.Session
	var total int64

	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at DESC")

	// 统计总数
	if err := query.Model(&model.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// Update 更新会话
func (r *SessionRepository) Update(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete 软删除会话
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Session{}, "id = ?", id).Error
}

// UpdateTitle 更新会话标题
func (r *SessionRepository) UpdateTitle(ctx context.Context, id uuid.UUID, title string) error {
	return r.db.WithContext(ctx).Model(&model.Session{}).
		Where("id = ?", id).
		Update("title", title).Error
}

