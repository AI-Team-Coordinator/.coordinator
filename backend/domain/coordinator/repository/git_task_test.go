package repository

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"coordinator/model"
)

func sampleServices() []model.ServiceNode {
	return []model.ServiceNode{
		{ID: "core", Name: "Core", Group: "platform", Kind: "api", Repo: "Core"},
		{ID: "common", Name: "Common", Group: "workspace", Kind: "workspace", Repo: "Common"},
		{ID: "cursor", Name: ".cursor", Group: "workspace", Kind: "workspace", Repo: ".cursor", GitHubRepo: ".cursor"},
	}
}

func TestReposToInspectClaimedCommonOnMain(t *testing.T) {
	got := reposToInspect(sampleServices(), []string{"Common"}, "main")
	if len(got) != 1 || got[0].ID != "common" {
		t.Fatalf("got %+v", got)
	}
}

func TestReposToInspectTrunkWithoutClaimIsEmpty(t *testing.T) {
	got := reposToInspect(sampleServices(), nil, "main")
	if len(got) != 0 {
		t.Fatalf("expected no product scan on main, got %+v", got)
	}
}

func TestReposToInspectTopicBranchSkipsWorkspace(t *testing.T) {
	got := reposToInspect(sampleServices(), nil, "feat/x")
	if len(got) != 1 || got[0].ID != "core" {
		t.Fatalf("got %+v", got)
	}
}

func TestClassifyProductRepoNewBranchIsLocal(t *testing.T) {
	state, ahead := classifyProductRepo(0, false, false, 0)
	if state != "local" || ahead != 0 {
		t.Fatalf("got %s +%d", state, ahead)
	}
}

func TestClassifyProductRepoMergedOnlyWhenOnMain(t *testing.T) {
	state, ahead := classifyProductRepo(0, true, true, 0)
	if state != "merged" || ahead != 0 {
		t.Fatalf("got %s +%d", state, ahead)
	}
}

func TestClassifyProductRepoPushed(t *testing.T) {
	state, ahead := classifyProductRepo(3, false, true, 0)
	if state != "pushed" || ahead != 0 {
		t.Fatalf("got %s +%d", state, ahead)
	}
}

func TestClassifyProductRepoLocalUnpushed(t *testing.T) {
	state, ahead := classifyProductRepo(3, false, true, 2)
	if state != "local" || ahead != 2 {
		t.Fatalf("got %s +%d", state, ahead)
	}
}

func TestInspectOneRepoNewBranchIsLocalNotMerged(t *testing.T) {
	dir := initProductRepo(t)
	gitRun(t, dir, "checkout", "-b", "feat/x")
	work, ok := inspectOneRepo(dir, sampleServices()[0], "feat/x", "TASK-1")
	if !ok {
		t.Fatal("expected repo work")
	}
	if work.State != "local" || work.Ahead != 0 || work.Dirty {
		t.Fatalf("got state=%s ahead=%d dirty=%v", work.State, work.Ahead, work.Dirty)
	}
}

func TestInspectOneRepoDirtyOnTaskBranch(t *testing.T) {
	dir := initProductRepo(t)
	gitRun(t, dir, "checkout", "-b", "feat/x")
	if err := os.WriteFile(filepath.Join(dir, "wip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	work, ok := inspectOneRepo(dir, sampleServices()[0], "feat/x", "TASK-1")
	if !ok || work.State != "local" || !work.Dirty || !work.HasLocalBranch {
		t.Fatalf("got ok=%v state=%s dirty=%v local=%v", ok, work.State, work.Dirty, work.HasLocalBranch)
	}
}

func TestInspectOneRepoDirtyIgnoredOffBranch(t *testing.T) {
	dir := initProductRepo(t)
	gitRun(t, dir, "checkout", "-b", "feat/x")
	gitRun(t, dir, "checkout", "main")
	if err := os.WriteFile(filepath.Join(dir, "wip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	work, ok := inspectOneRepo(dir, sampleServices()[0], "feat/x", "TASK-1")
	if !ok || work.Dirty {
		t.Fatalf("got ok=%v dirty=%v", ok, work.Dirty)
	}
}

func TestInspectOneRepoMergedWhenTaskOnMain(t *testing.T) {
	dir := initProductRepo(t)
	gitRun(t, dir, "commit", "--allow-empty", "-m", "merge TASK-1")
	gitRun(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, dir, "checkout", "-b", "feat/x")
	work, ok := inspectOneRepo(dir, sampleServices()[0], "feat/x", "TASK-1")
	if !ok || work.State != "merged" {
		t.Fatalf("got ok=%v state=%s", ok, work.State)
	}
}

func TestServiceLabelMatchesCursorDot(t *testing.T) {
	svc := sampleServices()[2]
	if !serviceLabelMatches(".cursor", svc) || !serviceLabelMatches("cursor", svc) {
		t.Fatal("cursor labels")
	}
}

func initProductRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-b", "main")
	gitRun(t, dir, "config", "user.email", "t@t")
	gitRun(t, dir, "config", "user.name", "t")
	gitRun(t, dir, "commit", "--allow-empty", "-m", "base")
	gitRun(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t",
		"GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t",
		"GIT_COMMITTER_EMAIL=t@t",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
