package model

// TaskDoc is a read-only Common/docs markdown file for a task id.
type TaskDoc struct {
	TaskID   string
	Title    string
	RelPath  string
	Markdown string
}
