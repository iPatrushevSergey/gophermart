package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config holds the grouped application configuration.
type Config struct {
	Server  ServerConfig
	DB      DBConfig
	Auth    AuthConfig
	Accrual AccrualConfig
	Log     LogConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Address string
}

// DBConfig holds database connection pool settings.
type DBConfig struct {
	URI         string
	MaxConns    int32
	MinConns    int32
	MaxConnLife time.Duration
	MaxConnIdle time.Duration
	HealthCheck time.Duration
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	JWTSecret  string
	JWTTTL     time.Duration
	BCryptCost int
}

// AccrualConfig holds accrual system settings.
type AccrualConfig struct {
	Address string
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string
}

type internalConfig struct {
	ServerAddress  Address    `env:"RUN_ADDRESS"`
	AccrualAddress Address    `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI    string     `env:"DATABASE_URI"`
	JWTSecret      string     `env:"JWT_SECRET"`
	JWTTTL         Duration   `env:"JWT_TTL"`
	LogLevel       string     `env:"LOG_LEVEL"`
	BCryptCost     BCryptCost `env:"BCRYPT_COST"`
	DBMaxConns     int32      `env:"DB_MAX_CONNS"`
	DBMinConns     int32      `env:"DB_MIN_CONNS"`
	DBMaxConnLife  Duration   `env:"DB_MAX_CONN_LIFE"`
	DBMaxConnIdle  Duration   `env:"DB_MAX_CONN_IDLE"`
	DBHealthCheck  Duration   `env:"DB_HEALTH_CHECK"`
}

// LoadConfig loads config from flags and environment (env overrides flags).
func LoadConfig() (Config, error) {
	cfg := internalConfig{}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)

	// Defaults
	cfg.ServerAddress = Address{Host: "127.0.0.1", Port: 8080}
	cfg.AccrualAddress = Address{Schema: "http", Host: "127.0.0.1", Port: 8081}
	cfg.JWTTTL = Duration{Duration: 24 * time.Hour}
	cfg.BCryptCost = 10
	cfg.DBMaxConns = 25
	cfg.DBMinConns = 5
	cfg.DBMaxConnLife = Duration{Duration: time.Hour}
	cfg.DBMaxConnIdle = Duration{Duration: 30 * time.Minute}
	cfg.DBHealthCheck = Duration{Duration: time.Minute}

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
		Server: ServerConfig{
			Address: cfg.ServerAddress.String(),
		},
		DB: DBConfig{
			URI:         cfg.DatabaseURI,
			MaxConns:    cfg.DBMaxConns,
			MinConns:    cfg.DBMinConns,
			MaxConnLife: cfg.DBMaxConnLife.Duration,
			MaxConnIdle: cfg.DBMaxConnIdle.Duration,
			HealthCheck: cfg.DBHealthCheck.Duration,
		},
		Auth: AuthConfig{
			JWTSecret:  cfg.JWTSecret,
			JWTTTL:     cfg.JWTTTL.Duration,
			BCryptCost: int(cfg.BCryptCost),
		},
		Accrual: AccrualConfig{
			Address: cfg.AccrualAddress.URL(),
		},
		Log: LogConfig{
			Level: cfg.LogLevel,
		},
	}, nil
}
