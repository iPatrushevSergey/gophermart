package vo

// OrderNumberValidator validates order numbers in balance context.
type OrderNumberValidator interface {
	Valid(s string) bool
}
