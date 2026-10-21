package bus

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"relay/model"
)

func TestAppendSetsReceivedAtAndSourcePath(t *testing.T) {
	dir := t.TempDir()
	j := JSONL{Root: dir}
	ev := model.ExternalEvent{
		ID:         "github:ci:1",
		Source:     "github",
		ExternalID: "1",
		Class:      model.ClassWarning,
		Kind:       "ci.failed",
	}
	if err := j.Append(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "progress", "external", "events", "github")
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries=%v", entries)
	}
	raw, err := os.ReadFile(filepath.Join(path, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(string(raw))
	var row model.ExternalEvent
	if err := json.Unmarshal([]byte(line), &row); err != nil {
		t.Fatal(err)
	}
	if row.SchemaVersion != model.ExternalEventSchemaVersion || row.ReceivedAt == "" {
		t.Fatalf("%+v", row)
	}
	if row.Data != nil && len(row.Data) > 0 {
		t.Fatalf("unexpected data")
	}
}
