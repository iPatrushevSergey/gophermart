package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// OrderRepository persists orders.
type OrderRepository interface {
	Create(ctx context.Context, o *entity.Order) error
	FindByNumber(ctx context.Context, number vo.OrderNumber) (*entity.Order, error)
	ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Order, error)
	ListByStatuses(ctx context.Context, statuses []entity.OrderStatus, limit int) ([]entity.Order, error)
	Update(ctx context.Context, o *entity.Order) error
}
