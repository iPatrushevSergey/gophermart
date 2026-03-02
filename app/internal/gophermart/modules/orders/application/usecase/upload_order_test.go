package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"gophermart/internal/gophermart/application"
	appmocks "gophermart/internal/gophermart/application/port/mocks"
	"gophermart/internal/gophermart/modules/orders/application/dto"
	ordersportmocks "gophermart/internal/gophermart/modules/orders/application/port/mocks"
	"gophermart/internal/gophermart/modules/orders/domain/entity"
	"gophermart/internal/gophermart/modules/orders/domain/vo"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type stubOrderNumberValidator struct {
	valid bool
}

func (s stubOrderNumberValidator) Valid(string) bool {
	return s.valid
}

func TestUploadOrder_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("new order accepted", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := ordersportmocks.NewMockOrderReader(ctrl)
		orderWriter := ordersportmocks.NewMockOrderWriter(ctrl)
		validator := stubOrderNumberValidator{valid: true}
		clk := appmocks.NewMockClock(ctrl)

		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(nil, application.ErrNotFound)
		clk.EXPECT().Now().Return(fixedTime)
		orderWriter.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, o *entity.Order) error {
				assert.Equal(t, fixedTime, o.UploadedAt)
				return nil
			},
		)

		uc := NewUploadOrder(orderReader, orderWriter, validator, clk)
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.NoError(t, err)
	})

	t.Run("already uploaded by same user", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := ordersportmocks.NewMockOrderReader(ctrl)
		validator := stubOrderNumberValidator{valid: true}
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(
			&entity.Order{UserID: vo.UserID(1)}, nil,
		)

		uc := NewUploadOrder(orderReader, nil, validator, nil)
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.ErrorIs(t, err, application.ErrAlreadyExists)
	})

	t.Run("conflict with another user", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := ordersportmocks.NewMockOrderReader(ctrl)
		validator := stubOrderNumberValidator{valid: true}
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(
			&entity.Order{UserID: vo.UserID(2)}, nil,
		)

		uc := NewUploadOrder(orderReader, nil, validator, nil)
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.ErrorIs(t, err, application.ErrConflict)
	})

	t.Run("invalid order number", func(t *testing.T) {
		validator := stubOrderNumberValidator{valid: false}

		uc := NewUploadOrder(nil, nil, validator, nil)
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "123"})

		assert.ErrorIs(t, err, application.ErrInvalidOrderNumber)
	})

	t.Run("repo error on find", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := ordersportmocks.NewMockOrderReader(ctrl)
		validator := stubOrderNumberValidator{valid: true}
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(nil, errors.New("db error"))

		uc := NewUploadOrder(orderReader, nil, validator, nil)
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.Error(t, err)
	})
}
