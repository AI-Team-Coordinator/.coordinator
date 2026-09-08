package model

type LocalizedText struct {
	EN string `json:"en"`
	RU string `json:"ru"`
}

type ProjectMeta struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Tagline LocalizedText `json:"tagline"`
}

type ServiceGroup struct {
	ID    string        `json:"id"`
	Label LocalizedText `json:"label"`
}

type GitHubBinding struct {
	Host    string `json:"host"`
	Org     string `json:"org"`
	SSHHost string `json:"ssh_host,omitempty"`
}

type ServiceNode struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Group      string        `json:"group"`
	Kind       string        `json:"kind"`
	Repo       string        `json:"repo"`
	GitHubRepo string        `json:"github_repo,omitempty"`
	GitHubOrg  string        `json:"github_org,omitempty"`
	Purpose    LocalizedText `json:"purpose"`
}

type ServiceEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Via  string `json:"via"`
	Kind string `json:"kind"`
}

type ProjectProfile struct {
	Version  int            `json:"version"`
	Project  ProjectMeta    `json:"project"`
	GitHub   GitHubBinding  `json:"github"`
	Groups   []ServiceGroup `json:"groups"`
	Services []ServiceNode  `json:"services"`
	Edges    []ServiceEdge  `json:"edges"`
}
