package entity

import (
	"gophermart/internal/gophermart/domain/vo"
	"time"
)

// User — сущность пользователя системы лояльности.
type User struct {
	ID           vo.UserID
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
