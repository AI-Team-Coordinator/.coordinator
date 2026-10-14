package task

import "time"

// Task is a unit of work. It can sit in the backlog with no slot.
type Task struct {
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
