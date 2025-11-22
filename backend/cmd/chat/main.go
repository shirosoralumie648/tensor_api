package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/middleware"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
	logger "github.com/shirosoralumie648/Oblivious/backend/internal/logging"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.App.Env); err != nil {
		log.Fatal("Failed to init logger:", err)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.InitPostgres(&cfg.Database, cfg.App.Env); err != nil {
		logger.Fatal("Failed to init database", zap.Error(err))
	}
	defer database.Close()

	// 初始化 JWT
	utils.InitJWT(&cfg.JWT)

	// 初始化 Gin
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 添加中间件
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 初始化 Service
	chatService := service.NewChatService()

	// 注册路由 - 所有接口都需要鉴权
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// 创建会话
		api.POST("/chat/sessions", func(c *gin.Context) {
			userID := c.GetInt("user_id")

			var req service.CreateSessionRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			session, err := chatService.CreateSession(c.Request.Context(), userID, &req)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, session, "会话创建成功")
		})

		// 获取会话列表
		api.GET("/chat/sessions", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			sessions, total, err := chatService.GetUserSessions(c.Request.Context(), userID, page, pageSize)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, gin.H{
				"sessions": sessions,
				"total":    total,
				"page":     page,
				"pageSize": pageSize,
			}, "")
		})

		// 获取会话详情
		api.GET("/chat/sessions/:id", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			sessionID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				utils.BadRequest(c, "Invalid session ID")
				return
			}

			session, err := chatService.GetSessionByID(c.Request.Context(), sessionID, userID)
			if err != nil {
				utils.NotFound(c, err.Error())
				return
			}

			utils.Success(c, session, "")
		})

		// 更新会话
		api.PUT("/chat/sessions/:id", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			sessionID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				utils.BadRequest(c, "Invalid session ID")
				return
			}

			var req struct {
				Title string `json:"title"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			session, err := chatService.UpdateSession(c.Request.Context(), userID, sessionID, req.Title)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, session, "更新成功")
		})

		// 删除会话
		api.DELETE("/chat/sessions/:id", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			sessionID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				utils.BadRequest(c, "Invalid session ID")
				return
			}

			if err := chatService.DeleteSession(c.Request.Context(), userID, sessionID); err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, nil, "删除成功")
		})

		// 获取会话的消息列表
		api.GET("/chat/sessions/:id/messages", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			sessionID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				utils.BadRequest(c, "Invalid session ID")
				return
			}

			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

			messages, total, err := chatService.GetSessionMessages(c.Request.Context(), sessionID, userID, page, pageSize)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, gin.H{
				"messages": messages,
				"total":    total,
				"page":     page,
				"pageSize": pageSize,
			}, "")
		})

		// 发送消息（非流式）
		api.POST("/chat/messages", func(c *gin.Context) {
			userID := c.GetInt("user_id")

			var req service.SendMessageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			message, err := chatService.SendMessage(c.Request.Context(), userID, &req)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, message, "")
		})

		// 发送消息（流式 SSE）
		api.POST("/chat/messages/stream", func(c *gin.Context) {
			userID := c.GetInt("user_id")

			var req service.SendMessageRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			// 设置 SSE 响应头
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Transfer-Encoding", "chunked")

			// 获取响应写入器
			w := c.Writer

			// 通过流式服务发送消息
			if err := chatService.SendMessageStream(c.Request.Context(), userID, &req, w); err != nil {
				logger.Error("stream error", zap.Error(err))
				fmt.Fprintf(w, "event: error\n")
				fmt.Fprintf(w, "data: %s\n\n", err.Error())
				return
			}

			// 发送完成事件
			fmt.Fprintf(w, "event: done\n")
			fmt.Fprintf(w, "data: {\"status\":\"completed\"}\n\n")
		})
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	port := 8082 // 对话服务端口
	addr := fmt.Sprintf(":%d", port)
	logger.Info("Chat service starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
