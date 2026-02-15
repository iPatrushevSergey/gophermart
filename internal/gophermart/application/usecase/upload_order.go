package usecase

import (
	"context"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
)

// UploadOrder handles uploading a new order number for accrual calculation.
type UploadOrder struct {
	orderReader port.OrderReader
	orderWriter port.OrderWriter
	validator   vo.OrderNumberValidator
	orderSvc    service.OrderService
}

// NewUploadOrder returns the upload order use case.
func NewUploadOrder(
	orderReader port.OrderReader,
	orderWriter port.OrderWriter,
	validator vo.OrderNumberValidator,
	orderSvc service.OrderService,
) port.UseCase[dto.UploadOrderInput, struct{}] {
	return &UploadOrder{orderReader: orderReader, orderWriter: orderWriter, validator: validator, orderSvc: orderSvc}
}

// Execute validates the order number and creates it.
//
// Errors:
//   - application.ErrInvalidOrderNumber — order number failed Luhn check
//   - application.ErrAlreadyExists — same user already uploaded this order
//   - application.ErrConflict — another user uploaded this order
func (uc *UploadOrder) Execute(ctx context.Context, in dto.UploadOrderInput) (struct{}, error) {
	orderNumber, err := vo.NewOrderNumber(uc.validator, in.OrderNumber)
	if err != nil {
		return struct{}{}, application.ErrInvalidOrderNumber
	}

	existing, err := uc.orderReader.FindByNumber(ctx, orderNumber)
	if err != nil && err != application.ErrNotFound {
		return struct{}{}, err
	}

	if existing != nil {
		if existing.UserID == in.UserID {
			return struct{}{}, application.ErrAlreadyExists
		}
		return struct{}{}, application.ErrConflict
	}

	order := uc.orderSvc.CreateOrder(orderNumber, in.UserID, time.Now())

	if err := uc.orderWriter.Create(ctx, order); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
