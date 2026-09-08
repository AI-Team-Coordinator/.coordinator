package repository

import (
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
