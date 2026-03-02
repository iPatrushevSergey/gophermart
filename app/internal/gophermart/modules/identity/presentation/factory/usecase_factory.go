package factory

import (
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/identity/application/dto"
	"gophermart/internal/gophermart/modules/identity/domain/vo"
)

// UseCaseFactory provides identity use cases to the presentation layer.
type UseCaseFactory interface {
	RegisterUseCase() port.UseCase[dto.RegisterInput, vo.UserID]
	LoginUseCase() port.UseCase[dto.LoginInput, vo.UserID]
}
