package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"coordinator/model"
)

func TestOverlayReportedReposAddsLocalFromReport(t *testing.T) {
	report := &model.GitReport{Repos: []model.GitReportRepo{{
		ID: "core", Repo: "Core", Name: "Core", HasLocalBranch: true, Dirty: true, Unpushed: 2,
	}}}
	got := overlayReportedRepos(nil, report)
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].State != "local" || !got[0].Dirty || got[0].Ahead != 2 {
		t.Fatalf("got %+v", got[0])
	}
}

func TestOverlayReportedReposKeepsPushedAndOverlaysDirty(t *testing.T) {
	live := []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "pushed"}}
	report := &model.GitReport{Repos: []model.GitReportRepo{{
		ID: "core", Repo: "Core", Dirty: true,
	}}}
	got := overlayReportedRepos(live, report)
	if len(got) != 1 || got[0].State != "pushed" || !got[0].Dirty {
		t.Fatalf("got %+v", got)
	}
}

func TestOverlayReportedReposUnpushedTurnsPushedIntoLocal(t *testing.T) {
	live := []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "pushed"}}
	report := &model.GitReport{Repos: []model.GitReportRepo{{
		ID: "core", Repo: "Core", HasLocalBranch: true, Unpushed: 1,
	}}}
	got := overlayReportedRepos(live, report)
	if len(got) != 1 || got[0].State != "local" || got[0].Ahead != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestOverlayReportedReposMergedIgnoresDirty(t *testing.T) {
	live := []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "merged", Dirty: true}}
	report := &model.GitReport{Repos: []model.GitReportRepo{{
		ID: "core", Repo: "Core", Dirty: true, Unpushed: 1,
	}}}
	got := overlayReportedRepos(live, report)
	if len(got) != 1 || got[0].State != "merged" || got[0].Dirty {
		t.Fatalf("got %+v", got)
	}
}

func TestOverlayReportedReposStripsLocalDirtyWithoutReport(t *testing.T) {
	live := []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "local", Dirty: true, Ahead: 1}}
	got := overlayReportedRepos(live, nil)
	if len(got) != 1 || got[0].Dirty || got[0].Ahead != 1 {
		t.Fatalf("got %+v", got[0])
	}
}

func TestGitReportEqualIgnoresReportedAt(t *testing.T) {
	now := time.Date(2026, 9, 8, 16, 0, 0, 0, time.UTC)
	repos := []model.RepoWork{{
		ID: "core", Repo: "Core", State: "local", Ahead: 1, Dirty: true, HasLocalBranch: true,
	}}
	a := buildGitReportFromRepos(repos, now)
	b := buildGitReportFromRepos(repos, now.Add(time.Minute))
	if !gitReportEqual(a, b) {
		t.Fatal("expected equal")
	}
	b.Repos[0].Dirty = false
	if gitReportEqual(a, b) {
		t.Fatal("expected different dirty")
	}
}

func TestWriteGitReportPreservesDocAndSummary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".current_task_EK")
	initial := map[string]any{
		"alias":    "EK",
		"task_id":  "FIX-20260908-1635-EK-AVATAR",
		"status":   "in_progress",
		"doc":      "docs/foo.md",
		"summary":  "restore dark header avatar",
		"services": []string{"Website"},
	}
	raw, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	report := &model.GitReport{Repos: []model.GitReportRepo{{ID: "website", Repo: "Website", Dirty: true}}}
	if err := writeGitReport(path, report); err != nil {
		t.Fatal(err)
	}
	var snap map[string]any
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatal(err)
	}
	if snap["doc"] != "docs/foo.md" || snap["summary"] != "restore dark header avatar" {
		t.Fatalf("intent dropped: %+v", snap)
	}
	gr, _ := snap["git_report"].(map[string]any)
	if gr == nil {
		t.Fatalf("git_report missing: %+v", snap)
	}
}
