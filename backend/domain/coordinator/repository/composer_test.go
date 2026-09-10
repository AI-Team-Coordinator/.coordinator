package repository

import "testing"

func TestCollectSessionIDsMergesAndDedups(t *testing.T) {
	got := collectSessionIDs("a", []string{"b", "a", " c ", ""})
	if len(got) != 3 || got[0] != "b" || got[1] != "a" || got[2] != "c" {
		t.Fatalf("got %#v", got)
	}
	got = collectSessionIDs("only", nil)
	if len(got) != 1 || got[0] != "only" {
		t.Fatalf("got %#v", got)
	}
}

func TestChatTabsUsesLookup(t *testing.T) {
	titles := map[string]string{"abc": "Calculator prompt update"}
	got := chatTabs([]string{"abc", "xyz"}, func(id string) string { return titles[id] })
	if len(got) != 2 {
		t.Fatalf("len %d", len(got))
	}
	if got[0].Title != "Calculator prompt update" || got[0].SessionID != "abc" {
		t.Fatalf("first %#v", got[0])
	}
	if got[1].Title != "" || got[1].SessionID != "xyz" {
		t.Fatalf("second %#v", got[1])
	}
}
