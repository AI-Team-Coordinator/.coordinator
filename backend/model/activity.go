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

// ExtendTailIfLive stretches the last window to now while the clock is still running.
func ExtendTailIfLive(windows []ActivityWindow, now time.Time) []ActivityWindow {
	if len(windows) == 0 {
		return windows
	}
	if now.IsZero() {
		now = time.Now()
	}
	out := append([]ActivityWindow(nil), windows...)
	last := &out[len(out)-1]
	end := last.EndedAt
	if end.IsZero() {
		end = last.StartedAt
	}
	if end.IsZero() {
		return out
	}
	if now.Sub(end) <= IdleTimeout {
		last.EndedAt = now
	}
	return out
}

// SlotWindows returns recorded activity windows, or a legacy started→last span.
func SlotWindows(windows []ActivityWindow, started, last, updated, now time.Time) []ActivityWindow {
	if len(windows) > 0 {
		return ExtendTailIfLive(windows, now)
	}
	if started.IsZero() {
		return nil
	}
	end := last
	if end.IsZero() {
		end = updated
	}
	if end.IsZero() {
		end = now
	}
	if end.IsZero() {
		end = time.Now()
	}
	return ExtendTailIfLive([]ActivityWindow{{StartedAt: started, EndedAt: end}}, now)
}

// ResearchWindows uses only recorded pings. An empty list means no overlap.
func ResearchWindows(windows []ActivityWindow, now time.Time) []ActivityWindow {
	if len(windows) == 0 {
		return nil
	}
	return ExtendTailIfLive(windows, now)
}

func windowSpan(w ActivityWindow) (time.Time, time.Time, bool) {
	start, end := w.StartedAt, w.EndedAt
	if end.IsZero() {
		end = start
	}
	if start.IsZero() || end.Before(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

// WindowsOverlap is true when two activity timelines share any instant.
func WindowsOverlap(a, b []ActivityWindow) bool {
	for _, left := range a {
		ls, le, ok := windowSpan(left)
		if !ok {
			continue
		}
		for _, right := range b {
			rs, re, ok := windowSpan(right)
			if !ok {
				continue
			}
			if ls.Before(re) && rs.Before(le) || ls.Equal(rs) {
				return true
			}
		}
	}
	return false
}

// WindowOverlapsRange is true when a window intersects [start, end).
func WindowOverlapsRange(w ActivityWindow, start, end time.Time) bool {
	ls, le, ok := windowSpan(w)
	if !ok {
		return false
	}
	return ls.Before(end) && !le.Before(start)
}
