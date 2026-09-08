package dto

type SyncStatusResponse struct {
	Alias      string   `json:"alias"`
	Branch     string   `json:"branch"`
	DirtyFiles []string `json:"dirty_files"`
}

type SyncResponse struct {
	Alias     string   `json:"alias"`
	Committed bool     `json:"committed"`
	Pushed    bool     `json:"pushed"`
	Files     []string `json:"files"`
	Message   string   `json:"message"`
}
