package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// BalanceAccountRepository persists user balance accounts.
type BalanceAccountRepository interface {
	Create(ctx context.Context, acc *entity.BalanceAccount) error
	FindByUserID(ctx context.Context, userID vo.UserID) (*entity.BalanceAccount, error)
	Update(ctx context.Context, acc *entity.BalanceAccount) error
}
