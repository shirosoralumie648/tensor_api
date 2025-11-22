package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// StreamChunk 流式数据块
type StreamChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

// ChunkChoice 流式选择
type ChunkChoice struct {
	Index        int        `json:"index"`
	Delta        ChunkDelta `json:"delta"`
	FinishReason *string    `json:"finish_reason"`
}

// ChunkDelta 流式增量
type ChunkDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamSender SSE流式发送器
type StreamSender struct {
	w       http.ResponseWriter
	flusher http.Flusher
	ctx     context.Context
}

// NewStreamSender 创建流式发送器
func NewStreamSender(w http.ResponseWriter, ctx context.Context) (*StreamSender, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用Nginx缓冲

	return &StreamSender{
		w:       w,
		flusher: flusher,
		ctx:     ctx,
	}, nil
}

// Send 发送数据块
func (s *StreamSender) Send(chunk *StreamChunk) error {
	// 检查上下文是否取消
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	// 序列化数据
	data, err := json.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	// 发送SSE格式数据
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	// 立即刷新
	s.flusher.Flush()
	return nil
}

// SendText 发送纯文本（用于错误消息等）
func (s *StreamSender) SendText(text string) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
	}

	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", text); err != nil {
		return err
	}

	s.flusher.Flush()
	return nil
}

// SendError 发送错误消息
func (s *StreamSender) SendError(errMsg string) error {
	errorChunk := map[string]interface{}{
		"error": map[string]string{
			"message": errMsg,
			"type":    "stream_error",
		},
	}

	data, _ := json.Marshal(errorChunk)
	return s.SendText(string(data))
}

// Close 关闭流式连接
func (s *StreamSender) Close() error {
	if _, err := fmt.Fprint(s.w, "data: [DONE]\n\n"); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// IsClientDisconnected 检查客户端是否断开连接
func (s *StreamSender) IsClientDisconnected() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}
