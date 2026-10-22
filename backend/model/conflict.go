package model

// ConflictTask is one side of an overlap (or related slot) for UI detail.
type ConflictTask struct {
	Alias    string
	TaskID   string
	Branch   string
	Title    string
	DocTitle string
	HasDoc   bool
}

// Conflict is a detected collision between active tasks or branches.
type Conflict struct {
	Severity        string
	Title           string
	Description     string
	AffectedAliases []string
	Service         string
	Tasks           []ConflictTask
}
