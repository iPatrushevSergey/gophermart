package port

import (
	"context"

	"gophermart/internal/gophermart/application/dto"
)

// AccrualClient communicates with the external accrual calculation system.
type AccrualClient interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*dto.AccrualOrderInfo, error)
}
