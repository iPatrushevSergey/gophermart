package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
	voMocks "gophermart/internal/gophermart/domain/vo/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWithdraw_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		balanceReader := mocks.NewMockBalanceAccountReader(ctrl)
		balanceWriter := mocks.NewMockBalanceAccountWriter(ctrl)
		withdrawalWriter := mocks.NewMockWithdrawalWriter(ctrl)
		transactor := mocks.NewMockTransactor(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)
		clk := mocks.NewMockClock(ctrl)

		validator.EXPECT().Valid("2377225624").Return(true)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		balanceReader.EXPECT().FindByUserID(ctx, vo.UserID(1)).Return(&entity.BalanceAccount{
			Current: vo.Points(500),
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		withdrawalWriter.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		balanceWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		uc := NewWithdraw(balanceReader, balanceWriter, withdrawalWriter, transactor, validator, clk, 3)
		_, err := uc.Execute(ctx, dto.WithdrawInput{UserID: 1, OrderNumber: "2377225624", Sum: 200})

		assert.NoError(t, err)
	})

	t.Run("invalid order number", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)

		validator.EXPECT().Valid("123").Return(false)

		uc := NewWithdraw(nil, nil, nil, nil, validator, nil, 3)
		_, err := uc.Execute(ctx, dto.WithdrawInput{UserID: 1, OrderNumber: "123", Sum: 100})

		assert.ErrorIs(t, err, application.ErrInvalidOrderNumber)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		balanceReader := mocks.NewMockBalanceAccountReader(ctrl)
		transactor := mocks.NewMockTransactor(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)
		clk := mocks.NewMockClock(ctrl)

		validator.EXPECT().Valid("2377225624").Return(true)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		balanceReader.EXPECT().FindByUserID(ctx, vo.UserID(1)).Return(&entity.BalanceAccount{
			Current: vo.Points(50),
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)

		uc := NewWithdraw(balanceReader, nil, nil, transactor, validator, clk, 3)
		_, err := uc.Execute(ctx, dto.WithdrawInput{UserID: 1, OrderNumber: "2377225624", Sum: 200})

		assert.ErrorIs(t, err, application.ErrInsufficientBalance)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		balanceReader := mocks.NewMockBalanceAccountReader(ctrl)
		transactor := mocks.NewMockTransactor(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)

		validator.EXPECT().Valid("2377225624").Return(true)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		balanceReader.EXPECT().FindByUserID(ctx, vo.UserID(1)).Return(nil, errors.New("db error"))

		uc := NewWithdraw(balanceReader, nil, nil, transactor, validator, nil, 3)
		_, err := uc.Execute(ctx, dto.WithdrawInput{UserID: 1, OrderNumber: "2377225624", Sum: 200})

		assert.Error(t, err)
	})
}
