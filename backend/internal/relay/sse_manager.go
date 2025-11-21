package relay

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// SSEMessage SSE 消息结构
type SSEMessage struct {
	// 消息 ID (可选，用于客户端识别)
	ID string

	// 事件类型 (可选)
	Event string

	// 消息数据
	Data string

	// 重试时间 (毫秒，可选)
	Retry int64

	// 注释 (可选，调试用)
	Comment string
}

// SSEClient 代表一个 SSE 客户端连接
type SSEClient struct {
	// 客户端 ID
	ID string

	// 消息通道
	MessageChan chan *SSEMessage

	// 关闭通道
	CloseChan chan struct{}

	// 连接创建时间
	CreatedAt time.Time

	// 最后活动时间
	LastActivityAt time.Time

	// 客户端信息
	UserID string
	IP     string

	// 发送消息计数
	MessageCount int64

	// 字节数
	BytesSent int64

	// 是否已关闭
	closed bool
	mu     sync.Mutex
}

// SSEManager SSE 管理器
type SSEManager struct {
	// 所有客户端
	clients map[string]*SSEClient

	// 客户端锁
	clientsMu sync.RWMutex

	// 心跳间隔
	heartbeatInterval time.Duration

	// 心跳内容
	heartbeatData string

	// 客户端超时时间
	clientTimeout time.Duration

	// 最大客户端数
	maxClients int

	// 统计信息
	totalConnections   int64
	activeConnections  int32
	totalMessagesSent  int64
	totalBytesSent     int64

	// 停止信号
	stopChan chan struct{}

	// 等待组
	wg sync.WaitGroup
}

// NewSSEManager 创建新的 SSE 管理器
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients:           make(map[string]*SSEClient),
		heartbeatInterval: 30 * time.Second,
		heartbeatData:     ": heartbeat\n",
		clientTimeout:     5 * time.Minute,
		maxClients:        10000,
		stopChan:          make(chan struct{}),
	}
}

// SetHeartbeatInterval 设置心跳间隔
func (sm *SSEManager) SetHeartbeatInterval(interval time.Duration) {
	sm.heartbeatInterval = interval
}

// SetClientTimeout 设置客户端超时时间
func (sm *SSEManager) SetClientTimeout(timeout time.Duration) {
	sm.clientTimeout = timeout
}

// SetMaxClients 设置最大客户端数
func (sm *SSEManager) SetMaxClients(max int) {
	sm.maxClients = max
}

// RegisterClient 注册新客户端
func (sm *SSEManager) RegisterClient(clientID string, userID string, ip string) (*SSEClient, error) {
	sm.clientsMu.Lock()
	defer sm.clientsMu.Unlock()

	// 检查是否已达到最大连接数
	if len(sm.clients) >= sm.maxClients {
		return nil, fmt.Errorf("max clients reached")
	}

	client := &SSEClient{
		ID:             clientID,
		MessageChan:    make(chan *SSEMessage, 100), // 缓冲通道
		CloseChan:      make(chan struct{}),
		CreatedAt:      time.Now(),
		LastActivityAt: time.Now(),
		UserID:         userID,
		IP:             ip,
	}

	sm.clients[clientID] = client
	atomic.AddInt32(&sm.activeConnections, 1)
	atomic.AddInt64(&sm.totalConnections, 1)

	return client, nil
}

// UnregisterClient 注销客户端
func (sm *SSEManager) UnregisterClient(clientID string) {
	sm.clientsMu.Lock()
	defer sm.clientsMu.Unlock()

	if client, exists := sm.clients[clientID]; exists {
		client.Close()
		delete(sm.clients, clientID)
		atomic.AddInt32(&sm.activeConnections, -1)
	}
}

// BroadcastMessage 广播消息给所有客户端
func (sm *SSEManager) BroadcastMessage(msg *SSEMessage) {
	sm.clientsMu.RLock()
	defer sm.clientsMu.RUnlock()

	for _, client := range sm.clients {
		select {
		case client.MessageChan <- msg:
			atomic.AddInt64(&sm.totalMessagesSent, 1)
			atomic.AddInt64(&sm.totalBytesSent, int64(len(msg.Data)))
			atomic.AddInt64(&client.MessageCount, 1)
			client.BytesSent += int64(len(msg.Data))
		default:
			// 通道满，丢弃消息
		}
	}
}

// SendMessageToClient 发送消息给特定客户端
func (sm *SSEManager) SendMessageToClient(clientID string, msg *SSEMessage) error {
	sm.clientsMu.RLock()
	client, exists := sm.clients[clientID]
	sm.clientsMu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found")
	}

	select {
	case client.MessageChan <- msg:
		atomic.AddInt64(&sm.totalMessagesSent, 1)
		atomic.AddInt64(&sm.totalBytesSent, int64(len(msg.Data)))
		atomic.AddInt64(&client.MessageCount, 1)
		client.BytesSent += int64(len(msg.Data))
		return nil
	default:
		return fmt.Errorf("message channel full")
	}
}

// GetActiveClientCount 获取活跃客户端数
func (sm *SSEManager) GetActiveClientCount() int {
	return int(atomic.LoadInt32(&sm.activeConnections))
}

// GetTotalConnections 获取总连接数
func (sm *SSEManager) GetTotalConnections() int64 {
	return atomic.LoadInt64(&sm.totalConnections)
}

// GetStatistics 获取统计信息
func (sm *SSEManager) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"active_connections": atomic.LoadInt32(&sm.activeConnections),
		"total_connections":  atomic.LoadInt64(&sm.totalConnections),
		"total_messages":     atomic.LoadInt64(&sm.totalMessagesSent),
		"total_bytes":        atomic.LoadInt64(&sm.totalBytesSent),
	}
}

// Start 启动 SSE 管理器
func (sm *SSEManager) Start() {
	sm.wg.Add(1)
	go sm.runHeartbeat()
	sm.wg.Add(1)
	go sm.runCleanup()
}

// Stop 停止 SSE 管理器
func (sm *SSEManager) Stop() {
	close(sm.stopChan)
	sm.wg.Wait()

	// 关闭所有客户端连接
	sm.clientsMu.Lock()
	for _, client := range sm.clients {
		client.Close()
	}
	sm.clients = make(map[string]*SSEClient)
	sm.clientsMu.Unlock()
}

// runHeartbeat 运行心跳循环
func (sm *SSEManager) runHeartbeat() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.clientsMu.RLock()
			clients := make([]*SSEClient, 0, len(sm.clients))
			for _, client := range sm.clients {
				clients = append(clients, client)
			}
			sm.clientsMu.RUnlock()

			for _, client := range clients {
				select {
				case client.MessageChan <- &SSEMessage{
					Comment: "heartbeat",
					Data:    sm.heartbeatData,
				}:
				default:
				}
			}
		}
	}
}

// runCleanup 运行清理循环
func (sm *SSEManager) runCleanup() {
	defer sm.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.clientsMu.Lock()
			now := time.Now()
			var toDelete []string

			for id, client := range sm.clients {
				// 检查客户端是否超时
				if now.Sub(client.LastActivityAt) > sm.clientTimeout {
					toDelete = append(toDelete, id)
				}
			}

			for _, id := range toDelete {
				client := sm.clients[id]
				client.Close()
				delete(sm.clients, id)
				atomic.AddInt32(&sm.activeConnections, -1)
			}
			sm.clientsMu.Unlock()
		}
	}
}

// Close 关闭客户端
func (client *SSEClient) Close() {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closed {
		return
	}

	client.closed = true
	close(client.CloseChan)
	close(client.MessageChan)
}

// IsClosed 检查客户端是否已关闭
func (client *SSEClient) IsClosed() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.closed
}

// UpdateActivity 更新最后活动时间
func (client *SSEClient) UpdateActivity() {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.LastActivityAt = time.Now()
}

// GetClientInfo 获取客户端信息
func (client *SSEClient) GetClientInfo() map[string]interface{} {
	client.mu.Lock()
	defer client.mu.Unlock()

	return map[string]interface{}{
		"id":                 client.ID,
		"user_id":            client.UserID,
		"ip":                 client.IP,
		"created_at":         client.CreatedAt,
		"last_activity_at":   client.LastActivityAt,
		"message_count":      client.MessageCount,
		"bytes_sent":         client.BytesSent,
		"uptime":             time.Since(client.CreatedAt).String(),
		"connected_duration": time.Since(client.CreatedAt).Seconds(),
	}
}

