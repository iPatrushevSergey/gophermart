package vo

import "gophermart/internal/shared/luhn"

// Order number.
type OrderNumber string

// NewOrderNumber creates OrderNumber only if s passes the luhn check.
// Otherwise, it returns an error.
func NewOrderNumber(s string) (OrderNumber, error) {
	if !luhn.Valid(s) {
		return "", ErrInvalidOrderNumber
	}
	return OrderNumber(s), nil
}

// String returns a string representation of a number.
func (n OrderNumber) String() string {
	return string(n)
}
