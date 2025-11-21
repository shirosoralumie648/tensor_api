package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/adapter"
	"github.com/oblivious/backend/internal/middleware"
)

// ChatCompletionRequest 聊天完成请求
type ChatCompletionRequest struct {
	Model       string                 `json:"model" binding:"required"`
	Messages    []adapter.Message      `json:"messages" binding:"required"`
	Temperature float32                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	TopP        float32                `json:"top_p,omitempty"`
	TopK        int                    `json:"top_k,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	User        string                 `json:"user,omitempty"`
}

// ChatHandler 处理聊天相关请求的处理器
type ChatHandler struct {
	adapterFactory  *adapter.AdapterFactory
	billingService  BillingService
	channelService  ChannelService
	auditService    AuditService
}

// BillingService 计费服务接口
type BillingService interface {
	RecordUsage(record *BillingRecord) error
	CalculateCost(userID, model string, promptTokens, completionTokens int) (float32, error)
}

// ChannelService 渠道服务接口
type ChannelService interface {
	GetChannel(channelID string) *Channel
	GetDefaultChannel() *Channel
}

// AuditService 审计服务接口
type AuditService interface {
	RecordAPICall(record *AuditRecord) error
}

// Channel 渠道信息
type Channel struct {
	ID       string
	Name     string
	Provider string
	APIKey   string
}

// BillingRecord 计费记录
type BillingRecord struct {
	UserID           string
	Model            string
	PromptTokens     int
	CompletionTokens int
	Cost             float32
	Timestamp        time.Time
}

// AuditRecord 审计记录
type AuditRecord struct {
	UserID    string
	Endpoint  string
	Method    string
	Model     string
	Status    int
	Timestamp time.Time
}

// NewChatHandler 创建新的聊天处理器
func NewChatHandler(
	factory *adapter.AdapterFactory,
	billing BillingService,
	channel ChannelService,
	audit AuditService,
) *ChatHandler {
	return &ChatHandler{
		adapterFactory: factory,
		billingService: billing,
		channelService: channel,
		auditService:   audit,
	}
}

// ChatCompletion 处理非流式聊天请求
// @Summary 聊天完成 (非流式)
// @Description 发送消息并获取完整响应
// @Tags Chat
// @Accept json
// @Produce json
// @Param request body ChatCompletionRequest true "请求体"
// @Success 200 {object} adapter.ChatResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /v1/chat/completions [post]
func (h *ChatHandler) ChatCompletion(c *gin.Context) {
	// 验证用户
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.TopP == 0 {
		req.TopP = 1.0
	}

	// 获取渠道信息
	channel := h.channelService.GetDefaultChannel()
	if channel == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no channel available"})
		return
	}

	// 获取适配器
	provider := h.adapterFactory.Get(channel.Provider)
	if provider == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("provider %s not supported", channel.Provider),
		})
		return
	}

	// 构建适配器请求
	adapterReq := &adapter.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      false,
		User:        userID,
	}

	// 调用适配器
	startTime := time.Now()
	resp, err := provider.Chat(c.Request.Context(), adapterReq)
	if err != nil {
		// 记录审计信息
		h.auditService.RecordAPICall(&AuditRecord{
			UserID:    userID,
			Endpoint:  "/v1/chat/completions",
			Method:    "POST",
			Model:     req.Model,
			Status:    http.StatusInternalServerError,
			Timestamp: time.Now(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录计费信息
	cost, _ := h.billingService.CalculateCost(
		userID,
		req.Model,
		resp.Tokens.PromptTokens,
		resp.Tokens.CompletionTokens,
	)

	h.billingService.RecordUsage(&BillingRecord{
		UserID:           userID,
		Model:            req.Model,
		PromptTokens:     resp.Tokens.PromptTokens,
		CompletionTokens: resp.Tokens.CompletionTokens,
		Cost:             cost,
		Timestamp:        time.Now(),
	})

	// 记录审计信息
	h.auditService.RecordAPICall(&AuditRecord{
		UserID:    userID,
		Endpoint:  "/v1/chat/completions",
		Method:    "POST",
		Model:     req.Model,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})

	// 返回响应
	resp.ResponseTime = time.Since(startTime).Milliseconds()
	c.JSON(http.StatusOK, resp)
}

// ChatCompletionStream 处理流式聊天请求
// @Summary 聊天完成 (流式)
// @Description 发送消息并获取流式响应
// @Tags Chat
// @Accept json
// @Produce text/event-stream
// @Param request body ChatCompletionRequest true "请求体"
// @Success 200 {object} adapter.StreamDelta
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /v1/chat/completions [post]
func (h *ChatHandler) ChatCompletionStream(c *gin.Context) {
	// 验证用户
	userID, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.TopP == 0 {
		req.TopP = 1.0
	}

	// 获取渠道信息
	channel := h.channelService.GetDefaultChannel()
	if channel == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no channel available"})
		return
	}

	// 获取适配器
	provider := h.adapterFactory.Get(channel.Provider)
	if provider == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("provider %s not supported", channel.Provider),
		})
		return
	}

	// 构建适配器请求
	adapterReq := &adapter.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      true,
		User:        userID,
	}

	// 获取流式响应
	deltaCh, err := provider.ChatStream(c.Request.Context(), adapterReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头用于 Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 记录审计信息
	h.auditService.RecordAPICall(&AuditRecord{
		UserID:    userID,
		Endpoint:  "/v1/chat/completions",
		Method:    "POST",
		Model:     req.Model,
		Status:    http.StatusOK,
		Timestamp: time.Now(),
	})

	totalPromptTokens := 0
	totalCompletionTokens := 0

	// 流式发送响应
	for delta := range deltaCh {
		if delta.Error != nil {
			// 错误处理
			errorData := gin.H{"error": delta.Error.Error()}
			data, _ := json.Marshal(errorData)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			break
		}

		// 收集 token 统计
		if delta.Tokens.PromptTokens > 0 {
			totalPromptTokens = delta.Tokens.PromptTokens
		}
		if delta.Tokens.CompletionTokens > 0 {
			totalCompletionTokens += delta.Tokens.CompletionTokens
		}

		// 发送增量数据
		data, _ := json.Marshal(delta)
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
		c.Writer.Flush()

		// 如果完成，发送最后的消息
		if delta.Done {
			// 记录计费信息
			cost, _ := h.billingService.CalculateCost(
				userID,
				req.Model,
				totalPromptTokens,
				totalCompletionTokens,
			)

			h.billingService.RecordUsage(&BillingRecord{
				UserID:           userID,
				Model:            req.Model,
				PromptTokens:     totalPromptTokens,
				CompletionTokens: totalCompletionTokens,
				Cost:             cost,
				Timestamp:        time.Now(),
			})

			// 发送完成标记
			fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
			break
		}
	}
}

// ListModels 获取所有可用模型列表
// @Summary 获取模型列表
// @Description 获取所有可用的 AI 模型列表
// @Tags Models
// @Produce json
// @Success 200 {object} map[string][]adapter.Model
// @Failure 401 {object} gin.H
// @Router /v1/models [get]
func (h *ChatHandler) ListModels(c *gin.Context) {
	// 验证用户
	_, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	allModels := h.adapterFactory.GetAllModels()
	c.JSON(http.StatusOK, allModels)
}

// GetModel 获取单个模型详细信息
// @Summary 获取模型详情
// @Description 根据模型 ID 获取模型详细信息
// @Tags Models
// @Param model_id path string true "模型 ID"
// @Produce json
// @Success 200 {object} adapter.Model
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /v1/models/:model_id [get]
func (h *ChatHandler) GetModel(c *gin.Context) {
	// 验证用户
	_, err := middleware.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	modelID := c.Param("model_id")
	model := h.adapterFactory.GetModelByID(modelID)

	if model == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, model)
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查所有 AI 提供商的连接状态
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /health/chat [get]
func (h *ChatHandler) HealthCheck(c *gin.Context) {
	status := make(map[string]bool)

	for _, providerName := range h.adapterFactory.List() {
		provider := h.adapterFactory.Get(providerName)
		err := provider.HealthCheck(c.Request.Context())
		status[providerName] = err == nil
	}

	c.JSON(http.StatusOK, status)
}

