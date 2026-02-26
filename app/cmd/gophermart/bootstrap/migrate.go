package bootstrap

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending database migrations from the given directory.
func RunMigrations(dsn, sourceDir string) error {
	m, err := migrate.New("file://"+sourceDir, dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
