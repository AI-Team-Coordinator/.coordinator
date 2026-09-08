package coordinator

import (
	"testing"
	"time"

	"coordinator/model"
)

func f64(v float64) *float64 { return &v }

func TestComputeStatsCursorCost(t *testing.T) {
	now := time.Now().Unix()
	old := now - 20*86400
	stats := computeStats([]model.Event{
		{Event: "task_completed", TaskID: "A", Timestamp: now, CostUSD: f64(1.25)},
		{Event: "task_completed", TaskID: "B", Timestamp: now, CostUSD: f64(0.75)},
		{Event: "task_completed", TaskID: "C", Timestamp: old, CostUSD: f64(4)},
		{Event: "task_completed", TaskID: "D", Timestamp: now},
	}, nil)
	if stats.TotalCompleted != 4 {
		t.Fatalf("completed=%d", stats.TotalCompleted)
	}
	if stats.CostTasks != 3 || stats.CostUSDTotal != 6 {
		t.Fatalf("cost tasks=%d total=%v", stats.CostTasks, stats.CostUSDTotal)
	}
	if stats.CostUSDToday != 2 {
		t.Fatalf("today=%v", stats.CostUSDToday)
	}
	if stats.CostUSDAvg != 2 {
		t.Fatalf("avg=%v", stats.CostUSDAvg)
	}
}

func TestComputeStatsBudgetShare(t *testing.T) {
	now := time.Now().Unix()
	stats := computeStats([]model.Event{
		{Event: "task_completed", TaskID: "A", Timestamp: now, BudgetUSD: f64(4.5), CostUSD: f64(0), OnDemandUSD: f64(0), SpendKind: "product"},
		{Event: "task_completed", TaskID: "B", Timestamp: now, BudgetUSD: f64(1.5), OnDemandUSD: f64(2.5), SpendKind: "infra"},
	}, nil)
	if stats.BudgetUSDToday != 6 || stats.BudgetTasks != 2 {
		t.Fatalf("budget today=%v tasks=%d", stats.BudgetUSDToday, stats.BudgetTasks)
	}
	if stats.BudgetUSDProductToday != 4.5 || stats.BudgetUSDInfraToday != 1.5 {
		t.Fatalf("product=%v infra=%v", stats.BudgetUSDProductToday, stats.BudgetUSDInfraToday)
	}
	if stats.OnDemandUSDToday != 2.5 {
		t.Fatalf("ondemand=%v", stats.OnDemandUSDToday)
	}
}

func TestComputeStatsIncludesOpenTasks(t *testing.T) {
	now := time.Now().Unix()
	stats := computeStats([]model.Event{
		{Event: "task_completed", TaskID: "A", Timestamp: now, BudgetUSD: f64(1), CostUSD: f64(0.5), SpendKind: "product"},
	}, []model.Member{
		{Status: "in_progress", SpendKind: "infra", BudgetUSD: f64(2), CostUSD: f64(0.1), OnDemandUSD: f64(0)},
	})
	if stats.BudgetUSDToday != 3 || stats.BudgetUSDOpen != 2 {
		t.Fatalf("budget today=%v open=%v", stats.BudgetUSDToday, stats.BudgetUSDOpen)
	}
	if stats.BudgetUSDInfraToday != 2 || stats.BudgetUSDProductToday != 1 {
		t.Fatalf("infra=%v product=%v", stats.BudgetUSDInfraToday, stats.BudgetUSDProductToday)
	}
	if stats.CostUSDToday != 0.6 || stats.CostUSDOpen != 0.1 {
		t.Fatalf("cost today=%v open=%v", stats.CostUSDToday, stats.CostUSDOpen)
	}
}

func TestComputeTasksStatusAndFilters(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	tasks := computeTasks([]model.Event{
		{Event: "task_started", TaskID: "20260908-1206-EK-COORDINATOR_CURSOR_USAGE", Alias: "EK", Branch: "main", Timestamp: now.Unix() - 3600},
		{Event: "task_completed", TaskID: "20260908-1206-EK-COORDINATOR_CURSOR_USAGE", Alias: "EK", Branch: "main", Timestamp: now.Unix() - 60, BudgetUSD: f64(0.45), SpendKind: "infra"},
		{Event: "task_started", TaskID: "FIX-20260907-1825-EK-AVATAR", Alias: "EK", Branch: "fix/avatar", Timestamp: now.Unix() - 120},
		{Event: "deploy_finished", TaskID: "20260908-1206-EK-COORDINATOR_CURSOR_USAGE", Timestamp: now.Unix() - 30},
	}, []model.Member{
		{Alias: "EK", Status: "in_progress", TaskID: "20260908-1326-EK-COORDINATOR_STATS_TASK_TABLE", Branch: "main", TaskTitle: "Coordinator Stats: таблица задач", UpdatedAt: now.Add(-5 * time.Minute)},
	}, now)

	if len(tasks) != 3 {
		t.Fatalf("tasks=%d", len(tasks))
	}
	if tasks[0].Status != "in_progress" || tasks[2].Status != "completed" {
		t.Fatalf("order=%s %s %s", tasks[0].Status, tasks[1].Status, tasks[2].Status)
	}
	var current *model.Task
	for i := range tasks {
		if tasks[i].TaskID == "20260908-1326-EK-COORDINATOR_STATS_TASK_TABLE" {
			current = &tasks[i]
			break
		}
	}
	if current == nil || current.Status != "in_progress" || current.DurationSeconds != 300 {
		t.Fatalf("current=%v", current)
	}

	open := filterTasks(tasks, model.TaskQuery{Status: "in_progress"})
	if len(open) != 2 {
		t.Fatalf("open=%d", len(open))
	}
	done := filterTasks(tasks, model.TaskQuery{Status: "completed"})
	if len(done) != 1 || done[0].BudgetUSD == nil || *done[0].BudgetUSD != 0.45 {
		t.Fatalf("completed=%v", done)
	}
	fixes := filterTasks(tasks, model.TaskQuery{Kind: "fix"})
	if len(fixes) != 1 || fixes[0].Status != "in_progress" {
		t.Fatalf("fixes=%v", fixes)
	}
	page := pageTasks(tasks, 1, 0)
	if len(page) != 1 || page[0].Status != "in_progress" {
		t.Fatalf("page=%v", page)
	}
}
