package usecase

import (
	"context"
	"errors"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// Withdraw handles deducting points from the user's balance for an order.
type Withdraw struct {
	balanceReader     port.BalanceAccountReader
	balanceWriter     port.BalanceAccountWriter
	withdrawalWriter  port.WithdrawalWriter
	transactor        port.Transactor
	validator         vo.OrderNumberValidator
	clock             port.Clock
	optimisticRetries int
}

// NewWithdraw returns the withdraw use case.
func NewWithdraw(
	balanceReader port.BalanceAccountReader,
	balanceWriter port.BalanceAccountWriter,
	withdrawalWriter port.WithdrawalWriter,
	transactor port.Transactor,
	validator vo.OrderNumberValidator,
	clock port.Clock,
	optimisticRetries int,
) port.UseCase[dto.WithdrawInput, struct{}] {
	return &Withdraw{
		balanceReader:     balanceReader,
		balanceWriter:     balanceWriter,
		withdrawalWriter:  withdrawalWriter,
		transactor:        transactor,
		validator:         validator,
		clock:             clock,
		optimisticRetries: optimisticRetries,
	}
}

// Execute validates the order number, deducts points, and creates a withdrawal record in a transaction.
// Retries the entire transaction on optimistic lock conflicts.
//
// Errors:
//   - application.ErrInvalidOrderNumber — order number failed Luhn check
//   - application.ErrInsufficientBalance — not enough points on the account
//   - application.ErrNotFound — balance account does not exist
func (uc *Withdraw) Execute(ctx context.Context, in dto.WithdrawInput) (struct{}, error) {
	orderNumber, err := vo.NewOrderNumber(uc.validator, in.OrderNumber)
	if err != nil {
		return struct{}{}, application.ErrInvalidOrderNumber
	}

	err = application.WithOptimisticRetry(uc.optimisticRetries, func() error {
		return uc.transactor.RunInTransaction(ctx, func(ctx context.Context) error {
			acc, err := uc.balanceReader.FindByUserID(ctx, in.UserID)
			if err != nil {
				return err
			}

			now := uc.clock.Now()

			if err := acc.Withdraw(vo.Points(in.Sum), now); err != nil {
				return err
			}

			w := entity.NewWithdrawal(in.UserID, orderNumber, vo.Points(in.Sum), now)

			if err := uc.withdrawalWriter.Create(ctx, w); err != nil {
				return err
			}

			return uc.balanceWriter.Update(ctx, acc)
		})
	})

	if err != nil {
		if errors.Is(err, entity.ErrInsufficientBalance) {
			return struct{}{}, application.ErrInsufficientBalance
		}
		return struct{}{}, err
	}

	return struct{}{}, nil
}
