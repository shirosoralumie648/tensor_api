package channel

import "github.com/gin-gonic/gin"

func Register(app *gin.RouterGroup) {
    app.GET("/info", GetInfo)
    app.GET("/feature/info", GetFeatureInfo)
    app.GET("/attachments/:hash", AttachmentService)

    // payment notify
    app.POST("/payment/notify/yipay", YiPayNotify)
    app.POST("/payment/notify/wechat", WechatNotify)
    app.POST("/payment/notify/alipay", AlipayNotify)
    app.POST("/payment/notify/stripe", StripeNotify)

    // payment order
    app.POST("/payment/order/create", CreatePaymentOrder)

    app.GET("/admin/channel/list", GetChannelList)
    app.POST("/admin/channel/create", CreateChannel)
    app.GET("/admin/channel/get/:id", GetChannel)
    app.POST("/admin/channel/update/:id", UpdateChannel)
    app.GET("/admin/channel/delete/:id", DeleteChannel)
    app.GET("/admin/channel/activate/:id", ActivateChannel)
    app.GET("/admin/channel/deactivate/:id", DeactivateChannel)

    app.GET("/admin/charge/list", GetChargeList)
    app.POST("/admin/charge/set", SetCharge)
    app.GET("/admin/charge/delete/:id", DeleteCharge)
    app.POST("/admin/charge/sync", SyncCharge)

    app.GET("/admin/config/view", GetConfig)
    app.POST("/admin/config/update", UpdateConfig)

    app.GET("/admin/oauth/view", GetOAuthConfig)
    app.POST("/admin/oauth/update", UpdateOAuthConfig)

    app.GET("/admin/feature/view", GetFeatureConfig)
    app.POST("/admin/feature/update", UpdateFeatureConfig)

    // payment admin
    app.GET("/admin/payment/config/view", GetPaymentConfig)
    app.POST("/admin/payment/config/update", UpdatePaymentConfig)
    app.GET("/admin/payment/orders", ListPaymentOrders)
    app.GET("/admin/payment/order/:orderNo", GetPaymentOrder)
    app.GET("/admin/payment/order/sync/:orderNo", SyncPaymentOrder)
    app.POST("/admin/payment/order/refund/:orderNo", RefundPaymentOrder)

    app.GET("/admin/plan/view", GetPlanConfig)
    app.POST("/admin/plan/update", UpdatePlanConfig)
}
