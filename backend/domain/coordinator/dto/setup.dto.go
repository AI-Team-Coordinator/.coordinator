package dto

type SetupStateResponse struct {
	Needed          bool   `json:"needed"`
	CoordinatorName string `json:"coordinator_name"`
	ProjectName     string `json:"project_name"`
	Alias           string `json:"alias,omitempty"`
	Name            string `json:"name,omitempty"`
	Role            string `json:"role,omitempty"`
	Language        string `json:"language"`
	Layout          string `json:"layout"`
	DocsDir         string `json:"docs_dir"`
	DataDir         string `json:"data_dir"`
	Local           bool   `json:"local"`
	Collaboration   string `json:"collaboration"`
}

type CompleteSetupRequest struct {
	ProjectName   string `json:"project_name"`
	Alias         string `json:"alias"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Language      string `json:"language"`
	DocsDir       string `json:"docs_dir"`
	DataDir       string `json:"data_dir"`
	Collaboration string `json:"collaboration"`
}
