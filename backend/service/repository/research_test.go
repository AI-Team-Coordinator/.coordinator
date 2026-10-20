package repository

import (
	"testing"
	"time"

	"coordinator/model"
)

func TestResearchIsStaleUsesSessionActivity(t *testing.T) {
	rs := &model.Research{
		Status:    "active",
		StartedAt: time.Now().Add(-2 * time.Hour),
		SessionID: "abc",
	}
	fresh := composerSessionsCache{Sessions: map[string]composerSessionRec{
		"abc": {LastAgentAt: time.Now().Add(-2 * time.Minute).Unix()},
	}}
	if researchIsStale(rs, fresh) {
		t.Fatal("fresh session should not be stale")
	}
	old := composerSessionsCache{Sessions: map[string]composerSessionRec{
		"abc": {LastAgentAt: time.Now().Add(-2 * time.Hour).Unix()},
	}}
	if !researchIsStale(rs, old) {
		t.Fatal("idle session should be stale")
	}
}

func TestResearchIsStaleFallsBackToStartedAt(t *testing.T) {
	rs := &model.Research{Status: "active", StartedAt: time.Now().Add(-2 * time.Hour)}
	if !researchIsStale(rs, composerSessionsCache{}) {
		t.Fatal("old start without session activity should be stale")
	}
	rs.StartedAt = time.Now().Add(-2 * time.Minute)
	if researchIsStale(rs, composerSessionsCache{}) {
		t.Fatal("recent start should not be stale")
	}
}
