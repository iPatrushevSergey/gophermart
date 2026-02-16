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
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
	voMocks "gophermart/internal/gophermart/domain/vo/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUploadOrder_Execute(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("new order accepted", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := mocks.NewMockOrderReader(ctrl)
		orderWriter := mocks.NewMockOrderWriter(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)
		clk := mocks.NewMockClock(ctrl)

		validator.EXPECT().Valid("12345678903").Return(true)
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(nil, application.ErrNotFound)
		clk.EXPECT().Now().Return(fixedTime)
		orderWriter.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, o *entity.Order) error {
				assert.Equal(t, fixedTime, o.UploadedAt)
				return nil
			},
		)

		uc := NewUploadOrder(orderReader, orderWriter, validator, clk, service.OrderService{})
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.NoError(t, err)
	})

	t.Run("already uploaded by same user", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := mocks.NewMockOrderReader(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)

		validator.EXPECT().Valid("12345678903").Return(true)
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(
			&entity.Order{UserID: vo.UserID(1)}, nil,
		)

		uc := NewUploadOrder(orderReader, nil, validator, nil, service.OrderService{})
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.ErrorIs(t, err, application.ErrAlreadyExists)
	})

	t.Run("conflict with another user", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := mocks.NewMockOrderReader(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)

		validator.EXPECT().Valid("12345678903").Return(true)
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(
			&entity.Order{UserID: vo.UserID(2)}, nil,
		)

		uc := NewUploadOrder(orderReader, nil, validator, nil, service.OrderService{})
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.ErrorIs(t, err, application.ErrConflict)
	})

	t.Run("invalid order number", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		validator := voMocks.NewMockOrderNumberValidator(ctrl)
		validator.EXPECT().Valid("123").Return(false)

		uc := NewUploadOrder(nil, nil, validator, nil, service.OrderService{})
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "123"})

		assert.ErrorIs(t, err, application.ErrInvalidOrderNumber)
	})

	t.Run("repo error on find", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		orderReader := mocks.NewMockOrderReader(ctrl)
		validator := voMocks.NewMockOrderNumberValidator(ctrl)

		validator.EXPECT().Valid("12345678903").Return(true)
		orderReader.EXPECT().FindByNumber(ctx, vo.OrderNumber("12345678903")).Return(nil, errors.New("db error"))

		uc := NewUploadOrder(orderReader, nil, validator, nil, service.OrderService{})
		_, err := uc.Execute(ctx, dto.UploadOrderInput{UserID: 1, OrderNumber: "12345678903"})

		assert.Error(t, err)
	})
}
