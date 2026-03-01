package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// WithdrawalReader provides read-only access to withdrawals for balance module.
type WithdrawalReader interface {
	ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Withdrawal, error)
}

// WithdrawalWriter provides write access to withdrawals for balance module.
type WithdrawalWriter interface {
	Create(ctx context.Context, w *entity.Withdrawal) error
}

// WithdrawalRepository combines reader and writer for balance DI wiring.
type WithdrawalRepository interface {
	WithdrawalReader
	WithdrawalWriter
}
