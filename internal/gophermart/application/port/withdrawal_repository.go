package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// WithdrawalRepository persists withdrawal records.
type WithdrawalRepository interface {
	Create(ctx context.Context, w *entity.Withdrawal) error
	ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Withdrawal, error)
}
