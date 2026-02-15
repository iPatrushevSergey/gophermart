package usecase

import (
	"context"

	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
)

// ListWithdrawals returns all withdrawals for the given user.
type ListWithdrawals struct {
	withdrawalReader port.WithdrawalReader
}

// NewListWithdrawals returns the list withdrawals use case.
func NewListWithdrawals(withdrawalReader port.WithdrawalReader) port.UseCase[vo.UserID, []dto.WithdrawalOutput] {
	return &ListWithdrawals{withdrawalReader: withdrawalReader}
}

// Execute fetches withdrawals and maps them to output DTOs.
// Returns an empty slice if the user has no withdrawals.
func (uc *ListWithdrawals) Execute(ctx context.Context, userID vo.UserID) ([]dto.WithdrawalOutput, error) {
	withdrawals, err := uc.withdrawalReader.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.WithdrawalOutput, 0, len(withdrawals))
	for _, w := range withdrawals {
		result = append(result, dto.WithdrawalOutput{
			OrderNumber: w.OrderNumber.String(),
			Sum:         float64(w.Amount),
			ProcessedAt: w.ProcessedAt,
		})
	}

	return result, nil
}
