package api

import (
	"context"
	"time"
)

// OpenAccountInput is a module API input for creating a user's balance account.
type OpenAccountInput struct {
	UserID    int64
	CreatedAt time.Time
}

// AccountAPI defines the balance account contract exposed to other modules.
type AccountAPI interface {
	OpenAccount(ctx context.Context, in OpenAccountInput) error
}
