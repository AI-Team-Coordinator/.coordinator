package coordinator

import (
	"testing"
	"time"

	"coordinator/model"
)

func TestComputeStatsActiveNowCountsSlots(t *testing.T) {
	stats := computeStats(nil, []model.Member{
		{
			Alias:  "EK",
			Status: "in_progress",
			Tasks: []model.MemberTask{
				{TaskID: "FIX-1"},
				{TaskID: "20260910-0732-EK-STATUS"},
			},
			Research: &model.Research{Status: "active"},
		},
		{Alias: "AS", Status: "idle"},
	})
	if stats.ActiveNow != 3 {
		t.Fatalf("active=%d want 2 tasks + research", stats.ActiveNow)
	}
}

func TestComputeStatsResearchSplit(t *testing.T) {
	now := time.Now().Unix()
	stats := computeStats([]model.Event{
		{Event: "task_completed", TaskID: "A", Timestamp: now, BudgetUSD: f64(6), SpendKind: "product"},
		{Event: "research_completed", Alias: "EK", Timestamp: now, BudgetUSD: f64(2), CostUSD: f64(0.5), SpendKind: "research"},
	}, []model.Member{
		{Research: &model.Research{Status: "active", BudgetUSD: f64(1)}},
	})
	if stats.BudgetUSDToday != 6 {
		t.Fatalf("task budget today=%v", stats.BudgetUSDToday)
	}
	if stats.BudgetUSDResearchToday != 3 {
		t.Fatalf("research budget today=%v", stats.BudgetUSDResearchToday)
	}
	if stats.BudgetUSDResearchOpen != 1 {
		t.Fatalf("research open=%v", stats.BudgetUSDResearchOpen)
	}
	if stats.CostUSDResearchToday != 0.5 {
		t.Fatalf("research cost=%v", stats.CostUSDResearchToday)
	}
	if stats.TotalCompleted != 1 {
		t.Fatalf("completed tasks=%d", stats.TotalCompleted)
	}
	if stats.ResearchCompleted != 1 || stats.ResearchCompletedToday != 1 {
		t.Fatalf("research completed=%d today=%d", stats.ResearchCompleted, stats.ResearchCompletedToday)
	}
	if stats.ActiveNow != 1 {
		t.Fatalf("active=%d", stats.ActiveNow)
	}
}

func TestComputeStatsPoolPctsWithoutPlanPrice(t *testing.T) {
	now := time.Now().Unix()
	stats := computeStats([]model.Event{
		{Event: "task_completed", TaskID: "A", Timestamp: now, CursorModelsPct: f64(4), OtherModelsPct: f64(10), SpendKind: "product"},
		{Event: "research_completed", Timestamp: now, CursorModelsPct: f64(1.5), OtherModelsPct: f64(2)},
	}, []model.Member{
		{Status: "in_progress", SpendKind: "infra", CursorModelsPct: f64(0.5), OtherModelsPct: f64(1)},
	})
	if stats.CursorModelsPctCycle != 4.5 || stats.OtherModelsPctCycle != 11 {
		t.Fatalf("task cycle cursor=%v other=%v", stats.CursorModelsPctCycle, stats.OtherModelsPctCycle)
	}
	if stats.CursorModelsPctProductCycle != 4 || stats.CursorModelsPctInfraCycle != 0.5 {
		t.Fatalf("product=%v infra=%v", stats.CursorModelsPctProductCycle, stats.CursorModelsPctInfraCycle)
	}
	if stats.CursorModelsPctResearchCycle != 1.5 || stats.OtherModelsPctResearchCycle != 2 {
		t.Fatalf("research cursor=%v other=%v", stats.CursorModelsPctResearchCycle, stats.OtherModelsPctResearchCycle)
	}
	if stats.BudgetUSDCycle != 0 {
		t.Fatalf("budget=%v", stats.BudgetUSDCycle)
	}
}
