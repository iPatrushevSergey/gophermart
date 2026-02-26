package vo

// OrderNumber is the order number value object.
type OrderNumber string

// NewOrderNumber parses s as OrderNumber; returns error if validation fails.
// Caller must inject OrderNumberValidator (e.g. from adapters).
func NewOrderNumber(v OrderNumberValidator, s string) (OrderNumber, error) {
	if v == nil || !v.Valid(s) {
		return "", ErrInvalidOrderNumber
	}
	return OrderNumber(s), nil
}

// String returns the order number as string.
func (n OrderNumber) String() string {
	return string(n)
}
