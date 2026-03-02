package factory

import (
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/orders/application/dto"
	"gophermart/internal/gophermart/modules/orders/domain/vo"
)

// UseCaseFactory provides orders use cases to the presentation layer.
type UseCaseFactory interface {
	UploadOrderUseCase() port.UseCase[dto.UploadOrderInput, struct{}]
	ListOrdersUseCase() port.UseCase[vo.UserID, []dto.OrderOutput]
	ProcessAccrualUseCase() port.BackgroundRunner
}
