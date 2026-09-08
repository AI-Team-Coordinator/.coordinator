package coordinator

import (
	"testing"
	"time"

	"coordinator/model"
)

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
