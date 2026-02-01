package entity

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// Withdrawal is the record of points deducted for an order.
type Withdrawal struct {
	UserID      vo.UserID
	OrderNumber vo.OrderNumber
	Amount      vo.Points
	ProcessedAt time.Time
}
