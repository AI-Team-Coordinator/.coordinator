package model

import "testing"

func TestMessageHasTaskID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"feat(core): add daily billing (20260907-2129-EK-ADMIN_STATS_MONTH_PACE)", true},
		{"fix(core): avatar (FIX-20260907-1825-EK-AVATAR_MISSING_STYLE)", true},
		{"chore(prompts): add local movetocascais.com DESCRIPTION_ONLINE mirror", false},
		{"test(core): assert ARC closed-lost flag persists via JSON settings round-trip", false},
	}
	for _, tc := range cases {
		if got := MessageHasTaskID(tc.in); got != tc.want {
			t.Errorf("%q: got %v want %v", tc.in, got, tc.want)
		}
	}
}

func TestValidTaskID(t *testing.T) {
	if !ValidTaskID("20260907-2129-EK-ADMIN_STATS_MONTH_PACE") {
		t.Fatal("valid")
	}
	if !ValidTaskID("FIX-20260907-1825-EK-AVATAR_MISSING_STYLE") {
		t.Fatal("fix")
	}
	if ValidTaskID("../secret") || ValidTaskID("a/b") || ValidTaskID("") {
		t.Fatal("invalid accepted")
	}
}

func TestMessageHasThisTask(t *testing.T) {
	id := "20260907-2129-EK-ADMIN_STATS_MONTH_PACE"
	if !MessageHasThisTask("feat(core): merge feat/x ("+id+")", id) {
		t.Fatal("expected match")
	}
	if MessageHasThisTask("feat(core): merge feat/x ("+id+")", "FIX-20260907-1825-EK-AVATAR_MISSING_STYLE") {
		t.Fatal("did not expect other task")
	}
}
