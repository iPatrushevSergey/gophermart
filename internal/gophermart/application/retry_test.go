package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithOptimisticRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		calls := 0
		err := WithOptimisticRetry(3, func() error {
			calls++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, calls)
	})

	t.Run("success after retry", func(t *testing.T) {
		calls := 0
		err := WithOptimisticRetry(3, func() error {
			calls++
			if calls < 3 {
				return ErrOptimisticLock
			}
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 3, calls)
	})

	t.Run("exhausted retries", func(t *testing.T) {
		calls := 0
		err := WithOptimisticRetry(2, func() error {
			calls++
			return ErrOptimisticLock
		})

		assert.ErrorIs(t, err, ErrOptimisticLock)
		assert.Equal(t, 3, calls) // initial + 2 retries
	})

	t.Run("non-retryable error returned immediately", func(t *testing.T) {
		dbErr := errors.New("db connection lost")
		calls := 0
		err := WithOptimisticRetry(3, func() error {
			calls++
			return dbErr
		})

		assert.ErrorIs(t, err, dbErr)
		assert.Equal(t, 1, calls)
	})
}
