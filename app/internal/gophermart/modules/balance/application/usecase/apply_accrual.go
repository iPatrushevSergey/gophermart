package usecase

import (
	"context"

	"gophermart/internal/gophermart/modules/balance/application/api"
	"gophermart/internal/gophermart/modules/balance/application/port"
	"gophermart/internal/gophermart/modules/balance/domain/vo"
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
) api.AccrualAPI {
	return &ApplyAccrual{
		balanceReader: balanceReader,
		balanceWriter: balanceWriter,
	}
}

// ApplyAccrual loads account, adds accrual and persists it.
func (uc *ApplyAccrual) ApplyAccrual(ctx context.Context, in api.ApplyAccrualInput) error {
	acc, err := uc.balanceReader.FindByUserID(ctx, vo.UserID(in.UserID))
	if err != nil {
		return err
	}

	acc.AddAccrual(vo.Points(in.Amount), in.ProcessedAt)
	return uc.balanceWriter.Update(ctx, acc)
}
