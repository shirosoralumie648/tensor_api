package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/shirosoralumie648/Oblivious/backend/internal/config"
	"github.com/shirosoralumie648/Oblivious/backend/internal/database"
	"github.com/shirosoralumie648/Oblivious/backend/internal/middleware"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
	"github.com/shirosoralumie648/Oblivious/backend/internal/utils"
	"github.com/shirosoralumie648/Oblivious/backend/pkg/logger"
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
	userService := service.NewUserService(&cfg.JWT)

	// 注册路由
	api := r.Group("/api/v1")
	{
		// 用户注册
		api.POST("/register", func(c *gin.Context) {
			var req service.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			user, err := userService.Register(c.Request.Context(), &req)
			if err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			utils.Success(c, user, "注册成功")
		})

		// 用户登录
		api.POST("/login", func(c *gin.Context) {
			var req service.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			resp, err := userService.Login(c.Request.Context(), &req)
			if err != nil {
				utils.Unauthorized(c, err.Error())
				return
			}

			utils.Success(c, resp, "登录成功")
		})

		// Token 刷新
		api.POST("/refresh", func(c *gin.Context) {
			var req service.RefreshTokenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			resp, err := userService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
			if err != nil {
				utils.Unauthorized(c, err.Error())
				return
			}

			utils.Success(c, resp, "Token 刷新成功")
		})
	}

	// 需要鉴权的接口
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// 获取当前用户信息
		protected.GET("/user/profile", func(c *gin.Context) {
			userID := c.GetInt("user_id")

			user, err := userService.GetUserByID(c.Request.Context(), userID)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			if user == nil {
				utils.NotFound(c, "用户不存在")
				return
			}

			utils.Success(c, user, "")
		})

		// 更新用户资料
		protected.PUT("/user/profile", func(c *gin.Context) {
			userID := c.GetInt("user_id")

			var req service.UpdateProfileRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			user, err := userService.UpdateProfile(c.Request.Context(), userID, &req)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, user, "更新成功")
		})

		// 根据 ID 获取用户信息（管理员功能）
		protected.GET("/user/:id", func(c *gin.Context) {
			// TODO: 添加管理员权限检查
			user, err := userService.GetUserByID(c.Request.Context(), 1) // 临时写死
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			if user == nil {
				utils.NotFound(c, "用户不存在")
				return
			}

			utils.Success(c, user, "")
		})
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	port := 8081 // 用户服务端口
	addr := fmt.Sprintf(":%d", port)
	logger.Info("User service starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
