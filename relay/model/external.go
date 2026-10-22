package model

// ExternalEvent is one normalized fact on the coordinator-state bus (not model.Event).
//
// Envelope fields are source-agnostic. Connector-specific payload lives in Data.
// ReceivedAt is set by the bus when the line is appended.
const ExternalEventSchemaVersion = 1

type ExternalEvent struct {
	SchemaVersion int            `json:"schema_version"`
	ID            string         `json:"id"`
	Source        string         `json:"source"`
	ExternalID    string         `json:"external_id"`
	ReceivedAt    string         `json:"received_at,omitempty"`
	OccurredAt    string         `json:"occurred_at,omitempty"`
	Class         string         `json:"class"`
	Kind          string         `json:"kind"`
	Title         string         `json:"title,omitempty"`
	Summary       string         `json:"summary,omitempty"`
	URL           string         `json:"url,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Data              map[string]any `json:"data,omitempty"`
	RequiresReaction  bool           `json:"requires_reaction,omitempty"`
}

// Seal fills schema version and defaults before the bus writes the line.
func (e *ExternalEvent) Seal(receivedAt string) {
	if e.SchemaVersion == 0 {
		e.SchemaVersion = ExternalEventSchemaVersion
	}
	if e.ReceivedAt == "" {
		e.ReceivedAt = receivedAt
	}
	if e.Class == "" {
		e.Class = ClassInfo
	}
}
