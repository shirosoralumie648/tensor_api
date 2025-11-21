package relay

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

// RetryStrategy 重试策略
type RetryStrategy int

const (
	// 指数退避策略
	RetryStrategyExponentialBackoff RetryStrategy = iota
	// 线性退避策略
	RetryStrategyLinearBackoff
	// 固定延迟策略
	RetryStrategyFixedDelay
)

// RetryPolicy 重试策略配置
type RetryPolicy struct {
	// 最大重试次数
	MaxRetries int

	// 重试策略
	Strategy RetryStrategy

	// 初始延迟 (毫秒)
	InitialDelayMs int64

	// 最大延迟 (毫秒)
	MaxDelayMs int64

	// 退避乘数 (用于指数退避和线性退避)
	BackoffMultiplier float64

	// 是否添加抖动 (防止雷鸣羊群效应)
	EnableJitter bool

	// 可重试的 HTTP 状态码
	RetryableStatusCodes map[int]bool

	// 统计信息
	totalRetries   int64
	successRetries int64
	failedRetries  int64
}

// NewRetryPolicy 创建新的重试策略
func NewRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:        3,
		Strategy:          RetryStrategyExponentialBackoff,
		InitialDelayMs:    100,
		MaxDelayMs:        10000,
		BackoffMultiplier: 2.0,
		EnableJitter:      true,
		RetryableStatusCodes: map[int]bool{
			408: true, // Request Timeout
			429: true, // Too Many Requests
			500: true, // Internal Server Error
			502: true, // Bad Gateway
			503: true, // Service Unavailable
			504: true, // Gateway Timeout
		},
	}
}

// CalculateDelay 计算延迟时间
// retryCount: 当前重试次数 (从 0 开始)
func (rp *RetryPolicy) CalculateDelay(retryCount int) time.Duration {
	var delayMs int64

	switch rp.Strategy {
	case RetryStrategyExponentialBackoff:
		// 指数退避: initialDelay * (multiplier ^ retryCount)
		delayMs = int64(float64(rp.InitialDelayMs) * math.Pow(rp.BackoffMultiplier, float64(retryCount)))

	case RetryStrategyLinearBackoff:
		// 线性退避: initialDelay * (1 + multiplier * retryCount)
		delayMs = int64(float64(rp.InitialDelayMs) * (1 + rp.BackoffMultiplier*float64(retryCount)))

	case RetryStrategyFixedDelay:
		// 固定延迟
		delayMs = rp.InitialDelayMs

	default:
		delayMs = rp.InitialDelayMs
	}

	// 限制最大延迟
	if delayMs > rp.MaxDelayMs {
		delayMs = rp.MaxDelayMs
	}

	// 添加抖动
	if rp.EnableJitter {
		jitter := rand.Int63n(delayMs / 2)
		delayMs += jitter
	}

	return time.Duration(delayMs) * time.Millisecond
}

// IsRetryable 检查是否应该重试
func (rp *RetryPolicy) IsRetryable(statusCode int, err error) bool {
	// 检查 HTTP 状态码
	if _, ok := rp.RetryableStatusCodes[statusCode]; ok {
		return true
	}

	// 网络错误也可以重试
	if err != nil {
		// 连接错误、超时等都应该重试
		return true
	}

	return false
}

// CanRetry 检查是否还可以重试
func (rp *RetryPolicy) CanRetry(retryCount int) bool {
	return retryCount < rp.MaxRetries
}

// RecordSuccess 记录成功的重试
func (rp *RetryPolicy) RecordSuccess() {
	atomic.AddInt64(&rp.totalRetries, 1)
	atomic.AddInt64(&rp.successRetries, 1)
}

// RecordFailure 记录失败的重试
func (rp *RetryPolicy) RecordFailure() {
	atomic.AddInt64(&rp.totalRetries, 1)
	atomic.AddInt64(&rp.failedRetries, 1)
}

// GetStatistics 获取重试统计
func (rp *RetryPolicy) GetStatistics() map[string]interface{} {
	total := atomic.LoadInt64(&rp.totalRetries)
	success := atomic.LoadInt64(&rp.successRetries)
	failed := atomic.LoadInt64(&rp.failedRetries)

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_retries":   total,
		"success_retries": success,
		"failed_retries":  failed,
		"success_rate":    successRate,
	}
}

// RetryableError 可重试的错误
type RetryableError struct {
	StatusCode int
	Message    string
	Err        error
	RetryAfter time.Duration
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: status=%d, message=%s", e.StatusCode, e.Message)
}

// IsRetryableError 检查是否是可重试错误
func IsRetryableError(err error) bool {
	_, ok := err.(*RetryableError)
	return ok
}

// Retry 执行重试逻辑
// fn: 需要执行的函数
// 返回: (成功标志, 错误)
func Retry(ctx context.Context, policy *RetryPolicy, fn func(context.Context) error) (bool, error) {
	var lastErr error

	for retryCount := 0; retryCount <= policy.MaxRetries; retryCount++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		// 执行函数
		err := fn(ctx)
		if err == nil {
			// 成功
			if retryCount > 0 {
				policy.RecordSuccess()
			}
			return true, nil
		}

		lastErr = err

		// 检查是否可以重试
		retryableErr, ok := err.(*RetryableError)
		if !ok || !policy.CanRetry(retryCount) {
			policy.RecordFailure()
			return false, err
		}

		if retryCount < policy.MaxRetries {
			// 计算延迟
			delay := policy.CalculateDelay(retryCount)

			// 如果错误中指定了 Retry-After，使用更长的延迟
			if retryableErr.RetryAfter > 0 && retryableErr.RetryAfter > delay {
				delay = retryableErr.RetryAfter
			}

			// 等待后重试
			select {
			case <-time.After(delay):
				// 继续重试
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}

	policy.RecordFailure()
	return false, lastErr
}

// RetryWithContext 带上下文的重试
type RetryContext struct {
	// 重试次数
	RetryCount int

	// 当前尝试的延迟
	Delay time.Duration

	// 最后一个错误
	LastError error

	// 是否是最后一次重试
	IsLastAttempt bool
}

// RetryWithCallback 带回调的重试
// fn: 需要执行的函数，接收 RetryContext 作为参数
func RetryWithCallback(
	ctx context.Context,
	policy *RetryPolicy,
	fn func(context.Context, *RetryContext) error,
) (bool, error) {
	var lastErr error
	var lastDelay time.Duration

	for retryCount := 0; retryCount <= policy.MaxRetries; retryCount++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		// 准备 RetryContext
		retryCtx := &RetryContext{
			RetryCount:    retryCount,
			Delay:         lastDelay,
			LastError:     lastErr,
			IsLastAttempt: retryCount == policy.MaxRetries,
		}

		// 执行函数
		err := fn(ctx, retryCtx)
		if err == nil {
			// 成功
			if retryCount > 0 {
				policy.RecordSuccess()
			}
			return true, nil
		}

		lastErr = err

		// 检查是否可以重试
		retryableErr, ok := err.(*RetryableError)
		if !ok || !policy.CanRetry(retryCount) {
			policy.RecordFailure()
			return false, err
		}

		if retryCount < policy.MaxRetries {
			// 计算延迟
			delay := policy.CalculateDelay(retryCount)
			lastDelay = delay

			// 如果错误中指定了 Retry-After，使用更长的延迟
			if retryableErr.RetryAfter > 0 && retryableErr.RetryAfter > delay {
				delay = retryableErr.RetryAfter
				lastDelay = delay
			}

			// 等待后重试
			select {
			case <-time.After(delay):
				// 继续重试
			case <-ctx.Done():
				return false, ctx.Err()
			}
		}
	}

	policy.RecordFailure()
	return false, lastErr
}

