package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// BalanceAccountRepository is a PostgreSQL implementation of port.BalanceAccountRepository.
type BalanceAccountRepository struct {
	transactor *Transactor
}

// NewBalanceAccountRepository creates a new BalanceAccountRepository.
func NewBalanceAccountRepository(transactor *Transactor) *BalanceAccountRepository {
	return &BalanceAccountRepository{transactor: transactor}
}

// Create inserts a new balance account.
func (r *BalanceAccountRepository) Create(ctx context.Context, acc *entity.BalanceAccount) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		_, err := q.Exec(ctx, `
			INSERT INTO balance_accounts (user_id, current, withdrawn_total, created_at, updated_at, version)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, acc.UserID, acc.Current, acc.WithdrawnTotal, acc.CreatedAt, acc.UpdatedAt, acc.Version)

		return err
	})
}

// FindByUserID returns the balance account for the given user or application.ErrNotFound.
func (r *BalanceAccountRepository) FindByUserID(ctx context.Context, userID vo.UserID) (*entity.BalanceAccount, error) {
	var acc entity.BalanceAccount

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		return q.QueryRow(ctx, `
			SELECT user_id, current, withdrawn_total, created_at, updated_at, version
			FROM balance_accounts
			WHERE user_id = $1
		`, userID).Scan(
			&acc.UserID,
			&acc.Current,
			&acc.WithdrawnTotal,
			&acc.CreatedAt,
			&acc.UpdatedAt,
			&acc.Version,
		)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	return &acc, nil
}

// Update updates the balance account with optimistic locking.
// Returns application.ErrOptimisticLock if the version in DB does not match acc.Version.
func (r *BalanceAccountRepository) Update(ctx context.Context, acc *entity.BalanceAccount) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		var newVersion int64
		err := q.QueryRow(ctx, `
			UPDATE balance_accounts
			SET current = $1, withdrawn_total = $2, updated_at = $3, version = version + 1
			WHERE user_id = $4 AND version = $5
			RETURNING version
		`, acc.Current, acc.WithdrawnTotal, acc.UpdatedAt, acc.UserID, acc.Version).Scan(&newVersion)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return application.ErrOptimisticLock
			}
			return err
		}

		acc.Version = newVersion
		return nil
	})
}
