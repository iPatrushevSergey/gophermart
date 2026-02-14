package usecase

import (
	"context"
	"time"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/service"
	"gophermart/internal/gophermart/domain/vo"
)

// RegisterUser registers a new user and creates an initial balance account.
type RegisterUser struct {
	userRepo    port.UserRepository
	balanceRepo port.BalanceAccountRepository
	transactor  port.Transactor
	hasher      port.PasswordHasher
	userSvc     service.UserService
	balanceSvc  service.BalanceService
}

// NewRegisterUser returns the register use case (interactor) as port abstraction.
func NewRegisterUser(
	userRepo port.UserRepository,
	balanceRepo port.BalanceAccountRepository,
	transactor port.Transactor,
	hasher port.PasswordHasher,
	userSvc service.UserService,
	balanceSvc service.BalanceService,
) port.UseCase[dto.RegisterInput, vo.UserID] {
	return &RegisterUser{
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
		transactor:  transactor,
		hasher:      hasher,
		userSvc:     userSvc,
		balanceSvc:  balanceSvc,
	}
}

// Execute creates a user and balance account in a single transaction.
//
// Errors:
//   - application.ErrAlreadyExists — login is already taken
func (uc *RegisterUser) Execute(ctx context.Context, in dto.RegisterInput) (vo.UserID, error) {
	existing, err := uc.userRepo.FindByLogin(ctx, in.Login)
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

	now := time.Now()
	u := uc.userSvc.CreateUser(in.Login, hash, now)

	err = uc.transactor.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := uc.userRepo.Create(ctx, u); err != nil {
			return err
		}

		acc := uc.balanceSvc.CreateAccount(u.ID, now)
		return uc.balanceRepo.Create(ctx, acc)
	})
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}
