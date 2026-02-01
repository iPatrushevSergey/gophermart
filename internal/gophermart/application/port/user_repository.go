package port

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, u *entity.User) error
	FindByID(ctx context.Context, id vo.UserID) (*entity.User, error)
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
}
