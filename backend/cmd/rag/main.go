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
			"service": "rag",
			"time":    time.Now().Unix(),
		})
	})

	// API路由
	v1 := router.Group("/api/v1")
	{
		// 向量化
		v1.POST("/embed", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "embed endpoint"})
		})

		// 向量搜索
		v1.POST("/search", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "search endpoint"})
		})

		// 索引文档
		v1.POST("/index", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "index endpoint"})
		})
	}

	// 启动服务器
	port := 8089

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("RAG service started on port %d", port)

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
