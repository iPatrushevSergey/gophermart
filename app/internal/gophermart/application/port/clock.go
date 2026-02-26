package port

import "time"

// Clock provides the current time. Allows injecting a fake in tests.
type Clock interface {
	Now() time.Time
}
