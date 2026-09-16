package timeline

import (
	"strings"

	"coordinator/domain/identity"
	"coordinator/model"
)

// Decision is a typed "why": the choice we kept, and what we rejected.
type Decision struct {
	At       int64
	TaskID   string
	Person   string
	Agent    identity.Agent
	Choice   string
	Rejected string
}

func (d Decision) Event() model.Event {
	return model.Event{
		Timestamp: d.At,
		Event:     KindDecision,
		TaskID:    strings.TrimSpace(d.TaskID),
		Alias:     strings.TrimSpace(d.Person),
		Summary:   strings.TrimSpace(d.Choice),
		Findings:  strings.TrimSpace(d.Rejected),
		AgentID:   strings.TrimSpace(d.Agent.ID),
	}
}

func ParseDecision(ev model.Event) (Decision, bool) {
	if ev.Event != KindDecision {
		return Decision{}, false
	}
	return Decision{
		At:       ev.Timestamp,
		TaskID:   ev.TaskID,
		Person:   ev.Alias,
		Agent:    identity.CursorChat(ev.Alias, ev.AgentID),
		Choice:   ev.Summary,
		Rejected: ev.Findings,
	}, true
}
