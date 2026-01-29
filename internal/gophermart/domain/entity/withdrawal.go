package entity

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// Withdrawal — факт списания баллов в счёт оплаты заказа.
// Иммутабельное событие: создаётся один раз, не меняется.
type Withdrawal struct {
	UserID      vo.UserID
	OrderNumber vo.OrderNumber
	Amount      vo.Points
	ProcessedAt time.Time
}
