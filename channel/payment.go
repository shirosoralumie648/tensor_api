package channel

import (
	"chat/connection"
	"chat/globals"
	"database/sql"
	"fmt"
	"strings"
)

// PaymentOrder represents a payment order persisted in DB
// status: pending | paid | expired | canceled | failed | refunded

type PaymentOrder struct {
	Id        int     `json:"id"`
	OrderNo   string  `json:"order_no"`
	TradeNo   string  `json:"trade_no"`
	UserId    int     `json:"user_id"`
	Gateway   string  `json:"gateway"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Subject   string  `json:"subject"`
	Body      string  `json:"body"`
	Status    string  `json:"status"`
	Metadata  string  `json:"metadata"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	PaidAt    string  `json:"paid_at"`
	RefundedAt string `json:"refunded_at"`
	ExpiresAt string  `json:"expires_at"`
}

type PaymentQuery struct {
	Status  string
	Gateway string
	Q       string
	UserId  int
	Start   string
	End     string
	Page    int
	Size    int
}

func ensurePageSize(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	return page, size
}

func scanOrder(row *sql.Row) (*PaymentOrder, error) {
	var o PaymentOrder
	var tradeNo, currency, subject, body, metadata, paidAt, refundedAt, expiresAt sql.NullString
	if err := row.Scan(
		&o.Id, &o.OrderNo, &tradeNo, &o.UserId, &o.Gateway, &o.Amount,
		&currency, &subject, &body, &o.Status, &metadata, &o.CreatedAt, &o.UpdatedAt,
		&paidAt, &refundedAt, &expiresAt,
	); err != nil {
		return nil, err
	}
	o.TradeNo = tradeNo.String
	o.Currency = currency.String
	o.Subject = subject.String
	o.Body = body.String
	o.Metadata = metadata.String
	o.PaidAt = paidAt.String
	o.RefundedAt = refundedAt.String
	o.ExpiresAt = expiresAt.String
	return &o, nil
}

func scanOrders(rows *sql.Rows) ([]*PaymentOrder, error) {
	defer rows.Close()
	list := make([]*PaymentOrder, 0)
	for rows.Next() {
		var o PaymentOrder
		var tradeNo, currency, subject, body, metadata, paidAt, refundedAt, expiresAt sql.NullString
		if err := rows.Scan(
			&o.Id, &o.OrderNo, &tradeNo, &o.UserId, &o.Gateway, &o.Amount,
			&currency, &subject, &body, &o.Status, &metadata, &o.CreatedAt, &o.UpdatedAt,
			&paidAt, &refundedAt, &expiresAt,
		); err != nil {
			return nil, err
		}
		o.TradeNo = tradeNo.String
		o.Currency = currency.String
		o.Subject = subject.String
		o.Body = body.String
		o.Metadata = metadata.String
		o.PaidAt = paidAt.String
		o.RefundedAt = refundedAt.String
		o.ExpiresAt = expiresAt.String
		list = append(list, &o)
	}
	return list, nil
}

func GetOrderByNo(orderNo string) (*PaymentOrder, error) {
	row := globals.QueryRowDb(connection.DB, `
		SELECT id, order_no, trade_no, user_id, gateway, amount, currency, subject, body, status,
		       metadata, created_at, updated_at, paid_at, refunded_at, expires_at
		FROM payment_order WHERE order_no = ?
	`, orderNo)
	return scanOrder(row)
}

func NewOrder(o *PaymentOrder) error {
	_, err := globals.ExecDb(connection.DB, `
		INSERT INTO payment_order (order_no, trade_no, user_id, gateway, amount, currency, subject, body, status, metadata, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, o.OrderNo, o.TradeNo, o.UserId, strings.ToLower(o.Gateway), o.Amount, o.Currency, o.Subject, o.Body, o.Status, o.Metadata, o.ExpiresAt)
	return err
}

func UpdateOrderStatus(orderNo, status, tradeNo string) error {
	_, err := globals.ExecDb(connection.DB, `
		UPDATE payment_order
		SET status = ?, trade_no = COALESCE(?, trade_no), updated_at = CURRENT_TIMESTAMP,
		    paid_at = CASE WHEN ? = 'paid' THEN CURRENT_TIMESTAMP ELSE paid_at END,
		    refunded_at = CASE WHEN ? = 'refunded' THEN CURRENT_TIMESTAMP ELSE refunded_at END
		WHERE order_no = ?
	`, status, nullIfEmpty(tradeNo), status, status, orderNo)
	return err
}

func MarkOrderSynced(orderNo string) error {
	_, err := globals.ExecDb(connection.DB, `
		UPDATE payment_order SET updated_at = CURRENT_TIMESTAMP WHERE order_no = ?
	`, orderNo)
	return err
}

func QueryPaymentOrders(q PaymentQuery) ([]*PaymentOrder, int, error) {
	page, size := ensurePageSize(q.Page, q.Size)
	offset := (page - 1) * size

	cond := []string{"1=1"}
	args := make([]interface{}, 0)
	if q.Status != "" {
		cond = append(cond, "status = ?")
		args = append(args, q.Status)
	}
	if q.Gateway != "" {
		cond = append(cond, "gateway = ?")
		args = append(args, strings.ToLower(q.Gateway))
	}
	if q.UserId > 0 {
		cond = append(cond, "user_id = ?")
		args = append(args, q.UserId)
	}
	if q.Q != "" {
		cond = append(cond, "(order_no LIKE ? OR trade_no LIKE ? OR subject LIKE ?)")
		kw := fmt.Sprintf("%%%s%%", q.Q)
		args = append(args, kw, kw, kw)
	}
	if q.Start != "" {
		cond = append(cond, "created_at >= ?")
		args = append(args, q.Start)
	}
	if q.End != "" {
		cond = append(cond, "created_at <= ?")
		args = append(args, q.End)
	}

	where := strings.Join(cond, " AND ")

	cntRow := globals.QueryRowDb(connection.DB, "SELECT COUNT(*) FROM payment_order WHERE "+where, args...)
	var total int
	if err := cntRow.Scan(&total); err != nil {
		return nil, 0, err
	}

	argsPage := append(args, size, offset)
	rows, err := globals.QueryDb(connection.DB, `
		SELECT id, order_no, trade_no, user_id, gateway, amount, currency, subject, body, status,
		       metadata, created_at, updated_at, paid_at, refunded_at, expires_at
		FROM payment_order WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?
	`, argsPage...)
	if err != nil {
		return nil, 0, err
	}
	list, err := scanOrders(rows)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
