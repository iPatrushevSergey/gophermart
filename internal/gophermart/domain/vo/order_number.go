package vo

import "gophermart/internal/shared/luhn"

// Order number.
type OrderNumber string

// NewOrderNumber parses s as OrderNumber; returns error if Luhn check fails.
func NewOrderNumber(s string) (OrderNumber, error) {
	if !luhn.Valid(s) {
		return "", ErrInvalidOrderNumber
	}
	return OrderNumber(s), nil
}

// String returns the order number as string.
func (n OrderNumber) String() string {
	return string(n)
}
