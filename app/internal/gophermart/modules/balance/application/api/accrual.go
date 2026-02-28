package api

import (
	"context"
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// ApplyAccrualInput is a module API input for crediting user balance from processed orders.
type ApplyAccrualInput struct {
	UserID      vo.UserID
	Amount      vo.Points
	ProcessedAt time.Time
}

// AccrualAPI defines the balance module contract exposed to other modules.
type AccrualAPI interface {
	ApplyAccrual(ctx context.Context, in ApplyAccrualInput) error
}
