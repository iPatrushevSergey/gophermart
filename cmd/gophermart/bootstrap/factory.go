package bootstrap

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/application/usecase"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
	"gophermart/internal/gophermart/presentation/factory"
)

// useCaseFactory implements factory.UseCaseFactory; built in composition root.
type useCaseFactory struct {
	register port.UseCase[dto.RegisterInput, vo.UserID]
	login    port.UseCase[dto.LoginInput, vo.UserID]
}

// NewUseCaseFactory builds the use case factory with the given dependencies.
func NewUseCaseFactory(
	userRepo port.UserRepository,
	hasher port.PasswordHasher,
	tokens port.TokenProvider,
	userSvc service.UserService,
) factory.UseCaseFactory {
	register := usecase.NewRegisterUser(userRepo, hasher, userSvc)
	login := usecase.NewLoginUser(userRepo, hasher)
	return &useCaseFactory{register: register, login: login}
}

// RegisterUseCase returns the register use case.
func (f *useCaseFactory) RegisterUseCase() port.UseCase[dto.RegisterInput, vo.UserID] {
	return f.register
}

// LoginUseCase returns the login use case.
func (f *useCaseFactory) LoginUseCase() port.UseCase[dto.LoginInput, vo.UserID] {
	return f.login
}
