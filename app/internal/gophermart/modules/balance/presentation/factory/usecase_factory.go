package factory

import (
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/balance/application/dto"
	"gophermart/internal/gophermart/modules/balance/domain/vo"
)

// UseCaseFactory provides balance use cases to the presentation layer.
type UseCaseFactory interface {
	GetBalanceUseCase() port.UseCase[vo.UserID, dto.BalanceOutput]
	WithdrawUseCase() port.UseCase[dto.WithdrawInput, struct{}]
	ListWithdrawalsUseCase() port.UseCase[vo.UserID, []dto.WithdrawalOutput]
}
