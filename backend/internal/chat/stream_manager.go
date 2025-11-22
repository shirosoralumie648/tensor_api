package chat

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// StreamMessage 流式消息
type StreamMessage struct {
	// 消息 ID
	MessageID string `json:"message_id"`

	// 会话 ID
	SessionID string `json:"session_id"`

	// 内容
	Content string `json:"content"`

	// 是否完成
	Done bool `json:"done"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`

	// 错误信息
	Error string `json:"error,omitempty"`

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StreamClient 流式客户端
type StreamClient struct {
	// 客户端 ID
	ClientID string

	// 消息通道
	MessageCh chan *StreamMessage

	// 缓冲区大小
	bufferSize int

	// 最后接收时间
	lastReceived time.Time
	lastMu       sync.RWMutex

	// 是否关闭
	closed  bool
	closeMu sync.Mutex

	// 关闭信号
	closeCh chan struct{}

	// 统计信息
	messageCount int64
	errorCount   int64
}

// NewStreamClient 创建流式客户端
func NewStreamClient(clientID string, bufferSize int) *StreamClient {
	return &StreamClient{
		ClientID:     clientID,
		MessageCh:    make(chan *StreamMessage, bufferSize),
		bufferSize:   bufferSize,
		lastReceived: time.Now(),
		closeCh:      make(chan struct{}),
	}
}

// SendMessage 发送消息
func (sc *StreamClient) SendMessage(msg *StreamMessage) error {
	sc.closeMu.Lock()
	if sc.closed {
		sc.closeMu.Unlock()
		return fmt.Errorf("client is closed")
	}
	sc.closeMu.Unlock()

	select {
	case sc.MessageCh <- msg:
		atomic.AddInt64(&sc.messageCount, 1)
		sc.lastMu.Lock()
		sc.lastReceived = time.Now()
		sc.lastMu.Unlock()
		return nil
	case <-sc.closeCh:
		return fmt.Errorf("client is closed")
	default:
		atomic.AddInt64(&sc.errorCount, 1)
		return fmt.Errorf("message channel is full")
	}
}

// Close 关闭客户端
func (sc *StreamClient) Close() error {
	sc.closeMu.Lock()
	defer sc.closeMu.Unlock()

	if sc.closed {
		return nil
	}

	sc.closed = true
	close(sc.closeCh)
	close(sc.MessageCh)

	return nil
}

// GetLastReceived 获取最后接收时间
func (sc *StreamClient) GetLastReceived() time.Time {
	sc.lastMu.RLock()
	defer sc.lastMu.RUnlock()
	return sc.lastReceived
}

// GetStatistics 获取统计信息
func (sc *StreamClient) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"message_count": atomic.LoadInt64(&sc.messageCount),
		"error_count":   atomic.LoadInt64(&sc.errorCount),
		"last_received": sc.GetLastReceived(),
	}
}

// StreamSession 流式会话
type StreamSession struct {
	// 会话 ID
	SessionID string

	// 用户 ID
	UserID string

	// 模型名称
	ModelName string

	// 活跃客户端列表
	clients   map[string]*StreamClient
	clientsMu sync.RWMutex

	// 创建时间
	CreatedAt time.Time

	// 最后活动时间
	LastActivityAt time.Time
	activityMu     sync.RWMutex

	// 统计信息
	totalMessages int64
	totalErrors   int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewStreamSession 创建流式会话
func NewStreamSession(sessionID, userID, modelName string) *StreamSession {
	return &StreamSession{
		SessionID:      sessionID,
		UserID:         userID,
		ModelName:      modelName,
		clients:        make(map[string]*StreamClient),
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		logFunc:        defaultLogFunc,
	}
}

// RegisterClient 注册客户端
func (ss *StreamSession) RegisterClient(clientID string, bufferSize int) (*StreamClient, error) {
	ss.clientsMu.Lock()
	defer ss.clientsMu.Unlock()

	if _, exists := ss.clients[clientID]; exists {
		return nil, fmt.Errorf("client %s already exists", clientID)
	}

	client := NewStreamClient(clientID, bufferSize)
	ss.clients[clientID] = client

	ss.updateActivity()

	ss.logFunc("info", fmt.Sprintf("Client %s registered for session %s", clientID, ss.SessionID))

	return client, nil
}

// UnregisterClient 注销客户端
func (ss *StreamSession) UnregisterClient(clientID string) error {
	ss.clientsMu.Lock()
	defer ss.clientsMu.Unlock()

	client, exists := ss.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	_ = client.Close()
	delete(ss.clients, clientID)

	ss.updateActivity()

	ss.logFunc("info", fmt.Sprintf("Client %s unregistered from session %s", clientID, ss.SessionID))

	return nil
}

// BroadcastMessage 广播消息给所有客户端
func (ss *StreamSession) BroadcastMessage(msg *StreamMessage) error {
	ss.clientsMu.RLock()
	clients := make(map[string]*StreamClient)
	for k, v := range ss.clients {
		clients[k] = v
	}
	ss.clientsMu.RUnlock()

	var errCount int
	for _, client := range clients {
		if err := client.SendMessage(msg); err != nil {
			errCount++
			atomic.AddInt64(&ss.totalErrors, 1)
		} else {
			atomic.AddInt64(&ss.totalMessages, 1)
		}
	}

	ss.updateActivity()

	if errCount > 0 {
		ss.logFunc("warn", fmt.Sprintf("Failed to send message to %d/%d clients in session %s", errCount, len(clients), ss.SessionID))
	}

	return nil
}

// SendMessageToClient 发送消息给特定客户端
func (ss *StreamSession) SendMessageToClient(clientID string, msg *StreamMessage) error {
	ss.clientsMu.RLock()
	client, exists := ss.clients[clientID]
	ss.clientsMu.RUnlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	err := client.SendMessage(msg)
	if err == nil {
		atomic.AddInt64(&ss.totalMessages, 1)
	} else {
		atomic.AddInt64(&ss.totalErrors, 1)
	}

	ss.updateActivity()

	return err
}

// updateActivity 更新活动时间
func (ss *StreamSession) updateActivity() {
	ss.activityMu.Lock()
	ss.LastActivityAt = time.Now()
	ss.activityMu.Unlock()
}

// GetClientCount 获取客户端数量
func (ss *StreamSession) GetClientCount() int {
	ss.clientsMu.RLock()
	defer ss.clientsMu.RUnlock()
	return len(ss.clients)
}

// GetStatistics 获取统计信息
func (ss *StreamSession) GetStatistics() map[string]interface{} {
	ss.clientsMu.RLock()
	clientStats := make(map[string]interface{})
	for clientID, client := range ss.clients {
		clientStats[clientID] = client.GetStatistics()
	}
	ss.clientsMu.RUnlock()

	ss.activityMu.RLock()
	lastActivity := ss.LastActivityAt
	ss.activityMu.RUnlock()

	return map[string]interface{}{
		"session_id":     ss.SessionID,
		"user_id":        ss.UserID,
		"model_name":     ss.ModelName,
		"client_count":   ss.GetClientCount(),
		"total_messages": atomic.LoadInt64(&ss.totalMessages),
		"total_errors":   atomic.LoadInt64(&ss.totalErrors),
		"created_at":     ss.CreatedAt,
		"last_activity":  lastActivity,
		"client_stats":   clientStats,
	}
}

// StreamManager 流式管理器
type StreamManager struct {
	// 会话映射
	sessions   map[string]*StreamSession
	sessionsMu sync.RWMutex

	// 客户端超时时间
	clientTimeout time.Duration

	// 会话超时时间
	sessionTimeout time.Duration

	// 统计信息
	totalSessions  int64
	activeSessions int64

	// 清理 goroutine 控制
	stopCh  chan struct{}
	running bool
	runMu   sync.Mutex

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// NewStreamManager 创建流式管理器
func NewStreamManager(clientTimeout, sessionTimeout time.Duration) *StreamManager {
	return &StreamManager{
		sessions:       make(map[string]*StreamSession),
		clientTimeout:  clientTimeout,
		sessionTimeout: sessionTimeout,
		stopCh:         make(chan struct{}),
		logFunc:        defaultLogFunc,
	}
}

// CreateSession 创建会话
func (sm *StreamManager) CreateSession(sessionID, userID, modelName string) (*StreamSession, error) {
	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	session := NewStreamSession(sessionID, userID, modelName)
	sm.sessions[sessionID] = session

	atomic.AddInt64(&sm.totalSessions, 1)
	atomic.AddInt64(&sm.activeSessions, 1)

	sm.logFunc("info", fmt.Sprintf("Session %s created for user %s", sessionID, userID))

	return session, nil
}

// GetSession 获取会话
func (sm *StreamManager) GetSession(sessionID string) (*StreamSession, error) {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// CloseSession 关闭会话
func (sm *StreamManager) CloseSession(sessionID string) error {
	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// 关闭所有客户端
	session.clientsMu.Lock()
	for _, client := range session.clients {
		_ = client.Close()
	}
	session.clients = make(map[string]*StreamClient)
	session.clientsMu.Unlock()

	delete(sm.sessions, sessionID)

	atomic.AddInt64(&sm.activeSessions, -1)

	sm.logFunc("info", fmt.Sprintf("Session %s closed", sessionID))

	return nil
}

// Start 启动管理器（包括清理 goroutine）
func (sm *StreamManager) Start() {
	sm.runMu.Lock()
	if sm.running {
		sm.runMu.Unlock()
		return
	}
	sm.running = true
	sm.runMu.Unlock()

	go sm.cleanupLoop()

	sm.logFunc("info", "Stream manager started")
}

// cleanupLoop 清理循环
func (sm *StreamManager) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopCh:
			sm.runMu.Lock()
			sm.running = false
			sm.runMu.Unlock()
			return

		case <-ticker.C:
			sm.cleanup()
		}
	}
}

// cleanup 执行清理
func (sm *StreamManager) cleanup() {
	sm.sessionsMu.Lock()
	defer sm.sessionsMu.Unlock()

	now := time.Now()
	var toDelete []string

	for sessionID, session := range sm.sessions {
		session.activityMu.RLock()
		lastActivity := session.LastActivityAt
		session.activityMu.RUnlock()

		// 检查会话是否超时
		if now.Sub(lastActivity) > sm.sessionTimeout {
			toDelete = append(toDelete, sessionID)
		} else {
			// 检查客户端是否超时
			session.clientsMu.Lock()
			var clientsToDelete []string
			for clientID, client := range session.clients {
				if now.Sub(client.GetLastReceived()) > sm.clientTimeout {
					clientsToDelete = append(clientsToDelete, clientID)
				}
			}

			for _, clientID := range clientsToDelete {
				_ = session.clients[clientID].Close()
				delete(session.clients, clientID)
				sm.logFunc("info", fmt.Sprintf("Timeout: Client %s removed from session %s", clientID, sessionID))
			}
			session.clientsMu.Unlock()
		}
	}

	for _, sessionID := range toDelete {
		_ = sm.CloseSession(sessionID)
	}
}

// Stop 停止管理器
func (sm *StreamManager) Stop() {
	sm.runMu.Lock()
	if !sm.running {
		sm.runMu.Unlock()
		return
	}
	sm.runMu.Unlock()

	close(sm.stopCh)

	sm.logFunc("info", "Stream manager stopped")
}

// GetStatistics 获取统计信息
func (sm *StreamManager) GetStatistics() map[string]interface{} {
	sm.sessionsMu.RLock()
	defer sm.sessionsMu.RUnlock()

	sessionStats := make(map[string]interface{})
	for sessionID, session := range sm.sessions {
		sessionStats[sessionID] = session.GetStatistics()
	}

	return map[string]interface{}{
		"total_sessions":  atomic.LoadInt64(&sm.totalSessions),
		"active_sessions": atomic.LoadInt64(&sm.activeSessions),
		"sessions":        sessionStats,
	}
}

// defaultLogFunc 默认日志函数
func defaultLogFunc(level, msg string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", level, msg)
}
