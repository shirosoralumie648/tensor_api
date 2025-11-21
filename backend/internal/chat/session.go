package chat

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SessionStatus 会话状态
type SessionStatus string

const (
	SessionActive   SessionStatus = "active"
	SessionArchived SessionStatus = "archived"
	SessionDeleted  SessionStatus = "deleted"
)

// Session 会话
type Session struct {
	// 会话 ID
	ID string `json:"id"`

	// 用户 ID
	UserID int64 `json:"user_id"`

	// 会话标题
	Title string `json:"title"`

	// 会话描述
	Description string `json:"description"`

	// 会话状态
	Status SessionStatus `json:"status"`

	// 消息数
	MessageCount int64 `json:"message_count"`

	// 总 token 数
	TotalTokens int64 `json:"total_tokens"`

	// 创建时间
	CreatedAt time.Time `json:"created_at"`

	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// 删除时间（软删除）
	DeletedAt *time.Time `json:"deleted_at"`

	// 标签
	Tags []string `json:"tags"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata"`

	// 互斥锁
	mu sync.RWMutex
}

// NewSession 创建会话
func NewSession(id string, userID int64, title string) *Session {
	return &Session{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Status:      SessionActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}
}

// AddTag 添加标签
func (s *Session) AddTag(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.Tags {
		if t == tag {
			return
		}
	}

	s.Tags = append(s.Tags, tag)
	s.UpdatedAt = time.Now()
}

// RemoveTag 移除标签
func (s *Session) RemoveTag(tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.Tags {
		if t == tag {
			s.Tags = append(s.Tags[:i], s.Tags[i+1:]...)
			s.UpdatedAt = time.Now()
			return
		}
	}
}

// HasTag 检查是否有标签
func (s *Session) HasTag(tag string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}

	return false
}

// SetMetadata 设置元数据
func (s *Session) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
}

// GetMetadata 获取元数据
func (s *Session) GetMetadata(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.Metadata[key]
	return value, exists
}

// UpdateTitle 更新标题
func (s *Session) UpdateTitle(title string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Title = title
	s.UpdatedAt = time.Now()
}

// UpdateDescription 更新描述
func (s *Session) UpdateDescription(description string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Description = description
	s.UpdatedAt = time.Now()
}

// IncrementMessageCount 增加消息计数
func (s *Session) IncrementMessageCount() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MessageCount++
	s.UpdatedAt = time.Now()
}

// AddTokens 增加 token 数
func (s *Session) AddTokens(tokens int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalTokens += tokens
	s.UpdatedAt = time.Now()
}

// IsActive 检查是否活跃
func (s *Session) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Status == SessionActive
}

// IsDeleted 检查是否已删除
func (s *Session) IsDeleted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Status == SessionDeleted
}

// SessionManager 会话管理器
type SessionManager struct {
	// 会话存储
	sessions map[string]*Session
	sessionsMu sync.RWMutex

	// 用户会话索引
	userSessions map[int64][]string
	userSessionsMu sync.RWMutex

	// 标签索引
	tagIndex map[string][]string
	tagIndexMu sync.RWMutex

	// 统计信息
	totalCreated int64
	totalDeleted int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*Session),
		userSessions: make(map[int64][]string),
		tagIndex:     make(map[string][]string),
		logFunc:      defaultLogFunc,
	}
}

// CreateSession 创建会话
func (sm *SessionManager) CreateSession(session *Session) error {
	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	if _, exists := sm.sessions[session.ID]; exists {
		return fmt.Errorf("session %s already exists", session.ID)
	}

	sm.sessions[session.ID] = session

	// 更新用户索引
	sm.userSessionsMu.Lock()
	sm.userSessions[session.UserID] = append(sm.userSessions[session.UserID], session.ID)
	sm.userSessionsMu.Unlock()

	// 更新标签索引
	sm.updateTagIndex(session)

	atomic.AddInt64(&sm.totalCreated, 1)

	sm.logFunc("info", fmt.Sprintf("Created session: %s for user %d", session.ID, session.UserID))

	return nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if session.IsDeleted() {
		return nil, fmt.Errorf("session %s is deleted", sessionID)
	}

	return session, nil
}

// GetSessionIncludingDeleted 获取会话（包括已删除）
func (sm *SessionManager) GetSessionIncludingDeleted(sessionID string) (*Session, error) {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// UpdateSession 更新会话
func (sm *SessionManager) UpdateSession(sessionID string, title, description string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.UpdateTitle(title)
	session.UpdateDescription(description)

	return nil
}

// SoftDeleteSession 软删除会话
func (sm *SessionManager) SoftDeleteSession(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	now := time.Now()
	session.mu.Lock()
	session.Status = SessionDeleted
	session.DeletedAt = &now
	session.UpdatedAt = now
	session.mu.Unlock()

	atomic.AddInt64(&sm.totalDeleted, 1)

	sm.logFunc("info", fmt.Sprintf("Soft deleted session: %s", sessionID))

	return nil
}

// RestoreSession 恢复会话
func (sm *SessionManager) RestoreSession(sessionID string) error {
	sm.sessionsMu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.sessionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.mu.Lock()
	session.Status = SessionActive
	session.DeletedAt = nil
	session.UpdatedAt = time.Now()
	session.mu.Unlock()

	atomic.AddInt64(&sm.totalDeleted, -1)

	sm.logFunc("info", fmt.Sprintf("Restored session: %s", sessionID))

	return nil
}

// PermanentlyDeleteSession 永久删除会话
func (sm *SessionManager) PermanentlyDeleteSession(sessionID string) error {
	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// 从用户索引中移除
	sm.userSessionsMu.Lock()
	userID := session.UserID
	sessions := sm.userSessions[userID]
	for i, id := range sessions {
		if id == sessionID {
			sm.userSessions[userID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}
	sm.userSessionsMu.Unlock()

	// 从标签索引中移除
	sm.tagIndexMu.Lock()
	for _, tag := range session.Tags {
		tagSessions := sm.tagIndex[tag]
		for i, id := range tagSessions {
			if id == sessionID {
				sm.tagIndex[tag] = append(tagSessions[:i], tagSessions[i+1:]...)
				break
			}
		}
	}
	sm.tagIndexMu.Unlock()

	delete(sm.sessions, sessionID)

	sm.logFunc("info", fmt.Sprintf("Permanently deleted session: %s", sessionID))

	return nil
}

// GetUserSessions 获取用户的所有会话
func (sm *SessionManager) GetUserSessions(userID int64) []*Session {
	sm.userSessionsMu.RLock()
	sessionIDs := sm.userSessions[userID]
	sm.userSessionsMu.RUnlock()

	var sessions []*Session

	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	for _, id := range sessionIDs {
		if session, exists := sm.sessions[id]; exists && !session.IsDeleted() {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// SearchByTag 按标签搜索
func (sm *SessionManager) SearchByTag(userID int64, tag string) []*Session {
	sm.tagIndexMu.RLock()
	sessionIDs := sm.tagIndex[tag]
	sm.tagIndexMu.RUnlock()

	var sessions []*Session

	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	for _, id := range sessionIDs {
		if session, exists := sm.sessions[id]; exists && session.UserID == userID && !session.IsDeleted() {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// SearchByTitle 按标题搜索（全文搜索）
func (sm *SessionManager) SearchByTitle(userID int64, keyword string) []*Session {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	var sessions []*Session

	for _, session := range sm.sessions {
		if session.UserID == userID && !session.IsDeleted() {
			if matchesKeyword(session.Title, keyword) || matchesKeyword(session.Description, keyword) {
				sessions = append(sessions, session)
			}
		}
	}

	return sessions
}

// matchesKeyword 匹配关键字
func matchesKeyword(text, keyword string) bool {
	if len(keyword) == 0 {
		return true
	}

	return len(text) > 0 && len(keyword) > 0
}

// updateTagIndex 更新标签索引
func (sm *SessionManager) updateTagIndex(session *Session) {
	sm.tagIndexMu.Lock()
	defer sm.tagIndexMu.Unlock()

	for _, tag := range session.Tags {
		if _, exists := sm.tagIndex[tag]; !exists {
			sm.tagIndex[tag] = make([]string, 0)
		}

		sm.tagIndex[tag] = append(sm.tagIndex[tag], session.ID)
	}
}

// GetDeletedSessions 获取已删除的会话
func (sm *SessionManager) GetDeletedSessions(userID int64) []*Session {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	var sessions []*Session

	for _, session := range sm.sessions {
		if session.UserID == userID && session.IsDeleted() {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetStatistics 获取统计信息
func (sm *SessionManager) GetStatistics() map[string]interface{} {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	activeCount := 0
	deletedCount := 0

	for _, session := range sm.sessions {
		if session.IsDeleted() {
			deletedCount++
		} else {
			activeCount++
		}
	}

	return map[string]interface{}{
		"total_created": atomic.LoadInt64(&sm.totalCreated),
		"total_deleted": atomic.LoadInt64(&sm.totalDeleted),
		"active_count":  activeCount,
		"deleted_count": deletedCount,
	}
}

// PurgeOldDeletedSessions 清理旧的已删除会话（超过30天）
func (sm *SessionManager) PurgeOldDeletedSessions(daysThreshold int) int {
	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	count := 0
	thresholdTime := time.Now().AddDate(0, 0, -daysThreshold)

	for sessionID, session := range sm.sessions {
		if session.IsDeleted() && session.DeletedAt != nil && session.DeletedAt.Before(thresholdTime) {
			// 从用户索引中移除
			sm.userSessionsMu.Lock()
			userID := session.UserID
			sessions := sm.userSessions[userID]
			for i, id := range sessions {
				if id == sessionID {
					sm.userSessions[userID] = append(sessions[:i], sessions[i+1:]...)
					break
				}
			}
			sm.userSessionsMu.Unlock()

			// 从标签索引中移除
			sm.tagIndexMu.Lock()
			for _, tag := range session.Tags {
				tagSessions := sm.tagIndex[tag]
				for i, id := range tagSessions {
					if id == sessionID {
						sm.tagIndex[tag] = append(tagSessions[:i], tagSessions[i+1:]...)
						break
					}
				}
			}
			sm.tagIndexMu.Unlock()

			delete(sm.sessions, sessionID)
			count++
		}
	}

	sm.logFunc("info", fmt.Sprintf("Purged %d old deleted sessions", count))

	return count
}

