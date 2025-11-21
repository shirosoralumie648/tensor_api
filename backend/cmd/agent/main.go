package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/config"
	"github.com/oblivious/backend/internal/database"
	"github.com/oblivious/backend/internal/handler"
	"github.com/oblivious/backend/internal/middleware"
	"github.com/oblivious/backend/internal/utils"
	"github.com/oblivious/backend/pkg/logger"
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

	// 初始化 Handler
	agentHandler := handler.NewAgentHandler()

	// 注册路由 - 所有接口都需要鉴权
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// 创建助手
		api.POST("/agents", agentHandler.CreateAgent)

		// 获取用户的助手列表
		api.GET("/agents/user", agentHandler.GetUserAgents)

		// 获取公开的助手列表
		api.GET("/agents/public", agentHandler.GetPublicAgents)

		// 获取精选助手
		api.GET("/agents/featured", agentHandler.GetFeaturedAgents)

		// 搜索助手
		api.GET("/agents/search", agentHandler.SearchAgents)

		// 获取助手详情
		api.GET("/agents/:id", agentHandler.GetAgent)

		// 获取助手统计
		api.GET("/agents/:id/stats", agentHandler.GetAgentStats)

		// 赞助手
		api.POST("/agents/:id/like", agentHandler.LikeAgent)

		// 复制助手
		api.POST("/agents/:id/fork", agentHandler.ForkAgent)

		// 更新助手
		api.PUT("/agents/:id", agentHandler.UpdateAgent)

		// 删除助手
		api.DELETE("/agents/:id", agentHandler.DeleteAgent)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	port := 8084 // 助手服务端口
	addr := fmt.Sprintf(":%d", port)
	logger.Info("Agent service starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

