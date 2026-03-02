package factory

import (
	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/orders/application/dto"
	"gophermart/internal/gophermart/modules/orders/application/port"
	"gophermart/internal/gophermart/modules/orders/application/usecase"
	"gophermart/internal/gophermart/modules/orders/domain/vo"
)

// Params contains dependencies required to build orders use cases.
type Params struct {
	OrderRepo         port.OrderRepository
	BalanceGateway    port.BalanceGateway
	Validator         vo.OrderNumberValidator
	AccrualClient     port.AccrualClient
	Transactor        appport.Transactor
	Clock             appport.Clock
	Log               appport.Logger
	BatchSize         int
	MaxWorkers        int
	OptimisticRetries int
}

// UseCases holds orders module use cases exposed to composition root.
type UseCases struct {
	UploadOrder    appport.UseCase[dto.UploadOrderInput, struct{}]
	ListOrders     appport.UseCase[vo.UserID, []dto.OrderOutput]
	ProcessAccrual appport.BackgroundRunner
}

// NewUseCases builds orders module use cases.
func NewUseCases(p Params) UseCases {
	return UseCases{
		UploadOrder: usecase.NewUploadOrder(p.OrderRepo, p.OrderRepo, p.Validator, p.Clock),
		ListOrders:  usecase.NewListOrders(p.OrderRepo),
		ProcessAccrual: usecase.NewProcessAccrual(
			p.OrderRepo, p.OrderRepo, p.BalanceGateway, p.AccrualClient,
			p.Transactor, p.Clock, p.Log, p.BatchSize, p.MaxWorkers, p.OptimisticRetries,
		),
	}
}
