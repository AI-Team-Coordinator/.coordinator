package model

// OrgTopology is settings/org_topology.json (crews and CrewConnector links).
type OrgTopology struct {
	Version        int             `json:"version"`
	Crews          []Crew          `json:"crews"`
	CrewConnectors []CrewConnector `json:"crew_connectors"`
}

type Crew struct {
	ID            string        `json:"id"`
	Label         LocalizedText `json:"label"`
	Kind          string        `json:"kind"`
	InterestBrief string        `json:"interest_brief,omitempty"`
	Notes         string        `json:"notes,omitempty"`
}

type CrewConnector struct {
	ID         string              `json:"id"`
	FromCrewID string              `json:"from_crew_id"`
	ToCrewID   string              `json:"to_crew_id"`
	Summary    string              `json:"summary,omitempty"`
	Notes      string              `json:"notes,omitempty"`
	Interests  []CrewConnectorItem `json:"interests"`
}

type CrewConnectorItem struct {
	Title  string   `json:"title"`
	Detail string   `json:"detail,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}
