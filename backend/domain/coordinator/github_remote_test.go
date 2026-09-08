package coordinator

import (
	"testing"

	"coordinator/model"
)

func TestParseGitHubRemote(t *testing.T) {
	cases := []struct {
		raw  string
		org  string
		slug string
	}{
		{"git@github.com-personal:Alina-Assist/core.git", "Alina-Assist", "core"},
		{"git@github.com:Alina-Assist/.cursor.git", "Alina-Assist", ".cursor"},
		{"git@github.com-personal:AI-Team-Coordinator/.coordinator.git", "AI-Team-Coordinator", ".coordinator"},
		{"https://github.com/Alina-Assist/inbox-panel.git", "Alina-Assist", "inbox-panel"},
		{"https://github.com/Alina-Assist/inbox-panel", "Alina-Assist", "inbox-panel"},
		{"ssh://git@github.com/Alina-Assist/common.git", "Alina-Assist", "common"},
		{"", "", ""},
	}
	for _, tc := range cases {
		org, slug := parseGitHubRemote(tc.raw)
		if org != tc.org || slug != tc.slug {
			t.Fatalf("%q: got %s/%s want %s/%s", tc.raw, org, slug, tc.org, tc.slug)
		}
	}
}

func TestGitHubRepoHTMLURL(t *testing.T) {
	binding := model.GitHubBinding{Host: "github.com", Org: "Alina-Assist"}
	got := githubRepoHTMLURL(binding, "core")
	if got != "https://github.com/Alina-Assist/core" {
		t.Fatalf("got %s", got)
	}
}

func TestServiceHTMLURLOverridesOrg(t *testing.T) {
	profile := model.GitHubBinding{Host: "github.com", Org: "Alina-Assist"}
	svc := model.ServiceNode{GitHubRepo: ".coordinator", GitHubOrg: "AI-Team-Coordinator"}
	got := serviceHTMLURL(profile, svc)
	if got != "https://github.com/AI-Team-Coordinator/.coordinator" {
		t.Fatalf("got %s", got)
	}
}
