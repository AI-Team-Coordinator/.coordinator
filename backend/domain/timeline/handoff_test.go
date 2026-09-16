package timeline

import (
	"testing"

	"coordinator/domain/coordination"
	"coordinator/domain/identity"
	"coordinator/model"
)

func TestHandoffRoundTrip(t *testing.T) {
	h := Handoff{
		At:         100,
		TaskID:     "T-PAY",
		Branch:     "feat/pay",
		FromPerson: "EK",
		FromAgent:  identity.CursorChat("EK", "sess-1"),
		ToPerson:   "AS",
		ToAgent:    identity.CursorChat("AS", "sess-2"),
		Resource:   coordination.Parse("billing"),
		Summary:    "passing billing after checkout",
	}
	got, ok := ParseHandoff(h.Event())
	if !ok || got.ToPerson != "AS" || got.Resource.Key != "billing" || got.FromAgent.ID != "sess-1" || got.ToAgent.ID != "sess-2" {
		t.Fatalf("got ok=%v %+v", ok, got)
	}
}

func TestParseHandoffRejectsOtherKind(t *testing.T) {
	if _, ok := ParseHandoff(model.Event{Event: KindDecision}); ok {
		t.Fatal("expected reject")
	}
}

func TestDecisionRoundTrip(t *testing.T) {
	d := Decision{
		At:       200,
		TaskID:   "T-DOM",
		Person:   "EK",
		Agent:    identity.CursorChat("EK", "sess-9"),
		Choice:   "advisory claims, not file locks",
		Rejected: "Synapse-style file claims",
	}
	got, ok := ParseDecision(d.Event())
	if !ok || got.Choice != d.Choice || got.Rejected != d.Rejected || got.Agent.ID != "sess-9" {
		t.Fatalf("got ok=%v %+v", ok, got)
	}
}
