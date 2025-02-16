package utils

import "time"

// TimeNow returns the current time in UTC
func TimeNow() time.Time {
	return time.Now().UTC()
}
