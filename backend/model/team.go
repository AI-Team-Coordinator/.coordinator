package model

// TeamPerson is a roster entry from Common/data/settings/team.json.
type TeamPerson struct {
	Alias       string   `json:"alias"`
	Name        string   `json:"name"`
	Role        string   `json:"role,omitempty"`
	Access      string   `json:"access,omitempty"`
	Focus       []string `json:"focus,omitempty"`
	Crews       []string `json:"crews,omitempty"`
	PrimaryCrew string   `json:"primary_crew,omitempty"`
	PulseScope  string   `json:"pulse_scope,omitempty"`
}

type TeamFile struct {
	Version int          `json:"version"`
	Members []TeamPerson `json:"members"`
}
