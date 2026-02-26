package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid 11-digit", "12345678903", true},
		{"valid 11-digit alt", "49927398716", true},
		{"valid 10-digit", "2377225624", true},
		{"valid 16-digit", "4539578763621486", true},
		{"invalid checksum", "12345678901", false},
		{"single digit", "0", false},
		{"empty string", "", false},
		{"non-digits", "12a45", false},
		{"spaces", "1234 5678", false},
		{"all zeros", "00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Valid(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
