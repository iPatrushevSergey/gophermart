package vo

import "errors"

var (
	// ErrInvalidOrderNumber — The order number was not verified.
	ErrInvalidOrderNumber = errors.New("invalid order number: luhn check failed")
)
