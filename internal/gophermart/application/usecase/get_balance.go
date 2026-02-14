package usecase

import (
	"context"

	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
)

// GetBalance returns the current balance for the given user.
type GetBalance struct {
	balanceRepo port.BalanceAccountRepository
}

// NewGetBalance returns the get balance use case.
func NewGetBalance(balanceRepo port.BalanceAccountRepository) port.UseCase[vo.UserID, dto.BalanceOutput] {
	return &GetBalance{balanceRepo: balanceRepo}
}

// Execute loads the balance account and maps it to BalanceOutput.
//
// Errors:
//   - application.ErrNotFound — balance account does not exist
func (uc *GetBalance) Execute(ctx context.Context, userID vo.UserID) (dto.BalanceOutput, error) {
	acc, err := uc.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return dto.BalanceOutput{}, err
	}

	return dto.BalanceOutput{
		Current:   float64(acc.Current),
		Withdrawn: float64(acc.WithdrawnTotal),
	}, nil
}
