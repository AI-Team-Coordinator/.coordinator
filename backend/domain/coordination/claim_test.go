package coordination

import (
	"testing"

	"coordinator/domain/identity"
)

func TestParseClaimsAttachesAgents(t *testing.T) {
	agents := identity.CursorChats("EK", []string{"sess-1", "sess-2"})
	got := ParseClaims("EK", "T1", "Dom", "main", []string{"Core"}, agents)
	if len(got) != 1 || len(got[0].Agents) != 2 || got[0].Agents[0].ID != "sess-1" {
		t.Fatalf("got %+v", got)
	}
}
