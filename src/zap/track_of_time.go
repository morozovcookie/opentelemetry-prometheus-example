package zap

import (
	"time"
)

func trackOfTime(fn func()) (time.Time, time.Time, time.Duration) {
	start := time.Now().UTC()
	fn()
	end := time.Now().UTC()

	return start, end, end.Sub(start).Round(time.Millisecond)
}
