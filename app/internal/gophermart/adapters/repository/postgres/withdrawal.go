package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"gophermart/internal/gophermart/adapters/repository/postgres/converter"
	"gophermart/internal/gophermart/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// WithdrawalRepository is a PostgreSQL implementation of port.WithdrawalRepository.
type WithdrawalRepository struct {
	transactor *Transactor
	conv       converter.WithdrawalConverter
}

// NewWithdrawalRepository creates a new WithdrawalRepository.
func NewWithdrawalRepository(transactor *Transactor) *WithdrawalRepository {
	return &WithdrawalRepository{
		transactor: transactor,
		conv:       &converter.WithdrawalConverterImpl{},
	}
}

// Create inserts a new withdrawal record.
func (r *WithdrawalRepository) Create(ctx context.Context, w *entity.Withdrawal) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		dbWithdrawal := r.conv.ToModel(*w)

		_, err := q.Exec(ctx, `
			INSERT INTO withdrawals (user_id, order_number, amount, processed_at)
			VALUES ($1, $2, $3, $4)
		`, dbWithdrawal.UserID, dbWithdrawal.OrderNumber, dbWithdrawal.Amount, dbWithdrawal.ProcessedAt)

		return err
	})
}

// ListByUserID returns all withdrawals for the given user, sorted by processed_at DESC.
func (r *WithdrawalRepository) ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Withdrawal, error) {
	var result []entity.Withdrawal

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT user_id, order_number, amount, processed_at
			FROM withdrawals
			WHERE user_id = $1
			ORDER BY processed_at DESC
		`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Withdrawal])
		if err != nil {
			return err
		}

		result = result[:0]
		for _, dbRow := range dbRows {
			result = append(result, r.conv.ToEntity(dbRow))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
