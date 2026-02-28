package factory

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
	balanceusecase "gophermart/internal/gophermart/modules/balance/application/usecase"
)

// Params contains dependencies required to build balance use cases.
type Params struct {
	BalanceRepo       port.BalanceAccountRepository
	WithdrawalRepo    port.WithdrawalRepository
	Transactor        port.Transactor
	Validator         vo.OrderNumberValidator
	Clock             port.Clock
	OptimisticRetries int
}

// UseCases holds balance module use cases exposed to composition root.
type UseCases struct {
	GetBalance      port.UseCase[vo.UserID, dto.BalanceOutput]
	Withdraw        port.UseCase[dto.WithdrawInput, struct{}]
	ListWithdrawals port.UseCase[vo.UserID, []dto.WithdrawalOutput]
}

// NewUseCases builds balance module use cases.
func NewUseCases(p Params) UseCases {
	return UseCases{
		GetBalance:      balanceusecase.NewGetBalance(p.BalanceRepo),
		Withdraw:        balanceusecase.NewWithdraw(p.BalanceRepo, p.BalanceRepo, p.WithdrawalRepo, p.Transactor, p.Validator, p.Clock, p.OptimisticRetries),
		ListWithdrawals: balanceusecase.NewListWithdrawals(p.WithdrawalRepo),
	}
}
