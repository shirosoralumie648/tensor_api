package relay

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryPolicyCalculateDelay(t *testing.T) {
	policy := NewRetryPolicy()

	t.Run("exponential_backoff", func(t *testing.T) {
		policy.Strategy = RetryStrategyExponentialBackoff
		policy.InitialDelayMs = 100
		policy.BackoffMultiplier = 2.0
		policy.EnableJitter = false

		delay0 := policy.CalculateDelay(0)
		delay1 := policy.CalculateDelay(1)
		delay2 := policy.CalculateDelay(2)

		assert.Equal(t, 100*time.Millisecond, delay0)
		assert.Equal(t, 200*time.Millisecond, delay1)
		assert.Equal(t, 400*time.Millisecond, delay2)
	})

	t.Run("linear_backoff", func(t *testing.T) {
		policy.Strategy = RetryStrategyLinearBackoff
		policy.InitialDelayMs = 100
		policy.BackoffMultiplier = 2.0
		policy.EnableJitter = false

		delay0 := policy.CalculateDelay(0)
		delay1 := policy.CalculateDelay(1)
		delay2 := policy.CalculateDelay(2)

		assert.Equal(t, 100*time.Millisecond, delay0)
		assert.Equal(t, 300*time.Millisecond, delay1)
		assert.Equal(t, 500*time.Millisecond, delay2)
	})

	t.Run("fixed_delay", func(t *testing.T) {
		policy.Strategy = RetryStrategyFixedDelay
		policy.InitialDelayMs = 100
		policy.EnableJitter = false

		delay0 := policy.CalculateDelay(0)
		delay1 := policy.CalculateDelay(1)
		delay2 := policy.CalculateDelay(2)

		assert.Equal(t, 100*time.Millisecond, delay0)
		assert.Equal(t, 100*time.Millisecond, delay1)
		assert.Equal(t, 100*time.Millisecond, delay2)
	})

	t.Run("max_delay_cap", func(t *testing.T) {
		policy.Strategy = RetryStrategyExponentialBackoff
		policy.InitialDelayMs = 1000
		policy.MaxDelayMs = 5000
		policy.BackoffMultiplier = 2.0
		policy.EnableJitter = false

		delay0 := policy.CalculateDelay(0)
		delay1 := policy.CalculateDelay(1)
		delay2 := policy.CalculateDelay(2)
		delay3 := policy.CalculateDelay(3)

		assert.Equal(t, 1000*time.Millisecond, delay0)
		assert.Equal(t, 2000*time.Millisecond, delay1)
		assert.Equal(t, 4000*time.Millisecond, delay2)
		// 超过最大延迟，应该被限制
		assert.Equal(t, 5000*time.Millisecond, delay3)
	})
}

func TestRetryPolicyIsRetryable(t *testing.T) {
	policy := NewRetryPolicy()

	t.Run("retryable_status_codes", func(t *testing.T) {
		assert.True(t, policy.IsRetryable(500, nil))
		assert.True(t, policy.IsRetryable(503, nil))
		assert.True(t, policy.IsRetryable(429, nil))
	})

	t.Run("non_retryable_status_codes", func(t *testing.T) {
		assert.False(t, policy.IsRetryable(200, nil))
		assert.False(t, policy.IsRetryable(404, nil))
		assert.False(t, policy.IsRetryable(401, nil))
	})

	t.Run("with_error", func(t *testing.T) {
		err := fmt.Errorf("network error")
		assert.True(t, policy.IsRetryable(0, err))
	})
}

func TestRetryPolicyCanRetry(t *testing.T) {
	policy := NewRetryPolicy()
	policy.MaxRetries = 3

	assert.True(t, policy.CanRetry(0))
	assert.True(t, policy.CanRetry(1))
	assert.True(t, policy.CanRetry(2))
	assert.False(t, policy.CanRetry(3))
}

func TestRetry(t *testing.T) {
	policy := NewRetryPolicy()
	policy.MaxRetries = 3
	policy.InitialDelayMs = 10
	policy.EnableJitter = false

	t.Run("immediate_success", func(t *testing.T) {
		callCount := 0
		success, err := Retry(context.Background(), policy, func(ctx context.Context) error {
			callCount++
			return nil
		})

		assert.True(t, success)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("retry_and_succeed", func(t *testing.T) {
		callCount := 0
		success, err := Retry(context.Background(), policy, func(ctx context.Context) error {
			callCount++
			if callCount < 3 {
				return &RetryableError{StatusCode: 503}
			}
			return nil
		})

		assert.True(t, success)
		assert.NoError(t, err)
		assert.Equal(t, 3, callCount)
	})

	t.Run("max_retries_exceeded", func(t *testing.T) {
		callCount := 0
		success, err := Retry(context.Background(), policy, func(ctx context.Context) error {
			callCount++
			return &RetryableError{StatusCode: 503}
		})

		assert.False(t, success)
		assert.Error(t, err)
		assert.Equal(t, 4, callCount) // 初始尝试 + 3 次重试
	})

	t.Run("non_retryable_error", func(t *testing.T) {
		callCount := 0
		success, err := Retry(context.Background(), policy, func(ctx context.Context) error {
			callCount++
			return fmt.Errorf("non-retryable error")
		})

		assert.False(t, success)
		assert.Error(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("context_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		success, err := Retry(ctx, policy, func(ctx context.Context) error {
			return &RetryableError{StatusCode: 503}
		})

		assert.False(t, success)
		assert.Error(t, err)
	})
}

func TestRetryWithCallback(t *testing.T) {
	policy := NewRetryPolicy()
	policy.MaxRetries = 3
	policy.InitialDelayMs = 10
	policy.EnableJitter = false

	t.Run("successful_with_retries", func(t *testing.T) {
		callCount := 0
		success, err := RetryWithCallback(
			context.Background(),
			policy,
			func(ctx context.Context, retryCtx *RetryContext) error {
				callCount++
				assert.Equal(t, callCount-1, retryCtx.RetryCount)

				if callCount < 2 {
					return &RetryableError{StatusCode: 503}
				}
				return nil
			},
		)

		assert.True(t, success)
		assert.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("retry_context_details", func(t *testing.T) {
		callCount := 0
		success, err := RetryWithCallback(
			context.Background(),
			policy,
			func(ctx context.Context, retryCtx *RetryContext) error {
				callCount++

				if callCount == 1 {
					assert.False(t, retryCtx.IsLastAttempt)
					return &RetryableError{StatusCode: 503}
				}

				if callCount == 4 {
					assert.True(t, retryCtx.IsLastAttempt)
				}

				if callCount < 5 {
					return &RetryableError{StatusCode: 503}
				}

				return nil
			},
		)

		assert.True(t, success)
		assert.NoError(t, err)
	})
}

func TestRetryableError(t *testing.T) {
	err := &RetryableError{
		StatusCode: 503,
		Message:    "Service Unavailable",
	}

	assert.True(t, IsRetryableError(err))
	assert.False(t, IsRetryableError(fmt.Errorf("regular error")))
	assert.NotEmpty(t, err.Error())
}

func TestRetryPolicyStatistics(t *testing.T) {
	policy := NewRetryPolicy()

	policy.RecordSuccess()
	policy.RecordSuccess()
	policy.RecordFailure()

	stats := policy.GetStatistics()
	assert.Equal(t, int64(3), stats["total_retries"])
	assert.Equal(t, int64(2), stats["success_retries"])
	assert.Equal(t, int64(1), stats["failed_retries"])
	assert.InDelta(t, 66.67, stats["success_rate"], 1)
}

func BenchmarkRetryPolicyCalculateDelay(b *testing.B) {
	policy := NewRetryPolicy()
	policy.EnableJitter = false

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		policy.CalculateDelay(i % 10)
	}
}

func BenchmarkRetry(b *testing.B) {
	policy := NewRetryPolicy()
	policy.MaxRetries = 3
	policy.InitialDelayMs = 1

	callCount := 0
	fn := func(ctx context.Context) error {
		callCount++
		if callCount%3 == 0 {
			callCount = 0
			return nil
		}
		return &RetryableError{StatusCode: 503}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Retry(context.Background(), policy, fn)
	}
}

