package usecase

import (
	"context"

	"gophermart/internal/gophermart/modules/balance/application/api"
	"gophermart/internal/gophermart/modules/balance/application/port"
	"gophermart/internal/gophermart/modules/balance/domain/service"
	"gophermart/internal/gophermart/modules/balance/domain/vo"
)

// OpenAccount creates a new balance account for a user.
type OpenAccount struct {
	balanceWriter port.BalanceAccountWriter
	balanceSvc    service.BalanceService
}

// NewOpenAccount returns balance module API for account creation.
func NewOpenAccount(
	balanceWriter port.BalanceAccountWriter,
	balanceSvc service.BalanceService,
) api.AccountAPI {
	return &OpenAccount{
		balanceWriter: balanceWriter,
		balanceSvc:    balanceSvc,
	}
}

// OpenAccount creates and persists a zero-balance account.
func (uc *OpenAccount) OpenAccount(ctx context.Context, in api.OpenAccountInput) error {
	acc := uc.balanceSvc.CreateAccount(vo.UserID(in.UserID), in.CreatedAt)
	return uc.balanceWriter.Create(ctx, acc)
}
