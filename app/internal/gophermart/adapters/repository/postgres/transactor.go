package postgres

import (
	"context"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trmmanager "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Transactor coordinates PostgreSQL transactions via go-transaction-manager
// and applies retry policy for transaction and repository operations.
type Transactor struct {
	// trManager opens/commits/rolls back transactions and stores tx in context.
	trManager *trmmanager.Manager
	// getter resolves either tx from context or fallback DB connection.
	getter *trmpgx.CtxGetter
	// pool is the default DB connection used when context has no active tx.
	pool *pgxpool.Pool
	// retryOpts configure retry behavior (attempts, backoff, retriable errors).
	retryOpts []RetryOption
}

// NewTransactor creates a transactor with transaction manager and retry options.
func NewTransactor(pool *pgxpool.Pool, opts ...RetryOption) *Transactor {
	return &Transactor{
		trManager: trmmanager.Must(trmpgx.NewDefaultFactory(pool)),
		getter:    trmpgx.DefaultCtxGetter,
		pool:      pool,
		retryOpts: opts,
	}
}

// RunInTransaction executes fn inside a transaction.
// The whole transaction is retried according to retryOpts on retriable errors.
func (t *Transactor) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return DoWithRetry(ctx, func() error {
		return t.trManager.Do(ctx, fn)
	}, t.retryOpts...)
}

// GetQuerier returns tx from context when inside transaction, otherwise pool.
func (t *Transactor) GetQuerier(ctx context.Context) Querier {
	return t.getter.DefaultTrOrDB(ctx, t.pool)
}

// DoWithRetry executes op with the transactor retry configuration.
func (t *Transactor) DoWithRetry(ctx context.Context, op func() error) error {
	return DoWithRetry(ctx, op, t.retryOpts...)
}
