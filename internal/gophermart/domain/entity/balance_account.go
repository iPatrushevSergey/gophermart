package entity

import (
	"errors"
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

// BalanceAccount — user loyalty account.
type BalanceAccount struct {
	UserID         vo.UserID
	Current        vo.Points
	WithdrawnTotal vo.Points
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int64 // for optimistic locking in the database
}

// AddAccrual adds points to the account. now is the time of the operation.
func (a *BalanceAccount) AddAccrual(amount vo.Points, now time.Time) {
	if amount <= 0 {
		return
	}
	a.Current += amount
	a.UpdatedAt = now
}

// Withdraw deducts points. Returns ErrInsufficientBalance if there are not enough funds.
func (a *BalanceAccount) Withdraw(amount vo.Points, now time.Time) error {
	if amount <= 0 {
		return nil
	}
	if a.Current < amount {
		return ErrInsufficientBalance
	}
	a.Current -= amount
	a.WithdrawnTotal += amount
	a.UpdatedAt = now
	return nil
}
