package model

import (
	"testing"
	"time"
)

func TestActiveDurationLiveLegacy(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	started := now.Add(-5 * time.Minute)
	sec, paused := ActiveDuration(nil, started, started, now)
	if paused || sec != 300 {
		t.Fatalf("sec=%d paused=%v", sec, paused)
	}
}

func TestActiveDurationIdleLegacyFreezes(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	started := now.Add(-12 * time.Hour)
	last := now.Add(-8 * time.Hour)
	sec, paused := ActiveDuration(nil, started, last, now)
	if !paused {
		t.Fatal("expected pause")
	}
	if sec != int64((4 * time.Hour).Seconds()) {
		t.Fatalf("sec=%d", sec)
	}
}

func TestActiveDurationWindowsSkipOvernight(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	windows := []ActivityWindow{
		{StartedAt: now.Add(-20 * time.Hour), EndedAt: now.Add(-18 * time.Hour)},
		{StartedAt: now.Add(-4 * time.Hour), EndedAt: now.Add(-3 * time.Hour)},
	}
	sec, paused := ActiveDuration(windows, windows[0].StartedAt, windows[1].EndedAt, now)
	if !paused {
		t.Fatal("expected pause")
	}
	if sec != int64((3 * time.Hour).Seconds()) {
		t.Fatalf("sec=%d", sec)
	}
}

func TestActiveDurationLiveTail(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	windows := []ActivityWindow{
		{StartedAt: now.Add(-10 * time.Minute), EndedAt: now.Add(-2 * time.Minute)},
	}
	sec, paused := ActiveDuration(windows, windows[0].StartedAt, windows[0].EndedAt, now)
	if paused {
		t.Fatal("expected running clock")
	}
	if sec != int64((10 * time.Minute).Seconds()) {
		t.Fatalf("sec=%d", sec)
	}
}

func TestSpendWindowsDoesNotExtendIdleTail(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	windows := []ActivityWindow{{
		StartedAt: now.Add(-15 * time.Minute),
		EndedAt:   now.Add(-10 * time.Minute),
	}}
	spend := SpendWindows(windows, windows[0].StartedAt, windows[0].EndedAt, windows[0].EndedAt, now)
	if len(spend) != 1 || !spend[0].EndedAt.Equal(windows[0].EndedAt) {
		t.Fatalf("spend windows=%+v", spend)
	}
	clock := SlotWindows(windows, windows[0].StartedAt, windows[0].EndedAt, windows[0].EndedAt, now)
	if len(clock) != 1 || !clock[0].EndedAt.Equal(now) {
		t.Fatalf("clock windows=%+v", clock)
	}
}

func TestSumUsageInWindowsUsesOnlyWindowTicks(t *testing.T) {
	loc := time.FixedZone("test", 0)
	start := time.Date(2026, 9, 14, 10, 0, 0, 0, loc)
	price := 60.0
	samples := []UsageSample{
		{TS: start.Unix(), CursorModelsPct: 10, OtherModelsPct: 20, PlanPriceUSD: &price, Plan: "pro_plus"},
		{TS: start.Add(20 * time.Minute).Unix(), CursorModelsPct: 12, OtherModelsPct: 20, PlanPriceUSD: &price, Plan: "pro_plus"},
		{TS: start.Add(80 * time.Minute).Unix(), CursorModelsPct: 22, OtherModelsPct: 20, PlanPriceUSD: &price, Plan: "pro_plus"},
	}
	delta := SumUsageInWindows(samples, []ActivityWindow{{
		StartedAt: start.Add(10 * time.Minute),
		EndedAt:   start.Add(30 * time.Minute),
	}})
	if delta == nil {
		t.Fatal("expected windowed delta")
	}
	if delta.CursorModelsPct != 2 {
		t.Fatalf("cursor=%v", delta.CursorModelsPct)
	}
	if delta.BudgetUSD != 0.6 {
		t.Fatalf("budget=%v", delta.BudgetUSD)
	}
}
