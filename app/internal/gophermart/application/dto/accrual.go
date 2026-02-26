package dto

// AccrualOrderInfo holds the response from the accrual system for a single order.
type AccrualOrderInfo struct {
	Status  string
	Accrual *float64
}
