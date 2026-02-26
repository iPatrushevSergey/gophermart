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

func TestListWithdrawals_Execute(t *testing.T) {
	ctx := context.Background()
	userID := vo.UserID(1)

	t.Run("returns mapped withdrawals", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockWithdrawalReader(ctrl)

		now := time.Now()
		reader.EXPECT().ListByUserID(ctx, userID).Return([]entity.Withdrawal{
			{OrderNumber: "111", Amount: vo.Points(200), ProcessedAt: now},
			{OrderNumber: "222", Amount: vo.Points(300), ProcessedAt: now},
		}, nil)

		uc := NewListWithdrawals(reader)
		result, err := uc.Execute(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "111", result[0].OrderNumber)
		assert.Equal(t, float64(200), result[0].Sum)
		assert.Equal(t, "222", result[1].OrderNumber)
	})

	t.Run("empty list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockWithdrawalReader(ctrl)

		reader.EXPECT().ListByUserID(ctx, userID).Return(nil, nil)

		uc := NewListWithdrawals(reader)
		result, err := uc.Execute(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := mocks.NewMockWithdrawalReader(ctrl)

		reader.EXPECT().ListByUserID(ctx, userID).Return(nil, errors.New("db error"))

		uc := NewListWithdrawals(reader)
		_, err := uc.Execute(ctx, userID)

		assert.Error(t, err)
	})
}
