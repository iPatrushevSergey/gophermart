package service

import (
	"testing"
	"time"

	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
)

func TestWithdrawalService_Create(t *testing.T) {
	svc := WithdrawalService{}
	now := time.Now()

	w := svc.Create(vo.UserID(1), vo.OrderNumber("12345678903"), vo.Points(200), now)

	assert.Equal(t, vo.UserID(1), w.UserID)
	assert.Equal(t, vo.OrderNumber("12345678903"), w.OrderNumber)
	assert.Equal(t, vo.Points(200), w.Amount)
	assert.Equal(t, now, w.ProcessedAt)
}
