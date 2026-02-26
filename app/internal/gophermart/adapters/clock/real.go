package clock

import "time"

// Real returns the actual system time.
type Real struct{}

// Now returns time.Now().
func (Real) Now() time.Time {
	return time.Now()
}
