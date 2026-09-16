package coordination

import "testing"

func TestSelfOverlapsProduct(t *testing.T) {
	claims := append(
		ParseClaims("EK", "T-ARC", "ARC", "feat/arc", []string{"Core"}, nil),
		ParseClaims("EK", "T-CORE2", "Other", "feat/other", []string{"Core", "InboxPanelWeb"}, nil)...,
	)
	got := SelfOverlaps(claims)
	if len(got) != 1 || got[0].Title != TitleSelfOverlap || got[0].Service != "Core" {
		t.Fatalf("got %+v", got)
	}
}

func TestSelfOverlapsCoordinatorAndCoreOK(t *testing.T) {
	claims := append(
		ParseClaims("EK", "T-ARC", "ARC", "feat/arc", []string{"Core"}, nil),
		ParseClaims("EK", "T-COORD", "Coord", "main", []string{"Common", ".cursor", ".coordinator"}, nil)...,
	)
	if got := SelfOverlaps(claims); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestSelfOverlapsWorkspace(t *testing.T) {
	claims := append(
		ParseClaims("EK", "T1", "A", "main", []string{".coordinator"}, nil),
		ParseClaims("EK", "T2", "B", "main", []string{".coordinator"}, nil)...,
	)
	got := SelfOverlaps(claims)
	if len(got) != 1 || got[0].Severity != "critical" {
		t.Fatalf("got %+v", got)
	}
}

func TestPeerOverlapsProduct(t *testing.T) {
	claims := append(
		ParseClaims("EK", "T-VOICE", "Web voice", "feat/web-voice-mvp", []string{"LLM", "WebChat"}, nil),
		ParseClaims("AS", "T-LLM", "LLM timeout", "fix/llm-timeout", []string{"LLM"}, nil)...,
	)
	got := PeerOverlaps(claims)
	if len(got) != 1 || got[0].Title != TitlePeerOverlap || got[0].Service != "LLM" || got[0].Severity != "warning" {
		t.Fatalf("got %+v", got)
	}
}

func TestPeerOverlapsBusAndWorkspaceIgnored(t *testing.T) {
	claims := append(
		ParseClaims("EK", "T-CLOCK", "Task clock", "main", []string{"Common", ".cursor", ".coordinator"}, nil),
		ParseClaims("AS", "T-RULES", "Rules", "main", []string{"Common", ".cursor", ".coordinator"}, nil)...,
	)
	if got := PeerOverlaps(claims); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestPeerOverlapsPrefixPath(t *testing.T) {
	claims := append(
		ParseClaims("EK", "T-PAY", "Payments", "feat/pay", []string{"billing"}, nil),
		ParseClaims("AS", "T-INV", "Invoices", "feat/inv", []string{"billing/invoices"}, nil)...,
	)
	got := PeerOverlaps(claims)
	if len(got) != 1 || got[0].Severity != "warning" {
		t.Fatalf("got %+v", got)
	}
}
