package relay

import (
	"testing"
	"time"
)

func TestHealthChecker(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "test")
	cache.AddChannel(ch)

	config := DefaultHealthCheckConfig()
	config.Interval = 100 * time.Millisecond

	checker := NewHealthChecker(cache, config)
	defer checker.Stop()

	var resultReceived *HealthCheckResult
	checker.SetResultCallback(func(r *HealthCheckResult) {
		resultReceived = r
	})

	checker.Start()

	// 等待检查完成
	time.Sleep(200 * time.Millisecond)

	// 应该收到结果
	if resultReceived == nil {
		t.Errorf("Expected health check result")
	}
}

func TestHealthCheckerStatus(t *testing.T) {
	cache := NewChannelCache(ChannelCacheLevelMemory)

	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "test")
	cache.AddChannel(ch)

	config := DefaultHealthCheckConfig()
	config.HealthyThreshold = 95.0
	config.DegradedThreshold = 50.0

	checker := NewHealthChecker(cache, config)

	// 记录成功
	for i := 0; i < 95; i++ {
		ch.RecordSuccess(100)
	}
	// 记录失败
	for i := 0; i < 5; i++ {
		ch.RecordFailure()
	}

	result := checker.performCheck(ch)

	// 成功率 95%，应该是 Healthy
	if result.Status != ChannelStatusHealthy {
		t.Errorf("Expected Healthy status, got %v", result.Status)
	}
}

func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker("ch-1", 5, 3, 1*time.Second)

	if !cb.IsAvailable() {
		t.Errorf("Expected circuit breaker to be available")
	}

	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected Closed state")
	}
}

func TestCircuitBreakerOpen(t *testing.T) {
	cb := NewCircuitBreaker("ch-1", 3, 2, 1*time.Second)

	// 记录 3 次失败
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.IsAvailable() {
		t.Errorf("Expected circuit breaker to be unavailable")
	}

	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected Open state")
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("ch-1", 3, 2, 100*time.Millisecond)

	// 打开断路器
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected Open state")
	}

	// 等待超时
	time.Sleep(150 * time.Millisecond)

	// 记录成功转换到半开
	cb.RecordSuccess()

	if cb.GetState() != CircuitHalfOpen {
		t.Errorf("Expected HalfOpen state")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker("ch-1", 2, 2, 100*time.Millisecond)

	// 打开断路器
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected Open state")
	}

	// 等待超时进入半开
	time.Sleep(150 * time.Millisecond)
	cb.RecordSuccess()

	if cb.GetState() != CircuitHalfOpen {
		t.Errorf("Expected HalfOpen state")
	}

	// 再成功一次应该关闭
	cb.RecordSuccess()

	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected Closed state")
	}

	if !cb.IsAvailable() {
		t.Errorf("Expected circuit breaker to be available")
	}
}

func TestHealthCheckConfig(t *testing.T) {
	config := DefaultHealthCheckConfig()

	if config.Interval != 5*time.Minute {
		t.Errorf("Expected 5 minute interval")
	}

	if config.MaxConsecutiveFailures != 5 {
		t.Errorf("Expected 5 max consecutive failures")
	}
}

func TestHealthCheckManager(t *testing.T) {
	manager := NewHealthCheckManager()

	cache := NewChannelCache(ChannelCacheLevelMemory)
	ch := NewChannel("ch-1", "Channel 1", "https://api.test.com", "test")
	cache.AddChannel(ch)

	checker := NewHealthChecker(cache, DefaultHealthCheckConfig())
	manager.RegisterChecker("test", checker)

	retrieved := manager.GetChecker("test")
	if retrieved == nil {
		t.Errorf("Expected checker to be registered")
	}

	stats := manager.GetAllStatistics()
	if len(stats) != 1 {
		t.Errorf("Expected 1 checker in statistics")
	}
}

func BenchmarkCircuitBreakerSuccess(b *testing.B) {
	cb := NewCircuitBreaker("ch-1", 5, 3, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordSuccess()
	}
}

func BenchmarkCircuitBreakerFailure(b *testing.B) {
	cb := NewCircuitBreaker("ch-1", 5, 3, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.RecordFailure()
	}
}

