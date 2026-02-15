package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// WithdrawalReader provides read-only access to withdrawals.
type WithdrawalReader interface {
	ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Withdrawal, error)
}

// WithdrawalWriter provides write access to withdrawals.
type WithdrawalWriter interface {
	Create(ctx context.Context, w *entity.Withdrawal) error
}

// WithdrawalRepository combines reader and writer for DI wiring.
type WithdrawalRepository interface {
	WithdrawalReader
	WithdrawalWriter
}
