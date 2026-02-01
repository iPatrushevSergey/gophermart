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

// RegisterUser registers a new user.
type RegisterUser struct {
	userRepo port.UserRepository
	hasher   port.PasswordHasher
	userSvc  service.UserService
}

// NewRegisterUser returns the register use case (interactor) as port abstraction.
func NewRegisterUser(userRepo port.UserRepository, hasher port.PasswordHasher, userSvc service.UserService) port.UseCase[dto.RegisterInput, vo.UserID] {
	return &RegisterUser{userRepo: userRepo, hasher: hasher, userSvc: userSvc}
}

// Execute creates a user from input; returns user ID or application.ErrAlreadyExists if login is taken.
func (uc *RegisterUser) Execute(ctx context.Context, in dto.RegisterInput) (vo.UserID, error) {
	existing, err := uc.userRepo.FindByLogin(ctx, in.Login)
	if err != nil {
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
	if err := uc.userRepo.Create(ctx, u); err != nil {
		return 0, err
	}
	return u.ID, nil
}
