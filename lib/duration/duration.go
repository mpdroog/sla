package duration

import (
	"time"
)

// Convert duration to milliseconds
func MilliSeconds(d time.Duration) float64 {
	sec := d / time.Millisecond
	nsec := d % time.Millisecond
	return float64(sec) + float64(nsec)*1e-3
}
