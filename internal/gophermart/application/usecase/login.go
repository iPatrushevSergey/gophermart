package usecase

import (
	"context"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/application/dto"
	"gophermart/internal/gophermart/application/port"
	"gophermart/internal/gophermart/domain/vo"
)

// LoginUser authenticates by login and password.
type LoginUser struct {
	userRepo port.UserRepository
	hasher   port.PasswordHasher
}

// NewLoginUser returns the login use case (interactor) as port abstraction.
func NewLoginUser(userRepo port.UserRepository, hasher port.PasswordHasher) port.UseCase[dto.LoginInput, vo.UserID] {
	return &LoginUser{userRepo: userRepo, hasher: hasher}
}

// Execute checks credentials; returns user ID or application.ErrInvalidCredentials.
func (uc *LoginUser) Execute(ctx context.Context, in dto.LoginInput) (vo.UserID, error) {
	u, err := uc.userRepo.FindByLogin(ctx, in.Login)
	if err != nil {
		return 0, err
	}
	if u == nil || !uc.hasher.Compare(in.Password, u.PasswordHash) {
		return 0, application.ErrInvalidCredentials
	}
	return u.ID, nil
}
