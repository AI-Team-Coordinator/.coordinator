package model

// Event is one append-only log line from data/progress/events/<ALIAS>/<YYYY>.jsonl.
type Event struct {
	Timestamp       int64    `json:"timestamp"`
	Event           string   `json:"event"`
	TaskID          string   `json:"task_id"`
	Branch          string   `json:"branch,omitempty"`
	Alias           string   `json:"alias,omitempty"`
	Repo            string   `json:"repo,omitempty"`
	Service         string   `json:"service,omitempty"`
	Status          string   `json:"status,omitempty"`
	CostUSD         *float64 `json:"cost_usd,omitempty"`
	BudgetUSD       *float64 `json:"budget_usd,omitempty"`
	OnDemandUSD     *float64 `json:"ondemand_usd,omitempty"`
	CursorModelsPct *float64 `json:"cursor_models_pct,omitempty"`
	OtherModelsPct  *float64 `json:"other_models_pct,omitempty"`
	UsagePlan       string   `json:"usage_plan,omitempty"`
	PlanPriceUSD    *float64 `json:"plan_price_usd,omitempty"`
	SpendKind       string   `json:"spend_kind,omitempty"`
}

// EventQuery filters the SQLite event cache. Limit 0 means no page cap.
type EventQuery struct {
	Alias   string
	Event   string
	Service string
	TaskID  string
	Since   int64
	Limit   int
	Offset  int
}

// RepoWork is live git status of one service repo for the active task.
type RepoWork struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Repo     string `json:"repo"`
	Kind     string `json:"kind,omitempty"`
	State    string `json:"state"`
	Ahead    int    `json:"ahead,omitempty"`
	Dirty    bool   `json:"dirty,omitempty"`
	Deployed bool   `json:"deployed,omitempty"`
}

// StrayRepo is leftover commits on a topic branch that mention no task id.
type StrayRepo struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Repo    string        `json:"repo"`
	Branch  string        `json:"branch"`
	Commits []StrayCommit `json:"commits"`
}

type StrayCommit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}
