package vo

// OrderNumber is withdrawal reference in balance context.
type OrderNumber string

// NewOrderNumber validates and returns order number value object.
func NewOrderNumber(v OrderNumberValidator, s string) (OrderNumber, error) {
	if v == nil || !v.Valid(s) {
		return "", ErrInvalidOrderNumber
	}
	return OrderNumber(s), nil
}

// String returns order number as string.
func (n OrderNumber) String() string {
	return string(n)
}
