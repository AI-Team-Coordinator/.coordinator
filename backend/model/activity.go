package model

import (
	"sort"
	"time"
)

// IdleTimeout is how long a task clock keeps running after the last activity ping.
const IdleTimeout = 20 * time.Minute

// ActivityWindow is one stretch of work on a task. Gaps longer than IdleTimeout start a new window.
type ActivityWindow struct {
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
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
// The live tail is stretched to now for the clock. Spend overlap uses SpendWindows.
func SlotWindows(windows []ActivityWindow, started, last, updated, now time.Time) []ActivityWindow {
	return ExtendTailIfLive(recordedWindows(windows, started, last, updated, now), now)
}

// SpendWindows is recorded activity for quota sharing. It does not extend an
// idle tail to now — two clocks still running after a switch are not parallel work.
func SpendWindows(windows []ActivityWindow, started, last, updated, now time.Time) []ActivityWindow {
	return recordedWindows(windows, started, last, updated, now)
}

// AttributionWindows is SpendWindows plus a live tail on this slot only, so
// usage ticks after the last ping still count toward the active task.
func AttributionWindows(windows []ActivityWindow, started, last, updated, now time.Time) []ActivityWindow {
	return ExtendTailIfLive(recordedWindows(windows, started, last, updated, now), now)
}

func recordedWindows(windows []ActivityWindow, started, last, updated, now time.Time) []ActivityWindow {
	if len(windows) > 0 {
		return append([]ActivityWindow(nil), windows...)
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
	return []ActivityWindow{{StartedAt: started, EndedAt: end}}
}

// ResearchWindows uses only recorded pings. An empty list means no overlap.
func ResearchWindows(windows []ActivityWindow, now time.Time) []ActivityWindow {
	if len(windows) == 0 {
		return nil
	}
	return ExtendTailIfLive(windows, now)
}

// ResearchSpendWindows is recorded research pings without an idle tail.
func ResearchSpendWindows(windows []ActivityWindow) []ActivityWindow {
	if len(windows) == 0 {
		return nil
	}
	return append([]ActivityWindow(nil), windows...)
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

// ClipWindows keeps the parts of windows that intersect [start, end).
func ClipWindows(windows []ActivityWindow, start, end time.Time) []ActivityWindow {
	if len(windows) == 0 || !start.Before(end) {
		return nil
	}
	out := make([]ActivityWindow, 0, len(windows))
	for _, w := range windows {
		ls, le, ok := windowSpan(w)
		if !ok || !ls.Before(end) || le.Before(start) {
			continue
		}
		if ls.Before(start) {
			ls = start
		}
		if le.After(end) {
			le = end
		}
		if !ls.Before(le) && !ls.Equal(le) {
			continue
		}
		out = append(out, ActivityWindow{StartedAt: ls, EndedAt: le})
	}
	return out
}

// SumUsageInWindows adds account-meter deltas whose ticks fall inside windows.
func SumUsageInWindows(samples []UsageSample, windows []ActivityWindow) *UsageDelta {
	if len(samples) == 0 || len(windows) == 0 {
		return nil
	}
	sorted := append([]UsageSample(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TS < sorted[j].TS })
	acc := &UsageDelta{}
	any := false
	for _, w := range windows {
		start, end, ok := windowSpan(w)
		if !ok {
			continue
		}
		hs, he := start.Unix(), end.Unix()
		var before *UsageSample
		var lastIn *UsageSample
		for i := range sorted {
			ts := sorted[i].TS
			if ts <= hs {
				before = &sorted[i]
				continue
			}
			if ts <= he {
				lastIn = &sorted[i]
			}
		}
		if lastIn == nil {
			continue
		}
		startSnap := before
		if startSnap == nil {
			for i := range sorted {
				if sorted[i].TS > hs && sorted[i].TS <= he {
					startSnap = &sorted[i]
					break
				}
			}
		}
		if startSnap == nil {
			continue
		}
		d := ComputeUsageDelta(CursorUsageFromSample(*startSnap), CursorUsageFromSample(*lastIn))
		if d == nil {
			continue
		}
		acc.Add(d)
		any = true
	}
	if !any {
		return nil
	}
	return acc
}

// Add folds another delta into this one.
func (d *UsageDelta) Add(other *UsageDelta) {
	if d == nil || other == nil {
		return
	}
	d.CostUSD = round4(d.CostUSD + other.CostUSD)
	d.BudgetUSD = round4(d.BudgetUSD + other.BudgetUSD)
	d.OnDemandUSD = round4(d.OnDemandUSD + other.OnDemandUSD)
	d.CursorModelsPct = round4(d.CursorModelsPct + other.CursorModelsPct)
	d.OtherModelsPct = round4(d.OtherModelsPct + other.OtherModelsPct)
	if d.UsagePlan == "" {
		d.UsagePlan = other.UsagePlan
	}
	if d.PlanPriceUSD == nil && other.PlanPriceUSD != nil {
		p := *other.PlanPriceUSD
		d.PlanPriceUSD = &p
	}
}
