package logger

import (
	"fmt"

	"go.uber.org/zap"

	"gophermart/internal/gophermart/application/port"
)

// ZapLogger implements port.Logger using zap.
type ZapLogger struct {
	zl *zap.Logger
}

// NewZapLogger creates a Logger from zap.Logger.
func NewZapLogger(zl *zap.Logger) port.Logger {
	return &ZapLogger{zl: zl}
}

// Initialize creates a zap.Logger with the given level and returns a port.Logger.
func Initialize(level string) (port.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return NewZapLogger(zl), nil
}

func (z *ZapLogger) Debug(msg string, args ...any) {
	z.zl.Debug(msg, toZapFields(args)...)
}

func (z *ZapLogger) Info(msg string, args ...any) {
	z.zl.Info(msg, toZapFields(args)...)
}

func (z *ZapLogger) Warn(msg string, args ...any) {
	z.zl.Warn(msg, toZapFields(args)...)
}

func (z *ZapLogger) Error(msg string, args ...any) {
	z.zl.Error(msg, toZapFields(args)...)
}

// Sync flushes buffered logs.
func (z *ZapLogger) Sync() error {
	return z.zl.Sync()
}

func toZapFields(args []any) []zap.Field {
	if len(args) == 0 {
		return nil
	}
	fields := make([]zap.Field, 0, len(args)/2+1)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("key%d", i/2)
		}
		val := args[i+1]
		fields = append(fields, toZapField(key, val))
	}
	return fields
}

func toZapField(key string, val any) zap.Field {
	switch v := val.(type) {
	case error:
		return zap.Error(v)
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case bool:
		return zap.Bool(key, v)
	default:
		return zap.Any(key, val)
	}
}
