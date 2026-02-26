//go:build integration

package testutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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

	applyMigrations(t, pool, migrationsDir())

	return pool
}

// applyMigrations reads all *.up.sql files from dir, sorts them by name
// and executes each one against the pool.
func applyMigrations(t *testing.T, pool *pgxpool.Pool, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err, "failed to read migrations directory: %s", dir)

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" && len(e.Name()) > 7 && e.Name()[len(e.Name())-7:] == ".up.sql" {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	ctx := context.Background()
	for _, f := range upFiles {
		sql, err := os.ReadFile(filepath.Join(dir, f))
		require.NoError(t, err, "failed to read migration file %s", f)
		_, err = pool.Exec(ctx, string(sql))
		require.NoError(t, err, "failed to apply migration %s", f)
	}
}

// migrationsDir returns an absolute path to the migrations directory.
func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations", "gophermart")
}
