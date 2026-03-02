package entity

import (
	"time"

	"gophermart/internal/gophermart/modules/orders/domain/vo"
)

// OrderStatus is accrual processing status.
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

// Order is orders module aggregate root.
type Order struct {
	Number      vo.OrderNumber
	UserID      vo.UserID
	Status      OrderStatus
	Accrual     *vo.Points
	UploadedAt  time.Time
	ProcessedAt *time.Time
}

// NewOrder creates a new order with NEW status.
func NewOrder(number vo.OrderNumber, userID vo.UserID, now time.Time) *Order {
	return &Order{
		Number:     number,
		UserID:     userID,
		Status:     OrderStatusNew,
		UploadedAt: now,
	}
}

// MarkProcessed transitions order to PROCESSED with accrual.
func (o *Order) MarkProcessed(accrual vo.Points, now time.Time) {
	o.Status = OrderStatusProcessed
	o.Accrual = &accrual
	o.ProcessedAt = &now
}

// MarkInvalid transitions order to INVALID.
func (o *Order) MarkInvalid(now time.Time) {
	o.Status = OrderStatusInvalid
	o.Accrual = nil
	o.ProcessedAt = &now
}

// MarkProcessing transitions order to PROCESSING.
func (o *Order) MarkProcessing() {
	o.Status = OrderStatusProcessing
}
