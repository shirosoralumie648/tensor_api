package service

import (
"context"
"fmt"
"math"
"strings"
"sync"
"time"

"gorm.io/gorm"

"github.com/shirosoralumie648/Oblivious/backend/internal/adapter"
"github.com/shirosoralumie648/Oblivious/backend/internal/model"
"github.com/shirosoralumie648/Oblivious/backend/internal/relay"
)

// HealthCheckService 健康检查服务接口
type HealthCheckService interface {
StartPeriodicCheck(ctx context.Context)
CheckChannel(ctx context.Context, channelID int) (*HealthCheckResult, error)
CalculateHealthScore(ctx context.Context, channelID int) (*HealthScore, error)
GetHealthStatus(ctx context.Context, channelID int) (*HealthStatus, error)
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
ChannelID    int       `json:"channel_id"`
Success      bool      `json:"success"`
ResponseTime int64     `json:"response_time"`
Error        string    `json:"error,omitempty"`
CheckedAt    time.Time `json:"checked_at"`
}

// HealthScore 渠道健康度评分
type HealthScore struct {
ChannelID       int     `json:"channel_id"`
SuccessRate     float64 `json:"success_rate"`
AvgResponseTime int     `json:"avg_response_time"`
RecentFailures  int     `json:"recent_failures"`
Score           float64 `json:"score"`
}

// HealthStatus 渠道健康状态
type HealthStatus struct {
ChannelID     int       `json:"channel_id"`
Status        string    `json:"status"`
FailureCount  int       `json:"failure_count"`
LastCheckTime time.Time `json:"last_check_time"`
LastSuccess   time.Time `json:"last_success"`
}

// DefaultHealthCheckService 默认实现
type DefaultHealthCheckService struct {
db               *gorm.DB
configManager    *adapter.ConfigManager
interval         time.Duration
failureThreshold int
failureCounts    map[int]int
checkHistory     map[int][]*HealthCheckResult
mu               sync.RWMutex
}

// NewHealthCheckService 创建健康检查服务
func NewHealthCheckService(
db *gorm.DB,
configManager *adapter.ConfigManager,
) *DefaultHealthCheckService {
return &DefaultHealthCheckService{
db:               db,
configManager:    configManager,
interval:         30 * time.Minute,
failureThreshold: 3,
failureCounts:    make(map[int]int),
checkHistory:     make(map[int][]*HealthCheckResult),
}
}

// StartPeriodicCheck 启动定期健康检查
func (s *DefaultHealthCheckService) StartPeriodicCheck(ctx context.Context) {
ticker := time.NewTicker(s.interval)
defer ticker.Stop()

fmt.Println("✅ Health check service started")
s.checkAllChannels(ctx)

for {
select {
case <-ticker.C:
s.checkAllChannels(ctx)
case <-ctx.Done():
fmt.Println("⏹️  Health check service stopped")
return
}

// checkAllChannels 检查所有启用的渠道
func (s *DefaultHealthCheckService) checkAllChannels(ctx context.Context) {
	var channels []model.Channel
	if err := s.db.Where("enabled = ? AND deleted_at IS NULL", true).Find(&channels).Error; err != nil {
		fmt.Printf("❌ Failed to query channels: %v\n", err)
		return
	}

	fmt.Printf("🔍 Checking %d channels...\n", len(channels))

	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(channel model.Channel) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := s.CheckChannel(ctx, channel.ID)
			if err != nil {
				fmt.Printf("❌ Health check failed for channel %s: %v\n", channel.Name, err)
				return
			}

			s.handleCheckResult(ctx, &channel, result)
		}(ch)
	}

	wg.Wait()
	fmt.Println("✅ Health check completed")
}

// CheckChannel 检查单个渠道
func (s *DefaultHealthCheckService) CheckChannel(ctx context.Context, channelID int) (*HealthCheckResult, error) {
	var channel model.Channel
	if err := s.db.First(&channel, channelID).Error; err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	apiKeys := strings.Split(channel.APIKey, ",")
	if len(apiKeys) == 0 || apiKeys[0] == "" {
		return &HealthCheckResult{
			ChannelID: channelID,
			Success:   false,
			Error:     "no API key configured",
			CheckedAt: time.Now(),
		}, nil
	}

	config := &adapter.AdapterConfig{
		Type:    channel.Type,
		BaseURL: channel.BaseURL,
		APIKey:  strings.TrimSpace(apiKeys[0]),
		Timeout: 30 * time.Second,
	}

	adapterInstance, err := s.configManager.GetAdapter(channel.Type, config)
	if err != nil {
		return &HealthCheckResult{
			ChannelID: channelID,
			Success:   false,
			Error:     fmt.Sprintf("failed to create adapter: %v", err),
			CheckedAt: time.Now(),
		}, nil
	}

	start := time.Now()

	testModel := "gpt-3.5-turbo"
	if channel.SupportModels != "" {
		models := strings.Split(channel.SupportModels, ",")
		if len(models) > 0 {
			testModel = strings.TrimSpace(models[0])
		}
	}

	testReq := &relay.ChatCompletionRequest{
		Model: testModel,
		Messages: []relay.ChatMessage{
			{Role: "user", Content: "Hi"},
		},
		MaxTokens: 5,
	}

	_, err = adapterInstance.ChatCompletion(ctx, testReq)
	duration := time.Since(start)

	result := &HealthCheckResult{
		ChannelID:    channelID,
		Success:      err == nil,
		ResponseTime: duration.Milliseconds(),
		CheckedAt:    time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
	}

	s.saveCheckResult(result)
	return result, nil
}

// handleCheckResult 处理检查结果
func (s *DefaultHealthCheckService) handleCheckResult(ctx context.Context, channel *model.Channel, result *HealthCheckResult) {
	if result.Success {
		s.resetFailureCount(channel.ID)

updates := map[string]interface{}{
"response_time": result.ResponseTime,
"test_time":     time.Now().Unix(),
}

if channel.Status == 3 {
updates["status"] = 0
fmt.Printf("✅ Channel '%s' recovered\n", channel.Name)
}

s.db.Model(&model.Channel{}).Where("id = ?", channel.ID).Updates(updates)

} else {
failCount := s.incrementFailureCount(channel.ID)

if failCount >= s.failureThreshold {
s.db.Model(&model.Channel{}).
Where("id = ?", channel.ID).
Update("status", 3)

fmt.Printf("⚠️  Channel '%s' auto-disabled (failures: %d)\n", channel.Name, failCount)
} else {
fmt.Printf("⚠️  Channel '%s' check failed (%d/%d): %s\n",
channel.Name, failCount, s.failureThreshold, result.Error)
}
}
}

func (s *DefaultHealthCheckService) saveCheckResult(result *HealthCheckResult) {
s.mu.Lock()
defer s.mu.Unlock()

history := s.checkHistory[result.ChannelID]
history = append([]*HealthCheckResult{result}, history...)

if len(history) > 100 {
history = history[:100]
}

s.checkHistory[result.ChannelID] = history
}

func (s *DefaultHealthCheckService) incrementFailureCount(channelID int) int {
s.mu.Lock()
defer s.mu.Unlock()

s.failureCounts[channelID]++
return s.failureCounts[channelID]
}

func (s *DefaultHealthCheckService) resetFailureCount(channelID int) {
s.mu.Lock()
defer s.mu.Unlock()

delete(s.failureCounts, channelID)
}

func (s *DefaultHealthCheckService) getFailureCount(channelID int) int {
s.mu.RLock()
defer s.mu.RUnlock()

return s.failureCounts[channelID]
}

// CalculateHealthScore 计算渠道健康度评分
func (s *DefaultHealthCheckService) CalculateHealthScore(ctx context.Context, channelID int) (*HealthScore, error) {
s.mu.RLock()
history := s.checkHistory[channelID]
s.mu.RUnlock()

records := history
if len(records) > 48 {
records = records[:48]
}

if len(records) == 0 {
return &HealthScore{
ChannelID: channelID,
Score:     100.0,
}, nil
}

successCount := 0
var totalResponseTime int64

for _, record := range records {
if record.Success {
successCount++
totalResponseTime += record.ResponseTime
}
}

successRate := float64(successCount) / float64(len(records))
avgResponseTime := 0
if successCount > 0 {
avgResponseTime = int(totalResponseTime / int64(successCount))
}
recentFailures := len(records) - successCount

scoreBySuccess := successRate * 70.0

scoreBySpeed := 30.0
if avgResponseTime > 0 {
scoreBySpeed = math.Max(0, 30.0*(1.0-float64(avgResponseTime-1000)/4000.0))
}

score := scoreBySuccess + scoreBySpeed

return &HealthScore{
ChannelID:       channelID,
SuccessRate:     successRate,
AvgResponseTime: avgResponseTime,
RecentFailures:  recentFailures,
Score:           score,
}, nil
}

// GetHealthStatus 获取渠道健康状态
func (s *DefaultHealthCheckService) GetHealthStatus(ctx context.Context, channelID int) (*HealthStatus, error) {
failCount := s.getFailureCount(channelID)

status := "healthy"
if failCount >= s.failureThreshold {
status = "unhealthy"
} else if failCount > 0 {
status = "degraded"
}

s.mu.RLock()
history := s.checkHistory[channelID]
s.mu.RUnlock()

var lastCheckTime, lastSuccess time.Time
if len(history) > 0 {
lastCheckTime = history[0].CheckedAt

if history[0].Success {
lastSuccess = lastCheckTime
} else {
for _, record := range history {
if record.Success {
lastSuccess = record.CheckedAt
break
}
}
}
}

return &HealthStatus{
ChannelID:     channelID,
Status:        status,
FailureCount:  failCount,
LastCheckTime: lastCheckTime,
LastSuccess:   lastSuccess,
}, nil
}
