package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	relaydto "coordinator/service/relay/dto"
	"coordinator/service/repository"
)

func TestIssueRelayTokenAdminOnly(t *testing.T) {
	bus := t.TempDir()
	dataDir := filepath.Join(bus, "data")
	settings := filepath.Join(dataDir, "settings")
	if err := os.MkdirAll(settings, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, ".current_author"), []byte("EK\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	team := `{"version":1,"members":[{"alias":"EK","name":"E","access":"admin"},{"alias":"AS","name":"A","access":"member"}]}`
	if err := os.WriteFile(filepath.Join(settings, "team.json"), []byte(team), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewFileRepository(bus, "")
	svc := &Service{repo: repo}
	ctx := context.Background()

	out, err := svc.IssueRelayToken(ctx, relaydto.RelayIssueTokenRequest{Alias: "AS", Label: "laptop"})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if out.Secret == "" || out.Token.Alias != "AS" {
		t.Fatalf("unexpected issue response: %+v", out)
	}
	list, err := svc.ListRelayIssuedTokens(ctx)
	if err != nil || len(list.Tokens) != 1 {
		t.Fatalf("list: %v len=%d", err, len(list.Tokens))
	}

	if err := os.WriteFile(filepath.Join(dataDir, ".current_author"), []byte("AS\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.IssueRelayToken(ctx, relaydto.RelayIssueTokenRequest{}); err == nil {
		t.Fatal("member should not issue tokens")
	}
}
