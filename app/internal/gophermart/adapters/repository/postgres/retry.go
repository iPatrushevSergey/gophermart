package postgres

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// BackoffFunc calculates the delay for the given attempt.
type BackoffFunc func(attempt int) time.Duration

// RetriableFunc determines whether the error is retriable.
type RetriableFunc func(error) bool

// retryConfig holds internal retry settings, configured via RetryOption.
type retryConfig struct {
	maxRetries  int
	backoff     BackoffFunc
	isRetriable RetriableFunc
}

// RetryOption configures retry behaviour.
type RetryOption func(*retryConfig)

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) RetryOption {
	return func(c *retryConfig) { c.maxRetries = n }
}

// WithExponentialBackoff sets exponential backoff with full jitter.
func WithExponentialBackoff(base, max time.Duration) RetryOption {
	return func(c *retryConfig) {
		c.backoff = func(attempt int) time.Duration {
			delay := time.Duration(float64(base) * math.Pow(2, float64(attempt)))
			if delay > max {
				delay = max
			}
			if delay > 0 {
				delay = time.Duration(rand.Int64N(int64(delay)))
			}
			return delay
		}
	}
}

// WithConstantBackoff sets a fixed delay between retries.
func WithConstantBackoff(d time.Duration) RetryOption {
	return func(c *retryConfig) {
		c.backoff = func(_ int) time.Duration { return d }
	}
}

// WithBackoffFunc sets a custom backoff strategy.
func WithBackoffFunc(fn BackoffFunc) RetryOption {
	return func(c *retryConfig) { c.backoff = fn }
}

// WithRetriableCheck overrides the function that decides if an error is retriable.
func WithRetriableCheck(fn RetriableFunc) RetryOption {
	return func(c *retryConfig) { c.isRetriable = fn }
}

// DoWithRetry executes the operation and retries on retriable errors.
// Respects context cancellation between attempts.
func DoWithRetry(ctx context.Context, op func() error, opts ...RetryOption) error {
	cfg := retryConfig{
		maxRetries:  3,
		backoff:     func(_ int) time.Duration { return 100 * time.Millisecond },
		isRetriable: IsRetriable,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var err error
	for attempt := 0; attempt <= cfg.maxRetries; attempt++ {
		err = op()
		if err == nil {
			return nil
		}

		if !cfg.isRetriable(err) {
			return err
		}

		if attempt == cfg.maxRetries {
			break
		}

		delay := cfg.backoff(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return err
}
