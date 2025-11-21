# AI 供应商适配器设置指南

## 📋 概述

Oblivious 项目实现了一个灵活的 AI 供应商适配器系统，支持多个 AI 提供商（OpenAI、Claude 等）。该系统允许动态切换和组合不同提供商的模型。

---

## 🏗️ 架构设计

### 核心接口

```go
type AIProvider interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamDelta, error)
    ListModels(ctx context.Context) ([]Model, error)
    HealthCheck(ctx context.Context) error
    GetName() string
}
```

### 适配器工厂模式

```go
factory := adapter.NewAdapterFactory()

// 注册 OpenAI
openaiAdapter := openai.NewOpenAIAdapter(os.Getenv("OPENAI_API_KEY"))
factory.Register("openai", openaiAdapter, openaiConfig)

// 注册 Claude
claudeAdapter := claude.NewClaudeAdapter(os.Getenv("ANTHROPIC_API_KEY"))
factory.Register("claude", claudeAdapter, claudeConfig)
```

---

## 🔧 集成步骤

### Step 1: 安装依赖

```bash
# 进入后端目录
cd backend

# 添加 Go 依赖
go get github.com/sashabaranov/go-openai
go get github.com/anthropics/sdk-go
```

### Step 2: 配置环境变量

创建 `.env` 文件：

```bash
# OpenAI 配置
OPENAI_API_KEY=sk-...
OPENAI_BASE_URL=https://api.openai.com/v1

# Claude (Anthropic) 配置
ANTHROPIC_API_KEY=sk-ant-...

# 其他配置
CHAT_TIMEOUT=30
MAX_RETRIES=3
```

### Step 3: 初始化适配器

在 `backend/main.go` 中：

```go
package main

import (
	"os"
	
	"github.com/gin-gonic/gin"
	"oblivious/internal/adapter"
	openai_adapter "oblivious/internal/adapter/openai"
	claude_adapter "oblivious/internal/adapter/claude"
	"oblivious/internal/handler"
)

func main() {
	// 创建适配器工厂
	factory := adapter.NewAdapterFactory()

	// 注册 OpenAI 适配器
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey != "" {
		openaiAdapter := openai_adapter.NewOpenAIAdapter(openaiKey)
		openaiConfig := &adapter.ProviderConfig{
			Name:   "openai",
			APIKey: openaiKey,
			Models: getOpenAIModels(),
		}
		factory.Register("openai", openaiAdapter, openaiConfig)
	}

	// 注册 Claude 适配器
	claudeKey := os.Getenv("ANTHROPIC_API_KEY")
	if claudeKey != "" {
		claudeAdapter := claude_adapter.NewClaudeAdapter(claudeKey)
		claudeConfig := &adapter.ProviderConfig{
			Name:   "claude",
			APIKey: claudeKey,
			Models: getClaudeModels(),
		}
		factory.Register("claude", claudeAdapter, claudeConfig)
	}

	// 创建路由
	router := gin.Default()

	// 初始化处理器
	chatHandler := handler.NewChatHandler(
		factory,
		billingService,
		channelService,
		auditService,
	)

	// 注册路由
	api := router.Group("/v1")
	{
		api.POST("/chat/completions", chatHandler.ChatCompletion)
		api.POST("/chat/stream", chatHandler.ChatCompletionStream)
		api.GET("/models", chatHandler.ListModels)
		api.GET("/models/:model_id", chatHandler.GetModel)
	}

	// 健康检查
	router.GET("/health", chatHandler.HealthCheck)

	// 启动服务器
	router.Run(":8000")
}
```

### Step 4: 使用适配器

```go
// 获取适配器
provider := factory.Get("openai")

// 构建请求
req := &adapter.ChatRequest{
	Model: "gpt-4",
	Messages: []adapter.Message{
		{Role: "user", Content: "Hello!"},
	},
	Temperature: 0.7,
	MaxTokens:   2048,
	Stream:      false,
}

// 调用 API
resp, err := provider.Chat(ctx, req)
if err != nil {
	// 处理错误
	log.Printf("Error: %v", err)
}

// 使用响应
fmt.Printf("Content: %s\n", resp.Content)
fmt.Printf("Tokens: %d\n", resp.Tokens.TotalTokens)
```

---

## 🌊 流式响应使用

```go
// 获取流式响应
deltaCh, err := provider.ChatStream(ctx, req)
if err != nil {
	log.Fatal(err)
}

// 处理增量数据
for delta := range deltaCh {
	if delta.Error != nil {
		log.Printf("Stream error: %v", delta.Error)
		break
	}

	if delta.Content != "" {
		fmt.Print(delta.Content)
	}

	if delta.Done {
		fmt.Println("\n[Stream completed]")
		break
	}
}
```

---

## 📊 API 端点

### 非流式聊天

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "temperature": 0.7,
    "max_tokens": 2048
  }'
```

### 流式聊天

```bash
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": true
  }'
```

### 获取模型列表

```bash
curl -X GET http://localhost:8000/v1/models \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 健康检查

```bash
curl -X GET http://localhost:8000/health/chat
```

---

## 🛡️ 错误处理

### 适配器错误类型

```go
type AdapterError struct {
	Provider string  // 提供商名称
	Code     string  // 错误代码
	Message  string  // 错误消息
	Err      error   // 原始错误
}
```

### 错误处理示例

```go
resp, err := provider.Chat(ctx, req)
if err != nil {
	if adapterErr, ok := err.(*adapter.AdapterError); ok {
		log.Printf("Provider: %s, Code: %s, Message: %s",
			adapterErr.Provider,
			adapterErr.Code,
			adapterErr.Message,
		)
	} else {
		log.Printf("Unexpected error: %v", err)
	}
}
```

---

## 🧪 测试

### 单元测试

```bash
# 测试 OpenAI 适配器
go test ./internal/adapter/openai -v

# 测试 Claude 适配器
go test ./internal/adapter/claude -v

# 测试适配器工厂
go test ./internal/adapter -v
```

### 集成测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./internal/handler -v
```

### 示例测试代码

```go
// backend/internal/adapter/openai/openai_test.go

package openai

import (
	"context"
	"testing"

	"oblivious/internal/adapter"
)

func TestOpenAIChat(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	adapter := NewOpenAIAdapter(apiKey)

	req := &adapter.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []adapter.Message{
			{Role: "user", Content: "Hello!"},
		},
		Temperature: 0.7,
		MaxTokens:   100,
	}

	resp, err := adapter.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Response content is empty")
	}

	if resp.Tokens.TotalTokens == 0 {
		t.Error("Token count should be greater than 0")
	}
}
```

---

## 🔄 模型切换策略

### 简单切换

```go
// 根据用户选择切换模型
func switchModel(factory *adapter.AdapterFactory, modelID string) string {
	provider := factory.FindProviderByModel(modelID)
	return provider
}
```

### 负载均衡

```go
// 根据负载选择提供商
type ProviderLoadBalancer struct {
	providers map[string]int // 提供商名称 -> 当前负载
}

func (lb *ProviderLoadBalancer) SelectProvider() string {
	// 选择负载最低的提供商
	minLoad := int(^uint(0) >> 1)
	var selected string

	for name, load := range lb.providers {
		if load < minLoad {
			minLoad = load
			selected = name
		}
	}

	return selected
}
```

### 成本优化

```go
// 根据成本选择模型
func selectCheapestModel(factory *adapter.AdapterFactory, capability string) *adapter.Model {
	allModels := factory.GetAllModels()
	
	var cheapest *adapter.Model
	for _, models := range allModels {
		for _, model := range models {
			if cheapest == nil || model.CostPer1KPrompt < cheapest.CostPer1KPrompt {
				cheapest = &model
			}
		}
	}

	return cheapest
}
```

---

## 📈 监控和日志

### 记录 API 调用

```go
// 在处理器中记录调用
auditService.RecordAPICall(&AuditRecord{
	UserID:    userID,
	Endpoint:  "/v1/chat/completions",
	Method:    "POST",
	Model:     req.Model,
	Provider:  provider.GetName(),
	Status:    200,
	Duration:  time.Since(startTime),
	Timestamp: time.Now(),
})
```

### 计费记录

```go
// 记录用户使用情况
billingService.RecordUsage(&BillingRecord{
	UserID:           userID,
	Model:            req.Model,
	Provider:         provider.GetName(),
	PromptTokens:     resp.Tokens.PromptTokens,
	CompletionTokens: resp.Tokens.CompletionTokens,
	Cost:             calculateCost(provider, resp),
	Timestamp:        time.Now(),
})
```

---

## 🚀 生产部署

### Docker 支持

```dockerfile
# Dockerfile

FROM golang:1.21 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o oblivious-backend ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /build/oblivious-backend .

# 设置环境变量
ENV OPENAI_API_KEY=""
ENV ANTHROPIC_API_KEY=""

EXPOSE 8000

CMD ["./oblivious-backend"]
```

### Docker Compose

```yaml
# docker-compose.yml

version: '3.8'

services:
  backend:
    build: .
    ports:
      - "8000:8000"
    environment:
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
    networks:
      - oblivious
    restart: unless-stopped

  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: postgres
    networks:
      - oblivious

networks:
  oblivious:
```

---

## 🤝 扩展新提供商

### 步骤 1: 创建新适配器

```go
// backend/internal/adapter/gemini/gemini.go

package gemini

import (
	"context"
	"oblivious/internal/adapter"
)

type GeminiAdapter struct {
	client *genai.Client
	config *adapter.ProviderConfig
}

func NewGeminiAdapter(apiKey string) *GeminiAdapter {
	// 初始化 Gemini 客户端
}

func (a *GeminiAdapter) Chat(ctx context.Context, req *adapter.ChatRequest) (*adapter.ChatResponse, error) {
	// 实现 Chat 方法
}

// 实现其他必需方法...
```

### 步骤 2: 注册适配器

```go
// 在 main.go 中
geminiAdapter := gemini.NewGeminiAdapter(os.Getenv("GEMINI_API_KEY"))
factory.Register("gemini", geminiAdapter, geminiConfig)
```

### 步骤 3: 添加测试

```go
// backend/internal/adapter/gemini/gemini_test.go

func TestGeminiChat(t *testing.T) {
	// 测试实现
}
```

---

## 📚 参考资源

- [OpenAI API 文档](https://platform.openai.com/docs)
- [Claude API 文档](https://docs.anthropic.com)
- [Gemini API 文档](https://ai.google.dev)

---

## ✅ 实现清单

- [x] 适配器接口定义
- [x] 适配器工厂实现
- [x] OpenAI 适配器
- [x] Claude 适配器
- [x] 聊天处理器
- [x] API 端点
- [x] 流式响应支持
- [x] 错误处理
- [x] 健康检查
- [x] 文档

---

**文档版本**: v1.0  
**最后更新**: 2024 年 11 月 21 日  
**作者**: Oblivious 开发团队

