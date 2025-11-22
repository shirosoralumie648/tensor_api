package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/shirosoralumie648/Oblivious/backend/internal/model"
)

// StatsHandler 统计监控Handler
type StatsHandler struct {
	db *gorm.DB
}

// NewStatsHandler 创建统计Handler
func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

// OverviewStats 总览统计
type OverviewStats struct {
	TotalChannels   int64   `json:"total_channels"`
	ActiveChannels  int64   `json:"active_channels"`
	TotalRequests   int64   `json:"total_requests"`
	TotalTokens     int64   `json:"total_tokens"`
	TotalQuota      int64   `json:"total_quota"`
	AvgResponseTime float64 `json:"avg_response_time"`
	SuccessRate     float64 `json:"success_rate"`
	Today           struct {
		Requests int64 `json:"requests"`
		Tokens   int64 `json:"tokens"`
		Quota    int64 `json:"quota"`
	} `json:"today"`
}

// GetOverview 获取总览统计
// @Summary 获取总览统计
// @Tags stats
// @Produce json
// @Success 200 {object} OverviewStats
// @Router /api/admin/stats/overview [get]
func (h *StatsHandler) GetOverview(c *gin.Context) {
	var stats OverviewStats

	// 1. 统计渠道数
	h.db.Model(&model.Channel{}).
		Where("deleted_at IS NULL").
		Count(&stats.TotalChannels)

	h.db.Model(&model.Channel{}).
		Where("deleted_at IS NULL AND enabled = ? AND status = ?", true, 0).
		Count(&stats.ActiveChannels)

	// 2. 统计今日数据
	today := time.Now().Truncate(24 * time.Hour)

	var todayLogs []model.UnifiedLog
	h.db.Where("created_at >= ?", today).Find(&todayLogs)

	for _, log := range todayLogs {
		stats.Today.Requests++
		stats.Today.Tokens += int64(log.PromptTokens + log.CompletionTokens)
		stats.Today.Quota += int64(log.Quota)
	}

	// 3. 统计总计数据
	var allLogs []model.UnifiedLog
	h.db.Select("prompt_tokens, completion_tokens, quota, use_time").
		Limit(10000). // 限制查询数量
		Find(&allLogs)

	stats.TotalRequests = int64(len(allLogs))
	var totalTime int64
	for _, log := range allLogs {
		stats.TotalTokens += int64(log.PromptTokens + log.CompletionTokens)
		stats.TotalQuota += int64(log.Quota)
		totalTime += int64(log.UseTime)
	}

	if stats.TotalRequests > 0 {
		stats.AvgResponseTime = float64(totalTime) / float64(stats.TotalRequests)
		stats.SuccessRate = 0.98 // TODO: 计算真实成功率
	}

	c.JSON(http.StatusOK, stats)
}

// ChannelStats 渠道统计
type ChannelStats struct {
	ChannelID   int     `json:"channel_id"`
	ChannelName string  `json:"channel_name"`
	Type        string  `json:"type"`
	Status      int     `json:"status"`
	Enabled     bool    `json:"enabled"`
	Requests    int64   `json:"requests"`
	Tokens      int64   `json:"tokens"`
	Quota       int64   `json:"quota"`
	AvgLatency  float64 `json:"avg_latency"`
	SuccessRate float64 `json:"success_rate"`
}

// GetChannelStats 获取渠道统计
// @Summary 获取渠道统计
// @Tags stats
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {array} ChannelStats
// @Router /api/admin/stats/channels [get]
func (h *StatsHandler) GetChannelStats(c *gin.Context) {
	days := c.DefaultQuery("days", "7")

	// 计算起始时间
	startTime := time.Now().AddDate(0, 0, -7)
	if d, err := time.ParseDuration(days + "d"); err == nil {
		startTime = time.Now().Add(-d)
	}

	// 查询所有渠道
	var channels []model.Channel
	h.db.Where("deleted_at IS NULL").Find(&channels)

	stats := make([]ChannelStats, 0, len(channels))

	for _, channel := range channels {
		stat := ChannelStats{
			ChannelID:   channel.ID,
			ChannelName: channel.Name,
			Type:        channel.Type,
			Status:      channel.Status,
			Enabled:     channel.Enabled,
		}

		// 查询该渠道的日志
		var logs []model.UnifiedLog
		h.db.Where("channel_id = ? AND created_at >= ?", channel.ID, startTime).
			Find(&logs)

		stat.Requests = int64(len(logs))
		var totalTime int64
		for _, log := range logs {
			stat.Tokens += int64(log.PromptTokens + log.CompletionTokens)
			stat.Quota += int64(log.Quota)
			totalTime += int64(log.UseTime)
		}

		if stat.Requests > 0 {
			stat.AvgLatency = float64(totalTime) / float64(stat.Requests)
			stat.SuccessRate = 0.98 // TODO: 计算真实成功率
		}

		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, stats)
}

// ModelStats 模型统计
type ModelStats struct {
	Model      string  `json:"model"`
	Requests   int64   `json:"requests"`
	Tokens     int64   `json:"tokens"`
	Quota      int64   `json:"quota"`
	AvgLatency float64 `json:"avg_latency"`
}

// GetModelStats 获取模型统计
// @Summary 获取模型统计
// @Tags stats
// @Produce json
// @Param days query int false "统计天数" default(7)
// @Success 200 {array} ModelStats
// @Router /api/admin/stats/models [get]
func (h *StatsHandler) GetModelStats(c *gin.Context) {
	days := 7
	startTime := time.Now().AddDate(0, 0, -days)

	// 按模型分组统计
	type ModelGroup struct {
		ModelName string
		Count     int64
		Tokens    int64
		Quota     int64
		AvgTime   float64
	}

	var results []ModelGroup
	h.db.Model(&model.UnifiedLog{}).
		Select("model_name, COUNT(*) as count, SUM(prompt_tokens + completion_tokens) as tokens, SUM(quota) as quota, AVG(use_time) as avg_time").
		Where("created_at >= ?", startTime).
		Group("model_name").
		Scan(&results)

	stats := make([]ModelStats, 0, len(results))
	for _, r := range results {
		stats = append(stats, ModelStats{
			Model:      r.ModelName,
			Requests:   r.Count,
			Tokens:     r.Tokens,
			Quota:      r.Quota,
			AvgLatency: r.AvgTime,
		})
	}

	c.JSON(http.StatusOK, stats)
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Date     string `json:"date"`
	Requests int64  `json:"requests"`
	Tokens   int64  `json:"tokens"`
	Quota    int64  `json:"quota"`
}

// GetTimeSeries 获取时间序列数据
// @Summary 获取时间序列数据
// @Tags stats
// @Produce json
// @Param days query int false "统计天数" default(30)
// @Success 200 {array} TimeSeriesData
// @Router /api/admin/stats/timeseries [get]
func (h *StatsHandler) GetTimeSeries(c *gin.Context) {
	days := 30
	startTime := time.Now().AddDate(0, 0, -days)

	// 按日期分组统计
	type DailyGroup struct {
		Date   string
		Count  int64
		Tokens int64
		Quota  int64
	}

	var results []DailyGroup
	h.db.Model(&model.UnifiedLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count, SUM(prompt_tokens + completion_tokens) as tokens, SUM(quota) as quota").
		Where("created_at >= ?", startTime).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results)

	series := make([]TimeSeriesData, 0, len(results))
	for _, r := range results {
		series = append(series, TimeSeriesData{
			Date:     r.Date,
			Requests: r.Count,
			Tokens:   r.Tokens,
			Quota:    r.Quota,
		})
	}

	c.JSON(http.StatusOK, series)
}

// RegisterRoutes 注册路由
func (h *StatsHandler) RegisterRoutes(r *gin.RouterGroup) {
	stats := r.Group("/stats")
	{
		stats.GET("/overview", h.GetOverview)
		stats.GET("/channels", h.GetChannelStats)
		stats.GET("/models", h.GetModelStats)
		stats.GET("/timeseries", h.GetTimeSeries)
	}
}
