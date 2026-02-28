package factory

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
	ordersusecase "gophermart/internal/gophermart/modules/orders/application/usecase"
)

// Params contains dependencies required to build orders use cases.
type Params struct {
	OrderRepo         port.OrderRepository
	BalanceRepo       port.BalanceAccountRepository
	Validator         vo.OrderNumberValidator
	AccrualClient     port.AccrualClient
	Transactor        port.Transactor
	Clock             port.Clock
	Log               port.Logger
	BatchSize         int
	MaxWorkers        int
	OptimisticRetries int
}

// UseCases holds orders module use cases exposed to composition root.
type UseCases struct {
	UploadOrder    port.UseCase[dto.UploadOrderInput, struct{}]
	ListOrders     port.UseCase[vo.UserID, []dto.OrderOutput]
	ProcessAccrual port.BackgroundRunner
}

// NewUseCases builds orders module use cases.
func NewUseCases(p Params) UseCases {
	return UseCases{
		UploadOrder: ordersusecase.NewUploadOrder(p.OrderRepo, p.OrderRepo, p.Validator, p.Clock),
		ListOrders:  ordersusecase.NewListOrders(p.OrderRepo),
		ProcessAccrual: ordersusecase.NewProcessAccrual(
			p.OrderRepo, p.OrderRepo, p.BalanceRepo, p.BalanceRepo, p.AccrualClient,
			p.Transactor, p.Clock, p.Log, p.BatchSize, p.MaxWorkers, p.OptimisticRetries,
		),
	}
}
