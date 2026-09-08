package dto

type TaskDocResponse struct {
	TaskID   string `json:"task_id"`
	Title    string `json:"title"`
	RelPath  string `json:"rel_path"`
	Markdown string `json:"markdown"`
}
