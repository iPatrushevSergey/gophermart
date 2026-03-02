package port

import (
	"context"
	"iter"

	"gophermart/internal/gophermart/modules/orders/domain/entity"
	"gophermart/internal/gophermart/modules/orders/domain/vo"
)

// OrderReader provides read-only access to orders for orders module.
type OrderReader interface {
	FindByNumber(ctx context.Context, number vo.OrderNumber) (*entity.Order, error)
	ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Order, error)
	ListByStatuses(ctx context.Context, statuses []entity.OrderStatus, limit int) ([]entity.Order, error)
	StreamByStatuses(ctx context.Context, statuses []entity.OrderStatus, limit int) iter.Seq2[entity.Order, error]
}

// OrderWriter provides write access to orders for orders module.
type OrderWriter interface {
	Create(ctx context.Context, o *entity.Order) error
	Update(ctx context.Context, o *entity.Order) error
}

// OrderRepository combines reader and writer for orders DI wiring.
type OrderRepository interface {
	OrderReader
	OrderWriter
}
