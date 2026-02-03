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
}
