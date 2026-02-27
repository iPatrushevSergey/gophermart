package convext

import "time"

func CopyTime(v time.Time) time.Time {
	return v
}

func CopyTimePtr(v *time.Time) *time.Time {
	return v
}
