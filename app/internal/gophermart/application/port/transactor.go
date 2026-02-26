package port

import "context"

// Transactor manages database transactions.
type Transactor interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
