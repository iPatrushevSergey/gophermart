package postgres

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

// statusToInt maps domain OrderStatus to DB SMALLINT.
var statusToInt = map[entity.OrderStatus]int16{
	entity.OrderStatusNew:        0,
	entity.OrderStatusProcessing: 1,
	entity.OrderStatusInvalid:    2,
	entity.OrderStatusProcessed:  3,
}

// intToStatus maps DB SMALLINT to domain OrderStatus.
var intToStatus = map[int16]entity.OrderStatus{
	0: entity.OrderStatusNew,
	1: entity.OrderStatusProcessing,
	2: entity.OrderStatusInvalid,
	3: entity.OrderStatusProcessed,
}

// OrderRepository is a PostgreSQL implementation of port.OrderRepository.
type OrderRepository struct {
	transactor *Transactor
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(transactor *Transactor) *OrderRepository {
	return &OrderRepository{transactor: transactor}
}

// Create inserts a new order.
func (r *OrderRepository) Create(ctx context.Context, o *entity.Order) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		_, err := q.Exec(ctx, `
			INSERT INTO orders (number, user_id, status, accrual, uploaded_at)
			VALUES ($1, $2, $3, $4, $5)
		`, o.Number.String(), o.UserID, statusToInt[o.Status], o.Accrual, o.UploadedAt)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return application.ErrAlreadyExists
			}
			return err
		}

		return nil
	})
}

// FindByNumber returns the order by its number or application.ErrNotFound.
func (r *OrderRepository) FindByNumber(ctx context.Context, number vo.OrderNumber) (*entity.Order, error) {
	var o entity.Order
	var statusInt int16
	var numStr string

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		return q.QueryRow(ctx, `
			SELECT number, user_id, status, accrual, uploaded_at, processed_at
			FROM orders
			WHERE number = $1
		`, number.String()).Scan(
			&numStr,
			&o.UserID,
			&statusInt,
			&o.Accrual,
			&o.UploadedAt,
			&o.ProcessedAt,
		)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	o.Number = vo.OrderNumber(numStr)
	o.Status = intToStatus[statusInt]

	return &o, nil
}

// ListByUserID returns all orders for the given user, sorted by uploaded_at DESC.
func (r *OrderRepository) ListByUserID(ctx context.Context, userID vo.UserID) ([]entity.Order, error) {
	var result []entity.Order

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT number, user_id, status, accrual, uploaded_at, processed_at
			FROM orders
			WHERE user_id = $1
			ORDER BY uploaded_at DESC
		`, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		result = result[:0] // reset on retry
		for rows.Next() {
			var o entity.Order
			var statusInt int16
			var numStr string

			if err := rows.Scan(
				&numStr,
				&o.UserID,
				&statusInt,
				&o.Accrual,
				&o.UploadedAt,
				&o.ProcessedAt,
			); err != nil {
				return err
			}

			o.Number = vo.OrderNumber(numStr)
			o.Status = intToStatus[statusInt]
			result = append(result, o)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListByStatuses returns orders matching any of the given statuses, limited by limit.
func (r *OrderRepository) ListByStatuses(ctx context.Context, statuses []entity.OrderStatus, limit int) ([]entity.Order, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	ints := make([]int16, len(statuses))
	for i, s := range statuses {
		v, ok := statusToInt[s]
		if !ok {
			return nil, fmt.Errorf("unknown order status: %s", s)
		}
		ints[i] = v
	}

	var result []entity.Order

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT number, user_id, status, accrual, uploaded_at, processed_at
			FROM orders
			WHERE status = ANY($1)
			ORDER BY uploaded_at ASC
			LIMIT $2
		`, ints, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		result = result[:0]
		for rows.Next() {
			var o entity.Order
			var statusInt int16
			var numStr string

			if err := rows.Scan(
				&numStr,
				&o.UserID,
				&statusInt,
				&o.Accrual,
				&o.UploadedAt,
				&o.ProcessedAt,
			); err != nil {
				return err
			}

			o.Number = vo.OrderNumber(numStr)
			o.Status = intToStatus[statusInt]
			result = append(result, o)
		}

		return rows.Err()
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// StreamByStatuses returns a lazy iterator over orders matching the given statuses.
// The DB cursor remains open while the caller iterates; rows are closed automatically
// when the iterator is exhausted or the caller breaks out of the range loop.
func (r *OrderRepository) StreamByStatuses(ctx context.Context, statuses []entity.OrderStatus, limit int) iter.Seq2[entity.Order, error] {
	return func(yield func(entity.Order, error) bool) {
		if len(statuses) == 0 {
			return
		}

		ints := make([]int16, len(statuses))
		for i, s := range statuses {
			v, ok := statusToInt[s]
			if !ok {
				yield(entity.Order{}, fmt.Errorf("unknown order status: %s", s))
				return
			}
			ints[i] = v
		}

		q := r.transactor.GetQuerier(ctx)
		rows, err := q.Query(ctx, `
			SELECT number, user_id, status, accrual, uploaded_at, processed_at
			FROM orders
			WHERE status = ANY($1)
			ORDER BY uploaded_at ASC
			LIMIT $2
		`, ints, limit)
		if err != nil {
			yield(entity.Order{}, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var o entity.Order
			var statusInt int16
			var numStr string

			if err := rows.Scan(&numStr, &o.UserID, &statusInt, &o.Accrual, &o.UploadedAt, &o.ProcessedAt); err != nil {
				yield(entity.Order{}, err)
				return
			}

			o.Number = vo.OrderNumber(numStr)
			o.Status = intToStatus[statusInt]

			if !yield(o, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(entity.Order{}, err)
		}
	}
}

// Update updates the order status, accrual, and processed_at.
func (r *OrderRepository) Update(ctx context.Context, o *entity.Order) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		_, err := q.Exec(ctx, `
			UPDATE orders
			SET status = $1, accrual = $2, processed_at = $3
			WHERE number = $4
		`, statusToInt[o.Status], o.Accrual, o.ProcessedAt, o.Number.String())

		return err
	})
}
