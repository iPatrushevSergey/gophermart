package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// Transactor manages database transactions via pgxpool with retry support.
type Transactor struct {
	pool  *pgxpool.Pool
	retry RetryConfig
}

// NewTransactor creates a new Transactor.
func NewTransactor(pool *pgxpool.Pool, retry RetryConfig) *Transactor {
	return &Transactor{pool: pool, retry: retry}
}

// RunInTransaction executes the given function within a transaction.
// Retries the entire transaction on retriable errors (e.g. deadlock, serialization failure).
func (t *Transactor) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// If transaction already exists in context, reuse it (nested call).
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	return DoWithRetry(ctx, t.retry, func() error {
		return t.runTx(ctx, fn)
	})
}

func (t *Transactor) runTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctxWithTx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetQuerier returns the transaction from context if present, otherwise the pool.
func (t *Transactor) GetQuerier(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return t.pool
}

// RetryConfig returns the retry configuration for use by repositories.
func (t *Transactor) RetryConfig() RetryConfig {
	return t.retry
}
