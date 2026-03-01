package usecase

import (
	"context"

	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/service"
	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
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
) balanceapi.AccountAPI {
	return &OpenAccount{
		balanceWriter: balanceWriter,
		balanceSvc:    balanceSvc,
	}
}

// OpenAccount creates and persists a zero-balance account.
func (uc *OpenAccount) OpenAccount(ctx context.Context, in balanceapi.OpenAccountInput) error {
	acc := uc.balanceSvc.CreateAccount(in.UserID, in.CreatedAt)
	return uc.balanceWriter.Create(ctx, acc)
}
