package chat

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Message 对话消息
type Message struct {
	// 消息 ID
	ID string `json:"id"`

	// 角色（user/assistant/system）
	Role string `json:"role"`

	// 内容
	Content string `json:"content"`

	// Token 数
	Tokens int64 `json:"tokens"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationContext 对话上下文
type ConversationContext struct {
	// 会话 ID
	SessionID string

	// 消息列表
	messages []*Message
	messagesMu sync.RWMutex

	// 总 token 数
	totalTokens int64

	// 最大上下文大小（token）
	maxContextSize int64

	// 消息最大轮数
	maxRounds int

	// 当前轮数
	currentRound int

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time
	updateMu  sync.RWMutex

	// 统计信息
	messageCount int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewConversationContext 创建对话上下文
func NewConversationContext(sessionID string, maxContextSize int64, maxRounds int) *ConversationContext {
	return &ConversationContext{
		SessionID:      sessionID,
		messages:       make([]*Message, 0),
		maxContextSize: maxContextSize,
		maxRounds:      maxRounds,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		logFunc:        defaultLogFunc,
	}
}

// AddMessage 添加消息
func (cc *ConversationContext) AddMessage(msg *Message) error {
	cc.messagesMu.Lock()
	defer cc.messagesMu.Unlock()

	// 检查 token 数是否超过限制
	if cc.totalTokens+msg.Tokens > cc.maxContextSize {
		// 移除最老的消息以腾出空间
		cc.truncateOldestMessages(msg.Tokens)
	}

	cc.messages = append(cc.messages, msg)
	cc.totalTokens += msg.Tokens

	// 更新轮数（user 和 assistant 消息对算一轮）
	if msg.Role == "assistant" {
		cc.currentRound++
	}

	atomic.AddInt64(&cc.messageCount, 1)

	cc.updateTime()

	cc.logFunc("debug", fmt.Sprintf("Message added to session %s (role: %s, tokens: %d)", cc.SessionID, msg.Role, msg.Tokens))

	return nil
}

// truncateOldestMessages 删除最老的消息以腾出空间
func (cc *ConversationContext) truncateOldestMessages(requiredTokens int64) {
	for len(cc.messages) > 0 && cc.totalTokens+requiredTokens > cc.maxContextSize {
		oldMsg := cc.messages[0]
		cc.messages = cc.messages[1:]
		cc.totalTokens -= oldMsg.Tokens

		cc.logFunc("info", fmt.Sprintf("Truncated old message from session %s", cc.SessionID))

		// 如果移除的是 assistant 消息，轮数减 1
		if oldMsg.Role == "assistant" {
			cc.currentRound--
		}
	}
}

// GetMessages 获取消息列表
func (cc *ConversationContext) GetMessages() []*Message {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()

	// 返回副本以避免外部修改
	messages := make([]*Message, len(cc.messages))
	copy(messages, cc.messages)

	return messages
}

// GetRecentMessages 获取最近的 N 条消息
func (cc *ConversationContext) GetRecentMessages(count int) []*Message {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()

	if count > len(cc.messages) {
		count = len(cc.messages)
	}

	start := len(cc.messages) - count
	if start < 0 {
		start = 0
	}

	messages := make([]*Message, count)
	copy(messages, cc.messages[start:])

	return messages
}

// GetMessageCount 获取消息总数
func (cc *ConversationContext) GetMessageCount() int {
	return int(atomic.LoadInt64(&cc.messageCount))
}

// GetTotalTokens 获取总 token 数
func (cc *ConversationContext) GetTotalTokens() int64 {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()
	return cc.totalTokens
}

// GetCurrentRound 获取当前轮数
func (cc *ConversationContext) GetCurrentRound() int {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()
	return cc.currentRound
}

// IsContextFull 检查上下文是否满
func (cc *ConversationContext) IsContextFull() bool {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()
	return cc.totalTokens >= cc.maxContextSize
}

// IsMaxRoundsReached 检查是否达到最大轮数
func (cc *ConversationContext) IsMaxRoundsReached() bool {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()
	return cc.currentRound >= cc.maxRounds
}

// updateTime 更新时间戳
func (cc *ConversationContext) updateTime() {
	cc.updateMu.Lock()
	cc.UpdatedAt = time.Now()
	cc.updateMu.Unlock()
}

// Clear 清空上下文
func (cc *ConversationContext) Clear() {
	cc.messagesMu.Lock()
	defer cc.messagesMu.Unlock()

	cc.messages = make([]*Message, 0)
	cc.totalTokens = 0
	cc.currentRound = 0

	cc.updateTime()

	cc.logFunc("info", fmt.Sprintf("Context cleared for session %s", cc.SessionID))
}

// GetStatistics 获取统计信息
func (cc *ConversationContext) GetStatistics() map[string]interface{} {
	cc.messagesMu.RLock()
	defer cc.messagesMu.RUnlock()

	cc.updateMu.RLock()
	updatedAt := cc.UpdatedAt
	cc.updateMu.RUnlock()

	return map[string]interface{}{
		"session_id":      cc.SessionID,
		"message_count":   atomic.LoadInt64(&cc.messageCount),
		"total_tokens":    cc.totalTokens,
		"current_round":   cc.currentRound,
		"max_rounds":      cc.maxRounds,
		"context_usage":   float64(cc.totalTokens) / float64(cc.maxContextSize) * 100.0,
		"created_at":      cc.CreatedAt,
		"updated_at":      updatedAt,
	}
}

// ContextManager 上下文管理器
type ContextManager struct {
	// 会话上下文映射
	contexts map[string]*ConversationContext
	contextsMu sync.RWMutex

	// 默认最大上下文大小（token）
	defaultMaxContextSize int64

	// 默认最大轮数
	defaultMaxRounds int

	// 统计信息
	totalContexts int64
	activeContexts int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewContextManager 创建上下文管理器
func NewContextManager(defaultMaxContextSize int64, defaultMaxRounds int) *ContextManager {
	return &ContextManager{
		contexts:              make(map[string]*ConversationContext),
		defaultMaxContextSize: defaultMaxContextSize,
		defaultMaxRounds:      defaultMaxRounds,
		logFunc:               defaultLogFunc,
	}
}

// CreateContext 创建上下文
func (cm *ContextManager) CreateContext(sessionID string) (*ConversationContext, error) {
	cm.contextsMu.Lock()
	defer cm.contextsMu.Unlock()

	if _, exists := cm.contexts[sessionID]; exists {
		return nil, fmt.Errorf("context for session %s already exists", sessionID)
	}

	ctx := NewConversationContext(sessionID, cm.defaultMaxContextSize, cm.defaultMaxRounds)
	cm.contexts[sessionID] = ctx

	atomic.AddInt64(&cm.totalContexts, 1)
	atomic.AddInt64(&cm.activeContexts, 1)

	return ctx, nil
}

// GetContext 获取上下文
func (cm *ContextManager) GetContext(sessionID string) (*ConversationContext, error) {
	cm.contextsMu.RLock()
	defer cm.contextsMu.RUnlock()

	ctx, exists := cm.contexts[sessionID]
	if !exists {
		return nil, fmt.Errorf("context for session %s not found", sessionID)
	}

	return ctx, nil
}

// DeleteContext 删除上下文
func (cm *ContextManager) DeleteContext(sessionID string) error {
	cm.contextsMu.Lock()
	defer cm.contextsMu.Unlock()

	if _, exists := cm.contexts[sessionID]; !exists {
		return fmt.Errorf("context for session %s not found", sessionID)
	}

	delete(cm.contexts, sessionID)

	atomic.AddInt64(&cm.activeContexts, -1)

	cm.logFunc("info", fmt.Sprintf("Context deleted for session %s", sessionID))

	return nil
}

// GetStatistics 获取统计信息
func (cm *ContextManager) GetStatistics() map[string]interface{} {
	cm.contextsMu.RLock()
	defer cm.contextsMu.RUnlock()

	contextStats := make(map[string]interface{})
	for sessionID, ctx := range cm.contexts {
		contextStats[sessionID] = ctx.GetStatistics()
	}

	return map[string]interface{}{
		"total_contexts":  atomic.LoadInt64(&cm.totalContexts),
		"active_contexts": atomic.LoadInt64(&cm.activeContexts),
		"contexts":        contextStats,
	}
}

// SummarizeContext 总结上下文（简化实现）
func (cm *ContextManager) SummarizeContext(sessionID string) (string, error) {
	ctx, err := cm.GetContext(sessionID)
	if err != nil {
		return "", err
	}

	messages := ctx.GetMessages()

	if len(messages) == 0 {
		return "Empty context", nil
	}

	// 简化实现：仅统计消息信息
	userMsgCount := 0
	assistantMsgCount := 0
	totalLen := 0

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userMsgCount++
		case "assistant":
			assistantMsgCount++
		}
		totalLen += len(msg.Content)
	}

	summary := fmt.Sprintf(
		"Conversation summary: %d user messages, %d assistant messages, total length: %d chars, rounds: %d",
		userMsgCount,
		assistantMsgCount,
		totalLen,
		ctx.GetCurrentRound(),
	)

	return summary, nil
}

