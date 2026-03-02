package accrual

import "time"

// Config defines HTTP client settings for accrual adapter.
type Config struct {
	Address     string
	HTTPTimeout time.Duration
}
