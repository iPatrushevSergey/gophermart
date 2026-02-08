package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config holds the application configuration.
type Config struct {
	ServerAddress  string
	AccrualAddress string
	DatabaseURI    string
	JWTSecret      string
	JWTTTL         time.Duration
	LogLevel       string
	BCryptCost     int
}

type internalConfig struct {
	ServerAddress  Address    `env:"RUN_ADDRESS"`
	AccrualAddress Address    `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI    string     `env:"DATABASE_URI"`
	JWTSecret      string     `env:"JWT_SECRET"`
	JWTTTL         Duration   `env:"JWT_TTL"`
	LogLevel       string     `env:"LOG_LEVEL"`
	BCryptCost     BCryptCost `env:"BCRYPT_COST"`
}

// Load loads config from flags and environment (env overrides flags).
func LoadConfig() (Config, error) {
	cfg := internalConfig{}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)

	// Defaults
	cfg.ServerAddress = Address{Host: "127.0.0.1", Port: 8080}
	cfg.AccrualAddress = Address{Schema: "http", Host: "127.0.0.1", Port: 8081}
	cfg.JWTTTL = Duration{Duration: 24 * time.Hour}
	cfg.BCryptCost = 10

	fs.Var(&cfg.ServerAddress, "a", "server address (host:port)")
	fs.Var(&cfg.AccrualAddress, "r", "address of the accrual calculation system")
	fs.StringVar(
		&cfg.DatabaseURI, "d", "",
		"database dsn, example: postgres://user:password@localhost:5432/db?sslmode=disable",
	)
	fs.StringVar(&cfg.JWTSecret, "s", "dev-secret-change-in-prod", "JWT signing secret")
	fs.Var(&cfg.JWTTTL, "t", "JWT token TTL (e.g. 24h or 86400)")
	fs.StringVar(&cfg.LogLevel, "l", "info", "logging level")
	fs.Var(&cfg.BCryptCost, "bcrypt-cost", "bcrypt cost factor (4-31)")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return Config{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	return Config{
		ServerAddress:  cfg.ServerAddress.String(),
		AccrualAddress: cfg.AccrualAddress.URL(),
		DatabaseURI:    cfg.DatabaseURI,
		JWTSecret:      cfg.JWTSecret,
		JWTTTL:         cfg.JWTTTL.Duration,
		LogLevel:       cfg.LogLevel,
		BCryptCost:     int(cfg.BCryptCost),
	}, nil
}
