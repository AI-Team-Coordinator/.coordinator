package coordinator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"coordinator/domain/coordinator/dto"
	"coordinator/domain/coordinator/repository"
	"coordinator/model"
)

func TestNeedsSetup(t *testing.T) {
	if !needsSetup(nil) {
		t.Fatal("nil profile needs setup")
	}
	if needsSetup(&model.ProjectProfile{Setup: model.ProjectSetup{Completed: true}}) {
		t.Fatal("completed setup")
	}
	if needsSetup(&model.ProjectProfile{
		Project:  model.ProjectMeta{Name: "Alina Assist"},
		Services: []model.ServiceNode{{ID: "core"}},
	}) {
		t.Fatal("existing product map should skip wizard")
	}
	if !needsSetup(&model.ProjectProfile{Project: model.ProjectMeta{Name: "Acme"}}) {
		t.Fatal("named seed without services still needs confirm")
	}
}

func TestSetupCoordinatorName(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	ctx := context.Background()

	state, err := svc.GetSetup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.CoordinatorName != "Otto" {
		t.Fatalf("default name=%q", state.CoordinatorName)
	}

	settings := filepath.Join(repo.WorkspaceDir(), "coordinator-data", "settings")
	if err := os.WriteFile(filepath.Join(settings, "locale.json"), []byte(`{"version":1,"chat_language":"ru","docs_language":"ru"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	state, err = svc.GetSetup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.CoordinatorName != "Отто" {
		t.Fatalf("ru default name=%q", state.CoordinatorName)
	}

	if err := os.WriteFile(filepath.Join(settings, "coordinator.json"), []byte(`{"version":1,"name":"Ottilie"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	state, err = svc.GetSetup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.CoordinatorName != "Ottilie" {
		t.Fatalf("configured name=%q", state.CoordinatorName)
	}
}

func TestCompleteSetupWritesLocalFiles(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	ctx := context.Background()

	state, err := svc.GetSetup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Needed {
		t.Fatalf("expected needed setup: %+v", state)
	}
	if state.Layout != "in-repo" {
		t.Fatalf("layout=%s", state.Layout)
	}

	done, err := svc.CompleteSetup(ctx, dto.CompleteSetupRequest{
		ProjectName: "Acme Tools",
		Alias:       "ak",
		Name:        "Alex",
		Role:        "founder",
		Language:    "ru",
	})
	if err != nil {
		t.Fatal(err)
	}
	if done.Needed {
		t.Fatal("setup should be complete")
	}

	profile, err := repo.GetProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Project.Name != "Acme Tools" || !profile.Setup.Completed || profile.Project.ID != "acme-tools" {
		t.Fatalf("profile %+v", profile)
	}
	team, err := repo.GetTeam(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(team) != 1 || team[0].Alias != "AK" || team[0].Access != "admin" {
		t.Fatalf("team %+v", team)
	}
	alias, err := repo.CurrentAuthor(ctx)
	if err != nil || alias != "AK" {
		t.Fatalf("author %q %v", alias, err)
	}
	loc, err := repo.GetLocale(ctx)
	if err != nil || loc.ChatLanguage != "ru" {
		t.Fatalf("locale %+v %v", loc, err)
	}
	if repo.Collaboration() != model.CollaborationSolo {
		t.Fatalf("collaboration=%s", repo.Collaboration())
	}
	cfg := repo.CoordinatorFile(ctx)
	if cfg.Collaboration != model.CollaborationSolo {
		t.Fatalf("coordinator file %+v", cfg)
	}
	if cfg.Name != "Отто" {
		t.Fatalf("coordinator name=%q", cfg.Name)
	}

	again, err := svc.CompleteSetup(ctx, dto.CompleteSetupRequest{ProjectName: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	profile, _ = repo.GetProject(ctx)
	if profile.Project.Name != "Acme Tools" || again.Needed {
		t.Fatal("second save must not rewrite a finished setup")
	}
}

func TestCompleteSetupAppliesFolders(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	ctx := context.Background()

	if _, err := svc.CompleteSetup(ctx, dto.CompleteSetupRequest{
		ProjectName: "Acme",
		Alias:       "AK",
		Name:        "Alex",
		DocsDir:     "kb",
		DataDir:     "coord-data",
	}); err != nil {
		t.Fatal(err)
	}

	layout := repo.SetupLayout()
	if layout.DocsRel != "kb" || layout.DataRel != "coord-data" {
		t.Fatalf("layout %+v", layout)
	}
	if _, err := os.Stat(filepath.Join(repo.WorkspaceDir(), "kb")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo.WorkspaceDir(), "coord-data", "settings")); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteSetupRejectsEscapePath(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	_, err := svc.CompleteSetup(context.Background(), dto.CompleteSetupRequest{
		ProjectName: "Acme",
		Alias:       "AK",
		Name:        "Alex",
		DocsDir:     "../outside",
		DataDir:     "coordinator-data",
	})
	if err == nil {
		t.Fatal("expected path error")
	}
}

func TestSlugifyProjectID(t *testing.T) {
	if got := slugifyProjectID("Acme Tools"); got != "acme-tools" {
		t.Fatal(got)
	}
	if got := slugifyProjectID("  "); got != "project" {
		t.Fatal(got)
	}
}

func TestCompleteSetupWritesTeamCollaboration(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	if _, err := svc.CompleteSetup(context.Background(), dto.CompleteSetupRequest{
		ProjectName:   "Acme",
		Alias:         "AK",
		Name:          "Alex",
		Collaboration: "team",
	}); err != nil {
		t.Fatal(err)
	}
	if repo.Collaboration() != model.CollaborationTeam {
		t.Fatalf("collaboration=%s", repo.Collaboration())
	}
}

func TestCompleteSetupWritesCoordinatorName(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	if _, err := svc.CompleteSetup(context.Background(), dto.CompleteSetupRequest{
		ProjectName:     "Acme",
		Alias:           "AK",
		Name:            "Alex",
		Language:        "en",
		CoordinatorName: "Max",
	}); err != nil {
		t.Fatal(err)
	}
	cfg := repo.CoordinatorFile(context.Background())
	if cfg.Name != "Max" {
		t.Fatalf("coordinator name=%q", cfg.Name)
	}
}

func TestCollaborationMissingFileIsTeam(t *testing.T) {
	repo := newSetupTestRepo(t)
	if repo.Collaboration() != model.CollaborationTeam {
		t.Fatalf("got %s", repo.Collaboration())
	}
}

func TestNormalizeCollaboration(t *testing.T) {
	if model.NormalizeCollaboration("solo", model.CollaborationTeam) != model.CollaborationSolo {
		t.Fatal("solo")
	}
	if model.NormalizeCollaboration("", model.CollaborationSolo) != model.CollaborationSolo {
		t.Fatal("empty uses fallback")
	}
	if model.NormalizeCollaboration("nope", model.CollaborationTeam) != model.CollaborationTeam {
		t.Fatal("unknown uses fallback")
	}
}

func newSetupTestRepo(t *testing.T) *repository.FileRepository {
	t.Helper()
	root := t.TempDir()
	ws := filepath.Join(root, "app")
	data := filepath.Join(ws, "coordinator-data")
	docs := filepath.Join(ws, "docs")
	if err := os.MkdirAll(filepath.Join(data, "settings"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	return repository.NewFileRepositoryWithPaths(repository.Paths{
		Workspace: ws,
		Data:      data,
		Docs:      docs,
		Bus:       ws,
		Cursor:    filepath.Join(ws, ".cursor"),
	})
}
