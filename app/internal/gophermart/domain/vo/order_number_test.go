package vo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// stubValidator is a simple test stub for OrderNumberValidator.
type stubValidator struct{ valid bool }

func (s stubValidator) Valid(string) bool { return s.valid }

func TestNewOrderNumber(t *testing.T) {
	tests := []struct {
		name      string
		validator OrderNumberValidator
		input     string
		wantNum   OrderNumber
		wantErr   error
	}{
		{"valid number", stubValidator{true}, "12345678903", OrderNumber("12345678903"), nil},
		{"invalid number", stubValidator{false}, "123", "", ErrInvalidOrderNumber},
		{"nil validator", nil, "12345678903", "", ErrInvalidOrderNumber},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOrderNumber(tt.validator, tt.input)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantNum, got)
			}
		})
	}
}

func TestOrderNumber_String(t *testing.T) {
	n := OrderNumber("12345678903")
	assert.Equal(t, "12345678903", n.String())
}
