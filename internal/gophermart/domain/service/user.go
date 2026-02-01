package service

import (
	"time"

	"gophermart/internal/gophermart/domain/entity"
)

// UserService performs user-related domain operations.
type UserService struct{}

// CreateUser builds a new User entity.
func (UserService) CreateUser(login, passwordHash string, now time.Time) *entity.User {
	return &entity.User{
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
