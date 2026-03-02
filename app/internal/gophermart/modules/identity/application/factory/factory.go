package factory

import (
	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/identity/application/dto"
	"gophermart/internal/gophermart/modules/identity/application/port"
	"gophermart/internal/gophermart/modules/identity/application/usecase"
	"gophermart/internal/gophermart/modules/identity/domain/vo"
)

// Params contains dependencies required to build identity use cases.
type Params struct {
	UserRepo       port.UserRepository
	BalanceGateway port.BalanceGateway
	Transactor     appport.Transactor
	Hasher         appport.PasswordHasher
	Clock          appport.Clock
}

// UseCases holds identity module use cases exposed to composition root.
type UseCases struct {
	Register appport.UseCase[dto.RegisterInput, vo.UserID]
	Login    appport.UseCase[dto.LoginInput, vo.UserID]
}

// NewUseCases builds identity module use cases.
func NewUseCases(p Params) UseCases {
	return UseCases{
		Register: usecase.NewRegisterUser(
			p.UserRepo, p.UserRepo, p.BalanceGateway, p.Transactor, p.Hasher, p.Clock,
		),
		Login: usecase.NewLoginUser(p.UserRepo, p.Hasher),
	}
}
