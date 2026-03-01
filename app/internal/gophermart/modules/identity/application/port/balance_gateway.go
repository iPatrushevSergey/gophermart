package port

import (
	"context"
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// BalanceGateway is an identity-module port for opening user balance accounts.
type BalanceGateway interface {
	OpenAccount(ctx context.Context, userID vo.UserID, createdAt time.Time) error
}
