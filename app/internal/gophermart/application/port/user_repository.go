package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// UserReader provides read-only access to users.
type UserReader interface {
	FindByID(ctx context.Context, id vo.UserID) (*entity.User, error)
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

// UserWriter provides write access to users.
type UserWriter interface {
	Create(ctx context.Context, u *entity.User) error
}

// UserRepository combines reader and writer for DI wiring.
type UserRepository interface {
	UserReader
	UserWriter
}
