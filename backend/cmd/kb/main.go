package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/handler"
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

	// 获取 Embedding API 配置
	embeddingURL := os.Getenv("EMBEDDING_API_URL")
	if embeddingURL == "" {
		// 默认使用本地 Relay 服务的 embedding 端点
		embeddingURL = "http://relay:8083/v1/embeddings"
	}
	embeddingKey := os.Getenv("EMBEDDING_API_KEY")

	// 初始化服务
	ragService := service.NewRAGService(embeddingURL, embeddingKey)
	kbHandler := handler.NewKBHandler(ragService)

	// 注册路由 - 所有接口都需要鉴权
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// 知识库管理
		api.POST("/knowledge-bases", kbHandler.CreateKnowledgeBase)
		api.GET("/knowledge-bases", kbHandler.ListKnowledgeBases)
		api.GET("/knowledge-bases/:id", kbHandler.GetKnowledgeBase)
		api.DELETE("/knowledge-bases/:id", kbHandler.DeleteKnowledgeBase)

		// 文档管理
		api.POST("/knowledge-bases/:id/documents", kbHandler.UploadDocument)
		api.GET("/knowledge-bases/:id/documents", kbHandler.GetDocumentList)
		api.DELETE("/knowledge-bases/:id/documents/:doc_id", kbHandler.DeleteDocument)

		// 搜索
		api.POST("/knowledge-bases/:id/search", kbHandler.SearchDocuments)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	port := 8085 // 知识库服务端口
	addr := fmt.Sprintf(":%d", port)
	logger.Info("Knowledge Base service starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
