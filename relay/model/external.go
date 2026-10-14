package model

// ExternalEvent is one normalized fact written to the coordinator-state bus.
// It is not a human model.Event.
type ExternalEvent struct {
	ID          string   `json:"id"`
	Source      string   `json:"source"`
	OriginID    string   `json:"origin_id"`
	ExternalID  string   `json:"external_id"`
	OccurredAt  string   `json:"occurred_at,omitempty"`
	Kind        string   `json:"kind"`
	Title       string   `json:"title,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	URL         string   `json:"url,omitempty"`
	Repo        string   `json:"repo,omitempty"`
	Branch      string   `json:"branch,omitempty"`
	GitHubLogin string   `json:"github_login,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}
