package performance

import (
	"runtime"
	"testing"
	"time"
)

// TestPerformanceOptimizer 测试性能优化器
func TestPerformanceOptimizer(t *testing.T) {
	optimizer := NewPerformanceOptimizer()

	// 测试初始化
	if optimizer.dbPoolSize != 20 {
		t.Errorf("Expected dbPoolSize 20, got %d", optimizer.dbPoolSize)
	}

	if optimizer.goroutineLimit != 1000 {
		t.Errorf("Expected goroutineLimit 1000, got %d", optimizer.goroutineLimit)
	}
}

// TestGetMemoryStats 测试内存统计
func TestGetMemoryStats(t *testing.T) {
	optimizer := NewPerformanceOptimizer()

	stats := optimizer.GetMemoryStats()

	if stats.Alloc == 0 {
		t.Error("Expected non-zero Alloc")
	}

	if stats.TotalAlloc == 0 {
		t.Error("Expected non-zero TotalAlloc")
	}
}

// TestGetGoroutineCount 测试 Goroutine 数量
func TestGetGoroutineCount(t *testing.T) {
	optimizer := NewPerformanceOptimizer()

	count := optimizer.GetGoroutineCount()

	if count <= 0 {
		t.Errorf("Expected positive goroutine count, got %d", count)
	}

	if count != runtime.NumGoroutine() {
		t.Errorf("GetGoroutineCount() mismatch with runtime.NumGoroutine()")
	}
}

// TestGetCPUStats 测试 CPU 统计
func TestGetCPUStats(t *testing.T) {
	optimizer := NewPerformanceOptimizer()

	stats := optimizer.GetCPUStats()

	if _, ok := stats["num_cpu"]; !ok {
		t.Error("Expected num_cpu in stats")
	}

	if _, ok := stats["goroutine_count"]; !ok {
		t.Error("Expected goroutine_count in stats")
	}

	if _, ok := stats["num_goroutine_max"]; !ok {
		t.Error("Expected num_goroutine_max in stats")
	}
}

// TestCheckMemoryUsage 测试内存使用检查
func TestCheckMemoryUsage(t *testing.T) {
	optimizer := NewPerformanceOptimizer()

	used, limit, percentage := optimizer.CheckMemoryUsage()

	if used == 0 {
		t.Error("Expected non-zero used memory")
	}

	if limit == 0 {
		t.Error("Expected non-zero memory limit")
	}

	if percentage < 0 || percentage > 100 {
		t.Errorf("Expected percentage between 0 and 100, got %f", percentage)
	}
}

// TestOptimizeMemory 测试内存优化
func TestOptimizeMemory(t *testing.T) {
	optimizer := NewPerformanceOptimizer()

	// 这只是测试是否会 panic
	optimizer.OptimizeMemory()
}

// TestDatabaseOptimization 测试数据库优化配置
func TestDatabaseOptimization(t *testing.T) {
	dbOpt := NewDatabaseOptimization()

	if dbOpt.PoolSize != 20 {
		t.Errorf("Expected PoolSize 20, got %d", dbOpt.PoolSize)
	}

	if dbOpt.MaxOverflow != 40 {
		t.Errorf("Expected MaxOverflow 40, got %d", dbOpt.MaxOverflow)
	}

	if dbOpt.MaxIdleTime != 5*time.Minute {
		t.Errorf("Expected MaxIdleTime 5m, got %v", dbOpt.MaxIdleTime)
	}
}

// TestCacheOptimization 测试缓存优化配置
func TestCacheOptimization(t *testing.T) {
	cacheOpt := NewCacheOptimization()

	if cacheOpt.MaxSize != 100000 {
		t.Errorf("Expected MaxSize 100000, got %d", cacheOpt.MaxSize)
	}

	if cacheOpt.EvictionPolicy != "LRU" {
		t.Errorf("Expected EvictionPolicy LRU, got %s", cacheOpt.EvictionPolicy)
	}
}

// TestConcurrencyOptimization 测试并发优化配置
func TestConcurrencyOptimization(t *testing.T) {
	concOpt := NewConcurrencyOptimization()

	if concOpt.MaxGoroutines != 10000 {
		t.Errorf("Expected MaxGoroutines 10000, got %d", concOpt.MaxGoroutines)
	}

	if concOpt.RequestTimeout != 30*time.Second {
		t.Errorf("Expected RequestTimeout 30s, got %v", concOpt.RequestTimeout)
	}
}

// TestNetworkOptimization 测试网络优化配置
func TestNetworkOptimization(t *testing.T) {
	netOpt := NewNetworkOptimization()

	if !netOpt.EnableHTTP2 {
		t.Error("Expected EnableHTTP2 to be true")
	}

	if !netOpt.EnableKeepAlive {
		t.Error("Expected EnableKeepAlive to be true")
	}

	if netOpt.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns 100, got %d", netOpt.MaxIdleConns)
	}
}

// TestPerformanceMonitor 测试性能监控器
func TestPerformanceMonitor(t *testing.T) {
	monitor := &PerformanceMonitor{}

	// 记录请求
	monitor.RecordRequest(100 * time.Millisecond)
	monitor.RecordRequest(200 * time.Millisecond)

	// 记录缓存
	monitor.RecordCacheHit()
	monitor.RecordCacheHit()
	monitor.RecordCacheMiss()

	// 记录错误
	monitor.RecordError()

	// 获取统计
	stats := monitor.GetStats()

	if stats["request_count"] != int64(2) {
		t.Errorf("Expected request_count 2, got %d", stats["request_count"])
	}

	if stats["error_count"] != int64(1) {
		t.Errorf("Expected error_count 1, got %d", stats["error_count"])
	}

	if stats["cache_hits"] != int64(2) {
		t.Errorf("Expected cache_hits 2, got %d", stats["cache_hits"])
	}

	if stats["cache_misses"] != int64(1) {
		t.Errorf("Expected cache_misses 1, got %d", stats["cache_misses"])
	}

	cacheHitRate := stats["cache_hit_rate_percent"].(float64)
	expectedRate := float64(2) / float64(3) * 100
	if cacheHitRate < expectedRate-1 || cacheHitRate > expectedRate+1 {
		t.Errorf("Expected cache_hit_rate_percent ~66.67, got %f", cacheHitRate)
	}
}

// BenchmarkPerformanceMonitor 性能基准测试
func BenchmarkPerformanceMonitor(b *testing.B) {
	monitor := &PerformanceMonitor{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordRequest(time.Millisecond)
		if i%2 == 0 {
			monitor.RecordCacheHit()
		} else {
			monitor.RecordCacheMiss()
		}
	}
}

// BenchmarkMemoryStats 内存统计基准测试
func BenchmarkMemoryStats(b *testing.B) {
	optimizer := NewPerformanceOptimizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.GetMemoryStats()
	}
}

