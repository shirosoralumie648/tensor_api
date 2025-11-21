package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/config"
	"github.com/oblivious/backend/internal/database"
	"github.com/oblivious/backend/internal/middleware"
	"github.com/oblivious/backend/internal/relay"
	"github.com/oblivious/backend/internal/service"
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

	// 初始化服务
	relayService := service.NewRelayService()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由组
	api := r.Group("/v1")

	// 公开接口 - 中转 OpenAI 兼容的 API
	{
		// Chat Completion 接口（支持流式和非流式）
		api.POST("/chat/completions", func(c *gin.Context) {
			var req relay.ChatCompletionRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				utils.BadRequest(c, err.Error())
				return
			}

			// 检查 stream 参数
			if req.Stream {
				// 流式响应 - Week 7 实现
				c.Header("Content-Type", "text/event-stream")
				c.Header("Cache-Control", "no-cache")
				c.Header("Connection", "keep-alive")
				c.Header("Transfer-Encoding", "chunked")

				w := c.Writer

				err := relayService.RelayChatCompletionStream(c.Request.Context(), &req, func(chunk *relay.ChatCompletionResponse) error {
					// 格式化 SSE 数据
					if len(chunk.Choices) > 0 {
						data, _ := json.Marshal(chunk)
						fmt.Fprintf(w, "data: %s\n\n", string(data))
						if f, ok := w.(http.Flusher); ok {
							f.Flush()
						}
					}
					return nil
				})

				if err != nil {
					logger.Error("stream error", zap.Error(err))
					fmt.Fprintf(w, "event: error\n")
					fmt.Fprintf(w, "data: %s\n\n", err.Error())
					return
				}

				// 发送完成标记
				fmt.Fprintf(w, "data: [DONE]\n\n")
				return
			}

			// 非流式响应
			resp, err := relayService.RelayChatCompletion(c.Request.Context(), &req)
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, resp, "")
		})

		// 列出可用模型
		api.GET("/models", func(c *gin.Context) {
			channels, err := relayService.GetAvailableChannels(c.Request.Context())
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			// 构建模型列表
			modelMap := make(map[string]bool)
			var models []string

			for _, ch := range channels {
				if ch.SupportModels == "" {
					// 如果没有指定，默认支持所有常见模型
					defaultModels := []string{
						"gpt-3.5-turbo",
						"gpt-4",
						"gpt-4-turbo",
						"gpt-4o",
						"claude-3-opus",
						"claude-3-sonnet",
						"gemini-pro",
					}
					for _, m := range defaultModels {
						if !modelMap[m] {
							models = append(models, m)
							modelMap[m] = true
						}
					}
				}
			}

			utils.Success(c, gin.H{
				"object": "list",
				"data": models,
			}, "")
		})

		// 获取渠道列表（用于管理）
		api.GET("/channels", func(c *gin.Context) {
			channels, err := relayService.GetAvailableChannels(c.Request.Context())
			if err != nil {
				utils.InternalError(c, err.Error())
				return
			}

			utils.Success(c, channels, "")
		})
	}

	// 需要鉴权的管理接口
	admin := api.Group("")
	admin.Use(middleware.AuthMiddleware())
	{
		// 获取模型价格（用于计费）
		admin.GET("/model-price/:channel_id/:model", func(c *gin.Context) {
			channelID := c.Param("channel_id")
			modelName := c.Param("model")

			// TODO: 实现获取价格的逻辑
			utils.Success(c, gin.H{
				"channel_id":   channelID,
				"model":        modelName,
				"input_price":  0.0001,
				"output_price": 0.0003,
			}, "")
		})
	}

	// 启动服务
	port := 8083 // 中转服务端口
	addr := fmt.Sprintf(":%d", port)
	logger.Info("Relay service starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

