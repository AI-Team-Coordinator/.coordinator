package coordinator

import (
	"testing"

	"coordinator/domain/coordinator/dto"
	"coordinator/model"
)

func testProfile() *model.ProjectProfile {
	return &model.ProjectProfile{
		GitHub: model.GitHubBinding{Host: "github.com", Org: "Alina-Assist", SSHHost: "github.com-personal"},
		Groups: []model.ServiceGroup{{ID: "platform"}, {ID: "channels"}},
		Services: []model.ServiceNode{
			{ID: "core", Name: "Core", Repo: "Core", GitHubRepo: "core", Group: "platform"},
		},
	}
}

func TestNormalizeCreateService(t *testing.T) {
	node, err := normalizeCreateService(testProfile(), dto.CreateServiceRequest{
		ID:         "billing_worker",
		Name:       "Billing Worker",
		Group:      "platform",
		Kind:       "Worker",
		Repo:       "BillingWorker",
		GitHubRepo: "billing-worker",
		Purpose:    dto.LocalizedText{EN: "Invoices", RU: "Счета"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if node.Kind != "worker" || node.GitHubRepo != "billing-worker" {
		t.Fatalf("unexpected node: %+v", node)
	}

	if _, err := normalizeCreateService(testProfile(), dto.CreateServiceRequest{
		ID: "core", Name: "X", Group: "platform", Repo: "Other", GitHubRepo: "other",
	}); err == nil {
		t.Fatal("expected duplicate id")
	}

	if _, err := normalizeCreateService(&model.ProjectProfile{Groups: []model.ServiceGroup{{ID: "platform"}}}, dto.CreateServiceRequest{
		ID: "x", Name: "X", Group: "platform", Repo: "X", GitHubRepo: "x",
	}); err == nil {
		t.Fatal("expected missing org")
	}

	if _, err := normalizeCreateService(testProfile(), dto.CreateServiceRequest{
		ID: "bad/id", Name: "X", Group: "platform", Repo: "X", GitHubRepo: "x",
	}); err == nil {
		t.Fatal("expected bad id")
	}

	if _, err := normalizeCreateService(testProfile(), dto.CreateServiceRequest{
		ID: "ok", Name: "X", Group: "platform", Repo: "../etc", GitHubRepo: "x",
	}); err == nil {
		t.Fatal("expected bad folder")
	}
}

func TestSSHCloneURL(t *testing.T) {
	got := sshCloneURL(model.GitHubBinding{Org: "Alina-Assist", SSHHost: "github.com-personal"}, "billing-worker")
	want := "git@github.com-personal:Alina-Assist/billing-worker.git"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
