package orgdto

import "coordinator/model"

type OrgTopologyResponse struct {
	Version        int                     `json:"version"`
	Crews          []CrewResponse          `json:"crews"`
	CrewConnectors []CrewConnectorResponse `json:"crew_connectors"`
	Pushed         bool                    `json:"pushed,omitempty"`
	Warning        string                  `json:"warning,omitempty"`
}

type CrewResponse struct {
	ID            string              `json:"id"`
	Label         model.LocalizedText `json:"label"`
	Kind          string              `json:"kind"`
	InterestBrief string              `json:"interest_brief,omitempty"`
	Notes         string              `json:"notes,omitempty"`
}

type CrewConnectorResponse struct {
	ID         string                      `json:"id"`
	FromCrewID string                      `json:"from_crew_id"`
	ToCrewID   string                      `json:"to_crew_id"`
	Summary    string                      `json:"summary,omitempty"`
	Notes      string                      `json:"notes,omitempty"`
	Interests  []CrewConnectorItemResponse `json:"interests"`
}

type CrewConnectorItemResponse struct {
	Title  string   `json:"title"`
	Detail string   `json:"detail,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

type SaveOrgTopologyRequest struct {
	Version        int                    `json:"version"`
	Crews          []CrewRequest          `json:"crews"`
	CrewConnectors []CrewConnectorRequest `json:"crew_connectors"`
}

type CrewRequest struct {
	ID            string              `json:"id"`
	Label         model.LocalizedText `json:"label"`
	Kind          string              `json:"kind"`
	InterestBrief string              `json:"interest_brief"`
	Notes         string              `json:"notes"`
}

type CrewConnectorRequest struct {
	ID         string                     `json:"id"`
	FromCrewID string                     `json:"from_crew_id"`
	ToCrewID   string                     `json:"to_crew_id"`
	Summary    string                     `json:"summary"`
	Notes      string                     `json:"notes"`
	Interests  []CrewConnectorItemRequest `json:"interests"`
}

type CrewConnectorItemRequest struct {
	Title  string   `json:"title"`
	Detail string   `json:"detail"`
	Tags   []string `json:"tags"`
}
