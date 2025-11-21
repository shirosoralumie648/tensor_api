package ratelimit

import (
	"fmt"
	"time"
)

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	userLimiter   Limiter
	tokenLimiter  Limiter
	ipLimiter     Limiter
	modelLimiter  Limiter
	quotaManager  *QuotaManager
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(
	userLimiter Limiter,
	tokenLimiter Limiter,
	ipLimiter Limiter,
	modelLimiter Limiter,
	quotaManager *QuotaManager,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		userLimiter:  userLimiter,
		tokenLimiter: tokenLimiter,
		ipLimiter:    ipLimiter,
		modelLimiter: modelLimiter,
		quotaManager: quotaManager,
	}
}

// CheckRequest 检查请求
func (rlm *RateLimitMiddleware) CheckRequest(userID, tokenID, clientIP, model string, cost int64) *RateLimitResponse {
	response := &RateLimitResponse{
		Allowed: true,
		Headers: make(map[string]string),
	}

	// 检查用户限流
	if !rlm.userLimiter.Allow(userID) {
		response.Allowed = false
		response.Reason = "user_rate_limit_exceeded"
		response.RetryAfter = 60
		response.Headers["Retry-After"] = "60"
		return response
	}

	// 检查 Token 限流
	if !rlm.tokenLimiter.Allow(tokenID) {
		response.Allowed = false
		response.Reason = "token_rate_limit_exceeded"
		response.RetryAfter = 60
		response.Headers["Retry-After"] = "60"
		return response
	}

	// 检查 IP 限流
	if !rlm.ipLimiter.Allow(clientIP) {
		response.Allowed = false
		response.Reason = "ip_rate_limit_exceeded"
		response.RetryAfter = 60
		response.Headers["Retry-After"] = "60"
		return response
	}

	// 检查模型限流
	if !rlm.modelLimiter.Allow(model) {
		response.Allowed = false
		response.Reason = "model_rate_limit_exceeded"
		response.RetryAfter = 60
		response.Headers["Retry-After"] = "60"
		return response
	}

	// 检查配额
	quotaResp := rlm.quotaManager.CheckQuota(&QuotaRequest{
		UserID: userID,
		Cost:   cost,
		Type:   QuotaDaily,
	})

	if !quotaResp.Allowed {
		response.Allowed = false
		response.Reason = quotaResp.Message
		response.RetryAfter = int64(quotaResp.RetryAfter.Seconds())
		response.Headers["Retry-After"] = fmt.Sprintf("%d", response.RetryAfter)
		response.Headers["X-RateLimit-Remaining"] = fmt.Sprintf("%d", quotaResp.Remaining)
		return response
	}

	// 添加限流信息头
	response.Headers["X-RateLimit-Limit"] = "1000"
	response.Headers["X-RateLimit-Remaining"] = fmt.Sprintf("%d", rlm.userLimiter.GetRemaining(userID))
	response.Headers["X-RateLimit-Reset"] = fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix())

	return response
}

// RateLimitResponse 限流响应
type RateLimitResponse struct {
	Allowed    bool
	Reason     string
	RetryAfter int64
	Headers    map[string]string
	Message    string
}

// GetStatus 获取状态
func (rlm *RateLimitMiddleware) GetStatus(userID, tokenID, clientIP, model string) map[string]interface{} {
	return map[string]interface{}{
		"user":  map[string]interface{}{"remaining": rlm.userLimiter.GetRemaining(userID)},
		"token": map[string]interface{}{"remaining": rlm.tokenLimiter.GetRemaining(tokenID)},
		"ip":    map[string]interface{}{"remaining": rlm.ipLimiter.GetRemaining(clientIP)},
		"model": map[string]interface{}{"remaining": rlm.modelLimiter.GetRemaining(model)},
		"quota": rlm.quotaManager.GetQuotaStatus(userID),
	}
}

// Reset 重置限流
func (rlm *RateLimitMiddleware) Reset(userID, tokenID, clientIP, model string) {
	rlm.userLimiter.Reset(userID)
	rlm.tokenLimiter.Reset(tokenID)
	rlm.ipLimiter.Reset(clientIP)
	rlm.modelLimiter.Reset(model)
}


