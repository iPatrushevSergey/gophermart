package dto

import (
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// UploadOrderInput is the input for uploading an order number.
type UploadOrderInput struct {
	UserID      vo.UserID
	OrderNumber string
}

// OrderOutput is the output for a single order in the list.
type OrderOutput struct {
	Number     string
	Status     string
	Accrual    *float64
	UploadedAt time.Time
}
