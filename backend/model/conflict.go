package model

// Conflict is a detected collision between active tasks or branches.
type Conflict struct {
	Severity        string
	Title           string
	Description     string
	AffectedAliases []string
}
