package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/quota"
	"github.com/shirosoralumie648/Oblivious/backend/internal/tokenizer"
)

// StreamHandler 流式请求处理器
type StreamHandler struct {
	quotaService     quota.QuotaService
	tokenizerFactory *tokenizer.TokenizerFactory
}

// NewStreamHandler 创建流式处理器
func NewStreamHandler(quotaService quota.QuotaService, tokenizerFactory *tokenizer.TokenizerFactory) *StreamHandler {
	return &StreamHandler{
		quotaService:     quotaService,
		tokenizerFactory: tokenizerFactory,
	}
}

// StreamOptions 流式处理选项
type StreamOptions struct {
	RequestID    string
	UserID       int
	ChannelID    int
	Model        string
	PromptTokens int
	MaxTokens    int
	TotalTimeout time.Duration // 总超时时间
	IdleTimeout  time.Duration // 空闲超时时间
}

// StreamResult 流式处理结果
type StreamResult struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Duration         time.Duration
	ChunkCount       int
}

// HandleStreamResponse 处理流式响应
func (h *StreamHandler) HandleStreamResponse(
	ctx context.Context,
	sender *StreamSender,
	chunkChan <-chan *StreamChunk,
	errChan <-chan error,
	opts *StreamOptions,
) (*StreamResult, error) {
	startTime := time.Now()

	// 设置总超时
	totalTimeout := opts.TotalTimeout
	if totalTimeout == 0 {
		totalTimeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	// 创建Token计数器
	counter, err := h.tokenizerFactory.CreateStreamCounter(opts.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create token counter: %w", err)
	}

	// 设置空闲超时
	idleTimeout := opts.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = 30 * time.Second
	}
	timer := time.NewTimer(idleTimeout)
	defer timer.Stop()

	chunkCount := 0

	// 确保结束时执行后扣费
	defer func() {
		finalTokens := counter.Finalize()

		// 计算实际配额（需要从quotaService获取计算器）
		// 这里简化处理，实际应该调用calculator
		actualQuota := float64(opts.PromptTokens + finalTokens)

		// 执行后扣费
		postReq := &quota.PostConsumeRequest{
			RequestID:        opts.RequestID,
			UserID:           opts.UserID,
			ChannelID:        opts.ChannelID,
			Model:            opts.Model,
			PromptTokens:     opts.PromptTokens,
			CompletionTokens: finalTokens,
			ActualQuota:      actualQuota,
			IsStream:         true,
			ResponseTime:     time.Since(startTime).Milliseconds(),
		}

		if err := h.quotaService.PostConsumeQuota(postReq); err != nil {
			fmt.Printf("failed to post consume quota: %v\n", err)
		}
	}()

	// 处理流式数据
	for {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				// 通道关闭，流结束
				return &StreamResult{
					PromptTokens:     opts.PromptTokens,
					CompletionTokens: counter.GetCurrentCount(),
					TotalTokens:      opts.PromptTokens + counter.GetCurrentCount(),
					Duration:         time.Since(startTime),
					ChunkCount:       chunkCount,
				}, nil
			}

			// 提取内容并计数Token
			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					counter.AddChunk(content)
				}
			}

			// 发送给客户端
			if err := sender.Send(chunk); err != nil {
				return nil, fmt.Errorf("failed to send chunk: %w", err)
			}

			chunkCount++
			timer.Reset(idleTimeout) // 重置空闲超时

		case err := <-errChan:
			// 上游错误
			if sendErr := sender.SendError(err.Error()); sendErr != nil {
				return nil, fmt.Errorf("upstream error: %v, send error: %v", err, sendErr)
			}
			return nil, fmt.Errorf("upstream error: %w", err)

		case <-timer.C:
			// 空闲超时
			return nil, fmt.Errorf("stream idle timeout after %v", idleTimeout)

		case <-ctx.Done():
			// 上下文取消或总超时
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("stream total timeout after %v", totalTimeout)
			}
			return nil, ctx.Err()
		}
	}
}

// HandleStreamWithRetry 带重试的流式处理
func (h *StreamHandler) HandleStreamWithRetry(
	ctx context.Context,
	sender *StreamSender,
	streamFunc func() (<-chan *StreamChunk, <-chan error, error),
	opts *StreamOptions,
	maxRetries int,
) (*StreamResult, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 重试前等待
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		chunkChan, errChan, err := streamFunc()
		if err != nil {
			lastErr = err
			continue
		}

		result, err := h.HandleStreamResponse(ctx, sender, chunkChan, errChan, opts)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 检查是否是可重试的错误
		if !isRetryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("stream failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError 判断是否可重试
func isRetryableError(err error) bool {
	// 简化判断，实际应该根据具体错误类型
	errStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection reset",
		"temporary",
		"503",
		"502",
		"504",
	}

	for _, retryErr := range retryableErrors {
		if contains(errStr, retryErr) {
			return true
		}
	}

	return false
}

// contains 字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
