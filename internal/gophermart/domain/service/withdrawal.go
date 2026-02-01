package service

import (
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// WithdrawalService performs withdrawal-related domain operations.
type WithdrawalService struct{}

// Create builds a new Withdrawal entity.
func (WithdrawalService) Create(userID vo.UserID, orderNumber vo.OrderNumber, amount vo.Points, at time.Time) entity.Withdrawal {
	return entity.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Amount:      amount,
		ProcessedAt: at,
	}
}
