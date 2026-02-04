package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap/zapcore"
)

// Address - this is a custom flag type for the address 'host:port'.
type Address struct {
	Schema string
	Host   string
	Port   int
}

// Set implements the interface flag.Value
func (a *Address) Set(s string) error {
	if !(strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")) {
		s = "http://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	if u.Host == "" {
		return errors.New("host is empty")
	}

	hostName, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	a.Schema = u.Scheme
	a.Host = hostName
	a.Port = port

	return nil
}

// String implements the interface flag.Value
func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

func (a *Address) URL() string {
	return fmt.Sprintf("%s://%s:%d", a.Schema, a.Host, a.Port)
}

/////////////////////////////////////

type Duration struct {
	time.Duration
}

func (d *Duration) Set(s string) error {
	if val, err := strconv.Atoi(s); err == nil {
		d.Duration = time.Duration(val) * time.Second
		return nil
	}
	val, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = val
	return nil
}

func (d *Duration) String() string {
	return d.Duration.String()
}

func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}

// Config holds the application configuration.
type Config struct {
	ServerAddress  string
	AccrualAddress string
	DatabaseURI    string
	JWTSecret      string
	JWTTTL         time.Duration
	LogLevel       string
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.ServerAddress)
	enc.AddString("accrual_address", c.AccrualAddress)
	enc.AddDuration("jwt_ttl", c.JWTTTL)
	enc.AddString("log_level", c.LogLevel)
	return nil
}

type internalConfig struct {
	ServerAddress  Address  `env:"RUN_ADDRESS"`
	AccrualAddress Address  `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI    string   `env:"DATABASE_URI"`
	JWTSecret      string   `env:"KEY"`
	JWTTTL         Duration `env:"KEY"`
	LogLevel       string   `env:"LOG_LEVEL"`
}

// Load loads config from flags and environment (env overrides flags).
func LoadConfig() (Config, error) {
	cfg := internalConfig{}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)

	// Default
	cfg.ServerAddress = Address{Host: "127.0.0.1", Port: 8080}
	cfg.AccrualAddress = Address{Schema: "http", Host: "127.0.0.1", Port: 8081}
	cfg.JWTTTL = Duration{Duration: 24 * time.Hour}

	fs.Var(&cfg.ServerAddress, "a", "server address (host:port)")
	fs.Var(&cfg.AccrualAddress, "r", "address of the accrual calculation system")
	fs.StringVar(
		&cfg.DatabaseURI, "d", "",
		"database dsn, example: postgres://user:password@localhost:5432/db?sslmode=disable",
	)
	fs.StringVar(&cfg.JWTSecret, "s", "dev-secret-change-in-prod", "JWT signing secret")
	fs.Var(&cfg.JWTTTL, "t", "JWT token TTL (e.g. 24h or 86400)")
	fs.StringVar(&cfg.LogLevel, "l", "info", "logging level")

	// Flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		return Config{}, fmt.Errorf("flag parsing error: %w", err)
	}

	// Env
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("ENV parsing error: %w", err)
	}

	finalCfg := Config{
		ServerAddress:  cfg.ServerAddress.String(),
		AccrualAddress: cfg.AccrualAddress.URL(),
		DatabaseURI:    cfg.DatabaseURI,
		JWTSecret:      cfg.JWTSecret,
		JWTTTL:         cfg.JWTTTL.Duration,
		LogLevel:       cfg.LogLevel,
	}

	return finalCfg, nil
}

func getEnvOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
