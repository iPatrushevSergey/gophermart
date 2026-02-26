package postgres

import (
	"context"

	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// WithdrawalRepository is a PostgreSQL implementation of port.WithdrawalRepository.
type WithdrawalRepository struct {
	transactor *Transactor
}

// NewWithdrawalRepository creates a new WithdrawalRepository.
func NewWithdrawalRepository(transactor *Transactor) *WithdrawalRepository {
	return &WithdrawalRepository{transactor: transactor}
}

// Create inserts a new withdrawal record.
func (r *WithdrawalRepository) Create(ctx context.Context, w *entity.Withdrawal) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		_, err := q.Exec(ctx, `
			INSERT INTO withdrawals (user_id, order_number, amount, processed_at)
			VALUES ($1, $2, $3, $4)
		`, w.UserID, w.OrderNumber.String(), w.Amount, w.ProcessedAt)

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

		result = result[:0]
		for rows.Next() {
			var w entity.Withdrawal
			var numStr string

			if err := rows.Scan(
				&w.UserID,
				&numStr,
				&w.Amount,
				&w.ProcessedAt,
			); err != nil {
				return err
			}

			w.OrderNumber = vo.OrderNumber(numStr)
			result = append(result, w)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
