package service

import (
	"testing"
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
)

func TestOrderService_CreateOrder(t *testing.T) {
	svc := OrderService{}
	now := time.Now()

	order := svc.CreateOrder(vo.OrderNumber("12345678903"), vo.UserID(1), now)

	assert.Equal(t, vo.OrderNumber("12345678903"), order.Number)
	assert.Equal(t, vo.UserID(1), order.UserID)
	assert.Equal(t, entity.OrderStatusNew, order.Status)
	assert.Equal(t, now, order.UploadedAt)
	assert.Nil(t, order.Accrual)
	assert.Nil(t, order.ProcessedAt)
}
