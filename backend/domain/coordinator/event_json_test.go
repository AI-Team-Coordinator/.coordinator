package coordinator

import (
	"encoding/json"
	"testing"

	"coordinator/model"
)

func TestEventUnmarshalTaskID(t *testing.T) {
	raw := `{"timestamp": 1788813019, "event": "task_started", "task_id": "20260907-2129-EK-ADMIN_STATS_MONTH_PACE", "branch": "feat/admin-stats-month-pace"}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.TaskID != "20260907-2129-EK-ADMIN_STATS_MONTH_PACE" {
		t.Fatalf("task_id=%q", ev.TaskID)
	}
	if ev.Event != "task_started" || ev.Branch == "" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestEventUnmarshalCursorCost(t *testing.T) {
	raw := `{"timestamp": 1788813019, "event": "task_completed", "task_id": "T1", "cost_usd": 1.25, "cursor_models_pct": 0.4, "usage_plan": "ultra"}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.CostUSD == nil || *ev.CostUSD != 1.25 {
		t.Fatalf("cost_usd=%v", ev.CostUSD)
	}
	if ev.CursorModelsPct == nil || *ev.CursorModelsPct != 0.4 {
		t.Fatalf("cursor_models_pct=%v", ev.CursorModelsPct)
	}
	if ev.UsagePlan != "ultra" {
		t.Fatalf("plan=%q", ev.UsagePlan)
	}
}

func TestEventUnmarshalSpendKind(t *testing.T) {
	raw := `{"timestamp": 1, "event": "task_completed", "task_id": "T1", "spend_kind": "infra", "budget_usd": 1.2}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.SpendKind != "infra" {
		t.Fatalf("spend_kind=%q", ev.SpendKind)
	}
}

func TestEventUnmarshalActiveSeconds(t *testing.T) {
	raw := `{"timestamp": 1, "event": "task_completed", "task_id": "T1", "active_seconds": 420}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.ActiveSeconds == nil || *ev.ActiveSeconds != 420 {
		t.Fatalf("active_seconds=%v", ev.ActiveSeconds)
	}
}

func TestEventUnmarshalHandoff(t *testing.T) {
	raw := `{"timestamp": 1, "event": "handoff", "task_id": "T1", "alias": "EK", "to_alias": "AS", "agent_id": "sess-1", "to_agent": "sess-2", "service": "Core"}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.ToAlias != "AS" || ev.AgentID != "sess-1" || ev.ToAgent != "sess-2" {
		t.Fatalf("got %+v", ev)
	}
}

func TestEventUnmarshalActivityWindows(t *testing.T) {
	raw := `{"timestamp": 1, "event": "task_completed", "task_id": "T1", "activity_windows": [{"started_at": "2026-09-14T10:00:00Z", "ended_at": "2026-09-14T10:20:00Z"}]}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if len(ev.ActivityWindows) != 1 {
		t.Fatalf("windows=%+v", ev.ActivityWindows)
	}
	if ev.ActivityWindows[0].StartedAt.UTC().Hour() != 10 {
		t.Fatalf("started=%v", ev.ActivityWindows[0].StartedAt)
	}
}

func TestEventUnmarshalRelated(t *testing.T) {
	raw := `{"timestamp": 1, "event": "task_started", "task_id": "FIX-1", "related_tasks": ["T-PARENT"], "related_docs": ["docs/T-PARENT.md"]}`
	var ev model.Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if len(ev.RelatedTasks) != 1 || ev.RelatedTasks[0] != "T-PARENT" {
		t.Fatalf("tasks=%v", ev.RelatedTasks)
	}
	if len(ev.RelatedDocs) != 1 || ev.RelatedDocs[0] != "docs/T-PARENT.md" {
		t.Fatalf("docs=%v", ev.RelatedDocs)
	}
}
