package entity

import (
	"gophermart/internal/gophermart/domain/vo"
	"time"
)

// User is the loyalty system user.
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
