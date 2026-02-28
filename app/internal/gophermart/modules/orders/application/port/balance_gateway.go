package port

import (
	"context"
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// BalanceGateway is an orders-module port for cross-module balance accrual updates.
type BalanceGateway interface {
	ApplyAccrual(ctx context.Context, userID vo.UserID, amount vo.Points, processedAt time.Time) error
}
