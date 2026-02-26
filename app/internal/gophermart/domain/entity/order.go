package entity

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// OrderStatus is the order calculation status in the loyalty system.
const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrderStatus string

// Order is the order uploaded by the user for calculating points.
type Order struct {
	Number      vo.OrderNumber
	UserID      vo.UserID
	Status      OrderStatus
	Accrual     *vo.Points // nil (no accrual yet or INVALID)
	UploadedAt  time.Time
	ProcessedAt *time.Time
}

// NewOrder creates a new Order entity with NEW status.
func NewOrder(number vo.OrderNumber, userID vo.UserID, now time.Time) *Order {
	return &Order{
		Number:     number,
		UserID:     userID,
		Status:     OrderStatusNew,
		UploadedAt: now,
	}
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

// MarkProcessing sets the order status to processing.
func (o *Order) MarkProcessing() {
	o.Status = OrderStatusProcessing
}
