package admin

import (
	"chat/channel"
	"chat/globals"
	"chat/utils"
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

type UserTypeForm struct {
	Normal       int64 `json:"normal"`
	ApiPaid      int64 `json:"api_paid"`
	BasicPlan    int64 `json:"basic_plan"`
	StandardPlan int64 `json:"standard_plan"`
	ProPlan      int64 `json:"pro_plan"`
	Total        int64 `json:"total"`
}

// Group by gateway: net amount per gateway in last N days
func GetRevenueGroupByGateway(db *sql.DB, days int) RevenueGroupForm {
    if days <= 0 {
        days = 30
    }
    dates := getDays(days)
    start := time.Date(dates[0].Year(), dates[0].Month(), dates[0].Day(), 0, 0, 0, 0, dates[0].Location())
    end := time.Date(dates[len(dates)-1].Year(), dates[len(dates)-1].Month(), dates[len(dates)-1].Day(), 0, 0, 0, 0, dates[len(dates)-1].Location()).AddDate(0, 0, 1)

    paid := map[string]float64{}
    refunded := map[string]float64{}

    if rows, err := globals.QueryDb(db, `
        SELECT gateway, COALESCE(SUM(amount), 0) AS total
        FROM payment_order
        WHERE status = 'paid' AND paid_at >= ? AND paid_at < ?
        GROUP BY gateway
    `, start, end); err == nil {
        for rows.Next() {
            var gateway string
            var total float64
            if err := rows.Scan(&gateway, &total); err == nil {
                paid[gateway] += total
            }
        }
        rows.Close()
    }

    if rows, err := globals.QueryDb(db, `
        SELECT gateway, COALESCE(SUM(amount), 0) AS total
        FROM payment_order
        WHERE status = 'refunded' AND refunded_at >= ? AND refunded_at < ?
        GROUP BY gateway
    `, start, end); err == nil {
        for rows.Next() {
            var gateway string
            var total float64
            if err := rows.Scan(&gateway, &total); err == nil {
                refunded[gateway] += total
            }
        }
        rows.Close()
    }

    // union of keys
    items := make([]RevenueGroupItem, 0)
    keys := map[string]struct{}{}
    for k := range paid { keys[k] = struct{}{} }
    for k := range refunded { keys[k] = struct{}{} }
    for k := range keys {
        net := paid[k] - refunded[k]
        if net < 0 {
            net = 0 // avoid negative, optional
        }
        items = append(items, RevenueGroupItem{Name: k, Amount: float32(net)})
    }
    return RevenueGroupForm{Data: items}
}

// Group by plan in metadata (fallback to subject)
func GetRevenueGroupByPlan(db *sql.DB, days int) RevenueGroupForm {
    if days <= 0 {
        days = 30
    }
    dates := getDays(days)
    start := time.Date(dates[0].Year(), dates[0].Month(), dates[0].Day(), 0, 0, 0, 0, dates[0].Location())
    end := time.Date(dates[len(dates)-1].Year(), dates[len(dates)-1].Month(), dates[len(dates)-1].Day(), 0, 0, 0, 0, dates[len(dates)-1].Location()).AddDate(0, 0, 1)

    type row struct { plan string; amount float64 }
    paid := map[string]float64{}
    refunded := map[string]float64{}

    // helper to extract plan
    extract := func(meta string, subject string) string {
        var m map[string]interface{}
        if err := json.Unmarshal([]byte(meta), &m); err == nil && m != nil {
            if v, ok := m["plan"]; ok {
                if s, ok2 := v.(string); ok2 && s != "" { return s }
            }
        }
        if subject != "" { return subject }
        return "unknown"
    }

    if rows, err := globals.QueryDb(db, `
        SELECT metadata, subject, amount FROM payment_order
        WHERE status = 'paid' AND paid_at >= ? AND paid_at < ?
    `, start, end); err == nil {
        for rows.Next() {
            var meta, subject string
            var amount float64
            if err := rows.Scan(&meta, &subject, &amount); err == nil {
                key := extract(meta, subject)
                paid[key] += amount
            }
        }
        rows.Close()
    }

    if rows, err := globals.QueryDb(db, `
        SELECT metadata, subject, amount FROM payment_order
        WHERE status = 'refunded' AND refunded_at >= ? AND refunded_at < ?
    `, start, end); err == nil {
        for rows.Next() {
            var meta, subject string
            var amount float64
            if err := rows.Scan(&meta, &subject, &amount); err == nil {
                key := extract(meta, subject)
                refunded[key] += amount
            }
        }
        rows.Close()
    }

    items := make([]RevenueGroupItem, 0)
    keys := map[string]struct{}{}
    for k := range paid { keys[k] = struct{}{} }
    for k := range refunded { keys[k] = struct{}{} }
    for k := range keys {
        net := paid[k] - refunded[k]
        if net < 0 { net = 0 }
        items = append(items, RevenueGroupItem{Name: k, Amount: float32(net)})
    }
    return RevenueGroupForm{Data: items}
}

func getDates(t []time.Time) []string {
	return utils.Each[time.Time, string](t, func(date time.Time) string {
		return date.Format("1/2")
	})
}

func getFormat(t time.Time) string {
	return t.Format("2006-01-02")
}

func GetSubscriptionUsers(db *sql.DB) int64 {
	var count int64
	err := globals.QueryRowDb(db, `
   		SELECT COUNT(*) FROM subscription WHERE expired_at > NOW()
   	`).Scan(&count)
	if err != nil {
		return 0
	}

	return count
}

func GetBillingToday(cache *redis.Client) float32 {
	return float32(utils.MustInt(cache, getBillingFormat(getDay()))) / 100
}

func GetBillingMonth(cache *redis.Client) float32 {
	return float32(utils.MustInt(cache, getMonthBillingFormat(getMonth()))) / 100
}

func GetModelData(cache *redis.Client) ModelChartForm {
	dates := getDays(7)

	return ModelChartForm{
		Date: getDates(dates),
		Value: utils.EachNotNil[string, ModelData](globals.SupportModels, func(model string) *ModelData {
			data := ModelData{
				Model: model,
				Data: utils.Each[time.Time, int64](dates, func(date time.Time) int64 {
					return utils.MustInt(cache, getModelFormat(getFormat(date), model))
				}),
			}
			if utils.Sum(data.Data) == 0 {
				return nil
			}

			return &data
		}),
	}
}

func GetSortedModelData(cache *redis.Client) ModelChartForm {
	form := GetModelData(cache)
	data := utils.Sort(form.Value, func(a ModelData, b ModelData) bool {
		return utils.Sum(a.Data) > utils.Sum(b.Data)
	})

	form.Value = data

	return form
}

func GetRequestData(cache *redis.Client) RequestChartForm {
	dates := getDays(7)

	return RequestChartForm{
		Date: getDates(dates),
		Value: utils.Each[time.Time, int64](dates, func(date time.Time) int64 {
			return utils.MustInt(cache, getRequestFormat(getFormat(date)))
		}),
	}
}

func GetBillingData(cache *redis.Client) BillingChartForm {
	dates := getDays(30)

	return BillingChartForm{
		Date: getDates(dates),
		Value: utils.Each[time.Time, float32](dates, func(date time.Time) float32 {
			return float32(utils.MustInt(cache, getBillingFormat(getFormat(date)))) / 100.
		}),
	}
}

func GetErrorData(cache *redis.Client) ErrorChartForm {
	dates := getDays(7)

	return ErrorChartForm{
		Date: getDates(dates),
		Value: utils.Each[time.Time, int64](dates, func(date time.Time) int64 {
			return utils.MustInt(cache, getErrorFormat(getFormat(date)))
		}),
	}
}

func GetUserTypeData(db *sql.DB) (UserTypeForm, error) {
	var form UserTypeForm

	// get total users
	if err := globals.QueryRowDb(db, `
		SELECT COUNT(*) FROM auth
	`).Scan(&form.Total); err != nil {
		return form, err
	}

	// get subscription users count (current subscription)
	// level 1: basic plan, level 2: standard plan, level 3: pro plan
	if err := globals.QueryRowDb(db, `
		SELECT
			(SELECT COUNT(*) FROM subscription WHERE level = 1 AND expired_at > NOW()),
			(SELECT COUNT(*) FROM subscription WHERE level = 2 AND expired_at > NOW()),
			(SELECT COUNT(*) FROM subscription WHERE level = 3 AND expired_at > NOW())
	`).Scan(&form.BasicPlan, &form.StandardPlan, &form.ProPlan); err != nil {
		return form, err
	}

	// get normal users count (no subscription in `subscription` table and `quota` + `used` < initial quota in `quota` table)
	initialQuota := channel.SystemInstance.GetInitialQuota()
	if err := globals.QueryRowDb(db, `
		SELECT COUNT(*) FROM auth 
		WHERE id NOT IN (SELECT user_id FROM subscription WHERE total_month > 0)
		AND id IN (SELECT user_id FROM quota WHERE quota + used <= ?)
	`, initialQuota).Scan(&form.Normal); err != nil {
		return form, err
	}

	form.ApiPaid = form.Total - form.Normal - form.BasicPlan - form.StandardPlan - form.ProPlan

	return form, nil
}

// ===== Revenue (from payment_order) =====
// Net revenue = paid amount - refunded amount

func GetRevenueToday(db *sql.DB) float32 {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)

	var paid float64
	var refunded float64

	_ = globals.QueryRowDb(db, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payment_order
		WHERE status = 'paid' AND paid_at >= ? AND paid_at < ?
	`, start, end).Scan(&paid)

	_ = globals.QueryRowDb(db, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payment_order
		WHERE status = 'refunded' AND refunded_at >= ? AND refunded_at < ?
	`, start, end).Scan(&refunded)

	return float32(paid - refunded)
}

func GetRevenueMonth(db *sql.DB) float32 {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)

	var paid float64
	var refunded float64

	_ = globals.QueryRowDb(db, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payment_order
		WHERE status = 'paid' AND paid_at >= ? AND paid_at < ?
	`, start, end).Scan(&paid)

	_ = globals.QueryRowDb(db, `
		SELECT COALESCE(SUM(amount), 0)
		FROM payment_order
		WHERE status = 'refunded' AND refunded_at >= ? AND refunded_at < ?
	`, start, end).Scan(&refunded)

	return float32(paid - refunded)
}

func GetRevenueChart(db *sql.DB, days int) BillingChartForm {
	if days <= 0 {
		days = 30
	}
	dates := getDays(days)
	// [start, end)
	start := time.Date(dates[0].Year(), dates[0].Month(), dates[0].Day(), 0, 0, 0, 0, dates[0].Location())
	end := time.Date(dates[len(dates)-1].Year(), dates[len(dates)-1].Month(), dates[len(dates)-1].Day(), 0, 0, 0, 0, dates[len(dates)-1].Location()).AddDate(0, 0, 1)

	mPaid := map[string]float64{}
	mRefund := map[string]float64{}

	// aggregate paid by day
	if rows, err := globals.QueryDb(db, `
		SELECT paid_at, amount FROM payment_order
		WHERE status = 'paid' AND paid_at >= ? AND paid_at < ?
	`, start, end); err == nil {
		for rows.Next() {
			var paidAt []uint8
			var amount float64
			if err := rows.Scan(&paidAt, &amount); err == nil {
				d := utils.ConvertTime(paidAt).Format("2006-01-02")
				mPaid[d] += amount
			}
		}
		rows.Close()
	}

	// aggregate refunded by day
	if rows, err := globals.QueryDb(db, `
		SELECT refunded_at, amount FROM payment_order
		WHERE status = 'refunded' AND refunded_at >= ? AND refunded_at < ?
	`, start, end); err == nil {
		for rows.Next() {
			var refundedAt []uint8
			var amount float64
			if err := rows.Scan(&refundedAt, &amount); err == nil {
				d := utils.ConvertTime(refundedAt).Format("2006-01-02")
				mRefund[d] += amount
			}
		}
		rows.Close()
	}

	values := make([]float32, len(dates))
	for i, d := range dates {
		key := getFormat(d)
		net := mPaid[key] - mRefund[key]
		values[i] = float32(net)
	}

	return BillingChartForm{
		Date:  getDates(dates),
		Value: values,
	}
}
