package coordination

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

func TestBranchMatches(t *testing.T) {
	if !BranchMatches("EK", "feat/voice", "ek/feat/voice") {
		t.Fatal("prefixed git")
	}
	if !BranchMatches("EK", "ek/feat/voice", "feat/voice") {
		t.Fatal("canonical slot")
	}
	if BranchMatches("EK", "feat/voice", "as/feat/voice") {
		t.Fatal("wrong owner prefix")
	}
}
