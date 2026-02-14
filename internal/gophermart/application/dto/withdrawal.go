package dto

import "time"

// WithdrawalOutput is the output for a single withdrawal record.
type WithdrawalOutput struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}
