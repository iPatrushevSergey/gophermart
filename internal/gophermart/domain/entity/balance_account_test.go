package entity

import (
	"testing"
	"time"

	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
)

func TestBalanceAccount_AddAccrual(t *testing.T) {
	now := time.Now()

	t.Run("adds positive amount", func(t *testing.T) {
		acc := &BalanceAccount{Current: 100}
		acc.AddAccrual(50, now)

		assert.Equal(t, vo.Points(150), acc.Current)
		assert.Equal(t, now, acc.UpdatedAt)
	})

	t.Run("ignores zero amount", func(t *testing.T) {
		before := time.Now().Add(-time.Hour)
		acc := &BalanceAccount{Current: 100, UpdatedAt: before}
		acc.AddAccrual(0, now)

		assert.Equal(t, vo.Points(100), acc.Current)
		assert.Equal(t, before, acc.UpdatedAt)
	})

	t.Run("ignores negative amount", func(t *testing.T) {
		acc := &BalanceAccount{Current: 100}
		acc.AddAccrual(-10, now)

		assert.Equal(t, vo.Points(100), acc.Current)
	})
}

func TestBalanceAccount_Withdraw(t *testing.T) {
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		acc := &BalanceAccount{Current: 500}
		err := acc.Withdraw(200, now)

		assert.NoError(t, err)
		assert.Equal(t, vo.Points(300), acc.Current)
		assert.Equal(t, vo.Points(200), acc.WithdrawnTotal)
		assert.Equal(t, now, acc.UpdatedAt)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		acc := &BalanceAccount{Current: 100}
		err := acc.Withdraw(200, now)

		assert.ErrorIs(t, err, ErrInsufficientBalance)
		assert.Equal(t, vo.Points(100), acc.Current)
	})

	t.Run("zero amount is noop", func(t *testing.T) {
		acc := &BalanceAccount{Current: 100}
		err := acc.Withdraw(0, now)

		assert.NoError(t, err)
		assert.Equal(t, vo.Points(100), acc.Current)
	})

	t.Run("exact balance", func(t *testing.T) {
		acc := &BalanceAccount{Current: 100}
		err := acc.Withdraw(100, now)

		assert.NoError(t, err)
		assert.Equal(t, vo.Points(0), acc.Current)
		assert.Equal(t, vo.Points(100), acc.WithdrawnTotal)
	})
}
