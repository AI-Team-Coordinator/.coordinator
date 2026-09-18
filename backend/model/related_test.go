package model

import "testing"

func TestNormalizeRelated(t *testing.T) {
	tasks := NormalizeRelatedTasks([]string{" A ", "A", "SELF", "", "B"}, "SELF")
	if len(tasks) != 2 || tasks[0] != "A" || tasks[1] != "B" {
		t.Fatalf("tasks=%v", tasks)
	}
	docs := NormalizeRelatedDocs([]string{
		"Common/docs/FOO.md",
		"./docs/FOO.md",
		"InboxPanelWeb/BAR.md",
		"/etc/passwd",
		"",
	})
	if len(docs) != 2 || docs[0] != "docs/FOO.md" || docs[1] != "docs/InboxPanelWeb/BAR.md" {
		t.Fatalf("docs=%v", docs)
	}
}

func TestTaskIDFromDocPath(t *testing.T) {
	if got := TaskIDFromDocPath("docs/20260918-0043-EK-COORDINATOR_FIX_RELATIONS.md"); got != "20260918-0043-EK-COORDINATOR_FIX_RELATIONS" {
		t.Fatalf("got %q", got)
	}
	if got := TaskIDFromDocPath("docs/README.md"); got != "" {
		t.Fatalf("readme=%q", got)
	}
}
