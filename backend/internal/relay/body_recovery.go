package relay

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// RecoveryStrategy 恢复策略
type RecoveryStrategy int

const (
	// 重试策略
	RecoveryStrategyRetry RecoveryStrategy = iota
	// 备用渠道策略
	RecoveryStrategyAlternateChannel
	// 分块重发策略
	RecoveryStrategyChunked
)

// BodyRecoveryManager 请求体恢复管理器
type BodyRecoveryManager struct {
	// 缓存管理器
	cache *BodyCache

	// 恢复策略
	strategy RecoveryStrategy

	// 分块大小
	chunkSize int64

	// 最大恢复次数
	maxRecoveries int

	// 请求体恢复历史
	recoveryHistory map[string]*RecoveryHistory
	historyMu       sync.RWMutex

	// 统计信息
	totalRecoveries      int64
	successfulRecoveries int64
	failedRecoveries     int64
	totalBytesRecovered  int64

	// 日志函数
	logFunc func(level, msg string, args ...interface{})
}

// RecoveryHistory 恢复历史
type RecoveryHistory struct {
	RequestID        string        // 请求 ID
	CacheID          string        // 缓存 ID
	OriginalSize     int64         // 原始大小
	Attempts         int           // 尝试次数
	SuccessfulChunks int           // 成功的分块数
	FailedChunks     int           // 失败的分块数
	LastAttemptTime  time.Time     // 最后尝试时间
	Status           string        // 状态（pending, success, failed）
}

// NewBodyRecoveryManager 创建新的请求体恢复管理器
func NewBodyRecoveryManager(cache *BodyCache) *BodyRecoveryManager {
	return &BodyRecoveryManager{
		cache:               cache,
		strategy:            RecoveryStrategyRetry,
		chunkSize:           512 * 1024, // 512KB
		maxRecoveries:       3,
		recoveryHistory:     make(map[string]*RecoveryHistory),
		logFunc:             defaultLogFunc,
	}
}

// SetStrategy 设置恢复策略
func (brm *BodyRecoveryManager) SetStrategy(strategy RecoveryStrategy) {
	brm.strategy = strategy
}

// SetChunkSize 设置分块大小
func (brm *BodyRecoveryManager) SetChunkSize(size int64) {
	brm.chunkSize = size
}

// SetMaxRecoveries 设置最大恢复次数
func (brm *BodyRecoveryManager) SetMaxRecoveries(max int) {
	brm.maxRecoveries = max
}

// SetLogFunc 设置日志函数
func (brm *BodyRecoveryManager) SetLogFunc(logFunc func(level, msg string, args ...interface{})) {
	brm.logFunc = logFunc
}

// InitiateRecovery 启动恢复过程
func (brm *BodyRecoveryManager) InitiateRecovery(ctx context.Context, requestID, cacheID string) error {
	atomic.AddInt64(&brm.totalRecoveries, 1)

	// 记录恢复历史
	brm.historyMu.Lock()
	history := &RecoveryHistory{
		RequestID:       requestID,
		CacheID:         cacheID,
		Attempts:        0,
		Status:          "pending",
		LastAttemptTime: time.Now(),
	}
	brm.recoveryHistory[requestID] = history
	brm.historyMu.Unlock()

	// 获取缓存的请求体
	data, err := brm.cache.GetCachedBody(cacheID)
	if err != nil {
		brm.logFunc("error", "failed to get cached body for recovery",
			"requestID", requestID, "cacheID", cacheID, "error", err)
		brm.recordRecoveryFailure(requestID, err)
		return err
	}

	history.OriginalSize = int64(len(data))

	// 根据策略执行恢复
	switch brm.strategy {
	case RecoveryStrategyRetry:
		return brm.recoverByRetry(ctx, requestID, history, data)

	case RecoveryStrategyAlternateChannel:
		return brm.recoverByAlternateChannel(ctx, requestID, history, data)

	case RecoveryStrategyChunked:
		return brm.recoverByChunked(ctx, requestID, history, data)

	default:
		return fmt.Errorf("unknown recovery strategy: %d", brm.strategy)
	}
}

// recoverByRetry 通过重试恢复
func (brm *BodyRecoveryManager) recoverByRetry(ctx context.Context, requestID string,
	history *RecoveryHistory, data []byte) error {

	for attempt := 0; attempt < brm.maxRecoveries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		history.Attempts = attempt + 1
		history.LastAttemptTime = time.Now()

		// 模拟重试（实际应该调用重试逻辑）
		brm.logFunc("info", "recovery attempt",
			"requestID", requestID, "attempt", attempt+1, "size", len(data))

		// 这里应该调用实际的重试逻辑
		// 成功则标记为 success
		success := true
		if success {
			brm.recordRecoverySuccess(requestID, len(data))
			history.Status = "success"
			return nil
		}

		// 指数退避
		delay := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	brm.recordRecoveryFailure(requestID, fmt.Errorf("max recovery attempts exceeded"))
	return fmt.Errorf("recovery failed after %d attempts", brm.maxRecoveries)
}

// recoverByAlternateChannel 通过备用渠道恢复
func (brm *BodyRecoveryManager) recoverByAlternateChannel(ctx context.Context, requestID string,
	history *RecoveryHistory, data []byte) error {

	history.Attempts = 1
	history.LastAttemptTime = time.Now()

	brm.logFunc("info", "recovery with alternate channel",
		"requestID", requestID, "size", len(data))

	// 这里应该调用使用备用渠道的重试逻辑
	brm.recordRecoverySuccess(requestID, len(data))
	history.Status = "success"
	return nil
}

// recoverByChunked 通过分块恢复
func (brm *BodyRecoveryManager) recoverByChunked(ctx context.Context, requestID string,
	history *RecoveryHistory, data []byte) error {

	history.Attempts = 1
	history.LastAttemptTime = time.Now()

	// 计算分块数量
	chunkCount := (int64(len(data)) + brm.chunkSize - 1) / brm.chunkSize

	brm.logFunc("info", "recovery with chunked strategy",
		"requestID", requestID, "size", len(data), "chunks", chunkCount)

	// 发送每个分块
	for i := int64(0); i < chunkCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := i * brm.chunkSize
		end := start + brm.chunkSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}

		chunk := data[start:end]

		// 这里应该发送分块
		brm.logFunc("info", "sending chunk",
			"requestID", requestID, "chunk", i+1, "of", chunkCount, "size", len(chunk))

		// 成功则计数
		history.SuccessfulChunks++
	}

	brm.recordRecoverySuccess(requestID, len(data))
	history.Status = "success"
	return nil
}

// GetRecoveryHistory 获取恢复历史
func (brm *BodyRecoveryManager) GetRecoveryHistory(requestID string) *RecoveryHistory {
	brm.historyMu.RLock()
	defer brm.historyMu.RUnlock()

	if history, ok := brm.recoveryHistory[requestID]; ok {
		return history
	}
	return nil
}

// GetStatistics 获取统计信息
func (brm *BodyRecoveryManager) GetStatistics() map[string]interface{} {
	total := atomic.LoadInt64(&brm.totalRecoveries)
	successful := atomic.LoadInt64(&brm.successfulRecoveries)
	failed := atomic.LoadInt64(&brm.failedRecoveries)
	recovered := atomic.LoadInt64(&brm.totalBytesRecovered)

	successRate := 0.0
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_recoveries":       total,
		"successful_recoveries":  successful,
		"failed_recoveries":      failed,
		"success_rate":           successRate,
		"total_bytes_recovered":  recovered,
		"active_recoveries":      len(brm.recoveryHistory),
	}
}

// recordRecoverySuccess 记录恢复成功
func (brm *BodyRecoveryManager) recordRecoverySuccess(requestID string, size int) {
	atomic.AddInt64(&brm.successfulRecoveries, 1)
	atomic.AddInt64(&brm.totalBytesRecovered, int64(size))

	brm.historyMu.Lock()
	if history, ok := brm.recoveryHistory[requestID]; ok {
		history.Status = "success"
	}
	brm.historyMu.Unlock()

	brm.logFunc("info", "recovery succeeded", "requestID", requestID, "size", size)
}

// recordRecoveryFailure 记录恢复失败
func (brm *BodyRecoveryManager) recordRecoveryFailure(requestID string, err error) {
	atomic.AddInt64(&brm.failedRecoveries, 1)

	brm.historyMu.Lock()
	if history, ok := brm.recoveryHistory[requestID]; ok {
		history.Status = "failed"
	}
	brm.historyMu.Unlock()

	brm.logFunc("error", "recovery failed", "requestID", requestID, "error", err)
}

// BodyRecoveryStreamReader 可恢复的流读取器
type BodyRecoveryStreamReader struct {
	// 原始读取器
	reader io.Reader

	// 缓存管理器
	cache *BodyCache

	// 恢复管理器
	recovery *BodyRecoveryManager

	// 请求 ID
	requestID string

	// 缓冲区
	buffer *bytes.Buffer

	// 是否启用缓存
	enableCache bool

	// 当前位置
	position int64

	// 总大小
	totalSize int64

	// 缓存 ID
	cacheID string

	mu sync.Mutex
}

// NewBodyRecoveryStreamReader 创建新的可恢复流读取器
func NewBodyRecoveryStreamReader(reader io.Reader, cache *BodyCache,
	recovery *BodyRecoveryManager, requestID string) *BodyRecoveryStreamReader {

	return &BodyRecoveryStreamReader{
		reader:      reader,
		cache:       cache,
		recovery:    recovery,
		requestID:   requestID,
		buffer:      new(bytes.Buffer),
		enableCache: true,
		totalSize:   0,
	}
}

// Read 读取数据
func (brsr *BodyRecoveryStreamReader) Read(p []byte) (n int, err error) {
	brsr.mu.Lock()
	defer brsr.mu.Unlock()

	// 从原始读取器读取
	n, err = brsr.reader.Read(p)

	// 如果启用缓存，写入缓冲区
	if brsr.enableCache && n > 0 {
		brsr.buffer.Write(p[:n])
	}

	brsr.position += int64(n)

	// 如果读取完成，缓存整个请求体
	if err == io.EOF && brsr.enableCache {
		data := brsr.buffer.Bytes()
		cacheID, cacheErr := brsr.cache.CacheRequestBody(bytes.NewReader(data))
		if cacheErr == nil {
			brsr.cacheID = cacheID
		}
	}

	return n, err
}

// GetCacheID 获取缓存 ID
func (brsr *BodyRecoveryStreamReader) GetCacheID() string {
	brsr.mu.Lock()
	defer brsr.mu.Unlock()
	return brsr.cacheID
}

// defaultLogFunc 默认日志函数
func defaultLogFunc(level, msg string, args ...interface{}) {
	fmt.Printf("[%s] %s", level, msg)
	if len(args) > 0 {
		fmt.Printf(" %v", args)
	}
	fmt.Println()
}

