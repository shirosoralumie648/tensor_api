package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 1) // 10 tokens, 1 per second

	// 应该允许 10 个请求
	for i := 0; i < 10; i++ {
		if !limiter.Allow("user1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11 个请求应该被拒绝
	if limiter.Allow("user1") {
		t.Error("11th request should be denied")
	}

	limiter.Close()
}

func TestSlidingWindowLimiter(t *testing.T) {
	limiter := NewSlidingWindowLimiter(time.Minute, 10)

	// 应该允许 10 个请求
	for i := 0; i < 10; i++ {
		if !limiter.Allow("user1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11 个请求应该被拒绝
	if limiter.Allow("user1") {
		t.Error("11th request should be denied")
	}

	// 检查剩余请求数
	remaining := limiter.GetRemaining("user1")
	if remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", remaining)
	}
}

func TestFixedWindowLimiter(t *testing.T) {
	limiter := NewFixedWindowLimiter(time.Minute, 10)

	// 应该允许 10 个请求
	for i := 0; i < 10; i++ {
		if !limiter.Allow("user1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11 个请求应该被拒绝
	if limiter.Allow("user1") {
		t.Error("11th request should be denied")
	}
}

func TestQuotaManager(t *testing.T) {
	qm := NewQuotaManager(1000, 10000, 100)

	// 首次请求应该通过
	resp := qm.CheckQuota(&QuotaRequest{
		UserID: "user1",
		Cost:   100,
		Type:   QuotaDaily,
	})

	if !resp.Allowed {
		t.Error("First request should be allowed")
	}

	// 检查剩余配额
	if resp.Remaining != 900 {
		t.Errorf("Expected 900 remaining, got %d", resp.Remaining)
	}

	// 超过配额的请求应该被拒绝
	resp = qm.CheckQuota(&QuotaRequest{
		UserID: "user1",
		Cost:   950, // 总计 1050 > 1000
		Type:   QuotaDaily,
	})

	if resp.Allowed {
		t.Error("Request exceeding quota should be denied")
	}
}

func TestMultiLevelLimiter(t *testing.T) {
	userLimiter := NewFixedWindowLimiter(time.Minute, 100)
	tokenLimiter := NewFixedWindowLimiter(time.Minute, 50)

	mll := NewMultiLevelLimiter()
	mll.AddLimiter("user", userLimiter)
	mll.AddLimiter("token", tokenLimiter)

	// 应该允许请求
	if !mll.Allow("key1") {
		t.Error("First request should be allowed")
	}

	// 超过 token 限制
	for i := 0; i < 49; i++ {
		mll.Allow("key1")
	}

	// 第 51 个请求应该被拒绝
	if mll.Allow("key1") {
		t.Error("51st request should be denied (token limit)")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	userLimiter := NewFixedWindowLimiter(time.Minute, 100)
	tokenLimiter := NewFixedWindowLimiter(time.Minute, 100)
	ipLimiter := NewFixedWindowLimiter(time.Minute, 100)
	modelLimiter := NewFixedWindowLimiter(time.Minute, 100)
	quotaManager := NewQuotaManager(10000, 100000, 1000)

	middleware := NewRateLimitMiddleware(
		userLimiter,
		tokenLimiter,
		ipLimiter,
		modelLimiter,
		quotaManager,
	)

	// 检查请求
	resp := middleware.CheckRequest("user1", "token1", "127.0.0.1", "gpt-4", 100)

	if !resp.Allowed {
		t.Error("Request should be allowed")
	}

	if len(resp.Headers) == 0 {
		t.Error("Response should have headers")
	}
}

func TestQuotaStatus(t *testing.T) {
	qm := NewQuotaManager(1000, 10000, 100)

	status := qm.GetQuotaStatus("user1")

	if status == nil {
		t.Error("Status should not be nil")
	}

	daily, ok := status["daily"].(map[string]interface{})
	if !ok {
		t.Error("Daily status should be present")
	}

	if daily["quota"] != int64(1000) {
		t.Errorf("Expected quota 1000, got %v", daily["quota"])
	}
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	limiter := NewTokenBucketLimiter(10000, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("user1")
	}

	limiter.Close()
}

func BenchmarkSlidingWindowAllow(b *testing.B) {
	limiter := NewSlidingWindowLimiter(time.Minute, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow("user1")
	}
}

func BenchmarkQuotaCheck(b *testing.B) {
	qm := NewQuotaManager(1000000, 10000000, 100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qm.CheckQuota(&QuotaRequest{
			UserID: "user1",
			Cost:   100,
			Type:   QuotaDaily,
		})
	}
}


