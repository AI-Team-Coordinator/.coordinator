package model

import "strings"

const (
	CollaborationSolo = "solo"
	CollaborationTeam = "team"
)

// CoordinatorFile is settings/coordinator.json: who the Coordinator is
// and how task snapshots are stored.
// Name has no editor yet — the file is the only way to change it.
type CoordinatorFile struct {
	Version       int    `json:"version"`
	Name          string `json:"name"`
	Collaboration string `json:"collaboration,omitempty"`
}

// NormalizeCollaboration returns solo or team.
// Empty / unknown values use fallback (team at runtime, solo in the setup wizard).
func NormalizeCollaboration(raw, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case CollaborationSolo:
		return CollaborationSolo
	case CollaborationTeam:
		return CollaborationTeam
	default:
		if fallback == CollaborationSolo || fallback == CollaborationTeam {
			return fallback
		}
		return CollaborationTeam
	}
}
