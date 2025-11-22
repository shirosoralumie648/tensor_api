package analytics

import (
	"fmt"
	"sort"
	"time"
)

// AnalyticsAPI 数据分析 API
type AnalyticsAPI struct {
	logger  *UsageLogger
	realtime *RealtimeStatsEngine
}

// NewAnalyticsAPI 创建分析 API
func NewAnalyticsAPI(logger *UsageLogger, realtime *RealtimeStatsEngine) *AnalyticsAPI {
	return &AnalyticsAPI{
		logger:   logger,
		realtime: realtime,
	}
}

// QueryRequest 查询请求
type QueryRequest struct {
	UserID    string
	StartTime time.Time
	EndTime   time.Time
	Model     string
	Provider  string
	Status    string
	Limit     int
	Offset    int
}

// QueryResponse 查询响应
type QueryResponse struct {
	Total   int64
	Records []*UsageRecord
}

// Query 多维查询
func (aa *AnalyticsAPI) Query(req *QueryRequest) (*QueryResponse, error) {
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Limit > 10000 {
		req.Limit = 10000
	}

	filter := &UsageFilter{
		UserID:    req.UserID,
		Model:     req.Model,
		Provider:  req.Provider,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	records, err := aa.logger.Query(filter)
	if err != nil {
		return nil, err
	}

	return &QueryResponse{
		Total:   int64(len(records)),
		Records: records,
	}, nil
}

// GetCostBreakdown 获取成本分解
func (aa *AnalyticsAPI) GetCostBreakdown(userID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats, err := aa.logger.GetAggregatedStats(&AggregationFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   "model",
	})
	if err != nil {
		return nil, err
	}

	breakdown := map[string]interface{}{
		"total_cost":        stats.TotalCost,
		"total_requests":    stats.TotalRequests,
		"average_cost":      stats.AvgCostPerReq,
		"by_model":          nil,
		"by_provider":       nil,
		"by_status":         nil,
	}

	return breakdown, nil
}

// GetUsageTimeline 获取使用时间线
func (aa *AnalyticsAPI) GetUsageTimeline(userID string, startTime, endTime time.Time, granularity string) ([]*TimelinePoint, error) {
	var groupBy string
	switch granularity {
	case "hour":
		groupBy = "hour"
	case "day":
		groupBy = "day"
	default:
		groupBy = "day"
	}

	stats, err := aa.logger.GetAggregatedStats(&AggregationFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   groupBy,
	})
	if err != nil {
		return nil, err
	}

	return stats.Timeline, nil
}

// GetTopModels 获取热门模型
func (aa *AnalyticsAPI) GetTopModels(userID string, startTime, endTime time.Time, limit int) (map[string]*ModelUsage, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	stats, err := aa.logger.GetAggregatedStats(&AggregationFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		GroupBy:   "model",
	})
	if err != nil {
		return nil, err
	}

	result := make(map[string]*ModelUsage)
	
	// 构建模型使用统计
	if len(stats.Timeline) > 0 {
		for _, point := range stats.Timeline {
			model := fmt.Sprintf("model_%d", point.Timestamp.Unix())
			result[model] = &ModelUsage{
				Model:    model,
				Requests: point.Requests,
				Tokens:   point.Tokens,
				Cost:     point.Cost,
			}
		}
	}

	return result, nil
}

// ModelUsage 模型使用情况
type ModelUsage struct {
	Model    string  `json:"model"`
	Requests int64   `json:"requests"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}

// GetErrorStats 获取错误统计
func (aa *AnalyticsAPI) GetErrorStats(userID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 查询错误记录
	records, err := aa.logger.Query(&UsageFilter{
		UserID:    userID,
		Status:    "error",
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     10000,
	})
	if err != nil {
		return nil, err
	}

	// 统计错误类型
	errorTypes := make(map[string]int64)
	for _, record := range records {
		errorTypes[record.ErrorMsg]++
	}

	return map[string]interface{}{
		"total_errors": int64(len(records)),
		"by_type":      errorTypes,
	}, nil
}

// GetQuotaStatus 获取配额状态
func (aa *AnalyticsAPI) GetQuotaStatus(userID string, dailyQuota, monthlyQuota int64) (map[string]interface{}, error) {
	now := time.Now()

	// 获取今日统计
	dailyStats, err := aa.logger.GetDailyStats(userID, now)
	if err != nil {
		return nil, err
	}

	// 获取本月统计
	monthlyStats, err := aa.logger.GetMonthlyStats(userID, now.Year(), now.Month())
	if err != nil {
		return nil, err
	}

	dailyPercentage := float64(0)
	if dailyQuota > 0 {
		dailyPercentage = float64(dailyStats.TotalCost) / float64(dailyQuota) * 100
	}

	monthlyPercentage := float64(0)
	if monthlyQuota > 0 {
		monthlyPercentage = float64(monthlyStats.TotalCost) / float64(monthlyQuota) * 100
	}

	return map[string]interface{}{
		"daily": map[string]interface{}{
			"used":       dailyStats.TotalCost,
			"quota":      dailyQuota,
			"remaining":  dailyQuota - int64(dailyStats.TotalCost),
			"percentage": dailyPercentage,
		},
		"monthly": map[string]interface{}{
			"used":       monthlyStats.TotalCost,
			"quota":      monthlyQuota,
			"remaining":  monthlyQuota - int64(monthlyStats.TotalCost),
			"percentage": monthlyPercentage,
		},
	}, nil
}

// GetRealtimeMetrics 获取实时指标
func (aa *AnalyticsAPI) GetRealtimeMetrics(userID string) map[string]interface{} {
	stats := aa.realtime.GetUserStats(userID)

	return map[string]interface{}{
		"timestamp":    stats.Timestamp,
		"requests":     stats.Requests,
		"success":      stats.SuccessRequests,
		"errors":       stats.ErrorRequests,
		"error_rate":   fmt.Sprintf("%.2f%%", stats.ErrorRate*100),
		"qps":          fmt.Sprintf("%.2f", stats.QPS),
		"avg_duration": stats.AvgDuration,
		"total_tokens": stats.TotalTokens,
		"total_cost":   stats.TotalCost,
	}
}

// GetModelComparison 获取模型对比
func (aa *AnalyticsAPI) GetModelComparison(userID string, models []string, startTime, endTime time.Time) (map[string]*ModelComparison, error) {
	result := make(map[string]*ModelComparison)

	for _, model := range models {
		stats, err := aa.logger.GetAggregatedStats(&AggregationFilter{
			UserID:    userID,
			StartTime: startTime,
			EndTime:   endTime,
		})
		if err != nil {
			continue
		}

		result[model] = &ModelComparison{
			Model:           model,
			Requests:        stats.TotalRequests,
			AvgDuration:     stats.AvgDuration,
			AvgTokens:       stats.AvgTokensPerReq,
			AvgCost:         stats.AvgCostPerReq,
			TotalCost:       stats.TotalCost,
			SuccessRate:     float64(stats.SuccessRequests) / float64(stats.TotalRequests),
		}
	}

	return result, nil
}

// ModelComparison 模型对比
type ModelComparison struct {
	Model       string  `json:"model"`
	Requests    int64   `json:"requests"`
	AvgDuration int64   `json:"avg_duration"`
	AvgTokens   int64   `json:"avg_tokens"`
	AvgCost     float64 `json:"avg_cost"`
	TotalCost   float64 `json:"total_cost"`
	SuccessRate float64 `json:"success_rate"`
}

// ExportData 导出数据
func (aa *AnalyticsAPI) ExportData(userID string, startTime, endTime time.Time, format string) (interface{}, error) {
	records, err := aa.logger.Query(&UsageFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     100000,
	})
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return records, nil
	case "csv":
		return aa.convertToCSV(records), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// convertToCSV 转换为 CSV
func (aa *AnalyticsAPI) convertToCSV(records []*UsageRecord) string {
	csv := "ID,UserID,TokenID,Model,Provider,RequestTokens,ResponseTokens,TotalTokens,Cost,Duration,Status,ClientIP,Timestamp\n"

	for _, record := range records {
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%d,%d,%d,%.4f,%d,%s,%s,%s\n",
			record.ID, record.UserID, record.TokenID, record.Model, record.Provider,
			record.RequestTokens, record.ResponseTokens, record.TotalTokens,
			record.Cost, record.Duration, record.Status, record.ClientIP,
			record.Timestamp.Format(time.RFC3339))
		csv += line
	}

	return csv
}

// GetProviderComparison 获取提供商对比
func (aa *AnalyticsAPI) GetProviderComparison(userID string, startTime, endTime time.Time) (map[string]*ProviderStats, error) {
	result := make(map[string]*ProviderStats)

	records, err := aa.logger.Query(&UsageFilter{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     100000,
	})
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if _, exists := result[record.Provider]; !exists {
			result[record.Provider] = &ProviderStats{
				Provider: record.Provider,
			}
		}

		stats := result[record.Provider]
		stats.Requests++
		stats.TotalCost += record.Cost
		stats.TotalTokens += record.TotalTokens

		if record.Status == "success" {
			stats.SuccessRequests++
		}
	}

	// 计算衍生指标
	for _, stats := range result {
		if stats.Requests > 0 {
			stats.AvgCost = stats.TotalCost / float64(stats.Requests)
			stats.SuccessRate = float64(stats.SuccessRequests) / float64(stats.Requests)
		}
	}

	return result, nil
}

// ProviderStats 提供商统计
type ProviderStats struct {
	Provider         string  `json:"provider"`
	Requests         int64   `json:"requests"`
	SuccessRequests  int64   `json:"success_requests"`
	TotalCost        float64 `json:"total_cost"`
	AvgCost          float64 `json:"avg_cost"`
	TotalTokens      int64   `json:"total_tokens"`
	SuccessRate      float64 `json:"success_rate"`
}


