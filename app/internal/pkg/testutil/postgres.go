//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	testDBName = "gophermart_test"
	testDBUser = "test"
	testDBPass = "test"
)

// SetupPostgres starts a PostgreSQL container, applies migrations and returns
// a ready-to-use pgxpool.Pool. The container is terminated when t finishes.
func SetupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(ctx), "failed to terminate postgres container")
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	poolCfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	poolCfg.MaxConns = 5
	poolCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	require.NoError(t, err, "failed to create pool")

	t.Cleanup(func() {
		pool.Close()
	})

	applyMigrations(t, dsn, migrationsDir())

	return pool
}

// applyMigrations runs all pending goose migrations from dir.
func applyMigrations(t *testing.T, dsn, dir string) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err, "failed to initialize migrations")
	defer db.Close()

	err = goose.SetDialect("postgres")
	require.NoError(t, err, "failed to set migration dialect")

	err = goose.Up(db, dir)
	require.NoError(t, err, "failed to apply migrations")
}

// migrationsDir returns an absolute path to the migrations directory.
func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations", "gophermart")
}
