package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	logger "github.com/shirosoralumie648/Oblivious/backend/internal/logging"
	"github.com/shirosoralumie648/Oblivious/backend/internal/middleware"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
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
		target, err := joinURL(targetURL, c.Request.URL.Path, c.Request.URL.RawQuery)
		if err != nil {
			utils.InternalError(c, "创建请求失败")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
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

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			utils.InternalError(c, "请求上游服务失败")
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)WORK_PLAN.md，包含从环境搭建、后端组件、Relay/适配器、计费与异步任务、知识库、Token
			}
		}

		c.Status(resp.StatusCode)
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			logger.Error("Failed to copy response", zap.Error(err))
		}
	}
}

// proxyToServiceSSE 代理 SSE 流式请求到目标服务
func proxyToServiceSSE(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		target, err := joinURL(targetURL, c.Request.URL.Path, c.Request.URL.RawQuery)
		if err != nil {
			utils.InternalError(c, "创建请求失败")
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
		if err != nil {
			utils.InternalError(c, "创建请求失败")
			return
		}

		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		if userID, exists := c.Get("user_id"); exists {
			req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
			req.Header.Set("X-Username", c.GetString("username"))
			req.Header.Set("X-User-Role", fmt.Sprintf("%d", c.GetInt("role")))
		}

		client := &http.Client{Timeout: 300 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			utils.InternalError(c, "请求上游服务失败")
			return
		}
		defer resp.Body.Close()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		for key, values := range resp.Header {
			if key == "Content-Length" {
				continue
			}
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(resp.StatusCode)
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			logger.Error("Failed to copy SSE response", zap.Error(err))
		}
	}
}

// joinURL 组装目标 URL
func joinURL(base, path, rawQuery string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = path
	u.RawQuery = rawQuery
	return u.String(), nil
}
