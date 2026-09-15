package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"coordinator/model"
)

func TestStoreRebuildsFromJSONL(t *testing.T) {
	root := t.TempDir()
	eventsDir := filepath.Join(root, "events")
	yearDir := filepath.Join(eventsDir, "EK")
	if err := os.MkdirAll(yearDir, 0o755); err != nil {
		t.Fatal(err)
	}

	year := time.Now().Year()
	payload := `{"timestamp": 100, "event": "task_started", "task_id": "T1", "branch": "feat/x"}
{"timestamp": 200, "event": "deploy_finished", "task_id": "T1", "service": "Core", "status": "finished", "alias": "EK"}
{"timestamp": 300, "event": "task_completed", "task_id": "T1", "alias": "EK", "cost_usd": 1.5, "budget_usd": 4.5, "cursor_models_pct": 0.4, "usage_plan": "ultra", "spend_kind": "infra", "activity_windows": [{"started_at": "2026-09-14T10:00:00Z", "ended_at": "2026-09-14T10:20:00Z"}]}
`
	if err := os.WriteFile(filepath.Join(yearDir, itoa(year)+".jsonl"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "AB.jsonl"), []byte(`{"timestamp": 150, "event": "task_completed", "task_id": "T2"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cacheDir := filepath.Join(root, "cache")
	store, err := Open(cacheDir, eventsDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	items, total, err := store.List(ctx, model.EventQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || len(items) != 4 {
		t.Fatalf("total=%d len=%d", total, len(items))
	}
	if items[0].Timestamp != 300 || items[0].CostUSD == nil || *items[0].CostUSD != 1.5 {
		t.Fatalf("newest with cost: %+v", items[0])
	}
	if items[0].BudgetUSD == nil || *items[0].BudgetUSD != 4.5 {
		t.Fatalf("budget: %+v", items[0].BudgetUSD)
	}
	if items[0].SpendKind != "infra" {
		t.Fatalf("spend_kind: %q", items[0].SpendKind)
	}
	if len(items[0].ActivityWindows) != 1 {
		t.Fatalf("activity_windows: %+v", items[0].ActivityWindows)
	}
	if items[1].Timestamp != 200 || items[1].Service != "Core" {
		t.Fatalf("deploy: %+v", items[1])
	}
	if items[3].Alias != "EK" {
		t.Fatalf("alias filled from folder: %+v", items[3])
	}

	filtered, n, err := store.List(ctx, model.EventQuery{Alias: "AB", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 || len(filtered) != 1 || filtered[0].TaskID != "T2" {
		t.Fatalf("alias filter: n=%d items=%+v", n, filtered)
	}

	_ = store.Close()
	store, err = Open(cacheDir, eventsDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	again, total2, err := store.List(ctx, model.EventQuery{Event: "deploy_finished"})
	if err != nil {
		t.Fatal(err)
	}
	if total2 != 1 || again[0].Status != "finished" {
		t.Fatalf("reopen cache: n=%d %+v", total2, again)
	}

	if err := os.Remove(filepath.Join(cacheDir, dbFileName)); err != nil {
		t.Fatal(err)
	}
	rebuilt, err := Open(cacheDir, eventsDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rebuilt.Close() })
	all, total3, err := rebuilt.List(ctx, model.EventQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if total3 != 4 || len(all) != 4 {
		t.Fatalf("rebuild from jsonl: total=%d", total3)
	}
}

func itoa(n int) string {
	return time.Date(n, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}
