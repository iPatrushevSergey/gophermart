package api

import (
	"context"
	"time"
)

// ApplyAccrualInput is a module API input for crediting user balance from processed orders.
type ApplyAccrualInput struct {
	UserID      int64
	Amount      float64
	ProcessedAt time.Time
}

// AccrualAPI defines the balance module contract exposed to other modules.
type AccrualAPI interface {
	ApplyAccrual(ctx context.Context, in ApplyAccrualInput) error
}
