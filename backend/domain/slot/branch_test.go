package slot

import "testing"

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
