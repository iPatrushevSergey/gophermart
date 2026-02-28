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

	log, err := logger.Initialize(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync()

	log.Debug("starting server",
		"address", cfg.Server.Address,
		"accrual_address", cfg.Accrual.Client.Address,
		"database_configured", cfg.DB.Pool.URI != "",
		"db_max_conns", cfg.DB.Pool.MaxConns,
		"db_min_conns", cfg.DB.Pool.MinConns,
		"db_retry_max_retries", cfg.DB.Retry.MaxRetries,
		"jwt_ttl", cfg.Auth.JWTTTL,
		"jwt_secret_configured", cfg.Auth.JWTSecret != "",
		"log_level", cfg.Logger.Level,
		"bcrypt_cost", cfg.Auth.BCryptCost,
		"accrual_poll_interval", cfg.Accrual.PollInterval,
		"accrual_batch_size", cfg.Accrual.BatchSize,
		"accrual_max_workers", cfg.Accrual.MaxWorkers,
		"optimistic_retries", cfg.OptimisticRetries,
	)

	// Database
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DB.Pool)
	if err != nil {
		return fmt.Errorf("init database pool: %w", err)
	}
	defer pool.Close()

	transactor := postgres.NewTransactor(pool,
		postgres.WithMaxRetries(cfg.DB.Retry.MaxRetries),
		postgres.WithExponentialBackoff(cfg.DB.Retry.BaseDelay, cfg.DB.Retry.MaxDelay),
	)

	app := NewApp(cfg, log, transactor)

	// Start accrual background worker
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	app.AccrualWorker.Start(workerCtx)

	StartServer(app.Server, log)
	return WaitForShutdown(app.Server, cfg.Server.ShutdownTimeout, log)
}
