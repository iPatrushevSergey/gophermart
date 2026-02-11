package postgres

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// RetryConfig holds exponential backoff settings for database operations.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DoWithRetry executes the operation and retries on retriable PostgreSQL errors
// using exponential backoff with full jitter. Respects context cancellation.
func DoWithRetry(ctx context.Context, cfg RetryConfig, op func() error) error {
	var err error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		err = op()
		if err == nil {
			return nil
		}

		if !IsRetriable(err) {
			return err
		}

		if attempt == cfg.MaxRetries {
			break
		}

		delay := backoff(cfg.BaseDelay, cfg.MaxDelay, attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return err
}

// backoff calculates the delay for the given attempt using exponential backoff with full jitter.
// Formula: random(0, min(maxDelay, baseDelay * 2^attempt))
func backoff(baseDelay, maxDelay time.Duration, attempt int) time.Duration {
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))

	if delay > maxDelay {
		delay = maxDelay
	}

	if delay > 0 {
		delay = time.Duration(rand.Int64N(int64(delay)))
	}

	return delay
}
