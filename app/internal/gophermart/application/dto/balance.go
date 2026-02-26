package dto

import "gophermart/internal/gophermart/domain/vo"

// BalanceOutput is the output for the user's current balance.
type BalanceOutput struct {
	Current   float64
	Withdrawn float64
}

// WithdrawInput is the input for a withdrawal request.
type WithdrawInput struct {
	UserID      vo.UserID
	OrderNumber string
	Sum         float64
}
