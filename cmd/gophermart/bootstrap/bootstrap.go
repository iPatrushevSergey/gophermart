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

	log, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer log.Sync()

	log.Debug("starting server",
		"address", cfg.ServerAddress,
		"accrual_address", cfg.AccrualAddress,
		"jwt_ttl", cfg.JWTTTL,
		"log_level", cfg.LogLevel,
		"bcrypt_cost", cfg.BCryptCost,
	)

	app := NewApp(cfg, log)
	StartServer(app.Server, log)
	return WaitForShutdown(app.Server, log)
}
