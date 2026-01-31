package service

import (
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// BalanceService is a domain service for loyalty balance operations.
// Used when a use case needs to apply accrual or withdrawal: the service calls
// BalanceAccount entity methods (AddAccrual/Withdraw) without duplicating invariants.
// Implemented as a domain service (rather than entity methods only) so that later,
// if needed, a Clock or limits can be introduced without spreading logic across use cases.
type BalanceService struct{}

// ApplyAccrual adds points to the account.
func (BalanceService) ApplyAccrual(acc *entity.BalanceAccount, amount vo.Points, now time.Time) {
	acc.AddAccrual(amount, now)
}

// Withdraw deducts points. Returns entity.ErrInsufficientBalance when funds are low.
func (BalanceService) Withdraw(acc *entity.BalanceAccount, amount vo.Points, now time.Time) error {
	return acc.Withdraw(amount, now)
}
