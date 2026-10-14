package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	settingsdto "coordinator/service/settings/dto"
)

func TestGetUserSettingsRequiresAuthor(t *testing.T) {
	repo := newSetupTestRepo(t)
	svc := NewService(repo, nil)
	_, err := svc.GetUserSettings(context.Background())
	if err == nil {
		t.Fatal("expected error without current author")
	}
}

func TestSaveUserSettingsLanguage(t *testing.T) {
	repo := newSetupTestRepo(t)
	data := repo.WorkspaceDir()
	dataDir := filepath.Join(data, "coordinator-data")
	if err := os.WriteFile(filepath.Join(dataDir, ".current_author"), []byte("EK\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(dataDir, "settings")
	if err := os.WriteFile(filepath.Join(settings, "locale.json"), []byte(`{"version":1,"chat_language":"ru","docs_language":"ru"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(settings, "team.json"), []byte(`{"version":1,"members":[{"alias":"EK","name":"Test"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(settings, "project_profile.json"), []byte(`{"version":1,"project":{"name":"p"},"setup":{"completed":true},"services":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	got, err := svc.GetUserSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Language != "ru" {
		t.Fatalf("language %q", got.Language)
	}

	_, err = svc.SaveUserSettings(ctx, settingsdto.SaveUserSettingsRequest{
		Name:     "Test",
		Language: "en",
	})
	if err != nil {
		t.Fatal(err)
	}
	loc, err := repo.GetLocale(ctx)
	if err != nil || loc.ChatLanguage != "en" {
		t.Fatalf("locale %+v %v", loc, err)
	}
}
