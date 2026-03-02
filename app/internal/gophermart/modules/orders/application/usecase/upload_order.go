package usecase

import (
	"context"

	"gophermart/internal/gophermart/application"
	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/orders/application/dto"
	"gophermart/internal/gophermart/modules/orders/application/port"
	"gophermart/internal/gophermart/modules/orders/domain/entity"
	"gophermart/internal/gophermart/modules/orders/domain/vo"
)

// UploadOrder handles uploading a new order number for accrual calculation.
type UploadOrder struct {
	orderReader port.OrderReader
	orderWriter port.OrderWriter
	validator   vo.OrderNumberValidator
	clock       appport.Clock
}

// NewUploadOrder returns the upload order use case.
func NewUploadOrder(
	orderReader port.OrderReader,
	orderWriter port.OrderWriter,
	validator vo.OrderNumberValidator,
	clock appport.Clock,
) appport.UseCase[dto.UploadOrderInput, struct{}] {
	return &UploadOrder{orderReader: orderReader, orderWriter: orderWriter, validator: validator, clock: clock}
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

	order := entity.NewOrder(orderNumber, in.UserID, uc.clock.Now())

	if err := uc.orderWriter.Create(ctx, order); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
