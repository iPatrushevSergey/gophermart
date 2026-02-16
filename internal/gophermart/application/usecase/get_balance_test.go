package usecase

import (
	"context"
	"errors"
	"testing"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetBalance_Execute(t *testing.T) {
	ctx := context.Background()
	userID := vo.UserID(1)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		balanceReader := mocks.NewMockBalanceAccountReader(ctrl)

		balanceReader.EXPECT().FindByUserID(ctx, userID).Return(&entity.BalanceAccount{
			Current:        vo.Points(500),
			WithdrawnTotal: vo.Points(200),
		}, nil)

		uc := NewGetBalance(balanceReader)
		result, err := uc.Execute(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, float64(500), result.Current)
		assert.Equal(t, float64(200), result.Withdrawn)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		balanceReader := mocks.NewMockBalanceAccountReader(ctrl)

		balanceReader.EXPECT().FindByUserID(ctx, userID).Return(nil, application.ErrNotFound)

		uc := NewGetBalance(balanceReader)
		_, err := uc.Execute(ctx, userID)

		assert.ErrorIs(t, err, application.ErrNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		balanceReader := mocks.NewMockBalanceAccountReader(ctrl)

		balanceReader.EXPECT().FindByUserID(ctx, userID).Return(nil, errors.New("db error"))

		uc := NewGetBalance(balanceReader)
		_, err := uc.Execute(ctx, userID)

		assert.Error(t, err)
	})
}
