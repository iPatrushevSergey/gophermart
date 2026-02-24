package factory

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
)

// UseCaseFactory provides use cases to the presentation layer.
type UseCaseFactory interface {
	RegisterUseCase() port.UseCase[dto.RegisterInput, vo.UserID]
	LoginUseCase() port.UseCase[dto.LoginInput, vo.UserID]
	UploadOrderUseCase() port.UseCase[dto.UploadOrderInput, struct{}]
	ListOrdersUseCase() port.UseCase[vo.UserID, []dto.OrderOutput]
	GetBalanceUseCase() port.UseCase[vo.UserID, dto.BalanceOutput]
	WithdrawUseCase() port.UseCase[dto.WithdrawInput, struct{}]
	ListWithdrawalsUseCase() port.UseCase[vo.UserID, []dto.WithdrawalOutput]
	ProcessAccrualUseCase() port.BackgroundRunner
}
