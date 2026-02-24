package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListOrders_Execute(t *testing.T) {
	ctx := context.Background()
	userID := vo.UserID(1)

	t.Run("returns mapped orders", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader := mocks.NewMockOrderReader(ctrl)

		accrual := vo.Points(500)
		now := time.Now()
		orderReader.EXPECT().ListByUserID(ctx, userID).Return([]entity.Order{
			{Number: "111", Status: entity.OrderStatusProcessed, Accrual: &accrual, UploadedAt: now},
			{Number: "222", Status: entity.OrderStatusNew, UploadedAt: now},
		}, nil)

		uc := NewListOrders(orderReader)
		result, err := uc.Execute(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "111", result[0].Number)
		assert.Equal(t, "PROCESSED", result[0].Status)
		assert.NotNil(t, result[0].Accrual)
		assert.Equal(t, float64(500), *result[0].Accrual)
		assert.Nil(t, result[1].Accrual)
	})

	t.Run("empty list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader := mocks.NewMockOrderReader(ctrl)

		orderReader.EXPECT().ListByUserID(ctx, userID).Return(nil, nil)

		uc := NewListOrders(orderReader)
		result, err := uc.Execute(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader := mocks.NewMockOrderReader(ctrl)

		orderReader.EXPECT().ListByUserID(ctx, userID).Return(nil, errors.New("db error"))

		uc := NewListOrders(orderReader)
		_, err := uc.Execute(ctx, userID)

		assert.Error(t, err)
	})
}
