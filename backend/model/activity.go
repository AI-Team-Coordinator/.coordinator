package model

import "time"

// IdleTimeout is how long a task clock keeps running after the last activity ping.
const IdleTimeout = 20 * time.Minute

// ActivityWindow is one stretch of work on a task. Gaps longer than IdleTimeout start a new window.
type ActivityWindow struct {
	StartedAt time.Time
	EndedAt   time.Time
}

// ActiveDuration is effort time, not wall-clock from started_at.
// Windows, if present, are summed; a fresh last ping extends the tail to now.
// Without windows (legacy slots) the span is started→last, frozen when last is older than IdleTimeout.
func ActiveDuration(windows []ActivityWindow, started, last, now time.Time) (seconds int64, paused bool) {
	if now.IsZero() {
		now = time.Now()
	}
	if len(windows) > 0 {
		var sum int64
		for _, w := range windows {
			end := w.EndedAt
			if end.IsZero() {
				end = w.StartedAt
			}
			if w.StartedAt.IsZero() || end.Before(w.StartedAt) {
				continue
			}
			sum += int64(end.Sub(w.StartedAt).Seconds())
		}
		tail := windows[len(windows)-1].EndedAt
		if tail.IsZero() {
			tail = windows[len(windows)-1].StartedAt
		}
		if tail.IsZero() {
			return clampSec(sum), true
		}
		if now.Sub(tail) <= IdleTimeout {
			sum += int64(now.Sub(tail).Seconds())
			return clampSec(sum), false
		}
		return clampSec(sum), true
	}

	if started.IsZero() {
		started = last
	}
	if last.IsZero() {
		last = started
	}
	if started.IsZero() {
		return 0, true
	}
	if now.Sub(last) > IdleTimeout {
		return clampSec(int64(last.Sub(started).Seconds())), true
	}
	return clampSec(int64(now.Sub(started).Seconds())), false
}

func clampSec(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}
