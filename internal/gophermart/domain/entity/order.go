package entity

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// OrderStatus — статус расчёта заказа в системе лояльности.
// В API/домене — строки; в БД храним SMALLINT (маппинг в адаптере).
const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrderStatus string

// Order — the order uploaded by the user for calculating points.
// Identification by Number (natural key); surrogate id only in the database.
type Order struct {
	Number      vo.OrderNumber
	UserID      vo.UserID
	Status      OrderStatus
	Accrual     *vo.Points // nil, no accrual yet, or INVALID
	UploadedAt  time.Time
	ProcessedAt *time.Time
}

// MarkProcessed transfers the order to the PROCESSED status and records the charge and time.
func (o *Order) MarkProcessed(accrual vo.Points, now time.Time) {
	o.Status = OrderStatusProcessed
	o.Accrual = &accrual
	o.ProcessedAt = &now
}

// MarkInvalid converts the order to the INVALID status.
func (o *Order) MarkInvalid(now time.Time) {
	o.Status = OrderStatusInvalid
	o.Accrual = nil
	o.ProcessedAt = &now
}

// MarkProcessing transfers the order to processing (awaiting a response from accrual).
func (o *Order) MarkProcessing() {
	o.Status = OrderStatusProcessing
}
