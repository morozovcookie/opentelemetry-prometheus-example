package prometheus

import (
	"time"
)

func trackOfTime(fn func()) (time.Time, time.Time, time.Duration) {
	start := time.Now()

	fn()

	end := time.Now()

	return start, end, end.Sub(start).Round(time.Millisecond)
}
