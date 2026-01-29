package entity

import (
	"errors"
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

var ErrInsufficientBalance = errors.New("insufficient balance")

// BalanceAccount — счёт лояльности пользователя.
type BalanceAccount struct {
	UserID         vo.UserID
	Current        vo.Points
	WithdrawnTotal vo.Points
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int64 // для optimistic locking в БД
}

// AddAccrual начисляет баллы на счёт. now — время операции.
func (a *BalanceAccount) AddAccrual(amount vo.Points, now time.Time) {
	if amount <= 0 {
		return
	}
	a.Current += amount
	a.UpdatedAt = now
}

// Withdraw списывает баллы. Возвращает ErrInsufficientBalance, если средств недостаточно.
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
