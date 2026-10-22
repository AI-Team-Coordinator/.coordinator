package model

// ExternalEventTriageStatus is Coordinator overlay state (not on the git bus).
type ExternalEventTriageStatus string

const (
	ExternalTriageOpen       ExternalEventTriageStatus = "open"
	ExternalTriageInProgress ExternalEventTriageStatus = "in_progress"
	ExternalTriageResolved   ExternalEventTriageStatus = "resolved"
	ExternalTriageDismissed  ExternalEventTriageStatus = "dismissed"
)

func ValidExternalTriageStatus(s string) bool {
	switch ExternalEventTriageStatus(s) {
	case ExternalTriageOpen, ExternalTriageInProgress, ExternalTriageResolved, ExternalTriageDismissed:
		return true
	default:
		return false
	}
}

// ExternalEventTriage is keyed by source + external_id in local storage.
type ExternalEventTriage struct {
	Source        string                    `json:"source"`
	ExternalID    string                    `json:"external_id"`
	Status        ExternalEventTriageStatus   `json:"status"`
	AssignedAlias string                    `json:"assigned_alias,omitempty"`
	AssignedAt    string                    `json:"assigned_at,omitempty"`
	LinkedTaskID  string                    `json:"linked_task_id,omitempty"`
	ResolvedBy    string                    `json:"resolved_by,omitempty"`
	ResolvedAt    string                    `json:"resolved_at,omitempty"`
	DismissNote   string                    `json:"dismiss_note,omitempty"`
	UpdatedAt     string                    `json:"updated_at,omitempty"`
}

// ExternalEventRequiresReaction returns envelope flag or Coordinator defaults per source/kind.
func ExternalEventRequiresReaction(source, kind string, envelope bool) bool {
	if envelope {
		return true
	}
	if source == "llm_model_lifecycle" {
		switch kind {
		case "llm_model.deprecating", "llm_model.expired", "llm_model.removed":
			return true
		}
	}
	return false
}
