package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gophermart/internal/gophermart/adapters/repository/postgres/converter"
	"gophermart/internal/gophermart/adapters/repository/postgres/model"
	"gophermart/internal/gophermart/application"
	"gophermart/internal/gophermart/domain/entity"
	"gophermart/internal/gophermart/domain/vo"
)

type UserRepository struct {
	transactor *Transactor
	conv       converter.UserConverter
}

func NewUserRepository(transactor *Transactor) *UserRepository {
	return &UserRepository{
		transactor: transactor,
		conv:       &converter.UserConverterImpl{},
	}
}

func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		dbUser := r.conv.ToModel(*u)

		query := `
			INSERT INTO users (login, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`

		var dbID int64
		err := q.QueryRow(ctx, query, dbUser.Login, dbUser.PasswordHash, dbUser.CreatedAt, dbUser.UpdatedAt).Scan(&dbID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return application.ErrAlreadyExists
			}
			return err
		}
		u.ID = vo.UserID(dbID)

		return nil
	})
}

func (r *UserRepository) FindByID(ctx context.Context, id vo.UserID) (*entity.User, error) {
	var u entity.User

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		query := `
			SELECT id, login, password_hash, created_at, updated_at
			FROM users
			WHERE id = $1
		`

		rows, err := q.Query(ctx, query, id)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRow, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[model.User])
		if err != nil {
			return err
		}

		u = r.conv.ToEntity(dbRow)
		return nil
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

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		query := `
			SELECT id, login, password_hash, created_at, updated_at
			FROM users
			WHERE login = $1
		`

		rows, err := q.Query(ctx, query, login)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRow, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[model.User])
		if err != nil {
			return err
		}

		u = r.conv.ToEntity(dbRow)
		return nil
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}
