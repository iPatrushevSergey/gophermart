package logger

import "gophermart/internal/gophermart/application/port"

// NopLogger is a no-op logger for tests.
type NopLogger struct{}

// NewNopLogger returns a logger that discards all output.
func NewNopLogger() port.Logger {
	return &NopLogger{}
}

func (NopLogger) Debug(msg string, args ...any) {}
func (NopLogger) Info(msg string, args ...any)  {}
func (NopLogger) Warn(msg string, args ...any)  {}
func (NopLogger) Error(msg string, args ...any) {}
func (NopLogger) Sync() error                   { return nil }
