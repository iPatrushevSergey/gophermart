package usecase

import (
	"context"

	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
)

// ListOrders returns all orders uploaded by the given user.
type ListOrders struct {
	orderRepo port.OrderRepository
}

// NewListOrders returns the list orders use case.
func NewListOrders(orderRepo port.OrderRepository) port.UseCase[vo.UserID, []dto.OrderOutput] {
	return &ListOrders{orderRepo: orderRepo}
}

// Execute fetches orders and maps them to output DTOs.
// Returns an empty slice if the user has no orders.
func (uc *ListOrders) Execute(ctx context.Context, userID vo.UserID) ([]dto.OrderOutput, error) {
	orders, err := uc.orderRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.OrderOutput, 0, len(orders))
	for _, o := range orders {
		out := dto.OrderOutput{
			Number:     o.Number.String(),
			Status:     string(o.Status),
			UploadedAt: o.UploadedAt,
		}
		if o.Accrual != nil {
			v := float64(*o.Accrual)
			out.Accrual = &v
		}
		result = append(result, out)
	}

	return result, nil
}
