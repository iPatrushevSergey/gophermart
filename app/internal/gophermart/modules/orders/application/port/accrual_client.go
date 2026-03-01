package port

import (
	"context"

	ordersdto "gophermart/internal/gophermart/modules/orders/application/dto"
)

// AccrualClient communicates with the external accrual calculation system.
type AccrualClient interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (*ordersdto.AccrualOrderInfo, error)
}
