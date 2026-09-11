package dto

import "time"

type MemberResponse struct {
	Alias           string               `json:"alias"`
	Name            string               `json:"name"`
	Role            string               `json:"role,omitempty"`
	Access          string               `json:"access,omitempty"`
	Focus           []string             `json:"focus,omitempty"`
	Services        []string             `json:"services,omitempty"`
	Status          string               `json:"status"`
	TaskID          string               `json:"task_id,omitempty"`
	TaskTitle       string               `json:"task_title,omitempty"`
	TaskDoc         string               `json:"task_doc,omitempty"`
	TaskSummary     string               `json:"task_summary,omitempty"`
	Branch          string               `json:"branch,omitempty"`
	UpdatedAt       time.Time            `json:"updated_at"`
	DurationSeconds int64                `json:"duration_seconds,omitempty"`
	Repos           []RepoWorkResponse   `json:"repos,omitempty"`
	CostUSD         *float64             `json:"cost_usd,omitempty"`
	BudgetUSD       *float64             `json:"budget_usd,omitempty"`
	OnDemandUSD     *float64             `json:"ondemand_usd,omitempty"`
	CursorModelsPct *float64             `json:"cursor_models_pct,omitempty"`
	OtherModelsPct  *float64             `json:"other_models_pct,omitempty"`
	SpendKind       string               `json:"spend_kind,omitempty"`
	Research        *ResearchResponse    `json:"research,omitempty"`
	Tasks           []MemberTaskResponse `json:"tasks,omitempty"`
}

type MemberTaskResponse struct {
	TaskID          string             `json:"task_id"`
	TaskTitle       string             `json:"task_title,omitempty"`
	TaskDoc         string             `json:"task_doc,omitempty"`
	TaskSummary     string             `json:"task_summary,omitempty"`
	Branch          string             `json:"branch,omitempty"`
	Services        []string           `json:"services,omitempty"`
	StartedAt       time.Time          `json:"started_at,omitempty"`
	DurationSeconds int64              `json:"duration_seconds,omitempty"`
	Repos           []RepoWorkResponse `json:"repos,omitempty"`
	CostUSD         *float64           `json:"cost_usd,omitempty"`
	BudgetUSD       *float64           `json:"budget_usd,omitempty"`
	OnDemandUSD     *float64           `json:"ondemand_usd,omitempty"`
	CursorModelsPct *float64           `json:"cursor_models_pct,omitempty"`
	OtherModelsPct  *float64           `json:"other_models_pct,omitempty"`
	SpendKind       string             `json:"spend_kind,omitempty"`
	Chats           []ChatTabResponse  `json:"chats,omitempty"`
}

type ChatTabResponse struct {
	Title     string `json:"title"`
	SessionID string `json:"session_id,omitempty"`
}

type ResearchResponse struct {
	Status          string           `json:"status"`
	Summary         string           `json:"summary,omitempty"`
	StartedAt       time.Time        `json:"started_at,omitempty"`
	DurationSeconds int64            `json:"duration_seconds,omitempty"`
	CostUSD         *float64         `json:"cost_usd,omitempty"`
	BudgetUSD       *float64         `json:"budget_usd,omitempty"`
	OnDemandUSD     *float64         `json:"ondemand_usd,omitempty"`
	CursorModelsPct *float64         `json:"cursor_models_pct,omitempty"`
	OtherModelsPct  *float64         `json:"other_models_pct,omitempty"`
	Chat            *ChatTabResponse `json:"chat,omitempty"`
}

type RepoWorkResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Repo     string `json:"repo"`
	Kind     string `json:"kind,omitempty"`
	State    string `json:"state"`
	Ahead    int    `json:"ahead,omitempty"`
	Dirty    bool   `json:"dirty,omitempty"`
	Deployed bool   `json:"deployed,omitempty"`
}

type ConflictResponse struct {
	Severity        string   `json:"severity"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	AffectedAliases []string `json:"affected_aliases"`
}

type PulseResponse struct {
	UpdatedAt time.Time           `json:"updated_at"`
	Members   []MemberResponse    `json:"members"`
	Conflicts []ConflictResponse  `json:"conflicts"`
	Stray     []StrayRepoResponse `json:"stray"`
}

type StrayRepoResponse struct {
	ID      string                `json:"id"`
	Name    string                `json:"name"`
	Repo    string                `json:"repo"`
	Branch  string                `json:"branch"`
	Commits []StrayCommitResponse `json:"commits"`
}

type StrayCommitResponse struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

type MembersResponse struct {
	Members []MemberResponse `json:"members"`
}

type ConflictsResponse struct {
	Conflicts []ConflictResponse `json:"conflicts"`
}
