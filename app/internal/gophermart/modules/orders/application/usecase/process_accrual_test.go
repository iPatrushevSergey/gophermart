package usecase

import (
	"context"
	"errors"
	"iter"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/port"
	appmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/modules/orders/application/dto"
	ordersportmocks "gophermart/internal/gophermart/modules/orders/application/port/mocks"
	"gophermart/internal/gophermart/modules/orders/domain/entity"
	"gophermart/internal/gophermart/modules/orders/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var fixedTime = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

type stubBalanceGateway struct {
	apply func(ctx context.Context, userID vo.UserID, amount vo.Points, processedAt time.Time) error
}

func (s *stubBalanceGateway) ApplyAccrual(
	ctx context.Context,
	userID vo.UserID,
	amount vo.Points,
	processedAt time.Time,
) error {
	if s.apply != nil {
		return s.apply(ctx, userID, amount, processedAt)
	}
	return nil
}

// ordersIter builds an iter.Seq2 from a slice — handy for mocking StreamByStatuses.
func ordersIter(orders ...entity.Order) iter.Seq2[entity.Order, error] {
	return func(yield func(entity.Order, error) bool) {
		for _, o := range orders {
			if !yield(o, nil) {
				return
			}
		}
	}
}

func newTestProcessAccrual(
	ctrl *gomock.Controller,
) (
	*ordersportmocks.MockOrderReader,
	*ordersportmocks.MockOrderWriter,
	*stubBalanceGateway,
	*ordersportmocks.MockAccrualClient,
	*appmocks.MockTransactor,
	*appmocks.MockClock,
	*appmocks.MockLogger,
	port.BackgroundRunner,
) {
	orderReader := ordersportmocks.NewMockOrderReader(ctrl)
	orderWriter := ordersportmocks.NewMockOrderWriter(ctrl)
	balanceGateway := &stubBalanceGateway{}
	accrualClient := ordersportmocks.NewMockAccrualClient(ctrl)
	transactor := appmocks.NewMockTransactor(ctrl)
	clk := appmocks.NewMockClock(ctrl)
	logger := appmocks.NewMockLogger(ctrl)

	uc := NewProcessAccrual(orderReader, orderWriter, balanceGateway, accrualClient, transactor, clk, logger, 50, 5, 3)
	return orderReader, orderWriter, balanceGateway, accrualClient, transactor, clk, logger, uc
}

func TestProcessAccrual_Run(t *testing.T) {
	t.Run("empty batch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, _, _, _, _, uc := newTestProcessAccrual(ctrl)

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter())

		processed, err := uc.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
	})

	t.Run("order processed with accrual", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, orderWriter, balanceGateway, accrualClient, transactor, clk, _, uc := newTestProcessAccrual(ctrl)

		accrual := float64(700)
		order := entity.Order{Number: "12345678903", UserID: 1, Status: entity.OrderStatusNew}

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter(order))
		accrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "12345678903").Return(&dto.AccrualOrderInfo{
			Status: "PROCESSED", Accrual: &accrual,
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		transactor.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		orderWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		balanceGateway.apply = func(_ context.Context, userID vo.UserID, amount vo.Points, processedAt time.Time) error {
			assert.Equal(t, vo.UserID(1), userID)
			assert.Equal(t, vo.Points(accrual), amount)
			assert.Equal(t, fixedTime, processedAt)
			return nil
		}

		processed, err := uc.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("order invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, orderWriter, _, accrualClient, _, clk, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter(order))
		accrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "12345678903").Return(&dto.AccrualOrderInfo{
			Status: "INVALID",
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		orderWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		processed, err := uc.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("order still processing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, orderWriter, _, accrualClient, _, clk, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter(order))
		accrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "12345678903").Return(&dto.AccrualOrderInfo{
			Status: "PROCESSING",
		}, nil)
		clk.EXPECT().Now().Return(fixedTime)
		orderWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		processed, err := uc.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("order not registered in accrual", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, accrualClient, _, _, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter(order))
		accrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "12345678903").Return(nil, nil)

		processed, err := uc.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
	})

	t.Run("rate limit propagated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, accrualClient, _, _, _, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}
		rlErr := &application.ErrRateLimit{RetryAfter: 60 * time.Second}

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter(order))
		accrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "12345678903").Return(nil, rlErr)

		processed, err := uc.Run(context.Background())

		assert.ErrorAs(t, err, &rlErr)
		assert.Equal(t, 0, processed)
	})

	t.Run("accrual client error logged and skipped", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		orderReader, _, _, accrualClient, _, _, logger, uc := newTestProcessAccrual(ctrl)

		order := entity.Order{Number: "12345678903", Status: entity.OrderStatusNew}

		orderReader.EXPECT().StreamByStatuses(gomock.Any(), gomock.Any(), 50).Return(ordersIter(order))
		accrualClient.EXPECT().GetOrderAccrual(gomock.Any(), "12345678903").Return(nil, errors.New("timeout"))
		logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

		processed, err := uc.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
	})
}
