package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

type UserRepository struct {
	transactor *Transactor
}

func NewUserRepository(transactor *Transactor) *UserRepository {
	return &UserRepository{transactor: transactor}
}

func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	return DoWithRetry(ctx, r.transactor.RetryConfig(), func() error {
		q := r.transactor.GetQuerier(ctx)

		query := `
			INSERT INTO users (login, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`

		err := q.QueryRow(ctx, query, u.Login, u.PasswordHash, u.CreatedAt, u.UpdatedAt).Scan(&u.ID)
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

func (r *UserRepository) FindByID(ctx context.Context, id vo.UserID) (*entity.User, error) {
	var u entity.User

	err := DoWithRetry(ctx, r.transactor.RetryConfig(), func() error {
		q := r.transactor.GetQuerier(ctx)

		query := `
			SELECT id, login, password_hash, created_at, updated_at
			FROM users
			WHERE id = $1
		`

		return q.QueryRow(ctx, query, id).Scan(
			&u.ID,
			&u.Login,
			&u.PasswordHash,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	var u entity.User

	err := DoWithRetry(ctx, r.transactor.RetryConfig(), func() error {
		q := r.transactor.GetQuerier(ctx)

		query := `
			SELECT id, login, password_hash, created_at, updated_at
			FROM users
			WHERE login = $1
		`

		return q.QueryRow(ctx, query, login).Scan(
			&u.ID,
			&u.Login,
			&u.PasswordHash,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}
