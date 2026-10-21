package model

// ConnectorActivityEntry is an operational log line for one connector (not ExternalEvent).
type ConnectorActivityEntry struct {
	ID          int64          `json:"id"`
	ConnectorID string         `json:"connector_id"`
	Kind        string         `json:"kind"`
	Level       string         `json:"level"`
	Summary     string         `json:"summary"`
	Detail      map[string]any `json:"detail,omitempty"`
	CreatedAt   string         `json:"created_at"`
}
