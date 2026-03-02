package entity

import (
	"time"

	"gophermart/internal/gophermart/modules/identity/domain/vo"
)

// User is identity module aggregate root for authentication data.
type User struct {
	ID           vo.UserID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser creates a new User entity.
func NewUser(login, passwordHash string, now time.Time) *User {
	return &User{
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
