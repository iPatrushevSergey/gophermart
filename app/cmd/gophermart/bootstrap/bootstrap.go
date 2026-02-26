package bootstrap

import (
	"context"
	"fmt"

	"gophermart/internal/gophermart/adapters/logger"
	"gophermart/internal/gophermart/adapters/repository/postgres"
	"gophermart/internal/gophermart/config"
)

// Run loads config, initializes logger and DB, starts app, and waits for graceful shutdown.
func Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.Initialize(cfg.Log.Level)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync()

	log.Debug("starting server",
		"address", cfg.Server.Address,
		"accrual_address", cfg.Accrual.Address,
		"database", cfg.DB.URI,
		"jwt_ttl", cfg.Auth.JWTTTL,
		"log_level", cfg.Log.Level,
		"bcrypt_cost", cfg.Auth.BCryptCost,
	)

	// Database
	ctx := context.Background()

	pool, err := NewPool(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("init database pool: %w", err)
	}
	defer pool.Close()

	transactor := postgres.NewTransactor(pool,
		postgres.WithMaxRetries(cfg.Retry.MaxRetries),
		postgres.WithExponentialBackoff(cfg.Retry.BaseDelay, cfg.Retry.MaxDelay),
	)

	app := NewApp(cfg, log, transactor)

	// Start accrual background worker
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	app.AccrualWorker.Start(workerCtx)

	StartServer(app.Server, log)
	return WaitForShutdown(app.Server, cfg.Server.ShutdownTimeout, log)
}
