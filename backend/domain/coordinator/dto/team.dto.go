package dto

type TeamPersonResponse struct {
	Alias  string   `json:"alias"`
	Name   string   `json:"name"`
	Role   string   `json:"role,omitempty"`
	Access string   `json:"access,omitempty"`
	Focus  []string `json:"focus,omitempty"`
}

type TeamResponse struct {
	Members []TeamPersonResponse `json:"members"`
	Pushed  bool                 `json:"pushed,omitempty"`
	Warning string               `json:"warning,omitempty"`
}

type TeamPersonRequest struct {
	Alias  string   `json:"alias"`
	Name   string   `json:"name"`
	Role   string   `json:"role"`
	Access string   `json:"access"`
	Focus  []string `json:"focus"`
}

type SaveTeamRequest struct {
	Members []TeamPersonRequest `json:"members"`
}
