package port

import (
	"context"

	"gophermart/internal/gophermart/modules/identity/domain/entity"
	"gophermart/internal/gophermart/modules/identity/domain/vo"
)

// UserReader provides read-only access to users for identity module.
type UserReader interface {
	FindByID(ctx context.Context, id vo.UserID) (*entity.User, error)
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}

// UserWriter provides write access to users for identity module.
type UserWriter interface {
	Create(ctx context.Context, u *entity.User) error
}

// UserRepository combines reader and writer for identity DI wiring.
type UserRepository interface {
	UserReader
	UserWriter
}
