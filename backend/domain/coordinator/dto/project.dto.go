package dto

type LocalizedText struct {
	EN string `json:"en"`
	RU string `json:"ru"`
}

type ProjectMetaResponse struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Tagline LocalizedText `json:"tagline"`
}

type ServiceGroupResponse struct {
	ID    string        `json:"id"`
	Label LocalizedText `json:"label"`
}

type GitHubBindingResponse struct {
	Host    string `json:"host"`
	Org     string `json:"org"`
	SSHHost string `json:"ssh_host,omitempty"`
	HTMLURL string `json:"html_url"`
}

type GitHubLiveResponse struct {
	Auth   string `json:"auth"`
	Login  string `json:"login,omitempty"`
	Role   string `json:"role,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type ServiceNodeResponse struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Group      string        `json:"group"`
	Kind       string        `json:"kind"`
	Repo       string        `json:"repo"`
	GitHubRepo string        `json:"github_repo,omitempty"`
	GitHubOrg  string        `json:"github_org,omitempty"`
	HTMLURL    string        `json:"html_url,omitempty"`
	Purpose    LocalizedText `json:"purpose"`
}

type CreateServiceRequest struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Group      string        `json:"group"`
	Kind       string        `json:"kind"`
	Repo       string        `json:"repo"`
	GitHubRepo string        `json:"github_repo"`
	Purpose    LocalizedText `json:"purpose"`
}

type CreateServiceResponse struct {
	Project *ProjectResponse `json:"project"`
	Cloned  bool             `json:"cloned"`
	HTMLURL string           `json:"html_url"`
	Pushed  bool             `json:"pushed,omitempty"`
	Warning string           `json:"warning,omitempty"`
}

type ServiceEdgeResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
	Via  string `json:"via"`
	Kind string `json:"kind"`
}

type ProjectResponse struct {
	Version    int                    `json:"version"`
	Project    ProjectMetaResponse    `json:"project"`
	GitHub     *GitHubBindingResponse `json:"github,omitempty"`
	GitHubLive *GitHubLiveResponse    `json:"github_live,omitempty"`
	Groups     []ServiceGroupResponse `json:"groups"`
	Services   []ServiceNodeResponse  `json:"services"`
	Edges      []ServiceEdgeResponse  `json:"edges"`
}
