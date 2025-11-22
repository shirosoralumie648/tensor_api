package relay

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shirosoralumie648/Oblivious/backend/internal/quota"
	"github.com/shirosoralumie648/Oblivious/backend/internal/tokenizer"
)

// ExampleStreamUsage 流式请求使用示例
func ExampleStreamUsage(w http.ResponseWriter, r *http.Request, quotaService quota.QuotaService) {
	ctx := r.Context()

	// 1. 创建流式发送器
	sender, err := NewStreamSender(w, ctx)
	if err != nil {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	defer sender.Close()

	// 2. 创建Tokenizer工厂和流式处理器
	tokenizerFactory, _ := tokenizer.GetGlobalFactory()
	handler := NewStreamHandler(quotaService, tokenizerFactory)

	// 3. 预扣费
	preReq := &quota.PreConsumeRequest{
		RequestID:      "stream-req-001",
		UserID:         1,
		Model:          "gpt-4o",
		PromptTokens:   200,
		MaxTokens:      2000,
		EstimatedQuota: 1000,
		TrustThreshold: 5000,
	}

	_, err = quotaService.PreConsumeQuota(preReq)
	if err != nil {
		sender.SendError(fmt.Sprintf("预扣费失败: %v", err))
		return
	}

	// 确保失败时退款
	defer func() {
		if err != nil {
			quotaService.ReturnPreConsumedQuota("stream-req-001", 1)
		}
	}()

	// 4. 模拟调用上游API（流式）
	chunkChan := make(chan *StreamChunk, 10)
	errChan := make(chan error, 1)

	// 模拟生成流式数据
	go func() {
		defer close(chunkChan)
		defer close(errChan)

		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case <-time.After(100 * time.Millisecond):
				chunk := &StreamChunk{
					ID:      "chatcmpl-123",
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   "gpt-4o",
					Choices: []ChunkChoice{
						{
							Index: 0,
							Delta: ChunkDelta{
								Content: fmt.Sprintf("This is chunk %d. ", i),
							},
						},
					},
				}
				chunkChan <- chunk
			}
		}
	}()

	// 5. 处理流式响应
	opts := &StreamOptions{
		RequestID:    "stream-req-001",
		UserID:       1,
		ChannelID:    5,
		Model:        "gpt-4o",
		PromptTokens: 200,
		MaxTokens:    2000,
		TotalTimeout: 5 * time.Minute,
		IdleTimeout:  30 * time.Second,
	}

	result, err := handler.HandleStreamResponse(ctx, sender, chunkChan, errChan, opts)
	if err != nil {
		fmt.Printf("Stream处理失败: %v\n", err)
		return
	}

	fmt.Printf("Stream完成: 生成了%d个tokens，共%d个chunks\n",
		result.CompletionTokens, result.ChunkCount)
}

// ExampleStreamMonitor 流式监控示例
func ExampleStreamMonitor() {
	ctx := context.Background()
	monitor := NewStreamMonitor()

	// 启动定期清理
	go monitor.StartCleanupWorker(ctx, 1*time.Minute, 10*time.Minute)

	// 开始监控流
	monitor.StartStream("req-001", 1, "gpt-4o")

	// 更新活动
	monitor.UpdateActivity("req-001", 100, 5)

	// 获取指标
	metrics := monitor.GetMetrics("req-001")
	if metrics != nil {
		fmt.Printf("Stream metrics: chunks=%d, tokens=%d, bytes=%d\n",
			metrics.ChunkCount, metrics.TokenCount, metrics.BytesSent)
	}

	// 完成流
	monitor.CompleteStream("req-001")

	// 获取统计
	stats := monitor.GetStats()
	fmt.Printf("Monitor stats: %+v\n", stats)

	// 清理过期记录
	cleaned := monitor.CleanupStaleStreams(5 * time.Minute)
	fmt.Printf("Cleaned %d stale streams\n", cleaned)
}

// ExampleStreamWithRetry 带重试的流式示例
func ExampleStreamWithRetry(w http.ResponseWriter, r *http.Request, quotaService quota.QuotaService) {
	ctx := r.Context()

	sender, _ := NewStreamSender(w, ctx)
	defer sender.Close()

	tokenizerFactory, _ := tokenizer.GetGlobalFactory()
	handler := NewStreamHandler(quotaService, tokenizerFactory)

	// 定义流式函数（可能失败并重试）
	streamFunc := func() (<-chan *StreamChunk, <-chan error, error) {
		chunkChan := make(chan *StreamChunk, 10)
		errChan := make(chan error, 1)

		// 模拟可能失败的API调用
		// 这里返回真实的上游API流式通道
		return chunkChan, errChan, nil
	}

	opts := &StreamOptions{
		RequestID:    "retry-req-001",
		UserID:       1,
		ChannelID:    5,
		Model:        "gpt-4o",
		PromptTokens: 100,
		MaxTokens:    1000,
	}

	// 最多重试3次
	result, err := handler.HandleStreamWithRetry(ctx, sender, streamFunc, opts, 3)
	if err != nil {
		fmt.Printf("Stream失败（已重试）: %v\n", err)
		return
	}

	fmt.Printf("Stream成功: %d tokens\n", result.CompletionTokens)
}
