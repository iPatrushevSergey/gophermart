package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddress_Set(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSchema string
		wantHost   string
		wantPort   int
		wantErr    bool
	}{
		{"host:port", "localhost:8080", "http", "localhost", 8080, false},
		{"with http", "http://example.com:443", "http", "example.com", 443, false},
		{"with https", "https://example.com:443", "https", "example.com", 443, false},
		{"no port", "localhost", "", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Address
			err := a.Set(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantSchema, a.Schema)
			assert.Equal(t, tt.wantHost, a.Host)
			assert.Equal(t, tt.wantPort, a.Port)
		})
	}
}

func TestAddress_StringAndURL(t *testing.T) {
	a := Address{Schema: "http", Host: "localhost", Port: 8080}

	assert.Equal(t, "localhost:8080", a.String())
	assert.Equal(t, "http://localhost:8080", a.URL())
}

func TestDuration_Set(t *testing.T) {
	t.Run("integer seconds", func(t *testing.T) {
		var d Duration
		err := d.Set("60")
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, d.Duration)
	})

	t.Run("duration string", func(t *testing.T) {
		var d Duration
		err := d.Set("2h30m")
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Hour+30*time.Minute, d.Duration)
	})

	t.Run("invalid", func(t *testing.T) {
		var d Duration
		err := d.Set("abc")
		assert.Error(t, err)
	})
}

func TestBCryptCost_Set(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var b BCryptCost
		err := b.Set("10")
		assert.NoError(t, err)
		assert.Equal(t, BCryptCost(10), b)
	})

	t.Run("too low", func(t *testing.T) {
		var b BCryptCost
		err := b.Set("2")
		assert.Error(t, err)
	})

	t.Run("too high", func(t *testing.T) {
		var b BCryptCost
		err := b.Set("32")
		assert.Error(t, err)
	})

	t.Run("not a number", func(t *testing.T) {
		var b BCryptCost
		err := b.Set("abc")
		assert.Error(t, err)
	})
}
