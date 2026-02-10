package bootstrap

import (
	"fmt"

	"gophermart/internal/gophermart/adapters/logger"
	"gophermart/internal/gophermart/config"
)

// Run loads config, initializes logger and app, starts server and waits for graceful shutdown.
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

	app := NewApp(cfg, log)
	StartServer(app.Server, log)
	return WaitForShutdown(app.Server, log)
}
