package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsAlinaAssistProfile(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings")
	if err := os.MkdirAll(settings, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(settings, "project_profile.json")
	if err := os.WriteFile(path, []byte(`{"project":{"id":"alina-assist","name":"Alina Assist"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isAlinaAssistProfile(dir) {
		t.Fatal("expected Alina Assist profile")
	}
	if err := os.WriteFile(path, []byte(`{"project":{"id":"","name":"TestA"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if isAlinaAssistProfile(dir) {
		t.Fatal("empty id must not match")
	}
}

func TestDetectUIModeEnvWins(t *testing.T) {
	t.Setenv("UI_MODE", "static")
	if got := detectUIMode(t.TempDir()); got != "static" {
		t.Fatalf("got %q", got)
	}
	t.Setenv("UI_MODE", "vite")
	if got := detectUIMode(t.TempDir()); got != "vite" {
		t.Fatalf("got %q", got)
	}
}
