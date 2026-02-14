package dto

//go:generate easyjson -all $GOFILE

// WithdrawalResponse is the HTTP response body for a single withdrawal record.
type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
