package factory

import (
	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/balance/application/api"
	"gophermart/internal/gophermart/modules/balance/application/dto"
	"gophermart/internal/gophermart/modules/balance/application/port"
	"gophermart/internal/gophermart/modules/balance/application/usecase"
	"gophermart/internal/gophermart/modules/balance/domain/service"
	"gophermart/internal/gophermart/modules/balance/domain/vo"
)

// Params contains dependencies required to build balance use cases.
type Params struct {
	BalanceRepo       port.BalanceAccountRepository
	WithdrawalRepo    port.WithdrawalRepository
	Transactor        appport.Transactor
	Validator         vo.OrderNumberValidator
	Clock             appport.Clock
	BalanceSvc        service.BalanceService
	OptimisticRetries int
}

// UseCases holds balance module use cases exposed to composition root.
type UseCases struct {
	GetBalance      appport.UseCase[vo.UserID, dto.BalanceOutput]
	Withdraw        appport.UseCase[dto.WithdrawInput, struct{}]
	ListWithdrawals appport.UseCase[vo.UserID, []dto.WithdrawalOutput]
	ApplyAccrual    api.AccrualAPI
	OpenAccount     api.AccountAPI
}

// NewUseCases builds balance module use cases.
func NewUseCases(p Params) UseCases {
	return UseCases{
		GetBalance:      usecase.NewGetBalance(p.BalanceRepo),
		Withdraw:        usecase.NewWithdraw(p.BalanceRepo, p.BalanceRepo, p.WithdrawalRepo, p.Transactor, p.Validator, p.Clock, p.OptimisticRetries),
		ListWithdrawals: usecase.NewListWithdrawals(p.WithdrawalRepo),
		ApplyAccrual:    usecase.NewApplyAccrual(p.BalanceRepo, p.BalanceRepo),
		OpenAccount:     usecase.NewOpenAccount(p.BalanceRepo, p.BalanceSvc),
	}
}
