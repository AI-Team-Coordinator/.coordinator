package repository

import (
	"strings"
	"testing"

	"coordinator/model"
)

func TestRewriteAuthorMarkdown(t *testing.T) {
	src := `# Authors

Keep in sync.

| alias | name |
|-------|------|
| EK | Old Name |

- alias is stable.
`
	out := rewriteAuthorMarkdown(src, []model.TeamPerson{
		{Alias: "EK", Name: "Evgeny KOSIVTSOV"},
		{Alias: "AB", Name: "Ann B"},
	})
	if !strings.Contains(out, "| EK | Evgeny KOSIVTSOV |") {
		t.Fatalf("missing EK row:\n%s", out)
	}
	if !strings.Contains(out, "| AB | Ann B |") {
		t.Fatalf("missing AB row:\n%s", out)
	}
	if strings.Contains(out, "Old Name") {
		t.Fatalf("old row leaked:\n%s", out)
	}
	if !strings.Contains(out, "- alias is stable.") {
		t.Fatalf("prose lost:\n%s", out)
	}
}

func TestParseCurrentAuthor(t *testing.T) {
	got := parseCurrentAuthor([]byte("# comment\n\nEK\n"))
	if got != "EK" {
		t.Fatalf("got %q", got)
	}
}
