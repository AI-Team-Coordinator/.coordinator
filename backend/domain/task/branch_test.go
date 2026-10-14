package task

import "testing"

func TestCanonicalBranch(t *testing.T) {
	if CanonicalBranch("EK", "ek/feat/voice") != "feat/voice" {
		t.Fatal("strip prefix")
	}
	if CanonicalBranch("EK", "feat/voice") != "feat/voice" {
		t.Fatal("unchanged")
	}
	if CanonicalBranch("EK", "main") != "main" {
		t.Fatal("main")
	}
	if CanonicalBranch("EK", "as/feat/voice") != "as/feat/voice" {
		t.Fatal("foreign prefix kept")
	}
}
