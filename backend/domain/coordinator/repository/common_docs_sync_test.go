package repository

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSyncCommonDocsCommitsAndPushesDocsOnly(t *testing.T) {
	prev := docsSyncQuiet
	docsSyncQuiet = 0
	t.Cleanup(func() { docsSyncQuiet = prev })

	origin, local := docsSyncRepos(t)
	if err := os.MkdirAll(filepath.Join(local, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "docs", "NOTE.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "notes.txt"), []byte("leave me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, local, "add", "notes.txt")

	repo := NewFileRepository(local, "")
	got, err := repo.SyncCommonDocs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Committed || !got.Pushed || got.Pulled {
		t.Fatalf("sync = %+v", got)
	}

	msg := gitOutputTest(t, origin, "log", "-1", "--pretty=%s")
	if msg != "docs: sync working copy" {
		t.Fatalf("origin message %q", msg)
	}
	files := gitOutputTest(t, origin, "ls-tree", "-r", "--name-only", "HEAD")
	if !strings.Contains(files, "docs/NOTE.md") || strings.Contains(files, "notes.txt") {
		t.Fatalf("origin tree:\n%s", files)
	}
	staged := gitOutputTest(t, local, "diff", "--cached", "--name-only")
	if strings.TrimSpace(staged) != "notes.txt" {
		t.Fatalf("staged after sync %q", staged)
	}
}

func TestSyncCommonDocsSkipsFreshDocs(t *testing.T) {
	prev := docsSyncQuiet
	docsSyncQuiet = time.Hour
	t.Cleanup(func() { docsSyncQuiet = prev })

	_, local := docsSyncRepos(t)
	if err := os.MkdirAll(filepath.Join(local, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "docs", "FRESH.md"), []byte("now\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := NewFileRepository(local, "")
	got, err := repo.SyncCommonDocs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Committed || got.Pushed || got.Pulled {
		t.Fatalf("fresh file should wait, got %+v", got)
	}
}

func TestSyncCommonDocsPullsRemoteDocs(t *testing.T) {
	prev := docsSyncQuiet
	docsSyncQuiet = 0
	t.Cleanup(func() { docsSyncQuiet = prev })

	origin, local := docsSyncRepos(t)
	other := filepath.Join(t.TempDir(), "other")
	gitRun(t, local, "clone", origin, other)
	gitRun(t, other, "config", "user.email", "t@t")
	gitRun(t, other, "config", "user.name", "t")
	if err := os.MkdirAll(filepath.Join(other, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "docs", "REMOTE.md"), []byte("from other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, other, "add", "docs/REMOTE.md")
	gitRun(t, other, "commit", "-m", "docs: remote")
	gitRun(t, other, "push", "origin", "HEAD:main")

	repo := NewFileRepository(local, "")
	got, err := repo.SyncCommonDocs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Pulled || got.Committed || got.Pushed {
		t.Fatalf("sync = %+v", got)
	}
	body, err := os.ReadFile(filepath.Join(local, "docs", "REMOTE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "from other\n" {
		t.Fatalf("pulled body %q", body)
	}
}

func TestSyncCommonDocsSkipsNonMain(t *testing.T) {
	_, local := docsSyncRepos(t)
	gitRun(t, local, "checkout", "-b", "feat/x")
	if err := os.MkdirAll(filepath.Join(local, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(local, "docs", "NOTE.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := NewFileRepository(local, "")
	got, err := repo.SyncCommonDocs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Committed || got.Pulled || got.Pushed {
		t.Fatalf("non-main should no-op, got %+v", got)
	}
}

func docsSyncRepos(t *testing.T) (origin, local string) {
	t.Helper()
	root := t.TempDir()
	origin = filepath.Join(root, "origin.git")
	local = filepath.Join(root, "local")
	gitRun(t, root, "init", "--bare", "-b", "main", origin)
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, local, "init", "-b", "main")
	gitRun(t, local, "config", "user.email", "t@t")
	gitRun(t, local, "config", "user.name", "t")
	gitRun(t, local, "commit", "--allow-empty", "-m", "base")
	gitRun(t, local, "remote", "add", "origin", origin)
	gitRun(t, local, "push", "-u", "origin", "main")
	return origin, local
}

func gitOutputTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
