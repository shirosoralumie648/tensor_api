# 计费系统指南

## 概述

Oblivious 计费系统采用预付费模式，用户充值后按实际使用量扣费。系统支持实时扣费、账单记录和消费统计。

## 核心概念

### 额度单位

- **内部单位**：分（1元 = 100分）
- **显示单位**：元
- **美元兑换**：根据实时汇率

### 计费模型

```
总成本 = (输入token数 × 输入单价) + (输出token数 × 输出单价)
```

## 定价策略

### 默认定价（示例）

| 模型 | 输入价格（$/1K tokens）| 输出价格（$/1K tokens）|
|------|---------------------|---------------------|
| GPT-3.5 Turbo | 0.0015 | 0.002 |
| GPT-4 | 0.03 | 0.06 |
| Claude 3 Sonnet | 0.003 | 0.015 |

### 加价率

平台可配置加价率，例如：

```
用户价格 = 上游成本 × (1 + 加价率)
```

## 计费流程

### 1. 请求前检查

```go
// 检查用户额度
quota, err := billingService.CheckQuota(userID)
if quota < estimatedCost {
    return ErrInsufficientQuota
}
```

### 2. 请求中预扣费

```go
// 锁定额度（可选）
err := billingService.LockQuota(userID, estimatedCost)
```

### 3. 请求后扣费

```go
// 根据实际使用扣费
actualCost := calculateCost(inputTokens, outputTokens, model)
err := billingService.DeductQuota(userID, actualCost, billingLog)
```

### 4. 异步记录账单

```go
// 发送 MQ 消息
err := billingService.PublishBillingLog(billingLog)

// 消费者处理
func (c *BillingConsumer) ProcessBillingLog(log *BillingLog) {
    // 插入数据库
    db.Create(log)
    
    // 更新统计
    updateStatistics(log.UserID, log.Cost)
}
```

## 数据模型

### 用户额度

```go
type User struct {
    Quota      int64  // 当前可用额度（分）
    TotalQuota int64  // 累计获得额度
    UsedQuota  int64  // 已使用额度
}
```

### 计费日志

```go
type BillingLog struct {
    ID           int
    UserID       int
    SessionID    uuid.UUID
    MessageID    uuid.UUID
    Model        string
    InputTokens  int
    OutputTokens int
    TotalTokens  int
    Cost         int64   // 分
    CostUSD      float64 // 美元
    CreatedAt    time.Time
}
```

### 额度变更日志

```go
type QuotaLog struct {
    UserID        int
    OperationType string  // recharge, consume, refund, gift
    Amount        int64   // 正数=增加，负数=减少
    BalanceBefore int64
    BalanceAfter  int64
    Reason        string
}
```

## API 接口

### 查询额度

```http
GET /api/v1/billing/quota
Authorization: Bearer {token}
```

### 充值

```http
POST /api/v1/billing/recharge
{
  "amount": 100,
  "payment_method": "alipay"
}
```

### 消费记录

```http
GET /api/v1/billing/logs?start_date=2024-01-01&end_date=2024-01-31
```

## 管理功能

### 赠送额度

```bash
# 管理员操作
curl -X POST https://api.oblivious.ai/admin/users/{user_id}/gift-quota \
  -H "Authorization: Bearer admin-token" \
  -d '{"amount": 10000, "reason": "新用户奖励"}'
```

### 退款

```bash
curl -X POST https://api.oblivious.ai/admin/refund \
  -d '{"user_id": 123, "amount": 5000, "reason": "服务故障"}'
```

## 监控告警

### 关键指标

- 每日充值金额
- 每日消费金额
- 用户余额分布
- 异常扣费检测

### 告警规则

```yaml
alerts:
  - name: LowQuota
    expr: user_quota < 1000
    message: "用户额度不足"
  
  - name: HighCost
    expr: rate(billing_cost[5m]) > 1000
    message: "消费异常增长"
```

## 安全措施

### 防止重复扣费

```go
// 使用消息 ID 幂等性
func (s *BillingService) DeductQuota(messageID string, cost int64) error {
    // 检查是否已扣费
    exists := db.Where("message_id = ?", messageID).First(&BillingLog{}).Error
    if exists == nil {
        return ErrAlreadyCharged
    }
    
    // 执行扣费
    // ...
}
```

### 额度一致性

使用数据库事务确保额度变更的原子性。

## 相关文档

- [架构设计](ARCHITECTURE.md)
- [API 参考](API_REFERENCE.md)
