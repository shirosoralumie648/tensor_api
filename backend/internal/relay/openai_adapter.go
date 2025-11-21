package relay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/pkg/logger"
	"go.uber.org/zap"
)

// OpenAIAdapter OpenAI API 适配器
type OpenAIAdapter struct {
	channel *model.Channel
	client  *http.Client
}

// NewOpenAIAdapter 创建 OpenAI 适配器
func NewOpenAIAdapter(channel *model.Channel) *OpenAIAdapter {
	return &OpenAIAdapter{
		channel: channel,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Chat 调用 OpenAI API 进行对话
func (a *OpenAIAdapter) Chat(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// 构建请求 URL
	url := fmt.Sprintf("%s/v1/chat/completions", a.channel.BaseURL)

	// 序列化请求
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.channel.APIKey))

	// 发送请求
	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", httpResp.StatusCode, string(body))
	}

	// 解析响应
	var resp ChatCompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// ChatStream 调用 OpenAI API 进行流式对话
func (a *OpenAIAdapter) ChatStream(ctx context.Context, req *ChatCompletionRequest, handler func(chunk *ChatCompletionResponse) error) error {
	// 启用流式传输
	req.Stream = true

	// 构建请求 URL
	url := fmt.Sprintf("%s/v1/chat/completions", a.channel.BaseURL)

	// 序列化请求
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.channel.APIKey))

	// 发送请求
	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// 检查状态码
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("API error: status=%d, body=%s", httpResp.StatusCode, string(body))
	}

	// 处理流式响应（SSE 格式）
	// 格式为: data: {...}\n\n
	// 最后一条为: data: [DONE]
	scanner := bufio.NewScanner(httpResp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否是数据行
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// 提取数据部分
		data := strings.TrimPrefix(line, "data: ")

		// 检查是否是终止标记
		if data == "[DONE]" {
			break
		}

		// 解析 JSON 数据
		var chunk ChatCompletionResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			logger.Warn("failed to unmarshal stream chunk", zap.Error(err), zap.String("data", data))
			continue
		}

		// 调用处理函数
		if err := handler(&chunk); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}

	// 检查扫描器错误
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// GetModels 获取支持的模型列表
func (a *OpenAIAdapter) GetModels(ctx context.Context) ([]string, error) {
	// 这个接口可能因不同的 OpenAI 兼容 API 而有所不同
	// 目前返回配置中的支持模型列表
	models := a.channel.SupportModels
	if models == "" {
		// 如果未配置，返回常见的 OpenAI 模型
		return []string{
			"gpt-3.5-turbo",
			"gpt-4",
			"gpt-4-turbo",
			"gpt-4o",
		}, nil
	}
	
	// 解析逗号分隔的模型列表
	// TODO: 实现完整的解析逻辑
	return []string{}, nil
}

