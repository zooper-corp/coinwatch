package tools

import "time"

func AbsDuration(duration time.Duration) time.Duration {
	if duration > 0 {
		return duration
	}
	return -duration
}
