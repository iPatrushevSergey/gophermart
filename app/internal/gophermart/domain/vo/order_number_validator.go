package vo

// OrderNumberValidator validates order numbers (e.g. Luhn check).
// Implemented by adapters; domain defines the port.
type OrderNumberValidator interface {
	Valid(s string) bool
}
