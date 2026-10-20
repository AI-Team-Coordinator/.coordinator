package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHumanizeTaskID(t *testing.T) {
	got := humanizeTaskID("20260907-2129-EK-ADMIN_STATS_MONTH_PACE")
	if got != "Admin stats month pace" {
		t.Fatalf("got %q", got)
	}
	got = humanizeTaskID("FIX-20260907-1825-EK-AVATAR_MISSING_STYLE")
	if got != "Avatar missing style" {
		t.Fatalf("got %q", got)
	}
}

func TestGetTaskDoc(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	id := "20260907-2129-EK-ADMIN_STATS_MONTH_PACE"
	body := "# Admin stats\n\nPace of the month.\n"
	if err := os.WriteFile(filepath.Join(docs, id+".md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := NewFileRepository(root, "")
	doc, err := repo.GetTaskDoc(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "Admin stats" || doc.Markdown != body || doc.RelPath != "docs/"+id+".md" {
		t.Fatalf("%+v", doc)
	}
	if _, err := repo.GetTaskDoc(context.Background(), "../secret"); err != os.ErrInvalid {
		t.Fatalf("traversal: %v", err)
	}
	if _, err := repo.GetTaskDoc(context.Background(), "20260907-2129-EK-MISSING_DOC"); !os.IsNotExist(err) {
		t.Fatalf("missing: %v", err)
	}
}
