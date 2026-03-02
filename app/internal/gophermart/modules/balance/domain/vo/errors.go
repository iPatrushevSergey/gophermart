package vo

import "errors"

var (
	// ErrInvalidOrderNumber is returned when order number fails validation.
	ErrInvalidOrderNumber = errors.New("invalid order number: luhn check failed")
)
