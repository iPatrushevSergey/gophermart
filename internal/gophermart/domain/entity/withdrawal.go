package entity

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// Withdrawal — the fact that points were deducted from the payment for the order.
// Immutable event: created once, does not change.
type Withdrawal struct {
	UserID      vo.UserID
	OrderNumber vo.OrderNumber
	Amount      vo.Points
	ProcessedAt time.Time
}
