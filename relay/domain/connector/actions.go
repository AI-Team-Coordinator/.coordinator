package connector

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ActionID is a stable machine name (e.g. poll_now). Coordinator UI maps labels via catalog.
type ActionID string

const ActionPollNow ActionID = "poll_now"

// ActionDescriptor is exposed on marketplace tiles for the actions modal.
type ActionDescriptor struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Action runs on demand from POST /connectors/{id}/actions/{action}.
type Action struct {
	ID    ActionID
	Label string
	Run   func(ctx context.Context) error
}

// ActionRegistry maps connector id → named actions (thread-safe register at startup).
type ActionRegistry struct {
	mu   sync.RWMutex
	byID map[string]map[ActionID]Action
}

func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{byID: map[string]map[ActionID]Action{}}
}

func (r *ActionRegistry) Register(connectorID string, actions ...Action) {
	connectorID = strings.TrimSpace(connectorID)
	if connectorID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byID[connectorID] == nil {
		r.byID[connectorID] = map[ActionID]Action{}
	}
	for _, a := range actions {
		if a.Run == nil || a.ID == "" {
			continue
		}
		r.byID[connectorID][a.ID] = a
	}
}

func (r *ActionRegistry) Descriptors(connectorID string) []ActionDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m := r.byID[connectorID]
	if len(m) == 0 {
		return nil
	}
	out := make([]ActionDescriptor, 0, len(m))
	for _, a := range m {
		out = append(out, ActionDescriptor{ID: string(a.ID), Label: a.Label})
	}
	return out
}

func (r *ActionRegistry) Run(ctx context.Context, connectorID string, actionID string) error {
	connectorID = strings.TrimSpace(connectorID)
	actionID = strings.TrimSpace(actionID)
	if connectorID == "" || actionID == "" {
		return fmt.Errorf("connector id and action are required")
	}
	r.mu.RLock()
	m := r.byID[connectorID]
	var action Action
	if m != nil {
		action = m[ActionID(actionID)]
	}
	r.mu.RUnlock()
	if action.Run == nil {
		return fmt.Errorf("unknown action %q for connector %q", actionID, connectorID)
	}
	return action.Run(ctx)
}
