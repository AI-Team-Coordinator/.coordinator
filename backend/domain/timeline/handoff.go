package timeline

import (
	"strings"

	"coordinator/domain/coordination"
	"coordinator/domain/identity"
	"coordinator/model"
)

// Handoff transfers a zone (or the whole task) from one person/agent to another.
// It does not by itself rewrite services[] — that stays a live snapshot.
type Handoff struct {
	At         int64
	TaskID     string
	Branch     string
	FromPerson string
	FromAgent  identity.Agent
	ToPerson   string
	ToAgent    identity.Agent
	Resource   coordination.Resource
	Summary    string
}

func (h Handoff) Event() model.Event {
	ev := model.Event{
		Timestamp: h.At,
		Event:     KindHandoff,
		TaskID:    strings.TrimSpace(h.TaskID),
		Branch:    strings.TrimSpace(h.Branch),
		Alias:     strings.TrimSpace(h.FromPerson),
		Summary:   strings.TrimSpace(h.Summary),
		AgentID:   strings.TrimSpace(h.FromAgent.ID),
		ToAlias:   strings.TrimSpace(h.ToPerson),
		ToAgent:   strings.TrimSpace(h.ToAgent.ID),
	}
	if !h.Resource.IsZero() {
		ev.Service = h.Resource.Label
	}
	return ev
}

func ParseHandoff(ev model.Event) (Handoff, bool) {
	if ev.Event != KindHandoff {
		return Handoff{}, false
	}
	return Handoff{
		At:         ev.Timestamp,
		TaskID:     ev.TaskID,
		Branch:     ev.Branch,
		FromPerson: ev.Alias,
		FromAgent:  identity.CursorChat(ev.Alias, ev.AgentID),
		ToPerson:   strings.TrimSpace(ev.ToAlias),
		ToAgent:    identity.CursorChat(ev.ToAlias, ev.ToAgent),
		Resource:   coordination.Parse(ev.Service),
		Summary:    ev.Summary,
	}, true
}
