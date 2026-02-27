package model

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// Order is the DB projection of the orders table row.
type Order struct {
	Number      vo.OrderNumber
	UserID      vo.UserID
	Status      int16
	Accrual     *float64
	UploadedAt  time.Time
	ProcessedAt *time.Time
}
