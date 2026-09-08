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
