package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 包含所有 Prometheus 指标
type Metrics struct {
	// HTTP 指标
	HTTPRequestsTotal     prometheus.CounterVec
	HTTPRequestDuration   prometheus.HistogramVec
	HTTPResponseSize      prometheus.HistogramVec
	HTTPErrorsTotal       prometheus.CounterVec

	// API 调用指标
	APICallsTotal         prometheus.CounterVec
	APICallDuration       prometheus.HistogramVec
	APITokensUsed         prometheus.HistogramVec
	APICostTotal          prometheus.CounterVec

	// 缓存指标
	CacheHits             prometheus.CounterVec
	CacheMisses           prometheus.CounterVec
	CacheEvictions        prometheus.CounterVec
	CacheSize             prometheus.GaugeVec

	// 数据库指标
	DBConnectionsActive   prometheus.GaugeVec
	DBQueryDuration       prometheus.HistogramVec
	DBQueryErrors         prometheus.CounterVec
	DBRowsAffected        prometheus.HistogramVec

	// 队列指标
	QueueSize             prometheus.GaugeVec
	QueueMessages         prometheus.CounterVec
	QueueProcessingTime   prometheus.HistogramVec
	QueueErrors           prometheus.CounterVec

	// 系统指标
	GoroutineCount        prometheus.Gauge
	MemoryUsage           prometheus.GaugeVec
	CPUUsage              prometheus.Gauge
	UptimeSeconds         prometheus.Gauge

	// 业务指标
	ActiveSessions        prometheus.GaugeVec
	TotalUsers            prometheus.GaugeVec
	QuotaUsage            prometheus.GaugeVec
	ErrorRate             prometheus.GaugeVec
}

// NewMetrics 创建新的 Metrics 实例
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP 指标
		HTTPRequestsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "path"},
		),
		HTTPErrorsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors",
			},
			[]string{"method", "path", "error_type"},
		),

		// API 调用指标
		APICallsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_calls_total",
				Help: "Total number of API calls",
			},
			[]string{"model", "provider", "status"},
		),
		APICallDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_call_duration_seconds",
				Help:    "API call latency in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"model", "provider"},
		),
		APITokensUsed: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_tokens_used",
				Help:    "Tokens used per API call",
				Buckets: []float64{10, 100, 1000, 10000, 100000},
			},
			[]string{"model"},
		),
		APICostTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_cost_total",
				Help: "Total API cost in dollars",
			},
			[]string{"model", "provider"},
		),

		// 缓存指标
		CacheHits: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMisses: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type"},
		),
		CacheEvictions: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_evictions_total",
				Help: "Total number of cache evictions",
			},
			[]string{"cache_type"},
		),
		CacheSize: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cache_size_bytes",
				Help: "Cache size in bytes",
			},
			[]string{"cache_type"},
		),

		// 数据库指标
		DBConnectionsActive: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
			[]string{"database"},
		),
		DBQueryDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query latency in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
			},
			[]string{"operation"},
		),
		DBQueryErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_query_errors_total",
				Help: "Total number of database query errors",
			},
			[]string{"operation", "error_type"},
		),
		DBRowsAffected: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_rows_affected",
				Help:    "Number of rows affected per query",
				Buckets: []float64{1, 10, 100, 1000, 10000},
			},
			[]string{"operation"},
		),

		// 队列指标
		QueueSize: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "queue_size",
				Help: "Queue size (number of messages)",
			},
			[]string{"queue_name"},
		),
		QueueMessages: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "queue_messages_total",
				Help: "Total number of messages processed",
			},
			[]string{"queue_name", "status"},
		),
		QueueProcessingTime: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "queue_processing_time_seconds",
				Help:    "Queue message processing time in seconds",
				Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
			},
			[]string{"queue_name"},
		),
		QueueErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "queue_errors_total",
				Help: "Total number of queue processing errors",
			},
			[]string{"queue_name", "error_type"},
		),

		// 系统指标
		GoroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutine_count",
				Help: "Number of goroutines",
			},
		),
		MemoryUsage: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),
		UptimeSeconds: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "uptime_seconds",
				Help: "Application uptime in seconds",
			},
		),

		// 业务指标
		ActiveSessions: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_sessions",
				Help: "Number of active sessions",
			},
			[]string{"user_type"},
		),
		TotalUsers: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "total_users",
				Help: "Total number of users",
			},
			[]string{"status"},
		),
		QuotaUsage: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "quota_usage_percent",
				Help: "Quota usage percentage",
			},
			[]string{"quota_type"},
		),
		ErrorRate: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "error_rate_percent",
				Help: "Error rate percentage",
			},
			[]string{"service"},
		),
	}
}

