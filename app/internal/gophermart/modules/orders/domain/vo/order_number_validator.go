package vo

// OrderNumberValidator validates order numbers (e.g. Luhn check).
type OrderNumberValidator interface {
	Valid(s string) bool
}
