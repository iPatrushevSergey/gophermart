package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBCryptHasher(t *testing.T) {
	h := NewBCryptHasher(4)

	t.Run("hash and compare", func(t *testing.T) {
		hash, err := h.Hash("password123")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, "password123", hash)

		assert.True(t, h.Compare("password123", hash))
	})

	t.Run("wrong password", func(t *testing.T) {
		hash, err := h.Hash("correct")
		require.NoError(t, err)
		assert.False(t, h.Compare("wrong", hash))
	})

	t.Run("empty password", func(t *testing.T) {
		hash, err := h.Hash("")
		require.NoError(t, err)
		assert.True(t, h.Compare("", hash))
		assert.False(t, h.Compare("notempty", hash))
	})
}
