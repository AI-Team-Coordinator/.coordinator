package coordinator

import (
	"testing"
	"time"

	"coordinator/model"
)

func TestBuildHourlySpendExclusiveHour(t *testing.T) {
	loc := time.FixedZone("test", 0)
	hour := time.Date(2026, 9, 14, 10, 0, 0, 0, loc)
	now := hour.Add(50 * time.Minute)
	price := 60.0
	samples := []model.UsageSample{
		{TS: hour.Add(-5 * time.Minute).Unix(), CursorModelsPct: 10, OtherModelsPct: 20, PlanPriceUSD: &price, Plan: "pro_plus"},
		{TS: hour.Add(20 * time.Minute).Unix(), CursorModelsPct: 12, OtherModelsPct: 20, PlanPriceUSD: &price, Plan: "pro_plus", Kind: "task", TaskID: "TASK-A", Title: "A", Alias: "EK"},
	}
	hours := BuildHourlySpend(samples, nil, now, 2*time.Hour)
	if len(hours) != 1 {
		t.Fatalf("hours=%d %+v", len(hours), hours)
	}
	got := hours[0]
	if got.Hour != hour.Unix() {
		t.Fatalf("hour=%d want %d", got.Hour, hour.Unix())
	}
	if got.CursorModelsPct != 2 {
		t.Fatalf("cursor=%v", got.CursorModelsPct)
	}
	if got.BudgetUSD == nil || *got.BudgetUSD != 0.6 {
		t.Fatalf("budget=%v", got.BudgetUSD)
	}
	if len(got.Participants) != 1 || got.Participants[0].ID != "TASK-A" || got.Participants[0].Shared {
		t.Fatalf("participants=%+v", got.Participants)
	}
}

func TestBuildHourlySpendSharedFromWindows(t *testing.T) {
	loc := time.FixedZone("test", 0)
	hour := time.Date(2026, 9, 14, 11, 0, 0, 0, loc)
	now := hour.Add(40 * time.Minute)
	samples := []model.UsageSample{
		{TS: hour.Add(-1 * time.Minute).Unix(), CursorModelsPct: 1, OtherModelsPct: 1},
		{TS: hour.Add(10 * time.Minute).Unix(), CursorModelsPct: 3, OtherModelsPct: 1},
	}
	members := []model.Member{{
		Alias: "EK",
		Tasks: []model.MemberTask{{
			TaskID: "TASK-A",
			Title:  "A",
			ActivityWindows: []model.ActivityWindow{{
				StartedAt: hour.Add(5 * time.Minute),
				EndedAt:   hour.Add(25 * time.Minute),
			}},
		}},
		Research: &model.Research{
			Status:  "active",
			Summary: "look",
			ActivityWindows: []model.ActivityWindow{{
				StartedAt: hour.Add(8 * time.Minute),
				EndedAt:   hour.Add(12 * time.Minute),
			}},
		},
	}}
	hours := BuildHourlySpend(samples, members, now, 2*time.Hour)
	if len(hours) != 1 {
		t.Fatalf("hours=%d", len(hours))
	}
	if len(hours[0].Participants) != 2 {
		t.Fatalf("participants=%+v", hours[0].Participants)
	}
	for _, p := range hours[0].Participants {
		if !p.Shared {
			t.Fatalf("expected shared %+v", p)
		}
	}
}

func TestCompletedExclusiveNeedsSoloHours(t *testing.T) {
	hours := []model.HourSpend{{
		Hour: 1,
		Participants: []model.HourParticipant{
			{Kind: "task", ID: "A"},
		},
	}}
	if !completedExclusive("A", hours) {
		t.Fatal("solo hour should be exclusive")
	}
	hours[0].Participants = append(hours[0].Participants, model.HourParticipant{Kind: "research", ID: "r", Shared: true})
	hours[0].Participants[0].Shared = true
	if completedExclusive("A", hours) {
		t.Fatal("shared hour should not be exclusive")
	}
	if completedExclusive("B", hours) {
		t.Fatal("unseen task is not exclusive")
	}
}

func TestApplyCompletedSpendSharingHidesDollars(t *testing.T) {
	tasks := []model.Task{
		{TaskID: "A", Status: "completed", BudgetUSD: f64(4), CostUSD: f64(1), CursorModelsPct: f64(2)},
		{TaskID: "B", Status: "in_progress", BudgetUSD: f64(3)},
	}
	applyCompletedSpendSharing(tasks, nil, nil)
	if !tasks[0].SpendShared || tasks[0].BudgetUSD != nil || tasks[0].CostUSD != nil {
		t.Fatalf("completed without meter should be shared: %+v", tasks[0])
	}
	if tasks[0].CursorModelsPct == nil || *tasks[0].CursorModelsPct != 2 {
		t.Fatalf("keep percents %+v", tasks[0].CursorModelsPct)
	}
	if tasks[1].SpendShared || tasks[1].BudgetUSD == nil {
		t.Fatalf("open task unchanged %+v", tasks[1])
	}
}

func TestBuildDailySpendSumsHoursAndPadsDays(t *testing.T) {
	loc := time.FixedZone("test", 0)
	day := time.Date(2026, 9, 14, 0, 0, 0, 0, loc)
	now := day.Add(15 * time.Hour)
	b := 0.3
	hours := []model.HourSpend{
		{Hour: day.Add(10 * time.Hour).Unix(), Alias: "EK", CursorModelsPct: 1, OtherModelsPct: 2, BudgetUSD: &b, Participants: []model.HourParticipant{{Kind: "task", ID: "A", Alias: "EK"}}},
		{Hour: day.Add(11 * time.Hour).Unix(), Alias: "AS", CursorModelsPct: 0.5, Participants: []model.HourParticipant{{Kind: "task", ID: "B", Alias: "AS"}}},
	}
	days := BuildDailySpend(hours, now, 3)
	if len(days) != 3 {
		t.Fatalf("days=%d", len(days))
	}
	got := days[2]
	if got.Hour != day.Unix() {
		t.Fatalf("day=%d want %d", got.Hour, day.Unix())
	}
	if got.CursorModelsPct != 1.5 || got.OtherModelsPct != 2 {
		t.Fatalf("pcts cursor=%v other=%v", got.CursorModelsPct, got.OtherModelsPct)
	}
	if got.BudgetUSD == nil || *got.BudgetUSD != 0.3 {
		t.Fatalf("budget=%v", got.BudgetUSD)
	}
	if got.Alias != "" {
		t.Fatalf("mixed aliases should clear hour alias, got %q", got.Alias)
	}
	if len(got.Aliases) != 2 || got.Aliases[0] != "AS" || got.Aliases[1] != "EK" {
		t.Fatalf("aliases=%v", got.Aliases)
	}
	if days[0].CursorModelsPct != 0 || days[1].CursorModelsPct != 0 {
		t.Fatalf("empty pad days should be zero")
	}
}

func TestBuildHourlySpendSequentialSameHourNotShared(t *testing.T) {
	loc := time.FixedZone("test", 0)
	hour := time.Date(2026, 9, 14, 11, 0, 0, 0, loc)
	now := hour.Add(50 * time.Minute)
	samples := []model.UsageSample{
		{TS: hour.Add(-1 * time.Minute).Unix(), CursorModelsPct: 1, OtherModelsPct: 1},
		{TS: hour.Add(10 * time.Minute).Unix(), CursorModelsPct: 3, OtherModelsPct: 1},
	}
	members := []model.Member{{
		Alias: "EK",
		Tasks: []model.MemberTask{
			{
				TaskID: "TASK-A",
				Title:  "A",
				ActivityWindows: []model.ActivityWindow{{
					StartedAt: hour.Add(5 * time.Minute),
					EndedAt:   hour.Add(15 * time.Minute),
				}},
			},
			{
				TaskID: "TASK-B",
				Title:  "B",
				ActivityWindows: []model.ActivityWindow{{
					StartedAt: hour.Add(35 * time.Minute),
					EndedAt:   hour.Add(45 * time.Minute),
				}},
			},
		},
	}}
	hours := BuildHourlySpend(samples, members, now, 2*time.Hour)
	if len(hours) != 1 {
		t.Fatalf("hours=%d", len(hours))
	}
	if len(hours[0].Participants) != 2 {
		t.Fatalf("participants=%+v", hours[0].Participants)
	}
	for _, p := range hours[0].Participants {
		if p.Shared {
			t.Fatalf("sequential windows in one hour must not be shared: %+v", p)
		}
	}
}

func TestApplyCompletedSpendSharingSoloWindowsKeepMeter(t *testing.T) {
	loc := time.FixedZone("test", 0)
	start := time.Date(2026, 9, 14, 10, 0, 0, 0, loc)
	price := 60.0
	tasks := []model.Task{
		{
			TaskID:    "A",
			Status:    "completed",
			StartedAt: start.Unix(),
			CompletedAt: start.Add(2 * time.Hour).Unix(),
			BudgetUSD: f64(9),
			CostUSD:   f64(1),
			ActivityWindows: []model.ActivityWindow{{
				StartedAt: start.Add(10 * time.Minute),
				EndedAt:   start.Add(30 * time.Minute),
			}},
		},
		{
			TaskID:    "B",
			Status:    "in_progress",
			StartedAt: start.Unix(),
			ActivityWindows: []model.ActivityWindow{{
				StartedAt: start.Add(90 * time.Minute),
				EndedAt:   start.Add(100 * time.Minute),
			}},
		},
	}
	samples := []model.UsageSample{
		{TS: start.Unix(), CursorModelsPct: 10, OtherModelsPct: 10, PlanPriceUSD: &price, Plan: "pro_plus"},
		{TS: start.Add(20 * time.Minute).Unix(), CursorModelsPct: 12, OtherModelsPct: 10, PlanPriceUSD: &price, Plan: "pro_plus"},
		{TS: start.Add(95 * time.Minute).Unix(), CursorModelsPct: 20, OtherModelsPct: 10, PlanPriceUSD: &price, Plan: "pro_plus"},
	}
	applyCompletedSpendSharing(tasks, nil, samples)
	if tasks[0].SpendShared {
		t.Fatal("sequential windows must stay solo")
	}
	if tasks[0].BudgetUSD == nil || *tasks[0].BudgetUSD != 0.6 {
		t.Fatalf("solo window budget=%v", tasks[0].BudgetUSD)
	}
	if tasks[0].CursorModelsPct == nil || *tasks[0].CursorModelsPct != 2 {
		t.Fatalf("solo window cursor=%v", tasks[0].CursorModelsPct)
	}
}
