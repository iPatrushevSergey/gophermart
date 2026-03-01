package api

import (
	"context"
	"time"

	"gophermart/internal/gophermart/domain/vo"
)

// OpenAccountInput is a module API input for creating a user's balance account.
type OpenAccountInput struct {
	UserID    vo.UserID
	CreatedAt time.Time
}

// AccountAPI defines the balance account contract exposed to other modules.
type AccountAPI interface {
	OpenAccount(ctx context.Context, in OpenAccountInput) error
}
