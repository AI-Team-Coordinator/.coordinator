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
