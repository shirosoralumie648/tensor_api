package relay

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SSEHandler SSE 处理器
type SSEHandler struct {
	manager *SSEManager
	timeout time.Duration
}

// NewSSEHandler 创建新的 SSE 处理器
func NewSSEHandler(manager *SSEManager) *SSEHandler {
	return &SSEHandler{
		manager: manager,
		timeout: 30 * time.Second,
	}
}

// SetTimeout 设置超时时间
func (h *SSEHandler) SetTimeout(timeout time.Duration) {
	h.timeout = timeout
}

// HandleSSEConnect 处理 SSE 连接
// 使用方式: GET /api/sse/connect
func (h *SSEHandler) HandleSSEConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户 ID (从上下文或 token 中获取)
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userIDStr := fmt.Sprintf("%v", userID)

		// 生成客户端 ID
		clientID := uuid.New().String()

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 注册客户端
		client, err := h.manager.RegisterClient(clientID, userIDStr, clientIP)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 设置响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Client-ID", clientID)
		c.Header("X-Accel-Buffering", "no")

		// 获取响应写入器
		w := c.Writer

		// 设置分块编码
		w.(http.Flusher).Flush()

		// 发送连接成功消息
		fmt.Fprintf(w, "id: %s\n", clientID)
		fmt.Fprintf(w, "event: connected\n")
		fmt.Fprintf(w, "data: {\"client_id\":\"%s\",\"timestamp\":%d}\n\n", clientID, time.Now().UnixMilli())
		w.(http.Flusher).Flush()

		// 监听消息和连接关闭
		for {
			select {
			case msg := <-client.MessageChan:
				// 发送消息
				if err := h.writeMessage(w, msg); err != nil {
					// 错误发生，关闭连接
					h.manager.UnregisterClient(clientID)
					return
				}
				w.(http.Flusher).Flush()

			case <-client.CloseChan:
				// 客户端已关闭
				return

			case <-c.Request.Context().Done():
				// HTTP 连接已关闭
				h.manager.UnregisterClient(clientID)
				return

			case <-time.After(h.timeout):
				// 超时检查（可选）
				client.UpdateActivity()
			}
		}
	}
}

// HandleSSEDisconnect 处理 SSE 断开连接
// 使用方式: POST /api/sse/disconnect/:clientID
func (h *SSEHandler) HandleSSEDisconnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.Param("clientID")

		h.manager.UnregisterClient(clientID)

		c.JSON(http.StatusOK, gin.H{
			"message": "disconnected",
		})
	}
}

// HandleSSEBroadcast 广播消息给所有客户端
// 使用方式: POST /api/sse/broadcast
// 需要管理员权限
func (h *SSEHandler) HandleSSEBroadcast() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 权限检查（需要管理员权限）
		// 这里假设已通过权限中间件检查

		var req struct {
			Event string `json:"event" binding:"required"`
			Data  string `json:"data" binding:"required"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		msg := &SSEMessage{
			ID:    uuid.New().String(),
			Event: req.Event,
			Data:  req.Data,
		}

		h.manager.BroadcastMessage(msg)

		c.JSON(http.StatusOK, gin.H{
			"message": "broadcast sent",
		})
	}
}

// HandleSendToClient 发送消息给特定客户端
// 使用方式: POST /api/sse/send/:clientID
func (h *SSEHandler) HandleSendToClient() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.Param("clientID")

		var req struct {
			Event string `json:"event"`
			Data  string `json:"data" binding:"required"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		msg := &SSEMessage{
			ID:    uuid.New().String(),
			Event: req.Event,
			Data:  req.Data,
		}

		if err := h.manager.SendMessageToClient(clientID, msg); err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "message sent",
		})
	}
}

// HandleStatistics 获取 SSE 统计信息
// 使用方式: GET /api/sse/stats
func (h *SSEHandler) HandleStatistics() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := h.manager.GetStatistics()

		c.JSON(http.StatusOK, gin.H{
			"statistics": stats,
		})
	}
}

// writeMessage 写入 SSE 消息
func (h *SSEHandler) writeMessage(w http.ResponseWriter, msg *SSEMessage) error {
	// 写入消息 ID
	if msg.ID != "" {
		fmt.Fprintf(w, "id: %s\n", msg.ID)
	}

	// 写入事件类型
	if msg.Event != "" {
		fmt.Fprintf(w, "event: %s\n", msg.Event)
	}

	// 写入重试时间
	if msg.Retry > 0 {
		fmt.Fprintf(w, "retry: %d\n", msg.Retry)
	}

	// 写入注释
	if msg.Comment != "" {
		fmt.Fprintf(w, ": %s\n", msg.Comment)
	}

	// 写入数据
	if msg.Data != "" {
		// 处理多行数据
		fmt.Fprintf(w, "data: %s\n", msg.Data)
	}

	// 发送空行作为消息分隔符
	fmt.Fprintf(w, "\n")

	return nil
}

// RegisterSSERoutes 注册 SSE 路由
func (h *SSEHandler) RegisterSSERoutes(router *gin.Engine) {
	sseGroup := router.Group("/api/sse")

	sseGroup.GET("/connect", h.HandleSSEConnect())
	sseGroup.POST("/disconnect/:clientID", h.HandleSSEDisconnect())
	sseGroup.POST("/broadcast", h.HandleSSEBroadcast())
	sseGroup.POST("/send/:clientID", h.HandleSendToClient())
	sseGroup.GET("/stats", h.HandleStatistics())
}

// RegisterSSERoutesWithAuth 使用认证的方式注册 SSE 路由
func (h *SSEHandler) RegisterSSERoutesWithAuth(
	router *gin.Engine,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
) {
	sseGroup := router.Group("/api/sse")
	sseGroup.Use(authMiddleware)

	sseGroup.GET("/connect", h.HandleSSEConnect())
	sseGroup.POST("/disconnect/:clientID", h.HandleSSEDisconnect())
	sseGroup.POST("/broadcast", adminMiddleware, h.HandleSSEBroadcast())
	sseGroup.POST("/send/:clientID", h.HandleSendToClient())
	sseGroup.GET("/stats", adminMiddleware, h.HandleStatistics())
}

