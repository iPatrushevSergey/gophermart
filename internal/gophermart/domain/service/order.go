package service

import (
	"time"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// OrderService performs order-related domain operations.
type OrderService struct{}

// CreateOrder builds a new Order entity with NEW status.
func (OrderService) CreateOrder(number vo.OrderNumber, userID vo.UserID, now time.Time) *entity.Order {
	return &entity.Order{
		Number:     number,
		UserID:     userID,
		Status:     entity.OrderStatusNew,
		UploadedAt: now,
	}
}
