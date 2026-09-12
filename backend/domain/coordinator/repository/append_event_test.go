package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"coordinator/model"
)

func TestAppendEventWritesJSONL(t *testing.T) {
	root := t.TempDir()
	repo := NewFileRepositoryWithPaths(Paths{
		Bus:       root,
		Data:      filepath.Join(root, "data"),
		Workspace: root,
		App:       filepath.Join(root, "app"),
	})
	ev := model.Event{
		Timestamp: 1700000000,
		Event:     "coordinator_warning",
		TaskID:    "T-CLOCK",
		Alias:     "EK",
		Service:   "LLM",
		Summary:   "Peer Scope Overlap: EK and AS both claim LLM",
		Findings:  "peer-scope-overlap|as,ek|llm|…",
	}
	if err := repo.AppendEvent(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(repo.eventsFile("EK"))
	if err != nil {
		t.Fatal(err)
	}
	var got model.Event
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Event != ev.Event || got.Service != "LLM" || got.Findings != ev.Findings || got.Alias != "EK" {
		t.Fatalf("got %+v", got)
	}
}
