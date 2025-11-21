package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/config"
	"github.com/oblivious/backend/internal/database"
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

	// 初始化 Redis（用于限流）
	if err := database.InitRedis(&cfg.Redis); err != nil {
		logger.Fatal("Failed to init redis", zap.Error(err))
	}
	defer database.CloseRedis()

	// 初始化 JWT
	utils.InitJWT(&cfg.JWT)

	// 初始化 Gin
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	// API 路由组
	api := r.Group("/api/v1")

	// 公开接口（无需鉴权）
	public := api.Group("")
	public.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
		Rate:  10,
		Burst: 20,
		TTL:   time.Minute,
	}))
	{
		// 转发到用户服务
		public.POST("/register", proxyToService(cfg.Services.UserServiceURL))
		public.POST("/login", proxyToService(cfg.Services.UserServiceURL))
		public.POST("/refresh", proxyToService(cfg.Services.UserServiceURL))
	}

	// 需要鉴权的接口
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret)))
	protected.Use(middleware.RateLimitMiddleware(&middleware.RateLimitConfig{
		Rate:  100,
		Burst: 200,
		TTL:   time.Minute,
	}))
	{
		// 用户相关
		protected.GET("/user/profile", proxyToService(cfg.Services.UserServiceURL))
		protected.PUT("/user/profile", proxyToService(cfg.Services.UserServiceURL))

		// 对话相关
		protected.POST("/chat/sessions", proxyToService(cfg.Services.ChatServiceURL))
		protected.GET("/chat/sessions", proxyToService(cfg.Services.ChatServiceURL))
		protected.GET("/chat/sessions/:id", proxyToService(cfg.Services.ChatServiceURL))
		protected.PUT("/chat/sessions/:id", proxyToService(cfg.Services.ChatServiceURL))
		protected.DELETE("/chat/sessions/:id", proxyToService(cfg.Services.ChatServiceURL))
		protected.GET("/chat/sessions/:id/messages", proxyToService(cfg.Services.ChatServiceURL))
		protected.POST("/chat/messages", proxyToService(cfg.Services.ChatServiceURL))
		protected.POST("/chat/messages/stream", proxyToServiceSSE(cfg.Services.ChatServiceURL))

		// 计费相关（TODO: 当计费服务启动后启用）
		// protected.GET("/billing/history", proxyToService(cfg.Services.BillingServiceURL))
		// protected.GET("/billing/quota-history", proxyToService(cfg.Services.BillingServiceURL))
		// protected.GET("/billing/invoices", proxyToService(cfg.Services.BillingServiceURL))
		// protected.POST("/billing/invoices", proxyToService(cfg.Services.BillingServiceURL))

		// 管理员接口 (TODO: 实现 RoleMiddleware)
		// admin := protected.Group("")
		// admin.Use(middleware.RoleMiddleware(100)) // role >= 100 才能访问
		// {
		// 	// 管理员功能...
		// }
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	logger.Info("Gateway starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// proxyToService 代理请求到目标服务
func proxyToService(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			utils.InternalError(c, "读取请求失败")
			return
		}

		// 构建目标 URL
		target := targetURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}

		// 创建新请求
		req, err := http.NewRequest(c.Request.Method, target, bytes.NewReader(body))
		if err != nil {
			utils.InternalError(c, "创建请求失败")
			return
		}

		// 复制 Header
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// 传递用户信息（如果已鉴权）
		if userID, exists := c.Get("user_id"); exists {
			req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
			req.Header.Set("X-Username", c.GetString("username"))
			req.Header.Set("X-User-Role", fmt.Sprintf("%d", c.GetInt("role")))
		}

		// 发送请求
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			utils.InternalError(c, "请求上游服务失败")
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			utils.InternalError(c, "读取响应失败")
			return
		}

		// 复制响应 Header
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// 返回响应
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

// proxyToServiceSSE 代理 SSE 流式请求到目标服务
func proxyToServiceSSE(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			utils.InternalError(c, "读取请求失败")
			return
		}

		// 构建目标 URL
		target := targetURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}

		// 创建新请求
		req, err := http.NewRequest(c.Request.Method, target, bytes.NewReader(body))
		if err != nil {
			utils.InternalError(c, "创建请求失败")
			return
		}

		// 复制 Header
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// 传递用户信息（如果已鉴权）
		if userID, exists := c.Get("user_id"); exists {
			req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
			req.Header.Set("X-Username", c.GetString("username"))
			req.Header.Set("X-User-Role", fmt.Sprintf("%d", c.GetInt("role")))
		}

		// 发送请求
		client := &http.Client{Timeout: 300 * time.Second} // SSE 需要更长的超时时间
		resp, err := client.Do(req)
		if err != nil {
			utils.InternalError(c, "请求上游服务失败")
			return
		}
		defer resp.Body.Close()

		// 设置 SSE 响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Transfer-Encoding", "chunked")

		// 复制其他响应 Header（除了 Content-Length）
		for key, values := range resp.Header {
			if key != "Content-Length" {
				for _, value := range values {
					c.Header(key, value)
				}
			}
		}

		// 流式转发响应
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			logger.Error("Failed to copy SSE response", zap.Error(err))
			return
		}
	}
}
