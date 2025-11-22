package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/handler"
	"github.com/shirosoralumie648/Oblivious/backend/internal/middleware"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建服务
	billingService := service.NewBillingService()

	// 创建处理器
	billingHandler := handler.NewBillingHandler(billingService)

	// 设置路由
	router := gin.Default()

	// 全局中间件
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "billing",
			"time":    time.Now().Unix(),
		})
	})

	// API路由
	v1 := router.Group("/api/v1")
	{
		// 需要认证
		authFactory := middleware.NewAuthFactory([]byte(cfg.JWT.Secret), nil)
		v1.Use(authFactory.JWT())

		// 计费相关
		billing := v1.Group("/billing")
		{
			billing.GET("/logs", billingHandler.GetBillingLogs)
			billing.GET("/logs/:id", billingHandler.GetBillingLog)
			billing.POST("/refund/:id", billingHandler.Refund)
		}

		// 配额相关
		quota := v1.Group("/quota")
		{
			quota.GET("/logs", billingHandler.GetQuotaLogs)
			quota.POST("/recharge", billingHandler.Recharge)
		}
	}

	// 启动服务器
	port := cfg.Services.Billing.Port
	if port == 0 {
		port = 8084
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Billing service started on port %d", port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
