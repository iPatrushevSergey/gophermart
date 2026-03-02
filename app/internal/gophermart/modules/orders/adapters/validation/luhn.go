package validation

import (
	"gophermart/internal/pkg/luhn"
)

// LuhnValidator validates using the Luhn algorithm.
type LuhnValidator struct{}

// NewLuhnValidator returns a validator that satisfies vo.OrderNumberValidator.
func NewLuhnValidator() *LuhnValidator {
	return &LuhnValidator{}
}

// Valid returns true if s passes the Luhn check.
func (LuhnValidator) Valid(s string) bool {
	return luhn.Valid(s)
}
