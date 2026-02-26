package dto

//go:generate easyjson -all $GOFILE

// OrderResponse is the HTTP response body for a single order.
type OrderResponse struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}
