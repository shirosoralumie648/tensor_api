package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oblivious/backend/internal/model"
	"github.com/oblivious/backend/internal/service"
)

// BillingHandler 计费处理器
type BillingHandler struct {
	billingService *service.AdvancedBillingService
}

// NewBillingHandler 创建计费处理器
func NewBillingHandler(billingService *service.AdvancedBillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// GetBillingStats 获取计费统计信息
// @Summary 获取计费统计
// @Description 获取用户的计费统计信息
// @Tags Billing
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /v1/billing/stats [get]
func (h *BillingHandler) GetBillingStats(c *gin.Context) {
	userID, err := ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.billingService.GetUserBillingStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetInvoices 获取发票列表
// @Summary 获取发票列表
// @Description 获取用户的所有发票
// @Tags Billing
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /v1/billing/invoices [get]
func (h *BillingHandler) GetInvoices(c *gin.Context) {
	userID, err := ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if _, err := sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if _, err := sscanf(ps, "%d", &pageSize); err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}
	}

	// TODO: 从数据库查询发票列表
	// 这里为示例代码

	invoices := []map[string]interface{}{
		{
			"id":              "inv-001",
			"invoice_number":  "INV-user-2024-11",
			"billing_month":   "2024-11",
			"status":          "paid",
			"amount":          150.50,
			"issued_at":       time.Now().AddDate(0, 0, -5),
			"due_date":        time.Now().AddDate(0, 0, 25),
			"paid_at":         time.Now(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"total":    len(invoices),
	})
}

// ApplyCoupon 应用优惠券
// @Summary 应用优惠券
// @Description 为用户账户应用优惠券
// @Tags Billing
// @Accept json
// @Param request body ApplyCouponRequest true "优惠券代码"
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /v1/billing/coupons/apply [post]
func (h *BillingHandler) ApplyCoupon(c *gin.Context) {
	userID, err := ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		CouponCode string `json:"coupon_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, err := h.billingService.ApplyCoupon(c.Request.Context(), userID, req.CouponCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "coupon applied successfully",
		"coupon":   coupon,
		"discount": calculateCouponDiscount(coupon),
	})
}

// CreateInvoice 生成发票
// @Summary 生成发票
// @Description 为指定月份生成发票
// @Tags Billing
// @Accept json
// @Param request body CreateInvoiceRequest true "请求体"
// @Produce json
// @Success 200 {object} model.Invoice
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /v1/billing/invoices/generate [post]
func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	userID, err := ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		BillingMonth string `json:"billing_month" binding:"required"` // YYYY-MM
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.billingService.CreateInvoice(c.Request.Context(), userID, req.BillingMonth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// GetQuotaWarning 检查配额警告
// @Summary 检查配额警告
// @Description 检查用户是否需要配额警告
// @Tags Billing
// @Produce json
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /v1/billing/quota-warning [get]
func (h *BillingHandler) GetQuotaWarning(c *gin.Context) {
	userID, err := ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	warning, err := h.billingService.CheckQuotaWarning(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quota_warning": warning,
		"message": func() string {
			if warning {
				return "Your quota is running low. Please consider upgrading your plan."
			}
			return "Your quota is sufficient."
		}(),
	})
}

// GetSubscriptionPlans 获取订阅计划列表
// @Summary 获取订阅计划
// @Description 获取所有可用的订阅计划
// @Tags Billing
// @Produce json
// @Success 200 {object} []model.SubscriptionPlan
// @Router /v1/billing/plans [get]
func (h *BillingHandler) GetSubscriptionPlans(c *gin.Context) {
	plans := []map[string]interface{}{
		{
			"id":                "plan-basic",
			"name":              "Basic",
			"description":       "Perfect for getting started",
			"monthly_quota":     100000,
			"monthly_price":     9.99,
			"extra_token_price": 0.00002,
			"support_level":     "email",
			"features": []string{
				"Up to 100K tokens per month",
				"Email support",
				"Basic API access",
			},
		},
		{
			"id":                "plan-pro",
			"name":              "Pro",
			"description":       "For professionals",
			"monthly_quota":     1000000,
			"monthly_price":     99.99,
			"extra_token_price": 0.00015,
			"support_level":     "priority",
			"features": []string{
				"Up to 1M tokens per month",
				"Priority email support",
				"Advanced API access",
				"10% discount on overage",
			},
		},
		{
			"id":                "plan-enterprise",
			"name":              "Enterprise",
			"description":       "For large organizations",
			"monthly_quota":     10000000,
			"monthly_price":     999.99,
			"extra_token_price": 0.0001,
			"support_level":     "dedicated",
			"features": []string{
				"Up to 10M tokens per month",
				"24/7 phone support",
				"Dedicated account manager",
				"20% discount on overage",
				"Custom integrations",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// UpdateBillingSettings 更新计费设置
// @Summary 更新计费设置
// @Description 更新用户的计费设置
// @Tags Billing
// @Accept json
// @Param request body UpdateBillingSettingsRequest true "设置信息"
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /v1/billing/settings [put]
func (h *BillingHandler) UpdateBillingSettings(c *gin.Context) {
	userID, err := ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		BillingEmail       string `json:"billing_email"`
		AutoTopup          bool   `json:"auto_topup"`
		AutoTopupThreshold int64  `json:"auto_topup_threshold"`
		AutoTopupAmount    float32 `json:"auto_topup_amount"`
		EnableAlerts       bool   `json:"enable_alerts"`
		AlertThreshold     int64  `json:"alert_threshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: 更新数据库中的设置
	// settings := &model.BillingSettings{
	//     UserID: userID,
	//     ...
	// }

	c.JSON(http.StatusOK, gin.H{
		"message": "billing settings updated successfully",
	})
}

// 辅助函数

func ExtractUserID(c *gin.Context) (string, error) {
	userID, ok := c.Get("user_id")
	if !ok {
		return "", ErrUnauthorized
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", ErrUnauthorized
	}

	return userIDStr, nil
}

var ErrUnauthorized = &RequestError{
	Code:    401,
	Message: "unauthorized",
}

type RequestError struct {
	Code    int
	Message string
}

func (e *RequestError) Error() string {
	return e.Message
}

func calculateCouponDiscount(coupon *model.Coupon) map[string]interface{} {
	discount := map[string]interface{}{
		"type": coupon.Type,
	}

	switch coupon.Type {
	case "percentage":
		discount["value"] = coupon.Value
		discount["description"] = "Save " + string(rune(int(coupon.Value))) + "%"
	case "fixed":
		discount["value"] = coupon.Value
		discount["description"] = "Save $" + string(rune(int(coupon.Value)))
	}

	return discount
}
