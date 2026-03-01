package factory

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
)

// UseCaseFactory provides orders use cases to the presentation layer.
type UseCaseFactory interface {
	UploadOrderUseCase() port.UseCase[dto.UploadOrderInput, struct{}]
	ListOrdersUseCase() port.UseCase[vo.UserID, []dto.OrderOutput]
}
