package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gophermart/internal/gophermart/modules/balance/domain/entity"
	"gophermart/internal/gophermart/modules/balance/domain/vo"
)

func TestBalanceService_CreateAccount(t *testing.T) {
	svc := BalanceService{}
	now := time.Now()

	acc := svc.CreateAccount(42, now)

	assert.Equal(t, vo.UserID(42), acc.UserID)
	assert.Equal(t, vo.Points(0), acc.Current)
	assert.Equal(t, vo.Points(0), acc.WithdrawnTotal)
	assert.Equal(t, now, acc.CreatedAt)
}

func TestBalanceService_Withdraw(t *testing.T) {
	svc := BalanceService{}
	now := time.Now()

	acc := &entity.BalanceAccount{
		UserID:         1,
		Current:        vo.Points(100),
		WithdrawnTotal: vo.Points(0),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := svc.Withdraw(acc, vo.Points(40), now.Add(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, vo.Points(60), acc.Current)
	assert.Equal(t, vo.Points(40), acc.WithdrawnTotal)
}
