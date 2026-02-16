package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	appdto "gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var fixedTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestProcessAccrual(
	ctrl *gomock.Controller,
) (*mocks.MockOrderReader, *mocks.MockOrderWriter, *mocks.MockBalanceAccountReader, *mocks.MockBalanceAccountWriter, *mocks.MockAccrualClient, *mocks.MockTransactor, *mocks.MockClock, *mocks.MockLogger, *ProcessAccrual) {
	orderReader := mocks.NewMockOrderReader(ctrl)
	orderWriter := mocks.NewMockOrderWriter(ctrl)
	balanceReader := mocks.NewMockBalanceAccountReader(ctrl)
	balanceWriter := mocks.NewMockBalanceAccountWriter(ctrl)
	accrualClient := mocks.NewMockAccrualClient(ctrl)
	transactor := mocks.NewMockTransactor(ctrl)
	clk := mocks.NewMockClock(ctrl)
	logger := mocks.NewMockLogger(ctrl)

	uc := NewProcessAccrual(orderReader, orderWriter, balanceReader, balanceWriter, accrualClient, transactor, clk, logger, 50, 3)
	return orderReader, orderWriter, balanceReader, balanceWriter, accrualClient, transactor, clk, logger, uc
}

func TestProcessAccrual_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("empty batch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, _, _, _, _, _, uc := newTestProcessAccrual(ctrl)

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return(nil, nil)

		processed, err := uc.Run(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
	})

	t.Run("order processed with accrual", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, orderWriter, balanceReader, balanceWriter, accrualClient, transactor, clk, _, uc := newTestProcessAccrual(ctrl)

		accrual := float64(700)
		order := entity.Order{Number: "12345678903", UserID: 1, Status: entity.OrderStatusNew}

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return([]entity.Order{order}, nil)
		accrualClient.EXPECT().GetOrderAccrual(ctx, "12345678903").Return(&appdto.AccrualOrderInfo{
			Status: "PROCESSED", Accrual: &accrual,
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		orderWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		balanceReader.EXPECT().FindByUserID(ctx, vo.UserID(1)).Return(&entity.BalanceAccount{
			Current: vo.Points(0),
		}, nil)
		balanceWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		processed, err := uc.Run(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("order invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, orderWriter, _, _, accrualClient, _, clk, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return([]entity.Order{order}, nil)
		accrualClient.EXPECT().GetOrderAccrual(ctx, "12345678903").Return(&appdto.AccrualOrderInfo{
			Status: "INVALID",
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		orderWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		processed, err := uc.Run(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("order still processing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, orderWriter, _, _, accrualClient, _, clk, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return([]entity.Order{order}, nil)
		accrualClient.EXPECT().GetOrderAccrual(ctx, "12345678903").Return(&appdto.AccrualOrderInfo{
			Status: "PROCESSING",
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		orderWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		processed, err := uc.Run(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("order not registered in accrual", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, _, accrualClient, _, _, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return([]entity.Order{order}, nil)
		accrualClient.EXPECT().GetOrderAccrual(ctx, "12345678903").Return(nil, nil)

		processed, err := uc.Run(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("rate limit propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, _, accrualClient, _, _, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}
		rlErr := &application.ErrRateLimit{RetryAfter: 60 * time.Second}

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return([]entity.Order{order}, nil)
		accrualClient.EXPECT().GetOrderAccrual(ctx, "12345678903").Return(nil, rlErr)

		processed, err := uc.Run(ctx)

		assert.ErrorAs(t, err, &rlErr)
		assert.Equal(t, 0, processed)
	})

	t.Run("accrual client error logged and skipped", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, _, accrualClient, _, _, logger, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().ListByStatuses(ctx, gomock.Any(), 50).Return([]entity.Order{order}, nil)
		accrualClient.EXPECT().GetOrderAccrual(ctx, "12345678903").Return(nil, errors.New("timeout"))
		logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

		processed, err := uc.Run(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
	})
}
