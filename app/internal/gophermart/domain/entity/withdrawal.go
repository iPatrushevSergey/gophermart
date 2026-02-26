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

// NewWithdrawal creates a new Withdrawal entity.
func NewWithdrawal(userID vo.UserID, orderNumber vo.OrderNumber, amount vo.Points, at time.Time) *Withdrawal {
	return &Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Amount:      amount,
		ProcessedAt: at,
	}
}
