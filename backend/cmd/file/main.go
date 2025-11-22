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
			"service": "file",
			"time":    time.Now().Unix(),
		})
	})

	// API路由
	v1 := router.Group("/api/v1")
	{
		// 文件上传
		v1.POST("/upload", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "upload endpoint"})
		})

		// 文件下载
		v1.GET("/download/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "download endpoint"})
		})

		// 文件列表
		v1.GET("/files", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "files list endpoint"})
		})
	}

	// 启动服务器
	port := 8087

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("File service started on port %d", port)

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
