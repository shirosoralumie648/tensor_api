package billing

import (
	"context"
)

type BillingEngine struct {
	pricingManager *PricingManager
	accounting     *BillingEventQueue
	alertManager   *AlertManager
	quotaManager   *QuotaManager
}

func NewBillingEngine() *BillingEngine {
	qm := NewQuotaManager()
	return &BillingEngine{
		pricingManager: NewPricingManager(),
		accounting:     NewBillingEventQueue("billing", 1000),
		quotaManager:   qm,
		alertManager:   NewAlertManager(qm),
	}
}

func (e *BillingEngine) CalculateCost(ctx context.Context, modelName string, inputTokens, outputTokens int) (float64, error) {
	return e.pricingManager.CalculatePrice(modelName, int64(inputTokens), int64(outputTokens))
}

func (e *BillingEngine) RecordBilling(ctx context.Context, event *BillingEvent) error {
	return e.accounting.Enqueue(event)
}

func (e *BillingEngine) CheckQuotaAlert(ctx context.Context, userID string) error {
	return e.alertManager.CheckQuotaUsage(userID)
}

func (e *BillingEngine) GetPricingManager() *PricingManager {
	return e.pricingManager
}

func (e *BillingEngine) GetQuotaManager() *QuotaManager {
	return e.quotaManager
}

func (e *BillingEngine) GetAlertManager() *AlertManager {
	return e.alertManager
}

func (e *BillingEngine) GetAccountingQueue() *BillingEventQueue {
	return e.accounting
}

func (e *BillingEngine) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"accounting_stats": e.accounting.GetStatistics(),
		"pricing_stats":    e.pricingManager.GetStatistics(),
		"queue_size":       e.accounting.Size(),
	}
}
