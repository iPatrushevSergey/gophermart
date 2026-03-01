package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	appmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/modules/balance/application/dto"
	balanceportmocks "gophermart/internal/gophermart/modules/balance/application/port/mocks"
	"gophermart/internal/gophermart/modules/balance/domain/entity"
	"gophermart/internal/gophermart/modules/balance/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type stubOrderNumberValidator struct {
	valid bool
}

func (s stubOrderNumberValidator) Valid(string) bool {
	return s.valid
}

func TestWithdraw_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		balanceReader := balanceportmocks.NewMockBalanceAccountReader(ctrl)
		balanceWriter := balanceportmocks.NewMockBalanceAccountWriter(ctrl)
		withdrawalWriter := balanceportmocks.NewMockWithdrawalWriter(ctrl)
		transactor := appmocks.NewMockTransactor(ctrl)
		validator := stubOrderNumberValidator{valid: true}
		clk := appmocks.NewMockClock(ctrl)

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
		validator := stubOrderNumberValidator{valid: false}

		uc := NewWithdraw(nil, nil, nil, nil, validator, nil, 3)
		_, err := uc.Execute(ctx, dto.WithdrawInput{UserID: 1, OrderNumber: "123", Sum: 100})

		assert.ErrorIs(t, err, application.ErrInvalidOrderNumber)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		balanceReader := balanceportmocks.NewMockBalanceAccountReader(ctrl)
		transactor := appmocks.NewMockTransactor(ctrl)
		validator := stubOrderNumberValidator{valid: true}
		clk := appmocks.NewMockClock(ctrl)

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

		balanceReader := balanceportmocks.NewMockBalanceAccountReader(ctrl)
		transactor := appmocks.NewMockTransactor(ctrl)
		validator := stubOrderNumberValidator{valid: true}

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
