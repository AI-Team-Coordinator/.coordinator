package model

// CoordinatorFile is settings/coordinator.json: who the Coordinator is.
// Name has no editor yet — the file is the only way to change it.
type CoordinatorFile struct {
	Version int    `json:"version"`
	Name    string `json:"name"`
}
