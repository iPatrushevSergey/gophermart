package model

import "time"

// BalanceAccount is the DB projection of the balance_accounts table row.
type BalanceAccount struct {
	UserID         int64
	Current        float64
	WithdrawnTotal float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int64
}
