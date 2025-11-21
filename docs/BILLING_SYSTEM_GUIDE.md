# 高级计费系统完整指南

## 📋 概述

Oblivious 项目实现了一个企业级的高级计费系统，支持：
- ✅ Token 级别的精确计费
- ✅ 多层订阅计划 (Basic, Pro, Enterprise)
- ✅ 优惠券和促销码系统
- ✅ 自动发票生成
- ✅ 配额告警系统
- ✅ 自动充值功能

---

## 🏗️ 架构设计

### 核心模型

```
Subscription (订阅)
    ├── SubscriptionPlan (计划)
    │   ├── monthly_quota (月配额)
    │   └── monthly_price (月价)
    ├── BillingRecord (记录)
    │   ├── tokens_used (使用的 Token)
    │   └── cost (成本)
    ├── Invoice (发票)
    │   ├── items (项目)
    │   └── status (状态)
    └── BillingSettings (设置)
        ├── auto_topup (自动充值)
        └── alert_threshold (告警阈值)

ModelPrice (模型价格)
    ├── model_id
    ├── prompt_price_per_k
    └── completion_price_per_k

Coupon (优惠券)
    ├── code
    ├── type (percentage/fixed)
    └── value

CouponUsage (优惠券使用)
    ├── coupon_id
    ├── user_id
    └── discount_amount
```

---

## 💾 数据库模型

### SubscriptionPlan (订阅计划表)

```sql
CREATE TABLE subscription_plans (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    monthly_quota BIGINT NOT NULL,
    monthly_price FLOAT NOT NULL,
    extra_token_price FLOAT NOT NULL,
    support_level VARCHAR(50),
    max_concurrent_req INT,
    rate_limit_per_min INT,
    is_active BOOLEAN DEFAULT true,
    display_order INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- 预置计划
INSERT INTO subscription_plans VALUES
('plan-basic', 'Basic', 'Perfect for getting started', 100000, 9.99, 0.00002, 'email', 5, 60, true, 1, NOW(), NOW()),
('plan-pro', 'Pro', 'For professionals', 1000000, 99.99, 0.00015, 'priority', 20, 300, true, 2, NOW(), NOW()),
('plan-enterprise', 'Enterprise', 'For large organizations', 10000000, 999.99, 0.0001, 'dedicated', 100, 1000, true, 3, NOW(), NOW());
```

### Subscription (用户订阅表)

```sql
CREATE TABLE subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    plan_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    billing_cycle VARCHAR(50),
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    tokens_used BIGINT DEFAULT 0,
    auto_renew BOOLEAN DEFAULT true,
    next_billing_date TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(id),
    INDEX idx_user_status (user_id, status)
);
```

### BillingRecord (计费记录表)

```sql
CREATE TABLE billing_records (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    subscription_id VARCHAR(255),
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50),
    request_id VARCHAR(255) UNIQUE,
    prompt_tokens BIGINT NOT NULL,
    completion_tokens BIGINT NOT NULL,
    total_tokens BIGINT NOT NULL,
    cost_usd FLOAT NOT NULL,
    discount_rate FLOAT DEFAULT 1.0,
    discount_amount FLOAT DEFAULT 0,
    final_cost FLOAT NOT NULL,
    coupon_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'completed',
    billing_month VARCHAR(7),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX idx_user_month (user_id, billing_month),
    INDEX idx_created_date (created_at)
);
```

### Invoice (发票表)

```sql
CREATE TABLE invoices (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    billing_month VARCHAR(7) NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    billing_address TEXT,
    amount FLOAT NOT NULL,
    discount_amount FLOAT DEFAULT 0,
    tax_amount FLOAT DEFAULT 0,
    final_amount FLOAT NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    issued_at TIMESTAMP,
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    payment_method VARCHAR(50),
    transaction_id VARCHAR(255),
    invoice_number VARCHAR(50) UNIQUE,
    items JSON,
    notes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX idx_user_month (user_id, billing_month),
    INDEX idx_status (status)
);
```

### Coupon (优惠券表)

```sql
CREATE TABLE coupons (
    id VARCHAR(255) PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    value FLOAT NOT NULL,
    max_uses BIGINT NOT NULL,
    used_count BIGINT DEFAULT 0,
    max_usage_per_user INT DEFAULT 1,
    min_purchase_amount FLOAT DEFAULT 0,
    valid_from TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    description TEXT,
    created_by VARCHAR(255),
    applicable_plans JSON,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_active (is_active)
);
```

### BillingSettings (计费设置表)

```sql
CREATE TABLE billing_settings (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    billing_email VARCHAR(255),
    billing_address TEXT,
    tax_id VARCHAR(50),
    currency VARCHAR(3) DEFAULT 'USD',
    auto_topup BOOLEAN DEFAULT true,
    auto_topup_threshold BIGINT DEFAULT 0,
    auto_topup_amount FLOAT DEFAULT 100,
    invoice_frequency VARCHAR(50) DEFAULT 'monthly',
    send_invoice_email BOOLEAN DEFAULT true,
    alert_threshold BIGINT DEFAULT 0,
    enable_alerts BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## 🔌 API 端点

### 计费统计

```bash
GET /v1/billing/stats
Authorization: Bearer YOUR_TOKEN
```

**响应**:
```json
{
    "current_month_cost": 45.50,
    "current_month_tokens": 1250000,
    "subscription_plan": "pro",
    "tokens_quota": 1000000,
    "tokens_used": 750000,
    "tokens_remaining": 250000,
    "total_cost": 245.50
}
```

### 获取发票列表

```bash
GET /v1/billing/invoices?page=1&page_size=10
Authorization: Bearer YOUR_TOKEN
```

**响应**:
```json
{
    "invoices": [
        {
            "id": "inv-001",
            "invoice_number": "INV-user-2024-11",
            "billing_month": "2024-11",
            "status": "paid",
            "amount": 150.50,
            "issued_at": "2024-11-01T00:00:00Z",
            "due_date": "2024-11-30T00:00:00Z",
            "paid_at": "2024-11-15T00:00:00Z"
        }
    ],
    "total": 1
}
```

### 应用优惠券

```bash
POST /v1/billing/coupons/apply
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
    "coupon_code": "SAVE50"
}
```

**响应**:
```json
{
    "message": "coupon applied successfully",
    "coupon": {
        "code": "SAVE50",
        "type": "percentage",
        "value": 50
    },
    "discount": {
        "type": "percentage",
        "value": 50,
        "description": "Save 50%"
    }
}
```

### 生成发票

```bash
POST /v1/billing/invoices/generate
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
    "billing_month": "2024-11"
}
```

**响应**:
```json
{
    "id": "inv-001",
    "user_id": "user-123",
    "billing_month": "2024-11",
    "status": "issued",
    "amount": 150.50,
    "final_amount": 150.50,
    "currency": "USD",
    "invoice_number": "INV-user-2024-11",
    "items": [
        {
            "model": "gpt-4",
            "quantity": 500000,
            "unit_price": 0.03,
            "amount": 15.00,
            "description": "AI API usage - gpt-4"
        },
        {
            "model": "gpt-3.5-turbo",
            "quantity": 750000,
            "unit_price": 0.0015,
            "amount": 1.13,
            "description": "AI API usage - gpt-3.5-turbo"
        }
    ],
    "issued_at": "2024-11-01T00:00:00Z",
    "due_date": "2024-12-01T00:00:00Z"
}
```

### 获取订阅计划

```bash
GET /v1/billing/plans
```

**响应**:
```json
{
    "plans": [
        {
            "id": "plan-basic",
            "name": "Basic",
            "monthly_quota": 100000,
            "monthly_price": 9.99,
            "extra_token_price": 0.00002,
            "features": [
                "Up to 100K tokens per month",
                "Email support",
                "Basic API access"
            ]
        },
        {
            "id": "plan-pro",
            "name": "Pro",
            "monthly_quota": 1000000,
            "monthly_price": 99.99,
            "extra_token_price": 0.00015,
            "features": [
                "Up to 1M tokens per month",
                "Priority email support",
                "Advanced API access",
                "10% discount on overage"
            ]
        },
        {
            "id": "plan-enterprise",
            "name": "Enterprise",
            "monthly_quota": 10000000,
            "monthly_price": 999.99,
            "extra_token_price": 0.0001,
            "features": [
                "Up to 10M tokens per month",
                "24/7 phone support",
                "Dedicated account manager"
            ]
        }
    ]
}
```

### 检查配额警告

```bash
GET /v1/billing/quota-warning
Authorization: Bearer YOUR_TOKEN
```

**响应**:
```json
{
    "quota_warning": false,
    "message": "Your quota is sufficient."
}
```

### 更新计费设置

```bash
PUT /v1/billing/settings
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
    "billing_email": "billing@example.com",
    "auto_topup": true,
    "auto_topup_threshold": 10000,
    "auto_topup_amount": 100.0,
    "enable_alerts": true,
    "alert_threshold": 50000
}
```

---

## 🔧 集成示例

### 1. 记录 Token 使用

```go
// 在 AI 适配器响应后调用
record := &model.BillingRecord{
    ID:               generateID(),
    UserID:           userID,
    Model:            req.Model,
    Provider:         provider.GetName(),
    RequestID:        generateRequestID(),
    PromptTokens:     resp.Tokens.PromptTokens,
    CompletionTokens: resp.Tokens.CompletionTokens,
    TotalTokens:      resp.Tokens.TotalTokens,
    BillingMonth:     time.Now().Format("2006-01"),
    CreatedAt:        time.Now(),
}

if err := billingService.RecordUsage(ctx, record); err != nil {
    log.Printf("failed to record usage: %v", err)
}
```

### 2. 计算成本

```go
cost, err := billingService.CalculateCost(ctx, userID, model, promptTokens, completionTokens)
if err != nil {
    return err
}

// cost 现在包含了所有折扣
fmt.Printf("Cost: $%.2f\n", cost)
```

### 3. 应用优惠券

```go
coupon, err := billingService.ApplyCoupon(ctx, userID, couponCode)
if err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}

// 优惠券已应用，可以重新计算成本
```

### 4. 生成发票

```go
invoice, err := billingService.CreateInvoice(ctx, userID, "2024-11")
if err != nil {
    return err
}

// 发票已生成，可以发送给用户
fmt.Printf("Invoice: %s\n", invoice.InvoiceNumber)
```

---

## 📊 计费流程

### 标准计费流程

```
1. 用户发起 API 请求
   ↓
2. AI 适配器处理请求
   ↓
3. 获取响应的 Token 数
   ↓
4. 查询模型价格
   ↓
5. 计算成本（应用用户折扣）
   ↓
6. 记录计费
   ↓
7. 月末自动生成发票
   ↓
8. 发送给用户
```

### 优惠券应用流程

```
1. 用户提交优惠券代码
   ↓
2. 验证优惠券有效性
   ↓
3. 检查使用限制
   ↓
4. 记录优惠券使用
   ↓
5. 计算折扣金额
   ↓
6. 更新用户账单
```

### 自动充值流程

```
1. 检查用户设置（启用自动充值）
   ↓
2. 获取用户当前 Token 余额
   ↓
3. 检查是否低于阈值
   ↓
4. 如果是，触发支付系统
   ↓
5. 记录充值交易
   ↓
6. 更新用户 Token 余额
```

---

## 💳 支付集成

### 集成 Stripe

```go
// 示例：使用 Stripe 处理支付

import "github.com/stripe/stripe-go/v72"

func chargeUser(userID string, amount float64) error {
    // 获取用户的 Stripe 客户 ID
    customer, err := getStripeCustomer(userID)
    if err != nil {
        return err
    }

    // 创建收费
    params := &stripe.ChargeParams{
        Amount:      stripe.Int64(int64(amount * 100)), // 转换为美分
        Currency:    stripe.String(string(stripe.CurrencyUSD)),
        Customer:    stripe.String(customer.ID),
        Description: stripe.String("Oblivious AI API Usage"),
    }

    charge, err := charge.New(params)
    if err != nil {
        return err
    }

    // 记录交易
    return recordTransaction(userID, charge.ID, amount)
}
```

### 集成 PayPal

```go
// 示例：使用 PayPal 处理支付

import "github.com/plutov/paypal/v4"

func chargeViaPayPal(userID string, amount float64) error {
    client, err := paypal.NewClient()
    if err != nil {
        return err
    }

    // 创建订单
    order, err := client.CreateOrder(
        paypal.OrderIntentCapture,
        []paypal.Item{
            {
                Name:     "Oblivious AI Credits",
                Quantity: 1,
                Price:    fmt.Sprintf("%.2f", amount),
                Currency: "USD",
            },
        },
    )
    if err != nil {
        return err
    }

    // 返回订单 ID 给前端进行支付
    return nil
}
```

---

## 📈 监控和报告

### 获取计费报告

```go
func (s *AdvancedBillingService) GetMonthlyReport(ctx context.Context, month string) (map[string]interface{}, error) {
    var totalRevenue float32
    var totalTokens int64
    var recordCount int64

    s.db.WithContext(ctx).
        Model(&model.BillingRecord{}).
        Where("billing_month = ?", month).
        Select("SUM(final_cost), SUM(total_tokens), COUNT(*)").
        Row().
        Scan(&totalRevenue, &totalTokens, &recordCount)

    return map[string]interface{}{
        "month":           month,
        "total_revenue":   totalRevenue,
        "total_tokens":    totalTokens,
        "record_count":    recordCount,
        "avg_cost_per_record": totalRevenue / float32(recordCount),
    }, nil
}
```

---

## ✅ 检查清单

- [x] 数据模型设计完整
- [x] 数据库表结构完善
- [x] Token 级别计费实现
- [x] 优惠券系统实现
- [x] 自动发票生成
- [x] 配额告警系统
- [x] 自动充值框架
- [x] API 端点完整
- [x] 文档完善

---

## 📚 参考资源

- Stripe 文档: https://stripe.com/docs/api
- PayPal 文档: https://developer.paypal.com/
- GORM 文档: https://gorm.io/

---

**文档版本**: v1.0  
**最后更新**: 2024 年 11 月 21 日  
**作者**: Oblivious 开发团队

