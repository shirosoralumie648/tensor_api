package repository

import (
	"context"
	"fmt"

	"github.com/oblivious/backend/internal/database"
	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AgentRepository 处理 Agent 相关的数据库操作
type AgentRepository struct {
	db *gorm.DB
}

// NewAgentRepository 创建新的 Agent Repository
func NewAgentRepository() *AgentRepository {
	return &AgentRepository{
		db: database.DB,
	}
}

// Create 创建新的助手
func (r *AgentRepository) Create(ctx context.Context, agent *model.Agent) error {
	if err := r.db.WithContext(ctx).Create(agent).Error; err != nil {
		logger.Error("Failed to create agent", zap.Error(err))
		return err
	}
	return nil
}

// FindByID 根据 ID 获取助手
func (r *AgentRepository) FindByID(ctx context.Context, id int) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to find agent by ID", zap.Error(err), zap.Int("id", id))
		return nil, err
	}
	return &agent, nil
}

// FindByIdentifier 根据标识符获取助手
func (r *AgentRepository) FindByIdentifier(ctx context.Context, identifier string) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.WithContext(ctx).Where("identifier = ?", identifier).First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Error("Failed to find agent by identifier", zap.Error(err), zap.String("identifier", identifier))
		return nil, err
	}
	return &agent, nil
}

// FindByUserID 获取用户创建的所有助手
func (r *AgentRepository) FindByUserID(ctx context.Context, userID int, page, pageSize int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64

	// 统计总数
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Model(&model.Agent{}).
		Count(&total).Error; err != nil {
		logger.Error("Failed to count agents", zap.Error(err))
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&agents).Error; err != nil {
		logger.Error("Failed to find agents by user ID", zap.Error(err))
		return nil, 0, err
	}

	return agents, total, nil
}

// FindPublicAgents 获取公开的助手列表
func (r *AgentRepository) FindPublicAgents(ctx context.Context, category string, page, pageSize int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64

	query := r.db.WithContext(ctx).Where("is_public = true AND deleted_at IS NULL")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 统计总数
	if err := query.Model(&model.Agent{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count public agents", zap.Error(err))
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("is_featured DESC, likes DESC, views DESC, created_at DESC").
		Find(&agents).Error; err != nil {
		logger.Error("Failed to find public agents", zap.Error(err))
		return nil, 0, err
	}

	return agents, total, nil
}

// FindFeaturedAgents 获取精选助手
func (r *AgentRepository) FindFeaturedAgents(ctx context.Context, limit int) ([]*model.Agent, error) {
	var agents []*model.Agent
	if err := r.db.WithContext(ctx).
		Where("is_public = true AND is_featured = true AND deleted_at IS NULL").
		Order("likes DESC, views DESC, created_at DESC").
		Limit(limit).
		Find(&agents).Error; err != nil {
		logger.Error("Failed to find featured agents", zap.Error(err))
		return nil, err
	}
	return agents, nil
}

// Update 更新助手信息
func (r *AgentRepository) Update(ctx context.Context, agent *model.Agent) error {
	if err := r.db.WithContext(ctx).Save(agent).Error; err != nil {
		logger.Error("Failed to update agent", zap.Error(err))
		return err
	}
	return nil
}

// Delete 软删除助手
func (r *AgentRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Agent{}).Error; err != nil {
		logger.Error("Failed to delete agent", zap.Error(err))
		return err
	}
	return nil
}

// IncrementViews 增加浏览次数
func (r *AgentRepository) IncrementViews(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", id).Update("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment views", zap.Error(err))
		return err
	}
	return nil
}

// IncrementLikes 增加赞数
func (r *AgentRepository) IncrementLikes(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", id).Update("likes", gorm.Expr("likes + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment likes", zap.Error(err))
		return err
	}
	return nil
}

// IncrementForks 增加 Fork 次数
func (r *AgentRepository) IncrementForks(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", id).Update("forks", gorm.Expr("forks + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment forks", zap.Error(err))
		return err
	}
	return nil
}

// CreateFork 创建助手 Fork 记录
func (r *AgentRepository) CreateFork(ctx context.Context, fork *model.AgentFork) error {
	if err := r.db.WithContext(ctx).Create(fork).Error; err != nil {
		logger.Error("Failed to create agent fork", zap.Error(err))
		return err
	}
	return nil
}

// FindForksByOriginalID 根据原始助手 ID 获取所有 Fork 记录
func (r *AgentRepository) FindForksByOriginalID(ctx context.Context, originalID int) ([]*model.AgentFork, error) {
	var forks []*model.AgentFork
	if err := r.db.WithContext(ctx).
		Where("original_id = ? AND deleted_at IS NULL", originalID).
		Order("created_at DESC").
		Find(&forks).Error; err != nil {
		logger.Error("Failed to find agent forks", zap.Error(err))
		return nil, err
	}
	return forks, nil
}

// RecordUsage 记录助手使用情况
func (r *AgentRepository) RecordUsage(ctx context.Context, usage *model.AgentUsage) error {
	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		logger.Error("Failed to record agent usage", zap.Error(err))
		return err
	}
	return nil
}

// GetUsageStats 获取助手使用统计
func (r *AgentRepository) GetUsageStats(ctx context.Context, agentID int) (map[string]interface{}, error) {
	var result struct {
		TotalUsages   int64
		TotalMessages int64
		TotalTokens   int64
		TotalCost     float64
	}

	if err := r.db.WithContext(ctx).
		Table("agent_usages").
		Where("agent_id = ?", agentID).
		Select("COUNT(*) as total_usages, COALESCE(SUM(message_count), 0) as total_messages, COALESCE(SUM(token_count), 0) as total_tokens, COALESCE(SUM(cost), 0) as total_cost").
		Scan(&result).Error; err != nil {
		logger.Error("Failed to get usage stats", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"total_usages":   result.TotalUsages,
		"total_messages": result.TotalMessages,
		"total_tokens":   result.TotalTokens,
		"total_cost":     result.TotalCost,
	}, nil
}

// SearchAgents 搜索助手
func (r *AgentRepository) SearchAgents(ctx context.Context, keyword string, isPublic bool, page, pageSize int) ([]*model.Agent, int64, error) {
	var agents []*model.Agent
	var total int64

	query := r.db.WithContext(ctx)
	if isPublic {
		query = query.Where("is_public = true AND deleted_at IS NULL")
	}
	query = query.Where("name ILIKE ? OR description ILIKE ?", fmt.Sprintf("%%%s%%", keyword), fmt.Sprintf("%%%s%%", keyword))

	// 统计总数
	if err := query.Model(&model.Agent{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count agents by search", zap.Error(err))
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.
		Offset(offset).
		Limit(pageSize).
		Order("likes DESC, views DESC, created_at DESC").
		Find(&agents).Error; err != nil {
		logger.Error("Failed to search agents", zap.Error(err))
		return nil, 0, err
	}

	return agents, total, nil
}

