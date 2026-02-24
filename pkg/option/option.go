package option

// Option is a generic functional option type that can be reused
// across different constructors in the application.
type Option[T any] func(*T)

// Apply applies all options to the target.
func Apply[T any](t *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(t)
	}
}
