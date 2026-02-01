package service

import (
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// BalanceService performs loyalty balance operations.
type BalanceService struct{}

// ApplyAccrual adds points to the account.
func (BalanceService) ApplyAccrual(acc *entity.BalanceAccount, amount vo.Points, now time.Time) {
	acc.AddAccrual(amount, now)
}

// Withdraw deducts points.
func (BalanceService) Withdraw(acc *entity.BalanceAccount, amount vo.Points, now time.Time) error {
	return acc.Withdraw(amount, now)
}
