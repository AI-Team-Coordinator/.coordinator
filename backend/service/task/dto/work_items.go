package taskdto

type WorkItemsQuery struct {
	Status   string
	Assignee string
}

type WorkItemSlotResponse struct {
	Alias     string `json:"alias"`
	Status    string `json:"status"`
	Branch    string `json:"branch,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type WorkItemResponse struct {
	TaskID       string                 `json:"task_id"`
	Title        string                 `json:"title,omitempty"`
	Summary      string                 `json:"summary,omitempty"`
	Doc          string                 `json:"doc,omitempty"`
	Branch       string                 `json:"branch,omitempty"`
	Services     []string               `json:"services,omitempty"`
	CreatedBy    string                 `json:"created_by,omitempty"`
	Assignee     string                 `json:"assignee,omitempty"`
	Status       string                 `json:"status"`
	RelatedTasks []string               `json:"related_tasks,omitempty"`
	RelatedDocs  []string               `json:"related_docs,omitempty"`
	CreatedAt    string                 `json:"created_at,omitempty"`
	UpdatedAt    string                 `json:"updated_at,omitempty"`
	CompletedAt  string                 `json:"completed_at,omitempty"`
	Slots        []WorkItemSlotResponse `json:"slots,omitempty"`
}

type WorkItemsResponse struct {
	Items []WorkItemResponse `json:"items"`
	Total int                `json:"total"`
}

type CreateWorkItemRequest struct {
	TaskID       string   `json:"task_id"`
	Summary      string   `json:"summary"`
	Doc          string   `json:"doc"`
	Branch       string   `json:"branch"`
	Services     []string `json:"services"`
	Assignee     string   `json:"assignee"`
	RelatedTasks []string `json:"related_tasks"`
	RelatedDocs  []string `json:"related_docs"`
}
