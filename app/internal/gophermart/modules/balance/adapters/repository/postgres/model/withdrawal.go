package model

import "time"

// Withdrawal is the DB projection of the withdrawals table row.
type Withdrawal struct {
	UserID      int64
	OrderNumber string
	Amount      float64
	ProcessedAt time.Time
}
