package application

import "errors"

// WithOptimisticRetry retries fn when it returns ErrOptimisticLock.
// The entire fn (including reads) is re-executed on each attempt,
// allowing fresh data to be loaded.
func WithOptimisticRetry(maxRetries int, fn func() error) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil || !errors.Is(err, ErrOptimisticLock) {
			return err
		}
	}
	return err
}
