package entity

import (
	"gophermart/internal/gophermart/domain/vo"
	"time"
)

// User — the essence of the loyalty system user.
type User struct {
	ID           vo.UserID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
