package dto

type TaskResponse struct {
	TaskID          string   `json:"task_id"`
	Title           string   `json:"title,omitempty"`
	Status          string   `json:"status"`
	Kind            string   `json:"kind,omitempty"`
	Alias           string   `json:"alias,omitempty"`
	Branch          string   `json:"branch,omitempty"`
	Services        []string `json:"services,omitempty"`
	StartedAt       int64    `json:"started_at,omitempty"`
	CompletedAt     int64    `json:"completed_at,omitempty"`
	DurationSeconds int64    `json:"duration_seconds,omitempty"`
	ClockPaused     bool     `json:"clock_paused,omitempty"`
	CostUSD         *float64 `json:"cost_usd,omitempty"`
	BudgetUSD       *float64 `json:"budget_usd,omitempty"`
	OnDemandUSD     *float64 `json:"ondemand_usd,omitempty"`
	CursorModelsPct *float64 `json:"cursor_models_pct,omitempty"`
	OtherModelsPct  *float64 `json:"other_models_pct,omitempty"`
	SpendKind       string   `json:"spend_kind,omitempty"`
	SpendShared     bool     `json:"spend_shared,omitempty"`
}

type TasksResponse struct {
	Tasks  []TaskResponse `json:"tasks"`
	Total  int            `json:"total"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
}

type TasksQuery struct {
	Status string
	Alias  string
	Kind   string
	Limit  int
	Offset int
}
