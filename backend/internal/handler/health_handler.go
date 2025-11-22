package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/shirosoralumie648/Oblivious/backend/internal/service"
)

// HealthHandler 健康检查Handler
type HealthHandler struct {
	healthService service.HealthCheckService
}

// NewHealthHandler 创建健康Handler
func NewHealthHandler(healthService service.HealthCheckService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// CheckChannel 检查指定渠道
// @Summary 检查渠道健康状态
// @Tags health
// @Param id path int true "渠道ID"
// @Success 200 {object} service.HealthCheckResult
// @Router /api/admin/health/channels/{id} [post]
func (h *HealthHandler) CheckChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result, err := h.healthService.CheckChannel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetHealthStatus 获取渠道健康状态
// @Summary 获取渠道健康状态
// @Tags health
// @Param id path int true "渠道ID"
// @Success 200 {object} service.HealthStatus
// @Router /api/admin/health/channels/{id}/status [get]
func (h *HealthHandler) GetHealthStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	status, err := h.healthService.GetHealthStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetHealthScore 获取渠道健康评分
// @Summary 获取渠道健康评分
// @Tags health
// @Param id path int true "渠道ID"
// @Success 200 {object} service.HealthScore
// @Router /api/admin/health/channels/{id}/score [get]
func (h *HealthHandler) GetHealthScore(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	score, err := h.healthService.CalculateHealthScore(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, score)
}

// RegisterRoutes 注册路由
func (h *HealthHandler) RegisterRoutes(r *gin.RouterGroup) {
	health := r.Group("/health/channels")
	{
		health.POST("/:id", h.CheckChannel)
		health.GET("/:id/status", h.GetHealthStatus)
		health.GET("/:id/score", h.GetHealthScore)
	}
}
