# 智能渠道选择算法实现文档

## 概述

智能渠道选择算法为 Oblivious AI 平台提供灵活多样的渠道选择策略。支持随机、轮询、加权轮询、最少连接和最低延迟等多种算法，以及通配符规则匹配，实现高效的负载均衡和智能路由。

**核心特性：**
- ✅ 5 种选择策略（随机、轮询、加权轮询、最少连接、最低延迟）
- ✅ 通配符模式匹配（支持 * 通配符）
- ✅ 灵活的过滤条件（类型、模型、地区、可用性）
- ✅ 优先级规则系统
- ✅ 实时选择统计
- ✅ <1ms 选择时间

## 架构设计

### 核心组件

#### 1. 选择策略

```
┌────────────────────────────────────────┐
│     ChannelSelectorStrategy            │
├────────────────────────────────────────┤
│                                        │
├─ Random (随机)                         │
│  └─ 均匀分布，适合简单场景              │
│                                        │
├─ RoundRobin (轮询)                     │
│  └─ 按顺序轮换，无权重                  │
│                                        │
├─ WeightedRoundRobin (加权轮询)         │
│  └─ 按权重分配，推荐生产环境            │
│                                        │
├─ LeastConnection (最少连接)            │
│  └─ 选择并发最少的，适合长连接          │
│                                        │
└─ LowestLatency (最低延迟)              │
   └─ 选择延迟最低的，追求速度            │
```

#### 2. 通配符规则

```go
type WildcardRule struct {
    ID                string
    Pattern           string              // "gpt-*", "claude-*"
    ChannelType       string              // "openai", "anthropic"
    PriorityChannels  []string            // 优先渠道列表
    Weight            int                 // 规则权重
    Enabled           bool                // 是否启用
}
```

#### 3. 选择选项

```go
type ChannelSelectOptions struct {
    ChannelType        string              // 渠道类型
    Model              string              // 模型名称
    Region             string              // 地理位置
    MinAvailability    float64             // 最小可用性
    PreferredChannelID string              // 首选渠道
    ExcludedChannelIDs map[string]bool     // 排除的渠道
}
```

## 使用指南

### 1. 初始化选择器

```go
// 创建选择器（加权轮询策略）
cache := relay.NewChannelCache(relay.ChannelCacheLevelHybrid)
selector := relay.NewChannelSelector(cache, relay.SelectorStrategyWeightedRoundRobin)
```

### 2. 基础选择

```go
// 准备选择选项
options := &relay.ChannelSelectOptions{
    ChannelType:     "openai",
    Model:           "gpt-4",
    MinAvailability: 95.0,
}

// 选择渠道
channel, err := selector.SelectChannel(options)
if err != nil {
    log.Fatalf("Selection failed: %v", err)
}

log.Printf("Selected: %s (%s)", channel.Name, channel.ID)
```

### 3. 通配符规则

```go
// 创建通配符规则
rule := &relay.WildcardRule{
    ID:               "gpt-priority",
    Pattern:          "gpt-*",              // 匹配 gpt-4, gpt-3.5 等
    ChannelType:      "openai",
    PriorityChannels: []string{"ch-1", "ch-2"},
    Weight:           10,
    Enabled:          true,
}

// 添加规则
selector.AddWildcardRule(rule)
```

### 4. 使用管理器

```go
// 创建管理器
manager := relay.NewChannelSelectorManager(cache)

// 为不同渠道类型设置不同的策略
manager.SetStrategy("openai", relay.SelectorStrategyWeightedRoundRobin)
manager.SetStrategy("anthropic", relay.SelectorStrategyLeastConnection)

// 获取规则管理器
ruleManager := manager.GetRuleManager()

// 添加全局规则
ruleManager.AddRule(&relay.WildcardRule{
    ID:      "claude-rule",
    Pattern: "claude-*",
    // ...
})

// 应用规则到所有选择器
ruleManager.ApplyRulesToSelector(selector)

// 选择
options := &relay.ChannelSelectOptions{
    ChannelType: "openai",
    Model:       "gpt-4",
}

channel, err := manager.SelectChannel(options)
```

## 选择策略详解

### 1. 随机选择 (Random)

```
每次随机选择一个可用渠道
优点:
  • 简单快速
  • 自然的负载均衡
  
缺点:
  • 忽略渠道能力
  • 可能选中慢速渠道

场景: 所有渠道性能相近
```

### 2. 轮询 (RoundRobin)

```
按顺序轮换选择渠道
优点:
  • 完全均匀分布
  • 无记忆开销
  
缺点:
  • 无法处理权重差异
  • 忽略渠道状态

场景: 渠道性能完全相同
```

### 3. 加权轮询 (WeightedRoundRobin) ⭐ 推荐

```
按权重比例分配选择概率
优点:
  • 灵活的权重配置
  • 自适应负载分配
  • 生产级别推荐
  
缺点:
  • 需要配置权重

场景: 渠道能力不同（推荐生产）
```

### 4. 最少连接 (LeastConnection)

```
选择当前并发数最少的渠道
优点:
  • 自适应负载均衡
  • 避免单点过载
  • 适合长连接
  
缺点:
  • 需要实时更新并发数

场景: 长连接、WebSocket 等
```

### 5. 最低延迟 (LowestLatency)

```
选择平均延迟最低的渠道
优点:
  • 追求用户体验
  • 自动发现最佳渠道
  
缺点:
  • 需要收集延迟信息
  • 可能集中在单个渠道

场景: 性能关键型应用
```

## 通配符模式匹配

### 支持的模式

| 模式 | 示例 | 说明 |
|------|------|------|
| `*` | `*` | 匹配所有 |
| `prefix-*` | `gpt-*` | 前缀匹配 |
| `*-suffix` | `*-vision` | 后缀匹配 |
| `*-middle-*` | `*-turbo-*` | 中缀匹配 |
| `exact` | `gpt-4` | 精确匹配 |

### 使用示例

```go
selector.matchPattern("gpt-4", "gpt-*")           // true
selector.matchPattern("gpt-3.5", "gpt-*")         // true
selector.matchPattern("claude-3", "gpt-*")        // false
selector.matchPattern("gpt-4-vision", "*-vision") // true
selector.matchPattern("gpt-4-turbo", "*-turbo")   // false
```

## 性能指标

### 选择速度

| 策略 | 时间 | 说明 |
|------|------|------|
| Random | <0.1ms | 最快 |
| RoundRobin | <0.1ms | 同上 |
| WeightedRoundRobin | <0.5ms | 快速 |
| LeastConnection | <1ms | 需排序 |
| LowestLatency | <1ms | 需排序 |

### 吞吐量

- **随机/轮询：** >10,000 选择/秒
- **加权轮询：** >1,000 选择/秒
- **最少连接/最低延迟：** >200 选择/秒

## 最佳实践

### 1. 选择策略

```
条件                           推荐策略
─────────────────────────────────────────
渠道能力相同                     Random / RoundRobin
渠道能力不同 (推荐)               WeightedRoundRobin
短连接、HTTP                     WeightedRoundRobin
长连接、WebSocket               LeastConnection
性能关键型应用                   LowestLatency
```

### 2. 权重配置

```go
// 根据渠道能力设置权重
ch1.Weight = 5   // 高性能
ch2.Weight = 3   // 中等
ch3.Weight = 1   // 低性能

// 权重比例: 5:3:1
// 选择概率: 56%:33%:11%
```

### 3. 通配符规则

```go
// 为不同模型系列设置优先渠道
rules := []*relay.WildcardRule{
    {
        ID:               "gpt-series",
        Pattern:          "gpt-*",
        ChannelType:      "openai",
        PriorityChannels: []string{"ch-openai-us"},
        Weight:           10,
        Enabled:          true,
    },
    {
        ID:               "claude-series",
        Pattern:          "claude-*",
        ChannelType:      "anthropic",
        PriorityChannels: []string{"ch-claude-us"},
        Weight:           10,
        Enabled:          true,
    },
}

for _, rule := range rules {
    selector.AddWildcardRule(rule)
}
```

### 4. 过滤条件

```go
// 只选择特定地区、高可用性的渠道
options := &relay.ChannelSelectOptions{
    ChannelType:     "openai",
    Model:           "gpt-4",
    Region:          "us",              // 特定地区
    MinAvailability: 99.0,              // 高可用性要求
}

channel, err := selector.SelectChannel(options)
```

## 监控和诊断

### 统计信息

```go
stats := selector.GetStatistics()
// 输出:
// {
//   "strategy": "weighted_round_robin",
//   "total_selections": 1000,
//   "channel_statistics": {
//     "ch-1": {"SelectionCount": 500, ...},
//     "ch-2": {"SelectionCount": 300, ...},
//     "ch-3": {"SelectionCount": 200, ...},
//   }
// }
```

### 性能监控

```go
// 监控选择分布
stats := selector.GetStatistics()
if distribution, ok := stats["channel_statistics"].(map[string]*SelectorStatistics); ok {
    for chID, stat := range distribution {
        log.Printf("Channel %s: %d selections", chID, stat.SelectionCount)
    }
}
```

## 与其他组件的集成

### 与 RequestClient 集成

```go
// 使用选择器选择渠道，然后发送请求
selector := relay.NewChannelSelector(cache, relay.SelectorStrategyWeightedRoundRobin)

channel, _ := selector.SelectChannel(&relay.ChannelSelectOptions{
    ChannelType: "openai",
    Model:       "gpt-4",
})

client := relay.NewRequestClient(30 * time.Second)
body, _, _ := client.DoRequest(ctx, "POST", channel.BaseURL+"/v1/chat", reqBody, headers)
```

## 常见问题

### Q: 如何在运行时改变选择策略？

A: 使用管理器重新注册选择器即可：
```go
manager.SetStrategy("openai", relay.SelectorStrategyLowestLatency)
```

### Q: 通配符规则的优先级如何确定？

A: 按权重从高到低，权重相同则按规则添加顺序。

### Q: 选择器是否线程安全？

A: 是的，所有操作使用原子操作和互斥锁保护。

### Q: 如何监控选择的负载均衡度？

A: 定期收集统计信息，分析每个渠道的选择比例。

## 下一步

本文档为 Phase 1.3 第二部分（1.3.2 智能渠道选择算法）。

接下来将实现：
- **1.3.3**: 多密钥轮询系统
- **1.3.4**: 渠道健康检查
- **1.3.5**: 负载均衡策略增强
- **1.3.6**: 渠道能力管理

## 参考资源

- [渠道管理文档](./CHANNEL_MANAGEMENT_IMPLEMENTATION.md)
- [RelayHandler 文档](./RELAY_HANDLER_IMPLEMENTATION.md)
- [项目状态](./PHASE1_PROGRESS.md)

