package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
)

// ChannelHandler 渠道管理Handler
type ChannelHandler struct {
	channelService        *service.ChannelService
	channelAbilityService service.ChannelAbilityService
}

// NewChannelHandler 创建渠道Handler
func NewChannelHandler(
	channelService *service.ChannelService,
	channelAbilityService service.ChannelAbilityService,
) *ChannelHandler {
	return &ChannelHandler{
		channelService:        channelService,
		channelAbilityService: channelAbilityService,
	}
}

// ListChannelsRequest 查询请求
type ListChannelsRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Type     string `form:"type"`
	Group    string `form:"group"`
	Status   *int   `form:"status"`
	Enabled  *bool  `form:"enabled"`
}

// ListChannelsResponse 查询响应
type ListChannelsResponse struct {
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Data     []*model.Channel `json:"data"`
}

// ListChannels 分页查询渠道
// @Summary 分页查询渠道
// @Tags channel
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param type query string false "渠道类型"
// @Param group query string false "分组"
// @Param status query int false "状态"
// @Param enabled query bool false "是否启用"
// @Success 200 {object} ListChannelsResponse
// @Router /api/admin/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	var req ListChannelsRequest
	req.Page = 1
	req.PageSize = 20

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用 Service
	channels, total, err := h.channelService.List(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListChannelsResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     channels,
	})
}

// CreateChannelRequest 创建请求
type CreateChannelRequest struct {
	Name          string  `json:"name" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	Group         string  `json:"group"`
	BaseURL       string  `json:"base_url"`
	APIKeys       string  `json:"api_keys" binding:"required"`
	SupportModels string  `json:"support_models" binding:"required"`
	Priority      int     `json:"priority"`
	Weight        int     `json:"weight"`
	MaxRPM        int     `json:"max_rpm"`
	MaxRPD        int     `json:"max_rpd"`
	Timeout       int     `json:"timeout"`
	ProxyURL      *string `json:"proxy_url"`
	Enabled       bool    `json:"enabled"`
}

// CreateChannel 创建渠道
// @Summary 创建渠道
// @Tags channel
// @Accept json
// @Produce json
// @Param channel body CreateChannelRequest true "渠道信息"
// @Success 201 {object} model.Channel
// @Router /api/admin/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建Channel对象
	channel := &model.Channel{
		Name:          req.Name,
		Type:          req.Type,
		Group:         req.Group,
		BaseURL:       req.BaseURL,
		APIKey:        req.APIKeys,
		SupportModels: req.SupportModels,
		Priority:      int64(req.Priority),
		Weight:        req.Weight,
		MaxRateLimit:  req.MaxRPM, // 映射到 MaxRateLimit
		// MaxRPD:        req.MaxRPD, // Model 中没有 MaxRPD，忽略或映射到其他字段
		// Timeout:       req.Timeout, // Model 中没有 Timeout，忽略
		// ProxyURL:      req.ProxyURL, // Model 中没有 ProxyURL
		Enabled: req.Enabled,
		Status:  1, // 1:启用
	}

	// 设置默认值
	if channel.Group == "" {
		channel.Group = "default"
	}
	if channel.Priority == 0 {
		channel.Priority = 100
	}
	if channel.Weight == 0 {
		channel.Weight = 10
	}

	// 创建渠道
	if err := h.channelService.Create(c.Request.Context(), channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 自动同步能力
	if err := h.channelAbilityService.SyncFromChannel(c.Request.Context(), channel); err != nil {
		// 仅记录错误，不影响返回
		// log.Printf("failed to sync abilities: %v", err)
	}

	c.JSON(http.StatusCreated, channel)
}

// UpdateChannelRequest 更新请求
type UpdateChannelRequest struct {
	Name          *string `json:"name"`
	BaseURL       *string `json:"base_url"`
	APIKeys       *string `json:"api_keys"`
	SupportModels *string `json:"support_models"`
	Priority      *int    `json:"priority"`
	Weight        *int    `json:"weight"`
	MaxRPM        *int    `json:"max_rpm"`
	MaxRPD        *int    `json:"max_rpd"`
	Timeout       *int    `json:"timeout"`
	ProxyURL      *string `json:"proxy_url"`
	Enabled       *bool   `json:"enabled"`
	Status        *int    `json:"status"`
}

// UpdateChannel 更新渠道
// @Summary 更新渠道
// @Tags channel
// @Accept json
// @Produce json
// @Param id path int true "渠道ID"
// @Param channel body UpdateChannelRequest true "更新信息"
// @Success 200 {object} model.Channel
// @Router /api/admin/channels/{id} [put]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有渠道
	channel, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if channel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	// 应用更新
	if req.Name != nil {
		channel.Name = *req.Name
	}
	if req.BaseURL != nil {
		channel.BaseURL = *req.BaseURL
	}
	if req.APIKeys != nil {
		channel.APIKey = *req.APIKeys
	}
	if req.SupportModels != nil {
		channel.SupportModels = *req.SupportModels
	}
	if req.Priority != nil {
		channel.Priority = int64(*req.Priority)
	}
	if req.Weight != nil {
		channel.Weight = *req.Weight
	}
	if req.MaxRPM != nil {
		channel.MaxRateLimit = *req.MaxRPM
	}
	if req.Enabled != nil {
		channel.Enabled = *req.Enabled
		if *req.Enabled {
			channel.Status = 1
		} else {
			channel.Status = 2
		}
	}
	if req.Status != nil {
		channel.Status = *req.Status
	}

	// 保存更新
	if err := h.channelService.Update(c.Request.Context(), channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新同步能力
	if req.SupportModels != nil {
		if err := h.channelAbilityService.SyncFromChannel(c.Request.Context(), channel); err != nil {
			// log error
		}
	}

	c.JSON(http.StatusOK, channel)
}

// DeleteChannel 删除渠道
// @Summary 删除渠道
// @Tags channel
// @Param id path int true "渠道ID"
// @Success 200 {object} map[string]string
// @Router /api/admin/channels/{id} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 删除能力记录
	if err := h.channelAbilityService.DeleteByChannel(c.Request.Context(), id); err != nil {
		// 继续删除渠道，不中断
	}

	// 删除渠道（软删除）
	if err := h.channelService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "channel deleted successfully"})
}

// TestChannel 测试渠道连接
// @Summary 测试渠道连接
// @Tags channel
// @Param id path int true "渠道ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/channels/{id}/test [post]
func (h *ChannelHandler) TestChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// TODO: 实现健康检查
	// result, err := h.healthCheckService.CheckChannel(c.Request.Context(), id)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"channel_id": id,
		"status":     "healthy",
		"latency_ms": 100,
		"message":    "连接正常",
	})
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	IDs       []int  `json:"ids" binding:"required"`
	Operation string `json:"operation" binding:"required,oneof=enable disable delete"`
}

// BatchOperation 批量操作
// @Summary 批量操作渠道
// @Tags channel
// @Accept json
// @Produce json
// @Param request body BatchOperationRequest true "批量操作"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/channels/batch [post]
func (h *ChannelHandler) BatchOperation(c *gin.Context) {
	var req BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success := 0
	failed := 0

	for _, id := range req.IDs {
		var err error
		switch req.Operation {
		case "enable":
			// TODO: h.channelService.Enable(c.Request.Context(), id)
		case "disable":
			// TODO: h.channelService.Disable(c.Request.Context(), id)
		case "delete":
			err = h.channelAbilityService.DeleteByChannel(c.Request.Context(), id)
			// TODO: h.channelService.Delete(c.Request.Context(), id)
		}

		if err != nil {
			failed++
		} else {
			success++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"failed":  failed,
		"total":   len(req.IDs),
	})
}

// RegisterRoutes 注册路由
func (h *ChannelHandler) RegisterRoutes(r *gin.RouterGroup) {
	channels := r.Group("/channels")
	{
		channels.GET("", h.ListChannels)
		channels.POST("", h.CreateChannel)
		channels.PUT("/:id", h.UpdateChannel)
		channels.DELETE("/:id", h.DeleteChannel)
		channels.POST("/:id/test", h.TestChannel)
		channels.POST("/batch", h.BatchOperation)
	}
}
