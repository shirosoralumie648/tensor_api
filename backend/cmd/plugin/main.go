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
)

func main() {
	router := gin.Default()

	// 全局中间件
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "plugin",
			"time":    time.Now().Unix(),
		})
	})

	// API路由
	v1 := router.Group("/api/v1")
	{
		// 插件列表
		v1.GET("/plugins", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "plugins list endpoint"})
		})

		// 插件详情
		v1.GET("/plugins/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "plugin detail endpoint"})
		})

		// 执行插件
		v1.POST("/plugins/:id/execute", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "plugin execute endpoint"})
		})
	}

	// 启动服务器
	port := 8088

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Plugin service started on port %d", port)

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
