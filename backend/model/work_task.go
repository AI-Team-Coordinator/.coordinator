package model

import "time"

// WorkTask is a coordinator work item (backlog / open / done), separate from a member snapshot slot.
type WorkTask struct {
	TaskID       string
	Summary      string
	Doc          string
	Branch       string
	Services     []string
	CreatedBy    string
	Assignee     string
	Status       string // backlog, open, done
	RelatedTasks []string
	RelatedDocs  []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  time.Time
}

// WorkTaskSlot is a live or parked process bound to a work task.
type WorkTaskSlot struct {
	Alias    string
	Status   string
	Branch   string
	UpdatedAt time.Time
}
