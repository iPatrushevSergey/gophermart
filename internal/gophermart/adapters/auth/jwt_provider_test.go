package auth

import (
	"testing"
	"time"

	"gophermart/internal/gophermart/domain/vo"

	"github.com/stretchr/testify/assert"
)

func TestJWTProvider_IssueAndValidate(t *testing.T) {
	p := NewJWTProvider("test-secret", time.Hour)

	t.Run("round trip", func(t *testing.T) {
		token, err := p.Issue(vo.UserID(42))
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		id, err := p.Validate(token)
		assert.NoError(t, err)
		assert.Equal(t, vo.UserID(42), id)
	})

	t.Run("expired token", func(t *testing.T) {
		expired := NewJWTProvider("test-secret", -time.Hour)
		token, err := expired.Issue(vo.UserID(1))
		assert.NoError(t, err)

		_, err = p.Validate(token)
		assert.Error(t, err)
	})

	t.Run("invalid token string", func(t *testing.T) {
		_, err := p.Validate("not-a-jwt")
		assert.Error(t, err)
	})

	t.Run("wrong secret", func(t *testing.T) {
		other := NewJWTProvider("other-secret", time.Hour)
		token, err := other.Issue(vo.UserID(1))
		assert.NoError(t, err)

		_, err = p.Validate(token)
		assert.Error(t, err)
	})
}
