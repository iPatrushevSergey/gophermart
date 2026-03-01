package usecase

import (
	"context"

	"gophermart/internal/gophermart/application/port"
	balanceapi "gophermart/internal/gophermart/modules/balance/application/api"
)

// ApplyAccrual credits user balance for a processed order.
type ApplyAccrual struct {
	balanceReader port.BalanceAccountReader
	balanceWriter port.BalanceAccountWriter
}

// NewApplyAccrual returns balance module API for accrual crediting.
func NewApplyAccrual(
	balanceReader port.BalanceAccountReader,
	balanceWriter port.BalanceAccountWriter,
) balanceapi.AccrualAPI {
	return &ApplyAccrual{
		balanceReader: balanceReader,
		balanceWriter: balanceWriter,
	}
}

// ApplyAccrual loads account, adds accrual and persists it.
func (uc *ApplyAccrual) ApplyAccrual(ctx context.Context, in balanceapi.ApplyAccrualInput) error {
	acc, err := uc.balanceReader.FindByUserID(ctx, in.UserID)
	if err != nil {
		return err
	}

	acc.AddAccrual(in.Amount, in.ProcessedAt)
	return uc.balanceWriter.Update(ctx, acc)
}
