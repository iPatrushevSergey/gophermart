package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// BalanceAccountReader provides read-only access to balance accounts.
type BalanceAccountReader interface {
	FindByUserID(ctx context.Context, userID vo.UserID) (*entity.BalanceAccount, error)
}

// BalanceAccountWriter provides write access to balance accounts.
type BalanceAccountWriter interface {
	Create(ctx context.Context, acc *entity.BalanceAccount) error
	Update(ctx context.Context, acc *entity.BalanceAccount) error
}

// BalanceAccountRepository combines reader and writer for DI wiring.
type BalanceAccountRepository interface {
	BalanceAccountReader
	BalanceAccountWriter
}
