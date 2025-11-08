package channel

import (
	"chat/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/stripe/stripe-go/v76/webhook"
	"net/http"
	"strings"
)

type SyncChargeForm struct {
	Overwrite bool           `json:"overwrite"`
	Data      ChargeSequence `json:"data"`
}

// ===================== Payment Config =====================

func GetPaymentConfig(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": true,
        "data":   PaymentInstance,
    })
}

func UpdatePaymentConfig(c *gin.Context) {
    var cfg PaymentConfig
    if err := c.ShouldBindJSON(&cfg); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "status": false,
            "error":  err.Error(),
        })
        return
    }

    state := PaymentInstance.Update(&cfg)
    c.JSON(http.StatusOK, gin.H{
        "status": state == nil,
        "error":  utils.GetError(state),
    })
}

// ===================== Payment Orders =====================

func ListPaymentOrders(c *gin.Context) {
    query := PaymentQuery{
        Status:  c.Query("status"),
        Gateway: c.Query("gateway"),
        Q:       c.Query("q"),
        UserId:  utils.ParseInt(c.Query("user_id")),
        Start:   c.Query("start"),
        End:     c.Query("end"),
        Page:    utils.ParseInt(c.Query("page")),
        Size:    utils.ParseInt(c.Query("size")),
    }

    list, total, err := QueryPaymentOrders(query)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": true, "data": gin.H{
        "list":  list,
        "total": total,
    }})
}

func GetPaymentOrder(c *gin.Context) {
    orderNo := c.Param("orderNo")
    item, err := GetOrderByNo(orderNo)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": true, "data": item})
}

func SyncPaymentOrder(c *gin.Context) {
    orderNo := c.Param("orderNo")
    item, err := GetOrderByNo(orderNo)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
        return
    }

    switch strings.ToLower(item.Gateway) {
    case "stripe":
        if !PaymentInstance.Stripe.Enabled || strings.TrimSpace(PaymentInstance.Stripe.SecretKey) == "" {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": "stripe not configured"})
            return
        }
        stripe.Key = PaymentInstance.Stripe.SecretKey
        if strings.TrimSpace(item.TradeNo) == "" {
            _ = MarkOrderSynced(orderNo)
            c.JSON(http.StatusOK, gin.H{"status": true})
            return
        }
        s, err := session.Get(item.TradeNo, nil)
        if err != nil {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
            return
        }
        // update status based on session
        if s.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
            _ = UpdateOrderStatus(orderNo, "paid", s.ID)
        } else if s.Status == stripe.CheckoutSessionStatusExpired {
            _ = UpdateOrderStatus(orderNo, "closed", s.ID)
        }
        _ = MarkOrderSynced(orderNo)
        c.JSON(http.StatusOK, gin.H{"status": true})
        return
    default:
        err := MarkOrderSynced(orderNo)
        c.JSON(http.StatusOK, gin.H{"status": err == nil, "error": utils.GetError(err)})
        return
    }
}

func RefundPaymentOrder(c *gin.Context) {
    orderNo := c.Param("orderNo")
    item, err := GetOrderByNo(orderNo)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
        return
    }

    switch strings.ToLower(item.Gateway) {
    case "stripe":
        if !PaymentInstance.Stripe.Enabled || strings.TrimSpace(PaymentInstance.Stripe.SecretKey) == "" {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": "stripe not configured"})
            return
        }
        stripe.Key = PaymentInstance.Stripe.SecretKey
        if strings.TrimSpace(item.TradeNo) == "" {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": "missing trade_no for refund"})
            return
        }
        // get session to retrieve payment_intent
        s, err := session.Get(item.TradeNo, nil)
        if err != nil {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
            return
        }
        if s.PaymentIntent == nil || s.PaymentIntent.ID == "" {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": "missing payment_intent"})
            return
        }
        _, err = refund.New(&stripe.RefundParams{PaymentIntent: stripe.String(s.PaymentIntent.ID)})
        if err != nil {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
            return
        }
        _ = UpdateOrderStatus(orderNo, "refunded", s.ID)
        c.JSON(http.StatusOK, gin.H{"status": true})
        return
    default:
        c.JSON(http.StatusOK, gin.H{"status": false, "error": "refund not implemented"})
        return
    }
}

// ===================== Payment Notify (stubs) =====================

type CreateOrderForm struct {
    Gateway  string                 `json:"gateway"`
    Amount   float64                `json:"amount"`
    Currency string                 `json:"currency"`
    Subject  string                 `json:"subject"`
    Body     string                 `json:"body"`
    Metadata map[string]interface{} `json:"metadata"`
}

func CreatePaymentOrder(c *gin.Context) {
    var form CreateOrderForm
    if err := c.ShouldBindJSON(&form); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
        return
    }
    gw := strings.ToLower(strings.TrimSpace(form.Gateway))
    orderNo := "po_" + strings.ReplaceAll(uuid.NewString(), "-", "")
    metaBytes, _ := json.Marshal(form.Metadata)

    o := &PaymentOrder{
        OrderNo:  orderNo,
        TradeNo:  "",
        UserId:   0,
        Gateway:  gw,
        Amount:   form.Amount,
        Currency: form.Currency,
        Subject:  form.Subject,
        Body:     form.Body,
        Status:   "pending",
        Metadata: string(metaBytes),
    }

    if err := NewOrder(o); err != nil {
        c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
        return
    }

    switch gw {
    case "stripe":
        if !PaymentInstance.Stripe.Enabled || strings.TrimSpace(PaymentInstance.Stripe.SecretKey) == "" {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": "stripe not configured"})
            return
        }
        stripe.Key = PaymentInstance.Stripe.SecretKey
        unit := int64(form.Amount * 100)
        params := &stripe.CheckoutSessionParams{
            SuccessURL:        stripe.String(PaymentInstance.Stripe.SuccessURL + "?order_no=" + orderNo),
            CancelURL:         stripe.String(PaymentInstance.Stripe.CancelURL + "?order_no=" + orderNo),
            Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
            ClientReferenceID: stripe.String(orderNo),
            LineItems: []*stripe.CheckoutSessionLineItemParams{
                {
                    PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                        Currency:   stripe.String(strings.ToLower(form.Currency)),
                        ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                            Name: stripe.String(form.Subject),
                        },
                        UnitAmount: stripe.Int64(unit),
                    },
                    Quantity: stripe.Int64(1),
                },
            },
            PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
                Metadata: map[string]string{"order_no": orderNo},
            },
        }
        s, err := session.New(params)
        if err != nil {
            c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
            return
        }
        _ = UpdateOrderStatus(orderNo, "pending", s.ID)
        c.JSON(http.StatusOK, gin.H{"status": true, "data": gin.H{"order_no": orderNo, "url": s.URL, "id": s.ID}})
        return
    default:
        c.JSON(http.StatusOK, gin.H{"status": false, "error": "unsupported gateway"})
        return
    }
}

func YiPayNotify(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": false, "error": "not implemented"})
}

func WechatNotify(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": false, "error": "not implemented"})
}

func AlipayNotify(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": false, "error": "not implemented"})
}

func StripeNotify(c *gin.Context) {
    payload, err := c.GetRawData()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
        return
    }
    sig := c.GetHeader("Stripe-Signature")
    secret := strings.TrimSpace(PaymentInstance.Stripe.WebhookSecret)
    if secret == "" {
        c.JSON(http.StatusOK, gin.H{"status": false, "error": "webhook not configured"})
        return
    }
    evt, err := webhook.ConstructEvent(payload, sig, secret)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": false, "error": err.Error()})
        return
    }
    if evt.Type == "checkout.session.completed" {
        var obj struct {
            ID                 string `json:"id"`
            ClientReferenceID  string `json:"client_reference_id"`
            PaymentIntent      struct{ ID string `json:"id"` } `json:"payment_intent"`
        }
        _ = json.Unmarshal(evt.Data.Raw, &obj)
        if strings.TrimSpace(obj.ClientReferenceID) != "" {
            _ = UpdateOrderStatus(obj.ClientReferenceID, "paid", obj.ID)
        }
    }
    c.JSON(http.StatusOK, gin.H{"status": true})
}

func GetInfo(c *gin.Context) {
	c.JSON(http.StatusOK, SystemInstance.AsInfo())
}

func GetFeatureInfo(c *gin.Context) {
	c.JSON(http.StatusOK, FeatureInstance)
}

func AttachmentService(c *gin.Context) {
	// /attachments/:hash -> ~/storage/attachments/:hash
	hash := c.Param("hash")
	c.File(fmt.Sprintf("storage/attachments/%s", hash))
}

func DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.DeleteChannel(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func ActivateChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.ActivateChannel(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func DeactivateChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.DeactivateChannel(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetChannelList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   ConduitInstance.Sequence,
	})
}

func GetChannel(c *gin.Context) {
	id := c.Param("id")
	channel := ConduitInstance.Sequence.GetChannelById(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": channel != nil,
		"data":   channel,
	})
}

func CreateChannel(c *gin.Context) {
	var channel Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := ConduitInstance.CreateChannel(&channel)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func UpdateChannel(c *gin.Context) {
	var channel Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	id := c.Param("id")
	channel.Id = utils.ParseInt(id)

	state := ConduitInstance.UpdateChannel(channel.Id, &channel)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func SetCharge(c *gin.Context) {
	var charge Charge
	if err := c.ShouldBindJSON(&charge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := ChargeInstance.SetRule(charge)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetChargeList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   ChargeInstance.ListRules(),
	})
}

func DeleteCharge(c *gin.Context) {
	id := c.Param("id")
	state := ChargeInstance.DeleteRule(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func SyncCharge(c *gin.Context) {
	var form SyncChargeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
	}

	state := ChargeInstance.SyncRules(form.Data, form.Overwrite)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   SystemInstance,
	})
}

func GetOAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   OAuthInstance,
	})
}

func UpdateConfig(c *gin.Context) {
	var config SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := SystemInstance.UpdateConfig(&config)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func UpdateOAuthConfig(c *gin.Context) {
	var cfg OAuthConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := OAuthInstance.Update(&cfg)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetFeatureConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   FeatureInstance,
	})
}

func UpdateFeatureConfig(c *gin.Context) {
	var cfg FeatureConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := FeatureInstance.Update(&cfg)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetPlanConfig(c *gin.Context) {
	c.JSON(http.StatusOK, PlanInstance)
}

func UpdatePlanConfig(c *gin.Context) {
	var config PlanManager
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := PlanInstance.UpdateConfig(&config)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}
