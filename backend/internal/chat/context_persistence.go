package chat

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MessageSnapshot 消息快照
type MessageSnapshot struct {
	// 消息 ID
	ID string `json:"id"`

	// 角色
	Role string `json:"role"`

	// 内容
	Content string `json:"content"`

	// Token 数
	Tokens int64 `json:"tokens"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`

	// 序列号
	SequenceNumber int64 `json:"sequence_number"`
}

// ContextVersion 上下文版本
type ContextVersion struct {
	// 版本号
	VersionNumber int64 `json:"version_number"`

	// 会话 ID
	SessionID string `json:"session_id"`

	// 快照消息列表
	Messages []*MessageSnapshot `json:"messages"`

	// 总 token 数
	TotalTokens int64 `json:"total_tokens"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 备注
	Remark string `json:"remark"`
}

// ContextSnapshot 上下文快照
type ContextSnapshot struct {
	// 会话 ID
	SessionID string `json:"session_id"`

	// 消息快照
	Messages []*MessageSnapshot `json:"messages"`

	// 总 token 数
	TotalTokens int64 `json:"total_tokens"`

	// 当前轮数
	CurrentRound int `json:"current_round"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 快照时间
	SnapshotAt time.Time `json:"snapshot_at"`
}

// ContextPersistence 上下文持久化
type ContextPersistence struct {
	// 会话 ID
	SessionID string

	// 版本列表
	versions []*ContextVersion
	versionsMu sync.RWMutex

	// 当前版本号
	currentVersion int64

	// 版本历史大小限制
	maxVersions int

	// 统计信息
	totalSnapshots int64
	totalRestores  int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewContextPersistence 创建上下文持久化
func NewContextPersistence(sessionID string, maxVersions int) *ContextPersistence {
	return &ContextPersistence{
		SessionID:   sessionID,
		versions:    make([]*ContextVersion, 0),
		maxVersions: maxVersions,
		logFunc:     defaultLogFunc,
	}
}

// SaveVersion 保存版本
func (cp *ContextPersistence) SaveVersion(messages []*Message, totalTokens int64, remark string) error {
	cp.versionsMu.Lock()
	defer cp.versionsMu.Unlock()

	cp.currentVersion++

	msgSnapshots := make([]*MessageSnapshot, len(messages))
	for i, msg := range messages {
		msgSnapshots[i] = &MessageSnapshot{
			ID:             msg.ID,
			Role:           msg.Role,
			Content:        msg.Content,
			Tokens:         msg.Tokens,
			Timestamp:      msg.Timestamp,
			SequenceNumber: int64(i),
		}
	}

	version := &ContextVersion{
		VersionNumber: cp.currentVersion,
		SessionID:     cp.SessionID,
		Messages:      msgSnapshots,
		TotalTokens:   totalTokens,
		CreatedAt:     time.Now(),
		Remark:        remark,
	}

	cp.versions = append(cp.versions, version)

	// 限制版本数量
	if len(cp.versions) > cp.maxVersions {
		cp.versions = cp.versions[len(cp.versions)-cp.maxVersions:]
	}

	atomic.AddInt64(&cp.totalSnapshots, 1)

	cp.logFunc("info", fmt.Sprintf("Saved version %d for session %s", cp.currentVersion, cp.SessionID))

	return nil
}

// GetVersion 获取指定版本
func (cp *ContextPersistence) GetVersion(versionNumber int64) (*ContextVersion, error) {
	cp.versionsMu.RLock()
	defer cp.versionsMu.RUnlock()

	for _, version := range cp.versions {
		if version.VersionNumber == versionNumber {
			return version, nil
		}
	}

	return nil, fmt.Errorf("version %d not found", versionNumber)
}

// GetLatestVersion 获取最新版本
func (cp *ContextPersistence) GetLatestVersion() (*ContextVersion, error) {
	cp.versionsMu.RLock()
	defer cp.versionsMu.RUnlock()

	if len(cp.versions) == 0 {
		return nil, fmt.Errorf("no versions found")
	}

	return cp.versions[len(cp.versions)-1], nil
}

// GetVersions 获取所有版本
func (cp *ContextPersistence) GetVersions() []*ContextVersion {
	cp.versionsMu.RLock()
	defer cp.versionsMu.RUnlock()

	versions := make([]*ContextVersion, len(cp.versions))
	copy(versions, cp.versions)

	return versions
}

// RestoreFromVersion 从版本恢复
func (cp *ContextPersistence) RestoreFromVersion(versionNumber int64) ([]*Message, error) {
	version, err := cp.GetVersion(versionNumber)
	if err != nil {
		return nil, err
	}

	messages := make([]*Message, len(version.Messages))
	for i, snapshot := range version.Messages {
		messages[i] = &Message{
			ID:        snapshot.ID,
			Role:      snapshot.Role,
			Content:   snapshot.Content,
			Tokens:    snapshot.Tokens,
			Timestamp: snapshot.Timestamp,
		}
	}

	atomic.AddInt64(&cp.totalRestores, 1)

	cp.logFunc("info", fmt.Sprintf("Restored version %d for session %s", versionNumber, cp.SessionID))

	return messages, nil
}

// ToJSON 转换为 JSON
func (cp *ContextPersistence) ToJSON() (string, error) {
	cp.versionsMu.RLock()
	defer cp.versionsMu.RUnlock()

	data, err := json.MarshalIndent(cp.versions, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetStatistics 获取统计信息
func (cp *ContextPersistence) GetStatistics() map[string]interface{} {
	cp.versionsMu.RLock()
	defer cp.versionsMu.RUnlock()

	return map[string]interface{}{
		"session_id":       cp.SessionID,
		"current_version":  cp.currentVersion,
		"version_count":    len(cp.versions),
		"total_snapshots":  atomic.LoadInt64(&cp.totalSnapshots),
		"total_restores":   atomic.LoadInt64(&cp.totalRestores),
		"max_versions":     cp.maxVersions,
	}
}

// ContextRetrieval 上下文检索
type ContextRetrieval struct {
	// 会话上下文映射
	contexts map[string]*ConversationContext
	contextsMu sync.RWMutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewContextRetrieval 创建上下文检索
func NewContextRetrieval() *ContextRetrieval {
	return &ContextRetrieval{
		contexts: make(map[string]*ConversationContext),
		logFunc:  defaultLogFunc,
	}
}

// RegisterContext 注册上下文
func (cr *ContextRetrieval) RegisterContext(sessionID string, ctx *ConversationContext) error {
	cr.contextsMu.Lock()
	defer cr.contextsMu.Unlock()

	if _, exists := cr.contexts[sessionID]; exists {
		return fmt.Errorf("context for session %s already registered", sessionID)
	}

	cr.contexts[sessionID] = ctx

	return nil
}

// SearchMessages 搜索消息（简单实现）
func (cr *ContextRetrieval) SearchMessages(sessionID, keyword string) ([]*Message, error) {
	cr.contextsMu.RLock()
	ctx, exists := cr.contexts[sessionID]
	cr.contextsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context for session %s not found", sessionID)
	}

	messages := ctx.GetMessages()
	var results []*Message

	for _, msg := range messages {
		if len(keyword) == 0 || containsKeyword(msg.Content, keyword) {
			results = append(results, msg)
		}
	}

	return results, nil
}

// containsKeyword 检查关键字
func containsKeyword(content, keyword string) bool {
	return len(keyword) > 0 && len(content) > 0 && len([]rune(content)) >= len([]rune(keyword))
}

// GetContextSummary 获取上下文摘要
func (cr *ContextRetrieval) GetContextSummary(sessionID string) (map[string]interface{}, error) {
	cr.contextsMu.RLock()
	ctx, exists := cr.contexts[sessionID]
	cr.contextsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context for session %s not found", sessionID)
	}

	messages := ctx.GetMessages()
	userCount := 0
	assistantCount := 0

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}

	return map[string]interface{}{
		"session_id":       sessionID,
		"message_count":    ctx.GetMessageCount(),
		"total_tokens":     ctx.GetTotalTokens(),
		"current_round":    ctx.GetCurrentRound(),
		"user_messages":    userCount,
		"assistant_messages": assistantCount,
		"context_usage":    float64(ctx.GetTotalTokens()) / float64(ctx.maxContextSize) * 100.0,
	}, nil
}

// AdvancedContextManager 高级上下文管理器
type AdvancedContextManager struct {
	// 上下文管理器
	contextManager *ContextManager

	// 持久化存储映射
	persistences map[string]*ContextPersistence
	persistencesMu sync.RWMutex

	// 检索引擎
	retrieval *ContextRetrieval

	// 统计信息
	totalContexts int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewAdvancedContextManager 创建高级上下文管理器
func NewAdvancedContextManager(defaultMaxContextSize int64, defaultMaxRounds int) *AdvancedContextManager {
	return &AdvancedContextManager{
		contextManager: NewContextManager(defaultMaxContextSize, defaultMaxRounds),
		persistences:   make(map[string]*ContextPersistence),
		retrieval:      NewContextRetrieval(),
		logFunc:        defaultLogFunc,
	}
}

// CreateContext 创建上下文
func (acm *AdvancedContextManager) CreateContext(sessionID string) (*ConversationContext, error) {
	ctx, err := acm.contextManager.CreateContext(sessionID)
	if err != nil {
		return nil, err
	}

	// 创建持久化存储
	acm.persistencesMu.Lock()
	acm.persistences[sessionID] = NewContextPersistence(sessionID, 100)
	acm.persistencesMu.Unlock()

	// 注册检索引擎
	_ = acm.retrieval.RegisterContext(sessionID, ctx)

	atomic.AddInt64(&acm.totalContexts, 1)

	return ctx, nil
}

// GetContext 获取上下文
func (acm *AdvancedContextManager) GetContext(sessionID string) (*ConversationContext, error) {
	return acm.contextManager.GetContext(sessionID)
}

// SaveContextVersion 保存上下文版本
func (acm *AdvancedContextManager) SaveContextVersion(sessionID string, remark string) error {
	ctx, err := acm.contextManager.GetContext(sessionID)
	if err != nil {
		return err
	}

	acm.persistencesMu.RLock()
	persistence, exists := acm.persistences[sessionID]
	acm.persistencesMu.RUnlock()

	if !exists {
		return fmt.Errorf("persistence for session %s not found", sessionID)
	}

	messages := ctx.GetMessages()
	return persistence.SaveVersion(messages, ctx.GetTotalTokens(), remark)
}

// RestoreContextVersion 恢复上下文版本
func (acm *AdvancedContextManager) RestoreContextVersion(sessionID string, versionNumber int64) error {
	acm.persistencesMu.RLock()
	persistence, exists := acm.persistences[sessionID]
	acm.persistencesMu.RUnlock()

	if !exists {
		return fmt.Errorf("persistence for session %s not found", sessionID)
	}

	messages, err := persistence.RestoreFromVersion(versionNumber)
	if err != nil {
		return err
	}

	ctx, err := acm.contextManager.GetContext(sessionID)
	if err != nil {
		return err
	}

	// 清空上下文并恢复消息
	ctx.Clear()
	for _, msg := range messages {
		_ = ctx.AddMessage(msg)
	}

	return nil
}

// SearchMessages 搜索消息
func (acm *AdvancedContextManager) SearchMessages(sessionID, keyword string) ([]*Message, error) {
	return acm.retrieval.SearchMessages(sessionID, keyword)
}

// GetContextSummary 获取上下文摘要
func (acm *AdvancedContextManager) GetContextSummary(sessionID string) (map[string]interface{}, error) {
	return acm.retrieval.GetContextSummary(sessionID)
}

// GetContextVersions 获取上下文版本列表
func (acm *AdvancedContextManager) GetContextVersions(sessionID string) ([]*ContextVersion, error) {
	acm.persistencesMu.RLock()
	persistence, exists := acm.persistences[sessionID]
	acm.persistencesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("persistence for session %s not found", sessionID)
	}

	return persistence.GetVersions(), nil
}

// GetStatistics 获取统计信息
func (acm *AdvancedContextManager) GetStatistics() map[string]interface{} {
	acm.persistencesMu.RLock()
	defer acm.persistencesMu.RUnlock()

	persistenceStats := make(map[string]interface{})
	for sessionID, persistence := range acm.persistences {
		persistenceStats[sessionID] = persistence.GetStatistics()
	}

	return map[string]interface{}{
		"total_contexts":      atomic.LoadInt64(&acm.totalContexts),
		"persistence_stats":   persistenceStats,
		"context_manager":     acm.contextManager.GetStatistics(),
	}
}

