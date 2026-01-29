package entity

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// OrderStatus — статус расчёта заказа в системе лояльности.
// В API/домене — строки; в БД храним SMALLINT (маппинг в адаптере).
const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type OrderStatus string

// Order — заказ, загруженный пользователем для расчёта баллов.
// Идентификация по Number (natural key); суррогатный id только в БД.
type Order struct {
	Number      vo.OrderNumber
	UserID      vo.UserID
	Status      OrderStatus
	Accrual     *vo.Points // nil, пока нет начисления или при INVALID
	UploadedAt  time.Time
	ProcessedAt *time.Time
}

// MarkProcessed переводит заказ в статус PROCESSED и фиксирует начисление и время.
func (o *Order) MarkProcessed(accrual vo.Points, now time.Time) {
	o.Status = OrderStatusProcessed
	o.Accrual = &accrual
	o.ProcessedAt = &now
}

// MarkInvalid переводит заказ в статус INVALID.
func (o *Order) MarkInvalid(now time.Time) {
	o.Status = OrderStatusInvalid
	o.Accrual = nil
	o.ProcessedAt = &now
}

// MarkProcessing переводит заказ в обработку (ожидание ответа от accrual).
func (o *Order) MarkProcessing() {
	o.Status = OrderStatusProcessing
}
