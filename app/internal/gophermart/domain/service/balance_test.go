package service

import (
	"testing"
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
)

func TestBalanceService_CreateAccount(t *testing.T) {
	svc := BalanceService{}
	now := time.Now()

	acc := svc.CreateAccount(vo.UserID(1), now)

	assert.Equal(t, vo.UserID(1), acc.UserID)
	assert.Equal(t, vo.Points(0), acc.Current)
	assert.Equal(t, vo.Points(0), acc.WithdrawnTotal)
	assert.Equal(t, now, acc.CreatedAt)
}

func TestBalanceService_Withdraw(t *testing.T) {
	svc := BalanceService{}
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		acc := &entity.BalanceAccount{Current: 500}
		err := svc.Withdraw(acc, 200, now)

		assert.NoError(t, err)
		assert.Equal(t, vo.Points(300), acc.Current)
	})

	t.Run("insufficient", func(t *testing.T) {
		acc := &entity.BalanceAccount{Current: 100}
		err := svc.Withdraw(acc, 500, now)

		assert.ErrorIs(t, err, entity.ErrInsufficientBalance)
	})
}
