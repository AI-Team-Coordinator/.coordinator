package model

type SyncStatus struct {
	Alias      string   `json:"alias"`
	Branch     string   `json:"branch"`
	DirtyFiles []string `json:"dirty_files"`
}

type SyncResult struct {
	Alias     string   `json:"alias"`
	Committed bool     `json:"committed"`
	Pushed    bool     `json:"pushed"`
	Files     []string `json:"files"`
	Message   string   `json:"message"`
}
