package postgres

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gophermart/internal/gophermart/adapters/repository/postgres/converter"
	"gophermart/internal/gophermart/adapters/repository/postgres/model"
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

// OrderRepository is a PostgreSQL implementation of port.OrderRepository.
type OrderRepository struct {
	transactor *Transactor
	conv       converter.OrderConverter
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(transactor *Transactor) *OrderRepository {
	return &OrderRepository{
		transactor: transactor,
		conv:       &converter.OrderConverterImpl{},
	}
}

// Create inserts a new order.
func (r *OrderRepository) Create(ctx context.Context, o *entity.Order) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		dbOrder, err := r.conv.ToModel(*o)
		if err != nil {
			return err
		}

		_, err = q.Exec(ctx, `
			INSERT INTO orders (number, user_id, status, accrual, uploaded_at)
			VALUES ($1, $2, $3, $4, $5)
		`, dbOrder.Number, dbOrder.UserID, dbOrder.Status, dbOrder.Accrual, dbOrder.UploadedAt)
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

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT number, user_id, status, accrual, uploaded_at, processed_at
			FROM orders
			WHERE number = $1
		`, number.String())
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRow, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[model.Order])
		if err != nil {
			return err
		}

		o, err = r.conv.ToEntity(dbRow)
		return err
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

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

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Order])
		if err != nil {
			return err
		}

		result = result[:0] // reset on retry
		for _, dbRow := range dbRows {
			o, err := r.conv.ToEntity(dbRow)
			if err != nil {
				return err
			}
			result = append(result, o)
		}

		return nil
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

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Order])
		if err != nil {
			return err
		}

		result = result[:0]
		for _, dbRow := range dbRows {
			o, err := r.conv.ToEntity(dbRow)
			if err != nil {
				return err
			}
			result = append(result, o)
		}

		return nil
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
			dbRow, err := pgx.RowToStructByPos[model.Order](rows)
			if err != nil {
				yield(entity.Order{}, err)
				return
			}

			o, err := r.conv.ToEntity(dbRow)
			if err != nil {
				yield(entity.Order{}, err)
				return
			}

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
		dbOrder, err := r.conv.ToModel(*o)
		if err != nil {
			return err
		}

		_, err = q.Exec(ctx, `
			UPDATE orders
			SET status = $1, accrual = $2, processed_at = $3
			WHERE number = $4
		`, dbOrder.Status, dbOrder.Accrual, dbOrder.ProcessedAt, dbOrder.Number)

		return err
	})
}
