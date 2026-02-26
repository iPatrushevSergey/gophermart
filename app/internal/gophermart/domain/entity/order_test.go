package entity

import (
	"testing"
	"time"

	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
)

func TestOrder_MarkProcessed(t *testing.T) {
	now := time.Now()
	accrual := vo.Points(500)

	o := &Order{Status: OrderStatusNew}
	o.MarkProcessed(accrual, now)

	assert.Equal(t, OrderStatusProcessed, o.Status)
	assert.NotNil(t, o.Accrual)
	assert.Equal(t, accrual, *o.Accrual)
	assert.NotNil(t, o.ProcessedAt)
	assert.Equal(t, now, *o.ProcessedAt)
}

func TestOrder_MarkInvalid(t *testing.T) {
	now := time.Now()
	accrual := vo.Points(100)

	o := &Order{Status: OrderStatusProcessing, Accrual: &accrual}
	o.MarkInvalid(now)

	assert.Equal(t, OrderStatusInvalid, o.Status)
	assert.Nil(t, o.Accrual)
	assert.NotNil(t, o.ProcessedAt)
	assert.Equal(t, now, *o.ProcessedAt)
}

func TestOrder_MarkProcessing(t *testing.T) {
	o := &Order{Status: OrderStatusNew}
	o.MarkProcessing()

	assert.Equal(t, OrderStatusProcessing, o.Status)
}
