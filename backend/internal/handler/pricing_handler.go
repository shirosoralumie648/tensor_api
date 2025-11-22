package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
)

// PricingHandler 定价管理Handler
type PricingHandler struct {
	pricingService service.PricingService
}

// NewPricingHandler 创建定价Handler
func NewPricingHandler(pricingService service.PricingService) *PricingHandler {
	return &PricingHandler{
		pricingService: pricingService,
	}
}

// ListPricing 列出所有定价
// @Summary 列出所有定价
// @Tags pricing
// @Produce json
// @Param enabled query bool false "是否启用"
// @Success 200 {array} model.ModelPricing
// @Router /api/v1/pricing [get]
func (h *PricingHandler) ListPricing(c *gin.Context) {
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		val := enabledStr == "true"
		enabled = &val
	}

	pricings, err := h.pricingService.ListPricing(c.Request.Context(), enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pricings)
}

// GetPricing 获取指定模型的定价
// @Summary 获取模型定价
// @Tags pricing
// @Produce json
// @Param model path string true "模型名称"
// @Param group query string false "分组" default(default)
// @Success 200 {object} model.ModelPricing
// @Router /api/v1/pricing/{model} [get]
func (h *PricingHandler) GetPricing(c *gin.Context) {
	modelName := c.Param("model")
	group := c.DefaultQuery("group", "default")

	pricing, err := h.pricingService.GetPricing(c.Request.Context(), modelName, group)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pricing)
}

// CreatePricingRequest 创建定价请求
type CreatePricingRequest struct {
	Model           string   `json:"model" binding:"required"`
	Group           string   `json:"group"`
	QuotaType       int      `json:"quota_type"` // 0:按量 1:按次
	ModelPrice      *float64 `json:"model_price"`
	ModelRatio      *float64 `json:"model_ratio"`
	CompletionRatio float64  `json:"completion_ratio"`
	GroupRatio      float64  `json:"group_ratio"`
	VendorID        string   `json:"vendor_id"`
	Enabled         bool     `json:"enabled"`
	Description     string   `json:"description"`
}

// CreatePricing 创建定价
// @Summary 创建定价
// @Tags pricing
// @Accept json
// @Produce json
// @Param pricing body CreatePricingRequest true "定价信息"
// @Success 201 {object} model.ModelPricing
// @Router /api/v1/pricing [post]
func (h *PricingHandler) CreatePricing(c *gin.Context) {
	var req CreatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认值
	if req.Group == "" {
		req.Group = "default"
	}
	if req.CompletionRatio == 0 {
		req.CompletionRatio = 1.0
	}
	if req.GroupRatio == 0 {
		req.GroupRatio = 1.0
	}

	pricing := &model.ModelPricing{
		Model:           req.Model,
		Group:           req.Group,
		QuotaType:       req.QuotaType,
		ModelPrice:      req.ModelPrice,
		ModelRatio:      req.ModelRatio,
		CompletionRatio: req.CompletionRatio,
		GroupRatio:      req.GroupRatio,
		VendorID:        req.VendorID,
		Enabled:         req.Enabled,
		Description:     req.Description,
	}

	if err := h.pricingService.CreatePricing(c.Request.Context(), pricing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pricing)
}

// UpdatePricingRequest 更新定价请求
type UpdatePricingRequest struct {
	QuotaType       *int     `json:"quota_type"`
	ModelPrice      *float64 `json:"model_price"`
	ModelRatio      *float64 `json:"model_ratio"`
	CompletionRatio *float64 `json:"completion_ratio"`
	GroupRatio      *float64 `json:"group_ratio"`
	Enabled         *bool    `json:"enabled"`
	Description     *string  `json:"description"`
}

// UpdatePricing 更新定价
// @Summary 更新定价
// @Tags pricing
// @Accept json
// @Produce json
// @Param id path int true "定价ID"
// @Param pricing body UpdatePricingRequest true "更新信息"
// @Success 200 {object} model.ModelPricing
// @Router /api/v1/pricing/{id} [put]
func (h *PricingHandler) UpdatePricing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdatePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pricing := &model.ModelPricing{}
	if req.QuotaType != nil {
		pricing.QuotaType = *req.QuotaType
	}
	if req.ModelPrice != nil {
		pricing.ModelPrice = req.ModelPrice
	}
	if req.ModelRatio != nil {
		pricing.ModelRatio = req.ModelRatio
	}
	if req.CompletionRatio != nil {
		pricing.CompletionRatio = *req.CompletionRatio
	}
	if req.GroupRatio != nil {
		pricing.GroupRatio = *req.GroupRatio
	}
	if req.Enabled != nil {
		pricing.Enabled = *req.Enabled
	}
	if req.Description != nil {
		pricing.Description = *req.Description
	}

	if err := h.pricingService.UpdatePricing(c.Request.Context(), id, pricing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pricing updated successfully"})
}

// DeletePricing 删除定价
// @Summary 删除定价
// @Tags pricing
// @Param id path int true "定价ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/pricing/{id} [delete]
func (h *PricingHandler) DeletePricing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.pricingService.DeletePricing(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pricing deleted successfully"})
}

// CalculateQuotaRequest 计算配额请求
type CalculateQuotaRequest struct {
	Model            string `json:"model" binding:"required"`
	Group            string `json:"group"`
	PromptTokens     int    `json:"prompt_tokens" binding:"required"`
	CompletionTokens int    `json:"completion_tokens" binding:"required"`
}

// CalculateQuotaResponse 计算配额响应
type CalculateQuotaResponse struct {
	Model            string  `json:"model"`
	Group            string  `json:"group"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	Quota            int     `json:"quota"`
	GroupRatio       float64 `json:"group_ratio"`
}

// CalculateQuota 计算配额
// @Summary 计算配额
// @Tags pricing
// @Accept json
// @Produce json
// @Param request body CalculateQuotaRequest true "计算请求"
// @Success 200 {object} CalculateQuotaResponse
// @Router /api/v1/pricing/calculate [post]
func (h *PricingHandler) CalculateQuota(c *gin.Context) {
	var req CalculateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Group == "" {
		req.Group = "default"
	}

	quota, err := h.pricingService.CalculateQuota(
		c.Request.Context(),
		req.Model,
		req.Group,
		req.PromptTokens,
		req.CompletionTokens,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取分组倍率（如果service实现了该方法）
	groupRatio := 1.0
	if ps, ok := h.pricingService.(*service.DefaultPricingService); ok {
		groupRatio = ps.GetGroupRatio(req.Group)
	}

	c.JSON(http.StatusOK, CalculateQuotaResponse{
		Model:            req.Model,
		Group:            req.Group,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		Quota:            quota,
		GroupRatio:       groupRatio,
	})
}

// RefreshCache 刷新定价缓存
// @Summary 刷新定价缓存
// @Tags pricing
// @Success 200 {object} map[string]string
// @Router /api/v1/pricing/refresh [post]
func (h *PricingHandler) RefreshCache(c *gin.Context) {
	if err := h.pricingService.RefreshCache(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cache refreshed successfully"})
}

// RegisterRoutes 注册路由
func (h *PricingHandler) RegisterRoutes(r *gin.RouterGroup) {
	pricing := r.Group("/pricing")
	{
		pricing.GET("", h.ListPricing)
		pricing.GET("/:model", h.GetPricing)
		pricing.POST("", h.CreatePricing)
		pricing.PUT("/:id", h.UpdatePricing)
		pricing.DELETE("/:id", h.DeletePricing)
		pricing.POST("/calculate", h.CalculateQuota)
		pricing.POST("/refresh", h.RefreshCache)
	}
}
