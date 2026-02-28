package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"gophermart/internal/gophermart/adapters/accrual"
	"gophermart/internal/gophermart/adapters/logger"
	"gophermart/internal/gophermart/adapters/repository/postgres"
)

// Config holds the grouped application configuration.
type Config struct {
	Server  ServerConfig
	Auth    AuthConfig
	Logger  logger.Config
	DB      postgres.Config
	Accrual AccrualConfig
	// OptimisticRetries controls use-case retries on optimistic lock conflicts.
	OptimisticRetries int
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Address         string
	ShutdownTimeout time.Duration
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	JWTSecret  string
	JWTTTL     time.Duration
	BCryptCost int
}

// AccrualConfig groups adapter and worker settings for accrual processing.
type AccrualConfig struct {
	Client       accrual.Config
	PollInterval time.Duration
	BatchSize    int
	MaxWorkers   int
}

// LoadConfig loads config from flags/env/file/defaults.
// Priority: flags > env > file > defaults.
func LoadConfig() (Config, error) {
	// --- flags (CLI schema) ---
	fs := pflag.NewFlagSet("gophermart", pflag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringP("address", "a", "", "server address (host:port)")
	fs.StringP("database-uri", "d", "", "database dsn")
	fs.StringP("accrual-address", "r", "", "accrual system address")
	fs.StringP("jwt-secret", "s", "", "JWT signing secret")
	fs.StringP("jwt-ttl", "t", "", "JWT token TTL (e.g. 24h or 86400)")
	fs.StringP("log-level", "l", "", "logging level")
	fs.Int("bcrypt-cost", 0, "bcrypt cost factor (4-31)")
	fs.String("config", "app/configs/gophermart.yaml", "path to YAML config")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return Config{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// --- dotenv (local dev helper) ---
	dotenvPath, dotenvLoaded, err := loadDotEnv()
	if err != nil {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	if dotenvLoaded {
		_, _ = fmt.Fprintf(os.Stderr, "config: loaded dotenv file %s\n", dotenvPath)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "config: dotenv file not found (checked app/.env, .env), continuing with env/yaml/defaults")
	}

	// --- defaults (baseline) ---
	v := viper.New()
	setDefaults(v)

	// --- config_file (YAML) ---
	configPath, _ := fs.GetString("config")
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			if fs.Changed("config") {
				return Config{}, fmt.Errorf("config file not found: %s", configPath)
			}
			_, _ = fmt.Fprintf(os.Stderr, "config: default config file %s not found, continuing with env/defaults\n", configPath)
		} else {
			return Config{}, fmt.Errorf("read config file: %w", err)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "config: loaded config file %s\n", v.ConfigFileUsed())
	}

	// --- env binding ---
	bindEnv(v)

	// --- flags binding ---
	if err := bindFlags(v, fs); err != nil {
		return Config{}, fmt.Errorf("bind flags: %w", err)
	}

	// --- normalize + validate ---
	serverAddr, err := parseAddress(v.GetString("server.address"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid server address: %w", err)
	}
	accrualURL, err := parseURLAddress(v.GetString("accrual.address"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid accrual address: %w", err)
	}
	jwtTTL, err := parseDuration(v.Get("auth.jwt_ttl"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_TTL: %w", err)
	}
	dbMaxConnLife, err := parseDuration(v.Get("database.max_conn_life"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_MAX_CONN_LIFE: %w", err)
	}
	dbMaxConnIdle, err := parseDuration(v.Get("database.max_conn_idle"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_MAX_CONN_IDLE: %w", err)
	}
	dbHealthCheck, err := parseDuration(v.Get("database.health_check"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_HEALTH_CHECK: %w", err)
	}
	accrualPollInterval, err := parseDuration(v.Get("accrual.poll_interval"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid ACCRUAL_POLL_INTERVAL: %w", err)
	}
	accrualHTTPTimeout, err := parseDuration(v.Get("accrual.http_timeout"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid ACCRUAL_HTTP_TIMEOUT: %w", err)
	}
	retryBaseDelay, err := parseDuration(v.Get("database.retry.base_delay"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_RETRY_BASE_DELAY: %w", err)
	}
	retryMaxDelay, err := parseDuration(v.Get("database.retry.max_delay"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_RETRY_MAX_DELAY: %w", err)
	}
	shutdownTimeout, err := parseDuration(v.Get("server.shutdown_timeout"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SERVER_SHUTDOWN_TIMEOUT: %w", err)
	}
	bcryptCost, err := parseBCryptCost(v.Get("auth.bcrypt_cost"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid BCRYPT_COST: %w", err)
	}
	jwtSecret := strings.TrimSpace(v.GetString("auth.jwt_secret"))
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	databaseURI := strings.TrimSpace(v.GetString("database.uri"))
	if databaseURI == "" {
		return Config{}, fmt.Errorf("DATABASE_URI is required")
	}

	// --- assemble typed config ---
	return Config{
		Server: ServerConfig{
			Address:         serverAddr,
			ShutdownTimeout: shutdownTimeout,
		},
		Auth: AuthConfig{
			JWTSecret:  jwtSecret,
			JWTTTL:     jwtTTL,
			BCryptCost: bcryptCost,
		},
		Logger: logger.Config{
			Level: v.GetString("logger.level"),
		},
		DB: postgres.Config{
			Pool: postgres.PoolConfig{
				URI:         databaseURI,
				MaxConns:    int32(v.GetInt("database.max_conns")),
				MinConns:    int32(v.GetInt("database.min_conns")),
				MaxConnLife: dbMaxConnLife,
				MaxConnIdle: dbMaxConnIdle,
				HealthCheck: dbHealthCheck,
			},
			Retry: postgres.RetryConfig{
				MaxRetries: v.GetInt("database.retry.max_retries"),
				BaseDelay:  retryBaseDelay,
				MaxDelay:   retryMaxDelay,
			},
		},
		Accrual: AccrualConfig{
			Client: accrual.Config{
				Address:     accrualURL,
				HTTPTimeout: accrualHTTPTimeout,
			},
			PollInterval: accrualPollInterval,
			BatchSize:    v.GetInt("accrual.batch_size"),
			MaxWorkers:   v.GetInt("accrual.max_workers"),
		},
		OptimisticRetries: v.GetInt("optimistic_retries"),
	}, nil
}

func loadDotEnv() (string, bool, error) {
	for _, file := range []string{"app/.env", ".env"} {
		if _, err := os.Stat(file); err == nil {
			if err := godotenv.Load(file); err != nil {
				return "", false, fmt.Errorf("failed to load %s: %w", file, err)
			}
			return file, true, nil
		}
	}
	return "", false, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.address", "127.0.0.1:8080")
	v.SetDefault("server.shutdown_timeout", "5s")

	v.SetDefault("database.uri", "")
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("database.max_conn_life", "1h")
	v.SetDefault("database.max_conn_idle", "30m")
	v.SetDefault("database.health_check", "1m")
	v.SetDefault("database.retry.max_retries", 3)
	v.SetDefault("database.retry.base_delay", "100ms")
	v.SetDefault("database.retry.max_delay", "2s")

	v.SetDefault("auth.jwt_secret", "")
	v.SetDefault("auth.jwt_ttl", "24h")
	v.SetDefault("auth.bcrypt_cost", 10)

	v.SetDefault("logger.level", "info")

	v.SetDefault("accrual.address", "127.0.0.1:8081")
	v.SetDefault("accrual.poll_interval", "2s")
	v.SetDefault("accrual.http_timeout", "10s")
	v.SetDefault("accrual.batch_size", 50)
	v.SetDefault("accrual.max_workers", 5)

	v.SetDefault("optimistic_retries", 3)
}

func bindFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	bindings := map[string]string{
		"server.address":   "address",
		"database.uri":     "database-uri",
		"accrual.address":  "accrual-address",
		"auth.jwt_secret":  "jwt-secret",
		"auth.jwt_ttl":     "jwt-ttl",
		"logger.level":     "log-level",
		"auth.bcrypt_cost": "bcrypt-cost",
	}
	for key, flagName := range bindings {
		f := fs.Lookup(flagName)
		if f == nil {
			return fmt.Errorf("flag not found: %s", flagName)
		}
		if err := v.BindPFlag(key, f); err != nil {
			return err
		}
	}
	return nil
}

func bindEnv(v *viper.Viper) {
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.BindEnv("server.address", "RUN_ADDRESS")
	_ = v.BindEnv("database.uri", "DATABASE_URI")
	_ = v.BindEnv("accrual.address", "ACCRUAL_SYSTEM_ADDRESS")
	_ = v.BindEnv("auth.jwt_secret", "JWT_SECRET")
	_ = v.BindEnv("auth.jwt_ttl", "JWT_TTL")
	_ = v.BindEnv("logger.level", "LOG_LEVEL")
	_ = v.BindEnv("auth.bcrypt_cost", "BCRYPT_COST")

	_ = v.BindEnv("database.max_conns", "DB_MAX_CONNS")
	_ = v.BindEnv("database.min_conns", "DB_MIN_CONNS")
	_ = v.BindEnv("database.max_conn_life", "DB_MAX_CONN_LIFE")
	_ = v.BindEnv("database.max_conn_idle", "DB_MAX_CONN_IDLE")
	_ = v.BindEnv("database.health_check", "DB_HEALTH_CHECK")
	_ = v.BindEnv("database.retry.max_retries", "DB_RETRY_MAX_RETRIES")
	_ = v.BindEnv("database.retry.base_delay", "DB_RETRY_BASE_DELAY")
	_ = v.BindEnv("database.retry.max_delay", "DB_RETRY_MAX_DELAY")

	_ = v.BindEnv("accrual.poll_interval", "ACCRUAL_POLL_INTERVAL")
	_ = v.BindEnv("accrual.http_timeout", "ACCRUAL_HTTP_TIMEOUT")
	_ = v.BindEnv("accrual.batch_size", "ACCRUAL_BATCH_SIZE")
	_ = v.BindEnv("accrual.max_workers", "ACCRUAL_MAX_WORKERS")

	_ = v.BindEnv("optimistic_retries", "OPTIMISTIC_RETRIES")
}

func parseAddress(raw string) (string, error) {
	var addr Address
	if err := addr.Set(raw); err != nil {
		return "", err
	}
	return addr.String(), nil
}

func parseURLAddress(raw string) (string, error) {
	var addr Address
	if err := addr.Set(raw); err != nil {
		return "", err
	}
	return addr.URL(), nil
}

func parseDuration(raw any) (time.Duration, error) {
	var d Duration
	switch v := raw.(type) {
	case string:
		if err := d.Set(v); err != nil {
			return 0, err
		}
		return d.Duration, nil
	case int:
		return time.Duration(v) * time.Second, nil
	case int32:
		return time.Duration(v) * time.Second, nil
	case int64:
		return time.Duration(v) * time.Second, nil
	case float64:
		return time.Duration(v) * time.Second, nil
	default:
		return 0, fmt.Errorf("unsupported duration type %T", raw)
	}
}

func parseBCryptCost(raw any) (int, error) {
	var cost BCryptCost
	if err := cost.Set(fmt.Sprint(raw)); err != nil {
		return 0, err
	}
	return int(cost), nil
}
