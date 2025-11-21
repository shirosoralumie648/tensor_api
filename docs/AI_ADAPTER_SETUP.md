# AI 适配器设置指南

## 概述

Oblivious 使用适配器模式支持多个 AI 提供商。本文档介绍如何配置和添加新的 AI 适配器。

## 支持的提供商

- OpenAI (GPT-3.5, GPT-4)
- Anthropic (Claude)
- Google (Gemini)
- 通义千问
- 文心一言
- 讯飞星火
- Ollama (本地模型)

## 配置 OpenAI

### 1. 获取 API Key

访问 https://platform.openai.com/api-keys

### 2. 添加渠道

```sql
INSERT INTO channels (name, type, base_url, api_key, models, priority, weight, status)
VALUES (
    'OpenAI官方',
    1,  -- OpenAI类型
    'https://api.openai.com/v1',
    'sk-...',
    ARRAY['gpt-3.5-turbo', 'gpt-4', 'gpt-4-turbo'],
    100,
    100,
    1
);
```

### 3. 环境变量配置

```bash
# config.yaml
ai:
  providers:
    openai:
      api_key: sk-...
      base_url: https://api.openai.com/v1
      timeout: 60s
```

## 配置 Claude

### 1. 获取 API Key

访问 https://console.anthropic.com/

### 2. 添加渠道

```sql
INSERT INTO channels (name, type, base_url, api_key, models, priority, weight, status)
VALUES (
    'Claude官方',
    2,
    'https://api.anthropic.com',
    'sk-ant-...',
    ARRAY['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
    90,
    100,
    1
);
```

## 配置本地模型（Ollama）

### 1. 安装 Ollama

```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

### 2. 拉取模型

```bash
ollama pull llama2
ollama pull mistral
```

### 3. 配置渠道

```sql
INSERT INTO channels (name, type, base_url, api_key, models, priority)
VALUES (
    'Ollama本地',
    10,
    'http://localhost:11434',
    'ollama',
    ARRAY['llama2', 'mistral'],
    50
);
```

## 添加新的适配器

### 1. 实现适配器接口

```go
// backend/internal/adapter/custom/client.go
package custom

type CustomAdapter struct {
    client *http.Client
    apiKey string
    baseURL string
}

func NewCustomAdapter(apiKey, baseURL string) *CustomAdapter {
    return &CustomAdapter{
        client: &http.Client{Timeout: 60 * time.Second},
        apiKey: apiKey,
        baseURL: baseURL,
    }
}

func (a *CustomAdapter) CreateCompletion(req *CompletionRequest) (*CompletionResponse, error) {
    // 1. 转换请求格式
    customReq := a.convertRequest(req)
    
    // 2. 发送请求到上游
    resp, err := a.sendRequest(customReq)
    if err != nil {
        return nil, err
    }
    
    // 3. 转换响应格式
    return a.convertResponse(resp), nil
}

func (a *CustomAdapter) CreateCompletionStream(req *CompletionRequest) (io.ReadCloser, error) {
    // 实现流式响应
}
```

### 2. 注册适配器

```go
// backend/internal/relay/service/dispatcher.go
func (d *Dispatcher) getAdapter(channel *model.Channel) (Adapter, error) {
    switch channel.Type {
    case 1:
        return openai.NewAdapter(channel.APIKey, channel.BaseURL), nil
    case 2:
        return claude.NewAdapter(channel.APIKey), nil
    case 100:  // 自定义类型
        return custom.NewCustomAdapter(channel.APIKey, channel.BaseURL), nil
    default:
        return nil, ErrUnsupportedProvider
    }
}
```

## 渠道管理

### 优先级和权重

- **优先级（priority）**：数字越大优先级越高
- **权重（weight）**：用于负载均衡，权重越大被选中概率越高

### 选择策略

```go
func (d *Dispatcher) SelectChannel(model string) (*model.Channel, error) {
    // 1. 筛选支持该模型的渠道
    channels := d.filterChannelsByModel(model)
    
    // 2. 按优先级排序
    sort.Slice(channels, func(i, j int) bool {
        return channels[i].Priority > channels[j].Priority
    })
    
    // 3. 同优先级按权重随机选择
    return d.weightedRandom(channels), nil
}
```

## 健康检查

### 自动检测渠道状态

```go
func (s *ChannelService) HealthCheck() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        channels := s.GetAllChannels()
        
        for _, ch := range channels {
            // 发送测试请求
            start := time.Now()
            err := s.testChannel(ch)
            latency := time.Since(start).Milliseconds()
            
            if err != nil {
                // 标记为不可用
                s.UpdateChannelStatus(ch.ID, StatusDown)
            } else {
                s.UpdateChannelStatus(ch.ID, StatusUp)
                s.UpdateChannelLatency(ch.ID, int(latency))
            }
        }
    }
}
```

## 成本管理

### 配置模型定价

```sql
CREATE TABLE model_pricing (
    model VARCHAR(100) PRIMARY KEY,
    input_price DECIMAL(10, 6),   -- 每1K tokens价格（美元）
    output_price DECIMAL(10, 6),
    currency VARCHAR(10) DEFAULT 'USD'
);

INSERT INTO model_pricing VALUES
    ('gpt-3.5-turbo', 0.0015, 0.002, 'USD'),
    ('gpt-4', 0.03, 0.06, 'USD'),
    ('claude-3-sonnet', 0.003, 0.015, 'USD');
```

### 自动计算成本

```go
func CalculateCost(model string, inputTokens, outputTokens int) (int64, error) {
    pricing, err := db.GetModelPricing(model)
    if err != nil {
        return 0, err
    }
    
    inputCost := float64(inputTokens) / 1000 * pricing.InputPrice
    outputCost := float64(outputTokens) / 1000 * pricing.OutputPrice
    
    totalCostUSD := inputCost + outputCost
    totalCostCNY := totalCostUSD * exchangeRate  // 汇率
    
    return int64(totalCostCNY * 100), nil  // 转为分
}
```

## 错误处理

### 重试策略

```go
func (d *Dispatcher) CallWithRetry(req *CompletionRequest) (*CompletionResponse, error) {
    maxRetries := 3
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        channel, err := d.SelectChannel(req.Model)
        if err != nil {
            return nil, err
        }
        
        adapter := d.getAdapter(channel)
        resp, err := adapter.CreateCompletion(req)
        
        if err == nil {
            return resp, nil
        }
        
        lastErr = err
        
        // 根据错误类型决定是否重试
        if !isRetryableError(err) {
            break
        }
        
        // 标记渠道为临时不可用
        d.markChannelDown(channel.ID)
        
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return nil, lastErr
}
```

## 监控指标

### 记录关键指标

```go
var (
    requestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "relay_requests_total",
        },
        []string{"provider", "model", "status"},
    )
    
    latencyHistogram = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "relay_latency_seconds",
        },
        []string{"provider", "model"},
    )
)
```

## 相关文档

- [架构设计](ARCHITECTURE.md)
- [API 参考](API_REFERENCE.md)
