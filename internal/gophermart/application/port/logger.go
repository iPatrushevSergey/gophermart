package port

// Logger provides structured logging. Args are key-value pairs (e.g. "error", err).
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Sync() error
}
