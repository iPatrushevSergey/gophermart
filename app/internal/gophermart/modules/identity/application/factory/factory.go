package factory

import (
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
	identityusecase "gophermart/internal/gophermart/modules/identity/application/usecase"
)

// Params contains dependencies required to build identity use cases.
type Params struct {
	UserRepo    port.UserRepository
	BalanceRepo port.BalanceAccountRepository
	Transactor  port.Transactor
	Hasher      port.PasswordHasher
	Clock       port.Clock
	BalanceSvc  service.BalanceService
}

// UseCases holds identity module use cases exposed to composition root.
type UseCases struct {
	Register port.UseCase[dto.RegisterInput, vo.UserID]
	Login    port.UseCase[dto.LoginInput, vo.UserID]
}

// NewUseCases builds identity module use cases.
func NewUseCases(p Params) UseCases {
	return UseCases{
		Register: identityusecase.NewRegisterUser(
			p.UserRepo, p.UserRepo, p.BalanceRepo, p.Transactor, p.Hasher, p.Clock, p.BalanceSvc,
		),
		Login: identityusecase.NewLoginUser(p.UserRepo, p.Hasher),
	}
}
