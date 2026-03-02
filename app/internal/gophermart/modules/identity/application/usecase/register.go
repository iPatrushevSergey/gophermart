package usecase

import (
	"context"

	"gophermart/internal/gophermart/application"
	appport "gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/modules/identity/application/dto"
	"gophermart/internal/gophermart/modules/identity/application/port"
	"gophermart/internal/gophermart/modules/identity/domain/entity"
	"gophermart/internal/gophermart/modules/identity/domain/vo"
)

// RegisterUser registers a new user and creates an initial balance account.
type RegisterUser struct {
	userReader     port.UserReader
	userWriter     port.UserWriter
	balanceGateway port.BalanceGateway
	transactor     appport.Transactor
	hasher         appport.PasswordHasher
	clock          appport.Clock
}

// NewRegisterUser returns the register use case (interactor) as port abstraction.
func NewRegisterUser(
	userReader port.UserReader,
	userWriter port.UserWriter,
	balanceGateway port.BalanceGateway,
	transactor appport.Transactor,
	hasher appport.PasswordHasher,
	clock appport.Clock,
) appport.UseCase[dto.RegisterInput, vo.UserID] {
	return &RegisterUser{
		userReader:     userReader,
		userWriter:     userWriter,
		balanceGateway: balanceGateway,
		transactor:     transactor,
		hasher:         hasher,
		clock:          clock,
	}
}

// Execute creates a user and balance account in a single transaction.
//
// Errors:
//   - application.ErrAlreadyExists — login is already taken
func (uc *RegisterUser) Execute(ctx context.Context, in dto.RegisterInput) (vo.UserID, error) {
	existing, err := uc.userReader.FindByLogin(ctx, in.Login)
	if err != nil && err != application.ErrNotFound {
		return 0, err
	}
	if existing != nil {
		return 0, application.ErrAlreadyExists
	}

	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return 0, err
	}

	now := uc.clock.Now()
	u := entity.NewUser(in.Login, hash, now)

	err = uc.transactor.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := uc.userWriter.Create(ctx, u); err != nil {
			return err
		}

		return uc.balanceGateway.OpenAccount(ctx, u.ID, now)
	})
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}
