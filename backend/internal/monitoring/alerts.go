package monitoring

import "time"

// AlertRule 定义告警规则
type AlertRule struct {
	Name        string
	Description string
	Query       string        // Prometheus PromQL 查询
	Threshold   float64       // 阈值
	Duration    time.Duration // 保持时长
	Severity    string        // critical, warning, info
	Action      string        // 告警动作
}

// GetAlertRules 返回所有告警规则
func GetAlertRules() []AlertRule {
	return []AlertRule{
		// 服务可用性告警
		{
			Name:        "ServiceDown",
			Description: "服务不可用",
			Query:       "up == 0",
			Duration:    1 * time.Minute,
			Severity:    "critical",
			Action:      "立即告警+通知运维",
		},
		{
			Name:        "HighErrorRate",
			Description: "错误率过高 (>5%)",
			Query:       "rate(http_errors_total[5m]) / rate(http_requests_total[5m]) > 0.05",
			Duration:    5 * time.Minute,
			Severity:    "critical",
			Action:      "立即告警+日志分析",
		},

		// API 响应时间告警
		{
			Name:        "HighLatency",
			Description: "API 响应时间过长 (p99 > 1s)",
			Query:       "histogram_quantile(0.99, http_request_duration_seconds) > 1",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "记录告警+性能优化",
		},
		{
			Name:        "APITimeoutRate",
			Description: "API 超时率过高 (>2%)",
			Query:       "rate(http_errors_total{error_type=\"timeout\"}[5m]) > 0.02",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警+检查依赖服务",
		},

		// 缓存告警
		{
			Name:        "LowCacheHitRate",
			Description: "缓存命中率过低 (<80%)",
			Query:       "cache_hits_total / (cache_hits_total + cache_misses_total) < 0.8",
			Duration:    10 * time.Minute,
			Severity:    "warning",
			Action:      "记录告警+优化缓存策略",
		},
		{
			Name:        "HighCacheEviction",
			Description: "缓存驱逐频繁 (>100/min)",
			Query:       "rate(cache_evictions_total[1m]) > 100",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警+检查内存",
		},

		// 数据库告警
		{
			Name:        "DBConnPoolExhausted",
			Description: "数据库连接池耗尽 (>90%)",
			Query:       "db_connections_active / 100 > 0.9",
			Duration:    2 * time.Minute,
			Severity:    "critical",
			Action:      "立即告警+检查慢查询",
		},
		{
			Name:        "DBQueryLatency",
			Description: "数据库查询延迟过高 (p99 > 500ms)",
			Query:       "histogram_quantile(0.99, db_query_duration_seconds) > 0.5",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "记录告警+索引优化",
		},
		{
			Name:        "DBErrorRate",
			Description: "数据库错误率过高 (>1%)",
			Query:       "rate(db_query_errors_total[5m]) > 0.01",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警+检查数据库状态",
		},

		// 队列告警
		{
			Name:        "QueueBacklog",
			Description: "队列积压过多 (>10000 messages)",
			Query:       "queue_size > 10000",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警+增加消费者",
		},
		{
			Name:        "QueueProcessingDelay",
			Description: "队列处理延迟过高 (>10s)",
			Query:       "histogram_quantile(0.99, queue_processing_time_seconds) > 10",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "记录告警+优化处理",
		},
		{
			Name:        "QueueErrorRate",
			Description: "队列错误率过高 (>1%)",
			Query:       "rate(queue_errors_total[5m]) / rate(queue_messages_total[5m]) > 0.01",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警+查看错误日志",
		},

		// API 配额告警
		{
			Name:        "QuotaExhausted",
			Description: "配额即将耗尽 (>90%)",
			Query:       "quota_usage_percent > 90",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警通知用户+建议充值",
		},
		{
			Name:        "CostAnomaly",
			Description: "成本异常增长 (>200% of avg)",
			Query:       "api_cost_total / avg_over_time(api_cost_total[7d]) > 2",
			Duration:    10 * time.Minute,
			Severity:    "warning",
			Action:      "告警+查看成本明细",
		},

		// 系统资源告警
		{
			Name:        "HighMemoryUsage",
			Description: "内存使用过高 (>85%)",
			Query:       "memory_usage_bytes{type=\"rss\"} / memory_usage_bytes{type=\"total\"} > 0.85",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "记录告警+监控内存泄漏",
		},
		{
			Name:        "HighCPUUsage",
			Description: "CPU 使用过高 (>80%)",
			Query:       "cpu_usage_percent > 80",
			Duration:    5 * time.Minute,
			Severity:    "warning",
			Action:      "告警+检查热路径",
		},
		{
			Name:        "GoroutineLeaking",
			Description: "Goroutine 数量异常增长 (>2x baseline)",
			Query:       "goroutine_count / avg_over_time(goroutine_count[1h]) > 2",
			Duration:    15 * time.Minute,
			Severity:    "warning",
			Action:      "告警+检查 Goroutine 泄漏",
		},

		// 用户行为告警
		{
			Name:        "UnusualTraffic",
			Description: "流量异常 (>2x baseline)",
			Query:       "rate(http_requests_total[5m]) / avg_over_time(rate(http_requests_total[5m])[1h]) > 2",
			Duration:    5 * time.Minute,
			Severity:    "info",
			Action:      "记录告警+关注趋势",
		},
		{
			Name:        "LowActiveUsers",
			Description: "活跃用户过少 (<100)",
			Query:       "active_sessions < 100",
			Duration:    30 * time.Minute,
			Severity:    "info",
			Action:      "记录告警+业务分析",
		},
	}
}

// NotificationChannel 通知渠道
type NotificationChannel struct {
	Type      string // email, slack, pagerduty, webhook
	Endpoint  string
	Enabled   bool
	Template  string
}

// GetNotificationChannels 返回所有通知渠道
func GetNotificationChannels() []NotificationChannel {
	return []NotificationChannel{
		{
			Type:     "email",
			Endpoint: "devops@company.com",
			Enabled:  true,
			Template: "{{.Alert.Name}}: {{.Alert.Description}}",
		},
		{
			Type:     "slack",
			Endpoint: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
			Enabled:  true,
			Template: ":warning: {{.Alert.Severity}}: {{.Alert.Name}}",
		},
		{
			Type:     "pagerduty",
			Endpoint: "https://events.pagerduty.com/v2/enqueue",
			Enabled:  false,
			Template: "{{.Alert.Description}}",
		},
		{
			Type:     "webhook",
			Endpoint: "https://monitoring.company.com/alerts",
			Enabled:  true,
			Template: "{}",
		},
	}
}

