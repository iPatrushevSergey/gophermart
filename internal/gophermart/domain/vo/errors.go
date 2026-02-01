package vo

import "errors"

var (
	// ErrInvalidOrderNumber — order number failed Luhn check.
	ErrInvalidOrderNumber = errors.New("invalid order number: luhn check failed")
)
