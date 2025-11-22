package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/repository"
	"github.com/shirosoralumie648/Oblivious/backend/pkg/logger"
	"go.uber.org/zap"
)

// AgentService 处理助手相关的业务逻辑
type AgentService struct {
	agentRepo *repository.AgentRepository
}

// NewAgentService 创建新的 Agent Service
func NewAgentService() *AgentService {
	return &AgentService{
		agentRepo: repository.NewAgentRepository(),
	}
}

// CreateAgentRequest 创建助手的请求结构
type CreateAgentRequest struct {
	Name         string   `json:"name" binding:"required"`
	Avatar       string   `json:"avatar"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	SystemRole   string   `json:"system_role" binding:"required"`
	Model        string   `json:"model" binding:"required"`
	Temperature  float64  `json:"temperature"`
	TopP         float64  `json:"top_p"`
	MaxTokens    *int     `json:"max_tokens"`
	PluginIDs    []int64  `json:"plugin_ids"`
	KnowledgeBaseIDs []int64 `json:"knowledge_base_ids"`
	IsPublic     bool     `json:"is_public"`
}

// UpdateAgentRequest 更新助手的请求结构
type UpdateAgentRequest struct {
	Name         string   `json:"name"`
	Avatar       string   `json:"avatar"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	SystemRole   string   `json:"system_role"`
	Model        string   `json:"model"`
	Temperature  float64  `json:"temperature"`
	TopP         float64  `json:"top_p"`
	MaxTokens    *int     `json:"max_tokens"`
	PluginIDs    []int64  `json:"plugin_ids"`
	KnowledgeBaseIDs []int64 `json:"knowledge_base_ids"`
	IsPublic     bool     `json:"is_public"`
}

// CreateAgent 创建新的助手
func (s *AgentService) CreateAgent(ctx context.Context, userID int, req *CreateAgentRequest) (*model.Agent, error) {
	// 验证输入
	if req.Temperature < 0 || req.Temperature > 2 {
		return nil, fmt.Errorf("temperature must be between 0 and 2")
	}

	// 生成唯一的标识符
	identifier := generateIdentifier(req.Name)

	agent := &model.Agent{
		UserID:       &userID,
		Identifier:   identifier,
		Name:         req.Name,
		Avatar:       req.Avatar,
		Description:  req.Description,
		Category:     req.Category,
		SystemRole:   req.SystemRole,
		Model:        req.Model,
		Temperature:  req.Temperature,
		TopP:         req.TopP,
		MaxTokens:    req.MaxTokens,
		IsPublic:     req.IsPublic,
		Status:       1, // 默认启用
	}

	if err := s.agentRepo.Create(ctx, agent); err != nil {
		logger.Error("Failed to create agent", zap.Error(err))
		return nil, err
	}

	return agent, nil
}

// GetAgentByID 根据 ID 获取助手详情
func (s *AgentService) GetAgentByID(ctx context.Context, id int) (*model.Agent, error) {
	agent, err := s.agentRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get agent", zap.Error(err))
		return nil, err
	}

	if agent == nil {
		return nil, fmt.Errorf("agent not found")
	}

	// 增加浏览量
	_ = s.agentRepo.IncrementViews(ctx, id)

	return agent, nil
}

// GetAgentByIdentifier 根据标识符获取助手
func (s *AgentService) GetAgentByIdentifier(ctx context.Context, identifier string) (*model.Agent, error) {
	agent, err := s.agentRepo.FindByIdentifier(ctx, identifier)
	if err != nil {
		logger.Error("Failed to get agent by identifier", zap.Error(err))
		return nil, err
	}

	if agent == nil {
		return nil, fmt.Errorf("agent not found")
	}

	return agent, nil
}

// GetUserAgents 获取用户创建的所有助手
func (s *AgentService) GetUserAgents(ctx context.Context, userID int, page, pageSize int) ([]*model.Agent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	agents, total, err := s.agentRepo.FindByUserID(ctx, userID, page, pageSize)
	if err != nil {
		logger.Error("Failed to get user agents", zap.Error(err))
		return nil, 0, err
	}

	return agents, total, nil
}

// GetPublicAgents 获取公开的助手列表
func (s *AgentService) GetPublicAgents(ctx context.Context, category string, page, pageSize int) ([]*model.Agent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	agents, total, err := s.agentRepo.FindPublicAgents(ctx, category, page, pageSize)
	if err != nil {
		logger.Error("Failed to get public agents", zap.Error(err))
		return nil, 0, err
	}

	return agents, total, nil
}

// GetFeaturedAgents 获取精选助手
func (s *AgentService) GetFeaturedAgents(ctx context.Context, limit int) ([]*model.Agent, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	agents, err := s.agentRepo.FindFeaturedAgents(ctx, limit)
	if err != nil {
		logger.Error("Failed to get featured agents", zap.Error(err))
		return nil, err
	}

	return agents, nil
}

// UpdateAgent 更新助手信息
func (s *AgentService) UpdateAgent(ctx context.Context, userID int, id int, req *UpdateAgentRequest) (*model.Agent, error) {
	agent, err := s.agentRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find agent", zap.Error(err))
		return nil, err
	}

	if agent == nil {
		return nil, fmt.Errorf("agent not found")
	}

	// 检查权限
	if agent.UserID == nil || *agent.UserID != userID {
		return nil, fmt.Errorf("permission denied")
	}

	// 更新字段
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Avatar != "" {
		agent.Avatar = req.Avatar
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Category != "" {
		agent.Category = req.Category
	}
	if req.SystemRole != "" {
		agent.SystemRole = req.SystemRole
	}
	if req.Model != "" {
		agent.Model = req.Model
	}
	if req.Temperature > 0 {
		agent.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		agent.TopP = req.TopP
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = req.MaxTokens
	}

	agent.IsPublic = req.IsPublic

	if err := s.agentRepo.Update(ctx, agent); err != nil {
		logger.Error("Failed to update agent", zap.Error(err))
		return nil, err
	}

	return agent, nil
}

// DeleteAgent 删除助手
func (s *AgentService) DeleteAgent(ctx context.Context, userID int, id int) error {
	agent, err := s.agentRepo.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find agent", zap.Error(err))
		return err
	}

	if agent == nil {
		return fmt.Errorf("agent not found")
	}

	// 检查权限
	if agent.UserID == nil || *agent.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	if err := s.agentRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete agent", zap.Error(err))
		return err
	}

	return nil
}

// SearchAgents 搜索助手
func (s *AgentService) SearchAgents(ctx context.Context, keyword string, page, pageSize int) ([]*model.Agent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	agents, total, err := s.agentRepo.SearchAgents(ctx, keyword, true, page, pageSize)
	if err != nil {
		logger.Error("Failed to search agents", zap.Error(err))
		return nil, 0, err
	}

	return agents, total, nil
}

// LikeAgent 赞助手
func (s *AgentService) LikeAgent(ctx context.Context, id int) error {
	if err := s.agentRepo.IncrementLikes(ctx, id); err != nil {
		logger.Error("Failed to like agent", zap.Error(err))
		return err
	}
	return nil
}

// ForkAgent 复制助手
func (s *AgentService) ForkAgent(ctx context.Context, userID int, originalID int, forkName string) (*model.Agent, error) {
	// 获取原始助手
	originalAgent, err := s.agentRepo.FindByID(ctx, originalID)
	if err != nil {
		logger.Error("Failed to find original agent", zap.Error(err))
		return nil, err
	}

	if originalAgent == nil {
		return nil, fmt.Errorf("original agent not found")
	}

	// 创建新的助手（Fork）
	identifier := generateIdentifier(forkName)
	newAgent := &model.Agent{
		UserID:       &userID,
		Identifier:   identifier,
		Name:         forkName,
		Avatar:       originalAgent.Avatar,
		Description:  fmt.Sprintf("Fork of %s", originalAgent.Name),
		Category:     originalAgent.Category,
		SystemRole:   originalAgent.SystemRole,
		Model:        originalAgent.Model,
		Temperature:  originalAgent.Temperature,
		TopP:         originalAgent.TopP,
		MaxTokens:    originalAgent.MaxTokens,
		IsPublic:     false, // Fork 默认为私有
		Status:       1,
	}

	if err := s.agentRepo.Create(ctx, newAgent); err != nil {
		logger.Error("Failed to fork agent", zap.Error(err))
		return nil, err
	}

	// 记录 Fork
	fork := &model.AgentFork{
		OriginalID: originalID,
		UserID:     userID,
		ForkName:   forkName,
	}
	if err := s.agentRepo.CreateFork(ctx, fork); err != nil {
		logger.Warn("Failed to record fork", zap.Error(err))
		// 不影响主流程
	}

	// 增加原助手的 Fork 计数
	_ = s.agentRepo.IncrementForks(ctx, originalID)

	return newAgent, nil
}

// RecordAgentUsage 记录助手使用情况
func (s *AgentService) RecordAgentUsage(ctx context.Context, agentID int, userID int, sessionID string, messageCount, tokenCount int, cost float64) error {
	usage := &model.AgentUsage{
		AgentID:      agentID,
		UserID:       userID,
		SessionID:    sessionID,
		MessageCount: messageCount,
		TokenCount:   tokenCount,
		Cost:         cost,
	}

	if err := s.agentRepo.RecordUsage(ctx, usage); err != nil {
		logger.Error("Failed to record agent usage", zap.Error(err))
		return err
	}

	return nil
}

// GetUsageStats 获取助手使用统计
func (s *AgentService) GetUsageStats(ctx context.Context, agentID int) (map[string]interface{}, error) {
	stats, err := s.agentRepo.GetUsageStats(ctx, agentID)
	if err != nil {
		logger.Error("Failed to get usage stats", zap.Error(err))
		return nil, err
	}

	return stats, nil
}

// generateIdentifier 生成唯一的标识符
func generateIdentifier(name string) string {
	// 清理名称：转小写、只保留字母数字和下划线
	identifier := strings.ToLower(name)
	re := regexp.MustCompile("[^a-z0-9_-]")
	identifier = re.ReplaceAllString(identifier, "-")
	identifier = strings.Trim(identifier, "-")

	// 添加时间戳以确保唯一性
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", identifier, timestamp%1000)
}

